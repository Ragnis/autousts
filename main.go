package main

import (
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"os"
	"time"
)

// defaultDbFile returns path to the default database file
func defaultDbFile() string {
	return fmt.Sprintf("%s/.config/goautousts.db", os.Getenv("HOME"))
}

func main() {
	dbFile := flag.String("db-file", defaultDbFile(), "Path to the database file")
	flag.Parse()

	db, err := bolt.Open(*dbFile, 0600, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		fmt.Println("Could not open the database", err)
		return
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("Shows")); err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	var (
		rv     int
		sc     string
		scArgs []string
	)
	if na := flag.NArg(); na > 0 {
		sc = flag.Arg(0)
		scArgs = flag.Args()[1:]
	}

	switch sc {
	case "sync":
		rv = cmdSync(db, scArgs)
	case "view":
		rv = cmdView(db, scArgs)
	case "config":
		rv = cmdConfig(db, scArgs)
	default:
		rv = 1
		fmt.Println("allowed subcommands: {sync, view, config}")
	}

	os.Exit(rv)
}
