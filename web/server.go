// Package web — веб-интерфейс (html/template + htmx) поверх библиотек проекта.
package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"fuzzylab/cluster"
	"fuzzylab/fuzzy/core"
	"fuzzylab/fuzzy/metrics"
	"fuzzylab/gen"
	"fuzzylab/logger"
	"fuzzylab/storage"
)

//go:embed templates static
var assets embed.FS

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{"f": func(v float64) string {
	return strconv.FormatFloat(v, 'f', 1, 64)
}}).ParseFS(assets, "templates/*.html"))

// Server хранит текущий набор величин и обслуживает HTTP-запросы.
type Server struct {
	mu         sync.Mutex
	items      []core.Number
	genForm    genForm
	dir        string
	log        *logger.Logger
	lastLogged string
	lastOp     string
}

// genForm — параметры генерации набора.
type genForm struct {
	Mode      string
	N, Groups int
	ALo, AHi  float64
	WLo, WHi  float64
	Noise     float64
	Seed      int64
}

// ctl — параметры просмотра: уровень α и кластеризация.
type ctl struct {
	Alpha     float64
	Algo      string
	K         int
	Threshold float64
}

var (
	defaultGen = genForm{Mode: "clouds", N: 200, Groups: 4, ALo: -50, AHi: 100, WLo: 1, WHi: 20, Noise: 0.03, Seed: 1}
	defaultCtl = ctl{Algo: "pam", K: 4, Threshold: 10}
)

// New создаёт сервер, генерирует набор по умолчанию; dir — каталог для файлов.
func New(dir string, log *logger.Logger) (*Server, error) {
	s := &Server{dir: dir, log: log, genForm: defaultGen}
	if err := s.generate(defaultGen); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.FileServerFS(assets))
	mux.HandleFunc("GET /{$}", s.handle(noop, "index"))
	mux.HandleFunc("GET /view", s.handle(noop, "view"))
	mux.HandleFunc("GET /ops", s.opsHandler("ops"))
	mux.HandleFunc("GET /ops/calc", s.opsHandler("opsview"))
	mux.HandleFunc("POST /generate", s.handle(s.onGenerate, "view"))
	mux.HandleFunc("POST /save", s.handle(s.onSave, "view"))
	mux.HandleFunc("POST /load", s.handle(s.onLoad, "view"))
	return mux
}

// handle выполняет действие под замком и рисует шаблон page с сообщением или ошибкой.
func (s *Server) handle(act func(*http.Request) (string, error), page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		_ = r.ParseForm()
		msg, err := act(r)
		v := s.buildView(parseCtl(r))
		v.Message = msg
		if err != nil {
			v.Error = err.Error()
		}
		data := any(v)
		if page == "index" {
			data = struct {
				View viewModel
				Gen  genForm
				Ctl  ctl
				File string
			}{v, s.genForm, defaultCtl, "set.csv"}
		}
		if err := tmpl.ExecuteTemplate(w, page, data); err != nil {
			http.Error(w, "ошибка отображения страницы: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *Server) onGenerate(r *http.Request) (string, error) {
	f := genForm{Mode: r.Form.Get("mode")}
	var err error
	for _, p := range []struct {
		name, label string
		dst         any
	}{
		{"n", "число величин", &f.N}, {"groups", "облаков", &f.Groups},
		{"alo", "a от", &f.ALo}, {"ahi", "a до", &f.AHi},
		{"wlo", "ширина от", &f.WLo}, {"whi", "ширина до", &f.WHi},
		{"noise", "шум", &f.Noise}, {"seed", "seed", &f.Seed},
	} {
		if e := scan(r.Form.Get(p.name), p.dst); e != nil && err == nil {
			err = fmt.Errorf("поле «%s»: %w", p.label, e)
		}
	}
	if err != nil {
		return "", err
	}
	if err := s.generate(f); err != nil {
		return "", err
	}
	s.genForm = f
	return fmt.Sprintf("Сгенерировано величин: %d", len(s.items)), nil
}

func (s *Server) generate(f genForm) error {
	p := gen.Params{N: f.N, A: gen.Range{Lo: f.ALo, Hi: f.AHi}, W: gen.Range{Lo: f.WLo, Hi: f.WHi}, Seed: f.Seed}
	var items []core.Number
	var err error
	logData := map[string]any{
		"число": f.N, "a_от": f.ALo, "a_до": f.AHi, "ширина_от": f.WLo, "ширина_до": f.WHi, "seed": f.Seed,
	}
	if f.Mode == "uniform" {
		logData["режим"] = "равномерно"
		items, err = gen.Uniform(p)
	} else {
		logData["режим"], logData["облаков"], logData["шум"] = "облака", f.Groups, f.Noise
		items, _, err = gen.Clouds(gen.CloudParams{Params: p, K: f.Groups, Noise: f.Noise})
	}
	if err != nil {
		return err
	}
	s.items = items
	return s.log.Log("генерация", logData)
}

func (s *Server) onSave(r *http.Request) (string, error) {
	path := s.path(s.dir, r.Form.Get("file"))
	if err := storage.Save(path, s.items); err != nil {
		return "", err
	}
	return "Сохранено: " + path, s.log.Log("сохранение", map[string]any{"файл": path, "число": len(s.items)})
}

func (s *Server) onLoad(r *http.Request) (string, error) {
	path := s.path(filepath.Join(s.dir, "input"), r.Form.Get("file"))
	items, err := storage.Load(path)
	if err != nil {
		return "", err
	}
	s.items = items
	return "Загружено: " + path, s.log.Log("загрузка", map[string]any{"файл": path, "число": len(items)})
}

// path приводит имя файла к <dir>/<имя>.csv.
func (s *Server) path(dir, name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "." || name == string(filepath.Separator) {
		name = "set"
	}
	if filepath.Ext(name) != ".csv" {
		name += ".csv"
	}
	return filepath.Join(dir, name)
}

func parseCtl(r *http.Request) ctl {
	c := defaultCtl
	_ = scan(r.Form.Get("alpha"), &c.Alpha)
	_ = scan(r.Form.Get("k"), &c.K)
	_ = scan(r.Form.Get("threshold"), &c.Threshold)
	if r.Form.Get("algo") == "agglo" {
		c.Algo = "agglo"
	}
	c.Alpha = min(max(c.Alpha, 0), 1)
	return c
}

// cluster кластеризует набор по срезам на уровне α и пишет запуск в журнал.
func (s *Server) cluster(c ctl) (cluster.Result, error) {
	dm := metrics.DistanceMatrix(s.items, metrics.AlphaCutDistance{Alpha: c.Alpha})
	var res cluster.Result
	var err error
	algo := "k-медоидов"
	params := map[string]any{"α": c.Alpha}
	if c.Algo == "agglo" {
		algo, params["порог"] = "иерархический", c.Threshold
		res, err = cluster.Agglomerative(dm, c.Threshold)
	} else {
		params["k"] = min(max(c.K, 1), len(s.items))
		res, err = cluster.PAM(dm, params["k"].(int))
	}
	if err != nil {
		return res, err
	}
	if key := fmt.Sprint(algo, params, len(s.items)); key != s.lastLogged {
		s.lastLogged = key
		err = s.log.Clustering(algo, params, res)
	}
	return res, err
}

// scan разбирает строку в *int, *int64 или *float64; при ошибке dst не меняется.
func scan(s string, dst any) error {
	var err error
	switch d := dst.(type) {
	case *int:
		var v int
		if v, err = strconv.Atoi(s); err == nil {
			*d = v
		}
	case *int64:
		var v int64
		if v, err = strconv.ParseInt(s, 10, 64); err == nil {
			*d = v
		}
	case *float64:
		var v float64
		if v, err = strconv.ParseFloat(s, 64); err == nil {
			*d = v
		}
	default:
		return fmt.Errorf("неподдерживаемый тип %T", dst)
	}
	if err != nil {
		return fmt.Errorf("ожидалось число, получено %q", s)
	}
	return nil
}

func noop(*http.Request) (string, error) { return "", nil }
