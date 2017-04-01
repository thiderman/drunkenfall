package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/deckarep/golang-set"
	"github.com/drunkenfall/faking"
	"log"
	"math/rand"
	"time"
)

// Tournament is the main container of data for this app.
type Tournament struct {
	Name        string       `json:"name"`
	ID          string       `json:"id"`
	Players     []Player     `json:"players"` // See getTournamentPlayerObject()
	Winners     []Player     `json:"winners"` // TODO(thiderman): Refactor to pointer
	Runnerups   []*Person    `json:"runnerups"`
	Judges      []Judge      `json:"judges"`
	Tryouts     []*Match     `json:"tryouts"`
	Semis       []*Match     `json:"semis"`
	Final       *Match       `json:"final"`
	Current     CurrentMatch `json:"current"`
	Previous    CurrentMatch `json:"previous"`
	Opened      time.Time    `json:"opened"`
	Scheduled   time.Time    `json:"scheduled"`
	Started     time.Time    `json:"started"`
	Ended       time.Time    `json:"ended"`
	db          *Database
	server      *Server
	length      int
	finalLength int
}

// CurrentMatch holds the pointers needed to find the current match
type CurrentMatch struct {
	Kind  string `json:"kind"`
	Index int    `json:"index"`
}

const minPlayers = 8
const maxPlayers = 32
const matchLength = 10
const finalLength = 20

// NewTournament returns a completely new Tournament
func NewTournament(name, id string, scheduledStart time.Time, server *Server) (*Tournament, error) {
	t := Tournament{
		Name:        name,
		ID:          id,
		Opened:      time.Now(),
		Scheduled:   scheduledStart,
		db:          server.DB,
		server:      server,
		length:      matchLength,
		finalLength: finalLength,
	}

	// Set tryouts
	for i := 0; i < 4; i++ {
		match := NewMatch(&t, i, tryout)
		t.Tryouts = append(t.Tryouts, match)
	}

	// Set the predefined matches
	t.Semis = []*Match{NewMatch(&t, 0, semi), NewMatch(&t, 1, semi)}
	t.Final = NewMatch(&t, 0, final)

	// Mark the current match
	t.Current = CurrentMatch{tryout, 0}

	t.SetMatchPointers()
	t.Persist()
	return &t, nil
}

// LoadTournament loads a tournament from persisted JSON data
func LoadTournament(data []byte, db *Database) (t *Tournament, e error) {
	t = &Tournament{}
	err := json.Unmarshal(data, t)
	if err != nil {
		log.Print(err)
		return t, err
	}

	t.db = db
	t.server = db.Server

	t.SetMatchPointers()
	return
}

// Persist tells the database to save this tournament to disk
func (t *Tournament) Persist() error {
	if t.db == nil {
		// This might happen in tests.
		return errors.New("no database instantiated")
	}

	go t.server.SendWebsocketUpdate()

	return t.db.SaveTournament(t)
}

// JSON returns a JSON representation of the Tournament
func (t *Tournament) JSON() (out []byte, err error) {
	out, err = json.Marshal(t)
	return
}

// URL returns the URL for the tournament
func (t *Tournament) URL() string {
	out := fmt.Sprintf("/%s/", t.ID)
	return out
}

// AddPlayer adds a player into the tournament
//
// If the four default tryout matches are full, four more will be generated.
func (t *Tournament) AddPlayer(p *Player) error {
	p.Person.Correct()

	if err := t.CanJoin(p.Person); err != nil {
		log.Print(err)
		return err
	}

	t.Players = append(t.Players, *p)

	// If the tournament is already started, just add the player into the
	// runnerups so that they will be placed at the end immediately.
	if !t.Started.IsZero() {
		t.Runnerups = append(t.Runnerups, p.Person)
	}

	err := t.Persist()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func (t *Tournament) removePlayer(p Player) error {
	for i := 0; i < len(t.Players); i++ {
		r := t.Players[i]
		if r == p {
			t.Players = append(t.Players[:i], t.Players[i+1:]...)
			break
		}
	}

	err := t.Persist()
	return err
}

// TogglePlayer toggles a player in a tournament
func (t *Tournament) TogglePlayer(id string) error {
	ps := t.db.GetPerson(id)
	p, err := t.getTournamentPlayerObject(ps)

	if err != nil {
		// If there is an error, the player is not in the tournament and we should add them
		p = NewPlayer(ps)
		err = t.AddPlayer(p)
		return err
	}

	// If there was no error, the player is in the tournament and we should remove them!
	err = t.removePlayer(*p)
	return err
}

// ShufflePlayers will position players into matches
func (t *Tournament) ShufflePlayers() {
	// Shuffle all the players
	slice := t.Players
	for i := range slice {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}

	// Loop the players and set them into the matches
	for i, p := range slice {
		m := t.Tryouts[i/4]
		m.AddPlayer(p)
	}
}

// StartTournament will generate the tournament.
//
// This includes:
//  Generating Tryout matches
//  Setting Started date
//
// It will fail if there are not between 16 and max_players players.
func (t *Tournament) StartTournament() error {
	log.Printf("Starting %s...", t.Name)
	ps := len(t.Players)
	if ps < minPlayers || ps > maxPlayers {
		return fmt.Errorf("Tournament needs %d or more players and %d or less, got %d", minPlayers, maxPlayers, ps)
	}

	// More than 16 players - add four more tryouts
	if ps > 16 {
		for i := 16; i < ps; i += 4 {
			match := NewMatch(t, i/4, tryout)
			t.Tryouts = append(t.Tryouts, match)
		}
	}

	t.ShufflePlayers()
	t.Started = time.Now()

	// Get the first match and set the scheduled date to be now.
	m, err := t.NextMatch()
	if err != nil {
		log.Fatal(err)
	}
	m.SetTime(0)

	t.Persist()
	return nil
}

// Reshuffle shuffles the players of an already started tournament
func (t *Tournament) Reshuffle() error {
	// First we need to clear the player slots in the matches.
	for x := range t.Tryouts {
		t.Tryouts[x].Players = nil
		t.Tryouts[x].presentColors = mapset.NewSet()
	}

	t.ShufflePlayers()
	t.Persist()

	return nil
}

// UsurpTournament starts a fake tournament with all registered players
func (t *Tournament) UsurpTournament() error {
	t.db.LoadPeople()
	for _, ps := range t.db.People {
		p := NewPlayer(ps)
		t.AddPlayer(p)
	}
	return nil
}

// PopulateRunnerups fills a match with the runnerups with best scores
func (t *Tournament) PopulateRunnerups(m *Match) error {
	r, err := t.GetRunnerupPlayers()
	if err != nil {
		return err
	}

	for i := 0; len(m.Players) < 4; i++ {
		p := r[i]
		m.AddPlayer(p)
	}
	return nil
}

// GetRunnerupPlayers gets the runnerups for this tournament
//
// The returned list is sorted descending by score.
func (t *Tournament) GetRunnerupPlayers() (ps []Player, err error) {
	var l *Player
	err = t.UpdatePlayers()
	if err != nil {
		return
	}

	rs := len(t.Runnerups)
	p := make([]Player, 0, rs)
	for _, r := range t.Runnerups {
		l, err = t.getTournamentPlayerObject(r)
		if err != nil {
			return
		}

		p = append(p, *l)
	}
	bs := SortByRunnerup(p)
	return bs, nil
}

// UpdatePlayers updates all the player objects with their scores from
// all the matches they have participated in.
func (t *Tournament) UpdatePlayers() error {
	// Make sure all players have their score reset to nothing

	for i := range t.Players {
		t.Players[i].Reset()
	}

	for _, m := range t.Tryouts {
		for _, p := range m.Players {
			tp, err := t.getTournamentPlayerObject(p.Person)
			if err != nil {
				return err
			}
			tp.Update(p)
		}
	}

	for _, m := range t.Semis {
		for _, p := range m.Players {
			tp, err := t.getTournamentPlayerObject(p.Person)
			if err != nil {
				return err
			}
			tp.Update(p)
		}
	}

	for _, p := range t.Final.Players {
		tp, err := t.getTournamentPlayerObject(p.Person)
		if err != nil {
			return err
		}
		tp.Update(p)
	}

	return nil
}

// MovePlayers moves the winner(s) of a Match into the next bracket of matches
// or into the Runnerup bracket.
func (t *Tournament) MovePlayers(m *Match) error {
	if m.Kind == tryout {
		err := t.moveTryoutPlayers(m)
		if err != nil {
			return err
		}

		// If the next match is also a tryout and does not have enough players,
		// fill it up with runnerups.
		nm, err := t.NextMatch()
		if err != nil {
			return err
		}
		if nm.Kind == tryout && len(nm.Players) < 4 {
			log.Printf("Setting runnerups for %s", nm)
			err := t.PopulateRunnerups(nm)
			if err != nil {
				return err
			}
		}
	}

	// For the semis, just place the winner and silver into the final
	if m.Kind == semi {
		for i, p := range SortByKills(m.Players) {
			if i < 2 {
				t.Final.AddPlayer(p)
			}
		}
	}

	return nil
}

func (t *Tournament) moveTryoutPlayers(m *Match) error {
	ps := SortByKills(m.Players)
	for i := 0; i < len(ps); i++ {
		p := ps[i]
		// If we are in a four-match tryout, both the winner and the second-place
		// are to be sent to the semis.
		// If there are more than four matches, just send the winner
		if len(t.Tryouts) == 4 && i < 2 || i == 0 {
			// This spreads the winners into the semis so that the winners do not
			// face off immediately in the semis
			index := (i + m.Index) % 2
			t.Semis[index].AddPlayer(p)

			// If the player is also inside of the runnerups, move them from the
			// runnerup roster since they now have advanced to the finals. This
			// only happens for players that win the runnerup rounds.
			t.removeFromRunnerups(p.Person)
		} else {
			// For everyone else, add them into the Runnerup bracket unless they are
			// already in there.
			found := false
			for j := 0; j < len(t.Runnerups); j++ {
				r := t.Runnerups[j]
				if r.ID == p.Person.ID {
					found = true
					break
				}
			}
			if !found {
				t.Runnerups = append(t.Runnerups, p.Person)
			}
		}
	}

	return t.UpdateRunnerups()
}

// UpdateRunnerups updates the runnerup list
func (t *Tournament) UpdateRunnerups() error {
	// Get the runnerups and sort them into the Runnerup array
	ps, err := t.GetRunnerupPlayers()
	if err != nil {
		return err
	}
	t.Runnerups = make([]*Person, 0)
	for _, p := range ps {
		t.Runnerups = append(t.Runnerups, p.Person)
	}

	return nil
}

// BackfillSemis takes a few Person IDs and shuffles those into the remaining slots
// of the semi matches
func (t *Tournament) BackfillSemis(ids []string) error {
	// If we're on the last tryout, we should backfill the semis with runnerups
	// until they have have full seats.
	// The amount of players needed; 8 minus the current amount
	semiPlayers := 8 - (len(t.Semis[0].Players) + len(t.Semis[1].Players))
	if len(ids) != semiPlayers {
		return fmt.Errorf("Need %d players, got %d", semiPlayers, len(ids))
	}

	log.Printf("Backfilling %d semi players\n", semiPlayers)
	for _, id := range ids {
		index := 0
		if len(t.Semis[0].Players) == 4 {
			index = 1
		}

		ps := t.db.GetPerson(id)
		p, err := t.getTournamentPlayerObject(ps)
		if err != nil {
			return err
		}

		t.Semis[index].AddPlayer(*p)
		t.removeFromRunnerups(ps)
	}

	t.Persist()
	return nil
}

// SetCurrent sets the current match of the tournament
func (t *Tournament) SetCurrent(m *Match) {
	t.Current = CurrentMatch{m.Kind, m.Index}
}

// SetPrevious sets the previous match of the tournament
func (t *Tournament) SetPrevious(m *Match) {
	t.Previous = CurrentMatch{m.Kind, m.Index}
}

// NextMatch returns the next match
func (t *Tournament) NextMatch() (m *Match, err error) {
	// Firstly, check the tryouts
	for x := range t.Tryouts {
		m = t.Tryouts[x]
		if !m.IsEnded() {
			t.SetCurrent(m)
			return
		}
	}

	// If we don't have any tryouts, or there are no tryouts left,
	// check the semis
	for x := range t.Semis {
		m = t.Semis[x]
		if !m.IsEnded() {
			t.SetCurrent(m)
			return
		}
	}

	if !t.Final.IsEnded() {
		t.SetCurrent(t.Final)
		return t.Final, nil
	}

	return m, errors.New("all matches have been played")
}

// AwardMedals places the winning players in the Winners position
func (t *Tournament) AwardMedals(m *Match) error {
	if m.Kind != final {
		return errors.New("awarding medals outside of the final")
	}

	ps := SortByKills(m.Players)
	t.Winners = ps[0:3]

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
	if len(t.Players) >= maxPlayers {
		return false
	}
	return t.IsOpen() && t.Started.IsZero()
}

// IsStartable returns boolean true if the tournament can be started
func (t *Tournament) IsStartable() bool {
	p := len(t.Players)
	return t.IsOpen() && t.Started.IsZero() && p >= 16 && p <= maxPlayers
}

// IsRunning returns boolean true if the tournament is running or not
func (t *Tournament) IsRunning() bool {
	return !t.Started.IsZero() && t.Ended.IsZero()
}

// CanJoin checks if a player is allowed to join or is already in the tournament
func (t *Tournament) CanJoin(ps *Person) error {
	if len(t.Players) >= maxPlayers {
		return errors.New("tournament is full")
	}
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
	// log.Printf("%s: Setting match pointers...", t.ID)

	for i := range t.Tryouts {
		m = t.Tryouts[i]
		m.presentColors = mapset.NewSet()
		m.Tournament = t
		for j := range m.Players {
			m.Players[j].Match = m
		}
	}

	for i := range t.Semis {
		m = t.Semis[i]
		m.presentColors = mapset.NewSet()
		m.Tournament = t
		for j := range m.Players {
			m.Players[j].Match = m
		}
	}
	t.Final.Tournament = t
	t.Final.presentColors = mapset.NewSet()
	for i := range t.Final.Players {
		t.Final.Players[i].Match = t.Final
	}

	// log.Printf("%s: Pointers loaded.", t.ID)
	return nil
}

// SetupFakeTournament creates a fake tournament
func SetupFakeTournament(s *Server) *Tournament {
	title, id := faking.FakeTournamentTitle()

	t, err := NewTournament(title, id, time.Now().Add(time.Hour), s)
	if err != nil {
		// TODO this is the least we can do
		log.Printf("error creating tournament: %v", err)
	}

	// Fake between 14 and max_players players
	for i := 0; i < rand.Intn(18)+14; i++ {
		ps := &Person{
			ID:              faking.FakeName(),
			Name:            faking.FakeName(),
			Nick:            faking.FakeNick(),
			AvatarURL:       faking.FakeAvatar(),
			ColorPreference: []string{RandomColor(Colors)},
		}
		p := NewPlayer(ps)
		t.AddPlayer(p)
	}

	return t
}

// getTournamentPlayerObject returns the tournament-wide player object.
//
// The need for this distinction is that the ones that are stored in t.Players
// have scores from all the matches they have participated in, whereas the
// ones started in m.Players are local to that match only. This is also why
// the Match objects don't have pointers to their Player objects.
func (t *Tournament) getTournamentPlayerObject(ps *Person) (p *Player, err error) {
	for i := range t.Players {
		p := &t.Players[i]
		if ps.ID == p.Person.ID {
			return p, nil
		}
	}

	err = fmt.Errorf("no player found for %s", ps)
	return
}

func (t *Tournament) removeFromRunnerups(p *Person) {
	for j := 0; j < len(t.Runnerups); j++ {
		r := t.Runnerups[j]
		if r.ID == p.ID {
			t.Runnerups = append(t.Runnerups[:j], t.Runnerups[j+1:]...)
			break
		}
	}
}
