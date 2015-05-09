package main

import (
	"aragnis.com/autousts/db"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

func displayShow(show *db.Show) {
	table := tabwriter.NewWriter(os.Stdout, 0, 4, 0, '\t', 0)

	for _, row := range show.Table() {
		table.Write([]byte(row))
	}

	table.Flush()

	if len(show.Seasons) > 0 {
		fmt.Println("")
		fmt.Println("Season\tEpisodes\tBegin")

		for _, season := range show.Seasons {
			fmt.Println(season.TableRow())
		}
	}
}

func view(dbh *db.Database, args []string) {
	if len(args) == 0 {
		table := tabwriter.NewWriter(os.Stdout, 0, 4, 0, '\t', 0)

		for _, show := range dbh.Shows {
			table.Write([]byte(show.TableRow()))
		}

		table.Flush()
	} else {
		showName := args[0]
		show, ok := dbh.FindShow(showName)
		if !ok {
			fmt.Printf("The specified show '%s' not found.\n", showName)
			return
		}

		displayShow(show)
	}
}

func set(dbh *db.Database, args []string) {
	if len(args) != 3 {
		fmt.Println(`Usage:
  set SHOW PROP VALUE
  set SHOW:SEASON PROP VALUE

Show properties:
  query       : string, search query, must contain '%s' for pointer
  min-seeders : uint, minimum number of seeders allowed
  prefer-hq   : boolean
  pointer     : last downloaded episode

Season properties:
  epc   : uint, episode count
  begin : date, begin date`)
		return
	}

	split := strings.Split(args[0], ":")
	key := args[1]
	value := args[2]

	var (
		ok bool

		name   string = split[0]
		number uint

		show   *db.Show
		season *db.Season
	)

	if len(split) == 2 {
		v, err := strconv.Atoi(split[1])
		if err != nil || v <= 0 {
			fmt.Println("Invalid season specified")
			return
		}
		number = uint(v)
	}

	show, ok = dbh.FindShow(name)
	if !ok {
		show = &db.Show{
			Name: name,
		}
		dbh.Shows = append(dbh.Shows, show)
	}

	if number != 0 {
		season, ok = show.GetSeason(number)
		if !ok {
			season = &db.Season{
				Number: number,
			}
			show.Seasons = append(show.Seasons, season)
		}
	}

	if season != nil {
		switch key {
		case "epc":
			epc, err := strconv.Atoi(value)
			if err != nil || epc < 0 {
				fmt.Println("Invalid value")
				break
			}
			season.EpisodeCount = uint(epc)

		case "begin":
			begin, err := time.Parse("2006-01-02", value)
			if err != nil {
				fmt.Println("Invalid value: " + err.Error())
				break
			}
			season.Begin = begin

		default:
			fmt.Println("Invalid key")
		}
	} else {
		switch key {
		case "query":
			if strings.Count(value, "%s") != 1 {
				fmt.Println("The value must contain exactly one '%s'")
				break
			}
			show.Query = value

		case "seeders-min":
			v, err := strconv.Atoi(value)
			if err != nil || v < 0 {
				fmt.Println("Invalid value")
				break
			}
			show.SeedersMin = uint(v)

		case "prefer-hq":
			switch value {
			case "true":
				show.PreferHQ = true
			case "false":
				show.PreferHQ = false
			default:
				fmt.Println("Invalid value. Allowed values are: true, false")
			}

		case "pointer":
			pointer, err := db.PointerFromString(value)
			if err != nil {
				fmt.Println("Invalid value")
				break
			}
			show.Pointer = pointer

		default:
			fmt.Println("Invalid key")
		}
	}

	if err := dbh.Sync(); err != nil {
		fmt.Println("Error syncing database: " + err.Error())
	}
}

func main() {
	dbh, err := db.NewDatabase("testdb")
	if err != nil {
		fmt.Println("Could not open the database", err)
		return
	}

	if len(os.Args) <= 1 {
		fmt.Println("No verb specified: {sync, view, set}")
		return
	}

	verb := os.Args[1]

	switch verb {
	case "sync":
		fmt.Println("Not implemented")
	case "view":
		view(dbh, os.Args[2:])
	case "set":
		set(dbh, os.Args[2:])
	}

	dbh.Close()
}
