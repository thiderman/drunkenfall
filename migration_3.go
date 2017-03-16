package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"log"
	"sort"
	"time"
)

type mig3prevPlayer struct {
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

type mig3prevMatch struct {
	Players    []mig3prevPlayer `json:"players"`
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

type mig3prevTournament struct {
	Name      string           `json:"name"`
	ID        string           `json:"id"`
	Players   []mig3prevPlayer `json:"players"`
	Winners   []mig3prevPlayer `json:"winners"`
	Runnerups []*Person        `json:"runnerups"`
	Judges    []Judge          `json:"judges"`
	Tryouts   []*mig3prevMatch `json:"tryouts"`
	Semis     []*mig3prevMatch `json:"semis"`
	Final     *mig3prevMatch   `json:"final"`
	Current   CurrentMatch     `json:"current"`
	Opened    time.Time        `json:"opened"`
	Scheduled time.Time        `json:"scheduled"`
	Started   time.Time        `json:"started"`
	Ended     time.Time        `json:"ended"`
}

type mig3curMatch struct {
	Players    []mig3prevPlayer `json:"players"`
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
	Commits    []Round          `json:"commits"`
}

type mig3curTournament struct {
	Name      string           `json:"name"`
	ID        string           `json:"id"`
	Players   []mig3prevPlayer `json:"players"`
	Winners   []mig3prevPlayer `json:"winners"`
	Runnerups []*Person        `json:"runnerups"`
	Judges    []Judge          `json:"judges"`
	Tryouts   []*mig3curMatch  `json:"tryouts"`
	Semis     []*mig3curMatch  `json:"semis"`
	Final     *mig3curMatch    `json:"final"`
	Current   CurrentMatch     `json:"current"`
	Opened    time.Time        `json:"opened"`
	Scheduled time.Time        `json:"scheduled"`
	Started   time.Time        `json:"started"`
	Ended     time.Time        `json:"ended"`
}

// mig3ByKills is a sort.Interface that sorts mig3prevPlayers by their kills
type mig3ByKills []mig3prevPlayer

func (s mig3ByKills) Len() int {
	return len(s)

}
func (s mig3ByKills) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]

}
func (s mig3ByKills) Less(i, j int) bool {
	// Technically not Less, but we want biggest first...
	return s[i].Kills > s[j].Kills
}

// mig3SortByKills returns a list in order of the kills the mig3prevPlayers have
func mig3SortByKills(ps []mig3prevPlayer) []mig3prevPlayer {
	tmp := make([]mig3prevPlayer, len(ps))
	copy(tmp, ps)
	sort.Sort(mig3ByKills(tmp))
	return tmp
}

// MigrateMatchScoreOrderKillOrder updates the ScoreOrder since the
// implementation had a bugfix. Previously it sorted by the entertainment
// score and not the amount of kills. This has been fixed and all the scores
// need to be updated to reflect this in previous tournaments.
// nolint: gocyclo
func MigrateMatchScoreOrderKillOrder(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Load the tournaments
		var ts []*mig3prevTournament
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket(TournamentKey)
			err := b.ForEach(func(k []byte, v []byte) error {
				t := &mig3prevTournament{}
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

		var out []*mig3curTournament
		for _, pt := range ts {
			t := mig3curTournament{
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

			t.Tryouts = make([]*mig3curMatch, len(pt.Tryouts))
			for x, pm := range pt.Tryouts {
				t.Tryouts[x] = &mig3curMatch{
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
					Commits:    pm.Commits,
				}

				t.Tryouts[x].KillOrder = make([]int, 0)
				ps := mig3SortByKills(pt.Tryouts[x].Players)
				for _, p := range ps {
					for i, o := range pt.Tryouts[x].Players {
						if p.Person.ID == o.Person.ID {
							t.Tryouts[x].KillOrder = append(t.Tryouts[x].KillOrder, i)
							break
						}
					}
				}
			}

			t.Semis = make([]*mig3curMatch, len(pt.Semis))
			for x, pm := range pt.Semis {
				t.Semis[x] = &mig3curMatch{
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
					Commits:    pm.Commits,
				}

				t.Semis[x].KillOrder = make([]int, 0)
				ps := mig3SortByKills(pt.Semis[x].Players)
				for _, p := range ps {
					for i, o := range pt.Semis[x].Players {
						if p.Person.ID == o.Person.ID {
							t.Semis[x].KillOrder = append(t.Semis[x].KillOrder, i)
							break
						}
					}
				}
			}

			t.Final = &mig3curMatch{
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
				Commits:    pt.Final.Commits,
			}

			t.Final.KillOrder = make([]int, 0)
			ps := mig3SortByKills(pt.Final.Players)
			for _, p := range ps {
				for i, o := range pt.Final.Players {
					if p.Person.ID == o.Person.ID {
						t.Final.KillOrder = append(t.Final.KillOrder, i)
						break
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

		return setVersion(tx, 4)
	})
}
