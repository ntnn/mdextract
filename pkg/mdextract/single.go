package mdextract

import (
	"bytes"
	"flag"
	"os"
	"slices"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
)

// Single goes through a markdown document and extracts code blocks
// matching the criteria as one single concatenated string.
type Single struct {
	// Tags allows filtering code blocks. Everything on the first line
	// of a fenced code block is treated as a tag, including the
	// language.
	// Only code blocks that have all specified tags will be extracted.
	Tags []string
	// ExcludeTags allows excluding code blocks that contain any of
	// the specified tags. ExcludeTags supersedes Tags, e.g. if
	// a codeblock has both a tag in Tags and ExcludeTags, it will be
	// excluded.
	ExcludeTags []string
	// ExcludeComments disables extracting code blocks inside HTML
	// comments.
	// Default: false
	ExcludeComments bool
}

func split(s string) []string {
	if len(s) == 0 {
		return []string{}
	}

	return strings.Split(s, ",")
}

// FlagSet returns the flag set for the Single struct.
func (single *Single) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("single", flag.ExitOnError)
	fs.Func("tags", "Tags to filter code blocks, comma-separated", func(s string) error {
		single.Tags = split(s)
		return nil
	})
	fs.Func("exclude-tags", "Tags to exclude code blocks, comma-separated", func(s string) error {
		single.ExcludeTags = split(s)
		return nil
	})
	fs.BoolVar(&single.ExcludeComments, "exclude-comments", false, "Exclude code blocks inside HTML comments")

	return fs
}

func parseTag(b []byte) []string {
	if len(b) == 0 {
		return []string{}
	}

	ret := []string{}

	for tag := range strings.SplitSeq(string(b), " ") {
		tag = strings.TrimSpace(tag)
		if len(tag) == 0 {
			continue
		}

		ret = append(ret, tag)
	}

	return ret
}

func (single Single) acceptBlock(tags []string) bool {
	if len(single.Tags) > 0 {
		for _, tag := range single.Tags {
			if !slices.Contains(tags, tag) {
				return false
			}
		}
	}

	if len(single.ExcludeTags) > 0 {
		for _, exTag := range single.ExcludeTags {
			if slices.Contains(tags, exTag) {
				return false
			}
		}
	}

	return true
}

// ExtractFromFile reads a markdown file from the given path and
// extracts code block contents from it based on the specified tags.
func (single Single) ExtractFromFile(p string) (string, error) {
	b, err := os.ReadFile(p) //nolint:gosec
	if err != nil {
		return "", err
	}

	return single.Extract(b)
}

// Extract extracts code block contents from the given markdown data
// based on the specified tags.
func (single Single) Extract(data []byte) (string, error) {
	builder := &strings.Builder{}
	node := markdown.Parse(data, nil)

	ast.WalkFunc(node, ast.NodeVisitorFunc(func(node ast.Node, _ bool) ast.WalkStatus {
		switch n := node.(type) {
		case *ast.CodeBlock:
			tags := parseTag(n.Info)
			if !single.acceptBlock(tags) {
				return ast.GoToNext
			}

			builder.Write(n.Literal)
		case *ast.HTMLBlock:
			// an HTML block might be a comment with a code block that
			// should only be executed in e.g. CI
			if single.ExcludeComments {
				return ast.GoToNext
			}
			// Strip the comments, parse as markdown and add the code
			// blocks
			comment := bytes.TrimPrefix(n.Literal, []byte("<!--"))
			comment = bytes.TrimSuffix(comment, []byte("-->"))

			content, err := single.Extract(comment)
			if err != nil {
				return ast.Terminate
			}

			builder.WriteString(content)
		}

		return ast.GoToNext
	}))

	return builder.String(), nil
}
