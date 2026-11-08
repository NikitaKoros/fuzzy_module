package metrics

import "testing"

func TestAlphaCutDistance(t *testing.T) {
	a, b := u(t, 0, 10), u(t, 2, 6)
	// α=0: [0,10] и [2,6]; α=0.5: [5,10] и [4,6]
	for alpha, want := range map[float64]float64{0: 3, 0.5: 2.5} {
		if v := (AlphaCutDistance{Alpha: alpha}).Between(a, b); !near(v, want, 1e-12) {
			t.Errorf("α=%g: d = %g, want %g", alpha, v, want)
		}
	}
}
