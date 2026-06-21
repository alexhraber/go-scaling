# Module 012: Graceful Shutdown

This module introduces the next primitive of a Go program: servers should stop accepting work and exit cleanly when the operating system asks them to stop.

## Graceful Shutdown

Graceful shutdown means the server stops accepting new work and gives itself a short time to finish cleanly before the process exits.

This is different from disappearing immediately while the server is still handling work.

## Stop Accepting Work

A server listens for requests while it is running.

When the operating system asks it to stop, the server should stop accepting new requests before exiting.

## Operating-System Signal

An operating-system signal is a small message sent to a running process.

Signals are one way the operating system asks a process to do something.

## `SIGINT`

`SIGINT` usually means an interrupt request.

For a command-line program, it often comes from pressing `Ctrl+C`.

## `SIGTERM`

`SIGTERM` means a request to terminate.

It asks the process to stop.

## Why `docker stop` Matters

`docker stop` sends a stop signal to the running process in the container.

This module listens for that signal and uses it to begin graceful shutdown.

## `signal.NotifyContext`

`signal.NotifyContext` creates a context that is canceled when selected signals arrive.

This module waits for `SIGINT` or `SIGTERM`.

## `context.WithTimeout`

`context.WithTimeout` creates a context with a deadline.

This module uses it to give shutdown a short maximum amount of time.

## `server.Shutdown`

`server.Shutdown(shutdownCtx)` asks the HTTP server to stop accepting new requests and shut down cleanly.

The shutdown context controls how long the server has to finish.

## `http.ErrServerClosed`

`http.ErrServerClosed` is normal when a server is shut down on purpose.

This module does not treat that value as a failure.

## One Small Goroutine

The server starts in one small goroutine so `ListenAndServe` can run while `main` waits for the operating-system shutdown signal.

This module uses that goroutine only for the server start path.

## Request Logging Still Applies

Module 011 showed that servers explain what happened by writing request events to output.

This module keeps request logs with method, path, status, and duration.

## Configuration Still Applies

Module 010 showed that program behavior can change through configuration without changing source code.

This module keeps `APP_PORT`, the `-message` flag, and JSON responses.

## The Program

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	message := flag.String("message", "hello from graceful shutdown", "message returned by /config")
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
	server := &http.Server{Addr: address}
	serverErr := make(chan error, 1)

	go func() {
		log.Printf("event=startup address=%s routes=/healthz,/config", address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-shutdownSignal.Done():
		log.Printf("event=shutdown_start")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("event=shutdown_failed error=%v", err)
			return
		}

		if err := <-serverErr; err != nil {
			log.Printf("event=server_failed error=%v", err)
			return
		}

		log.Printf("event=shutdown_complete")
	case err := <-serverErr:
		if err != nil {
			log.Printf("event=server_failed error=%v", err)
		}
	}
}
```

## Build The Module Image

The shared runtime image is built once from the repo root earlier in the learning flow.

From this directory, build the docker image:

```bash
docker build -f Dockerfile -t go-scaling:module-012 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
docker rm -f go-scaling-module-012 >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-012 -p 8080:8080 -e APP_PORT=8080 go-scaling:module-012 -message "configured graceful shutdown"
docker logs -f go-scaling-module-012 &
LOG_PID=$!
```

The `-e APP_PORT=8080` part passes an environment variable into the container.

The `-message "configured graceful shutdown"` part passes a command-line flag to the compiled binary inside the container.

The server is now running in the background, and the log follower keeps container output visible in the same terminal.

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/config
```

Each request should produce a request log line in the same terminal.

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-012
wait "$LOG_PID" 2>/dev/null || true
```

`docker stop` sends a stop signal. The server begins graceful shutdown, shutdown logs appear in the same terminal, and stopping the named container causes the background log follower to finish.

## What Changed From Module 011

Module 011 added request logs.

This module keeps those request logs and adds graceful shutdown. The server waits for `SIGINT` or `SIGTERM`, creates a shutdown context with a timeout, calls `server.Shutdown`, and logs shutdown start and completion.

## First-Principles Chain

```text
source code → typed values → configuration → routes → handlers → request logs → operating-system signal → shutdown context → graceful shutdown → process exit
```

Source code uses typed values and configuration. Routes map paths to handlers. Handlers write JSON responses and request logs. An operating-system signal asks the process to stop. A shutdown context gives the server a deadline. Graceful shutdown lets the server stop cleanly before the process exits.

## Why Graceful Shutdown Matters For Scaling Later

Larger services are started, stopped, and replaced over time.

Graceful shutdown gives a server a clear way to stop accepting work and exit cleanly when the operating system asks it to stop.
