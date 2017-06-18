package search

import (
	"net/url"
	"strings"
)

// Result is a single search result
type Result struct {
	Name      string
	MagnetURL string
	Seeders   uint
	Size      uint64
}

// Searcher can search for torrents
type Searcher interface {
	Search(query string) ([]*Result, error)
}

// Aggregator aggregates search results from multiple Searchers
type Aggregator struct {
	searchers []Searcher
}

// AddSearcher adds a searcher to the aggregator. The caller must take care of
// not adding the save searcher multiple times.
func (ag *Aggregator) AddSearcher(s Searcher) {
	ag.searchers = append(ag.searchers, s)
}

// Search performs a search using all searchers and aggregates the results
// removing any duplicates.
func (ag Aggregator) Search(query string) ([]*Result, error) {
	var (
		ret []*Result
		xts = make(map[string]bool)
	)
	for _, s := range ag.searchers {
		list, err := s.Search(query)
		if err != nil {
			return nil, err
		}
		for _, r := range list {
			for _, xt := range parseExactTopics(r.MagnetURL) {
				if xts[xt] {
					continue
				}
				xts[xt] = true
			}
			ret = append(ret, r)
		}
	}
	return ret, nil
}

func parseExactTopics(magnet string) (ret []string) {
	u, err := url.Parse(magnet)
	if err != nil || u.Scheme != "magnet" {
		return
	}
	for k, vlist := range u.Query() {
		if k == "xt" || strings.HasPrefix(k, "xt.") {
			ret = append(ret, vlist...)
		}
	}
	return
}
