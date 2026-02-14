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
	fOutput := fs.String("output", "", "Output file ('-' for stdout, not compatible with -multi)")
	fMulti := fs.Bool("multi", false, "Extract multiple sections based on the file tag (not compatible with -output)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	if *fMulti && *fOutput != "" {
		fs.PrintDefaults()
		return fmt.Errorf("-multi and -output cannot be used together")
	}

	if !*fMulti && *fOutput == "" {
		fs.PrintDefaults()
		return fmt.Errorf("-multi or -output must be specified")
	}

	if fs.NArg() == 0 {
		fs.PrintDefaults()
		return fmt.Errorf("no input files specified")
	}

	if *fMulti {
		m := &mdextract.Multi{
			Single: *s,
		}
		return doMulti(ctx, m, fs.Args())
	}
	return doSingle(ctx, s, *fOutput, fs.Args())
}

func doSingle(ctx context.Context, s *mdextract.Single, outputPath string, args []string) error {
	f := os.Stdout
	if outputPath != "-" {
		var err error
		f, err = os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	for _, input := range args {
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

func doMulti(ctx context.Context, m *mdextract.Multi, args []string) error {
	for _, input := range args {
		out, err := m.ExtractFromFile(input)
		if err != nil {
			return err
		}
		for file, content := range out {
			if err := writeFileNoTruncate(file, []byte(content), 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

// copied from os.WriteFile but without truncation
func writeFileNoTruncate(name string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}
