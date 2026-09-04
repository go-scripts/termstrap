# Documentation des Fonctionnalités — Termstrap

Ce document constitue la référence exhaustive et à jour de toutes les fonctionnalités, balises HTML, classes CSS Bootstrap et capacités graphiques prises en charge par le moteur **Termstrap**.

---

## 1. Vue d'Ensemble de l'Architecture

Termstrap est un moteur de rendu **HTML/CSS natif vers terminal ANSI**. Il analyse l'arbre DOM HTML, résout les règles CSS (feuille de style Bootstrap intégrée + styles personnalisés) et calcule un modèle de boîte (*Box Model*) précis en cellules de caractères pour produire un affichage terminal soigné et responsive.

```
[ HTML + CSS Input ]
         │
         ▼
[ CSS Matcher (Cascadia + Douceur) ]
         │
         ▼
[ Layout Engine (Box Model 2D) ]
         │
         ▼
[ ANSI Render Engine (Lipgloss) ] ──► [ Text + Image Overlays ]
```

---

## 2. Grille & Flexbox Responsive (`.row` / `.col-*`)

Le système de grille repose sur une division en 12 colonnes avec calcul dynamique selon les seuils de largeur (*breakpoints*) du terminal.

### Breakpoints pris en charge

| Breakpoint | Préfixe de classe | Largeur minimale requise |
| :--- | :--- | :--- |
| **xs** (Extra Small) | `.col-*` | `< 60` colonnes |
| **sm** (Small) | `.col-sm-*` | $\ge 60$ colonnes |
| **md** (Medium) | `.col-md-*` | $\ge 80$ colonnes |
| **lg** (Large) | `.col-lg-*` | $\ge 120$ colonnes |
| **xl** (Extra Large) | `.col-xl-*` | $\ge 160$ colonnes |

### Comportement
- **Empilement automatique (*Stacking*)** : Si la largeur du terminal est inférieure au breakpoint spécifié (ex: un `.col-md-6` affiché sur un terminal de 50 colonnes), la colonne s'empile verticalement à 100% de la largeur.
- **Multi-breakpoints** : Vous pouvez cumuler plusieurs classes responsives (ex: `class="col-sm-12 col-md-6"`).
- **Imbrication (*Nesting*)** : Une `.row` peut être imbriquée à l'intérieur d'une `.col-*` à n'importe quel niveau de profondeur. Les largeurs sont réparties proportionnellement à la largeur de la colonne parente.

```html
<div class="row">
  <div class="col-sm-12 col-md-6 border p-1">
    <h3>Colonne Gauche</h3>
  </div>
  <div class="col-sm-12 col-md-6 border p-1">
    <h3>Colonne Droite</h3>
  </div>
</div>
```

---

## 3. Modèle de Boîte : Marges & Paddings

Les classes de marges et de paddings s'appliquent sur **n'importe quelle balise HTML** (`<div>`, `<p>`, `<img>`, `<table>`, `<span>`, etc.).

### Marges Externes (`m-*`)

| Type | Classes | Description |
| :--- | :--- | :--- |
| **Globales** | `.m-0` à `.m-5` | Marge sur les 4 côtés (de 0 à 5 cellules/lignes) |
| **Horizontales** | `.mx-0` à `.mx-5` | Marges gauche et droite |
| **Verticales** | `.my-0` à `.my-5` | Marges haut et bas |
| **Haut** | `.mt-0` à `.mt-5` | Marge supérieure |
| **Bas** | `.mb-0` à `.mb-5` | Marge inférieure |
| **Gauche / Start** | `.ms-0` à `.ms-5`, `.ml-0` à `.ml-5` | Marge gauche |
| **Droite / End** | `.me-0` à `.me-5`, `.mr-0` à `.mr-5` | Marge droite |

### Paddings Internes (`p-*`)

| Type | Classes | Description |
| :--- | :--- | :--- |
| **Globaux** | `.p-0` à `.p-5` | Espacement interne sur les 4 côtés |
| **Horizontaux** | `.px-0` à `.px-5` | Paddings gauche et droite |
| **Verticaux** | `.py-0` à `.py-5` | Paddings haut et bas |
| **Haut** | `.pt-0` à `.pt-5` | Padding supérieur |
| **Bas** | `.pb-0` à `.pb-5` | Padding inférieur |
| **Gauche / Start** | `.ps-0` à `.ps-5`, `.pl-0` à `.pl-5` | Padding gauche |
| **Droite / End** | `.pe-0` à `.pe-5`, `.pr-0` à `.pr-5` | Padding droit |

---

## 4. Bordures & Coins Arrondis

| Classe | Description |
| :--- | :--- |
| `.border` | Bordure rectangulaire complète sur les 4 côtés |
| `.rounded` | Coins arrondis utilisant les caractères Unicode `╭ ╮ ╰ ╯` |
| `.border-top` | Bordure supérieure uniquement |
| `.border-bottom` | Bordure inférieure uniquement |
| `.border-left`, `.border-start` | Bordure gauche uniquement |
| `.border-right`, `.border-end` | Bordure droite uniquement |

*Note : Les bordures partielles peuvent être librement combinées (ex: `class="border-top border-bottom"` ou `class="border-left border-right"`).*

---

## 5. Ombres Portées (*Box Shadows*)

| Classe | Niveau d'ombre | Effet visuel |
| :--- | :--- | :--- |
| `.shadow-sm` | Taille 1 | Ombre légère et discrète sous la boîte |
| `.shadow`, `.shadow-md` | Taille 2 | Ombre moyenne standard |
| `.shadow-lg` | Taille 3 | Ombre prononcée avec relief accentué |
| `.shadow-none` | Taille 0 | Supprime toute ombre |

*Système d'auto-clamping : Si une boîte approche de la limite droite du terminal, l'ombre se compresse automatiquement pour éviter tout débordement ou retour à la ligne forcé.*

---

## 6. Palette de Couleurs Bootstrap

Les couleurs sont prises en charge en **TrueColor 24-bit** (ou fallback ANSI 256).

### Couleurs d'Arrière-Plan (`bg-*`)
- `.bg-primary` (Bleu #0d6efd)
- `.bg-secondary` (Gris #6c757d)
- `.bg-success` (Vert #198754)
- `.bg-danger` (Rouge #dc3545)
- `.bg-warning` (Jaune #ffc107)
- `.bg-info` (Cyan #0dcaf0)
- `.bg-dark` (Noir/Anthracite #212529)
- `.bg-light` (Gris clair #f8f9fa)

### Couleurs de Texte (`text-*`)
- `.text-primary`, `.text-secondary`, `.text-success`, `.text-danger`, `.text-warning`, `.text-info`, `.text-dark`, `.text-light`, `.text-white`, `.text-muted`.

---

## 7. Typographie & Alignement

### Alignement Textuel
- `.text-start`, `.text-left` : Alignement à gauche (comportement par défaut).
- `.text-center` : Centrage du texte dans son conteneur.
- `.text-end`, `.text-right` : Alignement à droite.

### Styles de Police
- `.fw-bold`, `.text-bold`, `<b>`, `<strong>` : Texte en gras.
- `.fst-italic`, `.text-italic`, `<i>`, `<em>` : Texte en italique.
- `.text-decoration-underline` : Texte souligné.

---

## 8. Balises HTML Prises en Charge

| Balise | Rendu visuel dans le terminal |
| :--- | :--- |
| `<h1>` à `<h6>` | Titres en gras avec espacements verticaux automatiques |
| `<p>` | Paragraphe de texte avec retour à la ligne automatique (*word-wrap*) |
| `<table>`, `<tr>`, `<th>`, `<td>` | Tableau complet avec colonnes alignées et bordures d'en-tête |
| `<blockquote>` | Citation avec bordure latérale gauche et retrait |
| `<code>` | Code inline sur fond sombre avec texte contrasté |
| `<pre>` | Bloc de code préformaté préservant l'indentation |
| `<hr>` | Ligne horizontale de séparation |
| `<ul>`, `<ol>`, `<li>` | Listes ordonnées et non-ordonnées |
| `<img>` | Image rendue en demi-blocs ANSI ou transmise en overlay natif |

---

## 9. Rendu d'Images (`<img>`)

### Formats d'Images
- **PNG** (avec transparence alpha)
- **JPEG**
- **WebP**
- **GIF**

### Sources
- **Distantes** : URLs HTTP/HTTPS (`<img src="https://example.com/image.png" />`).
- **Locales** : Fichiers locaux résolus via `RootPath` (`<img src="poster.png" />`).

### Protocoles Terminaux
1. **HalfBlock (Unicode `▀`/`▄`)** : Protocole universel compatible avec tous les terminaux modernes. Rendu direct dans la grille de texte.
2. **Protocoles Graphiques Natifs (Kitty, iTerm2, Sixel)** : Système d'overlays retournant les coordonnées `(Row, Col, Width, Height)` pour intégration dans les frameworks TUI comme Bubble Tea.

---

## 10. Styles Personnalisés & Feuilles de Style Externes

### Styles Inline (`style="..."`)
Vous pouvez définir des propriétés CSS directement sur les éléments HTML :
```html
<div style="background-color: #ff5500; color: #ffffff; margin: 2; padding: 1;">
  Contenu stylé manuellement
</div>
```

### Feuilles de Style Externes (`WithStylesheets`)
Injectez vos propres règles CSS pour surcharger ou enrichir les classes Bootstrap :
```go
customCSS := `
  .badge-custom {
    background-color: #6f42c1;
    color: #ffffff;
    padding: 0 1;
    border-radius: true;
  }
`

m := termstrap.New(htmlContent, 
    termstrap.WithWidth(100),
    termstrap.WithStylesheets(customCSS),
)
output, err := m.Render()
```

---

## 11. Guide d'Utilisation de l'API Go

```go
package main

import (
    "fmt"
    "github.com/go-scripts/termstrap"
)

func main() {
    html := `
    <div class="row">
      <div class="col-md-6 bg-dark text-white p-2 border rounded shadow-lg">
        <h2 class="text-center">Serveur Actif</h2>
        <p>Le serveur fonctionne normalement.</p>
      </div>
      <div class="col-md-6 p-2 text-center">
        <img src="https://go.dev/doc/gopher/frontpage.png" alt="Gopher" />
      </div>
    </div>`

    // Initialisation du modèle
    m := termstrap.New(html, 
        termstrap.WithWidth(90),
        termstrap.WithRootPath("./assets"),
    )

    // Rendu en chaîne ANSI
    output, err := m.Render()
    if err != nil {
        panic(err)
    }

    fmt.Print(output)
}
```
