package cluster

import (
	"errors"
	"math"
	"testing"

	"fuzzylab/fuzzy/core"
	"fuzzylab/fuzzy/metrics"
	"fuzzylab/gen"
)

func lineMatrix(xs []float64) [][]float64 {
	dm := make([][]float64, len(xs))
	for i := range dm {
		dm[i] = make([]float64, len(xs))
		for j := range dm[i] {
			dm[i][j] = math.Abs(xs[i] - xs[j])
		}
	}
	return dm
}

var groups = lineMatrix([]float64{0, 1, 2, 50, 51, 52, 100, 101, 102})

func checkGroups(t *testing.T, r Result) {
	t.Helper()
	if len(r.Medoids) != 3 {
		t.Fatalf("clusters = %d, want 3", len(r.Medoids))
	}
	for g := 0; g < 9; g += 3 {
		if r.Labels[g] != r.Labels[g+1] || r.Labels[g] != r.Labels[g+2] {
			t.Errorf("group %d split: %v", g/3, r.Labels)
		}
	}
	if r.Labels[0] == r.Labels[3] || r.Labels[3] == r.Labels[6] {
		t.Errorf("groups merged: %v", r.Labels)
	}
}

func TestPAM(t *testing.T) {
	r, err := PAM(groups, 3)
	if err != nil {
		t.Fatal(err)
	}
	checkGroups(t, r)
	for c, m := range r.Medoids {
		if r.Labels[m] != c {
			t.Errorf("medoid %d not in its cluster", m)
		}
	}
}

func TestAgglomerative(t *testing.T) {
	r, err := Agglomerative(groups, 10)
	if err != nil {
		t.Fatal(err)
	}
	checkGroups(t, r)
	if r, _ := Agglomerative(groups, 1000); len(r.Medoids) != 1 {
		t.Errorf("large threshold: clusters = %d, want 1", len(r.Medoids))
	}
	if r, _ := Agglomerative(groups, 0); len(r.Medoids) != 9 {
		t.Errorf("zero threshold: clusters = %d, want 9", len(r.Medoids))
	}
}

func TestInvalid(t *testing.T) {
	if _, err := PAM(groups, 0); !errors.Is(err, core.ErrInvalidParams) {
		t.Errorf("k=0: err = %v", err)
	}
	if _, err := PAM(groups, 10); !errors.Is(err, core.ErrInvalidParams) {
		t.Errorf("k>n: err = %v", err)
	}
	if _, err := PAM(nil, 1); !errors.Is(err, core.ErrEmpty) {
		t.Errorf("empty: err = %v", err)
	}
	if _, err := Agglomerative([][]float64{{0, 1}}, 1); !errors.Is(err, core.ErrInvalidParams) {
		t.Errorf("non-square: err = %v", err)
	}
}

func TestPAMOnGeneratedClouds(t *testing.T) {
	items, truth, err := gen.Clouds(gen.CloudParams{
		Params: gen.Params{N: 200, A: gen.Range{Lo: 0, Hi: 1000}, W: gen.Range{Lo: 1, Hi: 5}, Seed: 3},
		K:      4, Noise: 0.005,
	})
	if err != nil {
		t.Fatal(err)
	}
	r, err := PAM(metrics.DistanceMatrix(items, metrics.CutsL1Distance{}), 4)
	if err != nil {
		t.Fatal(err)
	}
	pair := map[[2]int]bool{}
	for i := range truth {
		pair[[2]int{truth[i], r.Labels[i]}] = true
	}
	if len(pair) != 4 {
		t.Errorf("clusters do not match generated clouds: %v", pair)
	}
}
