package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScoreWithShots(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()
	p.AddShot()

	assert.Equal(3, p.Score())
	p.AddShot()
	assert.Equal(6, p.Score())
}

func TestScoreWithSweeps(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()
	p.AddSweep()

	assert.Equal(14, p.Score())
	p.AddSweep()
	assert.Equal(28, p.Score())
}

func TestScoreWithKills(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()
	p.AddKill()

	assert.Equal(2, p.Score())
	p.AddKill()
	assert.Equal(4, p.Score())
}

func TestScoreWithSelfs(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()
	p.AddSelf()

	assert.Equal(4, p.Score())
	p.AddSelf()
	assert.Equal(8, p.Score())
}

func TestScoreWithExplosions(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()
	p.AddExplosion()

	assert.Equal(6, p.Score())
	p.AddExplosion()
	assert.Equal(12, p.Score())
}

func TestScoreWithAll(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()
	p.AddShot()
	p.AddSweep()
	p.AddKill()
	p.AddSelf()
	p.AddExplosion()
	assert.Equal(27, p.Score())
}

func TestAddShot(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()

	p.AddShot()
	assert.Equal(1, p.Shots)
	p.AddShot()
	assert.Equal(2, p.Shots)
}

func TestRemoveShot(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()
	p.Shots = 1

	p.RemoveShot()
	assert.Equal(0, p.Shots)
	p.RemoveShot()
	assert.Equal(0, p.Shots)
}

func TestAddSweep(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()

	p.AddSweep()
	assert.Equal(1, p.Sweeps)
	assert.Equal(1, p.Shots)
	assert.Equal(3, p.Kills)

	p.AddSweep()
	assert.Equal(2, p.Sweeps)
	assert.Equal(2, p.Shots)
	assert.Equal(6, p.Kills)
}

func TestRemoveSweep(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()
	p.Sweeps = 1
	p.Shots = 1
	p.Kills = 3

	p.RemoveSweep()
	assert.Equal(0, p.Sweeps)
	assert.Equal(0, p.Shots)
	assert.Equal(0, p.Kills)
	p.RemoveSweep()
	assert.Equal(0, p.Sweeps)
	assert.Equal(0, p.Shots)
	assert.Equal(0, p.Kills)
}

func TestAddKill(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()

	p.AddKill()
	assert.Equal(1, p.Kills)
	p.AddKill()
	assert.Equal(2, p.Kills)
}

func TestRemoveKill(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()
	p.Kills = 1

	p.RemoveKill()
	assert.Equal(0, p.Kills)
	p.RemoveKill()
	assert.Equal(0, p.Kills)
}

func TestAddSelf(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()

	p.AddSelf()
	assert.Equal(1, p.Self)
	assert.Equal(1, p.Shots)
	p.AddSelf()
	assert.Equal(2, p.Self)
	assert.Equal(2, p.Shots)
}

func TestRemoveSelf(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()
	p.Self = 1
	p.Shots = 1

	p.RemoveSelf()
	assert.Equal(0, p.Self)
	assert.Equal(0, p.Shots)
	p.RemoveSelf()
	assert.Equal(0, p.Self)
	assert.Equal(0, p.Shots)
}

func TestAddExplosion(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()

	p.AddExplosion()
	assert.Equal(1, p.Explosions)
	assert.Equal(1, p.Shots)
	assert.Equal(1, p.Kills)
	p.AddExplosion()
	assert.Equal(2, p.Explosions)
	assert.Equal(2, p.Shots)
	assert.Equal(2, p.Kills)
}

func TestRemoveExplosion(t *testing.T) {
	assert := assert.New(t)
	p := testPlayer()
	p.Explosions = 1

	p.RemoveExplosion()
	assert.Equal(0, p.Explosions)
	p.RemoveExplosion()
	assert.Equal(0, p.Explosions)
}

// Same number of kills, but more pints for player p2.
func TestSortTiedPlayersByKills(t *testing.T) {
	assert := assert.New(t)
	p1 := testPlayer()
	p2 := testPlayer()

	p2.AddShot()

	ps := []Player{p1, p2}
	ret := SortByKills(ps)
	assert.Equal(ret[0], p2)
	assert.Equal(ret[1], p1)
}

func TestSortRunnerups(t *testing.T) {
	assert := assert.New(t)
	p1 := testPlayer() // 10 points, 1 match
	p2 := testPlayer() // 20 points, 2 Matches
	p3 := testPlayer() // 16 points, 1 match

	p1.Kills = 5
	p1.Matches = 1
	p1.Person.Nick = "second"
	p2.Kills = 10
	p2.Matches = 2
	p2.Person.Nick = "last"
	p3.Kills = 8
	p3.Matches = 1
	p3.Person.Nick = "first"

	ps := []Player{p1, p2, p3}
	ret := SortByRunnerup(ps)

	assert.Equal("first", ret[0].Name())
	assert.Equal("second", ret[1].Name())
	assert.Equal("last", ret[2].Name())
}
