// ── Search ─────────────────────────────────────────────────────────────────
//
// Loads search.json (pre-compiled at build time) and filters it client-side.
// No dependencies. Scores matches by field weight and highlights query terms.

let _searchIndex = null;

async function loadSearchIndex() {
    if (_searchIndex) return _searchIndex;
    const res = await fetch('search.json');
    _searchIndex = await res.json();
    return _searchIndex;
}

// Score a single entry against a query string.
// Returns 0 if no match, higher = better match.
function scoreEntry(entry, terms) {
    let score = 0;
    const fields = [
        { text: (entry.title   || '').toLowerCase(), weight: 10 },
        { text: (entry.tags    || '').toLowerCase(), weight:  6 },
        { text: (entry.preview || '').toLowerCase(), weight:  4 },
        { text: (entry.body    || '').toLowerCase(), weight:  1 },
    ];

    for (const term of terms) {
        let hit = false;
        for (const { text, weight } of fields) {
            if (text.includes(term)) {
                score += weight;
                hit = true;
            }
        }
        // All terms must match at least one field
        if (!hit) return 0;
    }
    return score;
}

// Wrap matched terms in <mark> within a plain text string.
function highlight(text, terms) {
    if (!text) return '';
    let result = text;
    for (const term of terms) {
        const re = new RegExp(`(${term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
        result = result.replace(re, '<mark>$1</mark>');
    }
    return result;
}

// Run a search and render results into #app.
async function runSearch(query) {
    const app   = document.getElementById('app');
    const index = await loadSearchIndex();
    const terms = query.toLowerCase().trim().split(/\s+/).filter(Boolean);

    if (!terms.length) {
        await showFeed();
        return;
    }

    const scored = index
        .map(entry => ({ entry, score: scoreEntry(entry, terms) }))
        .filter(r => r.score > 0)
        .sort((a, b) => b.score - a.score);

    if (!scored.length) {
        app.innerHTML = `<div class="feed"><div class="news-item"><p>No results for <strong>${query}</strong>.</p></div></div>`;
        return;
    }

    const items = scored.map(({ entry }) => {
        const title   = highlight(entry.title || 'Untitled', terms);
        const preview = highlight(entry.preview || '', terms);
        const tags    = highlight(entry.tags || '', terms);

        const metaItems = [];
        if (entry.date) metaItems.push(`<span class="meta-chip">${getRelativeTime(entry.date)}</span>`);
        if (tags)       metaItems.push(`<span class="meta-chip">${tags}</span>`);
        const metaMarkup = metaItems.length
            ? `<div class="meta-row">${metaItems.join('')}</div>`
            : '';

        return `
            <div class="news-item">
                <h2><a href="#${entry.slug}" onclick="openPost('${entry.slug}'); return false;"
                   style="text-decoration:none; color:inherit;">${title}</a></h2>
                ${metaMarkup}
                ${preview ? `<p>${preview}</p>` : ''}
                <hr class="item-divider">
            </div>
        `;
    });

    app.innerHTML = `
        <div class="search-meta">
            ${scored.length} result${scored.length === 1 ? '' : 's'} for <strong>${query}</strong>
        </div>
        <div class="feed">${items.join('')}</div>
    `;
}

// Mount the search bar into the navbar.
function mountSearchBar() {
    const navbar = document.querySelector('.navbar');
    if (!navbar) return;

    const form = document.createElement('form');
    form.className  = 'search-form';
    form.setAttribute('role', 'search');
    form.innerHTML  = `<input class="search-input" type="search" placeholder="Search…" aria-label="Search posts">`;

    // Insert before the <ul>
    const ul = navbar.querySelector('ul');
    navbar.insertBefore(form, ul);

    const input = form.querySelector('input');

    let debounce;
    input.addEventListener('input', () => {
        clearTimeout(debounce);
        const q = input.value.trim();
        debounce = setTimeout(() => {
            if (q) {
                history.pushState({ search: q }, '', `?q=${encodeURIComponent(q)}`);
                runSearch(q);
            } else {
                history.pushState({}, '', location.pathname);
                showFeed();
            }
        }, 180);
    });

    form.addEventListener('submit', e => e.preventDefault());

    // Restore query from URL on load
    const params = new URLSearchParams(location.search);
    const q = params.get('q');
    if (q) {
        input.value = q;
        runSearch(q);
    }
}
