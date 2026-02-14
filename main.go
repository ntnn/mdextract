// Package mdextract is a tool for extracting markdown sections from
// files. It can be used to extract documentation from source code
// files, or to extract specific sections from markdown files.
package main

import (
	"errors"
	"log"
	"os"

	"github.com/ntnn/mdextract/pkg/mdextract"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	multi := &mdextract.Multi{}
	fs := multi.FlagSet()

	fOutput := fs.String("output", "", "Output file ('-' for stdout, not compatible with -multi)")

	fMulti := fs.Bool("multi", false, "Extract multiple sections based on the file tag (not compatible with -output)")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	if *fMulti && *fOutput != "" {
		fs.PrintDefaults()
		return errors.New("-multi and -output cannot be used together")
	}

	if !*fMulti && *fOutput == "" {
		fs.PrintDefaults()
		return errors.New("-multi or -output must be specified")
	}

	if fs.NArg() == 0 {
		fs.PrintDefaults()
		return errors.New("no input files specified")
	}

	if *fMulti {
		return doMulti(multi, fs.Args())
	}

	return doSingle(&multi.Single, *fOutput, multi.FileMode, fs.Args())
}

func doSingle(s *mdextract.Single, outputPath string, fileMode uint32, args []string) error {
	f := os.Stdout

	closeFn := func() error { return nil }

	if outputPath != "-" {
		var err error

		f, err = os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(fileMode)) //nolint:gosec
		if err != nil {
			return err
		}

		closeFn = f.Close
	}

	for _, input := range args {
		out, err := s.ExtractFromFile(input)
		if err != nil {
			return err
		}

		if _, err := f.WriteString(out); err != nil {
			return err
		}
	}

	return closeFn()
}

// in contrast to Multi.ExtractFromFileAndWrite, this function does not
// truncate files when writing. This allows parsing multiple files in
// one go and accumulating their output in the same files.
func doMulti(m *mdextract.Multi, args []string) error {
	for _, input := range args {
		out, err := m.ExtractFromFile(input)
		if err != nil {
			return err
		}

		for file, content := range out {
			if err := writeFileNoTruncate(file, []byte(content), os.FileMode(m.FileMode)); err != nil {
				return err
			}
		}
	}

	return nil
}

// copied from os.WriteFile but without truncation.
func writeFileNoTruncate(name string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, perm) //nolint:gosec
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}

	return err
}
