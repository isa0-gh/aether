// Package feed loads all markdown posts from the content directory.
package feed

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/isa0-gh/aether/internal/markdown"
)

// Load reads every .md file in contentDir and returns posts sorted newest first.
func Load(contentDir string) ([]markdown.Post, error) {
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		return nil, err
	}

	var posts []markdown.Post
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		path := filepath.Join(contentDir, e.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		slug := strings.TrimSuffix(e.Name(), ".md")
		post := markdown.Parse(slug, string(raw))
		posts = append(posts, post)
	}

	// Sort newest first (ISO date strings compare lexicographically)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date > posts[j].Date
	})

	return posts, nil
}
