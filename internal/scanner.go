package internal

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const indexFile = "index.json"

func LoadLibrary() (*Library, error) {
	lib := &Library{Files: make(map[string]*Media), Folders: make(map[string]*Folder)}
	data, err := os.ReadFile(indexFile)
	if err != nil {
		if os.IsNotExist(err) {
			return lib, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, lib); err != nil {
		return nil, err
	}
	if lib.Files == nil {
		lib.Files = make(map[string]*Media)
	}
	if lib.Folders == nil {
		lib.Folders = make(map[string]*Folder)
	}
	return lib, nil
}

func SaveLibrary(lib *Library) error {
	data, err := json.MarshalIndent(lib, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexFile, data, 0644)
}

func ScanFolders(cfg *Config) *Library {
	lib := &Library{
		Files:   make(map[string]*Media),
		Folders: make(map[string]*Folder),
	}

	for _, root := range cfg.Folders {
		entries, err := os.ReadDir(root)
		if err != nil {
			log.Printf("Impossible de lire %s : %v", root, err)
			continue
		}

		rootID := uuid.New().String()
		lib.Folders[rootID] = &Folder{
			ID:       rootID,
			Name:     filepath.Base(root),
			Path:     root,
			ParentID: "",
			Files:    []string{},
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			fullPath := filepath.Join(root, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}

			mediaID := uuid.New().String()
			media := &Media{
				ID:       mediaID,
				Name:     entry.Name(),
				Path:     fullPath,
				Type:     GetMediaType(fullPath),
				Ext:      filepath.Ext(fullPath),
				Size:     info.Size(),
				Modified: info.ModTime().Unix(),
			}
			lib.Files[mediaID] = media
			lib.Folders[rootID].Files = append(lib.Folders[rootID].Files, mediaID)
		}
	}

	if err := SaveLibrary(lib); err != nil {
		log.Printf("Erreur sauvegarde index : %v", err)
	}
	return lib
}
