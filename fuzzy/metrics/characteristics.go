package metrics

import (
	"fuzzylab/fuzzy/core"
)

// Интегрирование выполняется численно методом трапеций по заданной α-сетке.

// Area — площадь под функцией принадлежности ∫ μ(x) dx = ∫ (Hi(α) - Lo(α)) dα.
func Area(n core.Number, cfg core.Config) float64 {
	cfg = cfg.WithDefaults()
	levels := core.AlphaLevels(n.Height(), cfg.AlphaLevels)
	return trapz(levels, func(a float64) float64 { return core.CutClamped(n, a).Width() })
}

// Centroid — центр тяжести ∫xμ dx / ∫μ dx = ∫ (Hi(α)² - Lo(α)²)/2 dα. Для величины нулевой ширины
// возвращает середину носителя.
func Centroid(n core.Number, cfg core.Config) float64 {
	cfg = cfg.WithDefaults()
	levels := core.AlphaLevels(n.Height(), cfg.AlphaLevels)
	area := trapz(levels, func(a float64) float64 { return core.CutClamped(n, a).Width() })
	if area <= 0 {
		return n.Support().Mid()
	}
	moment := trapz(levels, func(a float64) float64 {
		c := core.CutClamped(n, a)
		return (c.Hi*c.Hi - c.Lo*c.Lo) / 2
	})
	return moment / area
}

// Mean — возможностное среднее (Carlsson–Fuller):
//
//	E = ∫ α (Lo(α)+Hi(α)) dα / h².
func Mean(n core.Number, cfg core.Config) float64 {
	cfg = cfg.WithDefaults()
	h := n.Height()
	levels := core.AlphaLevels(h, cfg.AlphaLevels)
	return trapz(levels, func(a float64) float64 {
		c := core.CutClamped(n, a)
		return a * (c.Lo + c.Hi)
	}) / (h * h)
}

// Variance — возможностная дисперсия (Carlsson–Fuller):
//
//	Var = ∫ α (Hi(α)-Lo(α))² dα / (2h²).
func Variance(n core.Number, cfg core.Config) float64 {
	cfg = cfg.WithDefaults()
	h := n.Height()
	levels := core.AlphaLevels(h, cfg.AlphaLevels)
	return trapz(levels, func(a float64) float64 {
		w := core.CutClamped(n, a).Width()
		return a * w * w
	}) / (2 * h * h)
}

// Spread — ширина носителя (мера неопределённости величины).
func Spread(n core.Number) float64 { return n.Support().Width() }
