package main

import (
	"aragnis.com/autousts/db"
	"aragnis.com/autousts/search"
	"fmt"
	"github.com/longnguyen11288/go-transmission/transmission"
	"os"
	"strconv"
	"strings"
	"time"
)

func syncShow(show *db.Show, out chan<- *search.Result, fin chan<- bool) {
	var k search.Kickass

	for {
		pointer, ok := show.NextPointer()
		if !ok {
			break
		}

		query := fmt.Sprintf(show.Query, pointer)

		results, err := k.Search(query, search.Options{})
		if err != nil {
			fmt.Printf("Search error: " + err.Error())
			break
		}

		var chosen *search.Result

		for _, result := range results {
			if show.SeedersMin > 0 && result.Seeders < show.SeedersMin {
				continue
			}

			rptr, err := db.PointerFromString(result.Name)
			if err != nil || rptr.String() != pointer.String() {
				continue
			}

			chosen = result
			break
		}

		if chosen == nil {
			break
		}

		out <- chosen
		show.Pointer = pointer
	}

	fin <- true
}

func sync(dbh *db.Database) {
	tc := transmission.New("http://127.0.0.1:9091", "", "")
	if _, err := tc.GetTorrents(); err != nil {
		fmt.Println("Could not connect to Transmission RPC API: " + err.Error())
		return
	}

	var results []*search.Result

	waiting := 0
	fin := make(chan bool)
	rec := make(chan *search.Result)

	for _, show := range dbh.Shows {
		waiting += 1
		go syncShow(show, rec, fin)
	}

	for waiting > 0 {
		select {
		case result := <-rec:
			results = append(results, result)

		case <-fin:
			waiting -= 1
		}
	}

	for _, result := range results {
		fmt.Printf("Found torrent: '%s'\n", result.Name)

		if _, err := tc.AddTorrentByFilename(result.MagnetURL, ""); err != nil {
			fmt.Println("Error adding torrent: " + err.Error())
			fmt.Println("Stopping...")
			return
		}
	}

	fmt.Printf("Found %d torrent(s)\n", len(results))

	if err := dbh.Sync(); err != nil {
		fmt.Println("Error syncing database: " + err.Error())
	}
}

func viewAll(dbh *db.Database) {
	for _, show := range dbh.Shows {
		name := show.Name
		if len(name) > 15 {
			name = name[:15] + "..."
		}

		fmt.Printf("%-20s %s\n", name, show.Pointer)
	}
}

func view(dbh *db.Database, name string) {
	show, ok := dbh.FindShow(name)
	if !ok {
		fmt.Printf("The specified show '%s' not found.\n", name)
		return
	}

	fmt.Printf("Name        : %s\n", show.Name)
	fmt.Printf("Query       : %s\n", show.Query)
	fmt.Printf("Min seeders : %d\n", show.SeedersMin)
	fmt.Printf("Prefer HQ   : %t\n", show.PreferHQ)
	fmt.Printf("Pointer     : %s\n", show.Pointer)

	if len(show.Seasons) > 0 {
		fmt.Println("")
		fmt.Println("Season  Episodes  Begin")

		for _, season := range show.Seasons {
			fmt.Printf("%-7d %-9d %s\n", season.Number, season.EpisodeCount, season.Begin)
		}
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
		season, ok = show.FindSeason(number)
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
	dbh, err := db.NewDatabase(fmt.Sprintf("%s/.config/goautousts.zip", os.Getenv("HOME")))
	if err != nil {
		fmt.Println("Could not open the database", err)
		return
	}

	var (
		verb     string
		verbArgs []string
	)

	if len(os.Args) >= 2 {
		verb = os.Args[1]
		verbArgs = os.Args[2:]
	}

	switch verb {
	case "sync":
		sync(dbh)
	case "view":
		if len(verbArgs) == 1 {
			view(dbh, verbArgs[0])
		} else {
			viewAll(dbh)
		}
	case "set":
		set(dbh, verbArgs)
	default:
		fmt.Println("No verb specified: {sync, view, set}")
	}

	dbh.Close()
}
