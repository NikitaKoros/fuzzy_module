// Package ops — арифметика над нечёткими величинами.
//
// Метод: принцип обобщения через α-срезы. Для каждого уровня α берутся срезы
// операндов, над ними выполняется интервальная операция, результаты
// собираются в number.Cuts. Для выпуклых нечётких величин это эквивалентно
// принципу обобщения Заде.
package ops

import (
	"fmt"
	"math"

	"fuzzylab/fuzzy/core"
	"fuzzylab/fuzzy/number"
)

// Operation — бинарная операция над срезами.
type Operation func(a, b core.Interval) (core.Interval, error)

// UnaryOperation — унарная операция над срезом.
type UnaryOperation func(a core.Interval) core.Interval

// Готовые операции. Новая операция (степень, min, max...) — это ещё одна
// функция такого вида, ядро Apply менять не нужно.
func AddOp(a, b core.Interval) (core.Interval, error) { return a.Add(b), nil }
func SubOp(a, b core.Interval) (core.Interval, error) { return a.Sub(b), nil }
func MulOp(a, b core.Interval) (core.Interval, error) { return a.Mul(b), nil }
func DivOp(a, b core.Interval) (core.Interval, error) { return a.Div(b) }

// MinOp и MaxOp — поэлементные min/max нечётких величин.
func MinOp(a, b core.Interval) (core.Interval, error) {
	return core.Interval{Lo: math.Min(a.Lo, b.Lo), Hi: math.Min(a.Hi, b.Hi)}, nil
}
func MaxOp(a, b core.Interval) (core.Interval, error) {
	return core.Interval{Lo: math.Max(a.Lo, b.Lo), Hi: math.Max(a.Hi, b.Hi)}, nil
}

// Apply — ядро: применяет op на уровнях 0, h/N, ..., h, где
// h = min(высот операндов), N = cfg.AlphaLevels.
// Деление на величину, носитель которой содержит 0, даёт core.ErrDivByZero.
func Apply(a, b core.Number, op Operation, cfg core.Config) (number.Cuts, error) {
	cfg = cfg.WithDefaults()
	h := math.Min(a.Height(), b.Height())
	if !(h > 0) {
		return number.Cuts{}, core.ErrEmpty
	}
	levels := core.AlphaLevels(h, cfg.AlphaLevels)
	cuts := make([]core.Interval, len(levels))
	for i, alpha := range levels {
		ca, okA := a.AlphaCut(alpha)
		cb, okB := b.AlphaCut(alpha)
		if !okA || !okB {
			return number.Cuts{}, core.ErrEmpty
		}
		r, err := op(ca, cb)
		if err != nil {
			return number.Cuts{}, fmt.Errorf("α=%.4g: %w", alpha, err)
		}
		cuts[i] = r
	}
	return number.NewCuts(levels, cuts)
}

// ApplyUnary — то же для одного операнда.
func ApplyUnary(a core.Number, f UnaryOperation, cfg core.Config) (number.Cuts, error) {
	cfg = cfg.WithDefaults()
	h := a.Height()
	if !(h > 0) {
		return number.Cuts{}, core.ErrEmpty
	}
	levels := core.AlphaLevels(h, cfg.AlphaLevels)
	cuts := make([]core.Interval, len(levels))
	for i, alpha := range levels {
		c, ok := a.AlphaCut(alpha)
		if !ok {
			return number.Cuts{}, core.ErrEmpty
		}
		cuts[i] = f(c)
	}
	return number.NewCuts(levels, cuts)
}

func Add(a, b core.Number, cfg core.Config) (number.Cuts, error) { return Apply(a, b, AddOp, cfg) }
func Sub(a, b core.Number, cfg core.Config) (number.Cuts, error) { return Apply(a, b, SubOp, cfg) }
func Mul(a, b core.Number, cfg core.Config) (number.Cuts, error) { return Apply(a, b, MulOp, cfg) }
func Div(a, b core.Number, cfg core.Config) (number.Cuts, error) { return Apply(a, b, DivOp, cfg) }

// Neg — унарный минус.
func Neg(a core.Number, cfg core.Config) (number.Cuts, error) {
	return ApplyUnary(a, func(c core.Interval) core.Interval { return c.Neg() }, cfg)
}

// Scale — умножение на чёткое число k (может быть отрицательным или 0).
func Scale(a core.Number, k float64, cfg core.Config) (number.Cuts, error) {
	return ApplyUnary(a, func(c core.Interval) core.Interval { return c.Scale(k) }, cfg)
}

// Calculator — оболочка с зафиксированной конфигурацией для цепочек вычислений.
type Calculator struct{ Cfg core.Config }

func (c Calculator) Add(a, b core.Number) (number.Cuts, error) { return Add(a, b, c.Cfg) }
func (c Calculator) Sub(a, b core.Number) (number.Cuts, error) { return Sub(a, b, c.Cfg) }
func (c Calculator) Mul(a, b core.Number) (number.Cuts, error) { return Mul(a, b, c.Cfg) }
func (c Calculator) Div(a, b core.Number) (number.Cuts, error) { return Div(a, b, c.Cfg) }
func (c Calculator) Neg(a core.Number) (number.Cuts, error)    { return Neg(a, c.Cfg) }
func (c Calculator) Scale(a core.Number, k float64) (number.Cuts, error) {
	return Scale(a, k, c.Cfg)
}
func (c Calculator) Apply(a, b core.Number, op Operation) (number.Cuts, error) {
	return Apply(a, b, op, c.Cfg)
}
