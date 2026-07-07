package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := map[string]int{"video": 0, "audio": 0, "image": 0, "other": 0}
	for _, m := range s.Library.Files {
		counts[m.Type]++
	}
	recent := s.Library.RecentFiles(10)

	data := map[string]any{
		"Title":  "Accueil",
		"Counts": counts,
		"Recent": recent,
	}
	s.render(w, "index", data)
}

func (s *Server) handleLibrary(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mediaType := r.URL.Query().Get("type")
	var all []*Media
	if mediaType != "" {
		all = s.Library.ByType(mediaType)
	} else {
		for _, m := range s.Library.Files {
			all = append(all, m)
		}
	}

	perPage := 50
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	p := NewPagination(len(all), page, perPage)
	var files []*Media
	start := p.Offset()
	end := start + perPage
	if end > len(all) {
		end = len(all)
	}
	if start < len(all) {
		files = all[start:end]
	}

	base := "/library"
	q := r.URL.Query()
	q.Del("page")
	if len(q) > 0 {
		base += "?" + q.Encode()
	}

	data := map[string]any{
		"Title":      "Bibliothèque",
		"Roots":      s.Library.RootFolders(),
		"Files":      files,
		"MediaType":  mediaType,
		"Pagination": p,
		"BaseURL":    base,
	}
	s.render(w, "library", data)
}

func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	folderID := r.PathValue("id")
	folder, ok := s.Library.Folders[folderID]
	if !ok {
		http.Error(w, "Dossier introuvable", http.StatusNotFound)
		return
	}

	all := s.Library.AllFiles(folderID)

	perPage := 50
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	p := NewPagination(len(all), page, perPage)
	var files []*Media
	start := p.Offset()
	end := start + perPage
	if end > len(all) {
		end = len(all)
	}
	if start < len(all) {
		files = all[start:end]
	}

	base := "/browse/" + folderID
	q := r.URL.Query()
	q.Del("page")
	if len(q) > 0 {
		base += "?" + q.Encode()
	}

	if r.Header.Get("HX-Request") == "true" {
		data := map[string]any{
			"Folder":     folder,
			"Files":      files,
			"Pagination": p,
			"BaseURL":    base,
		}
		s.renderPartial(w, "folder", "folder_content", data)
		return
	}

	data := map[string]any{
		"Title":      folder.Name,
		"Folder":     folder,
		"Files":      files,
		"Pagination": p,
		"BaseURL":    base,
	}
	s.render(w, "folder", data)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := map[string]any{
		"Title":   "Paramètres",
		"Folders": s.Config.Folders,
	}
	s.render(w, "settings", data)
}

func (s *Server) handleAddFolder(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	if path == "" {
		var body struct{ Path string }
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			path = body.Path
		}
	}
	if path == "" {
		jsonError(w, "Chemin requis", http.StatusBadRequest)
		return
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		jsonError(w, "Chemin invalide", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		jsonError(w, "Le dossier n'existe pas", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.Config.AddFolder(absPath)
	if err := SaveConfig(s.Config); err != nil {
		s.mu.Unlock()
		jsonError(w, "Erreur sauvegarde", http.StatusInternalServerError)
		return
	}
	s.mu.Unlock()

	s.ReloadIndex()

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (s *Server) handleRemoveFolder(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	if path == "" {
		var body struct{ Path string }
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			path = body.Path
		}
	}
	if path == "" {
		jsonError(w, "Chemin requis", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.Config.RemoveFolder(path)
	if err := SaveConfig(s.Config); err != nil {
		s.mu.Unlock()
		jsonError(w, "Erreur sauvegarde", http.StatusInternalServerError)
		return
	}
	s.mu.Unlock()

	s.ReloadIndex()

	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	s.ReloadIndex()
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Scan terminé")
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	mediaID := r.PathValue("id")
	media, ok := s.Library.Files[mediaID]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Fichier introuvable", http.StatusNotFound)
		return
	}

	file, err := os.Open(media.Path)
	if err != nil {
		http.Error(w, "Erreur ouverture fichier", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Erreur stat fichier", http.StatusInternalServerError)
		return
	}
	fileSize := stat.Size()

	contentTypes := map[string]string{
		"video": "video/mp4",
		"audio": "audio/mpeg",
		"image": "image/jpeg",
	}
	contentType := contentTypes[media.Type]
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
		w.WriteHeader(http.StatusOK)
		io.Copy(w, file)
		return
	}

	var start, end int64
	_, err = fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
	if err != nil {
		_, err = fmt.Sscanf(rangeHeader, "bytes=%d-", &start)
		if err != nil {
			http.Error(w, "Range header invalide", http.StatusRequestedRangeNotSatisfiable)
			return
		}
		end = fileSize - 1
	}
	if start > end || start < 0 || end >= fileSize {
		http.Error(w, "Range invalide", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	contentLength := end - start + 1
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.Header().Set("Accept-Ranges", "bytes")
	w.WriteHeader(http.StatusPartialContent)

	file.Seek(start, io.SeekStart)
	io.CopyN(w, file, contentLength)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	mediaID := r.PathValue("id")
	media, ok := s.Library.Files[mediaID]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Fichier introuvable", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, media.Name))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, media.Path)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func (s *Server) handlePlayer(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	mediaID := r.PathValue("id")
	media, ok := s.Library.Files[mediaID]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Fichier introuvable", http.StatusNotFound)
		return
	}

	isHTMX := r.Header.Get("HX-Request") == "true"

	data := map[string]any{
		"Title": media.Name,
		"Media": media,
	}

	if isHTMX {
		s.renderPartial(w, "player", "player_content", data)
		return
	}
	s.render(w, "player", data)
}

func (s *Server) handleFolderList(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	folderID := r.PathValue("id")
	folder, ok := s.Library.Folders[folderID]
	if !ok {
		http.Error(w, "Dossier introuvable", http.StatusNotFound)
		return
	}

	files := s.Library.AllFiles(folderID)
	data := map[string]any{
		"Folder": folder,
		"Files":  files,
	}
	s.renderPartial(w, "folder", "folder_content", data)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := r.URL.Query().Get("q")
	if q == "" {
		s.handleHome(w, r)
		return
	}

	q = strings.ToLower(q)
	var all []*Media
	for _, m := range s.Library.Files {
		if strings.Contains(strings.ToLower(m.Name), q) {
			all = append(all, m)
		}
	}

	perPage := 50
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	p := NewPagination(len(all), page, perPage)

	var results []*Media
	start := p.Offset()
	end := start + perPage
	if end > len(all) {
		end = len(all)
	}
	if start < len(all) {
		results = all[start:end]
	}

	base := "/search?q=" + q
	htmx := r.Header.Get("HX-Request") == "true"

	data := map[string]any{
		"Query":      q,
		"Results":    results,
		"Title":      "Recherche",
		"Pagination": p,
		"BaseURL":    base,
	}
	if htmx {
		s.renderPartial(w, "search", "search_content", data)
		return
	}
	s.render(w, "search", data)
}

func (s *Server) handleThumbnail(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	mediaID := r.PathValue("id")
	media, ok := s.Library.Files[mediaID]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Fichier introuvable", http.StatusNotFound)
		return
	}
	if media.Type != "image" {
		http.Error(w, "Pas une image", http.StatusBadRequest)
		return
	}

	data, mime, err := s.Thumbs.Get(mediaID, media.Path)
	if err != nil || data == nil {
		http.Error(w, "Erreur génération miniature", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (s *Server) getLibraryStats() map[string]int {
	counts := map[string]int{"video": 0, "audio": 0, "image": 0, "other": 0}
	for _, m := range s.Library.Files {
		counts[m.Type]++
	}
	return counts
}
