package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ntnn/mdextract/pkg/mdextract"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	s := &mdextract.Single{}
	fs := s.FlagSet()
	fOutput := fs.String("output", "-", "Output file ('-' for stdout)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		fs.PrintDefaults()
		return fmt.Errorf("no input files specified")
	}

	f := os.Stdout
	if *fOutput != "-" {
		var err error
		f, err = os.OpenFile(*fOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	for _, input := range fs.Args() {
		out, err := s.ExtractFromFile(input)
		if err != nil {
			return err
		}
		if _, err := f.Write([]byte(out)); err != nil {
			return err
		}
	}

	return nil
}
