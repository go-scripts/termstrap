# Copilot Instructions — termstrap

## Projet

**termstrap** est une bibliothèque Go (`github.com/go-scripts/termstrap`) qui fournit un système de layout Bootstrap-like pour le terminal. Elle combine le rendu Markdown (glamour), un grid system 12 colonnes responsive, et un rendu d'images multi-protocole.

## Stack Technique

| Composant | Technologie |
|-----------|-------------|
| Langage | Go 1.25+ |
| Markdown → ANSI | `charmbracelet/glamour` |
| Styling terminal | `charmbracelet/lipgloss` |
| Parsing HTML | `PuerkitoBio/goquery` |
| Images halfblock | `stroborobo/aimg` |
| Images sixel | `mattn/go-sixel` |
| Détection terminal | `muesli/termenv` |
| Taille terminal | `golang.org/x/term` |
| Long-line wrap | `MichaelMure/go-term-markdown` |
| Tests | `testing` (stdlib) |

## Architecture

```
termstrap.go        → Model (point d'entrée public), Render()
html.go             → Extraction segments markdown/HTML
grid.go             → Parsing grid Bootstrap (row/col), résolution des largeurs
breakpoints.go      → Breakpoints responsifs (xs/sm/md/lg/xl)
classes.go          → Résolution CSS classes → style{} (padding, margin, border, colors, shadow)
renderer.go         → Pipeline de rendu : row → column → styling lipgloss, shadow, overlay
markdown.go         → Rendu markdown via glamour + injection images
images.go           → Extraction/remplacement placeholders images, rendu déféré (overlay)
utils.go            → Utilitaires ANSI : stripANSI, wrapLongLines, persistColors, hexToRGB
image/              → Sous-package : rendu images multi-protocole
  protocol.go       → Interface Renderer, types Protocol/Capabilities, constructeur NewRenderer
  detect.go         → Détection capacités terminal (env vars, sans escape sequences)
  halfblock.go      → Renderer Unicode ▄/▀ (fallback universel)
  sixel.go          → Renderer Sixel (DEC)
  iterm.go          → Renderer iTerm2 (OSC 1337)
  kitty.go          → Renderer Kitty (APC, chunked base64)
  resize.go         → Redimensionnement image + estimation hauteur visuelle
  renderer_test.go  → Tests unitaires du sous-package image
```

## Conventions Go

### Organisation des fichiers

- **Un fichier par responsabilité** : ne pas mélanger grid parsing et rendu dans le même fichier.
- **Package principal** (`package termstrap`) : flat, pas de sous-dossiers sauf `image/`.
- **Sous-package `image/`** : isolé avec sa propre interface `Renderer` et ses implémentations.
- **Tests** : `*_test.go` dans le même package (tests whitebox).
- **Exemples** : dans `examples/` avec un `main.go` principal et des sous-dossiers thématiques (`examples/image/detect/`, etc.).

### Nommage

- Types exportés : `PascalCase` (`Model`, `Protocol`, `Capabilities`).
- Types internes : `camelCase` (`gridRow`, `gridColumn`, `style`, `imageInfo`).
- Constantes : `camelCase` pour les iota (`HalfBlock`, `bpMD`), `camelCase` pour les seuils (`thresholdMD`).
- Constructeurs : `New*` pattern (`NewRenderer`).
- Méthodes privées du Model : `renderMarkdown`, `renderColumn`, `renderImages`.
- Variables : `camelCase` (`ansiRegex`, `htmlBlockRegex`). Jamais de variables sur une ou deux lettres (sauf si pertinente, ex: `i` dans `for i := 0; i < n; i++`). Le nom de la variable doit représenter ce quelle contient.

### Patterns utilisés

- **Functional options** : `NewRenderer(WithProtocol(Kitty), WithOutput(w))`.
- **Interface-based rendering** : `Renderer` interface avec `Render(img, width)` et `Protocol()`.
- **Protocol fallback hierarchy** : Kitty → iTerm2 → Sixel → HalfBlock.
- **Env var override** : `TERMSTRAP_IMAGE_PROTOCOL` pour forcer un protocole.
- **Placeholder pattern** : images extraites du markdown → placeholders → re-injection après rendu glamour.
- **Deferred rendering** : mode overlay curseur natif pour protocoles graphiques en layout multi-colonnes.

### Style de code

- **Commentaires** : doc-comments sur tous les types et fonctions exportés. Commentaires inline pour logique complexe (ANSI, shadow, overlay).
- **Gestion d'erreurs** : `log.Printf("Warning: ...")` + return gracieux (`"", false`) pour les images ; `error` propagé pour le rendu principal.
- **Immutabilité** : `Model` est une value struct (pas de pointeur receiver). Créer un nouveau `Model` pour les sous-rendus (colonnes).
- **Pas de globals mutables** : seuls `var ansiRegex` et `var htmlBlockRegex` (compiled regexp, thread-safe).
- **Imports groupés** : stdlib → espace → dépendances externes. Alias `termimage` pour le sous-package `image`.

### Grid system

- 12 colonnes, classes `col-{bp}-{n}` (Bootstrap 5).
- Breakpoints : xs (<60), sm (≥60), md (≥80), lg (≥120), xl (≥160) en colonnes terminal.
- Stacking vertical automatique sous le breakpoint.
- Classes supportées : `p-*`, `px-*`, `py-*`, `m-*`, `border`, `rounded`, `shadow`, `bg-*`, `text-*`, `text-center/end`.

### Images

- Extraction via regex `![alt](url =WxH)` avec support dimensions optionnelles.
- Chargement : URL (http/https) ou chemin local (relatif à `RootPath`).
- En layout mono-colonne : rendu direct via le `Renderer` configuré.
- En layout multi-colonnes : halfblock forcé (lipgloss alignment) OU overlay curseur natif (Kitty/iTerm2/Sixel).

### Shadow rendering

- 3 niveaux : `shadow-sm` (1), `shadow` (2), `shadow-lg` (3).
- Calcul intelligent : overflow detection, ajustement automatique, respect de `maxWidth`.
- Bottom shadow width = content width (les espaces d'offset complètent la largeur totale).

### ANSI color persistence

- `persistColors()` réinjecte bg/fg après chaque `\x1b[0m` (reset SGR) pour maintenir les couleurs à travers le rendu glamour.
- Remplacement spécifique de la couleur par défaut de glamour (palette 252) par la couleur custom.

## Tests

- **Structure AAA** : Arrange-Act-Assert.
- **Table-driven tests** : `[]struct{ name, input, expected }` avec `t.Run(tt.name, ...)`.
- **Helpers** : `newTestImage()` pour créer des images de test.
- **Assertions** : préférer `lipgloss.Width()` à `len()` pour mesurer les largeurs visuelles (ANSI-aware).
- **Validation** : `go test ./...` avant tout commit.
- **Couverture minimale** : 80%.

## Workflow de développement

1. **Analyser l'impact** : vérifier tous les fichiers concernés avant de coder.
2. **Coder** : respecter l'architecture fichier par responsabilité.
3. **Tester** : `go test ./...` — les tests doivent passer.
4. **Build** : `go build ./...` — zéro erreur.
5. **Commits** : conventional commits (`feat:`, `fix:`, `refactor:`, `chore:`, `test:`, `perf:`, `docs:`).
6. **Commits atomiques** : un commit par feature/fix logique, pas de commits fourre-tout.

## Règles critiques

- **Ne jamais committer de `fmt.Printf` de debug** — les retirer avant commit.
- **Ne jamais committer de binaires compilés** — le `.gitignore` les exclut.
- **Toujours tester la compilation** (`go build ./...`) avant de proposer un commit.
- **Préférer les value receivers** sur `Model` (pas de mutation).
- **Imports aliasés** : `termimage "github.com/go-scripts/termstrap/image"` partout dans le package principal.
