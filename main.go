package main

import (
	"aragnis.com/autousts/db"
	"fmt"
)

func main() {
	dbh, err := db.NewDatabase("testdb")
	if err != nil {
		fmt.Println("Could not open the database", err)
		return
	}

	dbh.Close()
}
