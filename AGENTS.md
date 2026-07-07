# MediaHub

Serveur média domestique en Go + HTMX + TailwindCSS.
Scanne des dossiers locaux et expose une interface web
pour naviguer, lire, streamer et télécharger des médias
depuis le navigateur (mobile/desktop) sur le réseau local.

## Stack

| Couche       | Technologie                          |
|--------------|--------------------------------------|
| Langage      | Go 1.26                              |
| HTTP         | `net/http` + `http.ServeMux` (Go 1.22+) |
| Templates    | `html/template`                      |
| Frontend     | HTMX 2.x + TailwindCSS (moteur CDN)  |
| Données      | `config.json` + `index.json`         |

## Arborescence

```
mediahub/
├── main.go                  # Entry point
├── go.mod
├── AGENTS.md
├── config.json              # Dossiers surveillés (persistant)
├── index.json               # Index des fichiers scannés (généré automatiquement)
├── internal/
│   ├── config.go            # Lecture/écriture config.json
│   ├── server.go            # Configuration et démarrage du serveur HTTP
│   ├── routes.go            # Enregistrement des routes
│   ├── handlers.go          # Handlers HTTP (pages, streaming, download)
│   ├── scanner.go           # Scan des dossiers, construction de l'index
│   ├── library.go           # Types Media, Folder, Library + logique
│   └── middleware.go         # Logging
├── templates/
│   ├── layout.html          # Layout global (nav, footer, head)
│   ├── index.html           # Page d'accueil (tuiles + récents)
│   ├── library.html         # Bibliothèque (arborescence + filtre type)
│   ├── folder.html          # Contenu d'un dossier
│   ├── player.html          # Lecteur vidéo/audio/image
│   └── settings.html        # Paramètres (gestion dossiers)
└── static/
    ├── css/
    │   └── style.css
    └── js/
        ├── tailwind.min.js  # Moteur TailwindCDN (local)
        └── main.js
```

## Modèles

### Media (fichier individuel)

```go
type Media struct {
    ID       string  // UUID v4
    Name     string  // Nom du fichier
    Path     string  // Chemin absolu
    Type     string  // "video" | "audio" | "image" | "other"
    Ext      string  // Extension (.mp4, .jpg, ...)
    Size     int64   // Taille en octets
    Modified int64   // Timestamp de modification
}
```

### Folder (dossier)

```go
type Folder struct {
    ID       string   // UUID v4
    Name     string   // Nom du dossier
    Path     string   // Chemin absolu
    ParentID string   // ID du dossier parent ("" si racine)
    Files    []string // IDs des fichiers contenus
}
```

### Library (index complet)

```go
type Library struct {
    Files   map[string]*Media
    Folders map[string]*Folder
}
```

### Config

```go
type Config struct {
    Folders []string // Chemins absolus des dossiers surveillés
}
```

## Types MIME (extension → type)

- `.mp4`, `.mkv`, `.avi`, `.mov`, `.webm`, `.m4v` → `video`
- `.mp3`, `.flac`, `.wav`, `.aac`, `.ogg`, `.wma`, `.m4a` → `audio`
- `.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, `.webp`, `.svg` → `image`
- tout autre → `other`

## Routes

| Méthode | Route                 | Handler            | Description                       |
|---------|-----------------------|--------------------|-----------------------------------|
| GET     | `/`                   | `handleHome`       | Accueil : tuiles + 10 récents     |
| GET     | `/library`            | `handleLibrary`    | Bibliothèque (arborescence)       |
| GET     | `/library?type=`      | `handleLibrary`    | Filtrer par type                  |
| GET     | `/browse/{id}`        | `handleBrowse`     | Contenu d'un dossier              |
| GET     | `/player/{id}`        | `handlePlayer`     | Lecteur adapté au type            |
| GET     | `/settings`           | `handleSettings`   | Paramètres                        |
| POST    | `/settings/folder`    | `handleAddFolder`  | Ajouter dossier surveillé         |
| DELETE  | `/settings/folder`    | `handleRemoveFolder`| Supprimer dossier surveillé      |
| POST    | `/scan`               | `handleScan`       | Rescanner l'index                 |
| GET     | `/stream/{id}`        | `handleStream`     | Streaming (support Range)         |
| GET     | `/download/{id}`      | `handleDownload`   | Téléchargement fichier            |
| GET     | `/folder-list/{id}`   | `handleFolderList` | Partial HTMX pour dossier         |
| GET     | `/static/*`           | file server        | Fichiers statiques                |

## Streaming (Range Support)

Le handler `/stream/{id}` implémente le RFC 7233 :
1. Lit le header `Range` de la requête
2. Ouvre le fichier via l'index
3. Utilise `file.Seek()` + `io.CopyN()` pour servir la plage demandée
4. Répond avec `206 Partial Content` et les headers appropriés
5. Si pas de Range, sert le fichier en entier avec `200 OK`

## Conventions de code

### Backend
- Les handlers prennent `(w http.ResponseWriter, r *http.Request)`
- Les routes sont définies avec la syntaxe `"METHOD /path"` (Go 1.22+)
- L'accès à Library est protégé par `sync.RWMutex` via `s.mu`
- Les templates sont rechargés après un rescann
- Les chemins Windows sont stockés tels quels (avec backslashes)

### Frontend
- Navigation HTMX : `hx-boost` sur les liens `<a>`
- Layout commun via `layout.html` avec bloc `{{define "content"}}`
- Templates partiels pour les réponses HTMX (`folder_content`, `player_content`)
- TailwindCSS via le moteur CDN local dans `static/js/tailwind.min.js`
- Interface responsive (Tailwind breakpoints)

### Templates
- Le layout principal est `layout.html` qui définit la nav + le bloc `content`
- Chaque page définit `{{define "content"}}...{{end}}` qui est injecté dans le layout
- Les réponses HTMX partielles (sans layout) définissent un bloc nommé spécifique
- Fonctions template disponibles : `typeEmoji`, `json`

## Fonctions template

- `typeEmoji(type string) string` : retourne l'emoji correspondant au type média
- `json(v any) string` : encode une valeur en JSON (pour hx-vals)

## Scan

Déclenché au démarrage (si index vide) et sur `POST /scan` :
1. Pour chaque dossier de `Config.Folders`
2. `os.ReadDir()` pour lister le contenu (1 niveau, pas de récursion)
3. Chaque fichier reçoit un UUID v4, détection de type MIME
4. Chaque dossier racine possède sa liste plate de fichiers
5. L'index est sauvegardé dans `index.json`
6. Remplacement atomique de l'index en mémoire

**Note :** Le scan est volontairement non-récursif. Seuls les fichiers directement dans les dossiers surveillés sont indexés. Pas de sous-dossiers.

## Navigation

L'interface est conçue pour être simple et plate :
- **Accueil** : 4 tuiles (compteurs par type) + 10 derniers fichiers ajoutés
- **Bibliothèque** : Liste des dossiers surveillés + possibilité de filtrer par type
- **Browse** : Affiche tous les fichiers d'un dossier racine (liste plate, pas d'arborescence)
- **Player** : Lecteur adapté au type du fichier (video/audio/image)

## Variables d'environnement

- `PORT` : port du serveur (défaut : `8080`)

## Construction et exécution

```powershell
# Build
go build -o mediahub.exe .

# Run
.\mediahub.exe
# → http://localhost:8080

# Dev (reload à chaque modification)
go run .
```

## Design decisions

- **Pas de watcher fichier** : rescann manuel via bouton "Actualiser"
- **Pas de base de données** : index en mémoire + fichier JSON
- **IDs UUID** : attribués au scan, pas de collision, pas de réutilisation
- **Streaming natif** : pas de transcodage, le navigateur gère les codecs
- **Pas d'auth** : prévu pour réseau local uniquement (MVP)

## Extensions futures (hors MVP)

- Recherche plein texte
- Miniatures générées pour les vidéos
- Mode mosaïque pour les images
- SQLite
- Auth basique (mot de passe unique)
- Filesystem watcher (fsnotify)
- Favoris / playlist
- Mode hors-ligne (PWA)
