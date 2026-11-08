// Package storage — сохранение и загрузка наборов равномерных величин в CSV.
package storage

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"fuzzylab/fuzzy/core"
	"fuzzylab/fuzzy/number"
)

var header = []string{"kind", "a", "b"}

// Save записывает набор в CSV (kind,a,b), создавая каталоги пути.
func Save(path string, items []core.Number) error {
	rows := [][]string{header}
	for i, it := range items {
		u, ok := it.(number.Uniform)
		if !ok {
			return fmt.Errorf("%w: элемент %d имеет тип %q, поддерживается только uniform",
				core.ErrInvalidParams, i, it.Kind())
		}
		rows = append(rows, []string{u.Kind(), format(u.A), format(u.B)})
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fileErr("создать каталог для", path, err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fileErr("создать", path, err)
	}
	if err := csv.NewWriter(f).WriteAll(rows); err != nil {
		f.Close()
		return fileErr("записать", path, err)
	}
	return f.Close()
}

func fileErr(action, path string, err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("файл не найден: %s", path)
	}
	return fmt.Errorf("не удалось %s файл %s: %w", action, path, err)
}

// Load читает набор, записанный Save.
func Load(path string) ([]core.Number, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fileErr("открыть", path, err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: файл не в формате CSV (kind,a,b)", core.ErrInvalidParams)
	}
	if len(rows) == 0 {
		return nil, core.ErrEmpty
	}
	items := make([]core.Number, 0, len(rows)-1)
	for i, row := range rows[1:] {
		if len(row) != len(header) || row[0] != "uniform" {
			return nil, fmt.Errorf("%w: строка %d (ожидалось uniform,a,b)", core.ErrInvalidParams, i+2)
		}
		a, errA := strconv.ParseFloat(row[1], 64)
		b, errB := strconv.ParseFloat(row[2], 64)
		if errA != nil || errB != nil {
			return nil, fmt.Errorf("%w: строка %d (a и b должны быть числами)", core.ErrInvalidParams, i+2)
		}
		u, err := number.NewUniform(a, b)
		if err != nil {
			return nil, fmt.Errorf("строка %d: %w", i+2, err)
		}
		items = append(items, u)
	}
	return items, nil
}

func format(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
