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
		port = "8080"
	}

	hostname, _ := os.Hostname()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from backend %s (port %s)\n", hostname, port)

		fmt.Fprintln(w, "\nIncoming Headers:")
		for name, values := range r.Header {
			for _, value := range values {
				fmt.Fprintf(w, "%s: %s\n", name, value)
				log.Printf("Header %s: %s", name, value)
			}
		}
	})

	log.Printf("Starting hello server on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
