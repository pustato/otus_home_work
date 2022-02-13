package main

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func captureStdout(t *testing.T, f func()) string {
	t.Helper()

	originalStdout := os.Stdout
	read, write, err := os.Pipe()
	if err != nil {
		t.Errorf("capture stdout error: %v", err)
	}
	os.Stdout = write

	var out []byte
	wait := make(chan struct{})
	go func() {
		out, _ = io.ReadAll(read)
		write.Close()
		close(wait)
	}()

	f()

	read.Close()
	os.Stdout = originalStdout

	<-wait
	return string(out)
}

func TestRunCmdExitCode(t *testing.T) {
	env := make(Environment)
	env["EXIT_CODE"] = EnvValue{
		Value: "7",
	}

	code := RunCmd([]string{
		"/bin/bash", "-c", "exit $EXIT_CODE",
	}, env)

	require.Equal(t, 7, code)
}

func TestRunCmdStdout(t *testing.T) {
	env := make(Environment)
	env["SAY"] = EnvValue{
		Value: "Hello!",
	}

	out := captureStdout(t, func() {
		RunCmd([]string{
			"/bin/bash", "-c", "echo $SAY",
		}, env)
	})
	require.Equal(t, "Hello!\n", out)
}

func TestRunCmdReplaceEnv(t *testing.T) {
	envVarName := "SOME_ENV_VARIABLE_THAT_MUST_BE_EMPTY"

	err := os.Setenv(envVarName, "must be empty")
	if err != nil {
		t.Skip("can't modify os environment")
	}

	env := make(Environment)
	env[envVarName] = EnvValue{
		NeedRemove: true,
	}

	out := captureStdout(t, func() {
		RunCmd([]string{
			"/bin/bash", "-c", "echo $" + envVarName,
		}, env)
	})
	require.Equal(t, "\n", out)
}
