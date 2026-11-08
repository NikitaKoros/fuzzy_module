package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"fuzzylab/cluster"
	"fuzzylab/fuzzy/core"
	"fuzzylab/fuzzy/number"
)

func TestLogger(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "log.json")
	l, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := number.NewUniform(1, 3)
	b, _ := number.NewUniform(2, 5)
	sum, _ := number.NewUniform(3, 8)
	if err := l.Operation("сложение", a, b, sum, nil); err != nil {
		t.Fatal(err)
	}
	if err := l.Operation("деление", a, b, nil, core.ErrDivByZero); err != nil {
		t.Fatal(err)
	}
	res := cluster.Result{Labels: []int{0, 0, 1}, Medoids: []int{0, 2}}
	if err := l.Clustering("pam", map[string]any{"k": 2}, res); err != nil {
		t.Fatal(err)
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var actions []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var e struct {
			Time   string         `json:"время"`
			Action string         `json:"действие"`
			Data   map[string]any `json:"данные"`
		}
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			t.Fatalf("invalid line %q: %v", sc.Text(), err)
		}
		actions = append(actions, e.Action)
		if e.Action == "деление" && e.Data["ошибка"] != core.ErrDivByZero.Error() {
			t.Errorf("ошибка деления не записана: %v", e.Data)
		}
		if e.Action == "кластеризация" && len(e.Data["размеры_кластеров"].([]any)) != 2 {
			t.Errorf("размеры = %v", e.Data["размеры_кластеров"])
		}
	}
	if len(actions) != 3 || actions[0] != "сложение" || actions[2] != "кластеризация" {
		t.Errorf("actions = %v", actions)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}
