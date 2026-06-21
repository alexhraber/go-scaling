# Module 015: Rate Limits And Capacity Protection

This module introduces the next scaling primitive of a Go program: a service must refuse excess work before accepted work exhausts the process.

The core primitive is:

```text
A scalable service protects capacity by deciding whether to admit work before doing the work.
```

## Admission Control

Admission control means deciding whether work is allowed to enter the service.

The decision happens before the expensive work begins.

## Rate Limit

A rate limit is a rule that allows only a certain amount of work during a period of time.

This module allows two `/work` requests every 10 seconds.

## Capacity Protection

Rate limits protect capacity by refusing excess work.

Rejecting work can be healthier than accepting everything because accepted work consumes CPU time, memory, and request slots.

If a process accepts more work than it can handle, the accepted work can exhaust the process and make every request slower.

## `429 Too Many Requests`

HTTP `429 Too Many Requests` means the client sent more requests than the service is willing to accept right now.

In this module, the third rapid `/work` request returns status `429`.

## `Retry-After`

`Retry-After` tells the client how long to wait before trying again.

This module sets:

```text
Retry-After: 10
```

The value matches the 10-second fixed window.

## Fixed-Window Limiter

A fixed-window limiter counts requests during a fixed period of time.

When the window expires, the count resets.

This module uses one 10-second window and allows two admitted `/work` requests in that window.

## Process-Local And Primitive

This limiter lives inside one process.

It is intentionally primitive. It does not coordinate with other processes and it does not use external storage.

This is not distributed rate limiting yet.

## Rate Limits And Timeouts

Timeouts limit how long accepted work can run. Rate limits decide whether work should be accepted at all.

Module 014 timeouts still matter after work is admitted because accepted work still needs time boundaries.

## Primitive Counters

This module still uses primitive counters instead of a metrics framework.

The counters are stored in process memory and protected with `sync.Mutex`.

## `/metrics`

The `/metrics` route returns JSON counters:

```json
{
  "requests_total": 5,
  "work_admitted_total": 2,
  "work_rejected_total": 1,
  "work_completed_total": 2
}
```

It is okay that `requests_total` includes calls to `/healthz` and `/metrics`.

## Earlier Modules Still Apply

Module 012 graceful shutdown still applies. The server still waits for an operating-system signal and exits cleanly.

Module 013 request lifetime and instrumentation still apply. The server still counts important outcomes and exposes those counts through `/metrics`.

Module 014 server timeouts still apply. The explicit `http.Server` timeout fields still protect the connection boundary.

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

type fixedWindowLimiter struct {
	mu          sync.Mutex
	limit       int
	window      time.Duration
	windowStart time.Time
	count       int
}

type metrics struct {
	mu                 sync.Mutex
	requestsTotal      int
	workAdmittedTotal  int
	workRejectedTotal  int
	workCompletedTotal int
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
	WorkAdmittedTotal  int `json:"work_admitted_total"`
	WorkRejectedTotal  int `json:"work_rejected_total"`
	WorkCompletedTotal int `json:"work_completed_total"`
}

func (l *fixedWindowLimiter) allow(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.windowStart) >= l.window {
		l.windowStart = now
		l.count = 0
	}

	if l.count >= l.limit {
		return false
	}

	l.count++
	return true
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

func incrementWorkAdmitted(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.workAdmittedTotal++
}

func incrementWorkRejected(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.workRejectedTotal++
}

func incrementWorkCompleted(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.workCompletedTotal++
}

func snapshotMetrics(appMetrics *metrics) metricsResponse {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()

	return metricsResponse{
		RequestsTotal:      appMetrics.requestsTotal,
		WorkAdmittedTotal:  appMetrics.workAdmittedTotal,
		WorkRejectedTotal:  appMetrics.workRejectedTotal,
		WorkCompletedTotal: appMetrics.workCompletedTotal,
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

func handleWork(appMetrics *metrics, limiter *fixedWindowLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		incrementRequests(appMetrics)

		if !limiter.allow(time.Now()) {
			status := http.StatusTooManyRequests
			incrementWorkRejected(appMetrics)
			w.Header().Set("Retry-After", "10")
			log.Printf("event=work_rejected path=/work reason=rate_limited")
			writeJSON(w, status, errorResponse{Error: "rate limit exceeded"})
			logRequest(r, status, start)
			return
		}

		incrementWorkAdmitted(appMetrics)
		log.Printf("event=work_admitted path=/work")

		timer := time.NewTimer(500 * time.Millisecond)
		defer timer.Stop()

		<-timer.C

		status := http.StatusOK
		incrementWorkCompleted(appMetrics)
		writeJSON(w, status, workResponse{Status: "completed"})
		logRequest(r, status, start)
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
	limiter := &fixedWindowLimiter{
		limit:       2,
		window:      10 * time.Second,
		windowStart: time.Now(),
	}

	http.HandleFunc("/healthz", handleHealthz(appMetrics))
	http.HandleFunc("/work", handleWork(appMetrics, limiter))
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
docker build -f Dockerfile -t go-scaling:module-015 .
```

The Dockerfile compiles `main.go` into a binary and copies that binary into the shared runtime image.

## Run The Module Image

From this module directory, run:

```bash
docker rm -f go-scaling-module-015 >/dev/null 2>&1 || true
docker run --rm -d --name go-scaling-module-015 -p 8080:8080 -e APP_PORT=8080 go-scaling:module-015
docker logs -f go-scaling-module-015 &
LOG_PID=$!
```

The server is now running in the background, and the log follower keeps container output visible in the same terminal.

From the same terminal, call the server:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/metrics
curl -i http://localhost:8080/work
curl -i http://localhost:8080/work
curl -i http://localhost:8080/work
curl http://localhost:8080/metrics
```

The first two `/work` requests should return `200 OK`.

The third rapid `/work` request should return `429 Too Many Requests`.

The rejected request should include:

```text
Retry-After: 10
```

The service should log admitted work, rejected work, and increment the corresponding counters.

When you are done, stop the container deterministically by name:

```bash
docker stop go-scaling-module-015
wait "$LOG_PID" 2>/dev/null || true
```

Graceful shutdown from Module 012 still applies.

## What Changed From Module 014

Module 014 added server timeouts and a handler-level deadline.

This module keeps the explicit `http.Server` timeout fields and adds admission control before `/work` starts. The fixed-window limiter admits the first two `/work` requests in a 10-second window and rejects the third.

## First-Principles Chain

```text
source code → server boundary → request → admission decision → admitted work or rejected work → outcome → counter → metrics response → logs → graceful shutdown → process exit
```

Source code configures the server boundary. A request arrives. The service makes an admission decision. Admitted work runs and completes. Rejected work returns before doing the work. Outcomes are counted, exposed through `/metrics`, logged, and the process still exits cleanly through graceful shutdown.

## Why Rate Limits And Capacity Protection Matter For Scaling Later

Larger services must protect themselves from more work than one process can handle.

Rate limits make the service choose which work enters the process before accepted work exhausts capacity.
