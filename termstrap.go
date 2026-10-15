package termstrap

import (
	"time"

	termimage "github.com/termstrap/termstrap/image"
)

// Theme represents a visual color scheme.
type Theme string

const (
	ThemeNone       Theme = "none"
	ThemeBootstrap  Theme = "bootstrap"
	ThemeTokyoNight Theme = "tokyonight"
	ThemeDracula    Theme = "dracula"
	ThemeGlamour    Theme = "glamour"
)

// LinkMode defines how hyperlinks are presented in the terminal output.
type LinkMode string

const (
	// LinkModeOSC8 emits native ANSI OSC 8 clickable terminal hyperlinks
	// on the anchor text without displaying the raw URL (default).
	LinkModeOSC8 LinkMode = "osc8"

	// LinkModeInline appends the raw URL directly after the link text (e.g. "Text https://..."),
	// matching Glamour / Glow CLI inline link behavior.
	LinkModeInline LinkMode = "inline"

	// LinkModeFootnotes references links via bracketed numbers (e.g. "Text [1]")
	// and appends a footnotes bibliography section at the bottom of the output.
	LinkModeFootnotes LinkMode = "footnotes"
)

// ImageMode defines how <img> elements are rendered.
type ImageMode string

const (
	// ImageModeAuto uses TermStrap's multi-protocol graphics engine
	// (Kitty, iTerm2, Sixel, Half-Block Unicode) (default).
	ImageModeAuto ImageMode = "auto"

	// ImageModeGlamour replaces all images with Glamour-style text placeholders:
	// "Image: [alt] →" followed by the source URL in underlined pink.
	ImageModeGlamour ImageMode = "glamour"

	// ImageModePlaceholder replaces images with simple "[alt](src)" markdown text.
	ImageModePlaceholder ImageMode = "placeholder"

	// ImageModeSmart renders images graphically when legible, but intelligently
	// falls back to Glamour-style placeholders when graphics would be degraded:
	// content width too narrow (< MinImageWidth cols), native pixels too tiny (< 32px),
	// extreme aspect ratios (> 5:1 or < 1:4), vector SVG, or network/404 load errors.
	ImageModeSmart ImageMode = "smart"
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

// WithLinkMode sets the presentation strategy for hyperlinks (osc8, inline, footnotes).
func WithLinkMode(mode LinkMode) Option {
	return func(m *Model) {
		m.LinkMode = mode
	}
}

// WithImageMode sets the presentation strategy for images (auto graphics, glamour placeholder, markdown placeholder).
func WithImageMode(mode ImageMode) Option {
	return func(m *Model) {
		m.ImageMode = mode
	}
}

// WithGlamourPreset configures TermStrap with the Glamour visual theme and Glamour's
// textual strategies for links and image placeholders (emulating Glow/Glamour CLI behavior).
func WithGlamourPreset() Option {
	return func(m *Model) {
		m.Theme = ThemeGlamour
		m.LinkMode = LinkModeInline
		m.ImageMode = ImageModeGlamour
	}
}

// WithMinImageWidth sets the minimum column width below which an image automatically
// falls back to a Glamour-style placeholder when ImageModeSmart is active (default: 12).
func WithMinImageWidth(minCols int) Option {
	return func(m *Model) {
		m.MinImageWidth = minCols
	}
}

// WithSmartImages enables or disables smart image fallback (ImageModeSmart).
func WithSmartImages(enable bool) Option {
	return func(m *Model) {
		if enable {
			m.ImageMode = ImageModeSmart
		} else {
			m.ImageMode = ImageModeAuto
		}
	}
}
