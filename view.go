package main

import (
	"fmt"

	"github.com/boltdb/bolt"
)

func cmdView(db *bolt.DB, argv []string) int {
	var rv int
	switch len(argv) {
	case 0:
		rv = cmdViewAll(db)
	case 1:
		rv = cmdViewShow(db, argv[0])
	default:
		fmt.Println("too many arguments")
		rv = 1
	}
	return rv
}

func cmdViewAll(db *bolt.DB) int {
	shows, err := Shows(db)
	if err != nil {
		fmt.Printf("error loading shows: %v\n", err)
		return 1
	}
	for _, show := range shows {
		var (
			name     = show.Name
			hasNext  string
			isPaused string
		)
		if len(name) > 15 {
			name = name[:15] + "..."
		}
		if _, ok := show.NextPointer(); ok {
			hasNext = "+"
		}
		if show.Paused {
			isPaused = "paused"
		}
		fmt.Printf("%-20s %s %s %s\n", name, show.Pointer, hasNext, isPaused)
	}
	return 0
}

func cmdViewShow(db *bolt.DB, name string) int {
	show, err := ShowByName(db, name)
	if err != nil {
		fmt.Printf("error querying show: %v\n", err)
		return 1
	}

	fmt.Printf("Name        : %s\n", show.Name)
	fmt.Printf("Query       : %s\n", show.Query)
	fmt.Printf("Min seeders : %d\n", show.SeedersMin)
	fmt.Printf("Prefer HQ   : %t\n", show.PreferHQ)
	fmt.Printf("Pointer     : %s\n", show.Pointer)
	fmt.Printf("Paused      : %t\n", show.Paused)

	if seasons := show.Seasons.Slice(); seasons != nil {
		fmt.Println("")
		fmt.Println("Season  Episodes  Begin")

		for _, season := range seasons {
			fmt.Printf("%-7d %-9d %s\n", season.Number, season.EpisodeCount, season.Begin)
		}
	}

	return 0
}
