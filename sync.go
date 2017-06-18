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
			results = append(results, result)

		case <-fin:
			waiting--
		}
	}

	for _, result := range results {
		fmt.Printf("Found torrent: '%s'\n", result.Name)

		if !*dryRun {
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
	var k search.Thepiratebay

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
