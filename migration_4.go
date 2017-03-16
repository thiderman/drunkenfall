package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

type mig4prevPlayer struct {
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

type mig4prevMatch struct {
	Players    []mig4prevPlayer `json:"players"`
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

type mig4prevTournament struct {
	Name      string           `json:"name"`
	ID        string           `json:"id"`
	Players   []mig4prevPlayer `json:"players"`
	Winners   []mig4prevPlayer `json:"winners"`
	Runnerups []*Person        `json:"runnerups"`
	Judges    []Judge          `json:"judges"`
	Tryouts   []*mig4prevMatch `json:"tryouts"`
	Semis     []*mig4prevMatch `json:"semis"`
	Final     *mig4prevMatch   `json:"final"`
	Current   CurrentMatch     `json:"current"`
	Opened    time.Time        `json:"opened"`
	Scheduled time.Time        `json:"scheduled"`
	Started   time.Time        `json:"started"`
	Ended     time.Time        `json:"ended"`
}

type mig4curMatch struct {
	Players    []mig4prevPlayer `json:"players"`
	Judges     []Judge          `json:"judges"`
	Kind       string           `json:"kind"`
	Index      int              `json:"index"`
	Length     int              `json:"length"`
	Pause      time.Duration    `json:"pause"`
	Scheduled  time.Time        `json:"scheduled"`
	Started    time.Time        `json:"started"`
	Ended      time.Time        `json:"ended"`
	Tournament *Tournament      `json:"-"`
	KillOrder  []int            `json:"kill_order"`
	Rounds     []Round          `json:"rounds"`
}

type mig4curTournament struct {
	Name      string           `json:"name"`
	ID        string           `json:"id"`
	Players   []mig4prevPlayer `json:"players"`
	Winners   []mig4prevPlayer `json:"winners"`
	Runnerups []*Person        `json:"runnerups"`
	Judges    []Judge          `json:"judges"`
	Tryouts   []*mig4curMatch  `json:"tryouts"`
	Semis     []*mig4curMatch  `json:"semis"`
	Final     *mig4curMatch    `json:"final"`
	Current   CurrentMatch     `json:"current"`
	Opened    time.Time        `json:"opened"`
	Scheduled time.Time        `json:"scheduled"`
	Started   time.Time        `json:"started"`
	Ended     time.Time        `json:"ended"`
}

// MigrateMatchCommitToRound changes the type signature of the object that
// stores differences between rounds and actually calls it... Round{}.
// nolint: gocyclo
func MigrateMatchCommitToRound(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Load the tournaments
		var ts []*mig4prevTournament
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(TournamentKey)
			err := b.ForEach(func(k []byte, v []byte) error {
				t := &mig4prevTournament{}
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

		var out []*mig4curTournament
		for _, pt := range ts {
			t := mig4curTournament{
				Name:      pt.Name,
				ID:        pt.ID,
				Players:   pt.Players,
				Winners:   pt.Winners,
				Runnerups: pt.Runnerups,
				Judges:    pt.Judges,
				Current:   pt.Current,
				Opened:    pt.Opened,
				Scheduled: pt.Scheduled,
				Started:   pt.Started,
				Ended:     pt.Ended,
			}

			t.Tryouts = make([]*mig4curMatch, len(pt.Tryouts))
			for x, pm := range pt.Tryouts {
				t.Tryouts[x] = &mig4curMatch{
					Players:    pm.Players,
					Judges:     pm.Judges,
					Kind:       pm.Kind,
					Index:      pm.Index,
					Length:     pm.Length,
					Pause:      pm.Pause,
					Scheduled:  pm.Scheduled,
					Started:    pm.Started,
					Ended:      pm.Ended,
					Tournament: pm.Tournament,
					Rounds:     pm.Commits,
				}
			}

			t.Semis = make([]*mig4curMatch, len(pt.Semis))
			for x, pm := range pt.Semis {
				t.Semis[x] = &mig4curMatch{
					Players:    pm.Players,
					Judges:     pm.Judges,
					Kind:       pm.Kind,
					Index:      pm.Index,
					Length:     pm.Length,
					Pause:      pm.Pause,
					Scheduled:  pm.Scheduled,
					Started:    pm.Started,
					Ended:      pm.Ended,
					Tournament: pm.Tournament,
					Rounds:     pm.Commits,
				}
			}

			t.Final = &mig4curMatch{
				Players:    pt.Final.Players,
				Judges:     pt.Final.Judges,
				Kind:       pt.Final.Kind,
				Index:      pt.Final.Index,
				Length:     pt.Final.Length,
				Pause:      pt.Final.Pause,
				Scheduled:  pt.Final.Scheduled,
				Started:    pt.Final.Started,
				Ended:      pt.Final.Ended,
				Tournament: pt.Final.Tournament,
				Rounds:     pt.Final.Commits,
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

		return setVersion(tx, 5)
	})
}
