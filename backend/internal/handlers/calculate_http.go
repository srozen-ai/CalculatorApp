package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strconv"
)

// maxBodyBytes caps the request body to prevent unbounded reads.
const maxBodyBytes = 1 << 20 // 1 MiB

// resultPrecision is the number of significant digits results are rounded to
// before encoding. Significant digits rather than decimal places, so that
// neither very large nor very small results are destroyed by rounding.
const resultPrecision = 15

// CalcResponse is the successful response body.
type CalcResponse struct {
	Result float64 `json:"result"`
}

// ErrorBody describes a failure in a stable, machine-readable shape.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse wraps ErrorBody so every failure has the same envelope.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE",
			"Content-Type must be application/json")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req CalcRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "BODY_TOO_LARGE",
				"request body is too large")
			return
		}
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "request body must be valid JSON")
		return
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "INVALID_BODY",
			"request body must contain a single JSON object")
		return
	}

	result, err := Calculate(req)
	if err != nil {
		status, code := responseFor(err)
		if status == http.StatusInternalServerError {
			log.Printf("calculate: unexpected error: %v", err)
			writeError(w, status, code, "an unexpected error occurred")
			return
		}
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, CalcResponse{Result: round(result)})
}

// round returns v rounded to resultPrecision significant digits.
func round(v float64) float64 {
	r, err := strconv.ParseFloat(strconv.FormatFloat(v, 'g', resultPrecision, 64), 64)
	if err != nil {
		return v
	}
	return r
}
