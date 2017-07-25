package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// This makes sure that the test output just has the filenames, making it
	// easier for tools that parses it to find where the log output happened.
	log.SetFlags(log.Lshortfile)
	os.Exit(m.Run())
}

// testTournament makes a test tournament with `count` players.
func testTournament(count int) (t *Tournament) {
	s := strconv.Itoa(count)
	server := MockServer()
	t, err := NewTournament("Tournament "+s, s,
		time.Now().Add(time.Hour), nil, server)
	if err != nil {
		log.Fatal("tournament creation failed")
	}

	for i := 1; i <= count; i++ {
		p := testPerson(i)
		err := t.AddPlayer(NewPlayer(p))
		if err != nil {
			log.Fatal(err)
		}

		// XXX: If we don't add the person to the database anything that tries to
		// grab from it will fail. Backfilling from the semis is one of those
		// cases. This should be refactored away and this should be removed.
		t.db.SavePerson(p)
	}

	return
}

func endTryouts(t *Tournament) error {
	for x := range t.Tryouts {
		if err := t.Tryouts[x].Start(nil); err != nil {
			return err
		}
		if err := t.Tryouts[x].End(nil); err != nil {
			return err
		}
	}
	return nil
}

func endSemis(t *Tournament) error {
	for x := range t.Semis {
		if err := t.Semis[x].Start(nil); err != nil {
			return err
		}
		if err := t.Semis[x].End(nil); err != nil {
			return err
		}
	}
	return nil
}

func TestStartingTournamentWithFewerThan8PlayersFail(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(7)
	err := tm.StartTournament(nil)
	assert.NotNil(err)
}

func TestStartingTournamentWith8PlayersWorks(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(8)
	err := tm.StartTournament(nil)
	assert.Nil(err)
}

// func TestStartingTournamentWithMoreThan32PlayersFail(t *testing.T) {
// 	assert := assert.New(t)
// 	tm := testTournament(33)
// 	err := tm.StartTournament(nil)
// 	assert.NotNil(err)
// }

func TestStartingTournamentWith24PlayersWorks(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(24)
	err := tm.StartTournament(nil)
	assert.Nil(err)
}

// TestStartingGivesTheRightAmountOfTryouts
func TestStartingGivesTheRightAmountOfTryouts(t *testing.T) {
	assert := assert.New(t)
	for x := 8; x <= 32; x++ {
		tm := testTournament(x)
		err := tm.StartTournament(nil)
		assert.Nil(err)

		if x <= 16 {
			// Until 16 players, it's just always 4 matches
			assert.Equal(4, len(tm.Tryouts))
		} else {
			// But if there's more, then we need to check that only the needed
			// amount of matches are added.
			// The -1 in the calculation is to balance it out. I can't really put it
			// into words, but it works exactly as intended. The logic in
			// Tournament.StartTournament() is much more straightforward, luckily.
			y := (x-16-1)/4 + 1

			assert.Equal(
				4+y,
				len(tm.Tryouts),
				fmt.Sprintf("%d player tournament had %d matches, not %d", x, len(tm.Tryouts), 4+y),
			)
		}
	}

}

func TestStartingTournamentSetsStartedTimestamp(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(8)

	tm.StartTournament(nil)
	assert.NotNil(tm.Started)
}

func TestStartingTournamentCreatesTenEvents(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(8)

	tm.StartTournament(nil)
	assert.Equal(1+8+1, len(tm.Events))
	assert.Equal(tm.Events[0].Kind, "new_tournament")
	assert.Equal(tm.Events[1].Kind, "player_join")
	assert.Equal(tm.Events[9].Kind, "start")
}

func TestPopulateMatchesPopulatesTryoutsFor8Players(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(8)
	tm.StartTournament(nil)

	assert.Equal(4, len(tm.Tryouts[0].Players))
	assert.Equal(4, len(tm.Tryouts[1].Players))
}

func TestPopulateMatchesPopulatesAllMatchesFor24Players(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(24)
	tm.StartTournament(nil)

	assert.Equal(4, len(tm.Tryouts[0].Players))
	assert.Equal(4, len(tm.Tryouts[1].Players))
	assert.Equal(4, len(tm.Tryouts[2].Players))
	assert.Equal(4, len(tm.Tryouts[3].Players))
	assert.Equal(4, len(tm.Tryouts[4].Players))
	assert.Equal(4, len(tm.Tryouts[5].Players))
}

func TestRunnerupInsertion(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(23)
	tm.StartTournament(nil)

	m, err := tm.NextMatch()
	assert.Nil(err)
	m.Start(nil)
	m.Players[0].AddKills(10)
	m.End(nil)

	assert.Equal(m.Players[1].Person.ID, tm.Runnerups[0].ID)
	assert.Equal(m.Players[2].Person.ID, tm.Runnerups[1].ID)
	assert.Equal(m.Players[3].Person.ID, tm.Runnerups[2].ID)
}

func TestNextMatchNoMatchesAreStartedWithTryouts(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(16)
	tm.StartTournament(nil)

	m, err := tm.NextMatch()
	assert.Nil(err)
	assert.Equal(0, m.Index)
	assert.Equal("tryout", m.Kind)

	m.Start(nil)
	m.End(nil)

	m, err = tm.NextMatch()
	assert.Nil(err)
	assert.Equal(1, m.Index)
	assert.Equal("tryout", m.Kind)
	assert.Equal(CurrentMatch{"tryout", 1}, tm.Current)
}

func TestNextMatchNoMatchesAreStartedWithTryoutsDone(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(16)
	tm.StartTournament(nil)
	err := endTryouts(tm)
	assert.Nil(err)

	m, err := tm.NextMatch()
	assert.Nil(err)
	assert.Equal(0, m.Index)
	assert.Equal("semi", m.Kind)
	assert.Equal(CurrentMatch{"semi", 0}, tm.Current)
}

func TestNextMatchNoMatchesAreStartedWithTryoutsAndSemisDone(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(16)
	tm.StartTournament(nil)
	endTryouts(tm)
	endSemis(tm)

	m, err := tm.NextMatch()
	assert.Nil(err)
	assert.Equal(0, m.Index)
	assert.Equal("final", m.Kind)
	assert.Equal(CurrentMatch{"final", 0}, tm.Current)
}

func TestNextMatchEverythingDone(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(16)
	tm.StartTournament(nil)
	endTryouts(tm)
	endSemis(tm)
	tm.Final.Start(nil)
	tm.Final.End(nil)

	_, err := tm.NextMatch()
	assert.NotNil(err)
}

func TestUpdatePlayer(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(11)
	tm.StartTournament(nil)
	m, err := tm.NextMatch()
	assert.Nil(err)

	m.Start(nil)

	t.Log(m)
	m.Players[0].AddKills(5)
	m.Players[1].AddKills(6)
	m.Players[2].AddKills(7)
	m.Players[3].AddKills(10)

	m.End(nil) // Calls tm.UpdatePlayers()

	p, err := tm.getTournamentPlayerObject(m.Players[3].Person)
	assert.Nil(err)
	assert.Equal(10, p.Kills)
	p, err = tm.getTournamentPlayerObject(m.Players[2].Person)
	assert.Nil(err)
	assert.Equal(7, p.Kills)
	p, err = tm.getTournamentPlayerObject(m.Players[1].Person)
	assert.Nil(err)
	assert.Equal(6, p.Kills)
	p, err = tm.getTournamentPlayerObject(m.Players[0].Person)
	assert.Nil(err)
	assert.Equal(5, p.Kills)
}

func TestEnd4MatchTryoutsPlacesWinnerAndSecondIntoSemisAndRestIntoRunnerups(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(16)
	tm.StartTournament(nil)
	m, err := tm.NextMatch()
	assert.Nil(err)

	m.Start(nil)

	m.Players[0].AddKills(5)
	m.Players[1].AddKills(6)
	m.Players[2].AddKills(7)
	m.Players[3].AddKills(10)
	winner := m.Players[3].Name()
	silver := m.Players[2].Name()

	m.End(nil)

	assert.Equal(1, len(tm.Semis[0].Players))
	assert.Equal(1, len(tm.Semis[1].Players))
	assert.Equal(2, len(tm.Runnerups))

	assert.Equal(winner, tm.Semis[0].Players[0].Name())
	assert.Equal(silver, tm.Semis[1].Players[0].Name())
}

func TestEndComplete16PlayerTournamentKillsOnly(t *testing.T) {
	assert := assert.New(t)

	tm := testTournament(16)
	tm.StartTournament(nil)

	// Tryout 1 (same as test above)
	m, err := tm.NextMatch()
	assert.Nil(err)

	m.Start(nil)

	m.Players[0].AddKills(5)
	m.Players[1].AddKills(6)
	m.Players[2].AddKills(7)
	m.Players[3].AddKills(10)
	winner := m.Players[3].Name()
	silver := m.Players[2].Name()

	m.End(nil)

	assert.Equal(1, len(tm.Semis[0].Players))
	assert.Equal(1, len(tm.Semis[1].Players))
	assert.Equal(2, len(tm.Runnerups))

	assert.Equal(winner, tm.Semis[0].Players[0].Name())
	assert.Equal(silver, tm.Semis[1].Players[0].Name())

	// Tryout 2
	m2, err2 := tm.NextMatch()
	assert.Nil(err2)

	m2.Start(nil)

	m2.Players[0].AddKills(2)
	m2.Players[1].AddKills(10)
	m2.Players[2].AddKills(8)
	m2.Players[3].AddKills(4)
	winner2 := m2.Players[1].Name()
	silver2 := m2.Players[2].Name()

	m2.End(nil)

	assert.Equal(2, len(tm.Semis[0].Players))
	assert.Equal(2, len(tm.Semis[1].Players))
	assert.Equal(4, len(tm.Runnerups))

	assert.Equal(winner2, tm.Semis[1].Players[1].Name())
	assert.Equal(silver2, tm.Semis[0].Players[1].Name())

	// Tryout 3
	m3, err3 := tm.NextMatch()
	assert.Nil(err3)

	m3.Start(nil)

	m3.Players[0].AddKills(10)
	m3.Players[1].AddKills(3)
	m3.Players[2].AddKills(3)
	m3.Players[3].AddKills(5)
	winner3 := m3.Players[0].Name()
	silver3 := m3.Players[3].Name()

	m3.End(nil)

	assert.Equal(3, len(tm.Semis[0].Players))
	assert.Equal(3, len(tm.Semis[1].Players))
	assert.Equal(6, len(tm.Runnerups))

	assert.Equal(winner3, tm.Semis[0].Players[2].Name())
	assert.Equal(silver3, tm.Semis[1].Players[2].Name())

	// Tryout 4
	m4, err4 := tm.NextMatch()
	assert.Nil(err4)

	m4.Start(nil)

	m4.Players[0].AddKills(9)
	m4.Players[1].AddKills(10)
	m4.Players[2].AddKills(5)
	m4.Players[3].AddKills(5)
	winner4 := m4.Players[1].Name()
	silver4 := m4.Players[0].Name()

	m4.End(nil)

	assert.Equal(4, len(tm.Semis[0].Players))
	assert.Equal(4, len(tm.Semis[1].Players))
	assert.Equal(8, len(tm.Runnerups))

	assert.Equal(winner4, tm.Semis[1].Players[3].Name())
	assert.Equal(silver4, tm.Semis[0].Players[3].Name())

	// Semi 1
	s1, serr1 := tm.NextMatch()
	assert.Nil(serr1)

	assert.Equal("semi", s1.Kind)

	s1.Start(nil)

	s1.Players[0].AddKills(10)
	s1.Players[1].AddKills(7)
	s1.Players[2].AddKills(9)
	s1.Players[3].AddKills(8)
	winners1 := s1.Players[0].Name()
	silvers1 := s1.Players[2].Name()

	s1.End(nil)

	assert.Equal(2, len(tm.Final.Players))

	assert.Equal(winners1, tm.Final.Players[0].Name())
	assert.Equal(silvers1, tm.Final.Players[1].Name())

	// Semi 2
	s2, serr2 := tm.NextMatch()
	assert.Nil(serr2)

	assert.Equal("semi", s2.Kind)

	s2.Start(nil)

	s2.Players[0].AddKills(8)
	s2.Players[1].AddKills(10)
	s2.Players[2].AddKills(8)
	s2.Players[3].AddKills(9)
	winners2 := s2.Players[1].Name()
	silvers2 := s2.Players[3].Name()

	s2.End(nil)

	assert.Equal(4, len(tm.Final.Players))

	assert.Equal(winners2, tm.Final.Players[2].Name())
	assert.Equal(silvers2, tm.Final.Players[3].Name())

	// Final!
	f, ferr := tm.NextMatch()
	assert.Nil(ferr)

	assert.Equal("final", f.Kind)

	f.Start(nil)

	f.Players[0].AddKills(7)
	f.Players[1].AddKills(2)
	f.Players[2].AddKills(9)
	f.Players[3].AddKills(10)
	gold := f.Players[3].Name()
	lowe := f.Players[2].Name()
	bronze := f.Players[0].Name()

	f.End(nil)

	assert.Equal(gold, tm.Winners[0].Name())
	assert.Equal(lowe, tm.Winners[1].Name())
	assert.Equal(bronze, tm.Winners[2].Name())
}

func TestEndComplete19PlayerTournamentKillsOnly(t *testing.T) {
	// This primarily tests the runnerup population for the fifth match
	// and that only the winners are propagated when there are more
	// than 16 players.
	assert := assert.New(t)

	tm := testTournament(19)
	tm.StartTournament(nil)

	assert.Equal(5, len(tm.Tryouts))

	// Tryout 1
	m, err := tm.NextMatch()
	assert.Nil(err)

	m.Start(nil)

	m.Players[0].AddKills(5)
	m.Players[1].AddKills(6)
	m.Players[2].AddKills(7)
	m.Players[3].AddKills(10)
	winner := m.Players[3].Name()

	m.End(nil)

	assert.Equal(1, len(tm.Semis[0].Players))
	assert.Equal(0, len(tm.Semis[1].Players))
	assert.Equal(3, len(tm.Runnerups))

	assert.Equal(winner, tm.Semis[0].Players[0].Name())

	// Tryout 2
	m2, err2 := tm.NextMatch()
	assert.Nil(err2)

	m2.Start(nil)

	m2.Players[0].AddKills(2)
	m2.Players[1].AddKills(10)
	m2.Players[2].AddKills(8)
	m2.Players[3].AddKills(4)
	winner2 := m2.Players[1].Name()

	m2.End(nil)

	assert.Equal(1, len(tm.Semis[0].Players))
	assert.Equal(1, len(tm.Semis[1].Players))
	assert.Equal(6, len(tm.Runnerups))

	assert.Equal(winner2, tm.Semis[1].Players[0].Name())

	// Tryout 3
	m3, err3 := tm.NextMatch()
	assert.Nil(err3)

	m3.Start(nil)

	m3.Players[0].AddKills(10)
	m3.Players[1].AddKills(3)
	m3.Players[2].AddKills(3)
	m3.Players[3].AddKills(5)
	winner3 := m3.Players[0].Name()

	m3.End(nil)

	assert.Equal(2, len(tm.Semis[0].Players))
	assert.Equal(1, len(tm.Semis[1].Players))
	assert.Equal(9, len(tm.Runnerups))

	assert.Equal(winner3, tm.Semis[0].Players[1].Name())

	// Tryout 4
	m4, err4 := tm.NextMatch()
	assert.Nil(err4)

	m4.Start(nil)

	m4.Players[0].AddKills(9)
	m4.Players[1].AddKills(10)
	m4.Players[2].AddKills(5)
	m4.Players[3].AddKills(5)
	winner4 := m4.Players[1].Name()

	m4.End(nil)

	assert.Equal(2, len(tm.Semis[0].Players))
	assert.Equal(2, len(tm.Semis[1].Players))
	assert.Equal(12, len(tm.Runnerups))

	assert.Equal(winner4, tm.Semis[1].Players[1].Name())

	// Tryout 5 / Runnerup 1
	m5, err5 := tm.NextMatch()
	assert.Nil(err5)
	assert.Equal("tryout", m5.Kind)

	m5.Start(nil)
	// Given the 19 player match, there are 3 players that have yet to contend
	// and therefore we need to pick one of the runnerups.
	assert.Equal(4, len(m5.Players))
	assert.Equal(12, len(tm.Runnerups))

	m5.Players[0].AddKills(8)
	m5.Players[1].AddKills(7)
	m5.Players[2].AddKills(2)
	m5.Players[3].AddKills(10)
	winner5 := m5.Players[3].Name()

	m5.End(nil)

	assert.Equal(winner5, tm.Semis[0].Players[2].Name())

	// We need to backfill the players, and since that is a judge action we need
	// to simulate that
	err = tm.BackfillSemis(nil, []string{
		tm.Runnerups[0].ID,
		tm.Runnerups[1].ID,
		tm.Runnerups[2].ID,
	})

	assert.Nil(err)

	assert.Equal(4, len(tm.Semis[0].Players))
	assert.Equal(4, len(tm.Semis[1].Players))
	assert.Equal(11, len(tm.Runnerups))

	// Semi 1
	s1, serr1 := tm.NextMatch()
	assert.Nil(serr1)

	assert.Equal("semi", s1.Kind)

	s1.Start(nil)

	s1.Players[0].AddKills(10)
	s1.Players[1].AddKills(7)
	s1.Players[2].AddKills(9)
	s1.Players[3].AddKills(8)
	winners1 := s1.Players[0].Name()
	silvers1 := s1.Players[2].Name()

	s1.End(nil)

	assert.Equal(2, len(tm.Final.Players))

	assert.Equal(winners1, tm.Final.Players[0].Name())
	assert.Equal(silvers1, tm.Final.Players[1].Name())

	// Semi 2
	s2, serr2 := tm.NextMatch()
	assert.Nil(serr2)

	assert.Equal("semi", s2.Kind)

	s2.Start(nil)

	s2.Players[0].AddKills(8)
	s2.Players[1].AddKills(10)
	s2.Players[2].AddKills(8)
	s2.Players[3].AddKills(9)
	winners2 := s2.Players[1].Name()
	silvers2 := s2.Players[3].Name()

	s2.End(nil)

	assert.Equal(4, len(tm.Final.Players))

	assert.Equal(winners2, tm.Final.Players[2].Name())
	assert.Equal(silvers2, tm.Final.Players[3].Name())

	// Final!
	f, ferr := tm.NextMatch()
	assert.Nil(ferr)

	assert.Equal("final", f.Kind)

	f.Start(nil)

	f.Players[0].AddKills(7)
	f.Players[1].AddKills(2)
	f.Players[2].AddKills(9)
	f.Players[3].AddKills(10)

	gold := f.Players[3].Name()
	lowe := f.Players[2].Name()
	bronze := f.Players[0].Name()

	f.End(nil)

	assert.Equal(gold, tm.Winners[0].Name())
	assert.Equal(lowe, tm.Winners[1].Name())
	assert.Equal(bronze, tm.Winners[2].Name())
}
