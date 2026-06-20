package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

type config struct {
	IP      string
	Port    string
	Message string
}

type healthResponse struct {
	Status string `json:"status"`
}

type configResponse struct {
	IP      string `json:"ip"`
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
			IP:      appConfig.IP,
			Port:    appConfig.Port,
			Message: appConfig.Message,
		})
	}
}

func main() {
	message := flag.String("message", "message configuration flag", "message returned by /config")
	flag.Parse()

	ip := os.Getenv("APP_IP")
	if ip == "" {
		ip = "127.0.0.1"
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	appConfig := config{
		IP:      ip,
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
