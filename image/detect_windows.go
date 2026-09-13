//go:build windows

package image

// terminalDimensions queries the terminal window dimensions on Windows.
func terminalDimensions() (cols, rows, termW, termH, cellW, cellH int) {
	cols, rows = terminalSizeFallback()
	return cols, rows, 0, 0, 0, 0
}
