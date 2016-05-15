package sample

import "time"

type Value float64

type Header []string

type Sample struct {
	Time   time.Time
	Values []Value
}
