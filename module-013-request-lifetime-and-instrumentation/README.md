# Module 013: Request Lifetime And Instrumentation

This module introduces the next scaling primitive of a Go program: a request has a lifetime, and scalable services stop doing work when that request no longer matters.

It also introduces ultra-primitive instrumentation: simple in-process counters exposed through a JSON `/metrics` endpoint.

## Request Lifetime

A request lifetime is the time between a client starting a request and that request finishing or being canceled.

The server should treat each request as bounded work.

## Canceled Requests

A request may stop mattering before the server finishes work.

The client might disconnect, give up, or hit its own timeout.

If the server keeps working after that, it wastes capacity.

## `r.Context()`

`r.Context()` returns the context for one HTTP request.

At a beginner level, the request context is the signal the handler can watch to know whether the request is still active.

## Context Cancellation

Context cancellation means the request context has been marked done.

For a handler, that means the request may no longer need a response.

## Watching Slow Work

The `/slow` handler simulates slow work with a timer.

It uses `select` to wait for either the timer to finish or `r.Context().Done()` to be canceled.

If the request is canceled first, the handler stops waiting on the slow work and returns without writing the success response.

Continuing canceled work wastes CPU time, memory, and request slots.

At scale, wasted work becomes dangerous because many canceled requests can consume capacity needed for active requests.

## Instrumentation

Instrumentation means the program records facts about what happened.

This module records a few counts in memory.

## Logs And Metrics

Logs show individual events. Metrics show patterns across requests.

The cancellation log shows one canceled request.

The `/metrics` response shows how many requests have started, completed, or canceled since the process started.

## Primitive Counters

This module uses simple counters instead of a metrics framework.

The counters are stored in the process memory and reset when the process exits.

## `sync.Mutex`

More than one request can arrive at the same time.

`sync.Mutex` protects the counters so one request updates them at a time.

This module uses the mutex only for the tiny counter updates and snapshots.

## `/metrics`

The `/metrics` route returns JSON counters:

```json
{
  "requests_total": 4,
  "slow_started_total": 2,
  "slow_completed_total": 1,
  "slow_canceled_total": 1
}
```

This is not Prometheus or OpenTelemetry yet.

It is only the primitive idea: count important outcomes and expose those counts as structured data.

## Graceful Shutdown Still Applies

Module 012 showed that a service process should exit cleanly when the operating system asks it to stop.

This module keeps that graceful shutdown shape with an explicit `http.Server`, `signal.NotifyContext`, `context.WithTimeout`, and `server.Shutdown`.

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
	slowStartedTotal   int
	slowCompletedTotal int
	slowCanceledTotal  int
}

type healthResponse struct {
	Status string `json:"status"`
}

type slowResponse struct {
	Status string `json:"status"`
}

type metricsResponse struct {
	RequestsTotal      int `json:"requests_total"`
	SlowStartedTotal   int `json:"slow_started_total"`
	SlowCompletedTotal int `json:"slow_completed_total"`
	SlowCanceledTotal  int `json:"slow_canceled_total"`
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

func incrementSlowStarted(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.slowStartedTotal++
}

func incrementSlowCompleted(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.slowCompletedTotal++
}

func incrementSlowCanceled(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.slowCanceledTotal++
}

func snapshotMetrics(appMetrics *metrics) metricsResponse {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()

	return metricsResponse{
		RequestsTotal:      appMetrics.requestsTotal,
		SlowStartedTotal:   appMetrics.slowStartedTotal,
		SlowCompletedTotal: appMetrics.slowCompletedTotal,
		SlowCanceledTotal:  appMetrics.slowCanceledTotal,
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

func handleSlow(appMetrics *metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		incrementRequests(appMetrics)
		incrementSlowStarted(appMetrics)
		log.Printf("event=slow_work_start path=/slow")

		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()

		select {
		case <-timer.C:
			status := http.StatusOK
			incrementSlowCompleted(appMetrics)
			log.Printf("event=slow_work_complete path=/slow")
			writeJSON(w, status, slowResponse{Status: "completed"})
			logRequest(r, status, start)
		case <-r.Context().Done():
			incrementSlowCanceled(appMetrics)
			log.Printf("event=request_canceled path=/slow error=%v", r.Context().Err())
			return
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
	http.HandleFunc("/slow", handleSlow(appMetrics))
	http.HandleFunc("/metrics", handleMetrics(appMetrics))

	address := ":" + appConfig.Port
	server := &http.Server{Addr: address}
	serverErr := make(chan error, 1)

	go func() {
		log.Printf("event=startup address=%s routes=/healthz,/slow,/metrics", address)
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
docker build -f Dockerfile -t go-scaling:module-013 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
docker rm -f go-scaling-module-013 >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-013 -p 8080:8080 -e APP_PORT=8080 go-scaling:module-013
docker logs -f go-scaling-module-013 &
LOG_PID=$!
```

The server is now running in the background, and the log follower keeps container output visible in the same terminal.

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/metrics
curl http://localhost:8080/slow
curl http://localhost:8080/metrics
```

The `/slow` request should take about 5 seconds when the request is allowed to finish.

From the same terminal, cancel a slow request:

```bash
curl --max-time 1 http://localhost:8080/slow || true
curl http://localhost:8080/metrics
```

`curl --max-time 1` gives up before the server finishes the simulated work. The request context is canceled. The server should stop the slow work, log the cancellation, and increment `slow_canceled_total`.

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-013
wait "$LOG_PID" 2>/dev/null || true
```

Graceful shutdown from Module 012 still applies.

## What Changed From Module 012

Module 012 added graceful shutdown for the service process.

This module keeps graceful shutdown and adds request lifetime handling. The `/slow` handler watches `r.Context().Done()` and stops simulated work when the request is canceled. The module also adds primitive counters and exposes them through `/metrics`.

## First-Principles Chain

```text
source code → request → request context → bounded work → cancellation → outcome → counter → metrics response → logs → graceful shutdown → process exit
```

Source code receives a request. The request has a context. The handler treats work as bounded by that context. Cancellation changes the outcome. The program counts outcomes, exposes counters as a metrics response, writes logs for individual events, and still exits cleanly through graceful shutdown.

## Why Request Lifetime And Instrumentation Matter For Scaling Later

Larger services handle many requests at the same time.

Request lifetime handling prevents canceled work from wasting capacity.

Instrumentation shows patterns across requests so a person can see whether the service is completing work or spending capacity on canceled work.
