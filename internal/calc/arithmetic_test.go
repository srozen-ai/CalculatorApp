package calc

import (
	"errors"
	"math"
	"testing"
)

// epsilon is the tolerance used when comparing floats. Tests that pin an
// exact value compare directly instead; see TestAddFloatPrecision.
const epsilon = 1e-9

// binaryCase is one test case for a two-operand operation.
type binaryCase struct {
	name    string
	a, b    float64
	want    float64
	wantErr error
}

// runBinary runs tests against f, reporting failures as op(a, b).
func runBinary(t *testing.T, op string, f func(a, b float64) (float64, error), tests []binaryCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f(tt.a, tt.b)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("%s(%v, %v) error = %v, want %v", op, tt.a, tt.b, err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if math.Abs(got-tt.want) > epsilon {
				t.Errorf("%s(%v, %v) = %v, want %v", op, tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	runBinary(t, "Add", Add, []binaryCase{
		{name: "positive operands", a: 30, b: 5, want: 35},
		{name: "overflow to infinity", a: math.MaxFloat64, b: math.MaxFloat64, wantErr: ErrNotFinite},
	})
}

func TestSubtract(t *testing.T) {
	runBinary(t, "Subtract", Subtract, []binaryCase{
		{name: "positive operands", a: 30, b: 5, want: 25},
		{name: "overflow to infinity", a: math.MaxFloat64, b: -math.MaxFloat64, wantErr: ErrNotFinite},
	})
}

func TestMultiply(t *testing.T) {
	runBinary(t, "Multiply", Multiply, []binaryCase{
		{name: "positive operands", a: 30, b: 5, want: 150},
		{name: "overflow to infinity", a: math.MaxFloat64, b: 2, wantErr: ErrNotFinite},
	})
}

func TestDivide(t *testing.T) {
	runBinary(t, "Divide", Divide, []binaryCase{
		{name: "positive operands", a: 10, b: 4, want: 2.5},
		{name: "divisor is zero", a: 1, b: 0, wantErr: ErrDivisionByZero},
		{name: "zero divided by zero", a: 0, b: 0, wantErr: ErrDivisionByZero},
		{name: "overflow to infinity", a: math.MaxFloat64, b: 0.5, wantErr: ErrNotFinite},
	})
}

func TestExponentiate(t *testing.T) {
	runBinary(t, "Exponentiate", Exponentiate, []binaryCase{
		{name: "positive operands", a: 2, b: 3, want: 8},
		{name: "zero to a negative power", a: 0, b: -1, wantErr: ErrZeroNegativePower},
		{name: "negative exponent", a: 2, b: -1, want: 0.5},
		{name: "zero to the power of zero", a: 0, b: 0, want: 1},
		{name: "negative base with fractional exponent", a: -8, b: 0.5, wantErr: ErrNotFinite},
		{name: "overflow to infinity", a: 10, b: 400, wantErr: ErrNotFinite},
	})
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		name    string
		a       float64
		want    float64
		wantErr error
	}{
		{name: "positive operands", a: 9, want: 3},
		{name: "negative number", a: -9, wantErr: ErrNegativeRoot},
		{name: "zero returns 0", a: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sqrt(tt.a)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Sqrt(%v) error = %v, want %v", tt.a, err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if math.Abs(got-tt.want) > epsilon {
				t.Errorf("Sqrt(%v) = %v, want %v", tt.a, got, tt.want)
			}
		})
	}
}

func TestPercent(t *testing.T) {
	runBinary(t, "Percent", Percent, []binaryCase{
		{name: "positive operands", a: 50, b: 200, want: 100},
		{name: "overflow to infinity", a: 200, b: math.MaxFloat64, wantErr: ErrNotFinite},
	})
}

func TestAddFloatPrecision(t *testing.T) {
	const want = 0.30000000000000004

	got, err := Add(0.1, 0.2)
	if err != nil {
		t.Fatalf("Add(0.1, 0.2) error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Add(0.1, 0.2) = %v, want %v", got, want)
	}
}

// TestPercentIntermediatePrecision pins the error introduced by dividing
// before multiplying. (a/100)*b is used so that a large b cannot overflow in
// an intermediate step, at the cost of sub-epsilon error on common inputs:
// 7% of 100 is 7.000000000000001, not 7. The handler's rounding absorbs this.
// Exact comparison is required here; a tolerance would accept either value
// and the test would prove nothing.
func TestPercentIntermediatePrecision(t *testing.T) {
	const want = 7.000000000000001

	got, err := Percent(7, 100)
	if err != nil {
		t.Fatalf("Percent(7, 100) error = %v, want nil", err)
	}
	if got != want {
		t.Errorf("Percent(7, 100) = %v, want %v", got, want)
	}
}
