package mdextract

import (
	"bytes"
	"os"
	"slices"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"k8s.io/utils/ptr"
)

type Options struct {
	// Langauge determines the language of the code blocks to extract.
	// If empty, all languages are accepted.
	Language string
	// Tags allows filtering code blocks. The content after the language
	// is split by spaces and treated as tags. Only tags that have all
	// specified tags will be extracted.
	Tags []string
	// IncludeEmpty determines whether code blocks without any tags
	// are included regardless of the Tags. The language still applies.
	// Default: true
	IncludeEmpty *bool
	// IncludeComments allows extracting code blocks that are inside
	// HTML comments.
	// Default: true
	IncludeComments *bool
}

func (o *Options) Validate() error {
	if o.IncludeEmpty == nil {
		o.IncludeEmpty = ptr.To(true)
	}
	if o.IncludeComments == nil {
		o.IncludeComments = ptr.To(true)
	}
	return nil
}

type File struct {
	opts Options

	Sections []string
}

func ParseFile(p string, opts Options) (*File, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return Parse(b, opts)
}

func Parse(data []byte, opts Options) (*File, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	node := markdown.Parse(data, nil)
	f := new(File)
	f.opts = opts
	f.Sections = make([]string, 0)

	ast.WalkFunc(node, ast.NodeVisitorFunc(func(node ast.Node, entering bool) ast.WalkStatus {
		switch n := node.(type) {
		case *ast.CodeBlock:
			f.addSection(n)
		case *ast.HTMLBlock:
			// an HTML block might be a comment with a code block that
			// should only be executed in e.g. CI
			if !*opts.IncludeComments {
				return ast.GoToNext
			}
			// Strip the comments, parse as markdown and add the code
			// blocks
			comment := bytes.TrimPrefix(n.Literal, []byte("<!--"))
			comment = bytes.TrimSuffix(comment, []byte("-->"))
			newNode, err := Parse(comment, opts)
			if err != nil {
				// TODO handle error
				return ast.Terminate
			}
			f.Sections = append(f.Sections, newNode.Sections...)
		}
		return ast.GoToNext
	}))

	return f, nil
}

func parseTag(b []byte) (string, []string) {
	if len(b) == 0 {
		return "", []string{}
	}
	s := strings.Split(string(b), " ")
	return s[0], s[1:]
}

func (f *File) acceptSection(block *ast.CodeBlock) bool {
	if len(block.Info) == 0 {
		return *f.opts.IncludeEmpty
	}
	lang, tags := parseTag(block.Info)
	if f.opts.Language != "" && lang != f.opts.Language {
		return false
	}
	if len(f.opts.Tags) > 0 {
		if *f.opts.IncludeEmpty && len(tags) == 0 {
			return true
		}
		for _, tag := range f.opts.Tags {
			if !slices.Contains(tags, tag) {
				return false
			}
		}
	}
	return true
}

func (f *File) addSection(block *ast.CodeBlock) {
	if !f.acceptSection(block) {
		return
	}
	f.Sections = append(f.Sections, string(block.Literal))
}
