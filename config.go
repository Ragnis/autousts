package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

type configFlags struct {
	fs *flag.FlagSet

	Show   string
	Season uint
	Delete bool

	Query      string
	Pointer    string
	MinSeeders uint
	PreferHQ   bool
	Paused     bool

	EpisodeCount uint
	BeginDate    string
}

func (cf *configFlags) Parse(argv []string) {
	cf.fs = flag.NewFlagSet("config", flag.ExitOnError)

	cf.fs.StringVar(&cf.Show, "show", "", "the show to configure")
	cf.fs.UintVar(&cf.Season, "season", 0, "the season to configure")
	cf.fs.BoolVar(&cf.Delete, "delete", false, "delete the item")

	cf.fs.StringVar(&cf.Query, "query", "", "search query; must contain '%s' for the pointer")
	cf.fs.StringVar(&cf.Pointer, "pointer", "", "last-downloaded episode")
	cf.fs.UintVar(&cf.MinSeeders, "min-seeders", 0, "minimum number of seeders allowed")
	cf.fs.BoolVar(&cf.PreferHQ, "prefer-hq", false, "prefer high-quality torrents")
	cf.fs.BoolVar(&cf.Paused, "paused", false, "if set, the show will not be downloaded")

	cf.fs.UintVar(&cf.EpisodeCount, "epc", 0, "number of episodes in the season")
	cf.fs.StringVar(&cf.BeginDate, "begin", "", "begin date of the season")

	cf.fs.Parse(argv)
}

// ParsedFlags returns a slice of flag names that were present in the argument
// list
func (cf configFlags) ParsedFlagNames() (flags []string) {
	cf.fs.Visit(func(f *flag.Flag) {
		flags = append(flags, f.Name)
	})
	return
}

func cmdConfig(db *DB, argv []string) int {
	cf := &configFlags{}
	cf.Parse(argv)

	if cf.Show == "" {
		fmt.Println("no show specified")
		return 1
	}
	if cf.Season > 0 {
		return cmdConfigSeason(db, cf)
	}
	return cmdConfigShow(db, cf)
}

func cmdConfigShow(db *DB, cf *configFlags) int {
	if cf.Delete {
		if err := db.DeleteShow(cf.Show); err != nil {
			fmt.Printf("error deleting show: %v\n", err)
			return 1
		}
		return 0
	}
	show, err := db.Show(cf.Show)
	if err != nil {
		if err != ErrNoShow {
			fmt.Printf("error querying show: %v\n", err)
			return 1
		}

		fmt.Printf("new show: %s\n", cf.Show)
		show = &Show{Name: cf.Show}
	}
	for _, name := range cf.ParsedFlagNames() {
		switch name {
		case "query":
			if !strings.Contains(cf.Query, "%s") {
				fmt.Printf("query did not contain '%%s'\n")
				return 1
			}
			show.Query = cf.Query
		case "min-seeders":
			show.SeedersMin = cf.MinSeeders
		case "prefer-hq":
			show.PreferHQ = cf.PreferHQ
		case "paused":
			show.Paused = cf.Paused
		case "pointer":
			ptr, err := PointerFromString(cf.Pointer)
			if err != nil {
				fmt.Printf("error parsing pointer: %v", err)
				return 1
			}
			show.Pointer = ptr
		}
	}
	if err := db.SaveShow(show); err != nil {
		fmt.Printf("error saving show: %v\n", err)
		return 1
	}
	return 0
}

func cmdConfigSeason(db *DB, cf *configFlags) int {
	show, err := db.Show(cf.Show)
	if err != nil {
		fmt.Printf("error querying show: %v\n", err)
		return 1
	}
	if cf.Delete {
		if ok := show.Seasons.Remove(cf.Season); !ok {
			fmt.Println("no such season")
			return 1
		}
	} else {
		season, ok := show.Seasons.Get(cf.Season)
		if !ok {
			fmt.Printf("new season: %d\n", cf.Season)
			season = &Season{
				Number: cf.Season,
			}
			show.Seasons.Put(season)
		}

		for _, name := range cf.ParsedFlagNames() {
			switch name {
			case "epc":
				season.EpisodeCount = cf.EpisodeCount
			case "begin":
				date, err := time.Parse("2006-01-02", cf.BeginDate)
				if err != nil {
					fmt.Println("invalid date value")
					return 1
				}
				season.Begin = date
			}
		}
	}
	if err := db.SaveShow(show); err != nil {
		fmt.Printf("error saving show: %v\n", err)
		return 1
	}
	return 0
}
