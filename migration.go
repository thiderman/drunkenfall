package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	errNoMigrationsYet = errors.New("no migrations have been added yet")
	levelKey           = []byte("level")
)

// migrations is a list of all the migration functions.
// Their indexes are used to determine when they are to be applied.
//
// When adding a new migration, make sure to add it to this list.
var migrations = []func(db *bolt.DB) error{
	InitialMigration,
	MigrateOriginalColorPreferredColor,
	MigrateTournamentRunnerupStringPerson,
	MigrateMatchScoreOrderKillOrder,
	MigrateMatchCommitToRound,
}

var EmptyDBError = errors.New("drunkenfall: Empty DB.")

// Migrate is the main migration entrypoint
//
// When called, it will check the database to see what migrations have already
// been applied. If that is lower than the const TopVersion, all migrations up
// to that point will sequentially be applied.
func Migrate(db *bolt.DB) error {
	err := db.View(func(tx *bolt.Tx) error {
		if tx.Bucket(TournamentKey) == nil {
			return EmptyDBError
		}
		return nil
	})
	if err == EmptyDBError {
		log.Print("Empty db detected, nothing to migrate.")
		err := db.Update(func(tx *bolt.Tx) error {
			return setVersion(tx, len(migrations))
		})
		return err
	}
	version, err := getVersion(db)

	if err == errNoMigrationsYet {
		// No migrations have been done yet, so lets do the first one by adding
		// the migration bucket.
		version = 0
	} else if err != nil {
		// Something actually went wrong. Oh noes.
		return err
	}

	// If version is lower than the latest known version, it's time to apply the
	// migrations!
	if version < len(migrations) {
		if err := applyMigrations(db, version); err != nil {
			log.Print("Error: Migration application failed ;'(")
			return err
		}
	} else {
		log.Print("Database up to date.")
	}

	return nil
}

func applyMigrations(db *bolt.DB, version int) error {
	log.Printf(" --- Migrating %d -> %d:\n", version, len(migrations))
	if err := backup(db, version, "db-migration-backup/"); err != nil {
		return err
	}

	// Run the new migrations and the new migrations only
	new := migrations[version:]
	for x, migration := range new {
		log.Printf("Applying migration %d\n", x+len(new))
		if err := migration(db); err != nil {
			log.Print("Migration failure: ", err)
			return err
		}
	}

	log.Printf(" --- Migrations applied successfully. <3\n")
	return nil
}

func backup(db *bolt.DB, version int, path string) error {
	_ = os.Mkdir(path, 0755)

	fn := fmt.Sprintf(
		"%d_from-v%d-to-v%d.db",
		time.Now().UnixNano(),
		version,
		len(migrations),
	)
	dst := filepath.Join(path, fn)

	data, _ := ioutil.ReadFile(db.Path())
	if err := ioutil.WriteFile(dst, data, 0644); err != nil {
		return err
	}

	log.Printf("Backed up into %s\n", fn)
	return nil
}

func getVersion(db *bolt.DB) (int, error) {
	var version int
	err := db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket(MigrationKey)
		if b == nil {
			return errNoMigrationsYet
		}

		x := b.Get(levelKey)
		version, err = strconv.Atoi(string(x))
		return err
	})

	return version, err
}

func setVersion(tx *bolt.Tx, version int) error {
	b, err := tx.CreateBucketIfNotExists(MigrationKey)
	if err != nil {
		return err
	}

	out, _ := json.Marshal(version)
	err = b.Put(levelKey, out)
	return err
}

// InitialMigration adds the db_version bucket with the sole entry
func InitialMigration(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		return setVersion(tx, 1)
	})
}
