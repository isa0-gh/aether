---
title: "Deploying with GitHub Pages"
date: "2026-08-02"
tags: "docs"
preview: Push to main and your site deploys automatically via the included GitHub Actions workflow.
---

The included workflow at `.github/workflows/deploy.yml` builds and deploys your site on every push to `main`.

## Enable GitHub Pages

1. Go to your repo → **Settings** → **Pages**
2. Set **Source** to **GitHub Actions**

That's it. Push to `main` and your site goes live at:

```
https://<username>.github.io/<repo>/
```

## What the workflow does

1. Checks out the repo
2. Sets up Go using the version from `go.mod`
3. Runs `go run ./cmd/aether/main.go`
4. Uploads `public/` and deploys it to GitHub Pages

You can also trigger a deploy manually from the **Actions** tab using **workflow_dispatch**.
