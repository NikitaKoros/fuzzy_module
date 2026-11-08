package metrics

import (
	"math"

	"fuzzylab/fuzzy/core"
)

// Distance — расстояние между двумя нечёткими величинами.
// Кластеризация принимает этот интерфейс и не зависит от конкретной метрики.
type Distance interface {
	Between(a, b core.Number) float64
	Name() string
}

// pairIntegral усредняет f по α ∈ [0, h], h = min(высот).
func pairIntegral(a, b core.Number, cfg core.Config, f func(ca, cb core.Interval) float64) float64 {
	cfg = cfg.WithDefaults()
	h := math.Min(a.Height(), b.Height())
	if !(h > 0) {
		return math.NaN()
	}
	levels := core.AlphaLevels(h, cfg.AlphaLevels)
	return trapz(levels, func(alpha float64) float64 {
		return f(core.CutClamped(a, alpha), core.CutClamped(b, alpha))
	}) / h
}

// CutsL1Distance — среднее по α смещение границ срезов:
//
//	d = (1/h) ∫ (|ΔLo(α)| + |ΔHi(α)|)/2 dα.
type CutsL1Distance struct{ Cfg core.Config }

func (d CutsL1Distance) Name() string { return "cuts-L1" }
func (d CutsL1Distance) Between(a, b core.Number) float64 {
	return pairIntegral(a, b, d.Cfg, func(x, y core.Interval) float64 {
		return (math.Abs(x.Lo-y.Lo) + math.Abs(x.Hi-y.Hi)) / 2
	})
}

// CutsL2Distance — среднеквадратичное смещение границ срезов:
//
//	d = sqrt( (1/h) ∫ (ΔLo(α)² + ΔHi(α)²)/2 dα ).
type CutsL2Distance struct{ Cfg core.Config }

func (d CutsL2Distance) Name() string { return "cuts-L2" }
func (d CutsL2Distance) Between(a, b core.Number) float64 {
	v := pairIntegral(a, b, d.Cfg, func(x, y core.Interval) float64 {
		dl, dh := x.Lo-y.Lo, x.Hi-y.Hi
		return (dl*dl + dh*dh) / 2
	})
	return math.Sqrt(v)
}

// HausdorffDistance — max по α расстояния Хаусдорфа между срезами:
//
//	d = max_α max(|ΔLo|, |ΔHi|).
type HausdorffDistance struct{ Cfg core.Config }

func (d HausdorffDistance) Name() string { return "hausdorff" }
func (d HausdorffDistance) Between(a, b core.Number) float64 {
	cfg := d.Cfg.WithDefaults()
	h := math.Min(a.Height(), b.Height())
	best := 0.0
	for _, alpha := range core.AlphaLevels(h, cfg.AlphaLevels) {
		x, y := core.CutClamped(a, alpha), core.CutClamped(b, alpha)
		best = math.Max(best, math.Max(math.Abs(x.Lo-y.Lo), math.Abs(x.Hi-y.Hi)))
	}
	return best
}

// AlphaCutDistance — смещение границ срезов на одном уровне α:
//
//	d = (|ΔLo(α)| + |ΔHi(α)|)/2.
type AlphaCutDistance struct{ Alpha float64 }

func (d AlphaCutDistance) Name() string { return "cut-L1" }
func (d AlphaCutDistance) Between(a, b core.Number) float64 {
	x, y := core.CutClamped(a, d.Alpha), core.CutClamped(b, d.Alpha)
	return (math.Abs(x.Lo-y.Lo) + math.Abs(x.Hi-y.Hi)) / 2
}

// CentroidDistance — |c_a - c_b| по центрам тяжести.
type CentroidDistance struct{ Cfg core.Config }

func (d CentroidDistance) Name() string { return "centroid" }
func (d CentroidDistance) Between(a, b core.Number) float64 {
	return math.Abs(Centroid(a, d.Cfg) - Centroid(b, d.Cfg))
}

// HammingDistance — ∫|μa(x) - μb(x)| dx на сетке из cfg.GridPoints точек
// (правило средних прямоугольников). У величин с разрывами или вертикальными
// границами погрешность зависит от шага сетки.
type HammingDistance struct{ Cfg core.Config }

func (d HammingDistance) Name() string { return "hamming" }
func (d HammingDistance) Between(a, b core.Number) float64 {
	cfg := d.Cfg.WithDefaults()
	hull := a.Support().Hull(b.Support())
	w := hull.Width()
	if w <= 0 {
		return 0
	}
	step := w / float64(cfg.GridPoints)
	sum := 0.0
	for i := 0; i < cfg.GridPoints; i++ {
		x := hull.Lo + (float64(i)+0.5)*step
		sum += math.Abs(a.Membership(x) - b.Membership(x))
	}
	return sum * step
}

// DistanceMatrix — симметричная матрица попарных расстояний (диагональ 0).
func DistanceMatrix(items []core.Number, d Distance) [][]float64 {
	n := len(items)
	m := make([][]float64, n)
	for i := range m {
		m[i] = make([]float64, n)
	}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			v := d.Between(items[i], items[j])
			m[i][j], m[j][i] = v, v
		}
	}
	return m
}
