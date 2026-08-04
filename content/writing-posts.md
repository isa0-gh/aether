---
title: "Writing posts"
date: "2026-08-03"
tags: "docs"
preview: Create a markdown file in content/, add frontmatter, and it appears in the feed automatically.
---

Create a `.md` file in `content/`:

```
content/my-first-post.md
```

```markdown
---
title: "My first post"
date: "2026-08-04"
tags: "go, web"
image: cover.png
preview: A short sentence shown in the feed.
---

Full post body goes here. **Markdown** is supported.
```

## Frontmatter fields

| Field | Required | Description |
|---|---|---|
| `title` | yes | Post title |
| `date` | yes | ISO date `YYYY-MM-DD`, used for sorting newest first |
| `tags` | no | Comma-separated tags shown as chips |
| `image` | no | Cover image filename — see Images |
| `preview` | no | Summary shown in the feed list |

## Images

Put image files in `static/images/` and reference them by filename only:

```
static/images/cover.png
```

```markdown
image: cover.png
```

Full URLs (`https://...`) are passed through as-is.
