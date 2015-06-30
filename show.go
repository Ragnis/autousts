package main

import (
	"encoding/json"
)

type Show struct {
	Name       string    `json:"name"`
	Query      string    `json:"query"`
	SeedersMin uint      `json:"seeders_min"`
	PreferHQ   bool      `json:"prefer_hq"`
	Pointer    Pointer   `json:"pointer"`
	Seasons    []*Season `json:"seasons"`
}

func (s Show) FindSeason(n uint) (*Season, bool) {
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

	season, ok := s.FindSeason(s.Pointer.Season)
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

func (s *Show) DeleteSeason(number uint) {
	keep := []*Season{}

	for _, season := range s.Seasons {
		if season.Number != number {
			keep = append(keep, season)
		}
	}

	s.Seasons = keep
}

func (s Show) Bytes() []byte {
	bytes, err := json.Marshal(s)

	if err != nil {
		return []byte{}
	}

	return bytes
}

func ShowFromBytes(bytes []byte) (*Show, error) {
	show := &Show{}

	if err := json.Unmarshal(bytes, show); err != nil {
		return nil, err
	}

	return show, nil
}
