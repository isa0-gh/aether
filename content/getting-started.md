---
title: "Getting started"
date: "2026-08-04"
tags: "docs"
preview: Clone the repo, edit config.json, add posts, and build in one command.
---

## Requirements

- Go 1.21+
- A GitHub repository (for deployment)

## Setup

Clone and enter the project:

```sh
git clone https://github.com/isa0-gh/aether my-site
cd my-site
```

Edit `config.json`:

```json
{
  "siteTitle": "My Site",
  "siteTitleSuffix": "",
  "siteDescription": "Short description shown in the header",
  "github": "your-github-username",
  "navLinks": [
    { "label": "Blog",  "url": "#" },
    { "label": "About", "url": "#" }
  ]
}
```

| Field | Description |
|---|---|
| `siteTitle` | Main heading and browser tab title |
| `siteTitleSuffix` | Optional dimmed suffix next to the title |
| `siteDescription` | Subtitle text in the header |
| `github` | Your GitHub username — avatar fetched at build time |
| `navLinks` | Array of `{ label, url }` navbar entries |

## Build

```sh
go run ./cmd/aether/main.go
```

Output lands in `public/`. Serve it locally:

```sh
cd public && python3 -m http.server
```

### CLI flags

```
-config     Path to config.json         (default: config.json)
-content    Directory of .md posts      (default: content)
-templates  Directory of HTML templates (default: templates)
-static     Directory of static assets  (default: static)
-output     Output directory            (default: public)
```
