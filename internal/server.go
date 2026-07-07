package internal

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sync"
)

type Server struct {
	Config    *Config
	Library   *Library
	Mux       *http.ServeMux
	templates map[string]*template.Template
	funcs     template.FuncMap
	Thumbs    *ThumbnailCache
	mu        sync.RWMutex
}

func NewServer() *Server {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Erreur chargement config : %v", err)
	}
	lib, err := LoadLibrary()
	if err != nil {
		log.Fatalf("Erreur chargement index : %v", err)
	}

	srv := &Server{
		Config:  cfg,
		Library: lib,
		Mux:     http.NewServeMux(),
		Thumbs:  NewThumbnailCache(200),
		funcs: template.FuncMap{
			"typeEmoji": TypeEmoji,
			"json":      func(v any) string { b, _ := json.Marshal(v); return string(b) },
		},
	}

	if len(cfg.Folders) > 0 && len(lib.Files) == 0 {
		log.Println("Premier scan des dossiers...")
		srv.Library = ScanFolders(cfg)
	}

	srv.loadTemplates()
	return srv
}

func (s *Server) loadTemplates() {
	s.templates = make(map[string]*template.Template)
	pages := []string{"index", "library", "folder", "player", "settings", "search"}
	for _, page := range pages {
		tmpl := template.New("").Funcs(s.funcs)
		template.Must(tmpl.ParseFiles("templates/layout.html", "templates/"+page+".html", "templates/pagination.html"))
		s.templates[page] = tmpl
	}
}

func (s *Server) render(w http.ResponseWriter, page string, data any) {
	tmpl, ok := s.templates[page]
	if !ok {
		http.Error(w, "Template introuvable", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "layout.html", data)
}

func (s *Server) renderPartial(w http.ResponseWriter, page string, block string, data any) {
	tmpl, ok := s.templates[page]
	if !ok {
		http.Error(w, "Template introuvable", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, block, data)
}

func (s *Server) ReloadIndex() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Library = ScanFolders(s.Config)
	s.loadTemplates()
}
