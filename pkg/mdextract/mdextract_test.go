package mdextract

import (
	"testing"

	"github.com/gomarkdown/markdown/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestFile_AcceptSection(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		expected bool
		info     string
		opts     Options
	}{
		"empty accepted":     {true, "", Options{}},
		"empty not accepted": {false, "", Options{IncludeEmpty: ptr.To(false)}},

		"tag match":              {true, "go ci", Options{Tags: []string{"ci"}}},
		"tag empty accepted":     {true, "", Options{Tags: []string{"ci"}}},
		"tag empty not accepted": {false, "", Options{Tags: []string{"ci"}, IncludeEmpty: ptr.To(false)}},
		"tag no match":           {false, "go test", Options{Tags: []string{"ci"}}},

		"tag match with multiple":   {true, "go ci test", Options{Tags: []string{"ci"}}},
		"tag no match with partial": {false, "go citation", Options{Tags: []string{"ci"}}},

		"langauge match":      {true, "go ci", Options{Language: "go"}},
		"language no match":   {false, "python ci", Options{Language: "go"}},
		"language empty":      {false, "ci", Options{Language: "go"}},
		"language only match": {true, "go", Options{Language: "go"}},

		"langauge tag match":     {true, "go ci", Options{Language: "go", Tags: []string{"ci"}}},
		"langauge tags match":    {true, "go ci export", Options{Language: "go", Tags: []string{"ci", "export"}}},
		"langauge tags no match": {false, "go export", Options{Language: "go", Tags: []string{"ci", "export"}}},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			f := &File{opts: cas.opts}
			require.NoError(t, f.opts.Validate())
			block := &ast.CodeBlock{
				Info: []byte(cas.info),
			}
			result := f.acceptSection(block)
			require.Equal(t, cas.expected, result)
		})
	}
}

func TestParseFile(t *testing.T) {
	t.Parallel()

	parsed, err := ParseFile("test.md", Options{})
	require.NoError(t, err)
	assert.Len(t, parsed.Sections, 9)

	parsed, err = ParseFile("test.md", Options{
		Tags: []string{"ci"},
	})
	require.NoError(t, err)
	assert.Len(t, parsed.Sections, 7)

	parsed, err = ParseFile("test.md", Options{
		Tags:         []string{"ci"},
		IncludeEmpty: ptr.To(false),
	})
	require.NoError(t, err)
	assert.Len(t, parsed.Sections, 2)

	parsed, err = ParseFile("test.md", Options{
		Tags:            []string{"ci"},
		IncludeEmpty:    ptr.To(false),
		IncludeComments: ptr.To(false),
	})
	require.NoError(t, err)
	assert.Len(t, parsed.Sections, 1)

	parsed, err = ParseFile("test.md", Options{
		Tags:            []string{"ci"},
		IncludeComments: ptr.To(false),
	})
	require.NoError(t, err)
	assert.Len(t, parsed.Sections, 5)

	parsed, err = ParseFile("test.md", Options{
		Tags: []string{"noci"},
	})
	require.NoError(t, err)
	assert.Len(t, parsed.Sections, 7)
}
