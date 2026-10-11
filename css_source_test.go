package termstrap

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWithStylesheets_Autodetect verifies that each WithStylesheets argument is
// classified and loaded correctly: raw CSS, local file path, or http(s) URL,
// with unreadable sources skipped.
func TestWithStylesheets_Autodetect(t *testing.T) {
	dir := t.TempDir()
	rel := filepath.Join(dir, "f.css")
	if err := os.WriteFile(rel, []byte(".fromfile{color:#00ff00;}"), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		_, _ = w.Write([]byte(".fromurl{font-style:italic;}"))
	}))
	defer srv.Close()

	cases := []struct {
		in   string
		want string
	}{
		{".raw{font-weight:bold;}", ".raw{font-weight:bold;}"},  // raw (braces)
		{":root{\n--x:1;\n}", ":root{\n--x:1;\n}"},              // raw (multiline)
		{"f.css", ".fromfile{color:#00ff00;}"},                 // relative path via RootPath
		{rel, ".fromfile{color:#00ff00;}"},                     // absolute path
		{srv.URL + "/x.css", ".fromurl{font-style:italic;}"},   // http URL
		{"does-not-exist.css", ""},                             // unreadable -> dropped
		{"https://127.0.0.1:1/nope.css", ""},                   // unreachable URL -> dropped
	}
	for _, c := range cases {
		if got := resolveStylesheetArg(c.in, dir); got != c.want {
			t.Errorf("resolveStylesheetArg(%q) = %q, want %q", c.in, got, c.want)
		}
	}

	// End-to-end: a path-based stylesheet must actually style the rendered output.
	html := `<div class="fromfile"><span>Styled</span></div>`
	out, err := New(html, WithWidth(30), WithRootPath(dir), WithStylesheets("f.css")).Render()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Styled") || !strings.Contains(out, "\x1b[") {
		t.Fatalf("expected styled 'Styled' output, got %q", out)
	}
}
