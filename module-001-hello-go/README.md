# Module 001: Hello Go

This module introduces the first primitive of a Go program: source code becomes a runnable operating-system process.

## The Program

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello from Go.")
}
```

## `package main`

`package main` tells Go that this source file belongs to a program that can be compiled into an executable binary.

A program that the operating system can run starts with package `main`.

## `import "fmt"`

`import "fmt"` tells Go that this file uses the standard library package named `fmt`.

The `fmt` package contains functions for formatting text and writing it to output.

## `func main()`

`func main()` defines the entry point of the program.

When the compiled binary starts as an operating-system process, Go begins by running this function.

## `fmt.Println(...)`

`fmt.Println(...)` writes one line of text to standard output.

Standard output is the stream where a command-line program normally prints its result.

## Build The Module Image

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build the docker image:

```bash
docker build -f Dockerfile -t go-scaling:module-001 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
docker run --rm go-scaling:module-001
```

The container starts the binary as an operating-system process and prints the program output.

## First-Principles Chain

```text
source code -> compiler -> binary -> operating-system process
```

The `.go` file is source code. The Go compiler turns that source code into a binary. The operating system starts that binary as a process. The process runs `main` and prints a line to standard output.

## Why This Matters For Scaling

Every larger Go program begins with this same chain.

Scaling comes later. First, the program must become a process that the operating system can run.
