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
	requestsTotal        int
	healthRequestsTotal  int
	readyRequestsTotal   int
	slowStartedTotal     int
	slowCompletedTotal   int
	slowCanceledTotal    int
}

type healthResponse struct {
	Status string `json:"status"`
}

type readyResponse struct {
	Status string `json:"status"`
}

type slowResponse struct {
	Status string `json:"status"`
}

type metricsResponse struct {
	RequestsTotal        int `json:"requests_total"`
	HealthRequestsTotal  int `json:"health_requests_total"`
	ReadyRequestsTotal   int `json:"ready_requests_total"`
	SlowStartedTotal     int `json:"slow_started_total"`
	SlowCompletedTotal   int `json:"slow_completed_total"`
	SlowCanceledTotal    int `json:"slow_canceled_total"`
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

func incrementHealthRequests(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.healthRequestsTotal++
}

func incrementReadyRequests(appMetrics *metrics) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	appMetrics.readyRequestsTotal++
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
		RequestsTotal:        appMetrics.requestsTotal,
		HealthRequestsTotal:  appMetrics.healthRequestsTotal,
		ReadyRequestsTotal:   appMetrics.readyRequestsTotal,
		SlowStartedTotal:     appMetrics.slowStartedTotal,
		SlowCompletedTotal:   appMetrics.slowCompletedTotal,
		SlowCanceledTotal:    appMetrics.slowCanceledTotal,
	}
}

func handleHealthz(appMetrics *metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := http.StatusOK
		incrementRequests(appMetrics)
		incrementHealthRequests(appMetrics)
		writeJSON(w, status, healthResponse{Status: "ok"})
		logRequest(r, status, start)
	}
}

func handleReadyz(appMetrics *metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		status := http.StatusOK
		incrementRequests(appMetrics)
		incrementReadyRequests(appMetrics)
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

		timer := time.NewTimer(3 * time.Second)
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
	http.HandleFunc("/readyz", handleReadyz(appMetrics))
	http.HandleFunc("/slow", handleSlow(appMetrics))
	http.HandleFunc("/metrics", handleMetrics(appMetrics))

	address := ":" + appConfig.Port
	server := &http.Server{Addr: address}
	serverErr := make(chan error, 1)

	go func() {
		log.Printf("event=startup address=%s routes=/healthz,/readyz,/slow,/metrics", address)
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
