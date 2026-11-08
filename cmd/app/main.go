// Command app запускает веб-интерфейс кластеризации нечётких величин.
package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"

	"fuzzylab/logger"
	"fuzzylab/web"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "адрес сервера")
	dir := flag.String("data", "data", "каталог для наборов и журнала")
	flag.Parse()

	l, err := logger.Open(filepath.Join(*dir, "log.json"))
	if err != nil {
		log.Fatalf("не удалось открыть журнал: %v", err)
	}
	defer l.Close()
	s, err := web.New(*dir, l)
	if err != nil {
		log.Fatalf("не удалось запустить сервер: %v", err)
	}
	log.Printf("сервер запущен: http://%s", *addr)
	log.Fatalf("сервер остановлен: %v", http.ListenAndServe(*addr, s.Handler()))
}
