package web

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"fuzzylab/cluster"
	"fuzzylab/fuzzy/core"
)

const palette = 10

type viewModel struct {
	Alpha    float64
	Count    int
	Message  string
	Error    string
	Clusters []clusterRow
	Cuts     chart
	Plane    chart
}

type clusterRow struct {
	Class  int
	Num    int
	Size   int
	Medoid string
}

type tick struct {
	Pos   float64
	Label string
}

type bar struct {
	ID         string
	X, Y, W, H float64
	Class      int
	Tip        string
}

type dot struct {
	ID     string
	X, Y   float64
	Class  int
	Medoid bool
	Tip    string
}

type poly struct {
	Class  int
	Points string
	Name   string
}

// chart — готовая к выводу геометрия SVG-графика.
type chart struct {
	W, H                     int
	Left, Right, Top, Bottom float64
	XTicks, YTicks           []tick
	Bars                     []bar
	Dots                     []dot
	Polys                    []poly
	Diag                     bool
	XLabel, YLabel           string
}

// Base и End — нижняя и правая границы области графика.
func (c chart) Base() float64 { return float64(c.H) - c.Bottom }
func (c chart) End() float64  { return float64(c.W) - c.Right }

type scale struct{ lo, hi, p0, p1 float64 }

func (s scale) at(v float64) float64 { return s.p0 + (v-s.lo)/(s.hi-s.lo)*(s.p1-s.p0) }

func (s scale) ticks(n int) []tick {
	digits := 0
	if s.hi-s.lo < 10 {
		digits = 1
	}
	out := make([]tick, n+1)
	for i := range out {
		v := s.lo + (s.hi-s.lo)*float64(i)/float64(n)
		out[i] = tick{Pos: s.at(v), Label: strconv.FormatFloat(v, 'f', digits, 64)}
	}
	return out
}

func (s *Server) buildView(c ctl) viewModel {
	v := viewModel{Alpha: c.Alpha, Count: len(s.items)}
	res, err := s.cluster(c)
	if err != nil {
		v.Error = err.Error()
		return v
	}
	cuts := make([]core.Interval, len(s.items))
	dom := s.items[0].Support()
	for i, it := range s.items {
		cuts[i] = core.CutClamped(it, c.Alpha)
		dom = dom.Hull(it.Support())
	}
	pad := (dom.Hi - dom.Lo) * 0.03
	dom = core.Interval{Lo: dom.Lo - pad, Hi: dom.Hi + pad}

	v.Clusters = clusterRows(s.items, res)
	v.Cuts = cutsChart(s.items, cuts, res, dom)
	v.Plane = planeChart(s.items, cuts, res, dom)
	return v
}

func clusterRows(items []core.Number, res cluster.Result) []clusterRow {
	rows := make([]clusterRow, len(res.Medoids))
	for c, m := range res.Medoids {
		s := items[m].Support()
		rows[c] = clusterRow{Class: c % palette, Num: c + 1, Medoid: fmt.Sprintf("[%.2f; %.2f]", s.Lo, s.Hi)}
	}
	for _, l := range res.Labels {
		rows[l].Size++
	}
	return rows
}

func tip(it core.Number, cut core.Interval, label int) string {
	s := it.Support()
	return fmt.Sprintf("[%.2f; %.2f] → срез [%.2f; %.2f], кластер %d", s.Lo, s.Hi, cut.Lo, cut.Hi, label+1)
}

// cutsChart рисует срез каждой величины полосой, сгруппированной по кластерам.
func cutsChart(items []core.Number, cuts []core.Interval, res cluster.Result, dom core.Interval) chart {
	ch := chart{W: 640, H: 380, Left: 16, Right: 16, Top: 12, Bottom: 34, XLabel: "x"}
	x := scale{dom.Lo, dom.Hi, ch.Left, float64(ch.W) - ch.Right}
	order := make([]int, len(items))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(p, q int) bool {
		i, j := order[p], order[q]
		if res.Labels[i] != res.Labels[j] {
			return res.Labels[i] < res.Labels[j]
		}
		return cuts[i].Lo < cuts[j].Lo
	})
	plotH := float64(ch.H) - ch.Top - ch.Bottom
	row := plotH / float64(len(items))
	for pos, i := range order {
		l := res.Labels[i]
		ch.Bars = append(ch.Bars, bar{
			ID: "b" + strconv.Itoa(i), X: x.at(cuts[i].Lo), W: math.Max(x.at(cuts[i].Hi)-x.at(cuts[i].Lo), 1.5),
			Y: ch.Top + float64(pos)*row, H: math.Max(row*0.8, 0.6), Class: l % palette, Tip: tip(items[i], cuts[i], l),
		})
	}
	ch.XTicks = x.ticks(6)
	return ch
}

// planeChart рисует каждую величину точкой (Lo(α), Hi(α)); медоиды выделены.
func planeChart(items []core.Number, cuts []core.Interval, res cluster.Result, dom core.Interval) chart {
	ch := chart{W: 420, H: 380, Left: 44, Right: 16, Top: 12, Bottom: 34, Diag: true, XLabel: "нижняя граница среза", YLabel: "верхняя"}
	x := scale{dom.Lo, dom.Hi, ch.Left, float64(ch.W) - ch.Right}
	y := scale{dom.Lo, dom.Hi, float64(ch.H) - ch.Bottom, ch.Top}
	isMedoid := map[int]bool{}
	for _, m := range res.Medoids {
		isMedoid[m] = true
	}
	for i, it := range items {
		l := res.Labels[i]
		ch.Dots = append(ch.Dots, dot{
			ID: "d" + strconv.Itoa(i), X: x.at(cuts[i].Lo), Y: y.at(cuts[i].Hi),
			Class: l % palette, Medoid: isMedoid[i], Tip: tip(it, cuts[i], l),
		})
	}
	ch.XTicks, ch.YTicks = x.ticks(5), y.ticks(5)
	return ch
}
