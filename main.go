package main

import (
	"aragnis.com/autousts/db"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

func displayShow(show *db.Show) {
	table := tabwriter.NewWriter(os.Stdout, 0, 4, 0, '\t', 0)
	seasonTable := tabwriter.NewWriter(os.Stdout, 0, 4, 0, '\t', 0)

	for _, row := range show.Table() {
		table.Write([]byte(row))
	}

	for _, season := range show.Seasons {
		seasonTable.Write([]byte(season.TableRow()))
	}

	table.Flush()

	if len(show.Seasons) > 0 {
		fmt.Println("Seasons:")
		seasonTable.Flush()
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

func edit(dbh *db.Database, args []string) {
	if len(args) == 0 {
		fmt.Println("No show specified.\n")
		return
	}

	show, ok := dbh.FindShow(args[0])
	if !ok {
		fmt.Printf("The specified show '%s' not found.\n", args[0])
		fmt.Println("Creating it...")

		show := &db.Show{
			Name: args[0],
		}
		dbh.Shows = append(dbh.Shows, show)
	}

	flags := flag.NewFlagSet("edit", flag.ExitOnError)
	var flagQuery = flags.String("query", show.Query, "Search query. Must contain '%s'")
	var flagSeedersMin = flags.Uint("min-seeders", show.SeedersMin, "Minimum amount of seeders")
	var flagPreferHQ = flags.Bool("prefer-hq", show.PreferHQ, "Prefer HQ")
	var flagPointer = flags.String("pointer", show.Pointer.String(), "Set the show pointer")

	flags.Parse(args[1:len(args)])

	pointer, err := db.PointerFromString(*flagPointer)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	show.Query = *flagQuery
	show.SeedersMin = *flagSeedersMin
	show.PreferHQ = *flagPreferHQ
	show.Pointer = pointer

	if err := dbh.Sync(); err != nil {
		fmt.Println("Error saving the database", err)
	} else {
		fmt.Println("Changes saved\n")
		displayShow(show)
	}
}

func main() {
	dbh, err := db.NewDatabase("testdb")
	if err != nil {
		fmt.Println("Could not open the database", err)
		return
	}

	if len(os.Args) <= 1 {
		fmt.Println("No verb specified: {sync, view, edit}")
		return
	}

	verb := os.Args[1]

	switch verb {
	case "sync":
		fmt.Println("Not implemented")
	case "view":
		view(dbh, os.Args[2:])
	case "edit":
		edit(dbh, os.Args[2:])
	}

	dbh.Close()
}
