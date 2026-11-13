package server

import (
	"embed"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/philyuchkoff/NoDebris/internal/analyzer"
	"github.com/philyuchkoff/NoDebris/internal/model"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

var (
	tmpl         *template.Template
	bookmarks    []model.Bookmark
	mu           sync.RWMutex
	checkRunning bool
	checkDone    int
	checkTotal   int
	checkResults []analyzer.LinkResult
	checkMu      sync.Mutex
	checkStop    chan struct{}
)

func init() {
	tmpl = template.Must(template.New("").Funcs(template.FuncMap{
		"mul": func(a, b int) int { return a * b },
		"add": func(a, b int) int { return a + b },
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				return nil
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil
				}
				dict[key] = values[i+1]
			}
			return dict
		},
		"statusClass": func(code int) string {
			if code == 0 {
				return ""
			}
			if code >= 200 && code < 300 {
				return "status-ok"
			}
			if code >= 300 && code < 400 {
				return "status-redirect"
			}
			return "status-broken"
		},
		"statusLabel": func(code int, errStr string) string {
			if errStr != "" {
				return "Error"
			}
			if code == 0 {
				return "\u2014"
			}
			if code >= 200 && code < 300 {
				return "OK"
			}
			if code >= 300 && code < 400 {
				return "Redirect"
			}
			return "Broken"
		},
	}).ParseFS(templateFS, "templates/*.html"))
}

func Run(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /static/", func(w http.ResponseWriter, r *http.Request) {
		http.FileServerFS(staticFS).ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /{$}", handleIndex)
	mux.HandleFunc("POST /upload", handleUpload)
	mux.HandleFunc("GET /bookmarks", handleBookmarks)
	mux.HandleFunc("POST /delete", handleDelete)
	mux.HandleFunc("POST /move", handleMove)
	mux.HandleFunc("GET /duplicates", handleDuplicates)
	mux.HandleFunc("POST /duplicates/resolve", handleDupResolve)
	mux.HandleFunc("GET /structure", handleStructure)
	mux.HandleFunc("GET /check-links", handleCheckLinksStart)
	mux.HandleFunc("GET /check-links/progress", handleCheckLinksProgress)
	mux.HandleFunc("GET /export", handleExport)
	mux.HandleFunc("GET /categorize", handleCategorize)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	return srv.ListenAndServe()
}
