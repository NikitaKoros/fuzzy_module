package number

import (
	"fmt"
	"math"
	"sort"

	"fuzzylab/fuzzy/core"
)

// Cuts — универсальная нечёткая величина, заданная набором вложенных α-срезов:
// на уровне levels[i] срез равен cuts[i]. Между уровнями границы срезов
// интерполируются линейно.
//
// Представление через срезы (а не через сетку по x) точно хранит величины
// с кусочно-линейными границами α-срезов (например, треугольные величины).
// Для нелинейных границ, возникающих у некоторых операций, используется
// кусочно-линейная аппроксимация с точностью, зависящей от числа уровней.
type Cuts struct {
	levels []float64       // строго возрастают, levels[last] = высота
	cuts   []core.Interval // вложены: cuts[i+1] ⊆ cuts[i]
}

var _ core.Number = Cuts{}

// NewCuts проверяет и копирует данные. Требования: уровни в [0,1] строго
// возрастают, последний > 0, срезы корректны и вложены друг в друга.
// Если первый уровень не 0, носителем считается первый срез.
func NewCuts(levels []float64, cuts []core.Interval) (Cuts, error) {
	n := len(levels)
	if n == 0 || n != len(cuts) {
		return Cuts{}, fmt.Errorf("%w: уровни и срезы должны быть непустыми и одной длины", core.ErrInvalidParams)
	}
	ls := append([]float64(nil), levels...)
	cs := append([]core.Interval(nil), cuts...)
	for i := range ls {
		if math.IsNaN(ls[i]) || ls[i] < 0 || ls[i] > 1+1e-12 {
			return Cuts{}, fmt.Errorf("%w: уровень %g вне [0,1]", core.ErrInvalidParams, ls[i])
		}
		if i > 0 && ls[i] <= ls[i-1] {
			return Cuts{}, fmt.Errorf("%w: уровни должны строго возрастать", core.ErrInvalidParams)
		}
		if !cs[i].Valid() {
			return Cuts{}, fmt.Errorf("%w: некорректный срез %v на уровне %g", core.ErrInvalidParams, cs[i], ls[i])
		}
		if i > 0 {
			p := cs[i-1]
			tol := 1e-9 * (1 + math.Max(math.Abs(p.Lo), math.Abs(p.Hi)))
			if cs[i].Lo < p.Lo-tol || cs[i].Hi > p.Hi+tol {
				return Cuts{}, fmt.Errorf("%w: срезы не вложены на уровне %g", core.ErrInvalidParams, ls[i])
			}
			// убираем погрешность округления в пределах допуска
			cs[i].Lo = math.Max(cs[i].Lo, p.Lo)
			cs[i].Hi = math.Min(cs[i].Hi, p.Hi)
			if cs[i].Lo > cs[i].Hi {
				m := (cs[i].Lo + cs[i].Hi) / 2
				cs[i] = core.Interval{Lo: m, Hi: m}
			}
		}
	}
	if ls[n-1] <= 0 {
		return Cuts{}, fmt.Errorf("%w: высота должна быть положительной", core.ErrInvalidParams)
	}
	return Cuts{levels: ls, cuts: cs}, nil
}

// FromNumber снимает с любой core.Number срезы на n+1 равноотстоящих уровнях
// от 0 до Height().
func FromNumber(num core.Number, n int) (Cuts, error) {
	h := num.Height()
	if !(h > 0) {
		return Cuts{}, core.ErrEmpty
	}
	levels := core.AlphaLevels(h, n)
	cuts := make([]core.Interval, len(levels))
	for i, a := range levels {
		c, ok := num.AlphaCut(a)
		if !ok {
			return Cuts{}, core.ErrEmpty
		}
		cuts[i] = c
	}
	return NewCuts(levels, cuts)
}

func (c Cuts) Len() int { return len(c.levels) }

// Levels возвращает копию уровней.
func (c Cuts) Levels() []float64 { return append([]float64(nil), c.levels...) }

// Intervals возвращает копию срезов.
func (c Cuts) Intervals() []core.Interval { return append([]core.Interval(nil), c.cuts...) }

func (c Cuts) Height() float64        { return c.levels[len(c.levels)-1] }
func (c Cuts) Support() core.Interval { return c.cuts[0] }
func (c Cuts) Kind() string           { return "cuts" }

// AlphaCut — линейная интерполяция границ между соседними уровнями.
func (c Cuts) AlphaCut(alpha float64) (core.Interval, bool) {
	if math.IsNaN(alpha) || alpha < 0 || alpha > c.Height()+1e-12 {
		return core.Interval{}, false
	}
	if alpha <= c.levels[0] {
		return c.cuts[0], true
	}
	n := len(c.levels)
	i := sort.SearchFloat64s(c.levels, alpha) // первый levels[i] >= alpha
	if i >= n {
		return c.cuts[n-1], true
	}
	if c.levels[i] == alpha {
		return c.cuts[i], true
	}
	t := (alpha - c.levels[i-1]) / (c.levels[i] - c.levels[i-1])
	lo := c.cuts[i-1].Lo + t*(c.cuts[i].Lo-c.cuts[i-1].Lo)
	hi := c.cuts[i-1].Hi + t*(c.cuts[i].Hi-c.cuts[i-1].Hi)
	return core.Interval{Lo: lo, Hi: hi}, true
}

// Membership восстанавливает μ(x) по срезам: левая и правая ветви
// обращаются отдельно, итог — минимум из них.
func (c Cuts) Membership(x float64) float64 {
	if math.IsNaN(x) || !c.cuts[0].Contains(x) {
		return 0
	}
	n := len(c.levels)

	// левая ветвь: Lo неубывает по α; ищем последний i с Lo[i] <= x
	i := sort.Search(n, func(k int) bool { return c.cuts[k].Lo > x }) - 1
	muL := c.levels[i]
	if i < n-1 {
		d := c.cuts[i+1].Lo - c.cuts[i].Lo
		muL += (x - c.cuts[i].Lo) / d * (c.levels[i+1] - c.levels[i])
	}

	// правая ветвь: Hi невозрастает по α; ищем последний j с Hi[j] >= x
	j := sort.Search(n, func(k int) bool { return c.cuts[k].Hi < x }) - 1
	muR := c.levels[j]
	if j < n-1 {
		d := c.cuts[j].Hi - c.cuts[j+1].Hi
		muR += (c.cuts[j].Hi - x) / d * (c.levels[j+1] - c.levels[j])
	}
	return math.Min(muL, muR)
}

// Normalized возвращает копию с высотой 1 (уровни делятся на высоту).
func (c Cuts) Normalized() Cuts {
	h := c.Height()
	ls := make([]float64, len(c.levels))
	for i, l := range c.levels {
		ls[i] = l / h
	}
	ls[len(ls)-1] = 1
	return Cuts{levels: ls, cuts: append([]core.Interval(nil), c.cuts...)}
}

func (c Cuts) String() string {
	s := c.Support()
	return fmt.Sprintf("Cuts{support=[%.4g; %.4g], height=%.3g, levels=%d}", s.Lo, s.Hi, c.Height(), len(c.levels))
}
