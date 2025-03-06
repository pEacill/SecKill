package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
