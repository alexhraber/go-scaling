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
