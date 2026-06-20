# Module 006: Files And Standard Streams

This module introduces the next primitive of a Go program: programs communicate through operating-system streams and files.

## Standard Output

Standard output is the stream where a program prints normal results.

Command-line programs usually write their main output to standard output so other tools can read it.

## Standard Error

Standard error is the stream where a program prints error-style messages.

Keeping error messages separate from normal output helps shell pipelines and scripts handle success and failure differently.

## Standard Input

Standard input is the stream where a program reads input.

Command-line programs often read from standard input so other programs can send them data without interactive typing.

## A File

From the program’s point of view, a file is data stored by the operating system that can be written and read again later.

Writing a file is different from printing to standard output because the data is stored on disk instead of flowing through the output stream.

Reading a file is different from reading standard input because the data comes from stored bytes instead of from the input stream connected to the process.

## Errors From File Operations

File operations can fail because the path may not exist, permissions may block access, or the filesystem may have a problem.

That is why `os.WriteFile` and `os.ReadFile` return errors.

The `if err != nil` pattern from Module 005 applies here too: check the error, print a clear message, and return early.

## The Program

```go
package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	fmt.Fprintln(os.Stdout, "stdout: module 006 is running")
	fmt.Fprintln(os.Stderr, "stderr: this is an error-style message")

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read stdin:", err)
		return
	}

	notePath := "/tmp/module-006-note.txt"
	if err := os.WriteFile(notePath, input, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write file:", err)
		return
	}

	note, err := os.ReadFile(notePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to read file:", err)
		return
	}

	fmt.Println("stdin input:")
	fmt.Print(string(input))
	fmt.Println("file path:", notePath)
	fmt.Println("file contents:")
	fmt.Print(string(note))
}
```

## Build The Module Image

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build the docker image:

```bash
docker build -f Dockerfile -t go-scaling:module-006 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
printf "hello from stdin\n" | docker run --rm -i go-scaling:module-006
```

The container starts the binary as an operating-system process, reads stdin, writes a file, and prints the program output.

## What Changed From Module 005

Module 005 used return values and errors to report failure.

This module applies the same `if err != nil` pattern to streams and files. It reads from standard input, writes to standard output and standard error, writes a file, reads that file back, and prints the results with clear labels.

## First-Principles Chain

```text
source code → typed values → functions → errors → standard streams → files → compiler checks → binary → operating-system process
```

Source code uses typed values and functions. Functions return errors when stream or file operations fail. Standard streams connect the process to the shell. Files store data for later reads. The compiler checks the program before building a binary. The operating system starts that binary as a process.

## Why Files And Standard Streams Matter For Scaling

Larger programs usually do not live only inside one function or one terminal command.

They read input from other programs, write normal output for the next tool, report failures separately, and persist data in files when state needs to survive beyond one process.
