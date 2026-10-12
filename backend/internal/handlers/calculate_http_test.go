package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCalculateHandler(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		contentType string //
		wantStatus  int
		wantCode    string  // empty when success
		wantResult  float64 // only checked on success
	}{
		{
			name: "valid request", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"add","a":0.1,"b":0.2}`,
			wantStatus: http.StatusOK, wantResult: 0.3,
		},
		{
			name: "division by zero", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"divide","a":1,"b":0}`,
			wantStatus: http.StatusBadRequest, wantCode: "DIVISION_BY_ZERO",
		},
		{
			name: "missing operand", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"add","a":1}`,
			wantStatus: http.StatusBadRequest, wantCode: "MISSING_OPERAND",
		},
		{
			name: "zero operand is not missing", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"add","a":1,"b":0}`,
			wantStatus: http.StatusOK, wantResult: 1,
		},
		{
			name: "unknown operation", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"blorp","a":1,"b":2}`,
			wantStatus: http.StatusBadRequest, wantCode: "UNKNOWN_OPERATION",
		},
		{
			name: "malformed json", method: http.MethodPost, path: "/api/calculate",
			body:       `not json`,
			wantStatus: http.StatusBadRequest, wantCode: "INVALID_BODY",
		},
		{
			name: "empty body", method: http.MethodPost, path: "/api/calculate",
			body:       ``,
			wantStatus: http.StatusBadRequest, wantCode: "INVALID_BODY",
		},
		{
			name: "non-numeric operand", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"add","a":"abc","b":2}`,
			wantStatus: http.StatusBadRequest, wantCode: "INVALID_BODY",
		},
		{
			name: "wrong method", method: http.MethodGet, path: "/api/calculate",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name: "unknown path", method: http.MethodPost, path: "/api/nope",
			wantStatus: http.StatusNotFound,
		},
		{
			name: "large result is not rounded", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"multiply","a":1e300,"b":10}`,
			wantStatus: http.StatusOK, wantResult: 1e301,
		},
		{
			name: "wrong content type", method: http.MethodPost, path: "/api/calculate",
			body: `{"operation":"add","a":1,"b":2}`, contentType: "text/plain",
			wantStatus: http.StatusUnsupportedMediaType, wantCode: "UNSUPPORTED_MEDIA_TYPE",
		},
		{
			name: "missing first operand", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"sqrt"}`,
			wantStatus: http.StatusBadRequest, wantCode: "MISSING_OPERAND",
		},
		{
			name: "unary operation without b", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"sqrt","a":9}`,
			wantStatus: http.StatusOK, wantResult: 3,
		},
		{
			name: "negative root", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"sqrt","a":-9}`,
			wantStatus: http.StatusBadRequest, wantCode: "NEGATIVE_ROOT",
		},
		{
			name: "zero to a negative power", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"exponent","a":0,"b":-1}`,
			wantStatus: http.StatusBadRequest, wantCode: "ZERO_NEGATIVE_POWER",
		},
		{
			name: "result overflows", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"multiply","a":1e308,"b":10}`,
			wantStatus: http.StatusBadRequest, wantCode: "NOT_FINITE",
		},
		{
			name: "small result is not rounded away", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"divide","a":1,"b":1e12}`,
			wantStatus: http.StatusOK, wantResult: 1e-12,
		},
		{
			name: "trailing data after object", method: http.MethodPost, path: "/api/calculate",
			body:       `{"operation":"add","a":1,"b":2} garbage`,
			wantStatus: http.StatusBadRequest, wantCode: "INVALID_BODY",
		},
	}

	router := NewRouter()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.contentType == "" {
				req.Header.Set("Content-Type", "application/json")
			} else {
				req.Header.Set("Content-Type", tt.contentType)
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus == http.StatusMethodNotAllowed || tt.wantStatus == http.StatusNotFound {
				return // these come from the mux and have no JSON body
			}

			if tt.wantCode != "" {
				var got ErrorResponse
				if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
					t.Fatalf("decoding error body: %v", err)
				}
				if got.Error.Code != tt.wantCode {
					t.Errorf("error code = %q, want %q", got.Error.Code, tt.wantCode)
				}
				return
			}

			var got CalcResponse
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decoding success body: %v", err)
			}
			if got.Result != tt.wantResult {
				t.Errorf("result = %v, want %v", got.Result, tt.wantResult)
			}
		})
	}
}

func TestCalculateHandlerOversizedBody(t *testing.T) {
	// MaxBytesReader cuts the body off past maxBodyBytes, surfacing a
	// *http.MaxBytesError that the handler maps to 413.
	padding := strings.Repeat("0", maxBodyBytes+1)
	body := `{"operation":"add","a":1,"b":2,"pad":"` + padding + `"}`

	req := httptest.NewRequest(http.MethodPost, "/api/calculate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestResponseForUnknownError(t *testing.T) {
	status, code := responseFor(errors.New("something unexpected"))

	if status != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", status, http.StatusInternalServerError)
	}
	if code != "INTERNAL_ERROR" {
		t.Errorf("code = %q, want INTERNAL_ERROR", code)
	}
}
