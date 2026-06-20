# Module 009: JSON Request Response

This module introduces the next primitive of a Go program: HTTP services exchange structured data by decoding JSON requests and encoding JSON responses.

## JSON

JSON is a text format for structured data.

It can represent named fields and values, such as:

```json
{"message":"hello from json"}
```

HTTP services use JSON because clients and servers can exchange typed-looking records without sharing one process.

## Request Body

A request body is data the client sends inside the HTTP request.

For `/echo`, the request body contains JSON with a `message` field.

## Response Body

A response body is data the server sends back to the client.

For `/echo`, the response body contains JSON with the received message.

## `Content-Type: application/json`

`Content-Type: application/json` tells the client that the response body is JSON.

The server sets this header before writing JSON responses.

## Request Struct

A request struct is a Go struct that describes the JSON fields the server expects to read.

```go
type echoRequest struct {
	Message string `json:"message"`
}
```

The `json:"message"` tag connects the JSON field name to the Go struct field.

## Response Struct

A response struct is a Go struct that describes the JSON fields the server writes back.

```go
type echoResponse struct {
	Received string `json:"received"`
}
```

The response struct keeps the JSON response shape explicit.

## `json.NewDecoder(...).Decode(...)`

`json.NewDecoder(r.Body).Decode(&request)` reads JSON from the request body and stores it in the request struct.

The `&request` value lets the decoder fill in the struct fields.

## `json.NewEncoder(...).Encode(...)`

`json.NewEncoder(w).Encode(value)` writes a Go value as JSON to the response body.

Here the response writer is the destination.

## JSON Decode Errors

JSON decoding can fail when the request body is not valid JSON.

The same `if err != nil` pattern from Module 005 still applies: check the error, handle the failure, and return before the success path runs.

Invalid JSON is a bad client request, so the server returns status `400`.

## Wrong Method

The `/echo` route accepts only `POST` because it reads a request body.

If a client uses another method, the server returns status `405` to say the method is not allowed for that route.

## How This Builds On Module 008

Module 008 showed that a server chooses behavior by matching request paths to handler functions.

This module keeps `/healthz` and `/echo` as explicit routes. The new part is that handlers read JSON from the request body and write JSON to the response body.

## The Program

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type echoRequest struct {
	Message string `json:"message"`
}

type echoResponse struct {
	Received string `json:"received"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type healthResponse struct {
	Status string `json:"status"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	var request echoRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	writeJSON(w, http.StatusOK, echoResponse{Received: request.Message})
}

func main() {
	http.HandleFunc("/healthz", handleHealthz)
	http.HandleFunc("/echo", handleEcho)

	fmt.Fprintln(os.Stdout, "stdout: starting HTTP server on :8080 with routes /healthz and /echo")

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
docker build -f Dockerfile -t go-scaling:module-009 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
docker rm -f go-scaling-module-009 >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-009 -p 8080:8080 go-scaling:module-009
docker logs -f go-scaling-module-009 &
LOG_PID=$!
```

The server is now running in the background, and the log follower keeps container output visible in the same terminal.

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz

curl -X POST http://localhost:8080/echo \
  -H 'Content-Type: application/json' \
  -d '{"message":"hello from json"}'

curl -i -X POST http://localhost:8080/echo \
  -H 'Content-Type: application/json' \
  -d '{bad json}'

curl -i http://localhost:8080/echo
```

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-009
wait "$LOG_PID" 2>/dev/null || true
```

Stopping the named container causes the background log follower to finish.

## What Changed From Module 008

Module 008 returned plain text responses from route handlers.

This module returns JSON responses, reads a JSON request body, decodes that body into a request struct, encodes response structs back to JSON, and uses status codes for invalid JSON and wrong methods.

## First-Principles Chain

```text
source code → typed values → structs → routes → handlers → request body → JSON decode → response body → JSON encode → status codes → long-running process → port → compiler checks → binary → operating-system process
```

Source code uses typed values and structs. Routes map request paths to handlers. Handlers read a request body, decode JSON into structs, choose status codes, and encode response structs back into the response body. The compiled binary runs as a long-running process on a port. The compiler checks the code before building the binary. The operating system starts that binary as a process.

## Why JSON Request Response Handling Matters For Scaling Later

Larger services need a clear data contract between clients and servers.

JSON request and response handling lets a service receive structured input, return structured output, and make failure cases explicit before the system grows.
