package search

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

type Eztv struct {
}

func (eztv *Eztv) Search(query string, options Options) ([]*Result, error) {
	url := url.URL{
		Scheme: "https",
		Host:   "eztv.ag",
		Path:   fmt.Sprintf("/search/%s", query),
	}

	doc, err := goquery.NewDocument(url.String())
	if err != nil {
		return nil, err
	}

	ret := []*Result{}

	doc.Find("table > tbody > tr").Each(func(i int, tr *goquery.Selection) {
		name, ok := tr.Find("a.epinfo").Text()
		if !ok {
			return
		}

		magnet, ok := tr.Find("a[href^=magnet]").Attr("href")
		if !ok {
			return
		}

		seeders, err := strconv.Atoi(tr.Find("font[color=green]").Text())
		if err != nil || seeders < 0 {
			return
		}

		ret = append(ret, &Result{
			Name:      name,
			MagnetURL: magnet,
			Seeders:   uint(seeders),
		})
	})

	return ret, nil
}
