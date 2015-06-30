package main

import (
	"encoding/json"
	"errors"
	"time"
)

type Season struct {
	Number       uint      `json:"number"`
	EpisodeCount uint      `json:"episode_count"`
	Begin        time.Time `json:"begin"`
}

type Seasons struct {
	items []*Season
}

func (ss Seasons) Slice() []*Season {
	return ss.items
}

func (ss *Seasons) Put(s *Season) {
	ss.items = append(ss.items, s)
}

func (ss Seasons) Get(n uint) (*Season, bool) {
	for _, s := range ss.items {
		if s.Number == n {
			return s, true
		}
	}

	return nil, false
}

func (ss Seasons) GetClosest(n uint) (*Season, bool) {
	var ret *Season

	for _, s := range ss.items {
		if s.Number == n {
			return s, true
		}

		if s.Number > n && (ret == nil || s.Number < ret.Number) {
			ret = s
		}
	}

	return ret, ret != nil
}

func (ss *Seasons) Remove(n uint) bool {
	keep := []*Season{}
	ok := false

	for _, s := range ss.items {
		if s.Number == n {
			ok = true
		} else {
			keep = append(keep, s)
		}
	}

	ss.items = keep
	return ok
}

func (ss Seasons) MarshalJSON() ([]byte, error) {
	return json.Marshal(ss.items)
}

func (ss *Seasons) UnmarshalJSON(data []byte) error {
	if ss == nil {
		return errors.New("Seasons: UnmarshalJSON to nil pointer")
	}

	return json.Unmarshal(data, &ss.items)
}
