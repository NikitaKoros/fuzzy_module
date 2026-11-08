package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"fuzzylab/fuzzy/core"
	"fuzzylab/gen"
)

func TestRoundTrip(t *testing.T) {
	items, err := gen.Uniform(gen.Params{N: 100, A: gen.Range{Lo: -50, Hi: 100}, W: gen.Range{Lo: 1, Hi: 20}, Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "data", "set.csv")
	if err := Save(path, items); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(items) {
		t.Fatalf("len = %d, want %d", len(got), len(items))
	}
	for i := range items {
		if got[i].Support() != items[i].Support() {
			t.Fatalf("item %d: %v, want %v", i, got[i].Support(), items[i].Support())
		}
	}
}

func TestLoadInvalid(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.csv")
	for _, body := range []string{
		"kind,a,b\nuniform,x,2\n",
		"kind,a,b\nuniform,3,2\n",
		"kind,a,b\ntriangle,1,2\n",
	} {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); !errors.Is(err, core.ErrInvalidParams) {
			t.Errorf("%q: err = %v", body, err)
		}
	}
}
