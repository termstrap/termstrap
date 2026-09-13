package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/go-scripts/termstrap"
)

func main() {
	jsonPath := "../torrent-aggregator/data/v3/catalog/00094B8214D3EFB1E6EF32D6C5B56067E2B718DA.raw.json"
	if len(os.Args) > 1 {
		jsonPath = os.Args[1]
	}

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading json file %q: %v\n", jsonPath, err)
		os.Exit(1)
	}

	var torrent map[string]interface{}
	if err := json.Unmarshal(data, &torrent); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing json: %v\n", err)
		os.Exit(1)
	}

	tmplPath := filepath.Join("examples", "torrent", "torrent_detail.html")
	if _, err := os.Stat(tmplPath); err != nil {
		tmplPath = "torrent_detail.html"
	}
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading template %q: %v\n", tmplPath, err)
		os.Exit(1)
	}

	tmpl, err := template.New("detail").Parse(string(tmplContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing html template: %v\n", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct {
		Title         string
		Category      string
		SubCategory   string
		Rating        float64
		PosterURL     string
		Seeders       int
		Leechers      int
		FormattedSize string
		Genres        []string
		Description   string
		InfoHash      string
		SourceID      string
		DetailsURL    string
		MagnetURI     string
		PublishDate   interface{}
	}{
		Title:         getString(torrent, "title"),
		Category:      getString(torrent, "category"),
		SubCategory:   getString(torrent, "subcategory"),
		Rating:        getFloat(torrent, "rating"),
		PosterURL:     getString(torrent, "poster_url"),
		Seeders:       getInt(torrent, "seeders"),
		Leechers:      getInt(torrent, "leechers"),
		FormattedSize: getString(torrent, "formatted_size"),
		Genres:        getStringSlice(torrent, "genres"),
		Description:   getString(torrent, "description"),
		InfoHash:      getString(torrent, "info_hash"),
		SourceID:      getString(torrent, "source_id"),
		DetailsURL:    getString(torrent, "details_url"),
		MagnetURI:     getString(torrent, "magnet_uri"),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	// Render through Termstrap ANSI Engine
	// Render through Termstrap ANSI Engine with TokyoNight Theme
	m := termstrap.New(buf.String(),
		termstrap.WithWidth(110),
		termstrap.WithTheme(termstrap.ThemeTokyoNight),
	)

	output, err := m.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering with termstrap: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(output)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key].([]interface{}); ok {
		var res []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				res = append(res, s)
			}
		}
		return res
	}
	return nil
}
