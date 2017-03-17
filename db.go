package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"log"
	"strings"
)

// Database is the persisting class
type Database struct {
	DB            *bolt.DB
	Server        *Server
	Tournaments   []*Tournament
	People        []*Person
	tournamentRef map[string]*Tournament
}

var (
	// TournamentKey defines the tournament buckets
	TournamentKey = []byte("tournaments")
	// PeopleKey defines the bucket of people
	PeopleKey = []byte("people")
	// MigrationKey defines the bucket of migration levels
	MigrationKey = []byte("migration")
)

// NewDatabase returns a new database object
func NewDatabase(fn string) (*Database, error) {
	bolt, err := bolt.Open(fn, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db := &Database{DB: bolt}
	db.tournamentRef = make(map[string]*Tournament)

	return db, nil
}

// LoadTournaments loads the tournaments from the database and into memory
func (d *Database) LoadTournaments() error {
	err := d.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(TournamentKey)
		if b == nil {
			// If there is no bucket, bail silently.
			// This only really happens in tests.
			// TODO: Fix pls
			return nil
		}

		err := b.ForEach(func(k []byte, v []byte) error {
			t, err := LoadTournament(v, d)
			if err != nil {
				return err
			}

			tournamentMutex.Lock()
			d.Tournaments = append(d.Tournaments, t)
			d.tournamentRef[t.ID] = t
			tournamentMutex.Unlock()
			return nil
		})
		return err
	})

	return err
}

// SaveTournament stores the current state of the tournaments into the db
func (d *Database) SaveTournament(t *Tournament) error {
	ret := d.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(TournamentKey)
		if err != nil {
			return err
		}

		json, _ := t.JSON()
		err = b.Put([]byte(t.ID), json)
		if err != nil {
			log.Fatal(err)
		}

		return nil
	})

	return ret
}

// OverwriteTournament takes a new foreign Tournament{} object and replaces
// the one with the same ID with that one.
//
// Used from the EditHandler()
func (d *Database) OverwriteTournament(t *Tournament) error {
	ret := d.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(TournamentKey)

		json, err := t.JSON()
		if err != nil {
			log.Fatal(err)
		}

		err = b.Put([]byte(t.ID), json)
		if err != nil {
			log.Fatal(err)
		}

		// Replace the tournament in the in-memory list
		for j := 0; j < len(d.Tournaments); j++ {
			ot := d.Tournaments[j]
			if t.ID == ot.ID {
				d.Tournaments = d.Tournaments[:j]
				d.Tournaments = append(d.Tournaments, t)
				d.Tournaments = append(d.Tournaments, d.Tournaments[j+1:]...)
				break
			}
		}
		// And lastly the reference
		d.tournamentRef[t.ID] = t

		return nil
	})

	return ret
}

// SavePerson stores a person into the DB
func (d *Database) SavePerson(p *Person) error {
	ret := d.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(PeopleKey)
		if err != nil {
			return err
		}

		json, _ := p.JSON()
		err = b.Put([]byte(p.ID), json)
		if err != nil {
			log.Fatal(err)
		}

		return nil
	})

	return ret
}

// GetPerson gets a Person{} from the DB
func (d *Database) GetPerson(id string) *Person {
	tx, err := d.DB.Begin(false)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer tx.Rollback()

	b := tx.Bucket(PeopleKey)
	out := b.Get([]byte(id))
	p := &Person{}
	_ = json.Unmarshal(out, p)
	return p
}

// LoadPeople loads the people from the database and into memory
func (d *Database) LoadPeople() error {
	d.People = make([]*Person, 0)
	err := d.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(PeopleKey)

		err := b.ForEach(func(k []byte, v []byte) error {
			p, err := LoadPerson(v)
			if err != nil {
				return err
			}

			d.People = append(d.People, p)
			return nil
		})
		return err
	})

	return err
}

// ClearTestTournaments deletes any tournament that doesn't begin with "DrunkenFall"
func (d *Database) ClearTestTournaments() error {
	err := d.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(TournamentKey)

		err := b.ForEach(func(k []byte, v []byte) error {
			t, err := LoadTournament(v, d)
			if err != nil {
				return err
			}

			if !strings.HasPrefix(t.Name, "DrunkenFall") {
				log.Print("Deleting ", t.ID)
				err := b.Delete([]byte(t.ID))
				if err != nil {
					return err
				}

				// Also remove the database from memory
				delete(d.tournamentRef, t.ID)
				for j := 0; j < len(d.Tournaments); j++ {
					ot := d.Tournaments[j]
					if t.ID == ot.ID {
						d.Tournaments = append(d.Tournaments[:j], d.Tournaments[j+1:]...)
						break
					}
				}
			}
			return nil
		})
		return err
	})

	return err
}

// Close closes the database
func (d *Database) Close() error {
	return d.DB.Close()
}
