package main

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type errReadWriteCloser struct {
	errRead  error
	errWrite error
}

func (e *errReadWriteCloser) Write(_ []byte) (n int, err error) {
	return 0, e.errWrite
}

func (e *errReadWriteCloser) Read(_ []byte) (n int, err error) {
	return 0, e.errRead
}

func (e *errReadWriteCloser) Close() error {
	return nil
}

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})

	t.Run("invalid host", func(t *testing.T) {
		client := NewTelnetClient(
			"some_host:65533",
			time.Second,
			io.NopCloser(&bytes.Buffer{}),
			&bytes.Buffer{})

		require.Errorf(t, client.Connect(), "connection error")
	})

	t.Run("eof", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		rw := errReadWriteCloser{
			errRead: io.EOF,
		}

		go func() {
			defer wg.Done()

			client := NewTelnetClient(l.Addr().String(), time.Second, &rw, &bytes.Buffer{})
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			err = client.Send()
			require.NoError(t, err)
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()
		}()

		wg.Wait()
	})

	t.Run("client read/write errors", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			errWrite := errors.New("write error")
			errRead := errors.New("read error")
			rw := errReadWriteCloser{
				errRead:  errRead,
				errWrite: errWrite,
			}

			client := NewTelnetClient(l.Addr().String(), time.Second, &rw, &rw)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			err = client.Send()
			require.ErrorIs(t, err, errRead)

			err = client.Receive()
			require.ErrorIs(t, err, errWrite)
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			n, err := conn.Write([]byte("str"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})
}
