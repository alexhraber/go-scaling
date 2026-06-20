# Module 007: First HTTP Server

This module introduces the next primitive of a Go program: a web server is a long-running process that listens for requests and writes responses.

## An HTTP Server

An HTTP server is a program that waits for HTTP requests and sends HTTP responses.

It keeps running because it must stay ready for the next request.

## Port

A port is a numbered doorway on a machine.

Here the server listens on port `8080`, so other programs can connect to that port and send requests.

## Request And Response

A request is a message a client sends to a server.

A response is the message the server sends back.

In this module, `curl` sends a request and prints the response.

## A Handler Function

A handler function runs when a request arrives for a route.

The function receives the response writer and the request.

## `http.HandleFunc`

`http.HandleFunc` registers a handler function for a route.

Here it maps the `/` route to `handleRoot`.

## `http.ListenAndServe`

`http.ListenAndServe` starts the HTTP server on a port.

It returns an error if startup fails, so the program must check that error.

## `if err != nil`

The `if err != nil` pattern from Module 005 still applies.

If `ListenAndServe` returns an error, print it and return early.

## How This Builds On Module 006

Module 006 showed that a program is a process that reads and writes through operating-system streams and files.

This module keeps that process model and adds a network port where the process can wait for requests and write responses.

## The Program

```go
package main

import (
	"fmt"
	"net/http"
	"os"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello from module 007")
}

func main() {
	http.HandleFunc("/", handleRoot)

	fmt.Fprintln(os.Stdout, "stdout: starting HTTP server on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Fprintln(os.Stderr, "server failed:", err)
		return
	}
}
```

## Build The Module Image

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build the docker image:

```bash
docker build -f Dockerfile -t go-scaling:module-007 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
docker run --rm -p 8080:8080 go-scaling:module-007
```

This keeps running because the server is waiting for requests.

In a second terminal, call the server:

```bash
curl http://localhost:8080/
```

## What Changed From Module 006

Module 006 read from standard input, wrote to standard output and standard error, and used files for stored data.

This module keeps the same process-and-error model but adds an HTTP listener on port `8080` and returns a plain text response to a request.

## First-Principles Chain

```text
source code → typed values → functions → errors → long-running process → port → request → response → compiler checks → binary → operating-system process
```

Source code uses typed values and functions. Functions can return errors. The compiled binary runs as a long-running process. The process listens on a port, receives a request, and writes a response. The compiler checks the code before building the binary. The operating system starts that binary as a process.

## Why A First HTTP Server Matters For Scaling

A first HTTP server shows how a program can stay alive, accept work from outside the process, and answer one request at a time.

That is the starting point for building larger services later.
