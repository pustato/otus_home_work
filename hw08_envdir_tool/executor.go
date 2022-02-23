package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func createEnv(env Environment) []string {
	originalEnv := os.Environ()

	resultEnvMap := make(map[string]string)
	result := make([]string, 0, len(env)+len(originalEnv))
	for _, envEntity := range originalEnv {
		entityArr := strings.SplitN(envEntity, "=", 2)
		if len(entityArr) != 2 {
			continue
		}
		name, value := entityArr[0], entityArr[1]

		if valToReplace, ok := env[name]; ok {
			if valToReplace.NeedRemove {
				continue
			}

			value = valToReplace.Value
		}

		resultEnvMap[name] = value
	}

	for name, val := range env {
		if _, ok := resultEnvMap[name]; ok || val.NeedRemove {
			continue
		}

		resultEnvMap[name] = val.Value
	}

	for name, val := range resultEnvMap {
		result = append(result, fmt.Sprintf("%s=%s", name, val))
	}

	return result
}

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	name := cmd[0]
	args := cmd[1:]

	command := exec.Command(name, args...)

	command.Env = createEnv(env)
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr

	var exitError *exec.ExitError
	if err := command.Run(); errors.As(err, &exitError) {
		return exitError.ExitCode()
	}

	return 0
}
