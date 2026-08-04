// Package builder renders the site into the output directory.
package builder

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// Write robots.txt
	if err := writeRobots(opts, cfg); err != nil {
		return err
	}

	// Write posts.json — the manifest the JS uses to discover posts
	if err := writePostsManifest(opts.OutputDir, posts); err != nil {
		return err
	}

	// Write search.json — pre-compiled search index
	if err := writeSearchIndex(opts.OutputDir, posts); err != nil {
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

	// Generate sitemap.xml and rss.xml
	if cfg.BaseURL != "" {
		if err := writeSitemap(opts.OutputDir, cfg.BaseURL, posts); err != nil {
			return err
		}
		if err := writeRSS(opts.OutputDir, cfg, posts); err != nil {
			return err
		}
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

// writeRobots renders templates/robots.txt into public/robots.txt.
func writeRobots(opts Options, cfg *config.Site) error {
	tmplPath := filepath.Join(opts.TemplateDir, "robots.txt")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(opts.OutputDir, "robots.txt"))
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

// searchEntry is one record in the search index.
type searchEntry struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Tags    string `json:"tags"`
	Preview string `json:"preview"`
	Date    string `json:"date"`
	// Body is included so full-text search works client-side.
	Body string `json:"body"`
}

// writeSearchIndex writes a JSON array of searchable post fields.
func writeSearchIndex(outputDir string, posts []markdown.Post) error {
	entries := make([]searchEntry, len(posts))
	for i, p := range posts {
		entries[i] = searchEntry{
			Slug:    p.Slug,
			Title:   p.Title,
			Tags:    p.Tags,
			Preview: p.Preview,
			Date:    p.Date,
			Body:    p.Body,
		}
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(outputDir, "search.json"), data, 0644)
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

// ── Sitemap ────────────────────────────────────────────────────────────────

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

func writeSitemap(outputDir, baseURL string, posts []markdown.Post) error {
	base := strings.TrimRight(baseURL, "/")

	urls := []sitemapURL{{Loc: base + "/"}}
	for _, p := range posts {
		u := sitemapURL{Loc: base + "/#" + p.Slug}
		if p.Date != "" {
			u.LastMod = p.Date
		}
		urls = append(urls, u)
	}

	sm := sitemap{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	data, err := xml.MarshalIndent(sm, "", "  ")
	if err != nil {
		return err
	}

	out := append([]byte(xml.Header), data...)
	return os.WriteFile(filepath.Join(outputDir, "sitemap.xml"), out, 0644)
}

// ── RSS ────────────────────────────────────────────────────────────────────

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate,omitempty"`
	GUID        string `xml:"guid"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	LastBuild   string    `xml:"lastBuildDate"`
	Items       []rssItem `xml:"item"`
}

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

func writeRSS(outputDir string, cfg *config.Site, posts []markdown.Post) error {
	base := strings.TrimRight(cfg.BaseURL, "/")

	items := make([]rssItem, 0, len(posts))
	for _, p := range posts {
		link := base + "/#" + p.Slug
		item := rssItem{
			Title:       p.Title,
			Link:        link,
			GUID:        link,
			Description: p.Preview,
		}
		if p.Date != "" {
			if t, err := time.Parse("2006-01-02", p.Date); err == nil {
				item.PubDate = t.UTC().Format(time.RFC1123Z)
			}
		}
		items = append(items, item)
	}

	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Title:       cfg.SiteTitle,
			Link:        base + "/",
			Description: cfg.SiteDescription,
			Language:    "en",
			LastBuild:   time.Now().UTC().Format(time.RFC1123Z),
			Items:       items,
		},
	}

	data, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return err
	}

	out := append([]byte(xml.Header), data...)
	return os.WriteFile(filepath.Join(outputDir, "rss.xml"), out, 0644)
}
