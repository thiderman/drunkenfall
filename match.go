package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// Match represents a game being played
type Match struct {
	Players    []Player    `json:"players"`
	Judges     []Judge     `json:"judges"`
	Kind       string      `json:"kind"`
	Index      int         `json:"index"`
	Length     int         `json:"length"`
	Started    time.Time   `json:"started"`
	Ended      time.Time   `json:"ended"`
	Tournament *Tournament `json:"-"`

	// Stores the index to the player in the relative position.  E.g. if player
	// 3 is in the lead, ScoreOrder[0] will be 2 (the index of player 3).
	ScoreOrder []int `json:"score_order"`

	// One commit per round - the changeset of what happened in it.
	Commits []MatchCommit `json:"commits"`

	presentColors map[string]bool
}

// MatchCommit is a state commit for a round of a match
type MatchCommit struct {
	Kills     [][]int `json:"kills"`
	Shots     []bool  `json:"shots"`
	Committed string  `json:"comitted"` // ISO-8601
}

// NewMatch creates a new Match for usage!
func NewMatch(t *Tournament, index int, kind string) *Match {
	m := Match{
		Index:      index,
		Kind:       kind,
		Tournament: t,
		Length:     10,
	}
	m.presentColors = make(map[string]bool)

	// Finals are longer <3
	if kind == "final" {
		m.Length = 20
	}

	return &m
}

func (m *Match) String() string {
	var tempo string
	var name string

	if !m.IsStarted() {
		tempo = "not started"
	} else if m.IsEnded() {
		tempo = "ended"
	} else {
		tempo = "playing"
	}

	if m.Kind == "final" {
		name = "Final"
	} else {
		name = fmt.Sprintf("%s %d", strings.Title(m.Kind), m.Index+1)
	}

	names := make([]string, 0, len(m.Players))
	for _, p := range m.Players {
		names = append(names, p.Name())
	}

	return fmt.Sprintf(
		"<%s: %s - %s>",
		name,
		strings.Join(names, " / "),
		tempo,
	)
}

// Title returns a title string
func (m *Match) Title() string {
	l := 2
	if m.Kind == "final" {
		return "Final"
	} else if m.Kind == "tryout" {
		l = len(m.Tournament.Tryouts)
	}

	out := fmt.Sprintf(
		"%s %d/%d",
		strings.Title(m.Kind),
		m.Index+1,
		l,
	)
	return out
}

// URL builds the URL to the match
func (m *Match) URL() string {
	out := fmt.Sprintf(
		"/%s/%s/%d/",
		m.Tournament.ID,
		m.Kind,
		m.Index,
	)
	return out
}

// AddPlayer adds a player to the match
func (m *Match) AddPlayer(p Player) error {
	if len(m.Players) == 4 {
		return errors.New("cannot add fifth player")
	}

	// Reset all possible scores
	p.Reset()

	c := p.PreferredColor()
	p.OriginalColor = c
	if _, ok := m.presentColors[c]; ok {
		// Color is already present - give the player a new random one.
		c = Colors.Available(m).Random()
		log.Printf("Corrected color of %s from %s to %s", p.Person.Nick, p.OriginalColor, c)
	}

	// Set the player color
	p.Color = c
	m.presentColors[c] = true

	// Also set the match pointer
	p.Match = m

	m.Players = append(m.Players, p)

	return nil
}

// UpdatePlayer updates a player for the given match
func (m *Match) UpdatePlayer(p Player) error {
	for i, o := range m.Players {
		if o.Name() == p.Name() {
			m.Players[i] = p
		}
	}
	return nil
}

// Commit adds a state of the players
func (m *Match) Commit(c MatchCommit) {
	for i, score := range c.Kills {
		shotGiven := false
		ups := score[0]
		downs := score[1]

		// If the score is 3, then this player killed everyone else.
		// Count that as a sweep.
		if ups == 3 {
			m.Players[i].AddSweep()
			// Since AddSweep gives a shot, we shouldn't give another further down.
			shotGiven = true

			// If we have a sweep and a down, we need to redact the
			// extra shot, because no player should get more than
			// one shot per round. Removing one from here makes
			// sure that the `downs` calculation below still adds
			// the one.
			if downs == 1 {
				m.Players[i].RemoveShot()
			}
		} else if ups > 0 {
			m.Players[i].AddKill(ups)
		}

		if downs != 0 {
			m.Players[i].AddSelf()
			shotGiven = true
		}

		if c.Shots[i] && !shotGiven {
			m.Players[i].AddShot()
		}
	}

	m.ScoreOrder = m.MakeScoreOrder()
	m.Commits = append(m.Commits, c)
	_ = m.Tournament.Persist()
}

// Start starts the match
func (m *Match) Start() error {
	if !m.Started.IsZero() {
		return errors.New("match already started")
	}

	for i := range m.Players {
		m.Players[i].Reset()
		m.Players[i].Match = m
	}

	m.Started = time.Now()
	if m.Tournament != nil {
		m.Tournament.Persist()
	}
	return nil
}

// End signals that the match has ended
//
// It is also the place that moves players into either the Runnerup bracket
// or into their place in the semis.
func (m *Match) End() error {
	if !m.Ended.IsZero() {
		return errors.New("match already ended")
	}

	// XXX(thiderman): In certain test cases a Commit() might not have been run
	// and therefore this might not have been set. Since the calculation is
	// quick and has no side effects, it's easier to just add it here now. In
	// the future, make the tests better.
	m.ScoreOrder = m.MakeScoreOrder()

	// Give the winner one last shot
	winner := m.ScoreOrder[0]
	m.Players[winner].AddShot()

	m.Ended = time.Now()
	// TODO: This is for the tests not to break. Fix by setting up better tests.
	if m.Tournament != nil {
		if m.Kind == "final" {
			m.Tournament.AwardMedals(m)
		} else {
			m.Tournament.MovePlayers(m)
		}

		m.Tournament.Persist()
	}
	return nil
}

// IsStarted returns boolean whether the match has started or not
func (m *Match) IsStarted() bool {
	return !m.Started.IsZero()
}

// IsEnded returns boolean whether the match has ended or not
func (m *Match) IsEnded() bool {
	return !m.Ended.IsZero()
}

// CanStart returns boolean the match can be controlled or not
func (m *Match) CanStart() bool {
	return !m.IsStarted() && !m.IsEnded()
}

// CanEnd returns boolean whether the match can be ended or not
func (m *Match) CanEnd() bool {
	if !m.IsOpen() {
		return false
	}
	for _, p := range m.Players {
		if p.Kills >= m.Length {
			return true
		}
	}
	return false
}

// IsOpen returns boolean the match can be controlled or not
func (m *Match) IsOpen() bool {
	return m.IsStarted() && !m.IsEnded()
}

// MakeScoreOrder returns the score order of the current state of the match
func (m *Match) MakeScoreOrder() (ret []int) {
	ps := SortByScore(m.Players)
	for _, p := range ps {
		for i, o := range m.Players {
			if p.Name() == o.Name() {
				ret = append(ret, i)
				break
			}
		}
	}

	return
}

// NewMatchCommit makes a new MatchCommit object from a CommitRequest
func NewMatchCommit(c CommitRequest) MatchCommit {
	states := c.State
	m := MatchCommit{
		[][]int{
			[]int{states[0].Ups, states[0].Downs},
			[]int{states[1].Ups, states[1].Downs},
			[]int{states[2].Ups, states[2].Downs},
			[]int{states[3].Ups, states[3].Downs},
		},
		[]bool{
			states[0].Shot,
			states[1].Shot,
			states[2].Shot,
			states[3].Shot,
		},
		// ISO-8601 timestamp
		time.Now().UTC().Format(time.RFC3339),
	}

	return m
}
