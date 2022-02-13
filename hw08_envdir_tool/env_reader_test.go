package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type testEnv = map[string]string

func prepareTestDir(t *testing.T, env testEnv) string {
	t.Helper()

	dir, err := os.MkdirTemp(os.TempDir(), "env_reader")
	if err != nil {
		t.Errorf("error with create test dir: %v", err)
	}

	for name, value := range env {
		err := os.WriteFile(fmt.Sprintf("%s/%s", dir, name), []byte(value), 0o700)
		if err != nil {
			t.Errorf("error with creating test file: %v", err)
		}
	}

	t.Cleanup(func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Errorf("error with clean up test data: %v", err)
		}
	})

	return dir
}

func createStringWithTermZeroes(lines []string) string {
	return strings.Join(lines, string([]byte{0x00}))
}

func TestReadDirSuccess(t *testing.T) {
	testVariables := make(testEnv)
	testVariables["NORMAL_VAR"] = "value"
	testVariables["TRIMMED_VAR"] = " \t value      \t\t"
	testVariables["VAR_WITH=WRONG_NAME"] = "value"
	testVariables["VAR_TO_DELETE"] = ""
	testVariables["MULTILINE_VAR"] = "line1\nline2\nline3"
	testVariables["VAR_WITH_TERM_ZEROES"] = createStringWithTermZeroes([]string{"line1", "line2", "line3"})

	dir := prepareTestDir(t, testVariables)

	env, err := ReadDir(dir)
	require.NoError(t, err)
	require.Equal(t, len(env), len(testVariables)-1)

	_, ok := env["VAR_WITH=WRONG_NAME"]
	require.False(t, ok)

	require.Equal(t, env["NORMAL_VAR"].Value, "value")
	require.Equal(t, env["TRIMMED_VAR"].Value, " \t value")
	require.True(t, env["VAR_TO_DELETE"].NeedRemove)
	require.Equal(t, env["MULTILINE_VAR"].Value, "line1")
	require.Equal(t, env["VAR_WITH_TERM_ZEROES"].Value, "line1\nline2\nline3")
}

func TestReadDirError(t *testing.T) {
	_, err := ReadDir("/this_dir_is_not_exists")
	require.ErrorIs(t, err, ErrDirectoryIsNotReadable)
}
