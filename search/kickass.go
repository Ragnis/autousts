package search

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"strconv"
)

type Kickass struct {
}

func (k *Kickass) Search(query string, options Options) ([]*Result, error) {
	url := url.URL{
		Scheme:   "https",
		Host:     "kickass.to",
		Path:     fmt.Sprintf("/usearch/%s/", query),
		RawQuery: "field=seeders&sorder=desc",
	}

	doc, err := goquery.NewDocument(url.String())
	if err != nil {
		return nil, err
	}

	ret := []*Result{}

	doc.Find("#mainSearchTable table.data tr[id]").Each(func(i int, s *goquery.Selection) {
		magnet, ok := s.Find("a.imagnet").Attr("href")
		if !ok {
			return
		}

		seeders, err := strconv.Atoi(s.Find("td:nth-child(5)").Text())
		if err != nil || seeders < 0 {
			return
		}

		ret = append(ret, &Result{
			Name:      s.Find(".cellMainLink").Text(),
			MagnetURL: magnet,
			Seeders:   uint(seeders),
		})
	})

	return ret, nil
}
