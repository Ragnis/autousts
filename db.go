package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

// ErrNoShow indicates that the queried show does not exist
var ErrNoShow = errors.New("no such show")

// DB is the database
type DB struct {
	bolt *bolt.DB
}

// OpenDB opens a database file
func OpenDB(path string) (*DB, error) {
	bdb, err := bolt.Open(path, 0600, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &DB{bdb}, nil
}

// Close closes the database
func (db *DB) Close() error {
	return db.bolt.Close()
}

// Show queries a show by it's name
func (db *DB) Show(name string) (show *Show, err error) {
	err = db.bolt.View(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket([]byte("Shows")).Get([]byte(name))
		if b == nil {
			err = ErrNoShow
			return
		}
		show, err = ShowFromBytes(b)
		return
	})
	return
}

// Shows queries all shows
func (db *DB) Shows() (shows []*Show, err error) {
	err = db.bolt.View(func(tx *bolt.Tx) error {
		cur := tx.Bucket([]byte("Shows")).Cursor()
		for k, v := cur.First(); k != nil; k, v = cur.Next() {
			show, err := ShowFromBytes(v)
			if err != nil {
				return fmt.Errorf("error loading show '%s': %v", k, err)
			}
			shows = append(shows, show)
		}
		return nil
	})
	return
}

// SaveShow saves a show to the database
func (db *DB) SaveShow(show *Show) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("Shows"))
		if err != nil {
			return fmt.Errorf("creating bucket: %v", err)
		}
		return b.Put([]byte(show.Name), show.Bytes())
	})
}

// DeleteShow deletes a show
func (db *DB) DeleteShow(name string) error {
	return db.bolt.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("Shows")).Delete([]byte(name))
	})
}
