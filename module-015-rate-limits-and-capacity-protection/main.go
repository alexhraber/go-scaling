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

	go func() {
		log.Printf("event=startup address=%s routes=/healthz,/work,/metrics", address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("event=server_failed error=%v", err)
		}
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-shutdownSignal.Done()
	log.Printf("event=shutdown_start")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("event=shutdown_failed error=%v", err)
		return
	}

	log.Printf("event=shutdown_complete")
}
