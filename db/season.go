package db

import (
	"time"
)

type Season struct {
	Number       uint
	EpisodeCount uint
	Begin        time.Time
}
