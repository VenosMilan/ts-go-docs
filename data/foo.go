package data

import "time"

// Foo this is comment
type Foo struct {
	Id                int64   `json:"id"`
	Name              *string `json:"name"`
	Sub               *Bar
	Subs              *[]Bar
	LongNameAttribute string
	Time              *time.Time
}

// Bar this is comment
// second row
type Bar struct {
	Id  uint8
	Yes bool
}
