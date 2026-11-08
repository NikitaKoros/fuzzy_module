package number

import (
	"fuzzylab/fuzzy/core"
)

// Outline возвращает контур графика μ(x) в виде замкнутой ломаной:
// сначала левая граница срезов снизу вверх, затем правая сверху вниз.
// Используется GUI для отрисовки любой core.Number.
func Outline(n core.Number, steps int) []core.Point {
	levels := core.AlphaLevels(n.Height(), steps)
	left := make([]core.Point, 0, len(levels))
	right := make([]core.Point, 0, len(levels))
	for _, a := range levels {
		c := core.CutClamped(n, a)
		left = append(left, core.Point{X: c.Lo, Y: a})
		right = append(right, core.Point{X: c.Hi, Y: a})
	}
	out := make([]core.Point, 0, 2*len(levels))
	push := func(p core.Point) {
		if len(out) == 0 || out[len(out)-1] != p {
			out = append(out, p)
		}
	}
	for _, p := range left {
		push(p)
	}
	for i := len(right) - 1; i >= 0; i-- {
		push(right[i])
	}
	return out
}
