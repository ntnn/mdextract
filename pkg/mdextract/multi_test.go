package mdextract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFileTag(t *testing.T) {
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
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			file, tags := parseFileTag([]byte(cas.input))
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
