package core

// Number — нечёткая величина (выпуклая, с функцией принадлежности μ: R → [0,1]).
// Все остальные пакеты работают только с этим интерфейсом и не знают,
// какая именно функция принадлежности скрыта за ним.
type Number interface {
	// Membership возвращает μ(x).
	Membership(x float64) float64
	// Support — носитель (замыкание множества, где μ > 0).
	Support() Interval
	// AlphaCut возвращает α-срез {x : μ(x) >= α}.
	// Соглашение: AlphaCut(0) — это носитель; ok == false, если α < 0,
	// α > Height() или α некорректно.
	AlphaCut(alpha float64) (cut Interval, ok bool)
	// Height — sup μ. Для нормированных величин равна 1.
	Height() float64
	// Kind — короткое имя типа для логов и GUI ("uniform", "cuts", ...).
	Kind() string
}

// Point — точка графика (для отрисовки в GUI).
type Point struct{ X, Y float64 }

// AlphaLevels возвращает n+1 уровней 0, h/n, 2h/n, ..., h.
// Последний уровень равен h точно (без накопления погрешности).
func AlphaLevels(h float64, n int) []float64 {
	if n < 1 {
		n = 1
	}
	levels := make([]float64, n+1)
	for k := 0; k < n; k++ {
		levels[k] = h * float64(k) / float64(n)
	}
	levels[n] = h
	return levels
}

// CutClamped возвращает α-срез, при α > Height() берёт срез на уровне Height().
// Удобно в алгоритмах, где уровень α задан извне.
func CutClamped(n Number, alpha float64) Interval {
	if h := n.Height(); alpha > h {
		alpha = h
	}
	if c, ok := n.AlphaCut(alpha); ok {
		return c
	}
	return n.Support()
}
