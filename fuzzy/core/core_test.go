package core

import (
	"errors"
	"testing"
)

func TestIntervalArithmetic(t *testing.T) {
	a, b := Interval{1, 2}, Interval{3, 5}
	if got := a.Add(b); got != (Interval{4, 7}) {
		t.Errorf("Add = %v", got)
	}
	if got := a.Sub(b); got != (Interval{-4, -1}) {
		t.Errorf("Sub = %v", got)
	}
	if got := (Interval{-2, 3}).Mul(Interval{-1, 4}); got != (Interval{-8, 12}) {
		t.Errorf("Mul = %v", got)
	}
	got, err := (Interval{2, 4}).Div(Interval{1, 2})
	if err != nil || got != (Interval{1, 4}) {
		t.Errorf("Div = %v, %v", got, err)
	}
	if _, err := a.Div(Interval{-1, 1}); !errors.Is(err, ErrDivByZero) {
		t.Errorf("want ErrDivByZero, got %v", err)
	}
	if got := a.Scale(-2); got != (Interval{-4, -2}) {
		t.Errorf("Scale = %v", got)
	}
}

func TestAlphaLevels(t *testing.T) {
	l := AlphaLevels(1, 4)
	want := []float64{0, 0.25, 0.5, 0.75, 1}
	if len(l) != len(want) {
		t.Fatalf("len = %d", len(l))
	}
	for i := range want {
		if l[i] != want[i] {
			t.Errorf("l[%d] = %g", i, l[i])
		}
	}
}

func TestConfigDefaults(t *testing.T) {
	c := Config{}.WithDefaults()
	if c.AlphaLevels != DefaultAlphaLevels || c.GridPoints != DefaultGridPoints || c.Eps != DefaultEps {
		t.Errorf("defaults not applied: %+v", c)
	}
}
