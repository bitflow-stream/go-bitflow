package sample

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"
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
		string(csv_separator), tag_replacement,
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

	tagsLock    sync.RWMutex
	tags        map[string]string
	orderedTags []string // All keys from tags, with consistent ordering
}

func (sample *Sample) lockRead(do func()) {
	sample.tagsLock.RLock()
	defer sample.tagsLock.RUnlock()
	do()
}

func (sample *Sample) lockWrite(do func()) {
	sample.tagsLock.Lock()
	defer sample.tagsLock.Unlock()
	do()
}

func (sample *Sample) HasTag(name string) (ok bool) {
	sample.lockRead(func() {
		_, ok = sample.tags[name]
	})
	return
}

func (sample *Sample) SetTag(name, value string) {
	sample.lockWrite(func() {
		if sample.tags == nil {
			sample.tags = make(map[string]string)
		}
		if _, exists := sample.tags[name]; !exists {
			sample.orderedTags = append(sample.orderedTags, name)
			sort.Strings(sample.orderedTags)
		}
		sample.tags[name] = value
	})
}

func (sample *Sample) DeleteTag(name string) {
	sample.lockWrite(func() {
		if _, ok := sample.tags[name]; ok {
			l := len(sample.orderedTags)
			index := sort.SearchStrings(sample.orderedTags, name)
			if index < l && index >= 0 && sample.orderedTags[index] == name {
				// Delete element at index
				sample.orderedTags = append(sample.orderedTags[:index], sample.orderedTags[index+1:]...)
			}
			delete(sample.tags, name)
		}
	})
}

func (sample *Sample) Tag(name string) (value string) {
	sample.lockRead(func() {
		value = sample.tags[name]
	})
	return
}

func (sample *Sample) NumTags() (l int) {
	sample.lockRead(func() {
		l = len(sample.tags)
	})
	return
}

func (sample *Sample) TagString() (res string) {
	sample.lockRead(func() {
		var b bytes.Buffer
		started := false
		for _, key := range sample.orderedTags {
			value := sample.tags[key]
			if started {
				b.Write([]byte(tag_separator))
			}
			b.Write([]byte(escapeTagString(key)))
			b.Write([]byte(tag_equals))
			b.Write([]byte(escapeTagString(value)))
			started = true
		}
		res = b.String()
	})
	return
}

func escapeTagString(str string) string {
	return tagStringEscaper.Replace(str)
}

func (sample *Sample) ParseTagString(tags string) (err error) {
	sample.lockWrite(func() {
		sample.tags = make(map[string]string)
		fields := strings.FieldsFunc(tags, func(r rune) bool {
			return r == tag_equals_rune || r == tag_separator_rune
		})
		if len(fields)%2 == 1 {
			err = fmt.Errorf("Illegal tags string: %v", tags)
			return
		}
		for i := 0; i < len(fields); i += 2 {
			sample.SetTag(fields[i], fields[i+1])
		}
	})
	return
}

// This must be called in the Sample() method of every MetricSink implementation
func (sample *Sample) Check(header Header) error {
	if len(sample.Values) != len(header.Fields) {
		return fmt.Errorf("Unexpected number of values in sample: %v, expected %v",
			len(sample.Values), len(header.Fields))
	}
	return nil
}

func (sample *Sample) CopyMetadataFrom(other Sample) {
	sample.lockWrite(func() {
		sample.Time = other.Time
		sample.tags = make(map[string]string, len(other.tags))
		sample.orderedTags = make([]string, len(other.orderedTags))
		copy(sample.orderedTags, other.orderedTags)
		for key, val := range other.tags {
			sample.tags[key] = val
		}
	})
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

type SampleMetadata struct {
	Time        time.Time
	Tags        map[string]string
	orderedTags []string
}

func (sample *Sample) Metadata() SampleMetadata {
	return SampleMetadata{
		Time:        sample.Time,
		Tags:        sample.tags,
		orderedTags: sample.orderedTags,
	}
}

func (meta *SampleMetadata) NewSample(values []Value) Sample {
	return Sample{Values: values, Time: meta.Time, tags: meta.Tags, orderedTags: meta.orderedTags}
}
