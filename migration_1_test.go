package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

// loadmig1Tournaments ...
func loadmig1Tournaments(db, ro *bolt.DB) (*mig1curTournament, *mig1prevTournament) {
	id := "moon"
	orig := &mig1prevTournament{}
	tx, err := ro.Begin(false)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	b := tx.Bucket(TournamentKey)
	bs := b.Get([]byte(id))
	_ = json.Unmarshal(bs, orig)

	mig := &mig1curTournament{}
	tx, err = db.Begin(false)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	b = tx.Bucket(TournamentKey)
	bs = b.Get([]byte(id))
	_ = json.Unmarshal(bs, mig)

	return mig, orig
}

func TestMigration1(t *testing.T) {
	assert := assert.New(t)

	db, ro := migrationDB(1)
	err := MigrateOriginalColorPreferredColor(db)
	assert.Nil(err)

	v, err := getVersion(db)
	assert.Nil(err)
	assert.Equal(2, v)

	mig, orig := loadmig1Tournaments(db, ro)

	assert.Equal(len(orig.Players), len(mig.Players))

	// Assert that the colors in the matches are correct
	for x, ot := range orig.Tryouts {
		mt := mig.Tryouts[x]

		for y := range ot.Players {
			assert.Equal(ot.Players[y].OriginalColor, mt.Players[y].PreferredColor)
			assert.Equal(ot.Players[y].Color, mt.Players[y].Color)
		}
	}
	for x, ot := range orig.Semis {
		mt := mig.Semis[x]

		for y := range ot.Players {
			assert.Equal(ot.Players[y].OriginalColor, mt.Players[y].PreferredColor)
			assert.Equal(ot.Players[y].Color, mt.Players[y].Color)
		}
	}

	for i := range orig.Final.Players {
		assert.Equal(orig.Final.Players[i].OriginalColor, mig.Final.Players[i].PreferredColor)
		assert.Equal(orig.Final.Players[i].Color, mig.Final.Players[i].Color)
	}

	db.Close()
	ro.Close()
}
