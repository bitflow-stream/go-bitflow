package sample

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

const (
	tag_equals_rune    = '='
	tag_separator_rune = ' '
	tag_equals         = string(tag_equals_rune)
	tag_separator      = string(tag_separator_rune)
	tag_replacement    = "_"
)

var (
	tagStringEscaper = strings.NewReplacer(
		tag_equals, tag_replacement,
		tag_separator, tag_replacement,
		csv_separator, tag_replacement,
		csv_newline, tag_replacement)
)

type Value float64

type Header struct {
	Fields  []string
	HasTags bool
}

type Sample struct {
	Values []Value
	Time   time.Time
	Tags   map[string]string
}

func (sample *Sample) TagString() string {
	var b bytes.Buffer
	started := false
	for key, value := range sample.Tags {
		if started {
			b.Write([]byte(tag_separator))
		}
		b.Write([]byte(escapeTagString(key)))
		b.Write([]byte(tag_equals))
		b.Write([]byte(escapeTagString(value)))
		started = true
	}
	return b.String()
}

func escapeTagString(str string) string {
	return tagStringEscaper.Replace(str)
}

func (sample *Sample) ParseTagString(tags string) error {
	sample.Tags = make(map[string]string)
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

func (sample *Sample) Check(header Header) error {
	if len(sample.Values) != len(header.Fields) {
		return fmt.Errorf("Unexpected number of values in sample: %v, expected %v",
			len(sample.Values), len(header.Fields))
	}
	return nil
}
