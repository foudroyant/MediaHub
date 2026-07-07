package internal

import "net/http"

func (s *Server) SetupRoutes() {
	fs := http.FileServer(http.Dir("static"))
	s.Mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	s.Mux.HandleFunc("GET /", s.handleHome)
	s.Mux.HandleFunc("GET /library", s.handleLibrary)
	s.Mux.HandleFunc("GET /browse/{id}", s.handleBrowse)
	s.Mux.HandleFunc("GET /settings", s.handleSettings)
	s.Mux.HandleFunc("POST /settings/folder", s.handleAddFolder)
	s.Mux.HandleFunc("DELETE /settings/folder", s.handleRemoveFolder)
	s.Mux.HandleFunc("POST /scan", s.handleScan)
	s.Mux.HandleFunc("GET /stream/{id}", s.handleStream)
	s.Mux.HandleFunc("GET /download/{id}", s.handleDownload)
	s.Mux.HandleFunc("GET /player/{id}", s.handlePlayer)
	s.Mux.HandleFunc("GET /folder-list/{id}", s.handleFolderList)
	s.Mux.HandleFunc("GET /search", s.handleSearch)
	s.Mux.HandleFunc("GET /thumbnail/{id}", s.handleThumbnail)
}
