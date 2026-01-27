package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ntnn/mdextract/pkg/mdextract"
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

var (
	fInput           = flag.String("input", "", "Path to the input file")
	fOutput          = flag.String("output", "-", "Path to the output file, defaults to stdout")
	fLanguage        = flag.String("language", "", "Language to filter code blocks")
	fTags            = flag.String("tags", "", "Tags to filter code blocks, comma-separated")
	fIncludeEmpty    = flag.Bool("include-empty", true, "Whether to include code blocks without any tags")
	fIncludeComments = flag.Bool("include-comments", true, "Whether to include code blocks inside HTML comments")
)

func split(s string) []string {
	if len(s) == 0 {
		return []string{}
	}
	return strings.Split(s, ",")
}

func run(ctx context.Context) error {
	if *fInput == "" {
		return fmt.Errorf("input file is required")
	}
	f, err := mdextract.ParseFile(*fInput, mdextract.Options{
		Language:        *fLanguage,
		Tags:            split(*fTags),
		IncludeEmpty:    fIncludeEmpty,
		IncludeComments: fIncludeComments,
	})
	if err != nil {
		return err
	}

	out := strings.Join(f.Sections, "\n")

	if *fOutput == "-" {
		fmt.Println(out)
		return nil
	}

	return os.WriteFile(*fOutput, []byte(out), 0o644)
}
