// Package builder renders the site into the output directory.
package builder

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/isa0-gh/aether/internal/config"
	"github.com/isa0-gh/aether/internal/feed"
	"github.com/isa0-gh/aether/internal/github"
	"github.com/isa0-gh/aether/internal/markdown"
)

// Options controls input/output paths.
type Options struct {
	ContentDir  string // e.g. "content"
	TemplateDir string // e.g. "templates"
	StaticDir   string // e.g. "static"
	OutputDir   string // e.g. "public"
	ConfigFile  string // e.g. "config.json"
}

// Build generates the full static site into OutputDir.
func Build(opts Options) error {
	// Load config
	cfg, err := config.Load(opts.ConfigFile)
	if err != nil {
		return err
	}

	// Resolve GitHub avatar at build time
	if cfg.GitHub != "" {
		avatarURL, err := github.AvatarURL(cfg.GitHub)
		if err != nil {
			fmt.Printf("warning: could not fetch GitHub avatar for %q: %v\n", cfg.GitHub, err)
		} else {
			cfg.AvatarURL = avatarURL
		}
	}

	// Load posts
	posts, err := feed.Load(opts.ContentDir)
	if err != nil {
		return err
	}

	// Prepare output directory
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return err
	}

	// Write index.html
	if err := writeIndex(opts, cfg); err != nil {
		return err
	}

	// Write posts.json — the manifest the JS uses to discover posts
	if err := writePostsManifest(opts.OutputDir, posts); err != nil {
		return err
	}

	// Copy markdown posts to public/posts/
	if err := copyPosts(opts.ContentDir, opts.OutputDir, posts); err != nil {
		return err
	}

	// Copy static assets (css, js, images)
	if err := copyStatic(opts.StaticDir, opts.OutputDir); err != nil {
		return err
	}

	return nil
}

// writeIndex renders templates/index.html into public/index.html.
func writeIndex(opts Options, cfg *config.Site) error {
	tmplPath := filepath.Join(opts.TemplateDir, "index.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(opts.OutputDir, "index.html"))
	if err != nil {
		return err
	}
	defer out.Close()

	return tmpl.Execute(out, cfg)
}

// writePostsManifest writes a JSON array of post slugs.
func writePostsManifest(outputDir string, posts []markdown.Post) error {
	slugs := make([]string, len(posts))
	for i, p := range posts {
		slugs[i] = p.Slug
	}

	data, err := json.Marshal(slugs)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(outputDir, "posts.json"), data, 0644)
}

// copyPosts copies .md files to public/posts/.
func copyPosts(contentDir, outputDir string, posts []markdown.Post) error {
	postsOut := filepath.Join(outputDir, "posts")
	if err := os.MkdirAll(postsOut, 0755); err != nil {
		return err
	}

	for _, p := range posts {
		src := filepath.Join(contentDir, p.Slug+".md")
		dst := filepath.Join(postsOut, p.Slug+".md")
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

// copyStatic mirrors the static/ tree into outputDir.
func copyStatic(staticDir, outputDir string) error {
	return filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(staticDir, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(outputDir, rel)

		if info.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		return copyFile(path, dest)
	})
}

// copyFile copies a single file src → dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
