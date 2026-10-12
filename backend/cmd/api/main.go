// Command api serves the calculator HTTP API.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/srozen-ai/CalculatorApp/backend/internal/handlers"
	"github.com/srozen-ai/CalculatorApp/backend/internal/httpx"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           httpx.CORS(handlers.NewRouter()),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	log.Printf("listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
