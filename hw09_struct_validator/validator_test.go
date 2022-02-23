package hw09structvalidator

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int8     `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
		Meta   UserMeta `validate:"nested"`
	}

	UserMeta struct {
		From      string `validate:"len:3|in:adv,prt,org"`
		PartnerID int64  `validate:"min:1000"`
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidateWithProgramErrors(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{
				ID:     "fe601a00-1acf-4ded-bb18-0e7f49828649",
				Name:   "John",
				Age:    23,
				Email:  "test@test.test",
				Role:   "stuff",
				Phones: []string{"79991111111", "79991111112", "79991111113"},
				Meta: UserMeta{
					From:      "adv",
					PartnerID: 18841,
				},
			},
			expectedErr: nil,
		},

		{
			in: struct {
				Field uint `validate:"min:22"`
			}{
				Field: 12,
			},
			expectedErr: ErrUnsupportedType,
		},

		{
			in: struct {
				Field int16 `validate:"min:z"`
			}{
				Field: 1,
			},
			expectedErr: ErrInvalidTagArgument,
		},

		{
			in: struct {
				Field int16 `validate:"in"`
			}{
				Field: 1,
			},
			expectedErr: ErrInvalidTagArgumentsCount,
		},

		{
			in: struct {
				Field string `validate:"regexp:z[}\\z"`
			}{
				Field: "fe601a00-1acf-4ded-bb18-0e7f49828649",
			},
			expectedErr: ErrRegexpCompilationError,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("program error case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		in     interface{}
		errors []ValidationError
	}{
		{
			in: Response{
				Code: 403,
				Body: "not found",
			},
			errors: []ValidationError{
				{"Code", ErrInValidationFailed},
			},
		},
		{
			in: App{
				Version: "1.0.22",
			},
			errors: []ValidationError{
				{"Version", ErrLenValidationFailed},
			},
		},
		{
			in: UserMeta{
				From:      "idk from where",
				PartnerID: 999,
			},
			errors: []ValidationError{
				{"From", ErrLenValidationFailed},
				{"From", ErrInValidationFailed},
				{"PartnerID", ErrMinValidationFailed},
			},
		},
		{
			in: UserMeta{
				From:      "adv",
				PartnerID: 0,
			},
			errors: []ValidationError{
				{"PartnerID", ErrMinValidationFailed},
			},
		},
		{
			in: User{
				ID:     "too short",
				Name:   "John",
				Age:    12,
				Email:  "zzzzzzzzzzzzzzzz",
				Role:   "undefined",
				Phones: []string{"79991111111", "555-55-55", "555-55-56"},
				Meta: UserMeta{
					From:      "z",
					PartnerID: 999,
				},
			},
			errors: []ValidationError{
				{"ID", ErrLenValidationFailed},
				{"Age", ErrMinValidationFailed},
				{"Email", ErrRegexValidationFailed},
				{"Role", ErrInValidationFailed},
				{"Phones.1", ErrLenValidationFailed},
				{"Phones.2", ErrLenValidationFailed},
				{"Meta.From", ErrLenValidationFailed},
				{"Meta.From", ErrInValidationFailed},
				{"Meta.PartnerID", ErrMinValidationFailed},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("validation case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)
			require.Error(t, err)

			var validationErrors ValidationErrors
			if !errors.As(err, &validationErrors) {
				t.Error("invalid error type")
			}
			require.Equal(t, len(tt.errors), len(validationErrors))

			for i, actual := range validationErrors {
				expected := tt.errors[i]

				require.Equal(t, expected.Field, actual.Field)
				require.ErrorIs(t, expected.Err, actual.Err)
			}
		})
	}
}
