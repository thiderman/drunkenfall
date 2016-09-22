package main

import (
	"fmt"
	"math/rand"
	"sort"
)

// ColorList is the representation of colors players can choose from
type ColorList []string

// Colors is a list of the available player colors
var Colors = ColorList{
	"green",
	"blue",
	"pink",
	"orange",
	"white",
	"yellow",
	"cyan",
	"purple",
	"red",
}

// ScoreData is a structured Key/Value pair list for scores
type ScoreData struct {
	Key    string
	Value  int
	Player *Player
}

// Player is a Participant that is actively participating in battles.
type Player struct {
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

// NewPlayer returns a new instance of a player
func NewPlayer(ps *Person) Player {
	p := Player{Person: ps}
	return p
}

func (p *Player) String() string {
	IsPointed := "."
	if p.Match != nil {
		IsPointed = "!"
	}
	return fmt.Sprintf(
		"<%s: %dsh %dsw %dk %ds %de %s>",
		p.Name(),
		p.Shots,
		p.Sweeps,
		p.Kills,
		p.Self,
		p.Explosions,
		IsPointed,
	)
}

// Name returns the nickname
func (p *Player) Name() string {
	return p.Person.Nick
}

// PreferredColor returns the color
func (p *Player) PreferredColor() string {
	return p.Person.PreferredColor()
}

// Score calculates the score to determine runnerup positions.
func (p *Player) Score() (out int) {
	// This algorithm is probably flawed, but at least it should be able to
	// determine who is the most entertaining.
	// When executed, a sweep is basically 14 points since scoring a sweep
	// also comes with a shot and three kills.

	out += p.Sweeps * 5
	out += p.Shots * 3
	out += p.Kills * 2
	out += p.Self
	out += p.Explosions

	return
}

// ScoreData returns this players set of ScoreData
func (p *Player) ScoreData() []ScoreData {
	sd := []ScoreData{
		{Key: "kills", Value: p.Kills, Player: p},
		{Key: "shots", Value: p.Shots, Player: p},
		{Key: "sweeps", Value: p.Sweeps, Player: p},
		{Key: "self", Value: p.Self, Player: p},
		{Key: "explosions", Value: p.Explosions, Player: p},
	}
	return sd
}

// URL returns the URL to this player
//
// Used for action URLs
func (p *Player) URL() string {
	out := fmt.Sprintf(
		"%s/%d",
		p.Match.URL(),
		p.Index(),
	)
	return out
}

// Classes returns the CSS color classes for the player
func (p *Player) Classes() string {
	if p.Match != nil && p.Match.IsEnded() {
		ps := ByScore(p.Match.Players)
		if ps[0].Name() == p.Name() {
			// Always gold for the winner
			return "gold"
		} else if ps[1].Name() == p.Name() {
			// Silver for the second, unless there is a short amount of tryouts
			if p.Match.Kind != "tryout" || len(p.Match.Tournament.Tryouts) <= 4 {
				return "silver"
			}
		} else if ps[2].Name() == p.Name() && p.Match.Kind == "final" {
			return "bronze"
		}

		return "out"
	}
	return p.PreferredColor()
}

// RandomizeColor sets the color of this player to an unused one
func (p *Player) RandomizeColor(m *Match) {
	p.Color = Colors.Available(m).Random()
}

// Index returns the index in the current match
func (p *Player) Index() int {
	if p.Match != nil {
		for i, o := range p.Match.Players {
			if p.Name() == o.Name() {
				return i
			}
		}
	}
	return -1
}

// Action performs an action for a player
func (p *Player) Action(action, dir string) error {
	// TODO: This could use with a refactoring...
	if dir == "up" {
		switch action {
		case "kills":
			p.AddKill()
		case "shots":
			p.AddShot()
		case "sweeps":
			p.AddSweep()
		case "self":
			p.AddSelf()
		case "explosions":
			p.AddExplosion()
		}
	} else {
		switch action {
		case "kills":
			p.RemoveKill()
		case "shots":
			p.RemoveShot()
		case "sweeps":
			p.RemoveSweep()
		case "self":
			p.RemoveSelf()
		case "explosions":
			p.RemoveExplosion()
		}
	}

	// Save the change to the database
	p.Match.Tournament.Persist()

	return nil
}

// AddShot increases the shot count
func (p *Player) AddShot() {
	p.Shots++
}

// RemoveShot decreases the shot count
// Fails silently if shots are zero.
func (p *Player) RemoveShot() {
	if p.Shots == 0 {
		return
	}
	p.Shots--
}

// AddSweep increases the sweep count, gives three kills and a shot.
func (p *Player) AddSweep() {
	p.Sweeps++
	p.AddShot()
	p.AddKill()
	p.AddKill()
	p.AddKill()
}

// RemoveSweep decreases the sweep count, three kills and a shot
// Fails silently if sweeps are zero.
func (p *Player) RemoveSweep() {
	if p.Sweeps == 0 {
		return
	}
	p.Sweeps--
	p.RemoveShot()
	p.RemoveKill()
	p.RemoveKill()
	p.RemoveKill()
}

// AddKill increases the kill count
func (p *Player) AddKill(kills ...int) {
	// This is basically only to help out with testing.
	// Adding an optional argument with the amount of kills lets us just use
	// one call to AddKill() rather than 10.
	if len(kills) > 0 {
		p.Kills += kills[0]
	} else {
		p.Kills++
	}
}

// RemoveKill decreases the kill count
// Fails silently if kills are zero.
func (p *Player) RemoveKill() {
	if p.Kills == 0 {
		return
	}
	p.Kills--
}

// AddSelf increases the self count, decreases the kill, and gives a shot
func (p *Player) AddSelf() {
	p.Self++
	p.RemoveKill()
	p.AddShot()
}

// RemoveSelf decreases the self count and a shot
// Fails silently if selfs are zero.
func (p *Player) RemoveSelf() {
	if p.Self == 0 {
		return
	}
	p.Self--
	p.AddKill()
	p.RemoveShot()
}

// AddExplosion increases the explosion count, the kill count and gives a shot
func (p *Player) AddExplosion() {
	p.Explosions++
	p.AddShot()
	p.AddKill()
}

// RemoveExplosion decreases the explosion count, a shot and a kill
// Fails silently if Explosions are zero.
func (p *Player) RemoveExplosion() {
	if p.Explosions == 0 {
		return
	}
	p.Explosions--
	p.RemoveShot()
	p.RemoveKill()
}

// Reset resets the stats on a Player to 0
//
// It is to be run in Match.Start()
func (p *Player) Reset() {
	p.Shots = 0
	p.Sweeps = 0
	p.Kills = 0
	p.Self = 0
	p.Explosions = 0
	p.Matches = 0
}

// Update updates a player with the scores of another
//
// This is primarily used by the tournament score calculator
func (p *Player) Update(other Player) {
	p.Shots += other.Shots
	p.Sweeps += other.Sweeps
	p.Kills += other.Kills
	p.Self += other.Self
	p.Explosions += other.Explosions
	p.TotalScore = p.Score()

	// Every call to this method is per match. Count every call
	// as if a match.
	p.Matches++
	// log.Printf("Updated player: %d, %d", p.TotalScore, p.Matches)
}

// HTML renders the HTML of a player
func (p *Player) HTML() (out string) {
	return
}

// ByScore is a sort.Interface that sorts players by their score
type ByScore []Player

func (s ByScore) Len() int {
	return len(s)

}
func (s ByScore) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]

}
func (s ByScore) Less(i, j int) bool {
	// Technically not Less, but we want biggest first...
	return s[i].Score() > s[j].Score()
}

// SortByScore returns a list in order of the score the players have
func SortByScore(ps []Player) []Player {
	tmp := make([]Player, len(ps))
	copy(tmp, ps)
	sort.Sort(ByScore(tmp))
	return tmp
}

// ByKills is a sort.Interface that sorts players by their kills
type ByKills []Player

func (s ByKills) Len() int {
	return len(s)

}
func (s ByKills) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]

}
func (s ByKills) Less(i, j int) bool {
	// Technically not Less, but we want biggest first...
	return s[i].Kills > s[j].Kills
}

// SortByKills returns a list in order of the kills the players have
func SortByKills(ps []Player) []Player {
	tmp := make([]Player, len(ps))
	copy(tmp, ps)
	sort.Sort(ByKills(tmp))
	return tmp
}

// ByRunnerup is a sort.Interface that sorts players by their runnerup status
//
// This is almost exactly the same as score, but the number of matches a player
// has played factors in, and players that have played less matches are sorted
// favorably.
type ByRunnerup []Player

func (s ByRunnerup) Len() int {
	return len(s)

}
func (s ByRunnerup) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]

}
func (s ByRunnerup) Less(i, j int) bool {
	if s[i].Matches == s[j].Matches {
		// Same as by kills
		return s[i].Score() > s[j].Score()
	}
	// Lower is better - the ones that have not played should be at the top
	return s[i].Matches < s[j].Matches
}

// SortByRunnerup returns a list in order of the kills the players have
func SortByRunnerup(ps []Player) []Player {
	sort.Sort(ByRunnerup(ps))
	return ps
}

// Random returns a random color from the ColorList
func (c ColorList) Random() string {
	return c[rand.Intn(len(c))]
}

// Available returns a ColorList with the colors not used in a match
func (c ColorList) Available(m *Match) ColorList {
	ret := make(ColorList, 0)
	matchcolors := make(ColorList, 0)

	// Make a list with all the colors that are in use
	for _, p := range m.Players {
		matchcolors = append(matchcolors, p.PreferredColor())
	}

	// Fill the return list with the other colors
	for _, color := range c {
		for _, mc := range matchcolors {
			if color != mc {
				ret = append(ret, color)
			}
		}
	}

	return ret
}

// Judge is a Participant that has access to the judge functions
type Judge struct {
}
