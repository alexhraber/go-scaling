# Module 011: Logging And Request Output

This module introduces the next primitive of a Go program: servers explain what happened by writing request events to output.

## Log

A log is a line of output that describes something the program did.

Servers write logs so a person can see what happened after requests arrive.

## Process Output

Logs go to process output because containers and operating systems already know how to collect output from a running process.

In this module, `docker logs -f ...` streams that output back into the same terminal.

## Request Log Line

A request log line is one line of output for one HTTP request.

It records the basic facts needed to understand the request.

## Key/Value Output

Key/value output writes each fact as a name and a value.

```text
method=GET path=/healthz status=200 duration=...
```

This stays readable while making each part of the line explicit.

## HTTP Method

The HTTP method describes what kind of request the client sent.

For these routes, `curl` sends `GET` requests.

## Request Path

The request path is the route the client asked for.

Here the paths are `/healthz` and `/config`.

## Response Status

The response status tells whether the server accepted the request.

Both routes in this module return status `200`.

## Request Duration

The request duration is how long the handler took to write the response and log the event.

Duration helps show how much time the server spent on a request.

## `log.Printf`

`log.Printf` writes a formatted log line to process output.

This module uses it for startup and for request log lines.

## `time.Now`

`time.Now()` records the current time.

Each handler records a start time near the beginning of the request.

## `time.Since`

`time.Since(start)` returns how much time has passed since `start`.

The request log uses that value as the request duration.

## Why Log After The Response

This module logs after writing the response so the log line can include the status code the handler actually used.

The handler starts the timer, writes the JSON response, then logs method, path, status, and duration.

## How This Builds On Module 010

Module 010 showed that program behavior can change through configuration without changing source code.

This module keeps `APP_PORT`, the `-message` flag, and JSON responses. It adds request logs so the server output explains what happened for each request.

## The Program

```go
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"
)

type config struct {
	Port    string
	Message string
}

type healthResponse struct {
	Status string `json:"status"`
}

type configResponse struct {
	Port    string `json:"port"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func logRequest(r *http.Request, status int, start time.Time) {
	log.Printf("method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, status, time.Since(start))
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := http.StatusOK
	writeJSON(w, status, healthResponse{Status: "ok"})
	logRequest(r, status, start)
}

func handleConfig(appConfig config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := http.StatusOK
		writeJSON(w, status, configResponse{
			Port:    appConfig.Port,
			Message: appConfig.Message,
		})
		logRequest(r, status, start)
	}
}

func main() {
	message := flag.String("message", "hello from logging", "message returned by /config")
	flag.Parse()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	appConfig := config{
		Port:    port,
		Message: *message,
	}

	http.HandleFunc("/healthz", handleHealthz)
	http.HandleFunc("/config", handleConfig(appConfig))

	address := ":" + appConfig.Port
	log.Printf("event=startup address=%s routes=/healthz,/config", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Printf("event=server_failed error=%v", err)
		return
	}
}
```

## Build The Module Image

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build the docker image:

```bash
docker build -f Dockerfile -t go-scaling:module-011 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
docker rm -f go-scaling-module-011 >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-011 -p 8080:8080 -e APP_PORT=8080 go-scaling:module-011 -message "configured logging"
docker logs -f go-scaling-module-011 &
LOG_PID=$!
```

The `-e APP_PORT=8080` part passes an environment variable into the container.

The `-message "configured logging"` part passes a command-line flag to the compiled binary inside the container.

The server is now running in the background, and the log follower keeps container output visible in the same terminal.

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/config
```

Each request should produce a log line in the same terminal, such as:

```text
method=GET path=/healthz status=200 duration=...
method=GET path=/config status=200 duration=...
```

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-011
wait "$LOG_PID" 2>/dev/null || true
```

Stopping the named container causes the background log follower to finish.

## What Changed From Module 010

Module 010 introduced environment variables, defaults, and a command-line flag.

This module keeps that configuration model and adds request output. Each handler records a start time, writes a JSON response, and logs the method, path, status code, and duration.

## First-Principles Chain

```text
source code → typed values → configuration → routes → handlers → JSON responses → request events → logs → process output → long-running process → port → compiler checks → binary → operating-system process
```

Source code uses typed values and configuration. Routes map paths to handlers. Handlers write JSON responses and produce request events. Logs write those events to process output. The compiled binary runs as a long-running process on a port. The compiler checks the code before building the binary. The operating system starts that binary as a process.

## Why Logging Matters For Scaling Later

Larger services handle many requests over time.

Logs let a person inspect what the server did without changing the program or attaching a debugger to the running process.
