// Package handlers exposes the calculator over HTTP: request decoding,
// dispatch to calc, and the mapping from domain errors to status codes.
package handlers

import (
	"errors"

	"github.com/srozen-ai/CalculatorApp/internal/calc"
)

// ErrUnknownOperation is returned when the operation is not supported.
var ErrUnknownOperation = errors.New("unknown operation")

// ErrMissingOperand is returned when a required operand is absent.
var ErrMissingOperand = errors.New("a required operand is missing")

// CalcRequest is the decoded body of a calculate request. Operands are
// pointers so an absent operand can be distinguished from an operand of zero.
type CalcRequest struct {
	Operation string   `json:"operation"`
	A         *float64 `json:"a"`
	B         *float64 `json:"b"`
}

// binary applies f to both operands, requiring each to be present.
func binary(req CalcRequest, f func(a, b float64) (float64, error)) (float64, error) {
	if req.A == nil || req.B == nil {
		return 0, ErrMissingOperand
	}
	return f(*req.A, *req.B)
}

// unary applies f to a, requiring it to be present. b is ignored.
func unary(req CalcRequest, f func(a float64) (float64, error)) (float64, error) {
	if req.A == nil {
		return 0, ErrMissingOperand
	}
	return f(*req.A)
}

// Calculate dispatches the request to the matching operation in calc.
func Calculate(req CalcRequest) (float64, error) {
	switch req.Operation {
	case "add":
		return binary(req, calc.Add)
	case "subtract":
		return binary(req, calc.Subtract)
	case "multiply":
		return binary(req, calc.Multiply)
	case "divide":
		return binary(req, calc.Divide)
	case "sqrt":
		return unary(req, calc.Sqrt)
	case "exponent":
		return binary(req, calc.Exponentiate)
	case "percent":
		return binary(req, calc.Percent)
	default:
		return 0, ErrUnknownOperation
	}
}
