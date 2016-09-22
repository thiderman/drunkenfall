package main

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

// MockMatch makes a mock Match{} with a dummy Tournament{}
func MockMatch(idx int, cat string) *Match {
	s := MockServer()
	tm, _ := NewTournament("test", "t", s)
	tm.SetMatchPointers()
	m := NewMatch(tm, idx, cat)
	m.Players = []Player{
		testPlayer(),
		testPlayer(),
		testPlayer(),
		testPlayer(),
	}
	return m
}

func testPlayer() Player {
	return NewPlayer(testPerson())
}

func testPerson() *Person {
	return &Person{
		ID:   FakeName(),
		Name: FakeName(),
		Nick: FakeNick(),
		ColorPreference: []string{
			Colors[rand.Intn(len(Colors))],
			Colors[rand.Intn(len(Colors))],
		},
	}
}

func TestAddPlayer(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(1, "test")
	m.Players = []Player{}
	p := testPlayer()

	err := m.AddPlayer(p)
	assert.Nil(err)

	assert.Equal(1, len(m.Players))
}

func TestAddFifthPlayer(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(1, "test")

	m.Players = []Player{
		testPlayer(),
		testPlayer(),
		testPlayer(),
		testPlayer(),
	}
	p := testPlayer()

	err := m.AddPlayer(p)
	assert.NotNil(err)
	assert.Equal(4, len(m.Players))
}

func TestStartAlreadyStartedMatch(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(1, "test")
	m.Started = time.Now()

	err := m.Start()
	assert.NotNil(err)
}

func TestStart(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(1, "test")
	m.Players = []Player{
		testPlayer(),
		testPlayer(),
		testPlayer(),
		testPlayer(),
	}

	err := m.Start()
	assert.Nil(err)
	assert.Equal(false, m.Started.IsZero())
}

func TestEndGivesShotToWinner(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(1, "tryout")
	// TODO(thiderman): This is terrible, but it works for now :(
	m.Tournament = nil
	m.Players = []Player{
		testPlayer(),
		testPlayer(),
		testPlayer(),
		testPlayer(),
	}

	err := m.Start()
	assert.Nil(err)
	m.Players[2].AddKill(10)
	m.ScoreOrder = m.MakeScoreOrder()

	err = m.End()
	assert.Nil(err)
	assert.Equal(1, m.Players[2].Shots)
}

func TestEndAlreadyEndedMatch(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(1, "test")
	m.Ended = time.Now()

	err := m.End()
	assert.NotNil(err)
}

func TestEnd(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(1, "test")

	err := m.End()
	assert.Nil(err)
	assert.Equal(false, m.Ended.IsZero())
}

func TestCommitSweepPlayer1(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "test")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := MatchCommit{
		Kills: [][]int{
			[]int{3, 0},
			[]int{0, 0},
			[]int{0, 0},
			[]int{0, 0},
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

	m := MockMatch(0, "test")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := MatchCommit{
		Kills: [][]int{
			[]int{0, 0},
			[]int{2, 0},
			[]int{0, 0},
			[]int{0, 0},
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

	m := MockMatch(0, "test")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := MatchCommit{
		Kills: [][]int{
			[]int{0, 0},
			[]int{0, 0},
			[]int{3, 1},
			[]int{0, 0},
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

	m := MockMatch(0, "test")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := MatchCommit{
		Kills: [][]int{
			[]int{0, 0},
			[]int{0, 0},
			[]int{0, 0},
			[]int{0, 1},
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

	m := MockMatch(0, "test")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := MatchCommit{
		Kills: [][]int{
			[]int{0, 0},
			[]int{0, 0},
			[]int{0, 0},
			[]int{0, 0},
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

	m := MockMatch(0, "test")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := MatchCommit{
		Kills: [][]int{
			[]int{3, 0},
			[]int{0, 0},
			[]int{0, 0},
			[]int{0, 0},
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

	m := MockMatch(0, "test")
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())
	_ = m.AddPlayer(testPlayer())

	c := MatchCommit{
		Kills: [][]int{
			[]int{0, 0},
			[]int{1, 0},
			[]int{1, 0},
			[]int{1, 0},
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

	assert.Equal(0, len(m.Commits))
	m.Commit(c)
	assert.Equal(1, len(m.Commits))
	assert.Equal(1, m.Commits[0].Kills[1][0])
	assert.Equal(1, m.Commits[0].Kills[2][0])
	assert.Equal(1, m.Commits[0].Kills[3][0])
	assert.Equal(true, m.Commits[0].Shots[0])
}

func TestCorrectColorConflictsNoScores(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "final")
	m.Players = make([]Player, 0)

	p1 := testPlayer()
	p1.Person.ColorPreference[0] = "green"
	p2 := testPlayer()
	p2.Person.ColorPreference[0] = "green"
	p3 := testPlayer()
	p3.Person.ColorPreference[0] = "blue"
	p4 := testPlayer()
	p4.Person.ColorPreference[0] = "red"

	_ = m.AddPlayer(p1)
	_ = m.AddPlayer(p2)
	_ = m.AddPlayer(p3)
	_ = m.AddPlayer(p4)

	assert.Equal("green", m.Players[0].Color)
	assert.Equal("green", m.Players[0].OriginalColor)
	assert.NotEqual("green", m.Players[1].Color)
	assert.Equal("green", m.Players[1].OriginalColor)
	assert.Equal("blue", m.Players[2].Color)
	assert.Equal("blue", m.Players[2].OriginalColor)
	assert.Equal("red", m.Players[3].Color)
	assert.Equal("red", m.Players[3].OriginalColor)
}

// This test was needed since somehow getting the color reset
func TestCorrectColorConflictsResetsToOriginalColor(t *testing.T) {
	assert := assert.New(t)

	m := MockMatch(0, "semi")
	m2 := MockMatch(0, "final")
	m.Players = make([]Player, 0)
	m2.Players = make([]Player, 0)

	p1 := testPlayer()
	p1.Person.ColorPreference[0] = "green"
	p2 := testPlayer()
	p2.Person.ColorPreference[0] = "green"
	p3 := testPlayer()
	p3.Person.ColorPreference[0] = "green"
	p4 := testPlayer()
	p4.Person.ColorPreference[0] = "green"
	p5 := testPlayer()
	p5.Person.ColorPreference[0] = "green"

	_ = m.AddPlayer(p1)
	_ = m.AddPlayer(p2)
	_ = m.AddPlayer(p3)
	_ = m.AddPlayer(p4)

	assert.Equal("green", m.Players[0].Color)
	assert.Equal("green", m.Players[0].OriginalColor)
	assert.NotEqual("green", m.Players[1].Color)
	assert.Equal("green", m.Players[1].OriginalColor)
	assert.NotEqual("green", m.Players[2].Color)
	assert.Equal("green", m.Players[2].OriginalColor)
	assert.NotEqual("green", m.Players[3].Color)
	assert.Equal("green", m.Players[3].OriginalColor)

	_ = m2.AddPlayer(p5)
	_ = m2.AddPlayer(p2)
	_ = m2.AddPlayer(p3)
	_ = m2.AddPlayer(p4)

	assert.Equal("green", m2.Players[0].Color)
	assert.Equal("green", m2.Players[0].OriginalColor)
	assert.NotEqual("green", m2.Players[1].Color)
	assert.Equal("green", m2.Players[1].OriginalColor)
	assert.NotEqual("green", m2.Players[2].Color)
	assert.Equal("green", m2.Players[2].OriginalColor)
	assert.NotEqual("green", m2.Players[3].Color)
	assert.Equal("green", m2.Players[3].OriginalColor)

}

// func TestCorrectColorConflictsNoScoresDoubleConflict(t *testing.T) {
// 	assert := assert.New(t)

// 	m := MockMatch(0, "final")
// 	_ = m.AddPlayer(testPlayer())
// 	_ = m.AddPlayer(testPlayer())
// 	_ = m.AddPlayer(testPlayer())
// 	_ = m.AddPlayer(testPlayer())

// 	err := m.CorrectColorConflicts()
// 	assert.Nil(err)

// 	assert.Equal("green", m.Players[0].PreferredColor())
// 	assert.NotEqual("green", m.Players[1].PreferredColor())
// 	assert.Equal("blue", m.Players[2].PreferredColor())
// 	assert.NotEqual("blue", m.Players[3].PreferredColor())
// }

// func TestCorrectColorConflictPlayerTwoHasHigherScore(t *testing.T) {
// 	assert := assert.New(t)

// 	m := MockMatch(0, "final")
// 	_ = m.AddPlayer(testPlayer())
// 	_ = m.AddPlayer(testPlayer())
// 	_ = m.AddPlayer(testPlayer())
// 	_ = m.AddPlayer(testPlayer())

// 	// Add some score to player 2 so that it has preference over green.
// 	m.Players[1].AddKill(3)

// 	err := m.CorrectColorConflicts()
// 	assert.Nil(err)

// 	// TODO(thiderman): Fix this
// 	// assert.NotEqual("green", m.Players[0].PreferredColor())
// 	// assert.Equal("green", m.Players[1].PreferredColor())
// 	// assert.Equal("blue", m.Players[2].PreferredColor())
// 	// assert.Equal("cyan", m.Players[3].PreferredColor())
// }

func TestSortByScoreDoesNotAlterSourceList(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(0, "test")

	m.Players[0].AddKill(1)
	m.Players[1].AddKill(4)
	m.Players[2].AddKill(5)
	m.Players[3].AddKill(10)

	ps := SortByScore(m.Players)

	// As long as the order is reversed, this test is proven.
	assert.Equal(ps[0].Name(), m.Players[3].Name())
	assert.Equal(ps[1].Name(), m.Players[2].Name())
	assert.Equal(ps[2].Name(), m.Players[1].Name())
	assert.Equal(ps[3].Name(), m.Players[0].Name())
}

func TestMakeScoreOrder(t *testing.T) {
	assert := assert.New(t)
	m := MockMatch(0, "test")

	m.Players[0].AddKill(1)
	m.Players[1].AddKill(4)
	m.Players[2].AddKill(5)
	m.Players[3].AddKill(10)

	so := m.MakeScoreOrder()

	// As long as the order is reversed, this test is proven.
	// ...just like above. <3
	assert.Equal(so[0], 3)
	assert.Equal(so[1], 2)
	assert.Equal(so[2], 1)
	assert.Equal(so[3], 0)
}
