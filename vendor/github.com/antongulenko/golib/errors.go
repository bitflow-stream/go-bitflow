package golib

import (
	"bytes"
	"fmt"
)

type MultiError []error

func (err MultiError) NilOrError() error {
	if len(err) == 0 {
		return nil
	} else if len(err) == 1 {
		return err[0]
	}
	return err
}

func (err *MultiError) Add(errOrNil error) {
	if err != nil && errOrNil != nil {
		*err = append(*err, errOrNil)
	}
}

func (err MultiError) Error() string {
	switch len(err) {
	case 0:
		return "No error"
	case 1:
		return err[0].Error()
	default:
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "Multiple errors:\n")
		for i, e := range err {
			if i > 0 {
				fmt.Fprintf(&buf, "\n")
			}
			fmt.Fprintf(&buf, "\t%v. %v", i+1, e)
		}
		return buf.String()
	}
}
