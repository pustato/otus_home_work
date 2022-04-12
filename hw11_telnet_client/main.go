package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"
)

type options struct {
	timeout time.Duration
	address string
}

func main() {
	opts, err := parseOptions()
	if err != nil {
		fmt.Printf("%v. usage: %s [address string] [port int]\n", err, os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	client := NewTelnetClient(opts.address, opts.timeout, os.Stdin, os.Stdout)
	err = client.Connect()
	if err != nil {
		mustPrintlnf("%v", err.Error())
		os.Exit(1)
	}
	mustPrintlnf("...Connected to %s", opts.address)
	defer client.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	go func() {
		defer cancel()

		if err := client.Receive(); err != nil {
			select {
			case <-ctx.Done():
			default:
				mustPrintlnf("%v", err)
			}
			return
		}

		mustPrintlnf("...Connection was closed by peer")
	}()

	go func() {
		defer cancel()

		if err := client.Send(); err != nil {
			mustPrintlnf("%v", err)
			return
		}

		mustPrintlnf("...EOF")
	}()

	<-ctx.Done()
}

func mustPrintlnf(format string, a ...interface{}) {
	_, err := fmt.Fprintf(os.Stderr, format+"\n", a...)
	if err != nil {
		panic(err)
	}
}

func parseOptions() (*options, error) {
	opts := &options{}

	flag.DurationVar(&opts.timeout, "timeout", time.Second*10, "connection timeout")
	flag.Parse()

	if flag.NArg() != 2 {
		return nil, errors.New("not enough arguments")
	}
	opts.address = net.JoinHostPort(flag.Arg(0), flag.Arg(1))

	return opts, nil
}
