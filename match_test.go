package main

import (
	"fmt"
	"github.com/drunkenfall/faking"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"testing"
	"time"
)

// MockMatch makes a mock Match{} with a dummy Tournament{}
func MockMatch(idx int, cat string) (m *Match) {
	tm := testTournament(8)
	tm.SetMatchPointers()
	err := tm.StartTournament()
	if err != nil {
		log.Fatal(err)
	}

	switch cat {
	case "tryout":
		m = tm.Tryouts[idx]
	case "semi":
		m = tm.Semis[idx]
	case "final":
		m = tm.Final
	default:
		log.Fatalf("Unknown match type: %s", cat)
	}

	return m
}

func testPlayer() Player {
	return *NewPlayer(testPerson(rand.Int()))
}

func testPerson(i int) *Person {
	return &Person{
		ID:   fmt.Sprintf("%d: %s", i, faking.FakeName()),
		Name: fmt.Sprintf("%d: %s", i, faking.FakeName()),
		Nick: faking.FakeNick(),
		ColorPreference: []string{
			RandomColor(Colors),
			RandomColor(Colors),
		},
	}
}

func TestAddPlayer(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(0, "tryout")
	m.Players = []Player{}
	p := testPlayer()

	err := m.AddPlayer(p)
	assert.Nil(err)

	assert.Equal(1, len(m.Players))
}

func TestAddFifthPlayer(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(1, "tryout")

	p := testPlayer()

	err := m.AddPlayer(p)
	assert.NotNil(err)
	assert.Equal(4, len(m.Players))
}

func TestStartAlreadyStartedMatch(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(1, "tryout")
	m.Started = time.Now()

	err := m.Start()
	assert.NotNil(err)
}

func TestStart(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(1, "tryout")

	err := m.Start()
	assert.Nil(err)
	assert.Equal(false, m.Started.IsZero())
}

func TestEndGivesShotToWinner(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(1, "tryout")

	err := m.Start()
	assert.Nil(err)
	m.Players[2].AddKills(10)
	m.KillOrder = m.MakeKillOrder()

	err = m.End()
	assert.Nil(err)
	assert.Equal(1, m.Players[2].Shots)
}

func TestEndAlreadyEndedMatch(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(1, "tryout")
	m.Ended = time.Now()

	err := m.End()
	assert.NotNil(err)
}

func TestCommitSweepPlayer1(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "tryout")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := Round{
		Kills: [][]int{
			{3, 0},
			{0, 0},
			{0, 0},
			{0, 0},
		},
		Shots: []bool{
			false,
			false,
			false,
			false,
		},
	}

	m.Commit(c)
	assert.Equal(1, m.Players[0].Sweeps)
}

func TestCommitDoubleKillPlayer2(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "tryout")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := Round{
		Kills: [][]int{
			{0, 0},
			{2, 0},
			{0, 0},
			{0, 0},
		},
		Shots: []bool{
			false,
			false,
			false,
			false,
		},
	}

	m.Commit(c)
	assert.Equal(2, m.Players[1].Kills)
}

func TestCommitSweepAndSuicidePlayer3(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "tryout")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := Round{
		Kills: [][]int{
			{0, 0},
			{0, 0},
			{3, -1},
			{0, 0},
		},
		Shots: []bool{
			false,
			false,
			true,
			false,
		},
	}

	m.Commit(c)
	assert.Equal(1, m.Players[2].Sweeps)
	assert.Equal(2, m.Players[2].Kills)
	assert.Equal(1, m.Players[2].Shots)
}

func TestCommitSuicidePlayer4(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "tryout")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := Round{
		Kills: [][]int{
			{0, 0},
			{0, 0},
			{0, 0},
			{0, -1},
		},
		Shots: []bool{
			false,
			false,
			false,
			false,
		},
	}

	m.Commit(c)
	assert.Equal(1, m.Players[3].Self)
	assert.Equal(1, m.Players[3].Shots)
}

func TestCommitShotsForPlayer2and3(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "tryout")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := Round{
		Kills: [][]int{
			{0, 0},
			{0, 0},
			{0, 0},
			{0, 0},
		},
		Shots: []bool{
			false,
			true,
			true,
			false,
		},
	}

	m.Commit(c)
	assert.Equal(1, m.Players[1].Shots)
	assert.Equal(1, m.Players[2].Shots)
}

func TestCommitSweepForPlayer1(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "tryout")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := Round{
		Kills: [][]int{
			{3, 0},
			{0, 0},
			{0, 0},
			{0, 0},
		},
		// For the frontend it makes sense that a sweep marks a shot, therefore we
		// need to make sure that we don't add another shot.
		Shots: []bool{
			true,
			false,
			false,
			false,
		},
	}

	m.Commit(c)
	assert.Equal(3, m.Players[0].Kills)
	assert.Equal(1, m.Players[0].Shots)
}

func TestCommitStoredOnMatch(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "tryout")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := Round{
		Kills: [][]int{
			{0, 0},
			{1, 0},
			{1, 0},
			{1, 0},
		},
		// For the frontend it makes sense that a sweep marks a shot, therefore we
		// need to make sure that we don't add another shot.
		Shots: []bool{
			true,
			false,
			false,
			false,
		},
	}

	assert.Equal(0, len(m.Rounds))
	m.Commit(c)
	assert.Equal(1, len(m.Rounds))
	assert.Equal(1, m.Rounds[0].Kills[1][0])
	assert.Equal(1, m.Rounds[0].Kills[2][0])
	assert.Equal(1, m.Rounds[0].Kills[3][0])
	assert.Equal(true, m.Rounds[0].Shots[0])
}

func TestCorrectColorConflictsNoScores(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "final")
	m.Players = make([]Player, 0)

	m.Tournament.Players[0].PreferredColor = "green"
	m.Tournament.Players[1].PreferredColor = "green"
	m.Tournament.Players[2].PreferredColor = "blue"
	m.Tournament.Players[3].PreferredColor = "red"

	assert.Nil(m.AddPlayer(m.Tournament.Players[0]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[1]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[2]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[3]))

	assert.Equal("green", m.Players[0].Color)
	assert.Equal("green", m.Players[0].PreferredColor)
	assert.NotEqual("green", m.Players[1].Color)
	assert.Equal("green", m.Players[1].PreferredColor)
	assert.Equal("blue", m.Players[2].Color)
	assert.Equal("blue", m.Players[2].PreferredColor)
	assert.Equal("red", m.Players[3].Color)
	assert.Equal("red", m.Players[3].PreferredColor)
}

func TestCorrectColorConflictsUserLevels(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "final")
	m.Players = make([]Player, 0)

	m.Tournament.Players[0].PreferredColor = "green"
	m.Tournament.Players[0].Person.Userlevel = 10000
	m.Tournament.Players[1].PreferredColor = "green"
	m.Tournament.Players[2].PreferredColor = "red"
	m.Tournament.Players[2].Person.Userlevel = -10000
	m.Tournament.Players[3].PreferredColor = "red"

	assert.Nil(m.AddPlayer(m.Tournament.Players[0]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[1]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[2]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[3]))

	assert.Equal("green", m.Players[0].Color)
	assert.Equal("green", m.Players[0].PreferredColor)
	assert.NotEqual("green", m.Players[1].Color)
	assert.Equal("green", m.Players[1].PreferredColor)
	assert.NotEqual("red", m.Players[2].Color)
	assert.Equal("red", m.Players[2].PreferredColor)
	assert.Equal("red", m.Players[3].Color)
	assert.Equal("red", m.Players[3].PreferredColor)
}

// This test was needed since somehow the color were being kept
func TestCorrectColorConflictsResetsToPreferredColor(t *testing.T) {
	assert := assert.New(t)

	tm := testTournament(12)
	m := tm.Tryouts[0]
	m2 := tm.Tryouts[1]
	m.Players = make([]Player, 0)
	m2.Players = make([]Player, 0)

	m.Tournament.Players[0].PreferredColor = "green"
	m.Tournament.Players[1].PreferredColor = "green"
	m.Tournament.Players[2].PreferredColor = "green"
	m.Tournament.Players[3].PreferredColor = "green"
	m.Tournament.Players[4].PreferredColor = "green"

	assert.Nil(m.AddPlayer(m.Tournament.Players[0]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[1]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[2]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[3]))

	assert.Nil(m2.AddPlayer(m.Tournament.Players[4]))
	assert.Nil(m2.AddPlayer(m.Tournament.Players[1]))
	assert.Nil(m2.AddPlayer(m.Tournament.Players[2]))
	assert.Nil(m2.AddPlayer(m.Tournament.Players[3]))

	assert.Equal("green", m.Players[0].Color)
	assert.Equal("green", m.Players[0].PreferredColor)
	assert.NotEqual("green", m.Players[1].Color)
	assert.Equal("green", m.Players[1].PreferredColor)
	assert.NotEqual("green", m.Players[2].Color)
	assert.Equal("green", m.Players[2].PreferredColor)
	assert.NotEqual("green", m.Players[3].Color)

	assert.Equal("green", m2.Players[0].Color)
	assert.Equal("green", m2.Players[0].PreferredColor)
	assert.NotEqual("green", m2.Players[1].Color)
	assert.Equal("green", m2.Players[1].PreferredColor)
	assert.NotEqual("green", m2.Players[2].Color)
	assert.Equal("green", m2.Players[2].PreferredColor)
	assert.NotEqual("green", m2.Players[3].Color)
}

func TestCorrectColorConflictsNoScoresDoubleConflict(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "final")
	m.Players = make([]Player, 0)

	m.Tournament.Players[0].PreferredColor = "green"
	m.Tournament.Players[1].PreferredColor = "green"
	m.Tournament.Players[2].PreferredColor = "blue"
	m.Tournament.Players[3].PreferredColor = "blue"

	assert.Nil(m.AddPlayer(m.Tournament.Players[0]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[1]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[2]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[3]))

	assert.Equal("green", m.Players[0].Color)
	assert.Equal("green", m.Players[0].PreferredColor)
	assert.NotEqual("green", m.Players[1].Color)
	assert.Equal("green", m.Players[1].PreferredColor)
	assert.Equal("blue", m.Players[2].Color)
	assert.Equal("blue", m.Players[2].PreferredColor)
	assert.NotEqual("blue", m.Players[3].Color)
	assert.Equal("blue", m.Players[3].PreferredColor)
}

func TestCorrectColorConflictsWithScoresDoubleConflict(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "final")
	m.Players = make([]Player, 0)

	m.Tournament.Players[0].PreferredColor = "green"
	m.Tournament.Players[0].Person.Nick = "GreenCorrected"

	m.Tournament.Players[1].Kills = 3
	m.Tournament.Players[1].PreferredColor = "green"

	m.Tournament.Players[2].PreferredColor = "blue"
	m.Tournament.Players[2].Person.Nick = "BlueCorrected"

	m.Tournament.Players[3].Kills = 3
	m.Tournament.Players[3].PreferredColor = "blue"

	assert.Nil(m.AddPlayer(m.Tournament.Players[0]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[1]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[2]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[3]))

	assert.NotEqual("green", m.Players[0].Color)
	assert.Equal("green", m.Players[0].PreferredColor)
	assert.Equal("green", m.Players[1].Color)
	assert.Equal("green", m.Players[1].PreferredColor)
	assert.NotEqual("blue", m.Players[2].Color)
	assert.Equal("blue", m.Players[2].PreferredColor)
	assert.Equal("blue", m.Players[3].Color)
	assert.Equal("blue", m.Players[3].PreferredColor)
}

func TestCorrectColorConflictsWithScoresTripleConflict(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "final")
	m.Players = make([]Player, 0)

	m.Tournament.Players[0].PreferredColor = "green"
	m.Tournament.Players[0].Person.Nick = "Green1Corrected"

	m.Tournament.Players[1].Kills = 3
	m.Tournament.Players[1].PreferredColor = "green"
	m.Tournament.Players[1].Person.Nick = "Green2Corrected"

	m.Tournament.Players[2].PreferredColor = "blue"
	m.Tournament.Players[2].Person.Nick = "BlueKeep"

	m.Tournament.Players[3].Kills = 10
	m.Tournament.Players[3].PreferredColor = "green"
	m.Tournament.Players[3].Person.Nick = "GreenKeep"

	assert.Nil(m.AddPlayer(m.Tournament.Players[0]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[1]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[2]))
	assert.Nil(m.AddPlayer(m.Tournament.Players[3]))

	assert.NotEqual("green", m.Players[0].Color)
	assert.NotEqual("green", m.Players[1].Color)
	assert.Equal("blue", m.Players[2].Color)
	assert.Equal("green", m.Players[3].Color)
}

func TestMakeKillOrder(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(0, "tryout")

	m.Players[0].AddKills(1)
	m.Players[1].AddKills(4)
	m.Players[2].AddKills(5)
	m.Players[3].AddKills(10)

	ko := m.MakeKillOrder()

	// As long as the order is reversed, this test is proven.
	// ...just like above. <3
	assert.Equal(ko[0], 3)
	assert.Equal(ko[1], 2)
	assert.Equal(ko[2], 1)
	assert.Equal(ko[3], 0)
}
