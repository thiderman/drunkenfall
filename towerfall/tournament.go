package towerfall

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/deckarep/golang-set"
	"github.com/drunkenfall/drunkenfall/faking"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var ErrPublishDisconnected = errors.New("not connected; will not publish")

// Tournament is the main container of data for this app.
type Tournament struct {
	ID            uint             `json:"tournament_id"`
	Name          string           `json:"name"`
	Slug          string           `json:"slug"`
	Players       []PlayerSummary  `json:"-"`
	Winners       []Player         `json:"-" sql:"-"`
	Runnerups     []*PlayerSummary `json:"-" sql:"-"`
	Casters       []*Person        `json:"-" sql:"-"`
	Matches       []*Match         `json:"-"`
	Current       CurrentMatch     `json:"-"`
	Opened        time.Time        `json:"opened"`
	Scheduled     time.Time        `json:"scheduled"`
	Started       time.Time        `json:"started"`
	QualifyingEnd time.Time        `json:"qualifying_end"`
	Ended         time.Time        `json:"ended"`
	Events        []*Event         `json:"-" sql:"-"`
	Color         string           `json:"color"`
	// Levels      Levels       `json:"levels"`
	Cover       string `json:"cover"`
	Length      int    `json:"length"`
	FinalLength int    `json:"final_length"`
	connected   bool
	db          *Database
	server      *Server
}

// CurrentMatch holds the pointers needed to find the current match
type CurrentMatch int

const minPlayers = 12
const matchLength = 10
const finalLength = 20

// NewTournament returns a completely new Tournament
func NewTournament(name, id, cover string, scheduledStart time.Time, c *gin.Context, server *Server) (*Tournament, error) {
	t := Tournament{
		Name:        name,
		Slug:        id,
		Opened:      time.Now(),
		Scheduled:   scheduledStart,
		Cover:       cover,
		Length:      matchLength,
		FinalLength: finalLength,
		db:          server.DB,
		server:      server,
	}

	// t.SetMatchPointers()
	t.LogEvent(
		"new_tournament", "{name} ({id}) created",
		"name", name,
		"id", id,
		"person", PersonFromSession(t.server, c))

	err := t.db.NewTournament(&t)
	return &t, err
}

// Semi returns one of the two semi matches
func (t *Tournament) Semi(index int) *Match {
	return t.Matches[len(t.Matches)-3+index]
}

// Final returns the final match
func (t *Tournament) Final() *Match {
	return t.Matches[len(t.Matches)-1]
}

// Persist tells the database to save this tournament to disk
func (t *Tournament) Persist() error {
	if t.db == nil {
		// This might happen in tests.
		return errors.New("no database instantiated")
	}

	return t.db.SaveTournament(t)
}

// PublishNext sends information about the next match to the game
//
// It only does this if the match already has four players. If it does
// not, it's a semi that needs backfilling, and then the backfilling
// will make the publish. This should always be called before the
// match is started, so t.NextMatch() can always safely be used.
func (t *Tournament) PublishNext() error {
	if !t.connected {
		t.server.log.Info("Not publishing disconnected tournament")
		return ErrPublishDisconnected
	}

	next, err := t.NextMatch()
	if err != nil {
		return err
	}

	if len(next.Players) != 4 {
		return ErrPublishIncompleteMatch
	}

	msg := GameMatchMessage{
		Tournament: t.Slug,
		Level:      next.realLevel(),
		Length:     next.Length,
		Ruleset:    next.Ruleset,
		Kind:       next.Kind,
	}

	for _, p := range next.Players {
		gp := GamePlayer{
			TopName:    p.DisplayNames[0],
			BottomName: p.DisplayNames[1],
			Color:      p.NumericColor(),
			ArcherType: p.Person.ArcherType,
		}
		msg.Players = append(msg.Players, gp)
	}

	t.server.log.Info("Sending publish", zap.Any("match", msg))
	return t.server.publisher.Publish(gMatch, msg)
}

// connect sets the connect variable
func (t *Tournament) connect(connected bool) {
	if t.connected == connected {
		return
	}

	t.server.log.Info(
		"Tournament connection changed",
		zap.Bool("connected", connected),
	)
	t.connected = connected
}

// JSON returns a JSON representation of the Tournament
func (t *Tournament) JSON() (out []byte, err error) {
	return json.Marshal(t)
}

// URL returns the URL for the tournament
func (t *Tournament) URL() string {
	out := fmt.Sprintf("/%s/", t.Slug)
	return out
}

// LogEvent makes an event and stores it on the tournament object
func (t *Tournament) LogEvent(kind, message string, items ...interface{}) {
	ev, err := NewEvent(kind, message, items...)
	if err != nil {
		log.Fatal(err)
	}

	t.Events = append(t.Events, ev)
}

// AddPlayer adds a player into the tournament
func (t *Tournament) AddPlayer(p *PlayerSummary) error {
	p.Person.Correct()

	if err := t.CanJoin(p.Person); err != nil {
		log.Print(err)
		return err
	}

	t.Players = append(t.Players, *p)

	// If the tournament is already started, just add the player into the
	// runnerups so that they will be placed at the end immediately.
	if !t.Started.IsZero() {
		t.Runnerups = append(t.Runnerups, p)
	}

	t.LogEvent(
		"player_join", "{nick} has joined",
		"nick", p.Person.Nick,
		"person", p.Person)

	return t.db.AddPlayer(t, p)
}

func (t *Tournament) removePlayer(p Player) error {
	for i := 0; i < len(t.Players); i++ {
		r := t.Players[i]
		if r.Person.PersonID == p.Person.PersonID {
			t.Players = append(t.Players[:i], t.Players[i+1:]...)
			break
		}
	}

	t.LogEvent(
		"player_remove", "{nick} has left",
		"nick", p.Person.Nick,
		"person", p.Person)

	err := t.Persist()
	return err
}

// TogglePlayer toggles a player in a tournament
func (t *Tournament) TogglePlayer(id string) error {
	// FIXME(thiderman): Adapt this so it works again
	ps, _ := t.db.GetPerson(id)
	// p, err := t.getTournamentPlayerObject(ps)

	// if err != nil {
	// If there is an error, the player is not in the tournament and we should add them
	p := NewPlayerSummary(ps)
	err := t.AddPlayer(p)
	// return err
	// }

	// If there was no error, the player is in the tournament and we should remove them!
	// err = t.removePlayer(*p)
	return err
}

// SetCasters resets the casters to the input people
func (t *Tournament) SetCasters(ids []string) {
	t.Casters = make([]*Person, 0)
	for _, id := range ids {
		ps, _ := t.db.GetPerson(id)
		t.Casters = append(t.Casters, ps)
	}

	t.Persist()
}

// ShufflePlayers will position players into matches
func (t *Tournament) ShufflePlayers() error {
	// Shuffle all the players
	slice := t.Players
	for i := range slice {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}

	// Loop the players and set them into the matches. This exhausts the
	// list before it leaves the playoffs.
	for i, p := range slice {
		m := t.Matches[i/4]
		pla := *NewPlayer(p.Person)
		err := m.AddPlayer(pla)
		if err != nil {
			return err
		}
	}

	return nil
}

// StartTournament will generate the tournament.
func (t *Tournament) StartTournament(c *gin.Context) error {
	ps := len(t.Players)
	if ps < minPlayers {
		return fmt.Errorf("tournament needs %d or more players, got %d", minPlayers, ps)
	}

	if t.IsRunning() {
		return errors.New("tournament is already running")
	}

	// Set the two first matches
	m1 := NewMatch(t, qualifying)
	t.Matches = append(t.Matches, m1)

	m2 := NewMatch(t, qualifying)
	t.Matches = append(t.Matches, m2)

	// Add the first 8 players, in modulated joining order (first to
	// first match, second to second, third to first etc)
	for i, p := range t.Players[0:8] {
		ps := p.Player()
		err := t.Matches[i%2].AddPlayer(ps)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	t.Started = time.Now()

	// Get the first match and set the scheduled date to be now.
	t.Matches[0].SetTime(c, 0)
	t.LogEvent(
		"start", "Tournament started",
		"person", PersonFromSession(t.server, c))

	err := t.PublishNext()
	if err != nil && err != ErrPublishDisconnected {
		return err
	}
	return t.Persist()
}

// EndQualifyingRounds marks when the qualifying rounds end
func (t *Tournament) EndQualifyingRounds(ts time.Time) error {
	t.QualifyingEnd = ts
	return t.Persist()
}

// Reshuffle shuffles the players of an already started tournament
func (t *Tournament) Reshuffle(c *gin.Context) error {
	if !t.IsRunning() || t.Matches[0].IsStarted() {
		return errors.New("cannot reshuffle")
	}

	// First we need to clear the player slots in the matches.
	for x := 0; x < len(t.Matches)-3; x++ {
		t.Matches[x].Players = nil
		t.Matches[x].presentColors = mapset.NewSet()
	}

	err := t.ShufflePlayers()
	if err != nil {
		return err
	}

	t.LogEvent(
		"reshuffle", "Players reshuffled",
		"person", PersonFromSession(t.server, c))

	return t.Persist()
}

// UsurpTournament adds a batch of eight random players
func (t *Tournament) UsurpTournament() error {
	t.server.DisableWebsocketUpdates()

	// t.db.LoadPeople()
	// rand.Seed(time.Now().UnixNano())
	// for x := 0; x < 8; x++ {
	// 	p := NewPlayer(t.db.People[rand.Intn(len(t.db.People))])
	// 	err := t.AddPlayer(p)
	// 	if err != nil {
	// 		x--
	// 	}
	// }

	t.server.EnableWebsocketUpdates()
	t.server.SendWebsocketUpdate("tournament", t)
	return nil
}

// AutoplaySection runs through all the matches in the current section
// of matches
//
// E.g. if we're in the playoffs, they will all be finished and we're
// left at semi 1.
func (t *Tournament) AutoplaySection() {
	t.server.DisableWebsocketUpdates()

	if !t.IsRunning() {
		t.StartTournament(nil)
	}

	m := t.Matches[t.Current]
	kind := m.Kind

	for kind == m.Kind {
		m.Autoplay()

		if int(t.Current) == len(t.Matches) {
			// If we just finished the finals, then we should just exit
			break
		}

		m = t.Matches[t.Current]
	}

	if kind == playoff {
		needed := t.BackfillsNeeded()
		if needed > 0 {
			ids := make([]string, needed)
			for x := 0; x < needed; x++ {
				ids[x] = t.Runnerups[x].PersonID
			}
			t.BackfillSemis(nil, ids)
		}
	}

	t.server.EnableWebsocketUpdates()
	t.server.SendWebsocketUpdate("tournament", t)
}

// MatchIndex returns the index of the match
func (t *Tournament) MatchIndex(m *Match) int {
	var x int
	var o *Match

	for x, o = range t.Matches {
		if o == m {
			break
		}
	}

	return x
}

// PopulateRunnerups fills a match with the runnerups with best scores
func (t *Tournament) PopulateRunnerups(m *Match) error {
	_, err := t.GetRunnerupPlayers()
	if err != nil {
		return err
	}

	// for i := 0; len(m.Players) < 4; i++ {
	// p := r[i]
	// m.AddPlayer(p)
	// }
	// return errors.New("population of runnerups not implemented")
	return nil
}

// GetRunnerupPlayers gets the runnerups for this tournament
//
// The returned list is sorted descending by score.
// XXX(thiderman): Should be removed in favor of GetRunnerups()
func (t *Tournament) GetRunnerupPlayers() ([]PlayerSummary, error) {
	var ret []PlayerSummary

	// err := t.UpdatePlayers()
	// if err != nil {
	// 	return ret, err
	// }

	rs := len(t.Runnerups)
	p := make([]PlayerSummary, 0, rs)
	for _, r := range t.Runnerups {
		tp, err := t.GetPlayerSummary(r.Person)
		if err != nil {
			return ret, err
		}
		p = append(p, *tp)
	}
	ret = SortByRunnerup(p)
	return ret, nil
}

// GetRunnerups gets the runnerups for this tournament
//
// The returned list is sorted descending by matches and ascending by
// score.
func (t *Tournament) GetRunnerups() ([]*PlayerSummary, error) {
	return t.db.GetRunnerups(t)
}

// UpdatePlayers updates all the player objects with their scores from
// all the matches they have participated in.
func (t *Tournament) UpdatePlayers() error {
	// Make sure all players have their score reset to nothing
	for i := range t.Players {
		t.Players[i].Reset()
	}

	for _, m := range t.Matches {
		for _, p := range m.Players {
			tp, err := t.GetPlayerSummary(p.Person)
			if err != nil {
				return err
			}
			tp.Update(p.Summary())
		}
	}

	return nil
}

// MovePlayers moves the winner(s) of a Match into the next bracket of matches
// or into the Runnerup bracket.
func (t *Tournament) MovePlayers(m *Match) error {
	if m.Kind == qualifying {
		// If we have not yet set the qualifying end, we will keep making
		// matches and we have not passed it
		if t.QualifyingEnd.IsZero() || time.Now().Before(t.QualifyingEnd) {
			log.Print("Scheduling the next match")

			nm := NewMatch(t, qualifying)
			t.Matches = append(t.Matches, nm)
			rups, err := t.GetRunnerups()
			if err != nil {
				return err
			}

			// The runnerups are in order for the next match - add them
			for _, p := range rups {
				nm.AddPlayer(p.Player())
			}

			return nil
		}

		done, err := t.db.QualifyingMatchesDone(t)
		if err != nil {
			return errors.WithStack(err)
		}
		if done {
			t.ScheduleEndgame()
		}
	}

	// For the playoffs, just place the winner into the final
	if m.Kind == playoff {
		p := SortByKills(m.Players)[0]
		t.Matches[len(t.Matches)-1].AddPlayer(p)
	}

	return nil
}

// UpdateRunnerups updates the runnerup list
func (t *Tournament) UpdateRunnerups() error {
	// Get the runnerups and sort them into the Runnerup array
	ps, err := t.GetRunnerupPlayers()
	if err != nil {
		return err
	}
	t.Runnerups = make([]*PlayerSummary, 0)
	for _, p := range ps {
		tp, err := t.GetPlayerSummary(p.Person)
		if err != nil {
			return err
		}
		t.Runnerups = append(t.Runnerups, tp)
	}
	t.Persist()

	return nil
}

// ScheduleEndgame makes the playoff matches and the final
func (t *Tournament) ScheduleEndgame() error {
	log.Print("Scheduling endgame")

	// Get sorted list of players that made it to the playoffs
	players, err := t.db.GetPlayoffPlayers(t)
	if err != nil {
		return errors.WithStack(err)
	}

	lp := len(players)
	if lp != 16 {
		return errors.New(fmt.Sprintf("Needed 16 players, got %d", lp))
	}

	// Bucket the players for inclusion in the playoffs
	buckets, err := DividePlayoffPlayers(players)
	if err != nil {
		return errors.WithStack(err)
	}

	// Add four playoff matches
	for x := 0; x < 4; x++ {
		m := NewMatch(t, playoff)

		// Add players to the match
		for _, p := range buckets[x] {
			m.AddPlayer(p.Player())
		}

		t.Matches = append(t.Matches, m)
	}

	// Add the final, without players for now
	m := NewMatch(t, final)
	t.Matches = append(t.Matches, m)

	return nil
}

// BackfillSemis takes a few Person IDs and shuffles those into the remaining slots
// of the semi matches
func (t *Tournament) BackfillSemis(c *gin.Context, ids []string) error {
	// If we're on the last playoff, we should backfill the semis with runnerups
	// until they have have full seats.
	// The amount of players needed; 8 minus the current amount
	offset := len(t.Matches) - 3
	semiPlayers := t.BackfillsNeeded()
	if len(ids) != semiPlayers {
		return fmt.Errorf("Need %d players, got %d", semiPlayers, len(ids))
	}

	publish := true
	if semiPlayers == 1 {
		// If we only have one player to add, that means that the previous
		// match already has four players and therefore the update message
		// has already been sent
		publish = false
	}

	added := make([]*Person, semiPlayers)
	for x, id := range ids {
		index := 0
		if len(t.Matches[offset].Players) == 4 {
			index = 1
		}

		ps, err := t.db.GetPerson(id)
		if err != nil {
			return errors.WithStack(err)
		}

		// p, err := t.getTournamentPlayerObject(ps)
		// if err != nil {
		// 	return errors.WithStack(err)
		// }

		t.Matches[offset+index].AddPlayer(*NewPlayer(ps))
		added[x] = ps
		err = t.removeFromRunnerups(ps)
		if err != nil {
			log.Fatal(err)
		}

	}

	// If we haven't already sent a message about the next match to the
	// game, it is high time to do so!
	if publish {
		err := t.PublishNext()
		if err != nil && err != ErrPublishDisconnected {
			t.server.log.Info("Publishing next match failed", zap.Error(err))
		}
	}

	t.LogEvent(
		"backfill_semi", "Backfilling {count} semi players",
		"count", semiPlayers,
		"players", added,
		"person", PersonFromSession(t.server, c))

	return t.Persist()
}

// BackfillsNeeded returns the number of players needed to be backfilled
func (t *Tournament) BackfillsNeeded() int {
	offset := len(t.Matches) - 3
	semiPlayers := 8 - (len(t.Matches[offset].Players) + len(t.Matches[offset+1].Players))
	return semiPlayers
}

// NextMatch returns the next match
func (t *Tournament) NextMatch() (*Match, error) {
	if !t.IsRunning() {
		return nil, errors.New("tournament not running")
	}

	return t.db.NextMatch(t)
}

// AwardMedals places the winning players in the Winners position
func (t *Tournament) AwardMedals(c *gin.Context, m *Match) error {
	if m.Kind != final {
		return errors.New("awarding medals outside of the final")
	}

	ps := SortByKills(m.Players)
	t.Winners = ps[0:3]

	t.LogEvent(
		"tournament_end",
		"Tournament finished",
		"person", PersonFromSession(t.server, c))

	t.Ended = time.Now()
	t.Persist()

	return nil
}

// IsOpen returns boolean true if the tournament is open for registration
func (t *Tournament) IsOpen() bool {
	return !t.Opened.IsZero()
}

// IsJoinable returns boolean true if the tournament is joinable
func (t *Tournament) IsJoinable() bool {
	return t.IsOpen() && t.Started.IsZero()
}

// IsStartable returns boolean true if the tournament can be started
func (t *Tournament) IsStartable() bool {
	p := len(t.Players)
	return t.IsOpen() && t.Started.IsZero() && p >= minPlayers
}

// IsRunning returns boolean true if the tournament is running or not
func (t *Tournament) IsRunning() bool {
	return !t.Started.IsZero() && t.Ended.IsZero()
}

// CanJoin checks if a player is allowed to join or is already in the tournament
func (t *Tournament) CanJoin(ps *Person) error {
	for _, p := range t.Players {
		if p.Person.Nick == ps.Nick {
			return errors.New("already in tournament")
		}
	}
	return nil
}

// SetMatchPointers loops over all matches in the tournament and sets the tournament reference
//
// When loading tournaments from the database, these references will not be set.
// This also sets *Match pointers for Player objects.
func (t *Tournament) SetMatchPointers() error {
	var m *Match

	for i := range t.Matches {
		m = t.Matches[i]
		m.presentColors = mapset.NewSet()
		m.Tournament = t
		for j := range m.Players {
			m.Players[j].Match = m
		}
	}

	return nil
}

// GetCredits returns the credits object needed to display the credits roll.
func (t *Tournament) GetCredits() (*Credits, error) {
	// if t.Ended.IsZero() {
	// 	return nil, errors.New("cannot roll credits for unfinished tournament")
	// }

	// // TODO(thiderman): Many of these values are currently hardcoded or
	// // broadly grabs everything.
	// // We should move towards specifying these live when setting
	// // up the tournament itself.

	// executive := t.db.GetSafePerson("1279099058796903") // thiderman
	// producers := []*Person{
	// 	t.db.GetSafePerson("10153943465786915"), // GoosE
	// 	t.db.GetSafePerson("10154542569541289"), // Queen Obscene
	// 	t.db.GetSafePerson("10153964695568099"), // Karl-Astrid
	// 	t.db.GetSafePerson("10153910124391516"), // Hest
	// 	t.db.GetSafePerson("10154040229117471"), // Skolpadda
	// 	t.db.GetSafePerson("10154011729888111"), // Moijra
	// 	t.db.GetSafePerson("10154296655435218"), // Dalan
	// }

	// players := []*Person{
	// 	t.Winners[0].Person,
	// 	t.Winners[1].Person,
	// 	t.Winners[2].Person,
	// }
	// players = append(players, t.Runnerups...)

	// c := &Credits{
	// 	Executive:     executive,
	// 	Producers:     producers,
	// 	Players:       players,
	// 	ArchersHarmed: t.ArchersHarmed(),
	// }

	return &Credits{}, nil
}

// ArchersHarmed returns the amount of killed archers during the tournament
func (t *Tournament) ArchersHarmed() int {
	ret := 0
	for _, m := range t.Matches {
		ret += m.ArchersHarmed()
	}

	return ret
}

// SetupFakeTournament creates a fake tournament
func SetupFakeTournament(c *gin.Context, s *Server, req *NewRequest) *Tournament {
	title, id := req.Name, req.ID

	t, err := NewTournament(title, id, "", time.Now().Add(time.Hour), c, s)
	if err != nil {
		// TODO this is the least we can do
		log.Printf("error creating tournament: %v", err)
	}

	// Fake between 14 and max_players players
	for i := 0; i < rand.Intn(18)+14; i++ {
		ps := &Person{
			PersonID:       faking.FakeName(),
			Name:           faking.FakeName(),
			Nick:           faking.FakeNick(),
			AvatarURL:      faking.FakeAvatar(),
			PreferredColor: RandomColor(Colors),
		}
		p := NewPlayer(ps)
		sum := p.Summary()
		t.AddPlayer(&sum)
	}

	t.Persist()
	return t
}

// GetPlayerSummary returns the tournament-wide player object.
func (t *Tournament) GetPlayerSummary(ps *Person) (*PlayerSummary, error) {
	return t.db.GetPlayerSummary(t, ps.PersonID)
}

func (t *Tournament) removeFromRunnerups(p *Person) error {
	for j := 0; j < len(t.Runnerups); j++ {
		r := t.Runnerups[j]
		if r.PersonID == p.PersonID {
			t.Runnerups = append(t.Runnerups[:j], t.Runnerups[j+1:]...)
			return nil
		}
	}
	return nil
}
