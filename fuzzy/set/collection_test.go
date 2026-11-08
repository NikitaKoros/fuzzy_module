package set

import (
	"errors"
	"testing"

	"fuzzylab/fuzzy/core"
	"fuzzylab/fuzzy/number"
)

func u(t *testing.T, a, b float64) core.Number {
	t.Helper()
	x, err := number.NewUniform(a, b)
	if err != nil {
		t.Fatal(err)
	}
	return x
}

func TestCollection(t *testing.T) {
	c := New(u(t, 0, 1), u(t, 5, 8))
	c.Add(u(t, -2, 0.5))
	if c.Len() != 3 {
		t.Fatalf("Len = %d", c.Len())
	}

	b, err := c.Bounds()
	if err != nil || b != (core.Interval{Lo: -2, Hi: 8}) {
		t.Errorf("Bounds = %v, %v", b, err)
	}

	if _, err := c.Get(3); !errors.Is(err, core.ErrIndex) {
		t.Errorf("Get(3) err = %v", err)
	}
	if err := c.Remove(1); err != nil || c.Len() != 2 {
		t.Errorf("Remove: %v, len %d", err, c.Len())
	}

	wide := c.Filter(func(n core.Number) bool { return n.Support().Width() > 1 })
	if wide.Len() != 1 {
		t.Errorf("Filter len = %d", wide.Len())
	}

	// Slice возвращает копию
	s := c.Slice()
	s[0] = nil
	if g, _ := c.Get(0); g == nil {
		t.Error("Slice must return a copy")
	}

	if _, err := New().Bounds(); !errors.Is(err, core.ErrEmpty) {
		t.Errorf("empty Bounds err = %v", err)
	}
}
