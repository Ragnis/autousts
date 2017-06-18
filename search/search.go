package search

import (
	"net/url"
	"sort"
	"strings"
)

// Result is a single search result
type Result struct {
	Name      string
	MagnetURL string
	Seeders   uint
	Size      uint64
}

// Results is a slice of search results
type Results []*Result

// ResultFilterFunc is used to filter a list of results
type ResultFilterFunc func(*Result) bool

// Searcher can search for torrents
type Searcher interface {
	Search(query string) (Results, error)
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
func (ag Aggregator) Search(query string) (Results, error) {
	var (
		ret Results
		xts = make(map[string]bool)
	)
	for _, s := range ag.searchers {
		list, err := s.Search(query)
		if err != nil {
			return nil, err
		}
		for _, r := range list {
			var dup bool
			for _, xt := range parseExactTopics(r.MagnetURL) {
				if xts[xt] {
					dup = true
				}
				xts[xt] = true
			}
			if !dup {
				ret = append(ret, r)
			}
		}
	}
	return ret, nil
}

// Filter creates a new results list using the filter function
func (rs *Results) Filter(fun ResultFilterFunc) (out Results) {
	for _, r := range *rs {
		if fun(r) {
			out = append(out, r)
		}
	}
	return
}

// Sort sorts the result list
func (rs *Results) Sort() {
	sort.Slice(*rs, func(i, j int) bool {
		return (*rs)[i].Seeders < (*rs)[j].Seeders
	})
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
