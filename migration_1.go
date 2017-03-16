package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

type mig1prevPlayer struct {
	Person        *Person `json:"person"`
	Color         string  `json:"color"`
	OriginalColor string  `json:"original_color"`
	Shots         int     `json:"shots"`
	Sweeps        int     `json:"sweeps"`
	Kills         int     `json:"kills"`
	Self          int     `json:"self"`
	Explosions    int     `json:"explosions"`
	Matches       int     `json:"matches"`
	TotalScore    int     `json:"score"`
	Match         *Match  `json:"-"`
}

type mig1prevMatch struct {
	Players    []mig1prevPlayer `json:"players"`
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

type mig1prevTournament struct {
	Name      string           `json:"name"`
	ID        string           `json:"id"`
	Players   []mig1prevPlayer `json:"players"`
	Winners   []mig1prevPlayer `json:"winners"`
	Runnerups []string         `json:"runnerups"`
	Judges    []Judge          `json:"judges"`
	Tryouts   []*mig1prevMatch `json:"tryouts"`
	Semis     []*mig1prevMatch `json:"semis"`
	Final     *mig1prevMatch   `json:"final"`
	Current   CurrentMatch     `json:"current"`
	Opened    time.Time        `json:"opened"`
	Scheduled time.Time        `json:"scheduled"`
	Started   time.Time        `json:"started"`
	Ended     time.Time        `json:"ended"`
}

type mig1curPlayer struct {
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

type mig1curMatch struct {
	Players    []mig1curPlayer `json:"players"`
	Judges     []Judge         `json:"judges"`
	Kind       string          `json:"kind"`
	Index      int             `json:"index"`
	Length     int             `json:"length"`
	Pause      time.Duration   `json:"pause"`
	Scheduled  time.Time       `json:"scheduled"`
	Started    time.Time       `json:"started"`
	Ended      time.Time       `json:"ended"`
	Tournament *Tournament     `json:"-"`
	ScoreOrder []int           `json:"score_order"`
	Commits    []Round         `json:"commits"`
}

type mig1curTournament struct {
	Name      string          `json:"name"`
	ID        string          `json:"id"`
	Players   []mig1curPlayer `json:"players"`
	Winners   []mig1curPlayer `json:"winners"`
	Runnerups []string        `json:"runnerups"`
	Judges    []Judge         `json:"judges"`
	Tryouts   []*mig1curMatch `json:"tryouts"`
	Semis     []*mig1curMatch `json:"semis"`
	Final     *mig1curMatch   `json:"final"`
	Current   CurrentMatch    `json:"current"`
	Opened    time.Time       `json:"opened"`
	Scheduled time.Time       `json:"scheduled"`
	Started   time.Time       `json:"started"`
	Ended     time.Time       `json:"ended"`
}

// MigrateOriginalColorPreferredColor changes Player{} to have a preferred
// color rather than an original color
// nolint: gocyclo
func MigrateOriginalColorPreferredColor(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Load the tournaments
		var ts []*mig1prevTournament
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(TournamentKey)
			err := b.ForEach(func(k []byte, v []byte) error {
				t := &mig1prevTournament{}
				err := json.Unmarshal(v, t)
				if err != nil {
					log.Print(err)
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

		var out []*mig1curTournament
		for _, pt := range ts {
			t := mig1curTournament{
				Name:      pt.Name,
				ID:        pt.ID,
				Judges:    pt.Judges,
				Runnerups: pt.Runnerups,
				Current:   pt.Current,
				Opened:    pt.Opened,
				Scheduled: pt.Scheduled,
				Started:   pt.Started,
				Ended:     pt.Ended,
			}

			// Update the player objects
			t.Players = make([]mig1curPlayer, len(pt.Players))
			for x, pp := range pt.Players {
				t.Players[x] = mig1curPlayer{
					Person:         pp.Person,
					Color:          pp.Person.ColorPreference[0],
					PreferredColor: pp.Person.ColorPreference[0],
					Shots:          pp.Shots,
					Sweeps:         pp.Sweeps,
					Kills:          pp.Kills,
					Self:           pp.Self,
					Explosions:     pp.Explosions,
					Matches:        pp.Matches,
					TotalScore:     pp.TotalScore,
				}
			}

			// Update the tryouts
			t.Tryouts = make([]*mig1curMatch, len(pt.Tryouts))
			for x, pm := range pt.Tryouts {
				t.Tryouts[x] = &mig1curMatch{
					Judges:     pm.Judges,
					Kind:       pm.Kind,
					Index:      pm.Index,
					Length:     pm.Length,
					Pause:      pm.Pause,
					Scheduled:  pm.Scheduled,
					Started:    pm.Started,
					Ended:      pm.Ended,
					Tournament: pm.Tournament,
					ScoreOrder: pm.ScoreOrder,
					Commits:    pm.Commits,
				}

				// For each match, also update each of the player objects
				t.Tryouts[x].Players = make([]mig1curPlayer, len(pt.Tryouts[x].Players))
				for y, pp := range pt.Tryouts[x].Players {
					t.Tryouts[x].Players[y] = mig1curPlayer{
						Person:         pp.Person,
						Color:          pp.Color,
						PreferredColor: pp.OriginalColor,
						Shots:          pp.Shots,
						Sweeps:         pp.Sweeps,
						Kills:          pp.Kills,
						Self:           pp.Self,
						Explosions:     pp.Explosions,
						Matches:        pp.Matches,
						TotalScore:     pp.TotalScore,
					}
				}
			}

			// Update the semis
			t.Semis = make([]*mig1curMatch, len(pt.Semis))
			for x, pm := range pt.Semis {
				t.Semis[x] = &mig1curMatch{
					Judges:     pm.Judges,
					Kind:       pm.Kind,
					Index:      pm.Index,
					Length:     pm.Length,
					Pause:      pm.Pause,
					Scheduled:  pm.Scheduled,
					Started:    pm.Started,
					Ended:      pm.Ended,
					Tournament: pm.Tournament,
					ScoreOrder: pm.ScoreOrder,
					Commits:    pm.Commits,
				}

				// For each match, also update each of the player objects
				t.Semis[x].Players = make([]mig1curPlayer, len(pt.Semis[x].Players))
				for y, pp := range pt.Semis[x].Players {
					t.Semis[x].Players[y] = mig1curPlayer{
						Person:         pp.Person,
						Color:          pp.Person.ColorPreference[0],
						PreferredColor: pp.Person.ColorPreference[0],
						Shots:          pp.Shots,
						Sweeps:         pp.Sweeps,
						Kills:          pp.Kills,
						Self:           pp.Self,
						Explosions:     pp.Explosions,
						Matches:        pp.Matches,
						TotalScore:     pp.TotalScore,
					}
				}
			}

			// Update the final
			t.Final = &mig1curMatch{
				Judges:     pt.Final.Judges,
				Kind:       pt.Final.Kind,
				Index:      pt.Final.Index,
				Length:     pt.Final.Length,
				Pause:      pt.Final.Pause,
				Scheduled:  pt.Final.Scheduled,
				Started:    pt.Final.Started,
				Ended:      pt.Final.Ended,
				Tournament: pt.Final.Tournament,
				ScoreOrder: pt.Final.ScoreOrder,
				Commits:    pt.Final.Commits,
			}
			t.Final.Players = make([]mig1curPlayer, len(pt.Final.Players))

			// For each match, also update each of the player objects
			for y, pp := range pt.Final.Players {
				t.Final.Players[y] = mig1curPlayer{
					Person:         pp.Person,
					Color:          pp.Person.ColorPreference[0],
					PreferredColor: pp.Person.ColorPreference[0],
					Shots:          pp.Shots,
					Sweeps:         pp.Sweeps,
					Kills:          pp.Kills,
					Self:           pp.Self,
					Explosions:     pp.Explosions,
					Matches:        pp.Matches,
					TotalScore:     pp.TotalScore,
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

		return setVersion(tx, 2)
	})
}
