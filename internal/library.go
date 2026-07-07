package internal

import (
	"path/filepath"
	"strings"
)

type Media struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
	Ext      string `json:"ext"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
}

type Folder struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	ParentID string   `json:"parent_id"`
	Files    []string `json:"files"`
}

type Library struct {
	Files   map[string]*Media  `json:"files"`
	Folders map[string]*Folder `json:"folders"`
}

func (l *Library) RecentFiles(n int) []*Media {
	var all []*Media
	for _, m := range l.Files {
		all = append(all, m)
	}
	var sorted []*Media
	for len(sorted) < n && len(sorted) < len(all) {
		idx := 0
		for i := 1; i < len(all); i++ {
			if all[i].Modified > all[idx].Modified {
				idx = i
			}
		}
		sorted = append(sorted, all[idx])
		all = append(all[:idx], all[idx+1:]...)
	}
	return sorted
}

func (l *Library) ByType(mediaType string) []*Media {
	var res []*Media
	for _, m := range l.Files {
		if m.Type == mediaType {
			res = append(res, m)
		}
	}
	return res
}

func (l *Library) AllFiles(folderID string) []*Media {
	var files []*Media
	folder, ok := l.Folders[folderID]
	if !ok {
		return files
	}
	for _, fid := range folder.Files {
		if m, ok := l.Files[fid]; ok {
			files = append(files, m)
		}
	}
	return files
}

type Pagination struct {
	CurrentPage int
	TotalPages  int
	PerPage     int
	Total       int
}

func NewPagination(total, page, perPage int) Pagination {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}
	totalPages := (total + perPage - 1) / perPage
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}
	return Pagination{
		CurrentPage: page,
		TotalPages:  totalPages,
		PerPage:     perPage,
		Total:       total,
	}
}

func (p Pagination) Offset() int {
	return (p.CurrentPage - 1) * p.PerPage
}

func (p Pagination) HasPrev() bool {
	return p.CurrentPage > 1
}

func (p Pagination) HasNext() bool {
	return p.CurrentPage < p.TotalPages
}

func (p Pagination) PrevPage() int {
	if p.HasPrev() {
		return p.CurrentPage - 1
	}
	return 1
}

func (p Pagination) NextPage() int {
	if p.HasNext() {
		return p.CurrentPage + 1
	}
	return p.TotalPages
}

func (p Pagination) Pages() []int {
	var pages []int
	start := p.CurrentPage - 2
	if start < 1 {
		start = 1
	}
	end := start + 4
	if end > p.TotalPages {
		end = p.TotalPages
		start = end - 4
		if start < 1 {
			start = 1
		}
	}
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	return pages
}

func (l *Library) RootFolders() []*Folder {
	var roots []*Folder
	for _, f := range l.Folders {
		if f.ParentID == "" {
			roots = append(roots, f)
		}
	}
	return roots
}

var mimeExtensions = map[string]string{
	".mp4":  "video",
	".mkv":  "video",
	".avi":  "video",
	".mov":  "video",
	".webm": "video",
	".m4v":  "video",
	".mp3":  "audio",
	".flac": "audio",
	".wav":  "audio",
	".aac":  "audio",
	".ogg":  "audio",
	".wma":  "audio",
	".m4a":  "audio",
	".jpg":  "image",
	".jpeg": "image",
	".png":  "image",
	".gif":  "image",
	".bmp":  "image",
	".webp": "image",
	".svg":  "image",
}

func GetMediaType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if t, ok := mimeExtensions[ext]; ok {
		return t
	}
	return "other"
}

var mediaEmoji = map[string]string{
	"video": "🎬",
	"audio": "🎵",
	"image": "📷",
	"other": "📄",
}

func TypeEmoji(t string) string {
	if e, ok := mediaEmoji[t]; ok {
		return e
	}
	return "📄"
}
