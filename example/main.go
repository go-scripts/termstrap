package main

import (
	"fmt"
	"os"

	"github.com/go-scripts/termstrap"
	"golang.org/x/term"
)

func main() {
	// Detect terminal width
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	content := `# Termstrap Demo

This is a **markdown** section rendered with [glamour](https://github.com/charmbracelet/glamour).

- Supports *italics*, **bold**, and ` + "`inline code`" + `
- Lists, headers, [links](https://example.com), and more
- Visit the repo at https://github.com/go-scripts/termstrap

---

## 1. Film Card — Grid + Image + Table

<div class="row">
  <div class="col-md-4 border rounded p-1">

![poster](https://zimage.cc/uploads/screen/d0a8490a3e1217e5f3d53f92780465955cd0b65c.webp)

  </div>
  <div class="col-md-8 ps-2">

### Gladiator II

**Ridley Scott** — 2024

| Info       | Valeur           |
|------------|------------------|
| Année      | 2024             |
| Genre      | Action, Drame    |
| Durée      | 2h28             |
| Qualité    | MULTI HDLight    |
| Langue     | FR / EN          |

> *"Those who are about to die, salute you."*

  </div>
</div>

---

## 2. Colored Alerts — Background & Text Colors

<div class="row">
  <div class="col-md-4 bg-success text-white p-2 rounded">

**Succès !** Le téléchargement est terminé. Le fichier a été enregistré dans le dossier de destination.

  </div>
  <div class="col-md-4 bg-warning text-dark p-2 rounded">

**Attention !** L'espace disque est presque plein. Veuillez libérer de l'espace pour continuer.

  </div>
  <div class="col-md-4 bg-danger text-white p-2 rounded">

**Erreur !** Impossible de se connecter au serveur. Vérifiez votre connexion réseau.

  </div>
</div>

---

## 3. Shadows — sm, normal, lg

<div class="row">
  <div class="col-md-4 border rounded shadow-sm p-2 m-1">

### Shadow SM

Petite ombre subtile.

  </div>
  <div class="col-md-4 border rounded shadow p-2 m-1">

### Shadow Normal

Ombre moyenne standard.

  </div>
  <div class="col-md-4 border rounded shadow-lg p-2 m-1">

### Shadow LG

Grande ombre prononcée.

  </div>
</div>

---

## 4. Text Alignment

<div class="row">
  <div class="col-md-4 border text-start p-1">

**Aligné à gauche**

Lorem ipsum dolor sit amet, consectetur adipiscing elit.

  </div>
  <div class="col-md-4 border text-center p-1">

**Centré**

Lorem ipsum dolor sit amet, consectetur adipiscing elit.

  </div>
  <div class="col-md-4 border text-end p-1">

**Aligné à droite**

Lorem ipsum dolor sit amet, consectetur adipiscing elit.

  </div>
</div>

---

## 5. Borders Variants

<div class="row">
  <div class="col-md-3 border p-1">

**border**

All sides

  </div>
  <div class="col-md-3 border-top p-1">

**border-top**

Top only

  </div>
  <div class="col-md-3 border-bottom p-1">

**border-bottom**

Bottom only

  </div>
  <div class="col-md-3 border-left border-right p-1">

**border-left + right**

Sides only

  </div>
</div>

---

## 6. Nested Markdown — Long Text, Code & Links

<div class="row">
  <div class="col-md-6 border rounded p-2">

### Installation

` + "```bash" + `
go get github.com/go-scripts/termstrap
` + "```" + `

Puis dans votre code :

` + "```go" + `
m := termstrap.Model{
    Content: content,
    Width:   80,
}
output, _ := m.Render()
fmt.Print(output)
` + "```" + `

  </div>
  <div class="col-md-6 border rounded p-2">

### Features

1. **Grid system** — Bootstrap 12-column responsive layout
2. **Markdown** — Full rendering via [glamour](https://github.com/charmbracelet/glamour)
3. **Images** — ANSI art from URLs or local files
4. **Styling** — Padding, margins, borders, shadows
5. **Colors** — Background and text colors (Bootstrap palette)
6. **Responsive** — Breakpoints: xs, sm, md, lg, xl

> Use ` + "`col-md-*`" + ` classes for medium terminals (80+ cols).

  </div>
</div>

---

## 7. Bold Text & Color Variants

<div class="row">
  <div class="col-md-3 bg-primary text-white p-2 fw-bold">

Primary

  </div>
  <div class="col-md-3 bg-secondary text-white p-2 fw-bold">

Secondary

  </div>
  <div class="col-md-3 bg-info text-dark p-2 fw-bold">

Info

  </div>
  <div class="col-md-3 bg-dark text-light p-2 fw-bold">

Dark

  </div>
</div>

---

## 8. Wide Layout — 2 Columns with Long Content

<div class="row">
  <div class="col-md-8 p-2 border-left">

### Description du film

**Gladiator II** poursuit la saga épique de pouvoir, d'intrigue et de vengeance dans la Rome antique. Des années après avoir assisté à la mort du vénéré héros Maximus aux mains de son oncle, Lucius est contraint d'entrer dans le Colisée après que sa patrie a été conquise par les empereurs tyranniques qui dirigent maintenant Rome d'une main de fer.

Le cœur brûlant de rage et l'avenir de l'Empire en jeu, Lucius doit se tourner vers son passé pour trouver la force et l'honneur de rendre la gloire de Rome à son peuple.

- **Réalisateur** : Ridley Scott
- **Acteurs** : Paul Mescal, Pedro Pascal, Denzel Washington
- **Budget** : 310 millions USD

  </div>
  <div class="col-md-4 bg-light text-dark p-2 rounded">

### Liens

- [IMDb](https://www.imdb.com/title/tt9218128/)
- [AlloCiné](https://www.allocine.fr/)
- [Rotten Tomatoes](https://www.rottentomatoes.com/)
- [Wikipedia](https://en.wikipedia.org/wiki/Gladiator_II)

### Tags

` + "`Action`" + ` ` + "`Drame`" + ` ` + "`Historique`" + ` ` + "`Aventure`" + ` ` + "`Péplum`" + `

  </div>
</div>

---

## 9. Single Column — Full Width Block

<div class="row">
  <div class="col-md-12 bg-dark text-white p-3 rounded shadow">

### Statistiques du serveur

| Métrique           | Valeur    |
|--------------------|-----------|
| Torrents actifs    | 1,247     |
| Seeders totaux     | 45,891    |
| Leechers totaux    | 12,034    |
| Bande passante     | 2.4 TB/h  |
| Uptime             | 99.97%    |

  </div>
</div>

---

[Lien Magnet](magnet:?xt=urn:btih:d0a8490a3e1217e5f3d53f92780465955cd0b65c&tr=udp://tracker.opentrackr.org:1337/announce&tr=udp://p4p.arenabg.com:1337/announce&tr=udp://open.stealth.si:80/announce&tr=udp://explodie.org:6969/announce&tr=udp://open.demonii.com:1337/announce&tr=udp://opentracker.io:6969/announce&tr=udp://www.torrent.eu.org:451/announce)

---

Back to regular **markdown** after all the layout blocks. The grid system handles everything above with proper alignment and styling.
`

	m := termstrap.Model{
		Content: content,
		Width:   width,
	}

	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(output)
}
