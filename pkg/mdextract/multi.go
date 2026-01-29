package mdextract

import (
	"bytes"
	"flag"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
)

// Multi goes through a markdown document and extracts code blocks
// matching the criteria, separated by the "file" tag.
// The "file" tag is used to determine the filename for each
// code block. If multiple code blocks have the same "file" tag,
// their contents are concatenated.
type Multi struct {
	Single
}

func (multi *Multi) FlagSet() *flag.FlagSet {
	return multi.Single.FlagSet()
}

func (multi *Multi) ExtractFromFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return multi.Extract(data)
}

func (multi *Multi) ExtractFromFileAndWrite(path string) error {
	contents, err := multi.ExtractFromFile(path)
	if err != nil {
		return err
	}
	for file, content := range contents {
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func parseFileTag(b []byte) (string, []string) {
	tags := parseTag(b)
	var file string
	var otherTags []string
	for _, tag := range tags {
		switch {
		case strings.HasPrefix(tag, "file="):
			file = strings.TrimPrefix(tag, "file=")
		default:
			otherTags = append(otherTags, tag)
		}
	}
	return file, otherTags
}

func (multi *Multi) Extract(data []byte) (map[string]string, error) {
	ret := make(map[string]string)
	node := markdown.Parse(data, nil)

	ast.WalkFunc(node, ast.NodeVisitorFunc(func(node ast.Node, entering bool) ast.WalkStatus {
		switch n := node.(type) {
		case *ast.CodeBlock:
			file, tags := parseFileTag(n.Info)
			if file == "" {
				return ast.GoToNext
			}
			if !multi.acceptBlock(tags) {
				return ast.GoToNext
			}
			ret[file] += string(n.Literal)
		case *ast.HTMLBlock:
			// an HTML block might be a comment with a code block that
			// should only be executed in e.g. CI
			if multi.Single.ExcludeComments {
				return ast.GoToNext
			}
			// Strip the comments, parse as markdown and add the code
			// blocks
			comment := bytes.TrimPrefix(n.Literal, []byte("<!--"))
			comment = bytes.TrimSuffix(comment, []byte("-->"))
			contents, err := multi.Extract(comment)
			if err != nil {
				return ast.Terminate
			}
			for k, v := range contents {
				ret[k] += v
			}
		}
		return ast.GoToNext
	}))

	return ret, nil
}
