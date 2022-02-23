package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	progressBarLength = 50
	copyBufferSize    = 1024
)

var (
	ErrUnsupportedFile              = errors.New("unsupported file")
	ErrOffsetExceedsFileSize        = errors.New("offset exceeds file size")
	ErrSourceFileIsNotReadable      = errors.New("source file is not readable")
	ErrDestinationFileIsNotWritable = errors.New("destination file is not writable")
	ErrNegativeOffset               = errors.New("negative offset")
	ErrNegativeLimit                = errors.New("negative limit")
)

type writerWithProgress struct {
	dst        io.Writer
	output     io.Writer
	totalBytes int64
	written    int64
}

func (w *writerWithProgress) Write(b []byte) (int, error) {
	n, err := w.dst.Write(b)
	if err != nil {
		return 0, err
	}

	w.written += int64(n)
	w.printProgress()

	return n, nil
}

func (w *writerWithProgress) printProgress() {
	percent := float32(w.written) / float32(w.totalBytes)
	filled := strings.Repeat("#", int(progressBarLength*percent))
	empty := strings.Repeat(" ", progressBarLength-len(filled))

	fmt.Fprintf(w.output, "\r[%s%s] %d/%d", filled, empty, w.written, w.totalBytes)
}

func (w *writerWithProgress) finish() {
	fmt.Fprintln(w.output, "")
}

func newWriterWithProgress(file *os.File, totalBytes int64, output io.Writer) *writerWithProgress {
	return &writerWithProgress{
		dst:        file,
		output:     output,
		totalBytes: totalBytes,
	}
}

func Copy(fromPath, toPath string, offset, limit int64, output io.Writer) error {
	if offset < 0 {
		return ErrNegativeOffset
	}

	if limit < 0 {
		return ErrNegativeLimit
	}

	var srcFile, dstFile *os.File
	success := false
	defer func() {
		if !success && dstFile != nil {
			os.Remove(dstFile.Name())
		}
	}()

	srcFile, err := os.Open(fromPath)
	if err != nil {
		return ErrSourceFileIsNotReadable
	}
	defer srcFile.Close()

	stat, err := srcFile.Stat()
	if err != nil || !stat.Mode().IsRegular() || stat.Size() == 0 {
		return ErrUnsupportedFile
	}
	size := stat.Size()

	if offset > size {
		return ErrOffsetExceedsFileSize
	}

	_, err = srcFile.Seek(offset, 0)
	if err != nil {
		return ErrUnsupportedFile
	}

	dstFile, err = os.Create(toPath)
	if err != nil {
		return ErrDestinationFileIsNotWritable
	}
	defer dstFile.Close()

	bytesToCopy := size - offset
	if limit == 0 || limit > bytesToCopy {
		limit = bytesToCopy
	}

	reader := io.LimitReader(srcFile, limit)
	writer := newWriterWithProgress(dstFile, limit, output)
	defer writer.finish()

	buffer := make([]byte, copyBufferSize)
	_, err = io.CopyBuffer(writer, reader, buffer)
	if err != nil {
		return ErrDestinationFileIsNotWritable
	}

	success = true
	return nil
}
