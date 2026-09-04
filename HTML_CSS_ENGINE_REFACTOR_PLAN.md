# Refonte Moteur HTML/CSS vers ANSI

## Context
Termstrap abandonne le Markdown et Glamour pour devenir un pur moteur de rendu HTML/CSS pour terminal. L'objectif est de supporter des feuilles de style CSS arbitraires, d'appliquer le modèle de boîte (Box Model) standard sur n'importe quelle balise (pas seulement les grilles Bootstrap), et de résoudre les conflits d'ANSI tout en préservant le système d'overlay natif des images (Kitty/iTerm2/Sixel).

## Approach

1. **Nettoyage et refonte de l'API (`termstrap.go`)**
   - Retirer toutes les dépendances à `glamour` et au Markdown.
   - Renommer le champ d'entrée de `Content` vers `HTML`.
   - Ajouter un champ `Stylesheets []string` au `Model` pour permettre l'injection de CSS externe.
   - Supprimer les utilitaires liés aux hacks ANSI (ex: `persistColors`, `extractSegments`).

2. **Moteur CSS et Résolution des Styles (`css.go` - Nouveau)**
   - Ajouter la dépendance `github.com/aymerick/douceur` pour parser le CSS.
   - Charger une feuille de style par défaut `bootstrap.css` (contenant les règles pour `.row`, `.col-*`, `.m-*`, `.p-*`, `.bg-*`, `.text-*`, `.border`).
   - Créer une fonction `ComputeStyles(doc *goquery.Document, stylesheets []string)` qui évalue les sélecteurs CSS (via `cascadia` ou `goquery.Is`) et attache un objet `ComputedStyle` (Marges, Paddings, Couleurs, Display) à chaque noeud de l'arbre.

3. **Layout Engine : Modèle de Boîte (`layout.go` - Nouveau)**
   - Convertir le DOM `goquery` en un arbre de rendu (`RenderTree` de type `Node`).
   - **Passe Descendante (Width)** : Le parent transmet sa largeur disponible. 
     - `display: block` prend 100% de la largeur du parent (moins les marges).
     - `.row` (flexbox simplifié) distribue la largeur à ses `.col-*` (calcul fractionnaire sur 12 colonnes).
   - **Passe Ascendante (Height & Text)** :
     - Utiliser `github.com/muesli/reflow/wordwrap` pour wrapper les noeuds textuels à la largeur exacte de leur conteneur (ComputedWidth - Paddings).
     - La hauteur du conteneur parent s'adapte à la somme de ses enfants.

4. **Pipeline de Rendu ANSI (`render.go` - Refonte)**
   - Parcourir le `RenderTree` de bas en haut.
   - Convertir chaque `ComputedStyle` en un `lipgloss.Style`.
   - `display: block` : Joindre les enfants avec `lipgloss.JoinVertical`.
   - `display: flex` (ex: `.row`) : Joindre les enfants avec `lipgloss.JoinHorizontal`.
   - Appliquer le style `lipgloss.Style.Render(...)` sur le bloc joint pour générer les bordures, couleurs et marges.
   - **Images (`<img>`)** : Extraire les attributs, calculer la largeur maximale à partir du `LayoutBox` du parent, générer le placeholder vide. L'offset `Row/Col` pour l'overlay `deferredImage` est calculé en accumulant les hauteurs et largeurs des noeuds précédents dans l'arbre.

5. **Nettoyage du code obsolète**
   - Supprimer `markdown.go`, `html.go` (extractSegments), `classes.go` (remplacé par CSS), et `grid.go` (le système de grille devient un simple comportement CSS Flexbox dans le Layout Engine).

## Critical files & anchors
- `termstrap.go` : Point d'entrée de l'API. L'assemblage global HTML -> DOM -> CSS -> Layout -> Render s'y fera.
- `layout.go` (à créer) : C'est le cœur du système. Contient les structures `Node`, `Box` et les fonctions récursives de calcul géométrique.
- `render.go` : Le remplacement de l'actuel `renderer.go`. Ne manipule plus de chaînes ANSI brutes ou de "lignes", mais assemble exclusivement des blocs `lipgloss` pré-calculés.
- `go.mod` : Ajouter `github.com/aymerick/douceur` et `github.com/muesli/reflow`.

## Verification
- **Test géométrique global** : Créer un layout avec un `div` englobant `.p-2 .m-3 .bg-dark`, contenant un `.row` et deux `.col-6`. Vérifier via une assertion que la largeur totale ne dépasse pas la limite du terminal, et que les caractères de fond ANSI sont bien appliqués jusqu'aux bords du padding interne, mais pas dans la marge.
- **Test Overlays d'Images** : Assurer que les coordonnées (Row/Col) d'une image `<img class="m-2">` sont correctement décalées par sa marge et par le padding de ses parents.
- **Validation Build** : Tous les utilitaires de test doivent utiliser des string HTML bruts au lieu du markdown existant.

## Assumptions & contingencies
- **Support CSS** : Seul un sous-ensemble critique des propriétés CSS sera supporté (Marges, Padding, Couleurs bg/fg, Text-Align, Bordures, Display block/inline/flex). Si l'utilisateur fournit des styles complexes (ex: `position: absolute` ou `float`), ils seront ignorés en fallbackant sur le comportement de bloc standard.
- **Performance** : Si le parsing CSS à la volée est trop lent, nous pré-compilerons la stylesheet par défaut de Bootstrap en objets `ComputedStyle` mis en cache par nom de classe.