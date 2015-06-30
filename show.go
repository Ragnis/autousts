package main

import (
	"encoding/json"
)

type Show struct {
	Name       string  `json:"name"`
	Query      string  `json:"query"`
	SeedersMin uint    `json:"seeders_min"`
	PreferHQ   bool    `json:"prefer_hq"`
	Pointer    Pointer `json:"pointer"`
	Seasons    Seasons `json:"seasons"`
}

func (s Show) NextPointer() (Pointer, bool) {
	ret := Pointer{}

	season, ok := s.Seasons.Get(s.Pointer.Season)
	if ok && s.Pointer.Episode >= season.EpisodeCount {
		// The current season has ended, look for a next one
		ok = false
	}

	if !ok {
		season, ok = s.Seasons.GetClosest(s.Pointer.Season + 1)
		if !ok {
			return Pointer{}, false
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
