package termstrap

import (
	"time"

	termimage "github.com/go-scripts/termstrap/image"
)

// Theme represents a visual color scheme.
type Theme string

const (
	ThemeNone       Theme = "none"
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

// WithStylesheets adds custom CSS merged on top of the Bootstrap/theme defaults.
// Each argument is auto-detected: raw CSS (contains "{"/"}" or newlines) is used
// as-is; an http(s):// URL is fetched; anything else is treated as a file path
// (relative paths resolve against WithRootPath). Unreadable paths/URLs are
// skipped silently. Resolution happens at Render time, so option order with
// WithRootPath does not matter.
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

// WithDisableImages controls whether images are replaced by alt text.
func WithDisableImages(disable bool) Option {
	return func(m *Model) {
		m.DisableImages = disable
	}
}

// WithDisableHyperlinks disables emitting OSC 8 ANSI hyperlinks for <a> tags.
func WithDisableHyperlinks(disable bool) Option {
	return func(m *Model) {
		m.DisableHyperlinks = disable
	}
}

// WithSyntaxHighlight toggles chroma syntax colorization of <pre> code blocks.
// Highlighting is enabled by default; pass false to disable it.
func WithSyntaxHighlight(enable bool) Option {
	return func(m *Model) {
		m.DisableSyntaxHighlight = !enable
	}
}

// WithSyntaxTheme sets the chroma highlighting style used for code blocks
// (e.g. "monokai", "dracula", "nord", "github-dark"). Empty derives a style
// from the active visual Theme.
func WithSyntaxTheme(name string) Option {
	return func(m *Model) {
		m.SyntaxTheme = name
	}
}

// WithStep sets the animation frame step index (for <spinner>, <shimmer>, .spinner, .shimmer).
func WithStep(step int) Option {
	return func(m *Model) {
		m.Step = step
	}
}
