# MediaHub

Serveur média domestique en Go + HTMX + TailwindCSS.

Scanne des dossiers locaux et expose une interface web pour naviguer, lire, streamer et télécharger des médias depuis le navigateur (mobile/desktop) sur le réseau local.

## Fonctionnalités

- 📂 **Bibliothèque** — parcourez vos dossiers de médias
- 🎬 **Lecteur vidéo** — streaming avec support Range (avance rapide)
- 🎵 **Lecteur audio** — écoutez votre musique
- 📷 **Visualiseur d'images** — avec miniatures générées
- 🔍 **Recherche** — trouvez vos fichiers instantanément
- ⬇ **Téléchargement** — récupérez vos fichiers
- 📱 **Responsive** — fonctionne sur mobile et desktop
- 🌐 **Réseau local** — accessible depuis tous vos appareils

## Stack

| Couche       | Technologie                          |
|--------------|--------------------------------------|
| Langage      | Go 1.26                              |
| HTTP         | `net/http` + `http.ServeMux`         |
| Templates    | `html/template`                      |
| Frontend     | HTMX 2.x + TailwindCSS (moteur CDN)  |
| Données      | `config.json` + `index.json`         |

## Démarrage rapide

```powershell
# Lancer le serveur
go run .

# Ou construire l'exécutable
go build -o mediahub.exe .
./mediahub.exe
```

Le serveur démarre sur `http://localhost:8080`.

Ajoutez des dossiers dans **Paramètres** → le scan indexe vos fichiers → naviguez dans la **Bibliothèque**.

## Structure

```
mediahub/
├── main.go
├── internal/        # Logique serveur (handlers, scanner, etc.)
├── templates/       # Templates HTML (layout, pages, partials)
├── static/          # Fichiers statiques (CSS, JS, favicon)
├── config.json      # Dossiers surveillés
└── index.json       # Index des fichiers scannés
```

## Crédits

**Développé par Stéphane Bamby**  
Fondateur de Bambyno  
Développeur Fullstack  
stephanebazebibouta@gmail.com

## Licence

Projet personnel — usage domestique.
