//go:build !windows

package image

import (
	"os"

	"golang.org/x/sys/unix"
)

// terminalDimensions queries the terminal window and pixel dimensions using TIOCGWINSZ ioctl.
// Returns columns, rows, terminal width in pixels, terminal height in pixels,
// cell width in pixels, and cell height in pixels.
func terminalDimensions() (cols, rows, termW, termH, cellW, cellH int) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil || ws == nil {
		cols, rows = terminalSizeFallback()
		return cols, rows, 0, 0, 0, 0
	}

	cols = int(ws.Col)
	rows = int(ws.Row)
	if cols <= 0 || rows <= 0 {
		cols, rows = terminalSizeFallback()
	}

	termW = int(ws.Xpixel)
	termH = int(ws.Ypixel)

	if termW > 0 && cols > 0 {
		cellW = termW / cols
	}
	if termH > 0 && rows > 0 {
		cellH = termH / rows
	}

	return cols, rows, termW, termH, cellW, cellH
}
