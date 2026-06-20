# Module 008: Routes, Handlers, And Responses

This module introduces the next primitive of a Go program: a server chooses behavior by matching request paths to handler functions.

## Route

A route is a path a server recognizes.

Here the routes are `/`, `/healthz`, and `/module`.

## Request Path

A request path is the part of the request that identifies which route the client wants.

The server uses the path to choose which handler should run.

## Handler Functions

A handler function runs when a request matches a route.

This module uses more than one handler so each route can return its own plain text response.

## `http.HandleFunc`

`http.HandleFunc` registers a handler function for a specific path.

Here it connects different paths to different behavior.

## Response Body

The response body is the text the server sends back to the client.

In this module, the body is plain text.

## Response Status Code

A response status code tells the client whether the request succeeded or failed.

`http.StatusOK` means success.

`http.StatusNotFound` means the server did not find a matching route.

## `w.WriteHeader`

`w.WriteHeader` sets the status code before the body is written.

The `/healthz` route uses `http.StatusOK`, and the not-found path uses `http.StatusNotFound`.

## How This Builds On Module 007

Module 007 showed one handler and one route on a long-running HTTP server.

This module keeps the same server model and adds route-based behavior so one process can answer different requests differently.

## The Program

```go
package main

import (
	"fmt"
	"net/http"
	"os"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "welcome to module 008")
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func handleModule(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "module 008: routes, handlers, and responses")
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "not found")
}

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/healthz", handleHealthz)
	http.HandleFunc("/module", handleModule)

	fmt.Fprintln(os.Stdout, "stdout: starting HTTP server on :8080 with routes /, /healthz, and /module")

	if err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/", "/healthz", "/module":
			http.DefaultServeMux.ServeHTTP(w, r)
		default:
			handleNotFound(w, r)
		}
	})); err != nil {
		fmt.Fprintln(os.Stderr, "server failed:", err)
		return
	}
}
```

## Build The Module Image

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build the docker image:

```bash
docker build -f Dockerfile -t go-scaling:module-008 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
docker rm -f go-scaling-module-008 >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-008 -p 8080:8080 go-scaling:module-008
docker logs -f go-scaling-module-008 &
LOG_PID=$!
```

The server is now running in the background, and the log follower keeps container output visible in the same terminal.

From the same terminal, call each route:

```bash
curl http://localhost:8080/
curl http://localhost:8080/healthz
curl http://localhost:8080/module
curl -i http://localhost:8080/missing
```

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-008
wait "$LOG_PID" 2>/dev/null || true
```

Stopping the named container causes the background log follower to finish.

## What Changed From Module 007

Module 007 introduced one route and one handler on a running server.

This module adds multiple routes, multiple handlers, status codes, and not-found behavior so the server can choose different responses based on the request path.

## First-Principles Chain

```text
source code → typed values → functions → routes → handlers → status codes → responses → long-running process → port → request → compiler checks → binary → operating-system process
```

Source code uses typed values and functions. Routes map request paths to handlers. Handlers set status codes and write responses. The compiled binary runs as a long-running process on a port, receives requests, and returns responses. The compiler checks the code before building the binary. The operating system starts that binary as a process.

## Why Routes, Handlers, And Responses Matter For Scaling

Larger services need one process to answer many requests in different ways.

Routes and handlers give the server a simple way to choose behavior while keeping the response rules explicit.
