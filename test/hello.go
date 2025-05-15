package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9001"
	}

	hostname, _ := os.Hostname()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from backend %s (port %s)\n", hostname, port)
	})

	log.Printf("Starting hello server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
