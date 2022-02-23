package hw10programoptimization

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var (
	ErrInvalidEmail  = errors.New("invalid email")
	ErrMalformedJSON = errors.New("malformed json")

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	domainSuffix := "." + domain
	scanner := bufio.NewScanner(r)
	result := make(DomainStat)
	user := &User{}

	for scanner.Scan() {
		if err := json.Unmarshal(scanner.Bytes(), user); err != nil {
			return nil, fmt.Errorf("%s: %w", string(scanner.Bytes()), ErrMalformedJSON)
		}

		email := strings.ToLower(user.Email)
		if strings.HasSuffix(email, domainSuffix) {
			idx := strings.Index(email, "@")
			if idx == -1 {
				return nil, fmt.Errorf("%s :%w", user.Email, ErrInvalidEmail)
			}
			result[email[idx+1:]]++
		}

		*user = User{}
	}

	return result, nil
}
