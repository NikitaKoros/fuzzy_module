// Package compare — сравнение и ранжирование нечётких величин.
//
// «Больше/меньше» для нечётких величин неоднозначно, поэтому есть степени
// возможности и необходимости и набор взаимозаменяемых стратегий.
package compare

import (
	"math"
	"sort"

	"fuzzylab/fuzzy/core"
	"fuzzylab/fuzzy/metrics"
)

const bisectIter = 60

// PossibilityGE — степень возможности Pos(a ≥ b) = sup_{x≥y} min(μa(x), μb(y)).
// Считается через α-срезы: наибольший α, при котором max(A_α) ≥ min(B_α).
func PossibilityGE(a, b core.Number) float64 {
	h := math.Min(a.Height(), b.Height())
	holds := func(alpha float64) bool {
		ca, okA := a.AlphaCut(alpha)
		cb, okB := b.AlphaCut(alpha)
		return okA && okB && ca.Hi >= cb.Lo
	}
	if !(h > 0) || !holds(0) {
		return 0
	}
	if holds(h) {
		return h
	}
	lo, hi := 0.0, h // holds(lo) = true, holds(hi) = false; условие монотонно по α
	for i := 0; i < bisectIter; i++ {
		mid := (lo + hi) / 2
		if holds(mid) {
			lo = mid
		} else {
			hi = mid
		}
	}
	return lo
}

// NecessityGE — степень необходимости Nec(a ≥ b) = 1 - Pos(b ≥ a)
// (для нормальных величин, h = 1).
func NecessityGE(a, b core.Number) float64 {
	return 1 - PossibilityGE(b, a)
}

// Result — итог сравнения.
type Result int

const (
	Less    Result = -1
	Equal   Result = 0
	Greater Result = 1
)

func (r Result) String() string {
	switch r {
	case Less:
		return "<"
	case Greater:
		return ">"
	default:
		return "="
	}
}

// Strategy — взаимозаменяемый способ сравнения.
type Strategy interface {
	Compare(a, b core.Number) Result
	Name() string
}

func sign(d float64) Result {
	switch {
	case d > 0:
		return Greater
	case d < 0:
		return Less
	default:
		return Equal
	}
}

// CentroidStrategy сравнивает центры тяжести.
type CentroidStrategy struct {
	Cfg core.Config
}

func (s CentroidStrategy) Name() string { return "centroid" }

func (s CentroidStrategy) Compare(a, b core.Number) Result {
	ca := metrics.Centroid(a, s.Cfg)
	cb := metrics.Centroid(b, s.Cfg)
	return sign(ca - cb)
}

// PossibilityStrategy сравнивает по разности Pos(a≥b) - Pos(b≥a).
// Сильно пересекающиеся величины считаются равными.
type PossibilityStrategy struct{}

func (s PossibilityStrategy) Name() string { return "possibility" }

func (s PossibilityStrategy) Compare(a, b core.Number) Result {
	return sign(PossibilityGE(a, b) - PossibilityGE(b, a))
}

// IntervalStrategy сравнивает середины α-срезов на уровне Alpha.
type IntervalStrategy struct {
	Alpha float64
}

func (s IntervalStrategy) Name() string { return "alpha-cut" }

func (s IntervalStrategy) Compare(a, b core.Number) Result {
	ca := core.CutClamped(a, s.Alpha)
	cb := core.CutClamped(b, s.Alpha)
	return sign(ca.Mid() - cb.Mid())
}

// Sort упорядочивает по возрастанию (устойчиво).
func Sort(items []core.Number, s Strategy) {
	sort.SliceStable(items, func(i, j int) bool { return s.Compare(items[i], items[j]) == Less })
}

// Max возвращает наибольшую величину (первую из равных).
func Max(items []core.Number, s Strategy) (core.Number, error) {
	if len(items) == 0 {
		return nil, core.ErrEmpty
	}
	best := items[0]
	for _, it := range items[1:] {
		if s.Compare(it, best) == Greater {
			best = it
		}
	}
	return best, nil
}

// Min возвращает наименьшую величину (первую из равных).
func Min(items []core.Number, s Strategy) (core.Number, error) {
	if len(items) == 0 {
		return nil, core.ErrEmpty
	}
	best := items[0]
	for _, it := range items[1:] {
		if s.Compare(it, best) == Less {
			best = it
		}
	}
	return best, nil
}
