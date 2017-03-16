package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
)

func testDB(fn string) *bolt.DB {
	// Reset it
	_ = os.Remove(fn)

	db, err := bolt.Open(fn, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func fatalError(err error) {
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
}

// migrationDB opens a fixture database, makes a copy and returns a bolt
// instance of the copy as well as the original database in RO
func migrationDB(version int) (*bolt.DB, *bolt.DB) {
	base := fmt.Sprintf("migration%d.fixture", version)
	src := "migration-dbs/" + base
	dst := "test/" + base

	_ = os.Remove(dst) // Clear it if it exists
	d, err := os.Create(dst)
	fatalError(err)

	s, err := os.Open(src)
	fatalError(err)

	_, err = io.Copy(d, s)
	fatalError(err)

	db, err := bolt.Open(d.Name(), 0600, nil)
	fatalError(err)

	ro, err := bolt.Open(s.Name(), 0400, nil)
	fatalError(err)

	return db, ro
}

func TestInitialMigration(t *testing.T) {
	assert := assert.New(t)

	db := testDB("test/migration1.db")
	err := InitialMigration(db)
	assert.Nil(err)

	v, err := getVersion(db)
	assert.Nil(err)
	assert.Equal(1, v)
}

func TestBackup(t *testing.T) {
	assert := assert.New(t)
	path := "test/backups/"
	_ = os.RemoveAll(path)

	db := testDB("test/backup.db")

	err := backup(db, 255, path)
	assert.Nil(err)

	files, err := ioutil.ReadDir(path)
	assert.Nil(err)
	assert.Equal(1, len(files))

	assert.True(
		strings.Contains(files[0].Name(), "v"+strconv.Itoa(255)+"-"),
		"Backup file did not contain the version number",
	)
}
