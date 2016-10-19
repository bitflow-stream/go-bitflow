package golib

import "fmt"

type StringSlice []string

func (i *StringSlice) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *StringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}
