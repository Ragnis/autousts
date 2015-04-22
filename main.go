package main

import (
	"aragnis.com/autousts/db"
	"fmt"
	"os"
	"text/tabwriter"
)

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
		fmt.Println("Not implemented")
	}

	dbh.Close()
}
