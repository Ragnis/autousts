package main

import (
	"flag"
	"fmt"
	"os"
)

// defaultDbFile returns path to the default database file
func defaultDbFile() string {
	return fmt.Sprintf("%s/.config/goautousts.db", os.Getenv("HOME"))
}

func main() {
	dbFile := flag.String("db-file", defaultDbFile(), "Path to the database file")
	flag.Parse()

	db, err := OpenDB(*dbFile)
	if err != nil {
		fmt.Printf("error opening the database: %v\n", err)
		return
	}
	defer db.Close()

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
