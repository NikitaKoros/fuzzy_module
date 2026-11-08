// Package number содержит реализации core.Number.
package number

import (
	"fmt"
	"math"

	"fuzzylab/fuzzy/core"
)

// Uniform хранит параметры a < b исходного равномерного распределения.
// Его функция принадлежности строится по функции распределения F, ограниченной
// исходным носителем [A, B]: μ(x)=F(x) на [A, B] и 0 вне него. Поэтому
// Uniform задаёт правостороннюю треугольную нечёткую величину (A, B, B),
// а не прямоугольную функцию принадлежности. Исходные Density и CDF доступны
// отдельно через соответствующие методы.
type Uniform struct{ A, B float64 }

var _ core.Number = Uniform{}

// NewUniform создаёт величину. Требуется конечные a < b.
func NewUniform(a, b float64) (Uniform, error) {
	if math.IsNaN(a) || math.IsNaN(b) || math.IsInf(a, 0) || math.IsInf(b, 0) || !(a < b) {
		return Uniform{}, fmt.Errorf("%w: равномерная величина требует конечных a < b, получено a=%g b=%g",
			core.ErrInvalidParams, a, b)
	}
	return Uniform{A: a, B: b}, nil
}

// CDF — функция распределения исходного равномерного распределения:
// 0 при x <= A, линейный рост на (A, B), 1 при x >= B.
func (u Uniform) CDF(x float64) float64 {
	if math.IsNaN(x) || x <= u.A {
		return 0
	}
	if x >= u.B {
		return 1
	}
	return (x - u.A) / (u.B - u.A)
}

// Membership — функция принадлежности, полученная ограничением CDF носителем:
// μ(x)=CDF(x) на [A,B] и 0 вне [A,B]. В частности, μ(A)=0, μ(B)=1.
func (u Uniform) Membership(x float64) float64 {
	if math.IsNaN(x) || x < u.A || x > u.B {
		return 0
	}
	return u.CDF(x)
}

// Density — исходная плотность равномерного распределения 1/(B-A) на [A,B].
func (u Uniform) Density(x float64) float64 {
	if x >= u.A && x <= u.B {
		return 1 / (u.B - u.A)
	}
	return 0
}

func (u Uniform) Support() core.Interval { return core.Interval{Lo: u.A, Hi: u.B} }

// AlphaCut возвращает срез правосторонней треугольной ФП:
// A_0=[A,B] (соглашение о носителе), A_α=[A+α(B-A), B] при 0<α<=1.
func (u Uniform) AlphaCut(alpha float64) (core.Interval, bool) {
	if math.IsNaN(alpha) || alpha < 0 || alpha > 1 {
		return core.Interval{}, false
	}
	if alpha == 0 {
		return u.Support(), true
	}
	if alpha == 1 {
		return core.Interval{Lo: u.B, Hi: u.B}, true
	}
	return core.Interval{Lo: u.A + alpha*(u.B-u.A), Hi: u.B}, true
}

func (u Uniform) Height() float64 { return 1 }
func (u Uniform) Kind() string    { return "uniform" }
func (u Uniform) Width() float64  { return u.B - u.A }
func (u Uniform) Mid() float64    { return (u.A + u.B) / 2 }

func (u Uniform) String() string { return fmt.Sprintf("U[%g; %g]", u.A, u.B) }
