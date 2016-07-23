package search

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"strconv"
)

type Thepiratebay struct {
}

func (tpb *Thepiratebay) Search(query string, options Options) ([]*Result, error) {
	url := url.URL{
		Scheme: "https",
		Host:   "thepiratebay.org",
		Path:   fmt.Sprintf("/search/%s/0/7/0", query),
	}

	doc, err := goquery.NewDocument(url.String())
	if err != nil {
		return nil, err
	}

	ret := []*Result{}

	doc.Find("#SearchResults table#searchResult > tbody > tr").Each(func(i int, tr *goquery.Selection) {
		magnet, ok := tr.Find("a[href^=magnet]").Attr("href")
		if !ok {
			return
		}

		seeders, err := strconv.Atoi(tr.Find("td:nth-child(3)").Text())
		if err != nil || seeders < 0 {
			return
		}

		ret = append(ret, &Result{
			Name:      tr.Find(".detName a.detLink").Text(),
			MagnetURL: magnet,
			Seeders:   uint(seeders),
		})
	})

	return ret, nil
}
