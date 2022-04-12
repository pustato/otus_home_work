package application

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

	for _, err := range e.errors {
		builder.WriteString(err.Error())
		builder.WriteString("\n")
	}

	return builder.String()
}
