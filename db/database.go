package db

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Database struct {
	file  *os.File
	Shows []*Show
}

// Open or create a database file
func NewDatabase(filename string) (*Database, error) {
	db := &Database{}

	file, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.Create(filename)
		}

		if err != nil {
			return db, err
		}
	}

	db.file = file

	if err = db.Read(); err != nil {
		return db, err
	}

	return db, nil
}

// Reads data from the database file overriding all unsaved changes
func (db *Database) Read() error {
	fstat, err := db.file.Stat()
	if err != nil {
		return err
	}

	if fstat.Size() == 0 {
		return nil
	}

	r, err := zip.NewReader(db.file, fstat.Size())
	if err != nil {
		return err
	}

	db.Shows = []*Show{}

	for _, file := range r.File {
		if !strings.HasPrefix(file.Name, "shows/") {
			continue
		}

		fr, err := file.Open()
		if err != nil {
			return err
		}
		defer fr.Close()

		var show *Show

		dec := json.NewDecoder(fr)
		if err = dec.Decode(&show); err != nil {
			return err
		}

		db.Shows = append(db.Shows, show)
	}

	return nil
}

// Write changes to the database
func (db *Database) Sync() error {
	w := zip.NewWriter(db.file)
	defer w.Close()

	for _, show := range db.Shows {
		fw, err := w.Create(fmt.Sprintf("shows/%s.json", show.Name))
		if err != nil {
			return err
		}

		enc := json.NewEncoder(fw)
		if err = enc.Encode(show); err != nil {
			return err
		}
	}

	return nil
}

// Close the database file
func (db Database) Close() error {
	return db.file.Close()
}

// Find a show by it's name
func (db Database) FindShow(name string) (*Show, bool) {
	for _, show := range db.Shows {
		if show.Name == name {
			return show, true
		}
	}

	return nil, false
}
