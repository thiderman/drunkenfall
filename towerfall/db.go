package towerfall

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var globalDB *Database

var ErrNoTourmamentRunning = errors.New("no tournament is running")

type Database struct {
	Server       *Server
	DB           *pg.DB
	log          *zap.Logger
	persistcalls int
}

// NewDatabase sets up the database reader and writer
func NewDatabase(c *Config) (*Database, error) {
	zlog, _ := zap.NewDevelopment()

	pgdb := pg.Connect(&pg.Options{
		Addr:     c.DbHost,
		User:     c.DbUser,
		Database: c.DbName,
	})

	db := &Database{
		DB:  pgdb,
		log: zlog,
	}

	if c.DbVerbose {
		zlog.Info("DRUNKENFALL_DBVERBOSE is set; enabling logger")
		pgdb.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
			query, err := event.FormattedQuery()
			if err != nil {
				panic(err)
			}

			log.Print(query)
		})
	}

	globalDB = db

	return db, nil
}

// SaveTournament stores the current state of the tournaments into the db
func (d *Database) SaveTournament(t *Tournament) error {
	d.persistcalls++
	err := d.DB.Update(t)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// SaveTournament stores the current state of the tournaments into the db
func (d *Database) NewTournament(t *Tournament) error {
	return d.DB.Insert(t)
}

// AddPlayer adds a player object to the tournament
func (d *Database) AddPlayer(t *Tournament, ps *PlayerSummary) error {
	ps.TournamentID = t.ID
	return d.DB.Insert(ps)
}

// RemovePlayer removes a player from a tourmament
func (d *Database) RemovePlayer(ps *PlayerSummary) error {
	return d.DB.Delete(ps)
}

// AddPlayerToMatch adds a player object to a match
func (d *Database) AddPlayerToMatch(m *Match, p *Player, idx int) error {
	p.MatchID = m.ID
	p.Index = idx

	// Reset the scores.
	// TODO(thiderman): Replace this
	p.Shots = 0
	p.Sweeps = 0
	p.Kills = 0
	p.Self = 0
	p.MatchScore = 0
	p.TotalScore = 0

	if p.State == nil {
		p.State = NewPlayerState()
	}

	return d.DB.Insert(p)
}

// AddMatch adds a match
func (d *Database) AddMatch(t *Tournament, m *Match) error {
	m.TournamentID = t.ID
	return d.DB.Insert(m)
}

// IsInTournament returns if the player is in the tournament or not
func (d *Database) IsInTournament(t *Tournament, p *Person) (bool, error) {
	q := d.DB.Model(&PlayerSummary{}).Where("tournament_id = ?", t.ID)
	count, err := q.Where("person_id = ?", p.PersonID).Count()
	if err != nil {
		return false, err
	}

	return count == 1, nil
}

// SaveMatch saves a match
func (d *Database) SaveMatch(m *Match) error {
	err := d.DB.Update(m)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// AddCommit adds a commit on a match
func (d *Database) AddCommit(m *Match, c *Commit) error {
	c.MatchID = m.ID
	return d.DB.Insert(c)
}

// StoreMessage stores a message for a match
func (d *Database) StoreMessage(m *Match, msg *Message) error {
	msg.MatchID = m.ID

	// Spin off as a goroutine, and print error if it fails; don't care
	// what the caller thinks. Without this, this operation becomes
	// crazy slow since we do it so often.
	go func() {
		out, err := json.Marshal(msg.Data)
		if err != nil {
			log.Printf("Error when marshalling JSON: %+v", err)
			return
		}

		msg.JSON = string(out)

		err = d.DB.Insert(msg)
		if err != nil {
			log.Printf("Error when saving message: %+v", err)
			return
		}
	}()

	return nil
}

// UpdatePlayer updates one player instance
func (d *Database) UpdatePlayer(m *Match, p *Player) error {
	if p.ID == 0 {
		panic(fmt.Sprintf("player id was zero: %+v", p))
	}

	// Set the computed score on every update
	p.TotalScore = p.Score()
	err := d.DB.Update(p)
	if err != nil {
		return err
	}

	return d.UpdatePlayerSummary(m.Tournament, p)
}

// UpdatePlayerSummary updates the total player data for the tournament
func (d *Database) UpdatePlayerSummary(t *Tournament, p *Player) error {
	query := `UPDATE player_summaries ps
   SET (shots, sweeps, kills, self, matches, total_score, skill_score)
   =
   (SELECT SUM(shots),
           SUM(sweeps),
           SUM(kills),
           SUM(self),
           COUNT(*),
           SUM(total_score),
           (SUM(total_score) / COUNT(*))
      FROM players P
      INNER JOIN matches M ON p.match_id = m.id
      WHERE m.tournament_id = ?
        AND m.started IS NOT NULL
        AND person_id = ?)
    WHERE person_id = ?
      AND tournament_id = ?;`

	_, err := d.DB.Exec(query, t.ID, p.PersonID, p.PersonID, t.ID)
	if err != nil {
		log.Printf("Summary update failed: %+v", err)
	}
	return err
}

// OverwriteTournament takes a new foreign Tournament{} object and replaces
// the one with the same ID with that one.
//
// Used from the EditHandler()
func (d *Database) OverwriteTournament(t *Tournament) error {
	return nil
}

// SavePerson stores a person into the DB
func (d *Database) SavePerson(p *Person) error {
	return d.DB.Update(p)
}

// GetPerson gets a Person{} from the DB
func (d *Database) GetPerson(id string) (*Person, error) {
	p := Person{
		PersonID: id,
	}
	err := d.DB.Select(&p)

	return &p, err
}

// GetRandomPerson gets a random Person{} from the DB
func (d *Database) GetRandomPerson(used []string) (*Person, error) {
	p := Person{}
	q := d.DB.Model(&p).Where("NOT disabled").OrderExpr("random()")

	if len(used) != 0 {
		args := make([]interface{}, 0)
		for _, u := range used {
			args = append(args, u)
		}
		q = q.WhereIn("person_id NOT IN (?)", args...)
	}

	err := q.First()
	return &p, err
}

// GetSafePerson gets a Person{} from the DB, while being absolutely
// sure there will be no error.
//
// This is only for hardcoded cases where error handling is just pointless.
func (d *Database) GetSafePerson(id string) *Person {
	p, _ := d.GetPerson(id)
	return p
}

// DisablePerson disables or re-enables a person
func (d *Database) DisablePerson(id string) error {
	// p, err := d.getPerson(id)
	// if err != nil {
	// 	return err
	// }

	// p.Disabled = !p.Disabled
	// d.SavePerson(p)

	return nil
}

// GetPeople loads the people from the database
func (d *Database) GetPeople() ([]*Person, error) {
	ret := make([]*Person, 0)
	err := d.DB.Model(&ret).Where("NOT disabled").Select()
	return ret, err
}

// GetTournament gets a tournament by id, or returns the cached one if there is one
func (d *Database) GetTournament(id uint) (*Tournament, error) {
	// if d.tournament == nil {
	t := Tournament{
		db:     d,
		server: d.Server,
	}

	q := d.DB.Model(&t).Column("tournament.*", "Matches", "Players")
	err := q.Where("id = ?", id).First()
	if err != nil {
		return nil, err
	}

	return &t, nil

	// 	d.tournament = &t
	// }

	// return d.tournament, nil
}

func (d *Database) GetTournaments() ([]*Tournament, error) {
	ret := make([]*Tournament, 0)
	q := d.DB.Model(&ret).Column("tournament.*", "Matches", "Players")
	err := q.Order("opened DESC").Select()
	return ret, err
}

// GetCurrentTournament gets the currently running tournament
func (d *Database) GetCurrentTournament() (*Tournament, error) {
	t := Tournament{
		db:     d,
		server: d.Server,
	}

	err := d.DB.Model(&t).Where("started IS NOT NULL").Where("ended IS NULL").Order("opened DESC").First()
	return &t, err
}

// GetMatch gets a match
func (d *Database) GetMatch(id uint) (*Match, error) {
	m := Match{}
	q := d.DB.Model(&m).Where("match.id = ?", id)
	err := q.Select()
	if err != nil {
		return nil, err
	}

	ps := []*Player{}
	q = d.DB.Model(&ps).Where("match_id = ?", id).Order("index")
	err = q.Select()
	m.Players = ps

	return &m, err
}

// GetMatches gets a slice of matches based on a kind
func (d *Database) GetMatches(t *Tournament, kind string) ([]*Match, error) {
	ret := []*Match{}

	q := d.DB.Model(&ret)

	if kind != "all" {
		q = q.Where("kind = ?", kind)
	}

	q = q.Where("tournament_id = ?", t.ID).Order("id ASC")
	err := q.Select(&ret)
	if err != nil {
		return ret, err
	}

	// XXX(thiderman): This is terrible, but without it the ordering of the players is wrong again
	for x, m := range ret {
		ps := []*Player{}
		q = d.DB.Model(&ps).Where("match_id = ?", m.ID).Order("index")
		err = q.Select()
		if err != nil {
			return ret, err
		}
		ret[x].Players = ps
	}

	return ret, err
}

// GetFinal get the final of a tournament
func (d *Database) GetFinal(t *Tournament) (*Match, error) {
	ret := Match{
		Tournament: t,
	}

	q := d.DB.Model(&ret).Column("match.*", "Players")
	err := q.Where("tournament_id = ?", t.ID).Where("kind = ?", "final").First()

	return &ret, err
}

// NextMatch the next playable match of a tournament
func (d *Database) NextMatch(t *Tournament) (*Match, error) {
	m := Match{
		Tournament: t,
	}

	q := t.db.DB.Model(&m).Where("tournament_id = ? AND started IS NULL", t.ID)
	q = q.Order("id").Limit(1).Column("match.*")

	err := q.Select()
	if err != nil {
		return nil, err
	}

	ps := []*Player{}
	q = d.DB.Model(&ps).Where("match_id = ?", m.ID).Order("index")
	err = q.Select()
	m.Players = ps

	return &m, err
}

// CurrentMatch the currently running match
func (d *Database) CurrentMatch(t *Tournament) (*Match, error) {
	m := Match{
		Tournament:   t,
		currentRound: NewRound(),
	}

	q := t.db.DB.Model(&m).Where("tournament_id = ? AND ended IS NULL", t.ID)
	q = q.Order("id").Limit(1)

	err := q.Select()
	if err != nil {
		return nil, err
	}

	ps := []*Player{}
	q = t.db.DB.Model(&ps).Where("match_id = ?", m.ID).Order("index ASC")
	err = q.Select()
	m.Players = ps

	return &m, err
}

// MatchMessages returns the messages from a match
func (d *Database) MatchMessages(m *Match) ([]*Message, error) {
	msgs := []*Message{}
	q := d.DB.Model(&msgs).Where("match_id = ?", m.ID)
	err := q.Select()

	return msgs, errors.WithStack(err)
}

// QualifyingMatchesDone returns if we're done with the qualifiers
func (d *Database) QualifyingMatchesDone(t *Tournament) (bool, error) {
	m := Match{
		Tournament: t,
	}

	// Get the count of the matches that have not ended
	q := t.db.DB.Model(&m).Where("tournament_id = ?", t.ID)
	q = q.Where("ended IS NULL")
	q = q.Where("kind = ?", qualifying)

	out, err := q.Count()
	if err != nil {
		return false, err
	}

	return out == 0, err
}

// GetRunnerups gets the next four runnerups, excluding those already
// booked to matches
func (d *Database) GetRunnerups(t *Tournament) ([]*PlayerSummary, error) {
	ps := []*Player{}
	subq := d.DB.Model(&ps).Column("person_id").Join("INNER JOIN matches m on m.id = match_id")
	subq = subq.Where("m.ended IS NULL").Where("m.tournament_id = ?", t.ID)

	ret := []*PlayerSummary{}
	q := d.DB.Model(&ret).Where("tournament_id = ?", t.ID)
	q = q.Where("person_id NOT IN (?)", subq)
	q = q.Order("matches ASC", "skill_score DESC").Limit(4)

	err := q.Select()
	return ret, err
}

// GetAllRunnerups gets all the runnerups that aren't already booked
// into matches
func (d *Database) GetAllRunnerups(t *Tournament) ([]*PlayerSummary, error) {
	ps := []*PlayerSummary{}
	q := `SELECT * FROM player_summaries
                WHERE id IN (SELECT id FROM runnerups(?))
             ORDER BY matches ASC, skill_score DESC`
	_, err := d.DB.Query(&ps, q, t.ID)
	return ps, err
}

// GetWinner gets the winner of the match
func (d *Database) GetWinner(m *Match) (*Player, error) {
	p := Player{}
	q := d.DB.Model(&p).Where("match_id = ?", m.ID).Order("kills DESC")
	q = q.Column("player.*", "Person").Limit(1)
	err := q.Select()
	if err != nil {
		return nil, err
	}

	if &p == nil {
		return nil, errors.New("player was nil")
	}

	return &p, nil
}

// GetSilver gets the winner of the match
func (d *Database) GetSilver(m *Match) (*Player, error) {
	p := Player{}
	q := d.DB.Model(&p).Where("match_id = ?", m.ID).Order("kills DESC")
	q = q.Column("player.*", "Person").Offset(1).Limit(1)
	err := q.Select()
	if err != nil {
		return nil, err
	}

	if &p == nil {
		return nil, errors.New("player was nil")
	}

	return &p, nil
}

// GetPlayoffPlayers gets the sixteen players that made it to the playoffs
func (d *Database) GetPlayoffPlayers(t *Tournament) ([]*PlayerSummary, error) {
	limit := 16
	if len(t.Players) < limit {
		log.Print("Backing down to two-match playoffs")
		limit = 8
	}

	ret := []*PlayerSummary{}
	q := d.DB.Model(&ret).Where("tournament_id = ?", t.ID)
	q = q.Order("skill_score DESC").Limit(limit)

	err := q.Select()
	return ret, err
}

// GetPlayerSummary gets a single player summary for a tourmanent
func (d *Database) GetPlayerSummary(t *Tournament, pid string) (*PlayerSummary, error) {
	ret := PlayerSummary{}
	q := d.DB.Model(&ret).Column("player_summary.*", "Person")
	q = q.Where("player_summary.person_id = ?", pid).Where("tournament_id = ?", t.ID)

	err := q.Select(&ret)
	return &ret, err
}

// GetPlayerSummaries gets all player summaries for a tourmanent
func (d *Database) GetPlayerSummaries(t *Tournament) ([]*PlayerSummary, error) {
	ret := []*PlayerSummary{}
	q := d.DB.Model(&ret).Column("player_summary.*", "Person")
	err := q.Where("player_summary.tournament_id = ?", t.ID).Order("id ASC").Select(&ret)
	if err != nil {
		return nil, err
	}

	return ret, err
}

// GetPlayerState gets the player state for a player
func (d *Database) GetPlayerState(m *Match, idx int) (*PlayerState, error) {
	sts, err := d.GetPlayerStates(m)
	if err != nil {
		return nil, err
	}
	return sts[idx], err
}

// GetPlayerStates gets the players for all players in a match
func (d *Database) GetPlayerStates(m *Match) ([]*PlayerState, error) {
	st := []*PlayerState{}
	q := d.DB.Model(&st)

	args := make([]interface{}, 0)
	for _, p := range m.Players {
		args = append(args, p.ID)
	}

	q = q.WhereIn("player_id IN (?)", args...)
	err := q.Order("index ASC").Select(&st)
	return st, err
}

// SetPlayerState gets the player state for a player
func (d *Database) SetPlayerState(st *PlayerState) error {
	return d.DB.Update(st)
}

// UsurpTournament adds testing players
func (d *Database) UsurpTournament(t *Tournament, x int) error {
	query := `INSERT INTO player_summaries (tournament_id, person_id)
  SELECT ?, person_id FROM people
   WHERE NOT disabled
     AND person_id NOT IN (
         SELECT person_id FROM player_summaries
          WHERE tournament_id = ?)
   ORDER BY random() LIMIT ?;`

	_, err := d.DB.Exec(query, t.ID, t.ID, x)
	if err != nil {
		log.Printf("Usurping failed: %+v", err)
	}
	return err
}

// ClearTestTournaments deletes any tournament that doesn't begin with "DrunkenFall"
func (d *Database) ClearTestTournaments() error {
	_, err := d.DB.Model(&Tournament{}).Where("name NOT LIKE 'DrunkenFall%'").Delete()
	return err
}

// Close closes the database
func (d *Database) Close() error {
	return d.DB.Close()
}
