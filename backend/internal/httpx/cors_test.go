package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	t.Setenv("ALLOWED_ORIGIN", "http://localhost:5173")

	var reached bool
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})

	t.Run("preflight is answered without reaching the handler", func(t *testing.T) {
		reached = false
		req := httptest.NewRequest(http.MethodOptions, "/api/calculate", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)
		rec := httptest.NewRecorder()

		CORS(next).ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
			t.Errorf("Access-Control-Allow-Origin = %q, want the configured origin", got)
		}
		if rec.Header().Get("Access-Control-Allow-Methods") == "" {
			t.Error("Access-Control-Allow-Methods is missing")
		}
		if reached {
			t.Error("preflight reached the next handler")
		}
	})

	t.Run("a normal request reaches the next handler", func(t *testing.T) {
		reached = false
		req := httptest.NewRequest(http.MethodPost, "/api/calculate", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		rec := httptest.NewRecorder()

		CORS(next).ServeHTTP(rec, req)

		if !reached {
			t.Error("request did not reach the next handler")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})
}
