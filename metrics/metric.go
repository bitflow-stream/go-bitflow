package metrics

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/go-ini/ini"
)

const (
	tagBytes    = 8
	valBytes    = 8
	metricBytes = tagBytes + valBytes
)

type Tag uint64
type Value float64

type Metric struct {
	Tag Tag
	Val Value
}

var Tags map[string]Tag
var TagNames map[Tag]string

func DownloadTagConfig(httpUrl string) error {
	resp, err := http.Get(httpUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := ioutil.TempFile("", "metric_tag_config.ini")
	if err != nil {
		return err
	}
	filename := file.Name()
	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("Failed to close temporary file %v: %v\n", filename, err)
		}
		err = os.Remove(filename)
		if err != nil {
			log.Printf("Failed to delete temporary file %v: %v\n", filename, err)
		}
	}()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return ReadTagConfig(filename)
}

func ReadTagConfig(filename string) error {
	Tags = make(map[string]Tag)
	TagNames = make(map[Tag]string)

	confIni, err := ini.Load(filename)
	if err != nil {
		return err
	}
	section, err := confIni.GetSection("")
	if err != nil {
		return err
	}
	for _, key := range section.Keys() {
		name := key.Name()
		tagVal, err := key.Uint64()
		if err != nil {
			return fmt.Errorf("Error parsing tag value of '%v': %v", name, err)
		}
		tag := Tag(tagVal)
		if _, ok := Tags[name]; ok {
			return fmt.Errorf("Duplicate tag name: %v", name)
		}
		if _, ok := TagNames[tag]; ok {
			return fmt.Errorf("Duplicate tag: %v", tag)
		}
		Tags[name] = tag
		TagNames[tag] = name
	}
	return nil
}

func (metric *Metric) MarshalBinary() (data []byte, err error) {
	if metric == nil {
		return nil, errors.New("Cannot marshal nil *Metric")
	}
	data = make([]byte, metricBytes, metricBytes)
	binary.BigEndian.PutUint64(data, uint64(metric.Tag))
	valBits := math.Float64bits(float64(metric.Val))
	binary.BigEndian.PutUint64(data[tagBytes:], valBits)
	return data, nil
}

func (metric *Metric) UnmarshalBinary(data []byte) error {
	if metric == nil {
		return errors.New("Cannot unmarshal into nil *Metric")
	}
	if len(data) != 16 {
		return fmt.Errorf("Metric.UnmarshalBinary: received %i bytes instead of %i", len(data), metricBytes)
	}
	metric.Tag = Tag(binary.BigEndian.Uint64(data))
	valBits := binary.BigEndian.Uint64(data[tagBytes:])
	metric.Val = Value(math.Float64frombits(valBits))
	return nil
}

func (tag Tag) Name() string {
	if tagName, ok := TagNames[tag]; ok {
		return tagName
	} else {
		return "Unknown Tag"
	}
}

func (tag Tag) String() string {
	return fmt.Sprintf("%v[%v]", tag.Name(), uint64(tag))
}

func (metric *Metric) String() string {
	if metric == nil {
		return "<nil> Metric"
	}
	return fmt.Sprintf("%v: %v", metric.Tag, metric.Val)
}
