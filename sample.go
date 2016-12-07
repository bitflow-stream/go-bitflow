package bitflow

import (
	"bytes"
	"errors"
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
		string(BinarySeparator), tag_replacement,
		string(CsvSeparator), tag_replacement,
		string(CsvNewline), tag_replacement)
)

// Value is a type alias for float64 and defines the type metric values.
type Value float64

// Header defines the structure of samples that belong to this header.
// When unmarshalling headers and sample, usually one header preceds a number of
// samples. Those samples are defined by the header.
type Header struct {
	// Fields defines the names of the metrics of samples belonging to this header.
	Fields []string

	// HasTags is true, if the samples belonging to this header include tags.
	// Tags are optional when marshalling/unmarshalling samples.
	HasTags bool
}

// Clone creates a copy of the Header receiver, using a new string-array as
// the header fields.
func (h *Header) Clone(newFields []string) *Header {
	return &Header{
		HasTags: h.HasTags,
		Fields:  newFields,
	}
}

// Sample contains an array of Values, a timestamp, and a string-to-string map of tags.
// The values are explained by the header belonging to this sample. There is no direct
// pointer from the sample to the header, so the header must be known from the context.
// In other words, the Sample should always be passed along with the header it belongs to.
// Without the header, the meaning of the Value array is not defined.
// The Values slice and the timestamp an be accessed and modified directly, but the tags
// map should only be manipulated through methods to ensure concurrency safe map operations.
// The Values and Tags should be modified by only one goroutine at a time.
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

// HasTag returns true if the receiving Sample includes a tag with the given name.
func (sample *Sample) HasTag(name string) (ok bool) {
	sample.lockRead(func() {
		_, ok = sample.tags[name]
	})
	return
}

// SetTag sets the tag of the given name to the given value in the receiving Sample.
func (sample *Sample) SetTag(name, value string) {
	sample.lockWrite(func() {
		sample.setTag(name, value)
	})
}

func (sample *Sample) setTag(name, value string) {
	if sample.tags == nil {
		sample.tags = make(map[string]string)
	}
	if _, exists := sample.tags[name]; !exists {
		sample.orderedTags = append(sample.orderedTags, name)
		sort.Strings(sample.orderedTags)
	}
	sample.tags[name] = value
}

// DeleteTag deleted the given tag from the receiving Sample.
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

// Tag returns the value of the given tag inside the receiving Sample.
// If the tag is not defined in the sample, an empty string is returned.
// HasTag can be used to find out if a tag is defined or not.
func (sample *Sample) Tag(name string) (value string) {
	sample.lockRead(func() {
		value = sample.tags[name]
	})
	return
}

// NumTags returns the number of tags defined in the receiving Sample.
func (sample *Sample) NumTags() (l int) {
	sample.lockRead(func() {
		l = len(sample.tags)
	})
	return
}

// TagString returns a string representationg of all the tags and tag values
// in the receiving Sample. This representation is used for marshalling by the
// CsvMarshaller and BinaryMarshaller. The format is a space-separated string
// of key-value pairs separated by '=' characters.
//
// Example:
//   tag1=value1 tag2=value2
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

// ParseTagString parses a string in the format produced by TagString().
// The resulting tags and tag values directly replace the tags inside the
// receiving Sample. Old tags are discarded.
//
// A non-nil error is returned if the format of the input string does not
// follow the defined format (see TagString).
//
// This method is used on freshly created Samples by CsvMarshaller and
// BniaryMarshaller when unmarshalling Samples from the respective format.
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
			sample.setTag(fields[i], fields[i+1])
		}
	})
	return
}

// Check returns a non-nil error when the receiving sample does not seem to belong
// to the header argument. The number of fields in the header are compared with
// the number of values in the sample.
// This is a sanity-check for ensuring correct format of files or TCP connections.
// This check must be called in the Sample() method of every MetricSink implementation.
func (sample *Sample) Check(header *Header) error {
	if sample == nil || header == nil {
		return errors.New("The sample or header is nil")
	}
	if len(sample.Values) != len(header.Fields) {
		return fmt.Errorf("Unexpected number of values in sample: %v, expected %v",
			len(sample.Values), len(header.Fields))
	}
	return nil
}

// HACK global lock to avoid potential deadlock in CopyMetadataFrom() because of acquiring 2 locks.
// TODO This sacrifices performance, find better solution.
var globalCopyMetadataLock sync.Mutex

// CopyMetadataFrom copies the timestamp and tags from the argument Sample into the
// receiving Sample. All previous tags in the receiving Sample are discarded.
func (sample *Sample) CopyMetadataFrom(other *Sample) {
	globalCopyMetadataLock.Lock()
	defer globalCopyMetadataLock.Unlock()

	sample.lockWrite(func() {
		other.lockRead(func() {
			sample.Time = other.Time
			sample.tags = make(map[string]string, len(other.tags))
			sample.orderedTags = make([]string, len(other.orderedTags))
			copy(sample.orderedTags, other.orderedTags)
			for key, val := range other.tags {
				sample.tags[key] = val
			}
		})
	})
}

// Clone returns a copy of the receiving sample. The metadata (timestamp and tags)
// is copied deeply, but values are referencing the old values. After using this,
// the old Sample should either not be used anymore, or the Values slice in the new
// Sample should be replaced by a new slice.
func (sample *Sample) Clone() (result *Sample) {
	result = &Sample{
		Values: sample.Values,
	}
	result.CopyMetadataFrom(sample)
	return
}

// Equals compares the receiving header with the argument header and returns true,
// if the two represent the same header. This method tried to optimize the comparison
// by first comparing the header pointers, values for HasTags, the length of the Fields slices,
// and pointers to the arrays backing the Fields slices.
// If all the checks fail, the last resort is to compare all the fields string-by-string.
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

// SampleMetadata is a helper type containing the timestamp and the tags of a Sample.
// It can be used to store samples in cases where the actual Values of the Sample are
// not relevant, or stored in a different location. The timestamp can be accessed
// directly, but the tags are hidden to ensure consistency. If a SampleMetadata instance
// is required with specific tags, a Sample with those tags should be created, and
// Metadata() should be called on that Sample to create the desired SampleMetadata instance.
type SampleMetadata struct {
	Time time.Time

	tags        map[string]string
	orderedTags []string
}

// Metadata returns an instance of SampleMetadata containing the tags and timestamp
// of the receiving Sample.
func (sample *Sample) Metadata() *SampleMetadata {
	return &SampleMetadata{
		Time:        sample.Time,
		tags:        sample.tags,
		orderedTags: sample.orderedTags,
	}
}

// NewSample returns a new Sample instances containing the tags and timestamp defined in the
// receiving SampleMetadata instance and the Values given as argument. The metadata
// is copied deeply, so the resulting Sample can be modified independently of the receiving
// SampleMetadata instance.
func (meta *SampleMetadata) NewSample(values []Value) *Sample {
	sample := &Sample{
		Values:      values,
		Time:        meta.Time,
		tags:        make(map[string]string, len(meta.tags)),
		orderedTags: make([]string, len(meta.orderedTags)),
	}
	copy(sample.orderedTags, meta.orderedTags)
	for tag, val := range meta.tags {
		sample.tags[tag] = val
	}
	return sample
}
