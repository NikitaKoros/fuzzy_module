// Package cluster — кластеризация по матрице расстояний.
package cluster

import (
	"fmt"
	"math"
	"slices"

	"fuzzylab/fuzzy/core"
)

// Result — метки кластеров 0..k−1 и индексы медоидов (по одному на кластер).
type Result struct {
	Labels  []int
	Medoids []int
}

func checkMatrix(dm [][]float64) error {
	if len(dm) == 0 {
		return core.ErrEmpty
	}
	for _, row := range dm {
		if len(row) != len(dm) {
			return fmt.Errorf("%w: матрица расстояний не квадратная", core.ErrInvalidParams)
		}
	}
	return nil
}

// PAM разбивает объекты на k кластеров методом k-медоидов (BUILD + SWAP).
func PAM(dm [][]float64, k int) (Result, error) {
	if err := checkMatrix(dm); err != nil {
		return Result{}, err
	}
	n := len(dm)
	if k < 1 || k > n {
		return Result{}, fmt.Errorf("%w: k=%d при %d объектах", core.ErrInvalidParams, k, n)
	}
	medoids := build(dm, k)
	cost := totalCost(dm, medoids)
	for improved := true; improved; {
		improved = false
		for m := range medoids {
			for h := 0; h < n; h++ {
				if slices.Contains(medoids, h) {
					continue
				}
				old := medoids[m]
				medoids[m] = h
				if c := totalCost(dm, medoids); c < cost-1e-12 {
					cost, improved = c, true
				} else {
					medoids[m] = old
				}
			}
		}
	}
	return Result{Labels: assign(dm, medoids), Medoids: medoids}, nil
}

// Agglomerative строит иерархию со средней связью и останавливается,
// когда расстояние между ближайшими кластерами достигает threshold.
func Agglomerative(dm [][]float64, threshold float64) (Result, error) {
	if err := checkMatrix(dm); err != nil {
		return Result{}, err
	}
	n := len(dm)
	d := make([][]float64, n)
	members := make([][]int, n)
	active := make([]bool, n)
	for i := range d {
		d[i] = append([]float64(nil), dm[i]...)
		members[i], active[i] = []int{i}, true
	}
	for {
		bi, bj, best := -1, -1, math.Inf(1)
		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				if active[i] && active[j] && d[i][j] < best {
					bi, bj, best = i, j, d[i][j]
				}
			}
		}
		if bi < 0 || best >= threshold {
			break
		}
		si, sj := float64(len(members[bi])), float64(len(members[bj]))
		for m := 0; m < n; m++ {
			if active[m] && m != bi && m != bj {
				d[bi][m] = (si*d[bi][m] + sj*d[bj][m]) / (si + sj)
				d[m][bi] = d[bi][m]
			}
		}
		members[bi] = append(members[bi], members[bj]...)
		active[bj] = false
	}
	res := Result{Labels: make([]int, n)}
	for i := range members {
		if !active[i] {
			continue
		}
		for _, m := range members[i] {
			res.Labels[m] = len(res.Medoids)
		}
		res.Medoids = append(res.Medoids, medoidOf(dm, members[i]))
	}
	return res, nil
}

// build жадно выбирает k начальных медоидов.
func build(dm [][]float64, k int) []int {
	n := len(dm)
	first, bestSum := 0, math.Inf(1)
	for i := 0; i < n; i++ {
		if s := sum(dm[i]); s < bestSum {
			first, bestSum = i, s
		}
	}
	medoids := []int{first}
	nearest := append([]float64(nil), dm[first]...)
	for len(medoids) < k {
		pick, bestGain := -1, -1.0
		for h := 0; h < n; h++ {
			if slices.Contains(medoids, h) {
				continue
			}
			gain := 0.0
			for j := 0; j < n; j++ {
				gain += math.Max(0, nearest[j]-dm[j][h])
			}
			if gain > bestGain {
				pick, bestGain = h, gain
			}
		}
		medoids = append(medoids, pick)
		for j := 0; j < n; j++ {
			nearest[j] = math.Min(nearest[j], dm[j][pick])
		}
	}
	return medoids
}

func totalCost(dm [][]float64, medoids []int) float64 {
	total := 0.0
	for j := range dm {
		best := math.Inf(1)
		for _, m := range medoids {
			best = math.Min(best, dm[j][m])
		}
		total += best
	}
	return total
}

func assign(dm [][]float64, medoids []int) []int {
	labels := make([]int, len(dm))
	for j := range dm {
		best := math.Inf(1)
		for c, m := range medoids {
			if dm[j][m] < best {
				labels[j], best = c, dm[j][m]
			}
		}
	}
	for c, m := range medoids {
		labels[m] = c
	}
	return labels
}

func medoidOf(dm [][]float64, members []int) int {
	best, bestSum := members[0], math.Inf(1)
	for _, i := range members {
		s := 0.0
		for _, j := range members {
			s += dm[i][j]
		}
		if s < bestSum {
			best, bestSum = i, s
		}
	}
	return best
}

func sum(xs []float64) float64 {
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s
}
