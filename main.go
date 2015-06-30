package main

import (
	"aragnis.com/autousts/search"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/longnguyen11288/go-transmission/transmission"
	"os"
	"strconv"
	"strings"
	"time"
)

func syncShow(show *Show, out chan<- *search.Result, fin chan<- bool) {
	var k search.Kickass

	for {
		pointer, ok := show.NextPointer()
		if !ok {
			break
		}

		query := fmt.Sprintf(show.Query, pointer)

		results, err := k.Search(query, search.Options{})
		if err != nil {
			fmt.Println("Search error: " + err.Error())
			break
		}

		var chosen *search.Result

		for _, result := range results {
			if show.SeedersMin > 0 && result.Seeders < show.SeedersMin {
				continue
			}

			rptr, err := PointerFromString(result.Name)
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

func sync(db *bolt.DB) {
	tc := transmission.New("http://127.0.0.1:9091", "", "")
	if _, err := tc.GetTorrents(); err != nil {
		fmt.Println("Could not connect to Transmission RPC API: " + err.Error())
		return
	}

	var (
		shows   []*Show
		results []*search.Result
	)

	waiting := 0
	fin := make(chan bool)
	rec := make(chan *search.Result)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Shows"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			show, err := ShowFromBytes(v)
			if err == nil {
				shows = append(shows, show)
			}
		}

		return nil
	})

	for _, show := range shows {
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

	err := db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Shows"))

		for _, show := range shows {
			if err := b.Put([]byte(show.Name), show.Bytes()); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error occurred while saving shows: %s", err.Error())
	}
}

func viewAll(db *bolt.DB) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Shows"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			show, err := ShowFromBytes(v)
			if err != nil {
				fmt.Println("Error loading show '%s'", k)
				continue
			}

			name := show.Name
			if len(name) > 15 {
				name = name[:15] + "..."
			}

			fmt.Printf("%-20s %s\n", name, show.Pointer)
		}

		return nil
	})
}

func view(db *bolt.DB, name string) {
	var (
		err  error
		show *Show
	)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Shows"))
		show, err = ShowFromBytes(b.Get([]byte(name)))
		return err
	})

	if show == nil {
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

func set(db *bolt.DB, args []string) {
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
		ok  bool
		err error

		name   string = split[0]
		number uint

		show   *Show
		season *Season
	)

	if len(split) == 2 {
		v, err := strconv.Atoi(split[1])
		if err != nil || v <= 0 {
			fmt.Println("Invalid season specified")
			return
		}
		number = uint(v)
	}

	// Attempt to read the show from the database
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Shows"))
		show, err = ShowFromBytes(b.Get([]byte(name)))
		return err
	})

	if show == nil {
		show = &Show{Name: name}
	}

	show.Name = name

	if number != 0 {
		season, ok = show.FindSeason(number)
		if !ok {
			season = &Season{
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
			pointer, err := PointerFromString(value)
			if err != nil {
				fmt.Println("Invalid value")
				break
			}
			show.Pointer = pointer

		default:
			fmt.Println("Invalid key")
		}
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Shows"))
		return b.Put([]byte(name), show.Bytes())
	})

	if err != nil {
		fmt.Printf("Error occurred while saving shows: %s", err.Error())
	}
}

func rm(db *bolt.DB, name string) {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Shows"))
		b.Delete([]byte(name))
		return nil
	})

	if err != nil {
		fmt.Printf("Database error: %s", err.Error())
		return
	}

	fmt.Printf("Show '%s' has been deleted, if it ever existed.\n", name)
}

func main() {
	dbFile := fmt.Sprintf("%s/.config/goautousts.db", os.Getenv("HOME"))
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{
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
		verb     string
		verbArgs []string
	)

	if len(os.Args) >= 2 {
		verb = os.Args[1]
		verbArgs = os.Args[2:]
	}

	switch verb {
	case "sync":
		sync(db)
	case "view":
		if len(verbArgs) == 1 {
			view(db, verbArgs[0])
		} else {
			viewAll(db)
		}
	case "set":
		set(db, verbArgs)
	case "rm":
		rm(db, verbArgs[0])
	default:
		fmt.Println("No verb specified: {sync, view, set, rm}")
	}
}
