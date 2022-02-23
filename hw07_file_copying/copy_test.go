package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type testData struct {
	tempDir, testFileName string
	testBytes             []byte
	fileSize              int64
}

func createTempDir(t *testing.T) string {
	t.Helper()

	dir := os.TempDir() + "/copy_test"
	err := os.Mkdir(dir, os.FileMode(0o700))
	if err != nil {
		t.Error("create test dir", err)
	}

	return dir
}

func createInputFile(t *testing.T, tempDir string, bytes []byte) string {
	t.Helper()

	f, err := os.CreateTemp(tempDir, "copy_test")
	if err != nil {
		t.Error("create test file", err)
	}
	_, err = f.Write(bytes)
	if err != nil {
		t.Error("write test file", err)
	}
	f.Close()

	return f.Name()
}

func generateTestData(t *testing.T, fileSize int64) *testData {
	t.Helper()

	testBytes := make([]byte, fileSize)
	rand.Read(testBytes)

	tempDir := createTempDir(t)
	testFileName := createInputFile(t, tempDir, testBytes)

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return &testData{
		tempDir,
		testFileName,
		testBytes,
		fileSize,
	}
}

func TestCopySuccess(t *testing.T) {
	var fileSize int64 = 256
	testData := generateTestData(t, fileSize)

	successCases := []struct {
		toPath        string
		offset, limit int64
	}{
		{
			toPath: testData.tempDir + "/success_case_0",
			offset: 0,
			limit:  0,
		},
		{
			toPath: testData.tempDir + "/success_case_1",
			offset: 63,
			limit:  0,
		},
		{
			toPath: testData.tempDir + "/success_case_2",
			offset: 0,
			limit:  113,
		},
		{
			toPath: testData.tempDir + "/success_case_3",
			offset: 63,
			limit:  113,
		},
		{
			toPath: testData.tempDir + "/success_case_4",
			offset: 113,
			limit:  1000,
		},
		{
			toPath: testData.tempDir + "/edge_case",
			offset: fileSize,
			limit:  0,
		},
	}

	for _, c := range successCases {
		c := c

		t.Run(fmt.Sprintf("success case offset %d limit %d", c.offset, c.limit), func(t *testing.T) {
			err := Copy(testData.testFileName, c.toPath, c.offset, c.limit, io.Discard)
			require.NoError(t, err)

			bytes, err := os.ReadFile(c.toPath)
			require.NoError(t, err)

			lowerBound, upperBound := c.offset, c.offset+c.limit
			if c.limit == 0 || c.limit > fileSize {
				upperBound = testData.fileSize
			}

			require.ElementsMatch(t, bytes, testData.testBytes[lowerBound:upperBound])
		})
	}
}

func TestCopyError(t *testing.T) {
	var fileSize int64 = 128
	testData := generateTestData(t, fileSize)

	errorCases := []struct {
		fromPath, toPath string
		offset, limit    int64
		err              error
	}{
		{
			fromPath: testData.testFileName,
			toPath:   testData.tempDir + "/errorCase",
			offset:   fileSize + 1,
			err:      ErrOffsetExceedsFileSize,
		},

		{
			fromPath: testData.testFileName,
			toPath:   testData.tempDir + "/errorCase",
			offset:   -1,
			err:      ErrNegativeOffset,
		},

		{
			fromPath: testData.testFileName,
			toPath:   testData.tempDir + "/errorCase",
			limit:    -1,
			err:      ErrNegativeLimit,
		},

		{
			fromPath: "/dev/urandom",
			toPath:   testData.tempDir + "/errorCase",
			err:      ErrUnsupportedFile,
		},

		{
			fromPath: "/tmp/this_file_is_not_exists",
			toPath:   testData.tempDir + "/errorCase",
			err:      ErrSourceFileIsNotReadable,
		},

		{
			fromPath: testData.testFileName,
			toPath:   "/dev/try_to_write_in_dev",
			err:      ErrDestinationFileIsNotWritable,
		},
	}

	for _, c := range errorCases {
		c := c

		t.Run(fmt.Sprintf("error case %v", c.err), func(t *testing.T) {
			err := Copy(c.fromPath, c.toPath, c.offset, c.limit, io.Discard)

			require.ErrorIs(t, c.err, err)
		})
	}
}
