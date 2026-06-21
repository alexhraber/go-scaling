# Module 014: Server Timeouts And Deadlines

This module introduces the next scaling primitive of a Go program: a service must enforce time boundaries at the server edge, not only inside handler code.

The core primitive is:

```text
A scalable service protects itself by putting deadlines around connections, headers, requests, responses, and idle clients.
```

## Server Timeout

A server timeout is a time boundary on part of the HTTP connection or response process.

Timeouts keep clients from holding server resources forever.

## Server-Edge Protection

The server edge is where client connections meet the service process.

Server-edge timeouts protect the service before handler logic becomes the only defense.

## `ReadHeaderTimeout`

`ReadHeaderTimeout` limits how long the server waits for request headers.

At a beginner level, it protects the server from clients that connect but send headers too slowly.

## `ReadTimeout`

`ReadTimeout` limits how long the server spends reading the whole request.

It covers more than the headers.

## `WriteTimeout`

`WriteTimeout` limits how long the server spends writing the response.

It protects the server from response writes that take too long.

## `IdleTimeout`

`IdleTimeout` limits how long an idle client connection can stay open between requests.

It protects the server from clients that keep connections open without doing useful work.

## Server Timeouts And Request Context

Request context tells a handler when the client is gone.

Server timeouts protect the connection boundary.

They are related, but they are not the same thing.

## Handler Deadline

Handler work may need its own deadline because the handler can spend too long on expensive work even when the client is still connected.

This module uses `context.WithTimeout(r.Context(), 2*time.Second)` inside `/work`.

The simulated work takes 3 seconds, so the handler deadline expires first.

## `context.WithTimeout`

`context.WithTimeout` creates a child context that is canceled when its deadline expires.

Inside a handler, it puts a time boundary around the handler's own work.

## Stopping Slow Work

Slow work should stop when its deadline expires.

Unbounded work wastes CPU time, memory, and request slots.

At scale, unbounded work becomes dangerous because too many slow requests can consume capacity needed for useful work.

## Primitive Counters

This module still uses primitive counters instead of a metrics framework.

The counters are stored in process memory and protected with `sync.Mutex`.

## `/metrics`

The `/metrics` route returns JSON counters:

```json
{
  "requests_total": 4,
  "work_started_total": 1,
  "work_completed_total": 0,
  "work_timed_out_total": 1
}
```

The metrics show how many requests arrived, how many `/work` requests started, how many completed, and how many hit the handler deadline.

## Graceful Shutdown Still Applies

Module 012 showed that a service process should exit cleanly when the operating system asks it to stop.

This module keeps that graceful shutdown shape with an explicit `http.Server`, `signal.NotifyContext`, `context.WithTimeout`, and `server.Shutdown`.

## Request Lifetime Still Applies

Module 013 showed that every request has a lifetime and handlers should stop work when `r.Context()` is canceled.

This module keeps request lifetime handling and adds a separate handler-level deadline.

The important distinction is:

```text
Request context tells a handler when the client is gone. Handler deadlines tell the handler when its own work has taken too long. Server timeouts protect the connection boundary.
```

## The Program

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type config struct {
	Port string
}

type metrics struct {
	mu                 sync.Mutex
	requestsTotal      int
	workStartedTotal   int
	workCompletedTotal int
	workTimedOutTotal  int
}

type healthResponse struct {
	Status string `json:"status"`
}

type workResponse struct {
	Status string `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type metricsResponse struct {
	RequestsTotal      int `json:"requests_total"`
	WorkStartedTotal   int `json:"work_started_total"`
	WorkCompletedTotal int `json:"work_completed_total"`
	WorkTimedOutTotal  int `json:"work_timed_out_total"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func logRequest(r *http.Request, status int, start time.Time) {
	log.Printf("method=%s path=%s status=%d duration=%s", r.Method, r.URL.Path, status, time.Since(start))
}

func incrementRequests(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.requestsTotal++
}

func incrementWorkStarted(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.workStartedTotal++
}

func incrementWorkCompleted(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.workCompletedTotal++
}

func incrementWorkTimedOut(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.workTimedOutTotal++
}

func snapshotMetrics(appMetrics *metrics) metricsResponse {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()

	return metricsResponse{
		RequestsTotal:      appMetrics.requestsTotal,
		WorkStartedTotal:   appMetrics.workStartedTotal,
		WorkCompletedTotal: appMetrics.workCompletedTotal,
		WorkTimedOutTotal:  appMetrics.workTimedOutTotal,
	}
}

func handleHealthz(appMetrics *metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := http.StatusOK
		incrementRequests(appMetrics)
		writeJSON(w, status, healthResponse{Status: "ok"})
		logRequest(r, status, start)
	}
}

func handleWork(appMetrics *metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		incrementRequests(appMetrics)
		incrementWorkStarted(appMetrics)
		log.Printf("event=work_start path=/work")

		workCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		timer := time.NewTimer(3 * time.Second)
		defer timer.Stop()

		select {
		case <-timer.C:
			status := http.StatusOK
			incrementWorkCompleted(appMetrics)
			log.Printf("event=work_complete path=/work")
			writeJSON(w, status, workResponse{Status: "completed"})
			logRequest(r, status, start)
		case <-workCtx.Done():
			status := http.StatusGatewayTimeout
			incrementWorkTimedOut(appMetrics)
			log.Printf("event=work_deadline_exceeded path=/work error=%v", workCtx.Err())
			writeJSON(w, status, errorResponse{Error: "work deadline exceeded"})
			logRequest(r, status, start)
		}
	}
}

func handleMetrics(appMetrics *metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := http.StatusOK
		incrementRequests(appMetrics)
		writeJSON(w, status, snapshotMetrics(appMetrics))
		logRequest(r, status, start)
	}
}

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	appConfig := config{Port: port}
	appMetrics := &metrics{}

	http.HandleFunc("/healthz", handleHealthz(appMetrics))
	http.HandleFunc("/work", handleWork(appMetrics))
	http.HandleFunc("/metrics", handleMetrics(appMetrics))

	address := ":" + appConfig.Port
	server := &http.Server{
		Addr:              address,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	serverErr := make(chan error, 1)

	go func() {
		log.Printf("event=startup address=%s routes=/healthz,/work,/metrics", address)
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
docker build -f Dockerfile -t go-scaling:module-014 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
docker rm -f go-scaling-module-014 >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-014 -p 8080:8080 -e APP_PORT=8080 go-scaling:module-014
docker logs -f go-scaling-module-014 &
LOG_PID=$!
```

The server is now running in the background, and the log follower keeps container output visible in the same terminal.

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/metrics
curl -i http://localhost:8080/work
curl http://localhost:8080/metrics
```

The `/work` request should return `504 Gateway Timeout` because the simulated work takes 3 seconds but the handler deadline is 2 seconds.

The service should log a timeout event and increment `work_timed_out_total`.

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-014
wait "$LOG_PID" 2>/dev/null || true
```

Graceful shutdown from Module 012 still applies.

## What Changed From Module 013

Module 013 added request lifetime handling and primitive instrumentation.

This module keeps request lifetime handling and counters. It adds explicit server-edge timeouts on `http.Server` and a handler-level deadline for `/work`.

## First-Principles Chain

```text
source code → server boundary → request → request context → handler deadline → bounded work → timeout outcome → counter → metrics response → logs → graceful shutdown → process exit
```

Source code configures the server boundary. A request arrives with a request context. The handler adds its own deadline around bounded work. If the deadline expires first, the timeout outcome is counted, returned as JSON, and logged. The server still exits cleanly through graceful shutdown.

## Why Server Timeouts And Deadlines Matter For Scaling Later

Larger services must protect their capacity from slow clients and slow work.

Server timeouts protect the connection boundary.

Handler deadlines protect expensive work inside the request.

Primitive metrics show whether work is completing or timing out before the system grows.
