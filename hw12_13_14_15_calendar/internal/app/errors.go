package app

import (
	"errors"
	"strings"
)

var (
	ErrTitleTooLong                  = errors.New("title is too long")
	ErrTimeEndMustBeGreaterThanStart = errors.New("time end must be greater than time start")
	ErrTimeIsBusy                    = errors.New("time is busy")
	ErrEventIsNotExists              = errors.New("event is not exists")
)

type ValidationErrors struct {
	errors []error
}

func (e *ValidationErrors) Errors() []error {
	return e.errors
}

func (e *ValidationErrors) Error() string {
	builder := strings.Builder{}

	for i, err := range e.errors {
		if i != 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(err.Error())
	}

	return builder.String()
}
