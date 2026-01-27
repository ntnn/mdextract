# mdextract

A Go tool and GitHub Action to extract code blocks from markdown files
with support for filtering by language and tags, as well as extracting
code blocks form withing HTML comments, e.g. to enable "hidden" code
blocks that are only relevant for CI but not for human readers.

`tags` are space-separated identifiers added after the language in code block fences:

    ```go ci test
    // This Go code block has tags: "ci" and "test"
    package main
    ```

### Examples

Extract all Go code blocks:
```bash
./bin/mdextract -input README.md -language go
```

Extract code blocks with specific tags:
```bash
./bin/mdextract -input README.md -tags ci,test -output extracted.sh
```

Extract Go code blocks tagged with "ci":
```bash
./bin/mdextract -input README.md -language go -tags ci -output test.go
```

## GitHub Action Usage

### Prerequisites

The action requires Go to be available in the workflow:

```yaml
- uses: actions/setup-go
```

The inputs are documented in the `action.yml` file.

Examples are available in [.github/workflows/example.yml](.github/workflows/example.yml).
