package main

import (
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	var (
		ok bool

		ss Seasons
		s  *Season
	)

	ss.Put(&Season{4, 0, time.Now()})
	ss.Put(&Season{1, 0, time.Now()})
	ss.Put(&Season{2, 0, time.Now()})
	ss.Put(&Season{6, 0, time.Now()})

	s, ok = ss.Get(2)
	if !ok {
		t.Errorf("Expected season 2, got nothing")
	} else if s.Number != 2 {
		t.Errorf("Expected season 2, got %s", s.Number)
	}

	s, ok = ss.Get(6)
	if !ok {
		t.Errorf("Expected season 6, got nothing")
	} else if s.Number != 6 {
		t.Errorf("Expected season 6, got %s", s.Number)
	}

	s, ok = ss.Get(3)
	if ok {
		t.Errorf("Expected nothing, got season %d", s.Number)
	}
}

func TestGetClosest(t *testing.T) {
	var (
		ok bool

		ss Seasons
		s  *Season
	)

	ss.Put(&Season{1, 0, time.Now()})
	ss.Put(&Season{2, 0, time.Now()})
	ss.Put(&Season{4, 0, time.Now()})

	s, ok = ss.GetClosest(2)
	if !ok {
		t.Errorf("Expected season 2, got nothing")
	} else if s.Number != 2 {
		t.Errorf("Expected season 2, got %s", s.Number)
	}

	s, ok = ss.GetClosest(3)
	if !ok {
		t.Errorf("Expected season 4, got nothing")
	} else if s.Number != 4 {
		t.Errorf("Expected season 4, got %s", s.Number)
	}
}
