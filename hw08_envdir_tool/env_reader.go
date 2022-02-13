package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	terminalZeroes = 0x00
	trimCutSet     = "\t "
	eol            = "\n"
)

var (
	ErrDirectoryIsNotReadable       = errors.New("directory is not readable")
	ErrFileInDirectoryIsNotReadable = errors.New("entry is not a file")
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

func readValue(fileName string) (value EnvValue, err error) {
	value = EnvValue{}

	file, err := os.Open(fileName)
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		value.NeedRemove = true
		return
	}

	contentBytes := bytes.ReplaceAll(scanner.Bytes(), []byte{terminalZeroes}, []byte(eol))
	value.Value = strings.TrimRight(string(contentBytes), trimCutSet)

	return
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	result := make(Environment)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, ErrDirectoryIsNotReadable
	}

	for _, e := range entries {
		name := e.Name()

		if e.IsDir() || !e.Type().IsRegular() || strings.Contains(name, "=") {
			continue
		}

		value, err := readValue(fmt.Sprintf("%s/%s", dir, name))
		if err != nil {
			return result, ErrFileInDirectoryIsNotReadable
		}

		result[name] = value
	}

	return result, nil
}
