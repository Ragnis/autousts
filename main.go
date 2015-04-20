package main

import (
	"aragnis.com/autousts/search"
	"fmt"
)

func main() {
	var kickass search.Kickass

	results, err := kickass.Search("shameless", search.Options{})
	if err != nil {
		fmt.Println(err)
	}

	for _, result := range results {
		fmt.Println(result)
	}
}
