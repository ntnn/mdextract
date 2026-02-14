package mdextract

import (
	"bytes"
	"flag"
	"os"
	"strconv"
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

	// FileMode determines the mode when writing files.
	// Default: 0600
	FileMode uint32
}

const defaultFileMode = 0o600

func (multi Multi) fileMode() os.FileMode {
	if multi.FileMode == 0 {
		return defaultFileMode
	}

	return os.FileMode(multi.FileMode)
}

// uint32Value is a custom flag value type for uint32 values.
type uint32Value struct {
	value *uint32
}

func (u *uint32Value) String() string {
	if u.value == nil {
		return "0"
	}

	return strconv.FormatUint(uint64(*u.value), 8)
}

func (u *uint32Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return err
	}

	uv := uint32(v)
	u.value = &uv

	return nil
}

// FlagSet returns the flag set for the Multi struct.
func (multi *Multi) FlagSet() *flag.FlagSet {
	fs := multi.Single.FlagSet()
	fs.Init("multi", flag.ExitOnError)

	multi.FileMode = uint32(multi.fileMode())
	flagValue := &uint32Value{value: &multi.FileMode}
	fs.Var(flagValue, "file-mode", "File mode to use when writing files (in octal, e.g. 0600)")

	return fs
}

// ExtractFromFile reads a markdown file from the given path and
// extracts code blocks from it.
func (multi *Multi) ExtractFromFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		return nil, err
	}

	return multi.Extract(data)
}

// ExtractFromFileAndWrite reads a markdown file from the given path,
// extracts code blocks from it and writes the contents to files based
// on the "file" tag.
func (multi *Multi) ExtractFromFileAndWrite(path string) error {
	contents, err := multi.ExtractFromFile(path)
	if err != nil {
		return err
	}

	for file, content := range contents {
		if err := os.WriteFile(file, []byte(content), multi.fileMode()); err != nil {
			return err
		}
	}

	return nil
}

func parseFileTag(b []byte) (string, []string) {
	tags := parseTag(b)

	var (
		file      string
		otherTags []string
	)

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

// Extract extracts code blocks from the given markdown data and returns
// a map of filenames to their corresponding code contents. The filename
// is determined by the "file" tag in the code block's info string.
func (multi *Multi) Extract(data []byte) (map[string]string, error) {
	ret := make(map[string]string)
	node := markdown.Parse(data, nil)

	ast.WalkFunc(node, ast.NodeVisitorFunc(func(node ast.Node, _ bool) ast.WalkStatus {
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
			if multi.ExcludeComments {
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
