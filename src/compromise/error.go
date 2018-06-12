package compromise

import (
	"fmt"
)

type SourceLocation interface {
	SourceLocation() (string, int, int)
}

type SpecError struct {
	Location SourceLocation
	Message  string
}

func NewSpecError(location SourceLocation, message string) SpecError {
	return SpecError{location, message}
}

func NewSpecErrorf(location SourceLocation, format string, args ...interface{}) SpecError {
	return SpecError{location, fmt.Sprintf(format, args...)}
}
