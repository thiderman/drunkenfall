package main

import (
	"errors"
	"fmt"
	"github.com/deckarep/golang-set"
	"log"
	"strings"
	"time"
)

// Match represents a game being played
//
// Match.ScoreOrder stores the index to the player in the relative position.
// E.g. if player 3 is in the lead, ScoreOrder[0] will be 2 (the index of
// player 3).
//
// Match.Commits is a list of one commit per round and represents the
// changeset of what happened in the match.
type Match struct {
	Players       []Player      `json:"players"`
	Judges        []Judge       `json:"judges"`
	Kind          string        `json:"kind"`
	Index         int           `json:"index"`
	Length        int           `json:"length"`
	Pause         time.Duration `json:"pause"`
	Scheduled     time.Time     `json:"scheduled"`
	Started       time.Time     `json:"started"`
	Ended         time.Time     `json:"ended"`
	Tournament    *Tournament   `json:"-"`
	KillOrder     []int         `json:"kill_order"`
	Rounds        []Round       `json:"commits"`
	presentColors mapset.Set
}

// Round is a state commit for a round of a match
type Round struct {
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
		Length:     t.length,
		Pause:      time.Minute * 5,
	}
	m.presentColors = mapset.NewSet()

	// The pause between tryout/semi brackets should be longer
	if kind == semi && index == 0 {
		m.Pause = time.Minute * 10
	}

	// Finals are longer <3
	if kind == final {
		m.Length = t.finalLength
		m.Pause = time.Minute * 10
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

	if m.Kind == final {
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
	if m.Kind == final {
		return "Final"
	} else if m.Kind == tryout {
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

	// Add all the previous players' colors.
	// This is to fix a bug with the presentColors map if the app has been
	// restarted. They cannot be added multiple tines anyway.
	for _, p := range m.Players {
		m.presentColors.Add(p.Color)
	}

	p.Color = p.PreferredColor
	m.presentColors.Add(p.Color)

	// Also set the match pointer
	p.Match = m

	m.Players = append(m.Players, p)

	// If we're adding the fourth player, it's time to correct the conflicts
	if len(m.Players) == 4 && len(m.presentColors.ToSlice()) != 4 {
		if err := m.CorrectFuckingColorConflicts(); err != nil {
			return err
		}
	}

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

// CorrectFuckingColorConflicts corrects color conflicts :@
func (m *Match) CorrectFuckingColorConflicts() error {
	// Make a map of conflicting players keyed on the color
	pairs := make(map[string][]Player)
	for _, color := range m.presentColors.ToSlice() {
		c := color.(string)
		for _, p := range m.Players {
			if p.PreferredColor == c {
				pairs[c] = append(pairs[c], p)
			}
		}
	}

	// Loop over the colors and
	for _, pair := range pairs {
		// If there are two or more players in the group, there is a conflict and
		// they need to be corrected.
		if len(pair) >= 2 {
			// We want to sort them by score, so that we can let the player with the
			// highest score keep their color.
			ps, err := SortByColorConflicts(pair)
			if err != nil {
				return err
			}

			for _, p := range ps[1:] {
				// For the players with lower scores, set their new colors
				new := RandomColor(AvailableColors(m))
				m.presentColors.Add(new)
				p.Color = new

				// Since we are using the tournament level Player object, the compound
				// scores from all other matches are currently on it. Reset that.
				p.Reset()

				if err := m.UpdatePlayer(p); err != nil {
					return err
				}
				log.Println(fmt.Sprintf("%s corrected from %s to %s", p.Name(), p.PreferredColor, new))
			}
		}
	}
	return nil
}

// Commit Applies the round actions to the state of the players
func (m *Match) Commit(round Round) {
	for i, score := range round.Kills {
		kills := score[0]
		self := score[1]

		m.Players[i].AddKills(kills)
		if self == -1 {
			m.Players[i].AddSelf()
		}
		if self == -1 || kills == 3 || round.Shots[i] {
			m.Players[i].AddShot()
		}
	}

	m.KillOrder = m.MakeKillOrder()
	m.Rounds = append(m.Rounds, round)
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
	m.KillOrder = m.MakeKillOrder()

	// Give the winner one last shot
	winner := m.KillOrder[0]
	m.Players[winner].AddShot()

	m.Ended = time.Now()
	// TODO: This is for the tests not to break. Fix by setting up better tests.
	if m.Tournament != nil {
		if m.Kind == final {
			if err := m.Tournament.AwardMedals(m); err != nil {
				return err
			}
		} else {
			if err := m.Tournament.MovePlayers(m); err != nil {
				return err
			}
		}

		m.Tournament.Persist()
	}
	return nil
}

// Reset resets all the player scores to zero and removes all the commits
func (m *Match) Reset() error {
	// Reset dem players
	for i := range m.Players {
		m.Players[i].Reset()
	}

	// And remove all the rounds
	m.Rounds = make([]Round, 0)

	m.Tournament.Persist()
	return nil
}

// SetTime sets the scheduled time based on the Pause attribute
func (m *Match) SetTime(minutes int) {
	log.Print(fmt.Sprintf("Setting time for %s in %d minutes", m, minutes))
	m.Scheduled = time.Now().Add(time.Minute * time.Duration(minutes))
	log.Print(m.Scheduled)
	m.Tournament.Persist()
}

// IsStarted returns boolean whether the match has started or not
func (m *Match) IsStarted() bool {
	return !m.Started.IsZero()
}

// IsEnded returns boolean whether the match has ended or not
func (m *Match) IsEnded() bool {
	return !m.Ended.IsZero()
}

// IsScheduled returns boolean whether the match has been scheduled or not
func (m *Match) IsScheduled() bool {
	return !m.Scheduled.IsZero()
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

// MakeKillOrder returns the score in order of the number of kills in the match.
func (m *Match) MakeKillOrder() (ret []int) {
	ps := SortByKills(m.Players)
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
func NewMatchCommit(c CommitRequest) Round {
	states := c.State
	m := Round{
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
