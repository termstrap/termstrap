// Package image provides multi-protocol terminal image rendering.
//
// It detects terminal graphics capabilities and renders images using the
// highest-fidelity protocol available, with the following fallback hierarchy:
//
//  1. Kitty Graphics Protocol (GPU-accelerated, pixel-perfect)
//  2. iTerm2 Inline Images (Base64 PNG via OSC 1337)
//  3. Sixel Graphics (for compatible terminals like foot, mlterm, xterm)
//  4. Sextants (Unicode 13.0 2x3 block sextants — 6 sub-pixels per cell)
//  5. ANSI Half-Block (Unicode ▄/▀ with TrueColor — universal fallback)
package image

import (
	"image"
	"io"
	"os"
)

// Protocol represents a terminal graphics protocol.
type Protocol int

const (
	// HalfBlock renders images using Unicode half-block characters (▄/▀)
	// with ANSI TrueColor. Works on virtually all modern terminals.
	HalfBlock Protocol = iota
	// Sixel renders images using the Sixel graphics protocol (DEC standard).
	Sixel
	// ITerm2 renders images using iTerm2's inline image protocol (OSC 1337).
	ITerm2
	// Kitty renders images using the Kitty graphics protocol (APC sequences).
	Kitty
	// Sextant renders images using Unicode 13.0 2x3 block sextants (6 sub-pixels per cell).
	Sextant
)

// String returns the human-readable name of the protocol.
func (p Protocol) String() string {
	switch p {
	case Kitty:
		return "kitty"
	case ITerm2:
		return "iterm2"
	case Sixel:
		return "sixel"
	case Sextant:
		return "sextant"
	case HalfBlock:
		return "halfblock"
	default:
		return "unknown"
	}
}

// IsGraphic reports whether the protocol renders as a pixel graphic overlay
// (Kitty, iTerm2, Sixel) rather than text character cells (HalfBlock, Sextant).
func (p Protocol) IsGraphic() bool {
	return p == Kitty || p == ITerm2 || p == Sixel
}

// Capabilities holds the detected terminal graphics capabilities.
type Capabilities struct {
	Protocol     Protocol // Best available graphics protocol
	TrueColor    bool     // Whether the terminal supports 24-bit color
	TermWidthPx  int      // Terminal width in pixels (0 if unknown)
	TermHeightPx int      // Terminal height in pixels (0 if unknown)
	CellWidthPx  int      // Single cell width in pixels (0 if unknown)
	CellHeightPx int      // Single cell height in pixels (0 if unknown)
	ColCount     int      // Terminal width in columns
	RowCount     int      // Terminal height in rows
}

// Renderer renders an image.Image as a terminal-ready string using
// a specific graphics protocol.
type Renderer interface {
	// Render converts an image to a terminal-displayable string.
	// width is the desired display width in terminal columns.
	Render(img image.Image, width int) (string, error)

	// Protocol returns which graphics protocol this renderer uses.
	Protocol() Protocol
}

// ConstrainedRenderer is optionally implemented by renderers that support
// both width and height constraints. This prevents images from overflowing
// their allocated space in bordered/padded containers.
type ConstrainedRenderer interface {
	Renderer
	// RenderConstrained renders an image constrained to fit within the given
	// width (columns) and height (rows). The image must not exceed these dimensions.
	RenderConstrained(img image.Image, width, height int) (string, error)
}

// Option configures a Renderer created by NewRenderer.
type Option func(*options)

type options struct {
	protocol     *Protocol // nil = auto-detect
	output       io.Writer // terminal output (for termenv detection)
	capabilities *Capabilities
	colorMode    ColorMode
	optimize     bool
	dither       bool
	cellWidthPx  int
	cellHeightPx int
}

// WithProtocol forces a specific graphics protocol instead of auto-detecting.
func WithProtocol(p Protocol) Option {
	return func(o *options) {
		o.protocol = &p
	}
}

// WithColorMode specifies the color depth (TrueColor, 256 colors, 16 colors) for half-block and sextant rendering.
func WithColorMode(cm ColorMode) Option {
	return func(o *options) {
		o.colorMode = cm
	}
}

// WithOptimizeSequences controls whether consecutive ANSI escape sequences with matching colors are deduplicated.
func WithOptimizeSequences(opt bool) Option {
	return func(o *options) {
		o.optimize = opt
	}
}

// WithDither enables or disables Floyd-Steinberg error diffusion dithering for quantized color modes.
func WithDither(dither bool) Option {
	return func(o *options) {
		o.dither = dither
	}
}

// WithCellSize explicitly configures the physical cell width and height in pixels (for HiDPI scaling).
func WithCellSize(cellWidthPx, cellHeightPx int) Option {
	return func(o *options) {
		o.cellWidthPx = cellWidthPx
		o.cellHeightPx = cellHeightPx
	}
}

// WithOutput sets the terminal output writer used for capability detection.
// Defaults to os.Stdout.
func WithOutput(w io.Writer) Option {
	return func(o *options) {
		o.output = w
	}
}

// WithCapabilities provides pre-detected capabilities, skipping detection.
func WithCapabilities(caps Capabilities) Option {
	return func(o *options) {
		o.capabilities = &caps
	}
}

// NewRenderer creates a Renderer using the best available graphics protocol.
// It auto-detects terminal capabilities unless overridden via options.
func NewRenderer(opts ...Option) Renderer {
	o := &options{
		output:    os.Stdout,
		colorMode: ColorModeTrueColor,
		optimize:  true,
		dither:    true,
	}
	for _, opt := range opts {
		opt(o)
	}

	var caps Capabilities
	var proto Protocol
	if o.capabilities != nil {
		caps = *o.capabilities
		proto = caps.Protocol
	} else {
		caps = Detect()
		proto = caps.Protocol
	}

	if o.protocol != nil {
		proto = *o.protocol
	}

	cellW := o.cellWidthPx
	cellH := o.cellHeightPx
	if cellW <= 0 && caps.CellWidthPx > 0 {
		cellW = caps.CellWidthPx
	}
	if cellH <= 0 && caps.CellHeightPx > 0 {
		cellH = caps.CellHeightPx
	}

	switch proto {
	case Kitty:
		return &kittyRenderer{cellWidthPx: cellW, cellHeightPx: cellH}
	case ITerm2:
		return &itermRenderer{cellWidthPx: cellW, cellHeightPx: cellH}
	case Sixel:
		return &sixelRenderer{cellWidthPx: cellW, cellHeightPx: cellH, dither: o.dither}
	case Sextant:
		return NewSextantRenderer(o.colorMode, o.optimize, o.dither)
	default:
		return NewHalfBlockRenderer(o.colorMode, o.optimize, o.dither)
	}
}
