package sample

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

type Value float64

var (
	tagStringEscaper   = strings.NewReplacer(",", "_", "=", "_", " ", "_", "\n", "_")
	tag_equals_rune    = '='
	tagEquals          = []byte("=")
	tag_separator_rune = ' '
	tagSeparator       = []byte(" ")
)

type Header struct {
	Fields  []string
	HasTags bool
}

type Sample struct {
	Values []Value
	Time   time.Time
	Tags   map[string]string
}

func NewSample(values []Value, time time.Time) Sample {
	return Sample{
		Values: values,
		Time:   time,
		Tags:   make(map[string]string),
	}
}

func (sample *Sample) TagString() string {
	var b bytes.Buffer
	started := false
	for key, value := range sample.Tags {
		if started {
			b.Write(tagSeparator)
		}
		b.Write([]byte(escapeTagString(key)))
		b.Write(tagEquals)
		b.Write([]byte(escapeTagString(value)))
		started = true
	}
	return b.String()
}

func escapeTagString(str string) string {
	return tagStringEscaper.Replace(str)
}

func (sample *Sample) ParseTagString(tags string) error {
	fields := strings.FieldsFunc(tags, func(r rune) bool {
		return r == tag_equals_rune || r == tag_separator_rune
	})
	if len(fields)%2 == 1 {
		return fmt.Errorf("Illegal tags string: %v", tags)
	}
	for i := 0; i < len(fields); i += 2 {
		sample.Tags[fields[i]] = fields[i+1]
	}
	return nil
}
