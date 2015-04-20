package db

import (
	"testing"
	"time"
)

func TestNextPointer(t *testing.T) {
	show := Show{
		Pointer: Pointer{0, 0},
		Seasons: []*Season{
			&Season{4, 24, time.Now()},
			&Season{3, 13, time.Now()},
		},
	}

	next, ok := show.NextPointer()
	if !ok {
		t.Error("Expected a next pointer")
	}

	if next.Season != 3 || next.Episode != 1 {
		t.Errorf("Expected Pointer{3 1}, got %v", next)
	}

	show.Pointer = Pointer{3, 13}
	next, ok = show.NextPointer()
	if !ok {
		t.Error("Expected a next pointer")
	}

	if next.Season != 4 || next.Episode != 1 {
		t.Errorf("Expected Pointer{4 1}, got %v", next)
	}

	show.Pointer = Pointer{4, 24}
	next, ok = show.NextPointer()
	if ok {
		t.Errorf("Expected no next pointer, got %v", next)
	}
}
