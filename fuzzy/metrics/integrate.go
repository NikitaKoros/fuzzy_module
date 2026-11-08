// Package metrics — числовые характеристики нечётких величин и расстояния
// между ними (метрики для кластеризации).
package metrics

// trapz — интеграл методом трапеций по узлам levels.
func trapz(levels []float64, f func(alpha float64) float64) float64 {
	if len(levels) < 2 {
		return 0
	}
	sum := 0.0
	prev := f(levels[0])
	for i := 1; i < len(levels); i++ {
		cur := f(levels[i])
		sum += (prev + cur) / 2 * (levels[i] - levels[i-1])
		prev = cur
	}
	return sum
}
