package reg

import (
	"fmt"
	"strconv"
	"time"
)

func ParameterError(name string, err error) error {
	return fmt.Errorf("Failed to parse '%v' parameter: %v", name, err)
}

type ParameterParser interface {
	String() string
	ParsePrimitive(val string) (interface{}, error)
	ParseList(val []string) (interface{}, error)
	ParseMap(val map[string]string) (interface{}, error)
	CorrectType(val interface{}) bool
}

type internalPrimitiveParser interface {
	String() string
	ParsePrimitive(val string) (interface{}, error)
	ParseList(val []string) (interface{}, error)
	ParseMap(val map[string]string) (interface{}, error)
	CorrectPrimitiveType(val interface{}) bool
	CorrectListType(val interface{}) bool
	CorrectMapType(val interface{}) bool
}

func Map(primitive ParameterParser) ParameterParser {
	return mapParser{primitive: extractPrimitive(primitive)}
}

func List(primitive ParameterParser) ParameterParser {
	return listParser{primitive: extractPrimitive(primitive)}
}

func String() ParameterParser {
	return primitiveParser{primitive: stringParser{}}
}

func Int() ParameterParser {
	return primitiveParser{primitive: intParser{}}
}

func Float() ParameterParser {
	return primitiveParser{primitive: floatParser{}}
}

func Duration() ParameterParser {
	return primitiveParser{primitive: durationParser{}}
}

func Bool() ParameterParser {
	return primitiveParser{primitive: boolParser{}}
}

func unexpectedType(parser ParameterParser, found string) error {
	return fmt.Errorf("Expected type %v, but found %v", parser.String(), found)
}

func extractPrimitive(parser ParameterParser) internalPrimitiveParser {
	if primParser, ok := parser.(primitiveParser); ok {
		return primParser.primitive
	} else {
		panic(fmt.Sprintf("Parser of type %T cannot be wrapped in list or map", parser))
	}
}

type stringParser struct {
}

func (stringParser) String() string {
	return "string"
}
func (stringParser) ParsePrimitive(val string) (interface{}, error) {
	return val, nil
}
func (stringParser) ParseList(val []string) (interface{}, error) {
	return val, nil
}
func (stringParser) ParseMap(val map[string]string) (interface{}, error) {
	return val, nil
}
func (stringParser) CorrectPrimitiveType(val interface{}) (res bool) {
	_, res = val.(string)
	return
}
func (stringParser) CorrectListType(val interface{}) (res bool) {
	_, res = val.([]string)
	return
}
func (stringParser) CorrectMapType(val interface{}) (res bool) {
	_, res = val.(map[string]string)
	return
}

type intParser struct {
}

func (intParser) String() string {
	return "int"
}
func (intParser) ParsePrimitive(val string) (interface{}, error) {
	return strconv.Atoi(val)
}
func (p intParser) ParseList(val []string) (interface{}, error) {
	result := make([]int, len(val))
	for i, v := range val {
		parsed, err := p.ParsePrimitive(v)
		if err != nil {
			return nil, err
		}
		result[i] = parsed.(int)
	}
	return result, nil
}
func (p intParser) ParseMap(val map[string]string) (interface{}, error) {
	result := make(map[string]int, len(val))
	for key, v := range val {
		parsed, err := p.ParsePrimitive(v)
		if err != nil {
			return nil, err
		}
		result[key] = parsed.(int)
	}
	return result, nil
}
func (intParser) CorrectPrimitiveType(val interface{}) (res bool) {
	_, res = val.(int)
	return
}
func (intParser) CorrectListType(val interface{}) (res bool) {
	_, res = val.([]int)
	return
}
func (intParser) CorrectMapType(val interface{}) (res bool) {
	_, res = val.(map[string]int)
	return
}

type floatParser struct {
}

func (floatParser) String() string {
	return "float"
}
func (floatParser) ParsePrimitive(val string) (interface{}, error) {
	return strconv.ParseFloat(val, 64)
}
func (p floatParser) ParseList(val []string) (interface{}, error) {
	result := make([]float64, len(val))
	for i, v := range val {
		parsed, err := p.ParsePrimitive(v)
		if err != nil {
			return nil, err
		}
		result[i] = parsed.(float64)
	}
	return result, nil
}
func (p floatParser) ParseMap(val map[string]string) (interface{}, error) {
	result := make(map[string]float64, len(val))
	for key, v := range val {
		parsed, err := p.ParsePrimitive(v)
		if err != nil {
			return nil, err
		}
		result[key] = parsed.(float64)
	}
	return result, nil
}
func (floatParser) CorrectPrimitiveType(val interface{}) (res bool) {
	_, res = val.(float64)
	return
}
func (floatParser) CorrectListType(val interface{}) (res bool) {
	_, res = val.([]float64)
	return
}
func (floatParser) CorrectMapType(val interface{}) (res bool) {
	_, res = val.(map[string]float64)
	return
}

type durationParser struct {
}

func (durationParser) String() string {
	return "duration"
}
func (durationParser) ParsePrimitive(val string) (interface{}, error) {
	return time.ParseDuration(val)
}
func (p durationParser) ParseList(val []string) (interface{}, error) {
	result := make([]time.Duration, len(val))
	for i, v := range val {
		parsed, err := p.ParsePrimitive(v)
		if err != nil {
			return nil, err
		}
		result[i] = parsed.(time.Duration)
	}
	return result, nil
}
func (p durationParser) ParseMap(val map[string]string) (interface{}, error) {
	result := make(map[string]time.Duration, len(val))
	for key, v := range val {
		parsed, err := p.ParsePrimitive(v)
		if err != nil {
			return nil, err
		}
		result[key] = parsed.(time.Duration)
	}
	return result, nil
}
func (durationParser) CorrectPrimitiveType(val interface{}) (res bool) {
	_, res = val.(time.Duration)
	return
}
func (durationParser) CorrectListType(val interface{}) (res bool) {
	_, res = val.([]time.Duration)
	return
}
func (durationParser) CorrectMapType(val interface{}) (res bool) {
	_, res = val.(map[string]time.Duration)
	return
}

type boolParser struct {
}

func (boolParser) String() string {
	return "bool"
}
func (boolParser) ParsePrimitive(val string) (interface{}, error) {
	return strconv.ParseBool(val)
}
func (p boolParser) ParseList(val []string) (interface{}, error) {
	result := make([]bool, len(val))
	for i, v := range val {
		parsed, err := p.ParsePrimitive(v)
		if err != nil {
			return nil, err
		}
		result[i] = parsed.(bool)
	}
	return result, nil
}
func (p boolParser) ParseMap(val map[string]string) (interface{}, error) {
	result := make(map[string]bool, len(val))
	for key, v := range val {
		parsed, err := p.ParsePrimitive(v)
		if err != nil {
			return nil, err
		}
		result[key] = parsed.(bool)
	}
	return result, nil
}
func (boolParser) CorrectPrimitiveType(val interface{}) (res bool) {
	_, res = val.(bool)
	return
}
func (boolParser) CorrectListType(val interface{}) (res bool) {
	_, res = val.([]bool)
	return
}
func (boolParser) CorrectMapType(val interface{}) (res bool) {
	_, res = val.(map[string]bool)
	return
}

type mapParser struct {
	primitive internalPrimitiveParser
}

func (p mapParser) String() string {
	return fmt.Sprintf("map(%v)", p.primitive)
}
func (p mapParser) ParsePrimitive(val string) (interface{}, error) {
	return nil, unexpectedType(p, "primitive")
}
func (p mapParser) ParseList(val []string) (interface{}, error) {
	return nil, unexpectedType(p, "list")
}
func (p mapParser) ParseMap(val map[string]string) (interface{}, error) {
	return p.primitive.ParseMap(val)
}
func (p mapParser) CorrectType(val interface{}) bool {
	return p.primitive.CorrectMapType(val)
}

type listParser struct {
	primitive internalPrimitiveParser
}

func (p listParser) String() string {
	return fmt.Sprintf("list(%v)", p.primitive)
}
func (p listParser) ParsePrimitive(val string) (interface{}, error) {
	return nil, unexpectedType(p, "primitive")
}
func (p listParser) ParseList(val []string) (interface{}, error) {
	return p.primitive.ParseList(val)
}
func (p listParser) ParseMap(val map[string]string) (interface{}, error) {
	return nil, unexpectedType(p, "map")
}
func (p listParser) CorrectType(val interface{}) bool {
	return p.primitive.CorrectListType(val)
}

type primitiveParser struct {
	primitive internalPrimitiveParser
}

func (p primitiveParser) String() string {
	return p.primitive.String()
}
func (p primitiveParser) ParsePrimitive(val string) (interface{}, error) {
	return p.primitive.ParsePrimitive(val)
}
func (p primitiveParser) ParseList(val []string) (interface{}, error) {
	return nil, unexpectedType(p, "list")
}
func (p primitiveParser) ParseMap(val map[string]string) (interface{}, error) {
	return nil, unexpectedType(p, "map")
}
func (p primitiveParser) CorrectType(val interface{}) bool {
	return p.primitive.CorrectPrimitiveType(val)
}
