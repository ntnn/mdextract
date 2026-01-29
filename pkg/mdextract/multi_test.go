package mdextract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFileTag(t *testing.T) {
	cases := map[string]struct {
		input      string
		lang, file string
		tags       []string
	}{
		"empty": {
			input: "",
		},
		"only lang": {
			input: "go",
			lang:  "go",
		},
		"only file": {
			input: " file=main.go",
			file:  "main.go",
		},
		"lang and file": {
			input: "python file=script.py",
			lang:  "python",
			file:  "script.py",
		},
		"lang, file and tags": {
			input: "js file=app.js tag1 tag2",
			lang:  "js",
			file:  "app.js",
			tags:  []string{"tag1", "tag2"},
		},
		"file and tags": {
			input: " file=config.yaml prod debug",
			file:  "config.yaml",
			tags:  []string{"prod", "debug"},
		},
		"lang and tags": {
			input: "ruby test unit",
			lang:  "ruby",
			tags:  []string{"test", "unit"},
		},
		"multiple file tags": {
			input: " file=first.txt file=second.txt tagA",
			file:  "second.txt",
			tags:  []string{"tagA"},
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			lang, file, tags := parseFileTag([]byte(cas.input))
			assert.Equal(t, cas.lang, lang)
			assert.Equal(t, cas.file, file)
			assert.Equal(t, cas.tags, tags)
		})
	}
}

func TestMulti_ExtractFromFile(t *testing.T) {
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
