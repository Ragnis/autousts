package db

import (
	"fmt"
)

type Show struct {
	Name       string
	Query      string
	SeedersMin uint
	PreferHQ   bool
	Pointer    Pointer
	Seasons    []*Season
}

func (s Show) getSeason(n uint) (*Season, bool) {
	for _, season := range s.Seasons {
		if season.Number == n {
			return season, true
		}
	}

	return nil, false
}

func (s Show) getNextSeason(n uint) (*Season, bool) {
	var ret *Season

	for _, season := range s.Seasons {
		if season.Number > n && (ret == nil || season.Number < ret.Number) {
			ret = season
		}
	}

	return ret, ret != nil
}

func (s Show) NextPointer() (Pointer, bool) {
	ret := Pointer{}

	season, ok := s.getSeason(s.Pointer.Season)
	if !ok {
		season, ok = s.getNextSeason(s.Pointer.Season)
		if !ok {
			return ret, false
		}
	}

	if s.Pointer.Episode >= season.EpisodeCount {
		season, ok = s.getNextSeason(s.Pointer.Season)
		if !ok {
			return ret, false
		}
	}

	ret.Season = season.Number

	if season.Number == s.Pointer.Season {
		ret.Episode = s.Pointer.Episode + 1
	} else {
		ret.Episode = 1
	}

	return ret, true
}

func (s Show) Table() []string {
	return []string{
		fmt.Sprintf("Name\t: %s\n", s.Name),
		fmt.Sprintf("Query\t: %s\n", s.Query),
		fmt.Sprintf("Min seeders\t: %d\n", s.SeedersMin),
		fmt.Sprintf("Prefer HQ\t: %t\n", s.PreferHQ),
		fmt.Sprintf("Pointer\t: %s\n", s.Pointer),
	}
}

func (s Show) TableRow() string {
	return fmt.Sprintf("%s\t%s\n", s.Name, s.Pointer)
}
