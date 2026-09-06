package termstrap

import (
	"time"

	termimage "github.com/go-scripts/termstrap/image"
)

// Theme represents a visual color scheme.
type Theme string

const (
	ThemeBootstrap  Theme = "bootstrap"
	ThemeTokyoNight Theme = "tokyonight"
	ThemeDracula    Theme = "dracula"
)

// New creates a new Model with the given HTML content and default settings.
func New(htmlContent string, opts ...Option) Model {
	m := Model{
		HTML: htmlContent,
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// Option configures a Model.
type Option func(*Model)

// WithWidth sets the terminal output width in characters.
func WithWidth(width int) Option {
	return func(m *Model) {
		m.Width = width
	}
}

// WithStylesheets sets custom CSS stylesheets to be merged with default Bootstrap styles.
func WithStylesheets(stylesheets ...string) Option {
	return func(m *Model) {
		m.Stylesheets = append(m.Stylesheets, stylesheets...)
	}
}

// WithRootPath sets the root directory for resolving local image files.
func WithRootPath(path string) Option {
	return func(m *Model) {
		m.RootPath = path
	}
}

// WithTheme sets the visual color theme for rendering.
func WithTheme(theme Theme) Option {
	return func(m *Model) {
		m.Theme = theme
	}
}

// WithImageRenderer sets a custom image renderer.
func WithImageRenderer(renderer termimage.Renderer) Option {
	return func(m *Model) {
		m.ImageRenderer = renderer
	}
}

// WithColorMode sets the color depth for rendering.
func WithColorMode(mode termimage.ColorMode) Option {
	return func(m *Model) {
		m.ColorMode = mode
	}
}

// WithCachePolicy sets the image caching policy.
func WithCachePolicy(policy CachePolicy) Option {
	return func(m *Model) {
		m.CachePolicy = policy
	}
}

// WithCacheTTL sets the time-to-live for cached items.
func WithCacheTTL(ttl time.Duration) Option {
	return func(m *Model) {
		m.CacheTTL = ttl
	}
}

