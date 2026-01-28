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
	// Langauge determines the language of the code blocks to extract.
	// If empty, all languages are accepted.
	Language string
	// Tags allows filtering code blocks. The content after the language
	// is split by spaces and treated as tags. Only code blocks that
	// have all specified tags will be extracted.
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

func (single *Single) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("single", flag.ExitOnError)
	fs.StringVar(&single.Language, "language", "", "Language to filter code blocks by")
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

func parseTag(b []byte) (string, []string) {
	if len(b) == 0 {
		return "", []string{}
	}
	s := strings.Split(string(b), " ")
	return s[0], s[1:]
}

func (single Single) acceptBlock(block *ast.CodeBlock) bool {
	lang, tags := parseTag(block.Info)
	if single.Language != "" && lang != single.Language {
		return false
	}
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

func (single Single) ExtractFromFile(p string) (string, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return "", err
	}
	return single.Extract(b)
}

func (single Single) Extract(data []byte) (string, error) {
	builder := &strings.Builder{}
	node := markdown.Parse(data, nil)

	ast.WalkFunc(node, ast.NodeVisitorFunc(func(node ast.Node, entering bool) ast.WalkStatus {
		switch n := node.(type) {
		case *ast.CodeBlock:
			if !single.acceptBlock(n) {
				return ast.GoToNext
			}
			builder.Write(n.Literal)
			// builder.WriteString("\n")
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
