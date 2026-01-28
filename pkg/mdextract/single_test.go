package mdextract

import (
	"strings"
	"testing"

	"github.com/gomarkdown/markdown/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSingle_AcceptBlock(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		expected bool
		info     string
		single   Single
	}{
		"empty accepted": {true, "", Single{}},

		"tag match":    {true, "go ci", Single{Tags: []string{"ci"}}},
		"tag no match": {false, "go test", Single{Tags: []string{"ci"}}},

		"tag match with multiple":   {true, "go ci test", Single{Tags: []string{"ci"}}},
		"tag no match with partial": {false, "go citation", Single{Tags: []string{"ci"}}},

		"langauge match":      {true, "go ci", Single{Language: "go"}},
		"language no match":   {false, "python ci", Single{Language: "go"}},
		"language empty":      {false, "ci", Single{Language: "go"}},
		"language only match": {true, "go", Single{Language: "go"}},

		"langauge tag match":     {true, "go ci", Single{Language: "go", Tags: []string{"ci"}}},
		"langauge tags match":    {true, "go ci export", Single{Language: "go", Tags: []string{"ci", "export"}}},
		"langauge tags no match": {false, "go export", Single{Language: "go", Tags: []string{"ci", "export"}}},

		"exclude tags match":    {false, "go ci export", Single{Language: "go", Tags: []string{"ci"}, ExcludeTags: []string{"export"}}},
		"exclude tags no match": {true, "go ci noexport", Single{Language: "go", Tags: []string{"ci"}, ExcludeTags: []string{"export"}}},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			block := &ast.CodeBlock{
				Info: []byte(cas.info),
			}
			lang, tags := parseTag(block.Info)
			result := cas.single.acceptBlock(lang, tags)
			require.Equal(t, cas.expected, result)
		})
	}
}

func TestSingle_Extract(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		single   Single
		input    []string
		expected []string
	}{
		"html comment with <!--": {
			single: Single{
				Language: "bash",
			},
			input: []string{
				"Some text",
				"<!--",
				"```bash",
				"code block inside comment",
				"```",
				"-->",
			},
			expected: []string{
				"code block inside comment",
			},
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()
			parsed, err := cas.single.Extract([]byte(strings.Join(cas.input, "\n")))
			require.NoError(t, err)
			assert.Equal(t, cas.expected, strings.Split(strings.TrimSpace(parsed), "\n"))
		})
	}
}

func TestSingle_ExtractFromFile(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		single   Single
		expected []string
	}{
		"no filter": {
			single: Single{},
			expected: []string{
				"code block with no tag",
				"code block with yaml",
				"code block with go",
				"code block with go with tag ci",
				"code block inside comment",
				"code block inside comment with tag ci",
				"code block inside comment with tag noci",
				"code block with go tag noci",
				`{"a": {"b": 1}}`,
			},
		},
		"tag ci": {
			single: Single{
				Tags: []string{"ci"},
			},
			expected: []string{
				"code block with go with tag ci",
				"code block inside comment with tag ci",
			},
		},
		"tag ci, excluding empty": {
			single: Single{
				Tags: []string{"ci"},
			},
			expected: []string{
				"code block with go with tag ci",
				"code block inside comment with tag ci",
			},
		},
		"tag ci, excluding empty and comments": {
			single: Single{
				Tags:            []string{"ci"},
				ExcludeComments: true,
			},
			expected: []string{
				"code block with go with tag ci",
			},
		},
		"tag ci, excluding comments": {
			single: Single{
				Tags:            []string{"ci"},
				ExcludeComments: true,
			},
			expected: []string{
				"code block with go with tag ci",
			},
		},
		"tag noci": {
			single: Single{
				Tags: []string{"noci"},
			},
			expected: []string{
				"code block inside comment with tag noci",
				"code block with go tag noci",
			},
		},
		"exclude noci": {
			single: Single{
				ExcludeTags: []string{"noci"},
			},
			expected: []string{
				"code block with no tag",
				"code block with yaml",
				"code block with go",
				"code block with go with tag ci",
				"code block inside comment",
				"code block inside comment with tag ci",
				`{"a": {"b": 1}}`,
			},
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()
			parsed, err := cas.single.ExtractFromFile("test.md")
			require.NoError(t, err)
			assert.Equal(t, cas.expected, strings.Split(strings.TrimSpace(parsed), "\n"))
		})
	}

}
