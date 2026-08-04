// Package markdown parses frontmatter and body from .md files.
package markdown

import (
	"bufio"
	"strings"
)

// Post represents a single parsed markdown file.
type Post struct {
	Slug    string
	Title   string
	Image   string
	Date    string
	Tags    string
	Preview string
	Body    string
}

// Parse extracts YAML-like frontmatter and body from raw markdown text.
// Frontmatter must be delimited by --- lines at the top of the file.
func Parse(slug, raw string) Post {
	post := Post{Slug: slug}

	scanner := bufio.NewScanner(strings.NewReader(raw))

	// Expect opening ---
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		post.Body = raw
		return post
	}

	// Read frontmatter lines until closing ---
	meta := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, `"'`)
		meta[key] = val
	}

	// Remaining lines are the body
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}
	post.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))

	post.Title   = meta["title"]
	post.Image   = meta["image"]
	post.Date    = meta["date"]
	post.Tags    = meta["tags"]
	post.Preview = meta["preview"]

	return post
}
