package towerfall

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

func endPlayoffs(t *Tournament) error {
	for x := 0; x < len(t.Matches)-3; x++ {
		if err := t.Matches[x].Start(nil); err != nil {
			return err
		}
		if err := t.Matches[x].End(nil); err != nil {
			return err
		}
	}
	return nil
}

func endSemis(t *Tournament) error {
	offset := len(t.Matches) - 3
	if err := t.Matches[offset].Start(nil); err != nil {
		return err
	}
	if err := t.Matches[offset].End(nil); err != nil {
		return err
	}
	if err := t.Matches[offset+1].Start(nil); err != nil {
		return err
	}
	if err := t.Matches[offset+1].End(nil); err != nil {
		return err
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

func TestStartingTournamentWith24PlayersWorks(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(24)
	err := tm.StartTournament(nil)
	assert.Nil(err)
}

func TestDoubleStartIsForbidden(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(16)
	err := tm.StartTournament(nil)
	assert.Nil(err)
	err = tm.StartTournament(nil)
	assert.EqualError(err, "tournament is already running")
}

func TestStartingGivesTheRightAmountOfPlayoffs(t *testing.T) {
	assert := assert.New(t)
	for x := 8; x <= 32; x++ {
		tm := testTournament(x)
		err := tm.StartTournament(nil)
		assert.Nil(err)

		if x == 8 {
			// A special case - we don't need any playoffs since we're ready
			// for semi-finals right away.
			assert.Equal(3, len(tm.Matches))
			continue
		}

		// The -1 is to shift so that when we have a player count
		// divisible by four an extra match isn't started. E.g. when we
		// have 16 players we want 4 matches, but without the -1 a fifth
		// match would be added.
		y := (x - 8 - 1) / 4
		m := len(tm.Matches) - 3
		compare := 3 + y
		assert.Equal(
			compare,
			m,
			fmt.Sprintf("%d player tournament had %d matches, not %d", x, m, compare),
		)
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

func TestPopulateMatchesPopulatesPlayoffsFor8Players(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(8)
	tm.StartTournament(nil)

	assert.Equal(4, len(tm.Matches[0].Players))
	assert.Equal(4, len(tm.Matches[1].Players))
}

func TestPopulateMatchesPopulatesAllMatchesFor24Players(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(24)
	tm.StartTournament(nil)

	assert.Equal(4, len(tm.Matches[0].Players))
	assert.Equal(4, len(tm.Matches[1].Players))
	assert.Equal(4, len(tm.Matches[2].Players))
	assert.Equal(4, len(tm.Matches[3].Players))
	assert.Equal(4, len(tm.Matches[4].Players))
	assert.Equal(4, len(tm.Matches[5].Players))
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

func TestNextMatchNoMatchesAreStartedWithPlayoffs(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(16)
	tm.StartTournament(nil)

	m, err := tm.NextMatch()
	assert.Nil(err)
	assert.Equal(0, m.Index)
	assert.Equal("playoff", m.Kind)

	m.Start(nil)
	m.End(nil)

	m, err = tm.NextMatch()
	assert.Nil(err)
	assert.Equal(1, m.Index)
	assert.Equal("playoff", m.Kind)
	assert.Equal(CurrentMatch(1), tm.Current)
}

func TestNextMatchNoMatchesAreStartedWithPlayoffsDone(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(16)
	tm.StartTournament(nil)
	err := endPlayoffs(tm)
	assert.Nil(err)

	m, err := tm.NextMatch()
	assert.Nil(err)
	assert.Equal(4, m.Index)
	assert.Equal("semi", m.Kind)
	assert.Equal(CurrentMatch(tm.MatchIndex(tm.Semi(0))), tm.Current)
}

func TestNextMatchNoMatchesAreStartedWithPlayoffsAndSemisDone(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(16)
	tm.StartTournament(nil)
	endPlayoffs(tm)
	endSemis(tm)

	m, err := tm.NextMatch()
	assert.Nil(err)
	assert.Equal(6, m.Index)
	assert.Equal("final", m.Kind)
	assert.Equal(CurrentMatch(tm.MatchIndex(tm.Final())), tm.Current)
}

func TestNextMatchEverythingDone(t *testing.T) {
	assert := assert.New(t)
	tm := testTournament(16)
	tm.StartTournament(nil)
	endPlayoffs(tm)
	endSemis(tm)
	tm.Final().Start(nil)
	tm.Final().End(nil)

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

func TestEnd4MatchPlayoffsPlacesWinnerAndSecondIntoSemisAndRestIntoRunnerups(t *testing.T) {
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

	assert.Equal(1, len(tm.Semi(0).Players))
	assert.Equal(1, len(tm.Semi(1).Players))
	assert.Equal(2, len(tm.Runnerups))

	assert.Equal(winner, tm.Semi(0).Players[0].Name())
	assert.Equal(silver, tm.Semi(1).Players[0].Name())
}

func TestEndComplete16PlayerTournamentKillsOnly(t *testing.T) {
	assert := assert.New(t)

	tm := testTournament(16)
	tm.StartTournament(nil)

	// Playoff 1 (same as test above)
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

	assert.Equal(1, len(tm.Semi(0).Players))
	assert.Equal(1, len(tm.Semi(1).Players))
	assert.Equal(2, len(tm.Runnerups))

	assert.Equal(winner, tm.Semi(0).Players[0].Name())
	assert.Equal(silver, tm.Semi(1).Players[0].Name())

	// Playoff 2
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

	assert.Equal(2, len(tm.Semi(0).Players))
	assert.Equal(2, len(tm.Semi(1).Players))
	assert.Equal(4, len(tm.Runnerups))

	assert.Equal(winner2, tm.Semi(1).Players[1].Name())
	assert.Equal(silver2, tm.Semi(0).Players[1].Name())

	// Playoff 3
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

	assert.Equal(3, len(tm.Semi(0).Players))
	assert.Equal(3, len(tm.Semi(1).Players))
	assert.Equal(6, len(tm.Runnerups))

	assert.Equal(winner3, tm.Semi(0).Players[2].Name())
	assert.Equal(silver3, tm.Semi(1).Players[2].Name())

	// Playoff 4
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

	assert.Equal(4, len(tm.Semi(0).Players))
	assert.Equal(4, len(tm.Semi(1).Players))
	assert.Equal(8, len(tm.Runnerups))

	assert.Equal(winner4, tm.Semi(1).Players[3].Name())
	assert.Equal(silver4, tm.Semi(0).Players[3].Name())

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

	assert.Equal(2, len(tm.Final().Players))

	assert.Equal(winners1, tm.Final().Players[0].Name())
	assert.Equal(silvers1, tm.Final().Players[1].Name())

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

	assert.Equal(4, len(tm.Final().Players))

	assert.Equal(winners2, tm.Final().Players[2].Name())
	assert.Equal(silvers2, tm.Final().Players[3].Name())

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

	// There should be 5 playoffs (and the predefineds)
	assert.Equal(5+3, len(tm.Matches))

	// Playoff 1
	m, err := tm.NextMatch()
	assert.NoError(err)

	m.Start(nil)

	m.Players[0].AddKills(5)
	m.Players[1].AddKills(6)
	m.Players[2].AddKills(7)
	m.Players[3].AddKills(10)
	winner := m.Players[3].Name()

	m.End(nil)

	assert.Equal(1, len(tm.Semi(0).Players))
	assert.Equal(0, len(tm.Semi(1).Players))
	assert.Equal(3, len(tm.Runnerups))

	assert.Equal(winner, tm.Semi(0).Players[0].Name())

	// Playoff 2
	m2, err2 := tm.NextMatch()
	assert.NoError(err2)

	m2.Start(nil)

	m2.Players[0].AddKills(2)
	m2.Players[1].AddKills(10)
	m2.Players[2].AddKills(8)
	m2.Players[3].AddKills(4)
	winner2 := m2.Players[1].Name()

	m2.End(nil)

	assert.Equal(1, len(tm.Semi(0).Players))
	assert.Equal(1, len(tm.Semi(1).Players))
	assert.Equal(6, len(tm.Runnerups))

	assert.Equal(winner2, tm.Semi(1).Players[0].Name())

	// Playoff 3
	m3, err3 := tm.NextMatch()
	assert.NoError(err3)

	m3.Start(nil)

	m3.Players[0].AddKills(10)
	m3.Players[1].AddKills(3)
	m3.Players[2].AddKills(3)
	m3.Players[3].AddKills(5)
	winner3 := m3.Players[0].Name()

	m3.End(nil)

	assert.Equal(2, len(tm.Semi(0).Players))
	assert.Equal(1, len(tm.Semi(1).Players))
	assert.Equal(9, len(tm.Runnerups))

	assert.Equal(winner3, tm.Semi(0).Players[1].Name())

	// Playoff 4
	m4, err4 := tm.NextMatch()
	assert.NoError(err4)

	m4.Start(nil)

	m4.Players[0].AddKills(9)
	m4.Players[1].AddKills(10)
	m4.Players[2].AddKills(5)
	m4.Players[3].AddKills(5)
	winner4 := m4.Players[1].Name()

	m4.End(nil)

	assert.Equal(2, len(tm.Semi(0).Players))
	assert.Equal(2, len(tm.Semi(1).Players))
	assert.Equal(12, len(tm.Runnerups))

	assert.Equal(winner4, tm.Semi(1).Players[1].Name())

	// Playoff 5 / Runnerup 1
	m5, err5 := tm.NextMatch()
	assert.NoError(err5)
	assert.Equal("playoff", m5.Kind)

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

	assert.Equal(winner5, tm.Semi(0).Players[2].Name())

	// We need to backfill the players, and since that is a judge action we need
	// to simulate that
	err = tm.BackfillSemis(nil, []string{
		tm.Runnerups[0].ID,
		tm.Runnerups[1].ID,
		tm.Runnerups[2].ID,
	})

	assert.NoError(err)

	assert.Equal(4, len(tm.Semi(0).Players))
	assert.Equal(4, len(tm.Semi(1).Players))
	assert.Equal(11, len(tm.Runnerups))

	// Semi 1
	s1, serr1 := tm.NextMatch()
	assert.NoError(serr1)

	assert.Equal("semi", s1.Kind)

	s1.Start(nil)

	s1.Players[0].AddKills(10)
	s1.Players[1].AddKills(7)
	s1.Players[2].AddKills(9)
	s1.Players[3].AddKills(8)
	winners1 := s1.Players[0].Name()
	silvers1 := s1.Players[2].Name()

	s1.End(nil)

	assert.Equal(2, len(tm.Final().Players))

	assert.Equal(winners1, tm.Final().Players[0].Name())
	assert.Equal(silvers1, tm.Final().Players[1].Name())

	// Semi 2
	s2, serr2 := tm.NextMatch()
	assert.NoError(serr2)

	assert.Equal("semi", s2.Kind)

	s2.Start(nil)

	s2.Players[0].AddKills(8)
	s2.Players[1].AddKills(10)
	s2.Players[2].AddKills(8)
	s2.Players[3].AddKills(9)
	winners2 := s2.Players[1].Name()
	silvers2 := s2.Players[3].Name()

	s2.End(nil)

	assert.Equal(4, len(tm.Final().Players))

	assert.Equal(winners2, tm.Final().Players[2].Name())
	assert.Equal(silvers2, tm.Final().Players[3].Name())

	// Final!
	f, ferr := tm.NextMatch()
	assert.NoError(ferr)

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
