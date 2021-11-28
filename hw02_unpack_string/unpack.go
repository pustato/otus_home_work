package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

const escapeRune = '\\'

func Unpack(input string) (string, error) {
	builder := strings.Builder{}

	runes := []rune(input)

	for i := 0; i < len(runes); {
		switch {
		case i+2 < len(runes) && isThreeRuneToken(runes[i], runes[i+1], runes[i+2]):
			count, _ := strconv.Atoi(string(runes[i+2]))
			builder.WriteString(repeatRune(runes[i+1], count))
			i += 3

		case i+1 < len(runes) && isTwoRuneToken(runes[i], runes[i+1]):
			if runes[i] == escapeRune {
				builder.WriteString(string(runes[i+1]))
			} else {
				count, _ := strconv.Atoi(string(runes[i+1]))
				builder.WriteString(repeatRune(runes[i], count))
			}
			i += 2

		case !isControlRune(runes[i]):
			builder.WriteString(string(runes[i]))
			i++

		default:
			return "", ErrInvalidString
		}
	}

	return builder.String(), nil
}

func isControlRune(r rune) bool {
	return r == escapeRune || unicode.IsDigit(r)
}

func isThreeRuneToken(r0, r1, r2 rune) bool {
	return r0 == escapeRune && isControlRune(r1) && unicode.IsDigit(r2)
}

func isTwoRuneToken(r0, r1 rune) bool {
	return (r0 == escapeRune && isControlRune(r1)) || (!isControlRune(r0) && unicode.IsDigit(r1))
}

func repeatRune(r rune, count int) string {
	return strings.Repeat(string(r), count)
}
