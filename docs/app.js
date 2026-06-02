/* ─────────────────────────────────────────────────────────────
   traCtl Developer Docs — Shared Site Logic
   ───────────────────────────────────────────────────────────── */

// ── Navigation tree ──────────────────────────────────────────
const NAV_ITEMS = [
  { type: 'group', label: 'GET STARTED' },
  { label: 'Overview',       icon: '📦', href: 'index.html' },
  { label: 'Architecture',   icon: '🏗️', href: 'pages/architecture.html' },
  { label: 'Schema & DTOs',  icon: '🔷', href: 'pages/schema-and-dto.html' },

  { type: 'group', label: 'ENGINE MODULES' },
  { label: 'Pipeline',    icon: '⚡', href: 'pages/module.html?id=pipeline' },   // orchestrator
  { label: 'Parser',      icon: '📝', href: 'pages/module.html?id=parser' },      // Stage 1
  { label: 'Overlay',     icon: '🔧', href: 'pages/module.html?id=overlay' },     // Stage 2
  { label: 'Validation',  icon: '🛡️', href: 'pages/module.html?id=validation' }, // Stage 3
  { label: 'Planner',     icon: '🗺️', href: 'pages/module.html?id=planner' },    // Stage 4
  { label: 'Compiler',    icon: '🔨', href: 'pages/module.html?id=compiler' },    // Stage 5
  { label: 'Scheduler',   icon: '🔄', href: 'pages/module.html?id=scheduler' },   // Stage 6
  { label: 'Executor',    icon: '🚀', href: 'pages/module.html?id=executor' },    // inside Stage 6
  { label: 'Runtime',     icon: '🧠', href: 'pages/module.html?id=runtime' },     // state during exec
  { label: 'Assertion',   icon: '✅', href: 'pages/module.html?id=assertion' },   // post-step
  { label: 'Extract',     icon: '🔍', href: 'pages/module.html?id=extract' },     // post-step
  { label: 'Sandbox (JS)',icon: '🏖️', href: 'pages/module.html?id=sandbox' },    // hooks
  { label: 'Diagnostics', icon: '📊', href: 'pages/module.html?id=diagnostics' }, // observability

  { type: 'group', label: 'SURFACES' },
  { label: 'Overview',   icon: '📡', href: 'pages/surface.html?id=overview' },
  { label: 'CLI',        icon: '💻', href: 'pages/surface.html?id=cli' },
  { label: 'Web + WASM', icon: '🌐', href: 'pages/surface.html?id=web-wasm' },
  { label: 'Desktop',    icon: '🖥️', href: 'pages/surface.html?id=desktop' },
  { label: 'Local API',  icon: '🔌', href: 'pages/surface.html?id=localapi' },

  { type: 'group', label: 'FRONTEND' },
  { label: 'React UI',          icon: '⚛️', href: 'pages/frontend/index.html' },
  { label: 'Platform Adapter',  icon: '🔀', href: 'pages/frontend/platform-adapter.html' },

  { type: 'group', label: 'CONTRIBUTING' },
  { label: 'Guidelines',     icon: '📋', href: 'pages/contributing/index.html' },
  { label: 'Code Standards', icon: '✏️', href: 'pages/contributing/standards.html' },
  { label: 'Git Workflow',   icon: '🌿', href: 'pages/contributing/git-workflow.html' },
  { label: 'Testing',        icon: '🧪', href: 'pages/contributing/testing.html' },

  { type: 'group', label: 'OPEN SOURCE' },
  { label: 'Overview',         icon: '🌍', href: 'pages/opensource/index.html' },
  { label: 'Code of Conduct',  icon: '🤝', href: 'pages/opensource/code-of-conduct.html' },
  { label: 'License (Apache)', icon: '📄', href: 'pages/opensource/license.html' },
  { label: 'Security Policy',  icon: '🔐', href: 'pages/opensource/security-policy.html' },
  { label: 'ADRs',             icon: '📐', href: 'pages/opensource/adrs.html' },

  { type: 'group', label: 'LEARN' },
  { label: 'Go for JS Devs',   icon: '🐹', href: 'pages/concepts.html' },
  { label: 'Security Review',  icon: '🔒', href: 'pages/security.html' },
  { label: 'Self-Check Quiz',  icon: '🎯', href: 'pages/quiz.html' },
  { label: 'Q&A',              icon: '💬', href: 'pages/qa.html' },

  { type: 'divider' },
  { label: 'User Docs →', icon: '📚', href: 'user/index.html', cls: 'external' },
];

// ── Theme ─────────────────────────────────────────────────────
const THEME_KEY = 'tractl-docs-theme';

function initTheme() {
  const saved = localStorage.getItem(THEME_KEY) || 'dark';
  document.documentElement.setAttribute('data-theme', saved);
  updateThemeBtn(saved);
}

function toggleTheme() {
  const current = document.documentElement.getAttribute('data-theme') || 'dark';
  const next = current === 'dark' ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', next);
  localStorage.setItem(THEME_KEY, next);
  updateThemeBtn(next);
}

function updateThemeBtn(theme) {
  const btn = document.getElementById('theme-toggle');
  if (btn) btn.textContent = theme === 'dark' ? '☀️' : '🌙';
}

// ── Path helpers ──────────────────────────────────────────────
// Compute the docs-root prefix based on the current page's depth.
// We detect depth from the <base href> tag set by each page.
function getBase() {
  const base = document.querySelector('base');
  return base ? base.getAttribute('href') : './';
}

// Resolve a docs-root-relative href to absolute (using <base href>).
// With <base href="../../">, all hrefs naturally resolve, so we just return as-is.
function resolveHref(href) { return href; }

// Determine if a nav item href matches the current page
function isCurrentPage(itemHref) {
  const base = getBase();
  const currentPath = window.location.pathname;
  const currentSearch = window.location.search;

  // Build the effective URL this nav item would navigate to
  const a = document.createElement('a');
  a.href = itemHref; // base href does the resolution
  const itemPath = a.pathname;
  const itemSearch = a.search;

  // Compare path (ignore trailing slash differences)
  const pathMatch = currentPath.replace(/\/$/, '') === itemPath.replace(/\/$/, '');
  const searchMatch = itemSearch === currentSearch;

  return pathMatch && searchMatch;
}

// ── Build nav ─────────────────────────────────────────────────
function buildNav() {
  const nav = document.getElementById('site-nav');
  if (!nav) return;

  let html = '';
  for (const item of NAV_ITEMS) {
    if (item.type === 'group') {
      html += `<div class="nav-group-label">${item.label}</div>`;
    } else if (item.type === 'divider') {
      html += `<hr class="nav-divider">`;
    } else {
      const active = isCurrentPage(item.href) ? ' active' : '';
      const cls = (item.cls || '') + active;
      html += `<a class="nav-link${cls ? ' ' + cls : ''}" href="${item.href}">
        <span class="nav-icon">${item.icon}</span>${item.label}</a>`;
    }
  }
  nav.innerHTML = html;
}

// ── Build site header ─────────────────────────────────────────
function buildHeader(title) {
  const el = document.getElementById('site-header');
  if (!el) return;
  document.title = title ? `${title} — traCtl Dev Docs` : 'traCtl Developer Docs';
  el.innerHTML = `
    <button id="mobile-menu-btn" onclick="toggleMobileNav()" aria-label="Menu">☰</button>
    <a class="header-logo" href="index.html">
      <span class="logo-text">traCtl</span>
      <span class="logo-sub">Developer Docs</span>
    </a>
    <div class="header-spacer"></div>
    <div class="header-links">
      <a href="pages/contributing/index.html">Contributing</a>
      <a href="pages/opensource/index.html">Open Source</a>
      <a href="https://github.com/tractl/tractl" target="_blank" rel="noopener">GitHub ↗</a>
    </div>
    <button class="theme-btn" id="theme-toggle" onclick="toggleTheme()" title="Toggle theme">☀️</button>
  `;
  updateThemeBtn(document.documentElement.getAttribute('data-theme') || 'dark');
}

// ── Mobile nav ────────────────────────────────────────────────
function toggleMobileNav() {
  const nav = document.getElementById('site-nav');
  if (nav) nav.classList.toggle('open');
}

// Close nav on outside click (mobile)
document.addEventListener('click', e => {
  const nav = document.getElementById('site-nav');
  const btn = document.getElementById('mobile-menu-btn');
  if (nav && nav.classList.contains('open') && !nav.contains(e.target) && e.target !== btn) {
    nav.classList.remove('open');
  }
});

// ── Page init ─────────────────────────────────────────────────
function initPage(opts = {}) {
  initTheme();
  buildHeader(opts.title || '');
  buildNav();
  if (typeof mermaid !== 'undefined') {
    initMermaid();
  }
}

// ── Mermaid ───────────────────────────────────────────────────
function initMermaid() {
  const dark = document.documentElement.getAttribute('data-theme') !== 'light';
  mermaid.initialize({
    startOnLoad: true,
    securityLevel: 'loose', // required for click hrefs
    theme: dark ? 'dark' : 'default',
    themeVariables: dark ? {
      background: '#161b22',
      primaryColor: '#1c2230',
      primaryTextColor: '#e6edf3',
      primaryBorderColor: '#30363d',
      lineColor: '#8b949e',
      secondaryColor: '#0d1117',
      tertiaryColor: '#1c2230',
      edgeLabelBackground: '#161b22',
      clusterBkg: '#1c2230',
      clusterBorder: '#30363d',
    } : {}
  });
}

// ── Markdown rendering ────────────────────────────────────────
async function renderMarkdown(containerId, mdUrl) {
  const el = document.getElementById(containerId);
  if (!el) return;

  el.innerHTML = '<p class="md-loading">Loading…</p>';
  try {
    const res = await fetch(mdUrl);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const text = await res.text();
    if (typeof marked === 'undefined') {
      el.innerHTML = `<pre style="white-space:pre-wrap;font-size:13px;color:var(--muted)">${escHtml(text)}</pre>`;
    } else {
      el.innerHTML = marked.parse(text);
    }
    el.classList.add('md-content');
  } catch (err) {
    const isFile = location.protocol === 'file:';
    el.innerHTML = `<div class="md-error">
      <strong>Could not load document.</strong><br>
      ${isFile
        ? 'Markdown rendering requires a local server. Run: <code>cd docs && python3 -m http.server 8080</code> then open <code>http://localhost:8080</code>'
        : 'Run <code>make sync-docs</code> to copy markdown files into <code>docs/_raw/</code>, then re-deploy.'
      }<br>
      <small>Error: ${escHtml(err.message)}</small>
    </div>`;
  }
}

function escHtml(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

// ── Tabs ──────────────────────────────────────────────────────
function switchTab(btn, paneId) {
  const bar  = btn.closest('.tab-bar');
  const tabs = bar.closest('.tabs');
  bar.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
  tabs.querySelectorAll('.tab-pane').forEach(p => p.classList.remove('active'));
  btn.classList.add('active');
  const pane = document.getElementById(paneId) || tabs.querySelector(`[data-pane="${paneId}"]`);
  if (pane) pane.classList.add('active');
}

function initTabs() {
  // auto-activate first tab in each .tabs block
  document.querySelectorAll('.tabs').forEach(tabs => {
    const firstBtn  = tabs.querySelector('.tab-btn');
    const firstPane = tabs.querySelector('.tab-pane');
    if (firstBtn  && !tabs.querySelector('.tab-btn.active'))  firstBtn.classList.add('active');
    if (firstPane && !tabs.querySelector('.tab-pane.active')) firstPane.classList.add('active');
  });
}

// ── Accordions ────────────────────────────────────────────────
function toggleAccordion(header) {
  const body = header.nextElementSibling;
  header.classList.toggle('open');
  body.classList.toggle('open');
}

function initAccordions() {} // auto via onclick in HTML

// ── Quiz ──────────────────────────────────────────────────────
let quizAnswered = 0;
let quizCorrect  = 0;
let quizTotal    = 0;

function initQuizCount() {
  quizTotal = document.querySelectorAll('.quiz-item').length;
}

function answerQuiz(optEl, qId, correct) {
  const item = document.getElementById(qId);
  if (item.dataset.answered) return;
  item.dataset.answered = '1';
  quizAnswered++;

  item.querySelectorAll('.quiz-opt').forEach(o => {
    if (o === optEl) {
      o.classList.add(correct ? 'correct' : 'wrong');
    } else {
      o.classList.add('dim');
    }
  });

  if (correct) {
    quizCorrect++;
    const exp = document.getElementById(qId + '-exp');
    if (exp) exp.classList.add('show');
  } else {
    // still show explanation on wrong answer
    const exp = document.getElementById(qId + '-exp');
    if (exp) exp.classList.add('show');
  }

  if (quizAnswered === quizTotal) showScore();
}

function showScore() {
  const el = document.getElementById('quiz-score');
  if (!el) return;
  const pct = Math.round((quizCorrect / quizTotal) * 100);
  let msg = pct === 100 ? '🎉 Perfect! You understand the codebase.' :
            pct >= 66  ? '👍 Solid understanding. Review the misses.' :
                          '📖 Re-read the sections for the ones you missed.';
  el.innerHTML = `<div class="score-num">${quizCorrect}/${quizTotal}</div>
    <p style="color:var(--muted);margin-top:8px">${msg}</p>`;
  el.classList.add('show');
}

// ── Data helpers — prefer window global (works on file://), fall back to fetch ──
async function getData(globalName, jsonPath) {
  if (typeof window[globalName] !== 'undefined') return window[globalName];
  const res = await fetch(jsonPath);
  if (!res.ok) throw new Error(`Failed to load ${jsonPath}: HTTP ${res.status}`);
  return res.json();
}

// ── Generic module page ───────────────────────────────────────
async function initModulePage() {
  const id = new URLSearchParams(window.location.search).get('id');
  if (!id) { document.getElementById('module-content').innerHTML = '<p>No module id specified.</p>'; return; }

  try {
    const data = await getData('MODULES_DATA', 'data/modules.json');
    const mod  = data[id];
    if (!mod) throw new Error('Module not found: ' + id);
    renderModuleDetail(mod);
  } catch (e) {
    document.getElementById('module-content').innerHTML = `<div class="callout danger"><p>${escHtml(e.message)}</p></div>`;
  }
}

function renderModuleDetail(m) {
  const el = document.getElementById('module-content');
  document.title = `${m.title} — traCtl Dev Docs`;
  document.getElementById('mod-breadcrumb-last').textContent = m.title;

  let html = `
    <div class="hero" style="margin-bottom:32px">
      <div class="hero-title">${m.title}</div>
      <div class="hero-desc">${m.tagline}</div>
      <div class="hero-tags">
        <span class="tag blue">${m.path}</span>
        <span class="tag ${m.category === 'critical' ? 'blue' : m.category === 'surface' ? 'green' : 'purple'}">${m.category}</span>
      </div>
    </div>
    <h2>What it does</h2>
    <p>${m.description}</p>`;

  if (m.keyTypes && m.keyTypes.length) {
    html += `<h2>Key Types</h2><div class="table-wrap"><table>
      <tr><th>Type</th><th>Kind</th><th>Purpose</th></tr>
      ${m.keyTypes.map(t => `<tr><td><code>${t.name}</code></td><td>${t.kind}</td><td>${t.purpose}</td></tr>`).join('')}
    </table></div>`;
  }

  if (m.interfaces && m.interfaces.length) {
    html += `<h2>Interface</h2>`;
    m.interfaces.forEach(i => {
      html += `<h4><code>${i.name}</code></h4><p>${i.desc}</p>`;
      if (i.methods) {
        html += `<pre>${escHtml(i.methods)}</pre>`;
      }
    });
  }

  if (m.goConceptsUsed && m.goConceptsUsed.length) {
    html += `<h2>Go Concepts Used Here</h2>
      <div style="display:flex;gap:8px;flex-wrap:wrap;margin-bottom:16px">
        ${m.goConceptsUsed.map(c => `<span class="tag blue">${c}</span>`).join('')}
      </div>`;
  }

  if (m.nodeAnalogy) {
    html += `<div class="callout tip"><p><strong>Node.js analogy:</strong> ${m.nodeAnalogy}</p></div>`;
  }

  if (m.codeExample) {
    html += `<h2>Code Example</h2><pre>${escHtml(m.codeExample)}</pre>`;
  }

  if (m.relatedModules && m.relatedModules.length) {
    html += `<h2>Related Modules</h2>
      <div class="card-grid">
        ${m.relatedModules.map(r => `<a class="card" href="${window.location.pathname}?id=${r.id}">
          <h4>${r.name}</h4><p>${r.rel}</p>
        </a>`).join('')}
      </div>`;
  }

  el.innerHTML = html;
}

// ── Generic surface page ──────────────────────────────────────
async function initSurfacePage() {
  const id = new URLSearchParams(window.location.search).get('id');
  if (!id) return;

  try {
    const data = await getData('SURFACES_DATA', 'data/surfaces.json');
    const surf = data[id];
    if (!surf) throw new Error('Surface not found: ' + id);
    renderSurfaceDetail(surf);
  } catch (e) {
    document.getElementById('surface-content').innerHTML = `<div class="callout danger"><p>${escHtml(e.message)}</p></div>`;
  }
}

function renderSurfaceDetail(s) {
  const el = document.getElementById('surface-content');
  document.title = `${s.title} — traCtl Dev Docs`;
  document.getElementById('surf-breadcrumb-last').textContent = s.title;

  let html = `
    <div class="hero" style="margin-bottom:32px">
      <div class="hero-title">${s.icon} ${s.title}</div>
      <div class="hero-desc">${s.tagline}</div>
      <div class="hero-tags">
        ${(s.tags||[]).map(t => `<span class="tag ${t.cls||''}">${t.label}</span>`).join('')}
      </div>
    </div>
    <h2>What it is</h2>
    <p>${s.description}</p>`;

  if (s.entryPoint) {
    html += `<h2>Entry Point</h2>
      <div class="callout info"><p><code>${s.entryPoint.path}</code> — ${s.entryPoint.desc}</p></div>`;
  }

  if (s.howFrontendTalks) {
    html += `<h2>How the Frontend Communicates</h2><p>${s.howFrontendTalks}</p>`;
  }

  if (s.keyFiles && s.keyFiles.length) {
    html += `<h2>Key Files</h2><div class="table-wrap"><table>
      <tr><th>File</th><th>Purpose</th></tr>
      ${s.keyFiles.map(f => `<tr><td><code>${f.path}</code></td><td>${f.purpose}</td></tr>`).join('')}
    </table></div>`;
  }

  if (s.routes && s.routes.length) {
    html += `<h2>API Routes</h2><div class="table-wrap"><table>
      <tr><th>Method</th><th>Path</th><th>What it does</th></tr>
      ${s.routes.map(r => `<tr><td><strong>${r.method}</strong></td><td><code>${r.path}</code></td><td>${r.desc}</td></tr>`).join('')}
    </table></div>`;
  }

  if (s.codeExample) {
    html += `<h2>Code Snippet</h2><pre>${escHtml(s.codeExample)}</pre>`;
  }

  if (s.whenToUse) {
    html += `<div class="callout tip"><p><strong>When to use this surface:</strong> ${s.whenToUse}</p></div>`;
  }

  el.innerHTML = html;
}

// ── Auto init on DOM ready ────────────────────────────────────
document.addEventListener('DOMContentLoaded', () => {
  initTheme();
  initTabs();
});
