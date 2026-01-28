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

<!--
```go ci
package main
```
-->

### Examples

<!--
```bash ci
make build
```
-->

Extract all Go code blocks:
```bash ci
./bin/mdextract -language go README.md
```

Extract code blocks with specific tags and write to a file:
```bash ci
./bin/mdextract -language go -tags ci -output extracted.go README.md
```

<!--
```bash ci
if ! grep -q 'package main' extracted.go; then
    echo "extracted.go does not contain expected content"
    cat extracted.go
    exit 1
fi
```
-->

Extract example files in code blocks:

    ```yaml ci
    this:
        is: an example
        of: a yaml file
    ```

<!--
```yaml ci
this:
    is: an example
    of: a yaml file
```
-->

```bash ci
./bin/mdextract -language yaml -tags ci -output example.yaml README.md
```

<!--
```bash ci
if ! grep -q 'is: an example' example.yaml; then
    echo "example.yaml does not contain expected content"
    cat example.yaml
    exit 1
fi
```
-->

## GitHub Action Usage

### Prerequisites

The action requires Go to be available in the workflow:

```yaml noci
- uses: actions/setup-go
```

<!--
```bash ci
if grep -q 'actions/setup-go' example.yaml; then
    echo "example.yaml contains unexpected content"
    cat example.yaml
    exit 1
fi
```
-->

The inputs are documented in the `action.yml` file.

Examples are available in [.github/workflows/example.yml](.github/workflows/example.yml).
