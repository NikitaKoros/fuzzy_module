package ops

import (
	"errors"
	"math"
	"testing"

	"fuzzylab/fuzzy/core"
	"fuzzylab/fuzzy/number"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func u(t *testing.T, a, b float64) number.Uniform {
	t.Helper()
	x, err := number.NewUniform(a, b)
	if err != nil {
		t.Fatal(err)
	}
	return x
}

func checkSupport(t *testing.T, name string, n core.Number, lo, hi float64) {
	t.Helper()
	s := n.Support()
	if !near(s.Lo, lo) || !near(s.Hi, hi) {
		t.Errorf("%s: support = %v, want [%g, %g]", name, s, lo, hi)
	}
}

func TestUniformCDFArithmetic(t *testing.T) {
	cfg := core.Config{}
	a, b := u(t, 1, 3), u(t, 2, 5)

	r, err := Add(a, b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	checkSupport(t, "Add", r, 3, 8)

	r, _ = Sub(a, b, cfg)
	checkSupport(t, "Sub", r, -4, 1)

	r, _ = Mul(a, b, cfg)
	checkSupport(t, "Mul", r, 2, 15)

	r, err = Div(a, b, cfg)
	if err != nil {
		t.Fatal(err)
	}
	checkSupport(t, "Div", r, 0.2, 1.5)

	r, _ = Neg(a, cfg)
	checkSupport(t, "Neg", r, -3, -1)

	r, _ = Scale(a, -2, cfg)
	checkSupport(t, "Scale", r, -6, -2)
}

func TestUniformCDFAdditionAlphaCuts(t *testing.T) {
	a, b := u(t, 1, 3), u(t, 2, 5)
	sum, err := Add(a, b, core.Config{})
	if err != nil {
		t.Fatal(err)
	}
	cut, ok := sum.AlphaCut(0.5)
	if !ok || !near(cut.Lo, 5.5) || !near(cut.Hi, 8) {
		t.Errorf("sum α-cut at 0.5 = %v, %v; want [5.5, 8]", cut, ok)
	}
	if !near(sum.Membership(5.5), 0.5) || !near(sum.Membership(8), 1) {
		t.Errorf("unexpected sum membership at left boundary or peak")
	}
}

func TestDivByZeroSupport(t *testing.T) {
	_, err := Div(u(t, 1, 2), u(t, -1, 1), core.Config{})
	if !errors.Is(err, core.ErrDivByZero) {
		t.Errorf("err = %v", err)
	}
}

type tri struct{ a, m, b float64 }

func (t tri) Membership(x float64) float64 {
	switch {
	case x < t.a || x > t.b:
		return 0
	case x <= t.m:
		return (x - t.a) / (t.m - t.a)
	default:
		return (t.b - x) / (t.b - t.m)
	}
}
func (t tri) Support() core.Interval { return core.Interval{Lo: t.a, Hi: t.b} }
func (t tri) AlphaCut(al float64) (core.Interval, bool) {
	if al < 0 || al > 1 {
		return core.Interval{}, false
	}
	return core.Interval{Lo: t.a + al*(t.m-t.a), Hi: t.b - al*(t.b-t.m)}, true
}
func (t tri) Height() float64 { return 1 }
func (t tri) Kind() string    { return "tri" }

// Сумма треугольных (0,1,2) и (1,2,3) — треугольная (1,3,5).
func TestTriangleSum(t *testing.T) {
	r, err := Add(tri{0, 1, 2}, tri{1, 2, 3}, core.Config{})
	if err != nil {
		t.Fatal(err)
	}
	want := tri{1, 3, 5}
	for _, x := range []float64{0.5, 1, 1.5, 2, 2.5, 3, 4, 5, 5.5} {
		if got := r.Membership(x); !near(got, want.Membership(x)) {
			t.Errorf("μ(%g) = %g, want %g", x, got, want.Membership(x))
		}
	}
}

// Цепочка: результат операции снова является core.Number.
func TestChain(t *testing.T) {
	calc := Calculator{}
	s, err := calc.Add(u(t, 1, 2), u(t, 3, 4)) // [4,6]
	if err != nil {
		t.Fatal(err)
	}
	p, err := calc.Mul(s, u(t, 2, 3)) // [8,18]
	if err != nil {
		t.Fatal(err)
	}
	checkSupport(t, "chain", p, 8, 18)
}

func TestCustomOperation(t *testing.T) {
	r, err := Calculator{}.Apply(u(t, 1, 5), u(t, 3, 4), MaxOp)
	if err != nil {
		t.Fatal(err)
	}
	checkSupport(t, "Max", r, 3, 5)
}
