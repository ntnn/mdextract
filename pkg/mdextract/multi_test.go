package mdextract

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFileTag(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		input string
		file  string
		tags  []string
	}{
		"empty": {
			input: "",
		},
		"only lang": {
			input: "go",
			tags:  []string{"go"},
		},
		"only file": {
			input: " file=main.go",
			file:  "main.go",
		},
		"tag and file": {
			input: "python file=script.py",
			file:  "script.py",
			tags:  []string{"python"},
		},
		"tag, file and tags": {
			input: "js file=app.js tag1 tag2",
			file:  "app.js",
			tags:  []string{"js", "tag1", "tag2"},
		},
		"file and tags": {
			input: " file=config.yaml prod debug",
			file:  "config.yaml",
			tags:  []string{"prod", "debug"},
		},
		"tags": {
			input: "ruby test unit",
			tags:  []string{"ruby", "test", "unit"},
		},
		"multiple file tags": {
			input: " file=first.txt file=second.txt tagA",
			file:  "second.txt",
			tags:  []string{"tagA"},
		},
		"file with path": {
			input: "go file=pkg/main.go",
			file:  "pkg/main.go",
			tags:  []string{"go"},
		},
		"file with spaces in name": {
			input: "file=my file.txt",
			file:  "my",
			tags:  []string{"file.txt"},
		},
		"empty file tag": {
			input: "file=",
			file:  "",
		},
		"only file prefix": {
			input: "file",
			tags:  []string{"file"},
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()
			file, tags := parseFileTag([]byte(cas.input))
			assert.Equal(t, cas.file, file)
			assert.Equal(t, cas.tags, tags)
		})
	}
}

func TestMulti_ExtractFromFile(t *testing.T) {
	t.Parallel()

	m := &Multi{}

	result, err := m.ExtractFromFile("multi.md")
	require.NoError(t, err)

	assert.Len(t, result, 4)
	assert.Contains(t, result, "example.yaml")
	assert.Contains(t, result, "example.go")
	assert.Contains(t, result, "example.txt")
	assert.Contains(t, result, "example.js")
}

func TestMulti_ExtractFromFile_Tags(t *testing.T) {
	t.Parallel()

	m := &Multi{
		Single: Single{
			Tags: []string{"ci"},
		},
	}

	result, err := m.ExtractFromFile("multi.md")
	require.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Contains(t, result, "example.go")
	assert.Contains(t, result, "example.js")
}

func TestMulti_ExtractFromFile_ExcludeTags(t *testing.T) {
	t.Parallel()

	m := &Multi{
		Single: Single{
			ExcludeTags: []string{"noci"},
		},
	}

	result, err := m.ExtractFromFile("multi.md")
	require.NoError(t, err)

	assert.Len(t, result, 3)
	assert.Contains(t, result, "example.yaml")
	assert.Contains(t, result, "example.go")
	assert.Contains(t, result, "example.js")
}

func TestMulti_Extract_TableDriven(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		multi    *Multi
		input    string
		expected map[string]string
	}{
		"empty": {
			multi:    &Multi{},
			input:    "",
			expected: map[string]string{},
		},
		"no code blocks": {
			multi:    &Multi{},
			input:    "# Just a heading\n\nSome text",
			expected: map[string]string{},
		},
		"code blocks without file tag": {
			multi:    &Multi{},
			input:    "```go\ncode without file\n```",
			expected: map[string]string{},
		},
		"duplicate file tags concatenate": {
			multi: &Multi{},
			input: "```go file=test.go\nfirst\n```\n\n```go file=test.go\nsecond\n```",
			expected: map[string]string{
				"test.go": "first\nsecond\n",
			},
		},
		"nested HTML comments": {
			multi: &Multi{},
			input: "<!--\n```go file=test.go\nouter\n```\n<!--\n```go file=test.go\ninner\n```\n-->\n-->",
			expected: map[string]string{
				"test.go": "outer\ninner\n",
			},
		},
		"mixed file and non-file code blocks": {
			multi: &Multi{},
			input: "```go\nno file\n```\n```go file=test.go\nwith file\n```",
			expected: map[string]string{
				"test.go": "with file\n",
			},
		},
		"code block with only whitespace": {
			multi: &Multi{},
			input: "```go file=test.go\n   \n\t\n```",
			expected: map[string]string{
				"test.go": "   \n\t\n",
			},
		},
		"file tag with path": {
			multi: &Multi{},
			input: "```go file=pkg/main.go\npackage main\n```",
			expected: map[string]string{
				"pkg/main.go": "package main\n",
			},
		},
		"file tag with language and extra tags": {
			multi: &Multi{},
			input: "```bash file=scripts/run.sh ci prod\necho ok\n```",
			expected: map[string]string{
				"scripts/run.sh": "echo ok\n",
			},
		},
		"multiple file tags uses last": {
			multi: &Multi{},
			input: "```go file=first.go file=second.go\npackage main\n```",
			expected: map[string]string{
				"second.go": "package main\n",
			},
		},
		"two files in separate blocks": {
			multi: &Multi{},
			input: "```go file=first.go\nfirst\n```\n```go file=second.go\nsecond\n```",
			expected: map[string]string{
				"first.go":  "first\n",
				"second.go": "second\n",
			},
		},
		"three files with mixed languages": {
			multi: &Multi{},
			input: "```go file=main.go\npackage main\n```\n```yaml file=config.yml\nkey: value\n```\n```bash file=scripts/build.sh\necho build\n```",
			expected: map[string]string{
				"main.go":          "package main\n",
				"config.yml":       "key: value\n",
				"scripts/build.sh": "echo build\n",
			},
		},
		"multiple files with duplicates": {
			multi: &Multi{},
			input: "```go file=one.go\nfirst\n```\n```go file=two.go\nsecond\n```\n```go file=one.go\nthird\n```",
			expected: map[string]string{
				"one.go": "first\nthird\n",
				"two.go": "second\n",
			},
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			result, err := cas.multi.Extract([]byte(cas.input))
			require.NoError(t, err)
			assert.Equal(t, cas.expected, result)
		})
	}
}

func TestMulti_ExtractFromFileAndWrite_Successful(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create a test markdown file with absolute path for output
	outputFile := filepath.Join(tmpDir, "test.go")
	mdContent := []byte("```go file=" + outputFile + "\npackage main\n```")
	mdFile := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(mdFile, mdContent, 0644)
	require.NoError(t, err)

	multi := &Multi{}
	err = multi.ExtractFromFileAndWrite(mdFile)
	require.NoError(t, err)

	// Verify the file was created
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, "package main\n", string(content))
}

func TestMulti_ExtractFromFileAndWrite_MultipleFiles(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	outputFile1 := filepath.Join(tmpDir, "one.go")
	outputFile2 := filepath.Join(tmpDir, "two.go")
	mdContent := []byte("```go file=" + outputFile1 + "\nfirst\n```\n```go file=" + outputFile2 + "\nsecond\n```")
	mdFile := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(mdFile, mdContent, 0644)
	require.NoError(t, err)

	multi := &Multi{}
	err = multi.ExtractFromFileAndWrite(mdFile)
	require.NoError(t, err)

	content1, err := os.ReadFile(outputFile1)
	require.NoError(t, err)
	assert.Equal(t, "first\n", string(content1))

	content2, err := os.ReadFile(outputFile2)
	require.NoError(t, err)
	assert.Equal(t, "second\n", string(content2))
}
