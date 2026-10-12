package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// / writeJSON encodes v and writes it with the given status. Encoding happens
// before the status is sent, so a marshaling failure cannot go out as a 200.
func writeJSON(w http.ResponseWriter, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		log.Printf("writeJSON: marshaling response: %v", err)
		body = []byte(`{"error":{"code":"INTERNAL_ERROR","message":"an unexpected error occurred"}}`)
		status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		log.Printf("writeJSON: writing response: %v", err)
	}
}

// writeError sends an error envelope with the given status, code and message.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{Error: ErrorBody{Code: code, Message: message}})
}
