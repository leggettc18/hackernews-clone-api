package errors

import (
	"bytes"
	"fmt"
)

type Errors []error

func (e Errors) Err() error {
	if len(e) == 0 {
		return nil
	}

	return e
}

func (e Errors) Error() string {
	var buf bytes.Buffer

	if n := len(e); n == 1 {
		buf.WriteString("1 error: ")
	} else {
		fmt.Fprintf(&buf, "%d errors: ", n)
	}

	for index, err := range e {
		if index != 0 {
			buf.WriteString("; ")
		}

		buf.WriteString(err.Error())
	}
	return buf.String()
}
