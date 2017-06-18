package main

import (
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
)

// ErrNoShow indicates that the queried show does not exist
var ErrNoShow = errors.New("no such show")

// ShowByName queries a show by it's name
func ShowByName(db *bolt.DB, name string) (show *Show, err error) {
	err = db.View(func(tx *bolt.Tx) (err error) {
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
func Shows(db *bolt.DB) (shows []*Show, err error) {
	err = db.View(func(tx *bolt.Tx) error {
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
func SaveShow(db *bolt.DB, show *Show) error {
	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("Shows")).Put([]byte(show.Name), show.Bytes())
	})
}

// DeleteShow deletes a show
func DeleteShow(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("Shows")).Delete([]byte(name))
	})
}
