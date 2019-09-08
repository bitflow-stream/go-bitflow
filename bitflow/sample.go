package bitflow

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/antongulenko/golib"
)

const (
	tag_equals_rune    = '='
	tag_separator_rune = ' '
	tag_equals         = string(tag_equals_rune)
	tag_separator      = string(tag_separator_rune)
	tag_replacement    = "_"
)

var (
	TagStringEscaper = strings.NewReplacer(
		tag_equals, tag_replacement,
		tag_separator, tag_replacement,
		string(BinarySeparator), tag_replacement,
		string(CsvSeparator), tag_replacement,
		string(CsvNewline), tag_replacement)
)

// Value is a type alias for float64 and defines the type for metric values.
type Value float64

// String formats the receiving float64 value.
func (v Value) String() string {
	return strconv.FormatFloat(float64(v), 'g', -1, 64)
}

// Header defines the structure of samples that belong to this header.
// When unmarshalling headers and sample, usually one header precedes a number of
// samples. Those samples are defined by the header.
type Header struct {
	// Fields defines the names of the metrics of samples belonging to this header.
	Fields []string
}

// Clone creates a copy of the Header receiver, using a new string-array as
// the header fields.
func (h *Header) Clone(newFields []string) *Header {
	return &Header{
		Fields: newFields,
	}
}

// String returns a human-readable string-representation of the header, including
// all meta-data and field names.
func (h *Header) String() string {
	return fmt.Sprintf("Header %v field(s): %v", len(h.Fields), strings.Join(h.Fields, " "))
}

// BuildIndex creates a dictionary of the header field names to their index in the header
// for optimize access to sample values.
func (h *Header) BuildIndex() map[string]int {
	result := make(map[string]int, len(h.Fields))
	for i, field := range h.Fields {
		result[field] = i
	}
	return result
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

// TagMap returns a copy of the tags stored in the receiving sample.
func (sample *Sample) TagMap() (res map[string]string) {
	sample.lockRead(func() {
		res = make(map[string]string, len(sample.tags))
		for key, val := range sample.tags {
			res[key] = val
		}
	})
	return res
}

// SortedTags returns a slice of key-value tag pairs, sorted by key
func (sample *Sample) SortedTags() (res []KeyValuePair) {
	sample.lockRead(func() {
		res = make([]KeyValuePair, len(sample.orderedTags))
		for i, key := range sample.orderedTags {
			res[i].Key = key
			res[i].Value = sample.tags[key]
		}
	})
	return res
}

// KeyValuePair represents a key-value string pair
type KeyValuePair struct {
	Key   string
	Value string
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

// TagString returns a string representation of all the tags and tag values
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
	return TagStringEscaper.Replace(str)
}

func EncodeTags(tags map[string]string) string {
	var s Sample
	for key, value := range tags {
		s.SetTag(key, value)
	}
	return s.TagString()
}

// ParseTagString parses a string in the format produced by TagString().
// The resulting tags and tag values directly replace the tags inside the
// receiving Sample. Old tags are discarded.
//
// A non-nil error is returned if the format of the input string does not
// follow the defined format (see TagString).
//
// This method is used on freshly created Samples by CsvMarshaller and
// BinaryMarshaller when unmarshalling Samples from the respective format.
func (sample *Sample) ParseTagString(tags string) (err error) {
	sample.lockWrite(func() {
		sample.tags = nil
		fields := strings.FieldsFunc(tags, func(r rune) bool {
			return r == tag_equals_rune || r == tag_separator_rune
		})
		if len(fields)%2 == 1 {
			err = fmt.Errorf("Illegal tags string: %v", tags)
			return
		}
		if len(fields) > 0 {
			sample.tags = make(map[string]string)
			for i := 0; i < len(fields); i += 2 {
				sample.setTag(fields[i], fields[i+1])
			}
		}
	})
	return
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

// CopyTagsFrom adds the tags of the parameter sample to the set of tags already present in the receiving sample.
// Colliding tags are overwritten, but other existing tags are not deleted.
func (sample *Sample) AddTagsFrom(other *Sample) {
	globalCopyMetadataLock.Lock()
	defer globalCopyMetadataLock.Unlock()

	sample.lockWrite(func() {
		other.lockRead(func() {
			for key, val := range other.tags {
				sample.setTag(key, val)
			}
		})
	})
}

// Clone returns a copy of the receiving sample. The metadata (timestamp and tags)
// is copied deeply, but values are referencing the old values. After using this,
// the old Sample should either not be used anymore, or the Values slice in the new
// Sample should be replaced by a new slice.
func (sample *Sample) Clone() *Sample {
	result := &Sample{
		Values: sample.Values,
	}
	result.CopyMetadataFrom(sample)
	return result
}

// DeepClone returns a deep copy of the receiving sample, including the timestamp,
// tags and actual metric values.
func (sample *Sample) DeepClone() *Sample {
	result := &Sample{
		Values: make([]Value, len(sample.Values), cap(sample.Values)),
	}
	result.CopyMetadataFrom(sample)
	copy(result.Values, sample.Values)
	return result
}

// Resize ensures that the Values slice of the sample has the given length.
// If possible, the current Values slice will be reused (shrinking or growing within the limits
// if its capacity). Otherwise, a new slice will be allocated, without copying any values.
// The result value will be true, if the current slice was reused, and false if a new slice was allocated.
func (sample *Sample) Resize(newSize int) bool {
	if cap(sample.Values) >= newSize {
		sample.Values = sample.Values[:newSize] // Grow or shrink, if possible
		return true
	} else {
		sample.Values = make([]Value, newSize)
		return false
	}
}

// Equals compares the receiving header with the argument header and returns true,
// if the two represent the same header. This method tried to optimize the comparison
// by first comparing the header pointers and the length of the Fields slices,
// and pointers to the arrays backing the Fields slices.
// If all the checks fail, the last resort is to compare all the fields string-by-string.
func (h *Header) Equals(other *Header) bool {
	switch {
	case h == nil && other == nil:
		return true
	case h == nil || other == nil:
		return false
	}
	return golib.EqualStrings(h.Fields, other.Fields)
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

// HeaderChecker is a helper type for implementations of SampleSink
// to find out, when the incoming header changes.
type HeaderChecker struct {
	LastHeader *Header
}

// HeaderChanged returns true, if the newHeader parameter represents a different header
// from the last time HeaderChanged was called. The result will also be true for
// the first time this method is called.
func (h *HeaderChecker) HeaderChanged(newHeader *Header) bool {
	changed := !newHeader.Equals(h.LastHeader)
	h.LastHeader = newHeader
	return changed
}

// InitializedHeaderChanged returns true, if the newHeader parameter represents a different header
// from the last time HeaderChanged was called. The first call to this method
// will return false, so this can be used in situations where the header has to be initialized.
func (h *HeaderChecker) InitializedHeaderChanged(newHeader *Header) bool {
	if h.LastHeader == nil {
		h.LastHeader = newHeader
		return false
	}
	changed := !newHeader.Equals(h.LastHeader)
	h.LastHeader = newHeader
	return changed
}

// SampleAndHeader is a convenience type combining pointers to a Sample and a Header.
type SampleAndHeader struct {
	*Sample
	*Header
}

func (s *SampleAndHeader) AddFields(names []string, values []Value) *SampleAndHeader {
	s.Sample = s.Sample.DeepClone()
	s.Sample.Values = append(s.Sample.Values, values...)

	s.Header = s.Header.Clone(append(s.Header.Fields, names...))
	return s
}

const TAG_TEMPLATE_ENV_PREFIX = "ENV_"

func ResolveTagTemplate(template string, missingValues string, sample *Sample) string {
	return TagTemplate{Template: template, MissingValue: missingValues}.Resolve(sample)
}

type TagTemplate struct {
	Template      string // Placeholders like ${xxx} will be replaced by tag values. Values matching ENV_* will be replaced by the environment variable.
	MissingValue  string // Replacement for missing values
	IgnoreEnvVars bool   // Set to true to not treat ENV_ replacement templates specially
}

var templateRegex = regexp.MustCompile("\\${[^{]*}") // Example: ${hello}, ${ENV_HOSTNAME}

func (t TagTemplate) Resolve(sample *Sample) string {
	return templateRegex.ReplaceAllStringFunc(t.Template, func(placeholder string) string {
		placeholder = placeholder[2 : len(placeholder)-1] // Strip the ${} prefix/suffix
		if sample.HasTag(placeholder) {
			return sample.Tag(placeholder)
		} else if strings.HasPrefix(placeholder, TAG_TEMPLATE_ENV_PREFIX) {
			if env, isSet := os.LookupEnv(placeholder[len(TAG_TEMPLATE_ENV_PREFIX):]); isSet {
				return env
			}
		}
		return t.MissingValue
	})
}

// SampleRing is a one-way circular queue of Sample instances. There is no dequeue operation.
// The stored samples can be copied into a correctly ordered slice.
type SampleRing struct {
	samples []*SampleAndHeader
	head    int
}

func NewSampleRing(capacity int) *SampleRing {
	return &SampleRing{
		samples: make([]*SampleAndHeader, capacity),
	}
}

func (r *SampleRing) Push(sample *Sample, header *Header) *SampleRing {
	return r.PushSampleAndHeader(&SampleAndHeader{sample, header})
}

func (r *SampleRing) PushSampleAndHeader(sample *SampleAndHeader) *SampleRing {
	if sample != nil && len(r.samples) > 0 {
		r.samples[r.head] = sample
		if r.head >= len(r.samples)-1 {
			r.head = 0
		} else {
			r.head++
		}
	}
	return r
}

func (r *SampleRing) Len() int {
	if r.IsFull() {
		return len(r.samples)
	}
	return r.head
}

func (r *SampleRing) IsFull() bool {
	if len(r.samples) == 0 {
		return true
	}
	return r.samples[r.head] != nil
}

func (r *SampleRing) Get() []*SampleAndHeader {
	if len(r.samples) == 0 {
		return nil
	}
	res := make([]*SampleAndHeader, r.Len())
	if r.IsFull() {
		copy(res, r.samples[r.head:])
		copy(res[len(r.samples)-r.head:], r.samples[:r.head])
	} else {
		copy(res, r.samples[:r.head])
	}
	return res
}

// SampleHeaderIndex builds a `map[string]int` index of the fields of a Header instance.
// The index is cached and only updated when the header changes. This allows efficient access to specific header fields.
type SampleHeaderIndex struct {
	header *Header
	index  map[string]int
}

func (index *SampleHeaderIndex) Update(header *Header) {
	if index.header.Equals(header) {
		return
	}
	index.header = header
	index.index = make(map[string]int)
	for i, field := range header.Fields {
		index.index[field] = i
	}
}

func (index *SampleHeaderIndex) GetSingle(sample *Sample, field string) (Value, bool) {
	i, ok := index.index[field]
	if !ok {
		return 0, false // No such field
	} else {
		return sample.Values[i], true
	}
}

func (index *SampleHeaderIndex) Get(sample *Sample, header *Header, field string) (Value, bool) {
	index.Update(header)
	return index.GetSingle(sample, field)
}
