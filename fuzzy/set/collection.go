// Package set — множество нечётких величин.
package set

import (
	"fmt"

	"fuzzylab/fuzzy/core"
)

// Collection — упорядоченный набор нечётких величин любых типов.
type Collection struct {
	items []core.Number
}

func New(items ...core.Number) *Collection {
	return &Collection{items: append([]core.Number(nil), items...)}
}

func (c *Collection) Add(items ...core.Number) { c.items = append(c.items, items...) }
func (c *Collection) Len() int                 { return len(c.items) }

func (c *Collection) Get(i int) (core.Number, error) {
	if i < 0 || i >= len(c.items) {
		return nil, fmt.Errorf("%w: %d (размер %d)", core.ErrIndex, i, len(c.items))
	}
	return c.items[i], nil
}

func (c *Collection) Remove(i int) error {
	if i < 0 || i >= len(c.items) {
		return fmt.Errorf("%w: %d (размер %d)", core.ErrIndex, i, len(c.items))
	}
	c.items = append(c.items[:i], c.items[i+1:]...)
	return nil
}

// Slice возвращает копию содержимого (для передачи в кластеризацию, метрики).
func (c *Collection) Slice() []core.Number { return append([]core.Number(nil), c.items...) }

func (c *Collection) Clone() *Collection { return &Collection{items: c.Slice()} }

func (c *Collection) Filter(pred func(core.Number) bool) *Collection {
	out := &Collection{}
	for _, it := range c.items {
		if pred(it) {
			out.items = append(out.items, it)
		}
	}
	return out
}

// Bounds — наименьший отрезок, покрывающий носители всех величин
// (границы для осей графика). Пустая коллекция — core.ErrEmpty.
func (c *Collection) Bounds() (core.Interval, error) {
	if len(c.items) == 0 {
		return core.Interval{}, core.ErrEmpty
	}
	b := c.items[0].Support()
	for _, it := range c.items[1:] {
		b = b.Hull(it.Support())
	}
	return b, nil
}
