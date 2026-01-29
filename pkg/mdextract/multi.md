# This is an example markdown file

It features

```yaml file=example.yaml
simple:
  yaml: documents
```

As well as code block:

```go file=example.go ci
package main

func main() {
    println("Hello, World!")
}
```

And would like to have a block without language, however the library
used for parsing markdown trims spaces.

```txt file=example.txt noci
This is a code block without a specified language.
```

And blocks without any file name

```python ci
def greet():
    print("Hello, World!")
```

<!--
And blocks in comments!
```javascript file=example.js ci
console.log("Hello, World!");
```
-->
