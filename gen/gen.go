// Package gen — генератор наборов равномерных нечётких величин.
package gen

import (
	"fmt"
	"math/rand"

	"fuzzylab/fuzzy/core"
	"fuzzylab/fuzzy/number"
)

// Range — отрезок допустимых значений параметра.
type Range struct{ Lo, Hi float64 }

// Params — общие параметры генерации: число величин, диапазоны a и ширины w = b − a.
type Params struct {
	N    int
	A, W Range
	Seed int64
}

// CloudParams — параметры режима «облака»: K центров и шум как доля диапазона.
type CloudParams struct {
	Params
	K     int
	Noise float64
}

func (p Params) validate() error {
	switch {
	case p.N < 1:
		return fmt.Errorf("%w: число величин N=%d", core.ErrInvalidParams, p.N)
	case p.A.Lo > p.A.Hi || p.W.Lo > p.W.Hi:
		return fmt.Errorf("%w: пустой диапазон (начало больше конца)", core.ErrInvalidParams)
	case !(p.W.Lo > 0):
		return fmt.Errorf("%w: ширина должна быть положительной", core.ErrInvalidParams)
	}
	return nil
}

func (r Range) draw(rnd *rand.Rand) float64 { return r.Lo + rnd.Float64()*(r.Hi-r.Lo) }

func (r Range) jitter(rnd *rand.Rand, noise float64) float64 {
	return (2*rnd.Float64() - 1) * noise * (r.Hi - r.Lo)
}

// Uniform строит N величин с независимыми a и w, равномерными на заданных диапазонах.
func Uniform(p Params) ([]core.Number, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	rnd := rand.New(rand.NewSource(p.Seed))
	out := make([]core.Number, p.N)
	for i := range out {
		a := p.A.draw(rnd)
		u, err := number.NewUniform(a, a+p.W.draw(rnd))
		if err != nil {
			return nil, err
		}
		out[i] = u
	}
	return out, nil
}

// Clouds строит N величин вокруг K случайных центров и возвращает номер центра каждой.
func Clouds(p CloudParams) ([]core.Number, []int, error) {
	if err := p.validate(); err != nil {
		return nil, nil, err
	}
	if p.K < 1 || p.K > p.N || p.Noise < 0 {
		return nil, nil, fmt.Errorf("%w: число облаков K=%d (нужно от 1 до N), шум=%g (нужно не меньше 0)", core.ErrInvalidParams, p.K, p.Noise)
	}
	rnd := rand.New(rand.NewSource(p.Seed))
	ca, cw := make([]float64, p.K), make([]float64, p.K)
	for k := range ca {
		ca[k], cw[k] = p.A.draw(rnd), p.W.draw(rnd)
	}
	out := make([]core.Number, p.N)
	labels := make([]int, p.N)
	for i := range out {
		k := i % p.K
		a := ca[k] + p.A.jitter(rnd, p.Noise)
		w := max(cw[k]+p.W.jitter(rnd, p.Noise), p.W.Lo)
		u, err := number.NewUniform(a, a+w)
		if err != nil {
			return nil, nil, err
		}
		out[i], labels[i] = u, k
	}
	return out, labels, nil
}
