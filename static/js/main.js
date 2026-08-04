// Parse YAML-like frontmatter from a markdown string.
// Returns { meta: {}, body: '' }
function parseFrontmatter(text) {
    const match = text.match(/^---\s*\n([\s\S]*?)\n---\s*\n?([\s\S]*)$/);
    if (!match) return { meta: {}, body: text };

    const meta = {};
    match[1].split(/\n/).forEach(line => {
        const sep = line.indexOf(':');
        if (sep === -1) return;
        const key = line.slice(0, sep).trim();
        const val = line.slice(sep + 1).trim();
        meta[key] = val.replace(/^['"]|['"]$/g, '');
    });

    return { meta, body: match[2] || '' };
}

// Human-readable relative time from an ISO date string.
function getRelativeTime(value) {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;

    const diffSec  = Math.floor((Date.now() - date) / 1000);
    const diffMin  = Math.floor(diffSec  / 60);
    const diffHour = Math.floor(diffMin  / 60);
    const diffDay  = Math.floor(diffHour / 24);

    if (diffMin  <  1) return 'just now';
    if (diffMin  < 60) return `${diffMin} minute${diffMin   === 1 ? '' : 's'} ago`;
    if (diffHour < 24) return `${diffHour} hour${diffHour   === 1 ? '' : 's'} ago`;
    if (diffDay  < 30) return `${diffDay} day${diffDay      === 1 ? '' : 's'} ago`;

    const diffMonth = Math.floor(diffDay / 30);
    if (diffMonth < 12) return `${diffMonth} month${diffMonth === 1 ? '' : 's'} ago`;

    const diffYear = Math.floor(diffMonth / 12);
    return `${diffYear} year${diffYear === 1 ? '' : 's'} ago`;
}

// Resolve an image filename from frontmatter to its public path.
function imgSrc(filename) {
    if (!filename) return '';
    // Already a full URL or absolute path — use as-is
    if (filename.startsWith('http') || filename.startsWith('/')) return filename;
    return `images/${filename}`;
}
function renderPost(post, slug) {
    const app = document.getElementById('app');
    const rendered = post.body ? marked.parse(post.body.trim()) : '';

    const metaItems = [];
    if (post.date) metaItems.push(`<span class="meta-chip">${getRelativeTime(post.date)}</span>`);
    if (post.tags) metaItems.push(`<span class="meta-chip">${post.tags}</span>`);
    const metaMarkup = metaItems.length
        ? `<div class="meta-row">${metaItems.join('')}</div>`
        : '';

    app.innerHTML = `
        <div class="post-view">
            <a class="back-link" onclick="showFeed(); return false;" href="#"><img src="images/back.svg" alt="Back" class="back-icon"></a>
            ${post.image ? `<div class="news-thumbnail-frame" style="margin-bottom:16px"><img src="${imgSrc(post.image)}" class="news-thumbnail" alt="${post.title}"></div>` : ''}
            <h1>${post.title || 'Untitled'}</h1>
            ${metaMarkup}
            <div class="content-body">${rendered}</div>
        </div>
    `;

    // Update URL without reload so the back button works.
    history.pushState({ slug }, '', `#${slug}`);
}

// Render the feed list inside #app.
function renderFeed(posts) {
    const app = document.getElementById('app');

    if (!posts.length) {
        app.innerHTML = '<div class="feed"><div class="news-item"><p>No posts found.</p></div></div>';
        return;
    }

    const items = posts.map(post => {
        const title   = (post.title || 'Untitled').replace(/\s+/g, ' ').trim();
        const metaItems = [];
        if (post.date) metaItems.push(`<span class="meta-chip">${getRelativeTime(post.date)}</span>`);
        if (post.tags) metaItems.push(`<span class="meta-chip">${post.tags}</span>`);
        const metaMarkup = metaItems.length
            ? `<div class="meta-row">${metaItems.join('')}</div>`
            : '';

        return `
            <div class="news-item">
                <h2><a href="#${post.slug}" onclick="openPost('${post.slug}'); return false;"
                   style="text-decoration:none; color:inherit;">${title}</a></h2>
                ${post.image ? `<div class="news-thumbnail-frame"><img src="${imgSrc(post.image)}" class="news-thumbnail" alt="${title}"></div>` : ''}
                ${metaMarkup}
                ${post.preview ? `<p>${post.preview}</p>` : ''}
                <hr class="item-divider">
            </div>
        `;
    });

    app.innerHTML = `<div class="feed">${items.join('')}</div>`;
    history.pushState({}, '', '#');
}

// ── State ──────────────────────────────────────────────────────────────────

let _posts = [];  // cached after first load

async function loadPosts() {
    if (_posts.length) return _posts;

    // The generator writes posts.json listing all content files.
    const index = await fetch('posts.json').then(r => r.json());

    const loaded = await Promise.all(index.map(async slug => {
        const res  = await fetch(`posts/${slug}.md`);
        const text = await res.text();
        const { meta, body } = parseFrontmatter(text);
        return {
            slug,
            title:   meta.title   || 'Untitled',
            image:   meta.image   || '',
            date:    meta.date    || '',
            tags:    meta.tags    || '',
            preview: meta.preview || '',
            body,
        };
    }));

    // Newest first
    _posts = loaded.sort((a, b) => (b.date || '').localeCompare(a.date || ''));
    return _posts;
}

async function showFeed() {
    const posts = await loadPosts();
    renderFeed(posts);
}

async function openPost(slug) {
    const posts = await loadPosts();
    const post  = posts.find(p => p.slug === slug);
    if (post) renderPost(post, slug);
}

// ── Bootstrap ──────────────────────────────────────────────────────────────

async function boot() {
    const hash = location.hash.slice(1);
    if (hash) {
        await openPost(hash);
    } else {
        await showFeed();
    }
}

window.addEventListener('popstate', async e => {
    if (e.state && e.state.slug) {
        await openPost(e.state.slug);
    } else {
        await showFeed();
    }
});

boot();
