package gen

import (
	"errors"
	"testing"

	"fuzzylab/fuzzy/core"
)

var params = Params{N: 150, A: Range{-50, 100}, W: Range{1, 20}, Seed: 1}

func TestUniform(t *testing.T) {
	items, err := Uniform(params)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != params.N {
		t.Fatalf("len = %d, want %d", len(items), params.N)
	}
	for _, it := range items {
		s := it.Support()
		if s.Lo < params.A.Lo || s.Lo > params.A.Hi || s.Width() < params.W.Lo || s.Width() > params.W.Hi {
			t.Fatalf("out of ranges: %v", s)
		}
	}
	again, _ := Uniform(params)
	if items[7].Support() != again[7].Support() {
		t.Error("same seed must give same data")
	}
}

func TestClouds(t *testing.T) {
	items, labels, err := Clouds(CloudParams{Params: params, K: 4, Noise: 0.02})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != params.N || len(labels) != params.N {
		t.Fatalf("len = %d/%d, want %d", len(items), len(labels), params.N)
	}
	seen := map[int]bool{}
	for _, l := range labels {
		seen[l] = true
	}
	if len(seen) != 4 {
		t.Errorf("clusters = %d, want 4", len(seen))
	}
}

func TestInvalid(t *testing.T) {
	bad := []Params{
		{N: 0, A: Range{0, 1}, W: Range{1, 2}},
		{N: 10, A: Range{1, 0}, W: Range{1, 2}},
		{N: 10, A: Range{0, 1}, W: Range{0, 2}},
	}
	for _, p := range bad {
		if _, err := Uniform(p); !errors.Is(err, core.ErrInvalidParams) {
			t.Errorf("%+v: err = %v", p, err)
		}
	}
	if _, _, err := Clouds(CloudParams{Params: params, K: 0}); !errors.Is(err, core.ErrInvalidParams) {
		t.Errorf("K=0: err = %v", err)
	}
}
