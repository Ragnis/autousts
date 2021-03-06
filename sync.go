package main

import (
	"flag"
	"fmt"

	"github.com/Ragnis/autousts/search"
	"github.com/longnguyen11288/go-transmission/transmission"
)

func cmdSync(db *DB, argv []string) int {
	var (
		fs     = flag.NewFlagSet("sync", flag.ExitOnError)
		dryRun = fs.Bool("dry-run", false, "do not actually download or save anything")

		tc transmission.TransmissionClient
	)
	fs.Parse(argv)

	if !*dryRun {
		tc = transmission.New("http://127.0.0.1:9091", "", "")
		if _, err := tc.GetTorrents(); err != nil {
			fmt.Printf("could not connect to Transmission RPC API: %s\n", err)
			return 1
		}
	}

	var (
		err     error
		shows   []*Show
		results []*search.Result
	)

	waiting := 0
	fin := make(chan bool)
	rec := make(chan *search.Result)

	shows, err = db.Shows()
	if err != nil {
		fmt.Printf("could not query shows: %v\n", err)
		return 1
	}

	for _, show := range shows {
		if show.Paused {
			continue
		}

		waiting++
		go syncShow(show, rec, fin)
	}

	for waiting > 0 {
		select {
		case result := <-rec:
			fmt.Printf("Found torrent: '%s'\n", result.Name)
			results = append(results, result)

		case <-fin:
			waiting--
		}
	}

	if !*dryRun {
		for _, result := range results {
			if _, err := tc.AddTorrentByFilename(result.MagnetURL, ""); err != nil {
				fmt.Println("Error adding torrent: " + err.Error())
				fmt.Println("Stopping...")
				return 1
			}
		}
	}

	fmt.Printf("Found %d torrent(s)\n", len(results))

	if !*dryRun {
		var ec int
		for _, show := range shows {
			if err := db.SaveShow(show); err != nil {
				fmt.Printf("error saving show '%s': %v\n", show.Name, err)
				ec++
			}
		}
		if ec > 0 {
			fmt.Printf("finished with %d errors\n", ec)
			return 1
		}
	}

	return 0
}

func syncShow(show *Show, out chan<- *search.Result, fin chan<- bool) {
	aggr := &search.Aggregator{}
	aggr.AddSearcher(&search.Thepiratebay{})

	for {
		pointer, ok := show.NextPointer()
		if !ok {
			break
		}

		query := fmt.Sprintf(show.Query, pointer)

		results, err := aggr.Search(query)
		if err != nil {
			fmt.Println("Search error: " + err.Error())
			break
		}
		results = results.Filter(func(r *search.Result) bool {
			if show.SeedersMin > 0 && r.Seeders < show.SeedersMin {
				return false
			}
			if !containsPointer(r.Name, pointer) {
				return false
			}
			return true
		})
		results.Sort()

		if len(results) == 0 {
			break
		}

		out <- results[0]
		show.Pointer = pointer
	}

	fin <- true
}

func containsPointer(str string, ptr Pointer) bool {
	nptr, err := PointerFromString(str)
	if err != nil {
		return false
	}
	return nptr.String() == ptr.String()
}
