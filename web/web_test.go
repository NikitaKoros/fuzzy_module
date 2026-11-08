package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fuzzylab/logger"
)

func newHandler(t *testing.T) (http.Handler, string) {
	t.Helper()
	dir := t.TempDir()
	l, err := logger.Open(filepath.Join(dir, "log.json"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	s, err := New(dir, l)
	if err != nil {
		t.Fatal(err)
	}
	return s.Handler(), dir
}

func do(h http.Handler, method, target string, form url.Values) string {
	var body *strings.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	} else {
		body = strings.NewReader("")
	}
	req := httptest.NewRequest(method, target, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Body.String()
}

func TestIndexAndView(t *testing.T) {
	h, _ := newHandler(t)
	if page := do(h, "GET", "/", nil); !strings.Contains(page, "<svg") || !strings.Contains(page, `name="alpha"`) {
		t.Error("index must contain charts and α control")
	}
	for _, q := range []string{"/view?alpha=0.5&algo=pam&k=3", "/view?alpha=1&algo=agglo&threshold=5", "/view?alpha=zzz&k=-4"} {
		out := do(h, "GET", q, nil)
		if !strings.Contains(out, "<circle") || strings.Contains(out, `class="error"`) {
			t.Errorf("%s: bad view: %.200s", q, out)
		}
	}
}

func TestGenerateSaveLoad(t *testing.T) {
	h, dir := newHandler(t)
	gen := url.Values{"mode": {"uniform"}, "n": {"120"}, "groups": {"3"}, "alo": {"0"}, "ahi": {"50"},
		"wlo": {"1"}, "whi": {"5"}, "noise": {"0.05"}, "seed": {"7"}}
	if out := do(h, "POST", "/generate", gen); !strings.Contains(out, "Величин: <b>120</b>") {
		t.Errorf("generate: %.200s", out)
	}
	gen.Set("n", "abc")
	if out := do(h, "POST", "/generate", gen); !strings.Contains(out, `class="error"`) {
		t.Error("invalid n must give an error")
	}
	if out := do(h, "POST", "/save", url.Values{"file": {"../x"}}); !strings.Contains(out, "Сохранено") {
		t.Errorf("save: %.200s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "x.csv")); err != nil {
		t.Error("file must be saved inside data dir:", err)
	}
	if out := do(h, "POST", "/load", url.Values{"file": {"x"}}); !strings.Contains(out, `class="error"`) {
		t.Error("загрузка должна читать только из data/input")
	}
	if err := os.MkdirAll(filepath.Join(dir, "input"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(dir, "x.csv"), filepath.Join(dir, "input", "x.csv")); err != nil {
		t.Fatal(err)
	}
	if out := do(h, "POST", "/load", url.Values{"file": {"x"}}); !strings.Contains(out, "Величин: <b>120</b>") {
		t.Errorf("load: %.200s", out)
	}
	if out := do(h, "POST", "/load", url.Values{"file": {"missing"}}); !strings.Contains(out, `class="error"`) {
		t.Error("missing file must give an error")
	}
}

func TestOps(t *testing.T) {
	h, _ := newHandler(t)
	if page := do(h, "GET", "/ops", nil); !strings.Contains(page, "<polygon") || !strings.Contains(page, `name="op"`) {
		t.Error("ops page must contain chart and operation selector")
	}
	q := func(op, a2, b2 string) string {
		return do(h, "GET", "/ops/calc?a1=1&b1=3&a2="+a2+"&b2="+b2+"&op="+op, nil)
	}
	if out := q("add", "2", "5"); strings.Count(out, "<polygon") != 3 || strings.Contains(out, `class="error"`) {
		t.Errorf("add: %.300s", out)
	}
	if out := q("div", "-1", "1"); !strings.Contains(out, `class="error"`) || strings.Count(out, "<polygon") != 2 {
		t.Errorf("div by interval with zero must report error and show only A and B: %.300s", out)
	}
	if out := q("add", "5", "2"); !strings.Contains(out, `class="error"`) {
		t.Error("a >= b must give an error")
	}
}

func TestStatic(t *testing.T) {
	h, _ := newHandler(t)
	if out := do(h, "GET", "/static/htmx.min.js", nil); !strings.Contains(out, "htmx") {
		t.Error("htmx must be served")
	}
}
