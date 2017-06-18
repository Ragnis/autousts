package main

import (
	"flag"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/longnguyen11288/go-transmission/transmission"

	"github.com/Ragnis/autousts/search"
)

func cmdSync(db *bolt.DB, argv []string) int {
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
