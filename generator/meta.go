package generator

import "time"

type Meta struct {
	Title      string
	Date       string
	Tags       []string
	Catalogue  string
	ParsedDate time.Time
}
