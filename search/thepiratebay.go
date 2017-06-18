package search

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"strconv"
)

// Thepiratebay is a Searcher backed by The Pirate Bay
type Thepiratebay struct {
}

// Search performs a search
func (tpb *Thepiratebay) Search(query string) (results Results, err error) {
	url := url.URL{
		Scheme: "https",
		Host:   "thepiratebay.org",
		Path:   fmt.Sprintf("/search/%s/0/7/0", query),
	}
	doc, err := goquery.NewDocument(url.String())
	if err != nil {
		err = fmt.Errorf("creating goquery document: %v", err)
		return
	}
	doc.Find("#SearchResults table#searchResult > tbody > tr").Each(func(i int, tr *goquery.Selection) {
		magnet, ok := tr.Find("a[href^=magnet]").Attr("href")
		if !ok {
			return
		}
		seeders, err := strconv.Atoi(tr.Find("td:nth-child(3)").Text())
		if err != nil || seeders < 0 {
			return
		}
		results = append(results, &Result{
			Name:      tr.Find(".detName a.detLink").Text(),
			MagnetURL: magnet,
			Seeders:   uint(seeders),
		})
	})
	return
}
