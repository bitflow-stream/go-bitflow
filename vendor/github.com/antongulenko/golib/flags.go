package golib

import (
	"fmt"
	"strings"
)

type StringSlice []string

func (i *StringSlice) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *StringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

const KeyValueSeparator = "="

type KeyValueStringSlice struct {
	Keys   []string
	Values []string
}

func (i *KeyValueStringSlice) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *KeyValueStringSlice) Set(value string) error {
	parts := strings.SplitN(value, KeyValueSeparator, 2)
	if len(parts) != 2 {
		return fmt.Errorf("Wrong format. Need key=value, got " + value)
	}
	i.Keys = append(i.Keys, parts[0])
	i.Values = append(i.Values, parts[1])
	return nil
}
