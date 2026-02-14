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

		"language match":      {true, "go ci", Single{Tags: []string{"go"}}},
		"language no match":   {false, "python ci", Single{Tags: []string{"go"}}},
		"language empty":      {false, "ci", Single{Tags: []string{"go"}}},
		"language only match": {true, "go", Single{Tags: []string{"go"}}},

		"langauge tag match":     {true, "go ci", Single{Tags: []string{"go", "ci"}}},
		"langauge tags match":    {true, "go ci export", Single{Tags: []string{"go", "ci", "export"}}},
		"langauge tags no match": {false, "go export", Single{Tags: []string{"go", "ci", "export"}}},

		"exclude tags match":    {false, "go ci export", Single{Tags: []string{"go", "ci"}, ExcludeTags: []string{"export"}}},
		"exclude tags no match": {true, "go ci noexport", Single{Tags: []string{"go", "ci"}, ExcludeTags: []string{"export"}}},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			block := &ast.CodeBlock{
				Info: []byte(cas.info),
			}
			tags := parseTag(block.Info)
			result := cas.single.acceptBlock(tags)
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
				Tags: []string{"bash"},
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
			parsed, err := cas.single.ExtractFromFile("single.md")
			require.NoError(t, err)
			assert.Equal(t, cas.expected, strings.Split(strings.TrimSpace(parsed), "\n"))
		})
	}

}

func TestParseTag(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input    string
		expected []string
	}{
		"empty": {
			input:    "",
			expected: []string{},
		},
		"single tag": {
			input:    "go",
			expected: []string{"go"},
		},
		"multiple tags": {
			input:    "go ci test",
			expected: []string{"go", "ci", "test"},
		},
		"multiple spaces": {
			input:    "go  ci   test",
			expected: []string{"go", "ci", "test"},
		},
		"leading and trailing spaces": {
			input:    " go ci test ",
			expected: []string{"go", "ci", "test"},
		},
		"tabs preserved in tag": {
			input:    "go\tci\ttest",
			expected: []string{"go\tci\ttest"},
		},
		"mixed whitespace": {
			input:    "  go \t ci  \n test  ",
			expected: []string{"go", "ci", "test"},
		},
		"only spaces": {
			input:    "   ",
			expected: []string{},
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()
			result := parseTag([]byte(cas.input))
			assert.Equal(t, cas.expected, result)
		})
	}
}

func TestSingle_ExtractFromFile_Error(t *testing.T) {
	t.Parallel()

	single := &Single{}
	_, err := single.ExtractFromFile("/non/existent/file.md")
	require.Error(t, err)

	tmpDir := t.TempDir()
	_, err = single.ExtractFromFile(tmpDir)
	require.Error(t, err)
}

func TestSingle_Extract_TableDriven(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input    string
		expected string
	}{
		"empty input": {
			input:    "",
			expected: "",
		},
		"no code blocks": {
			input:    "# Just a heading\n\nSome text",
			expected: "",
		},
		"multiple code blocks concatenation": {
			input:    "```\nfirst\n```\n\n```\nsecond\n```",
			expected: "first\nsecond\n",
		},
		"nested HTML comments": {
			input:    "<!--\n<!--\n```go\ncode\n```\n-->\n-->",
			expected: "code\n",
		},
		"malformed HTML comments": {
			input:    "<!-- unclosed comment\n```go\ncode\n```",
			expected: "code\n",
		},
		"code block with only whitespace": {
			input:    "```\n   \n\t\n```",
			expected: "   \n\t\n",
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			single := &Single{}
			result, err := single.Extract([]byte(cas.input))
			require.NoError(t, err)
			assert.Equal(t, cas.expected, result)
		})
	}
}

func TestSingle_AcceptBlock_AdditionalCases(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		single   Single
		tags     []string
		expected bool
	}{
		"empty tags with empty filter": {
			single:   Single{},
			tags:     []string{},
			expected: true,
		},
		"exclude supersedes": {
			single: Single{
				Tags:        []string{"go", "ci"},
				ExcludeTags: []string{"ci"},
			},
			tags:     []string{"go", "ci"},
			expected: false,
		},
		"partial match in exclude": {
			single: Single{
				ExcludeTags: []string{"test"},
			},
			tags:     []string{"testing"},
			expected: true,
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()
			result := cas.single.acceptBlock(cas.tags)
			assert.Equal(t, cas.expected, result)
		})
	}
}
