package main

import (
	"time"
)

type Season struct {
	Number       uint      `json:"number"`
	EpisodeCount uint      `json:"episode_count"`
	Begin        time.Time `json:"begin"`
}
