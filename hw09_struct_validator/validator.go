package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	validateTag = "validate"
	tagLen      = "len"
	tagRegexp   = "regexp"
	tagIn       = "in"
	tagMin      = "min"
	tagMax      = "max"
	tagNested   = "nested"
)

type tag struct {
	name string
	args []string
}

func (t *tag) argsInt() ([]int64, error) {
	result := make([]int64, len(t.args))

	for i, v := range t.args {
		iv, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("tag %s argument #%d: %w", t.name, i, ErrInvalidTagArgument)
		}

		result[i] = int64(iv)
	}

	return result, nil
}

func (t *tag) argInt(i int) (int64, error) {
	if len(t.args) < i+1 {
		return 0, fmt.Errorf("tag %s requires at least %d arg: %w", t.name, i+1, ErrInvalidTagArgumentsCount)
	}

	iv, err := strconv.Atoi(t.args[i])
	if err != nil {
		return 0, fmt.Errorf("tag %s argument #%d: %w", t.name, i+1, ErrInvalidTagArgument)
	}

	return int64(iv), nil
}

var (
	ErrValueIsNotAStructure     = errors.New("value is not a structure")
	ErrUnsupportedType          = errors.New("unsupported type")
	ErrInvalidTagArgument       = errors.New("invalid tag argument")
	ErrInvalidTagArgumentsCount = errors.New("invalid tag arguments count")
	ErrRegexpCompilationError   = errors.New("regexp compilation error")

	ErrMinValidationFailed   = errors.New("min validation failed")
	ErrMaxValidationFailed   = errors.New("max validation failed")
	ErrInValidationFailed    = errors.New("in validation failed")
	ErrLenValidationFailed   = errors.New("len validation failed")
	ErrRegexValidationFailed = errors.New("regex validation failed")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	builder := strings.Builder{}

	for _, e := range v {
		builder.Write([]byte(e.Field))
		builder.Write([]byte(": "))
		builder.Write([]byte(e.Err.Error()))
		builder.Write([]byte("\n"))
	}

	return builder.String()
}

func Validate(v interface{}) error {
	return validateStruct("", v)
}

func validateStruct(fieldPrefix string, v interface{}) error {
	t := reflect.TypeOf(v)
	if t.Kind() != reflect.Struct {
		return ErrValueIsNotAStructure
	}

	value := reflect.ValueOf(v)
	numField := t.NumField()
	var result ValidationErrors

	for i := 0; i < numField; i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		tagsStr, ok := field.Tag.Lookup(validateTag)
		if !ok {
			continue
		}

		tags := parseTags(tagsStr)

		fieldType := field.Type
		fieldVal := value.Field(i)
		fieldName := field.Name
		if fieldPrefix != "" {
			fieldName = strings.Join([]string{fieldPrefix, fieldName}, ".")
		}

		var err error
		//nolint:exhaustive
		switch fieldType.Kind() {
		case reflect.Slice, reflect.Array:
			err = validateSlice(fieldName, fieldVal, tags)
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			err = validateInt(fieldName, fieldVal, tags)
		case reflect.String:
			err = validateString(fieldName, fieldVal, tags)
		case reflect.Struct:
			if len(tags) == 1 && tags[0].name == tagNested {
				if fieldVal.CanInterface() {
					err = validateStruct(fieldName, fieldVal.Interface())
				} else {
					err = ErrUnsupportedType
				}
			}
		default:
			err = ErrUnsupportedType
		}

		if err != nil {
			var ve ValidationErrors
			if errors.As(err, &ve) {
				result = append(result, ve...)
			} else {
				return err
			}
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func validateSlice(fieldName string, v reflect.Value, tags []*tag) error {
	var result ValidationErrors

	fieldLen := v.Len()
	if fieldLen == 0 {
		return nil
	}
	kind := v.Index(0).Kind()

	for i := 0; i < fieldLen; i++ {
		fieldName := fmt.Sprintf("%s.%d", fieldName, i)
		value := v.Index(i)

		var err error
		//nolint:exhaustive
		switch kind {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
			err = validateInt(fieldName, value, tags)
		case reflect.String:
			err = validateString(fieldName, value, tags)
		default:
			err = ErrUnsupportedType
		}

		if err != nil {
			var ve ValidationErrors
			if errors.As(err, &ve) {
				result = append(result, ve...)
			} else {
				return err
			}
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func validateInt(fieldName string, v reflect.Value, tags []*tag) error {
	value := v.Int()
	var result ValidationErrors

	for _, t := range tags {
		args, err := t.argsInt()
		if err != nil {
			return fmt.Errorf("invalid tag arguments for field %s: %w", fieldName, err)
		}
		if len(args) == 0 {
			return fmt.Errorf("tag %s requires at least 1 argument: %w", t.name, ErrInvalidTagArgumentsCount)
		}

		switch t.name {
		case tagMin:
			if value < args[0] {
				result = append(result, ValidationError{fieldName, ErrMinValidationFailed})
			}
		case tagMax:
			if value > args[0] {
				result = append(result, ValidationError{fieldName, ErrMaxValidationFailed})
			}
		case tagIn:
			found := false
			for _, iv := range args {
				if value == iv {
					found = true
					break
				}
			}
			if !found {
				result = append(result, ValidationError{Field: fieldName, Err: ErrInValidationFailed})
			}
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func validateString(fieldName string, v reflect.Value, tags []*tag) error {
	value := v.String()
	var result ValidationErrors

	for _, t := range tags {
		if len(t.args) == 0 {
			return fmt.Errorf("tag %s requires at laest 1 argument: %w", t.name, ErrInvalidTagArgumentsCount)
		}

		switch t.name {
		case tagLen:
			arg, err := t.argInt(0)
			if err != nil {
				return fmt.Errorf("invalid tag arguments for field %s: %w", fieldName, err)
			}
			if int64(len(value)) != arg {
				result = append(result, ValidationError{Field: fieldName, Err: ErrLenValidationFailed})
			}

		case tagIn:
			found := false
			for _, iv := range t.args {
				if value == iv {
					found = true
					break
				}
			}
			if !found {
				result = append(result, ValidationError{Field: fieldName, Err: ErrInValidationFailed})
			}

		case tagRegexp:
			rx, err := regexp.Compile(t.args[0])
			if err != nil {
				return ErrRegexpCompilationError
			}
			if !rx.MatchString(value) {
				result = append(result, ValidationError{Field: fieldName, Err: ErrRegexValidationFailed})
			}
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}

func parseTags(t string) []*tag {
	result := make([]*tag, 0, strings.Count(t, "|")+1)

	for _, tt := range strings.Split(t, "|") {
		parts := strings.SplitN(tt, ":", 2)
		tagInstance := &tag{name: parts[0]}
		if len(parts) == 2 {
			tagInstance.args = strings.Split(parts[1], ",")
		}

		result = append(result, tagInstance)
	}

	return result
}
