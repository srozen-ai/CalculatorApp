package handlers

import "net/http"

// NewRouter returns the API router with all routes registered.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/calculate", CalculateHandler)
	return mux
}
