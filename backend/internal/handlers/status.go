package handlers

import (
	"errors"
	"net/http"

	"github.com/srozen-ai/CalculatorApp/backend/internal/calc"
)

// errorMapping ties a domain error to the status and code returned for it.
type errorMapping struct {
	err    error
	status int
	code   string
}

// errorMappings is the single source of truth for error responses. Adding a
// sentinel means adding one row here, so status and code cannot drift apart.
var errorMappings = []errorMapping{
	{ErrUnknownOperation, http.StatusBadRequest, "UNKNOWN_OPERATION"},
	{ErrMissingOperand, http.StatusBadRequest, "MISSING_OPERAND"},
	{calc.ErrDivisionByZero, http.StatusBadRequest, "DIVISION_BY_ZERO"},
	{calc.ErrNegativeRoot, http.StatusBadRequest, "NEGATIVE_ROOT"},
	{calc.ErrZeroNegativePower, http.StatusBadRequest, "ZERO_NEGATIVE_POWER"},
	{calc.ErrNotFinite, http.StatusBadRequest, "NOT_FINITE"},
}

// responseFor returns the status and code for err, defaulting to 500.
func responseFor(err error) (int, string) {
	for _, m := range errorMappings {
		if errors.Is(err, m.err) {
			return m.status, m.code
		}
	}
	return http.StatusInternalServerError, "INTERNAL_ERROR"
}
