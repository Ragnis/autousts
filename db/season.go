package db

import (
	"fmt"
	"time"
)

type Season struct {
	Number       uint
	EpisodeCount uint
	Begin        time.Time
}

func (s Season) TableRow() string {
	return fmt.Sprintf("%d\t%d\t%v", s.Number, s.EpisodeCount, s.Begin)
}
