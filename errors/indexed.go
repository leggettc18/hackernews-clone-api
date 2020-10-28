package errors

import "fmt"

type indexedError struct {
	cause error
	index int
}

func WithIndex(err error, index int) error {
	return indexedError{cause: err, index: index}
}

func (e indexedError) Error() string {
	return fmt.Sprintf("[%d]: %v", e.index, e.cause)
}
