package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type Pointer struct {
	Season  uint `json:"season"`
	Episode uint `json:"episode"`
}

func (p Pointer) String() string {
	return fmt.Sprintf("S%02dE%02d", p.Season, p.Episode)
}

func PointerFromString(s string) (Pointer, error) {
	re := regexp.MustCompile(`S(\d+)E(\d+)`)
	match := re.FindStringSubmatch(s)
	if match == nil {
		return Pointer{}, errors.New("Invalid pointer")
	}

	season, _ := strconv.Atoi(match[1])
	episode, _ := strconv.Atoi(match[2])

	return Pointer{
		Season:  uint(season),
		Episode: uint(episode),
	}, nil
}
