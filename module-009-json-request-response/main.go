package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type echoRequest struct {
	Message string `json:"message"`
}

type echoResponse struct {
	Received string `json:"received"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type healthResponse struct {
	Status string `json:"status"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	var request echoRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	writeJSON(w, http.StatusOK, echoResponse{Received: request.Message})
}

func main() {
	http.HandleFunc("/healthz", handleHealthz)
	http.HandleFunc("/echo", handleEcho)

	fmt.Fprintln(os.Stdout, "stdout: starting HTTP server on :8080 with routes /healthz and /echo")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Fprintln(os.Stderr, "server failed:", err)
		return
	}
}
