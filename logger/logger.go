// Package logger — журнал действий в формате JSON Lines.
package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fuzzylab/cluster"
	"fuzzylab/fuzzy/core"
)

// Logger дописывает по одной JSON-записи на действие.
type Logger struct {
	mu sync.Mutex
	f  *os.File
}

type entry struct {
	Time   string `json:"время"`
	Action string `json:"действие"`
	Data   any    `json:"данные"`
}

type span [2]float64

type operation struct {
	A      span   `json:"A"`
	B      *span  `json:"B,omitempty"`
	Result *span  `json:"результат,omitempty"`
	Error  string `json:"ошибка,omitempty"`
}

type clustering struct {
	Algo   string         `json:"алгоритм"`
	Params map[string]any `json:"параметры"`
	Sizes  []int          `json:"размеры_кластеров"`
}

// Open открывает файл журнала на дозапись, создавая каталоги пути.
func Open(path string) (*Logger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &Logger{f: f}, nil
}

func (l *Logger) Close() error { return l.f.Close() }

// Log записывает произвольное действие.
func (l *Logger) Log(action string, data any) error {
	line, err := json.Marshal(entry{Time: time.Now().Format(time.RFC3339), Action: action, Data: data})
	if err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err = l.f.Write(append(line, '\n'))
	return err
}

// Operation записывает операцию над величинами; b (унарная операция) и result (ошибка) могут быть nil.
func (l *Logger) Operation(op string, a, b, result core.Number, opErr error) error {
	d := operation{A: *toSpan(a), B: toSpan(b), Result: toSpan(result)}
	if opErr != nil {
		d.Error = opErr.Error()
	}
	return l.Log(op, d)
}

// Clustering записывает запуск кластеризации: алгоритм, параметры и размеры кластеров.
func (l *Logger) Clustering(algo string, params map[string]any, r cluster.Result) error {
	sizes := make([]int, len(r.Medoids))
	for _, c := range r.Labels {
		sizes[c]++
	}
	return l.Log("кластеризация", clustering{Algo: algo, Params: params, Sizes: sizes})
}

func toSpan(n core.Number) *span {
	if n == nil {
		return nil
	}
	s := n.Support()
	return &span{s.Lo, s.Hi}
}
