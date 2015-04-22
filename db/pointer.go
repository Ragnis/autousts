package db

import (
	"fmt"
)

type Pointer struct {
	Season  uint
	Episode uint
}

func (p Pointer) String() string {
	return fmt.Sprintf("S%02dE%02d", p.Season, p.Episode)
}
