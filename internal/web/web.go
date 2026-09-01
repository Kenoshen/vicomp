// Package web wires HTTP routes to the store and PDF renderer.
package web

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mwingfield/vicomp/internal/pdf"
	"github.com/mwingfield/vicomp/internal/store"
)

type Server struct {
	store     *store.Store
	pdf       *pdf.Client
	templates map[string]*template.Template
	staticDir string
}

// New parses every page template in templatesDir against the shared layout and
// partials, keyed by the page's base filename without extension.
func New(st *store.Store, pdfClient *pdf.Client, templatesDir, staticDir string) (*Server, error) {
	layout := filepath.Join(templatesDir, "layout.html")
	partials := filepath.Join(templatesDir, "partials.html")

	pages, err := filepath.Glob(filepath.Join(templatesDir, "pages", "*.html"))
	if err != nil {
		return nil, err
	}

	funcs := template.FuncMap{
		"money": func(v float64) string { return "$" + strconv.FormatFloat(v, 'f', 2, 64) },
		"date": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02")
		},
		"datep": func(t *time.Time) string {
			if t == nil {
				return ""
			}
			return t.Format("2006-01-02")
		},
		"deref": func(p *int64) int64 {
			if p == nil {
				return 0
			}
			return *p
		},
	}

	tset := make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		name := strings.TrimSuffix(filepath.Base(page), ".html")
		t, err := template.New("").Funcs(funcs).ParseFiles(layout, partials, page)
		if err != nil {
			return nil, err
		}
		tset[name] = t
	}

	return &Server{store: st, pdf: pdfClient, templates: tset, staticDir: staticDir}, nil
}

// page data shared by every template render.
type pageData struct {
	Title string
	Error string
	Data  any
}

func (s *Server) render(w http.ResponseWriter, name string, pd pageData) {
	t, ok := s.templates[name]
	if !ok {
		http.Error(w, "template not found: "+name, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout", pd); err != nil {
		log.Printf("render %s: %v", name, err)
	}
}

// renderFragment executes a named template (from partials) without the layout,
// for HTMX partial responses.
func (s *Server) renderFragment(w http.ResponseWriter, page, defName string, data any) {
	t, ok := s.templates[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, defName, data); err != nil {
		log.Printf("render fragment %s/%s: %v", page, defName, err)
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.staticDir))))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /{$}", s.handleIndex)

	// Settings (singleton)
	mux.HandleFunc("GET /settings", s.settingsEdit)
	mux.HandleFunc("POST /settings", s.settingsUpdate)

	// Counties
	mux.HandleFunc("GET /counties", s.countiesList)
	mux.HandleFunc("GET /counties/new", s.countyNew)
	mux.HandleFunc("POST /counties", s.countyCreate)
	mux.HandleFunc("GET /counties/{id}/edit", s.countyEdit)
	mux.HandleFunc("POST /counties/{id}", s.countyUpdate)
	mux.HandleFunc("DELETE /counties/{id}", s.countyDelete)

	// Clinicians
	mux.HandleFunc("GET /clinicians", s.cliniciansList)
	mux.HandleFunc("GET /clinicians/new", s.clinicianNew)
	mux.HandleFunc("POST /clinicians", s.clinicianCreate)
	mux.HandleFunc("GET /clinicians/{id}", s.clinicianShow)
	mux.HandleFunc("GET /clinicians/{id}/edit", s.clinicianEdit)
	mux.HandleFunc("POST /clinicians/{id}", s.clinicianUpdate)
	mux.HandleFunc("DELETE /clinicians/{id}", s.clinicianDelete)

	// Clients
	mux.HandleFunc("GET /clients", s.clientsList)
	mux.HandleFunc("GET /clients/new", s.clientNew)
	mux.HandleFunc("POST /clients", s.clientCreate)
	mux.HandleFunc("GET /clients/{id}/edit", s.clientEdit)
	mux.HandleFunc("POST /clients/{id}", s.clientUpdate)
	mux.HandleFunc("DELETE /clients/{id}", s.clientDelete)

	// Sessions
	mux.HandleFunc("GET /sessions", s.sessionsList)
	mux.HandleFunc("GET /sessions/new", s.sessionNew)
	mux.HandleFunc("POST /sessions", s.sessionCreate)
	mux.HandleFunc("GET /sessions/{id}/edit", s.sessionEdit)
	mux.HandleFunc("POST /sessions/{id}", s.sessionUpdate)
	mux.HandleFunc("DELETE /sessions/{id}", s.sessionDelete)

	// Invoices
	mux.HandleFunc("GET /invoices", s.invoicesList)
	mux.HandleFunc("GET /invoices/new", s.invoiceNew)
	mux.HandleFunc("GET /invoices/preview", s.invoicePreview)
	mux.HandleFunc("POST /invoices", s.invoiceCreate)
	mux.HandleFunc("GET /invoices/{id}", s.invoiceShow)
	mux.HandleFunc("GET /invoices/{id}/pdf", s.invoicePDF)
	mux.HandleFunc("POST /invoices/{id}/status", s.invoiceStatus)
	mux.HandleFunc("DELETE /invoices/{id}", s.invoiceDelete)

	return logRequests(mux)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.render(w, "index", pageData{Title: "Dashboard"})
}

// pathID parses the {id} path value.
func pathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}
