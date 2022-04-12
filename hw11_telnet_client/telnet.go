package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	address    string
	timeout    time.Duration
	in         io.ReadCloser
	out        io.Writer
	connection net.Conn
}

func (t *telnetClient) Connect() error {
	var err error
	t.connection, err = net.DialTimeout("tcp", t.address, t.timeout)
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}

	return nil
}

func (t *telnetClient) Close() error {
	return t.connection.Close()
}

func (t *telnetClient) Send() error {
	if _, err := io.Copy(t.connection, t.in); err != nil {
		return fmt.Errorf("sending error: %w", err)
	}

	return nil
}

func (t *telnetClient) Receive() error {
	if _, err := io.Copy(t.out, t.connection); err != nil {
		return fmt.Errorf("receiving error: %w", err)
	}

	return nil
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{address, timeout, in, out, nil}
}
