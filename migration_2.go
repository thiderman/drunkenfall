package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

type mig2prevPlayer struct {
	Person         *Person `json:"person"`
	Color          string  `json:"color"`
	PreferredColor string  `json:"preferred_color"`
	Shots          int     `json:"shots"`
	Sweeps         int     `json:"sweeps"`
	Kills          int     `json:"kills"`
	Self           int     `json:"self"`
	Explosions     int     `json:"explosions"`
	Matches        int     `json:"matches"`
	TotalScore     int     `json:"score"`
	Match          *Match  `json:"-"`
}

type mig2prevMatch struct {
	Players    []mig2prevPlayer `json:"players"`
	Judges     []Judge          `json:"judges"`
	Kind       string           `json:"kind"`
	Index      int              `json:"index"`
	Length     int              `json:"length"`
	Pause      time.Duration    `json:"pause"`
	Scheduled  time.Time        `json:"scheduled"`
	Started    time.Time        `json:"started"`
	Ended      time.Time        `json:"ended"`
	Tournament *Tournament      `json:"-"`
	ScoreOrder []int            `json:"score_order"`
	Commits    []Round          `json:"commits"`
}

type mig2prevTournament struct {
	Name      string           `json:"name"`
	ID        string           `json:"id"`
	Players   []mig2prevPlayer `json:"players"`
	Winners   []mig2prevPlayer `json:"winners"`
	Runnerups []string         `json:"runnerups"`
	Judges    []Judge          `json:"judges"`
	Tryouts   []*mig2prevMatch `json:"tryouts"`
	Semis     []*mig2prevMatch `json:"semis"`
	Final     *mig2prevMatch   `json:"final"`
	Current   CurrentMatch     `json:"current"`
	Opened    time.Time        `json:"opened"`
	Scheduled time.Time        `json:"scheduled"`
	Started   time.Time        `json:"started"`
	Ended     time.Time        `json:"ended"`
}

type mig2curPerson struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Email           string   `json:"email"`
	Nick            string   `json:"nick"`
	ColorPreference []string `json:"color_preference"`
	FacebookID      string   `json:"facebook_id"`
	FacebookToken   string   `json:"facebook_token"`
	AvatarURL       string   `json:"avatar_url"`
	Userlevel       int      `json:"userlevel"`
}

type mig2curTournament struct {
	Name      string           `json:"name"`
	ID        string           `json:"id"`
	Players   []mig2prevPlayer `json:"players"`
	Winners   []mig2prevPlayer `json:"winners"`
	Runnerups []*mig2curPerson `json:"runnerups"`
	Judges    []Judge          `json:"judges"`
	Tryouts   []*mig2prevMatch `json:"tryouts"`
	Semis     []*mig2prevMatch `json:"semis"`
	Final     *mig2prevMatch   `json:"final"`
	Current   CurrentMatch     `json:"current"`
	Opened    time.Time        `json:"opened"`
	Scheduled time.Time        `json:"scheduled"`
	Started   time.Time        `json:"started"`
	Ended     time.Time        `json:"ended"`
}

// MigrateTournamentRunnerupStringPerson changes Tournament.Runnerups to be a
// list of Person{} objects rather than a list of string names.
func MigrateTournamentRunnerupStringPerson(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Load the tournaments
		var ts []*mig2prevTournament
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(TournamentKey)
			err := b.ForEach(func(k []byte, v []byte) error {
				t := &mig2prevTournament{}
				err := json.Unmarshal(v, t)
				if err != nil {
					log.Print(err)
					return err
				}
				if err != nil {
					return err
				}

				ts = append(ts, t)
				return nil
			})
			return err
		})

		if err != nil {
			log.Fatal(err)
		}

		var out []*mig2curTournament
		for _, pt := range ts {
			t := mig2curTournament{
				Name:      pt.Name,
				ID:        pt.ID,
				Players:   pt.Players,
				Winners:   pt.Winners,
				Judges:    pt.Judges,
				Tryouts:   pt.Tryouts,
				Semis:     pt.Semis,
				Final:     pt.Final,
				Current:   pt.Current,
				Opened:    pt.Opened,
				Scheduled: pt.Scheduled,
				Started:   pt.Started,
				Ended:     pt.Ended,
			}

			load := func(id string) *mig2curPerson {
				// Prefixing tx and err with l because otherwise heavy shadowing will
				// break angrily.
				ltx, lerr := db.Begin(false)
				if lerr != nil {
					log.Fatal(lerr)
					return nil
				}
				defer ltx.Rollback()

				b := ltx.Bucket(PeopleKey)
				bs := b.Get([]byte(id))
				p := &mig2curPerson{}
				_ = json.Unmarshal(bs, p)
				return p
			}

			// Convert runnerups from strings to objects...
			for _, n := range pt.Runnerups {
				// ...by using the Person object on the players.
				for _, pl := range pt.Players {
					if pl.Person.Nick == n {
						p := load(pl.Person.ID)
						t.Runnerups = append(t.Runnerups, p)
					}
				}
			}

			out = append(out, &t)
		}

		// Loop the tournament results and save them into the db
		for _, t := range out {
			j, err := json.Marshal(t)
			if err != nil {
				log.Fatal(err)
			}

			err = tx.Bucket(TournamentKey).Put([]byte(t.ID), j)
			if err != nil {
				log.Fatal(err)
			}
		}

		return setVersion(tx, 3)
	})
}
