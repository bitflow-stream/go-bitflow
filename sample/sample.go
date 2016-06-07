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

func (h *Header) Clone(newFields []string) Header {
	return Header{
		HasTags: h.HasTags,
		Fields:  newFields,
	}
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

// TODO make sure this is called at consistent places
func (sample *Sample) Check(header Header) error {
	if len(sample.Values) != len(header.Fields) {
		return fmt.Errorf("Unexpected number of values in sample: %v, expected %v",
			len(sample.Values), len(header.Fields))
	}
	return nil
}

func (sample *Sample) CopyMetadataFrom(other Sample) {
	sample.Time = other.Time
	sample.Tags = make(map[string]string)
	for key, val := range other.Tags {
		sample.Tags[key] = val
	}
}

// All metadata is copied deeply, but values are referencing the old values
func (sample *Sample) Clone() (result Sample) {
	result.CopyMetadataFrom(*sample)
	result.Values = sample.Values
	return
}

func (header *Header) Equals(other *Header) bool {
	switch {
	case header == other:
		return true
	case header == nil && other == nil:
		return true
	case header == nil || other == nil:
		return false
	case header.HasTags != other.HasTags:
		return false
	case header.Fields == nil && other.Fields == nil:
		return true
	case header.Fields == nil || other.Fields == nil:
		return false
	case len(header.Fields) != len(other.Fields):
		return false
	}
	if len(header.Fields) >= 1 {
		// Compare the array backing the Fields slices
		if &(header.Fields[0]) == &(other.Fields[0]) {
			return true
		}
	}
	// Last resort: compare every string pair
	for i := range header.Fields {
		if header.Fields[i] != other.Fields[i] {
			return false
		}
	}
	return true
}
