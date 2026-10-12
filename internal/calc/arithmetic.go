// Package calc implements the arithmetic operations the API exposes.
// Every operation returns a finite float64 or one of the errors below;
// the package knows nothing about HTTP.
package calc

import (
	"errors"
	"math"
)

var (
	// ErrDivisionByZero is returned when the divisor is zero.
	ErrDivisionByZero = errors.New("division by zero")

	// ErrNegativeRoot is returned when the square root of a negative number is requested.
	ErrNegativeRoot = errors.New("square root of a negative number")

	// ErrNotFinite is returned when a result overflows or is undefined.
	ErrNotFinite = errors.New("result is not a finite number")

	// ErrZeroNegativePower is returned when zero is raised to a negative power.
	ErrZeroNegativePower = errors.New("zero raised to a negative power")
)

// Add returns the sum of a and b.
func Add(a, b float64) (float64, error) {
	return finite(a + b)
}

// Subtract returns the difference a - b.
func Subtract(a, b float64) (float64, error) {
	return finite(a - b)
}

// Multiply returns the product of a and b.
func Multiply(a, b float64) (float64, error) {
	return finite(a * b)
}

// Divide returns the quotient a / b.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return finite(a / b)
}

// Exponentiate returns a raised to the power of b.
func Exponentiate(a, b float64) (float64, error) {
	if a == 0 && b < 0 {
		return 0, ErrZeroNegativePower
	}
	return finite(math.Pow(a, b))
}

// Sqrt returns the square root of a.
func Sqrt(a float64) (float64, error) {
	if a < 0 {
		return 0, ErrNegativeRoot
	}
	return finite(math.Sqrt(a))
}

// Percent returns a percent of b (a% of b).
func Percent(a, b float64) (float64, error) {
	return finite((a / 100) * b)
}

// finite returns v, or ErrNotFinite if v is infinite or NaN.
func finite(v float64) (float64, error) {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return 0, ErrNotFinite
	}
	return v, nil
}
