package main

import (
	"fmt"
	"net/http"
	"os"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Response from go net/http HTTP server")
	fmt.Fprintln(w, "Thanks for stopping by")
}

func main() {
	http.HandleFunc("/", handleRoot)

	fmt.Fprintln(os.Stdout, "stdout: starting go net/http HTTP server on :8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Fprintln(os.Stderr, "server failed:", err)
		return
	}
}
