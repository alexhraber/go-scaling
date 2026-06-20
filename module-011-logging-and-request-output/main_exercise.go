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
