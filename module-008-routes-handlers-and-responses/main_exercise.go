package main

import (
	"fmt"
	"net/http"
	"os"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "welcome to module 008")
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "good")
}

func handleModule(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Module 008: routes, handlers, and responses")
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "route not found")
}

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/healthz", handleHealthz)
	http.HandleFunc("/module", handleModule)

	fmt.Fprintln(os.Stdout, "stdout: starting HTTP server on :8080 with routes /, /healthz, and /module")

	if err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			handleRoot(w, r)
		case "/healthz":
			handleHealthz(w, r)
		case "/module":
			handleModule(w, r)
		default:
			handleNotFound(w, r)
		}
	})); err != nil {
		fmt.Fprintln(os.Stderr, "server failed:", err)
		return
	}
}
