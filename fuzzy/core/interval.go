package core

import "math"

// Interval — замкнутый отрезок [Lo, Hi]. Это α-срез нечёткой величины
// и одновременно элемент интервальной арифметики (Мура).
type Interval struct{ Lo, Hi float64 }

// Valid сообщает, что границы конечны и Lo <= Hi.
func (i Interval) Valid() bool {
	return !math.IsNaN(i.Lo) && !math.IsNaN(i.Hi) &&
		!math.IsInf(i.Lo, 0) && !math.IsInf(i.Hi, 0) && i.Lo <= i.Hi
}

func (i Interval) Width() float64          { return i.Hi - i.Lo }
func (i Interval) Mid() float64            { return (i.Lo + i.Hi) / 2 }
func (i Interval) Contains(x float64) bool { return x >= i.Lo && x <= i.Hi }

// Hull — наименьший отрезок, содержащий оба.
func (i Interval) Hull(o Interval) Interval {
	return Interval{math.Min(i.Lo, o.Lo), math.Max(i.Hi, o.Hi)}
}

// Add: [a,b] + [c,d] = [a+c, b+d].
func (i Interval) Add(o Interval) Interval {
	return Interval{i.Lo + o.Lo, i.Hi + o.Hi}
}

// Sub: [a,b] - [c,d] = [a-d, b-c].
func (i Interval) Sub(o Interval) Interval {
	return Interval{i.Lo - o.Hi, i.Hi - o.Lo}
}

// Mul: границы — min и max из четырёх попарных произведений.
func (i Interval) Mul(o Interval) Interval {
	p1, p2, p3, p4 := i.Lo*o.Lo, i.Lo*o.Hi, i.Hi*o.Lo, i.Hi*o.Hi
	return Interval{
		math.Min(math.Min(p1, p2), math.Min(p3, p4)),
		math.Max(math.Max(p1, p2), math.Max(p3, p4)),
	}
}

// Div: [a,b] / [c,d] = [a,b] * [1/d, 1/c]. Если 0 ∈ [c,d] — ErrDivByZero.
func (i Interval) Div(o Interval) (Interval, error) {
	if o.Lo <= 0 && o.Hi >= 0 {
		return Interval{}, ErrDivByZero
	}
	return i.Mul(Interval{1 / o.Hi, 1 / o.Lo}), nil
}

// Neg: -[a,b] = [-b, -a].
func (i Interval) Neg() Interval { return Interval{-i.Hi, -i.Lo} }

// Scale умножает отрезок на число (в том числе отрицательное).
func (i Interval) Scale(k float64) Interval {
	a, b := i.Lo*k, i.Hi*k
	return Interval{math.Min(a, b), math.Max(a, b)}
}
