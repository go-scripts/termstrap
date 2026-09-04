// Package main demonstrates all features of termstrap:
// markdown rendering, Bootstrap-like HTML layouts, images,
// styling classes, and terminal capability auto-detection.
//
// Usage:
//
//	go run ./examples/
package main

import (
	"fmt"
	"os"

	"github.com/go-scripts/termstrap"
	termimage "github.com/go-scripts/termstrap/image"
	"golang.org/x/term"
)

func main() {
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	caps := termimage.Detect()
	fmt.Printf("Image protocol: %s | TrueColor: %v | Terminal: %dx%d\n\n",
		caps.Protocol, caps.TrueColor, width, 24)

	content := `<h1>Termstrap Demo</h1>

<p>This is a pure <b>HTML/CSS layout</b> rendered directly to ANSI terminal output.</p>
<ul>
  <li>Supports <i>italics</i>, <b>bold</b>, and <code>inline code</code></li>
  <li>Lists, headers, links, and more</li>
</ul>

<hr />

<h2>1. Film Card — Grid + Image + Table</h2>

<div class="row">
  <div class="col-md-4 border rounded p-1">
    <div><img src="https://zimage.cc/uploads/screen/d0a8490a3e1217e5f3d53f92780465955cd0b65c.webp" alt="poster" /></div>
  </div>
  <div class="col-md-8 ps-2">
    <h3>Gladiator II</h3>
    <p><b>Ridley Scott</b> — 2024</p>
    <table>
      <tr><th>Info</th><th>Valeur</th></tr>
      <tr><td>Année</td><td>2024</td></tr>
      <tr><td>Genre</td><td>Action, Drame</td></tr>
      <tr><td>Durée</td><td>2h28</td></tr>
      <tr><td>Qualité</td><td>MULTI HDLight</td></tr>
      <tr><td>Langue</td><td>FR / EN</td></tr>
    </table>
    <blockquote>"Those who are about to die, salute you."</blockquote>
  </div>
</div>

<hr />

<h2>2. Colored Alerts — Background &amp; Text Colors</h2>

<div class="row">
  <div class="col-md-4 bg-success text-white p-2 rounded">
    <div><b>Succès !</b> Le téléchargement est terminé.</div>
  </div>
  <div class="col-md-4 bg-warning text-dark p-2 rounded">
    <div><b>Attention !</b> L'espace disque est presque plein.</div>
  </div>
  <div class="col-md-4 bg-danger text-white p-2 rounded">
    <div><b>Erreur !</b> Impossible de se connecter au serveur.</div>
  </div>
</div>

<hr />

<h2>3. Shadows — sm, normal, lg</h2>

<div class="row">
  <div class="col-md-4 border rounded shadow-sm p-2 m-1">
    <h3>Shadow SM</h3>
    <p>Petite ombre subtile.</p>
  </div>
  <div class="col-md-4 border rounded shadow p-2 m-1">
    <h3>Shadow Normal</h3>
    <p>Ombre moyenne standard.</p>
  </div>
  <div class="col-md-4 border rounded shadow-lg p-2 m-1">
    <h3>Shadow LG</h3>
    <p>Grande ombre prononcée.</p>
  </div>
</div>

<hr />

<h2>4. Text Alignment</h2>

<div class="row">
  <div class="col-md-4 border text-start p-1">
    <div><b>Aligné à gauche</b></div>
    <p>Lorem ipsum dolor sit amet.</p>
  </div>
  <div class="col-md-4 border text-center p-1">
    <div><b>Centré</b></div>
    <p>Lorem ipsum dolor sit amet.</p>
  </div>
  <div class="col-md-4 border text-end p-1">
    <div><b>Aligné à droite</b></div>
    <p>Lorem ipsum dolor sit amet.</p>
  </div>
</div>

<hr />

<h2>5. Borders Variants</h2>

<div class="row">
  <div class="col-md-3 border p-1">
    <div><b>border</b></div>
    <p>All sides</p>
  </div>
  <div class="col-md-3 border-top p-1">
    <div><b>border-top</b></div>
    <p>Top only</p>
  </div>
  <div class="col-md-3 border-bottom p-1">
    <div><b>border-bottom</b></div>
    <p>Bottom only</p>
  </div>
  <div class="col-md-3 border-left border-right p-1">
    <div><b>border-left + right</b></div>
    <p>Sides only</p>
  </div>
</div>

<hr />

<h2>6. Code &amp; Preformatted Text</h2>

<div class="row">
  <div class="col-12 p-2 bg-dark text-white rounded">
    <pre>
package main

import "github.com/go-scripts/termstrap"

func main() {
    m := termstrap.New("&lt;div class=\"p-2\"&gt;Hello&lt;/div&gt;")
    out, _ := m.Render()
    println(out)
}
    </pre>
  </div>
</div>
`

	m := termstrap.Model{
		HTML:  content,
		Width: width,
	}
	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(output)
}
