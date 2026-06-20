# Module 010: Configuration, Env Vars, And Flags

This module introduces the next primitive of a Go program: program behavior can change through configuration without changing source code.

## Configuration

Configuration is data a program reads when it starts.

The source code stays the same, but configuration values can change how the running program behaves.

In this module, configuration controls the server port and the message returned by `/config`.

## Environment Variable

An environment variable is a named value provided by the operating system to a process.

Here the server reads `APP_PORT`.

If `APP_PORT` is set to `8080`, the server listens on port `8080`.

## `os.Getenv`

`os.Getenv("APP_PORT")` reads the environment variable named `APP_PORT`.

It returns a `string`.

If the variable is not set, it returns an empty string.

## Default Values

A default value is the value the program uses when configuration is not provided.

This module uses `8080` as the default port when `APP_PORT` is not set.

Defaults keep a program runnable without requiring every configuration value every time.

## Command-Line Flag

A command-line flag is a value passed to the program when the process starts.

Here the program accepts a `-message` flag.

The flag changes the message returned by `/config`.

## The `flag` Package

The `flag` package reads command-line flags.

```go
message := flag.String("message", "hello from configuration", "message returned by /config")
```

This defines a `-message` flag with a default value.

## `flag.Parse`

`flag.Parse()` reads the command-line arguments and stores flag values.

The program calls `flag.Parse()` before using the configured message.

## Configured Port

The configured port changes where the server listens.

The program builds the listen address with:

```go
address := ":" + appConfig.Port
```

If the port is `8080`, the address is `:8080`.

## Configured Message

The configured message changes the JSON response from `/config`.

For example:

```json
{"port":"8080","message":"configured from docker run"}
```

The handler returns the active configuration as JSON.

## How This Builds On Module 009

Module 009 showed that HTTP services exchange structured data by decoding JSON requests and encoding JSON responses.

This module keeps JSON responses and adds startup configuration. The server still returns structured data, but now some response values come from environment variables and command-line flags.

## The Program

```go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
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

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func handleConfig(appConfig config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, configResponse{
			Port:    appConfig.Port,
			Message: appConfig.Message,
		})
	}
}

func main() {
	message := flag.String("message", "hello from configuration", "message returned by /config")
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
	fmt.Fprintln(os.Stdout, "stdout: starting HTTP server on", address, "with routes /healthz and /config")

	if err := http.ListenAndServe(address, nil); err != nil {
		fmt.Fprintln(os.Stderr, "server failed:", err)
		return
	}
}
```

## Build The Module Image

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build the docker image:

```bash
docker build -f Dockerfile -t go-scaling:module-010 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
docker rm -f go-scaling-module-010 >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-010 -p 8080:8080 -e APP_PORT=8080 go-scaling:module-010 -message "configured from docker run"
docker logs -f go-scaling-module-010 &
LOG_PID=$!
```

The `-e APP_PORT=8080` part passes an environment variable into the container.

The `-message "configured from docker run"` part passes a command-line flag to the compiled binary inside the container.

The server is now running in the background, and the log follower keeps container output visible in the same terminal.

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/config
```

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-010
wait "$LOG_PID" 2>/dev/null || true
```

Stopping the named container causes the background log follower to finish.

## What Changed From Module 009

Module 009 introduced JSON request and response handling.

This module keeps JSON responses and adds configuration. The server reads `APP_PORT` from the environment, reads `-message` from command-line flags, applies default values, and returns the active configuration from `/config`.

## First-Principles Chain

```text
source code → typed values → flags → environment variables → configuration → routes → handlers → JSON responses → long-running process → port → compiler checks → binary → operating-system process
```

Source code uses typed values. Flags and environment variables provide configuration when the process starts. Configuration changes the routes and handlers by changing what port the server listens on and what JSON response it returns. The compiled binary runs as a long-running process on a port. The compiler checks the code before building the binary. The operating system starts that binary as a process.

## Why Configuration Matters For Scaling Later

Larger services need to run in more than one environment.

Configuration lets the same binary run with different ports, messages, and later settings without changing source code for every environment.
