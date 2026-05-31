# traCtl UI Specification

Authority: ADR-016
Companion: tractl_ui_screen_inventory.md
Phase: 15
Last updated: 2026-05-31

---

## Design principles

- Form-first: every engine feature exposed as structured form fields, never raw YAML
- Script Editor is the only screen where users type code — all other panels are forms
- Minimal chrome: only task-critical actions in primary chrome
- Context actions live in ··· menu, not the toolbar
- Light and dark theme mandatory from first build
- No decorative elements, no gradients, no shadows
- Progressive disclosure: start sparse, reveal complexity on demand
- Keyboard-first: every action has a keyboard shortcut; shortcuts reference always accessible
- Git-first: git status always visible, never blocking, never mandatory

---

## Persistence model

All requests and workflows auto-save to the traCtl app directory on first interaction.
No explicit Save button. Changes persist automatically as YAML files.

OS-appropriate locations:
- macOS:   ~/Library/Application Support/tractl/
- Windows: %APPDATA%\tractl\
- Linux:   ~/.config/tractl/   (XDG compliant)

Files are human-readable YAML, Git-compatible, and CLI-runnable without modification.

Export (··· menu): writes to a user-chosen path in YAML, JSON, or TOON format.
Export is the only action that prompts for a file location.

---

## Git integration

### Alpha (status bar indicator only)

Git status indicator in status bar, right zone. Three states:
- git: untracked — workspace not inside a git repo (grey, advisory)
- git: clean — all traCtl files committed (green)
- git: N uncommitted — files changed since last commit (amber)

Clicking indicator opens minimal git panel:
- Current branch name (read-only)
- List of changed traCtl YAML files
- "Open in terminal" shortcut
- "Show in Finder / Show in Explorer" shortcut

Advisory when workspace is not a git repo — informational only, never blocks execution.
traCtl does not implement git. It reads git status via shell and links to system tooling.

··· context menu additions for sidebar items (alpha):
- Show in Finder / Show in Explorer
- Copy file path

### Deferred to v1
- Inline commit from UI, branch switcher, diff viewer
- Remote integration (GitHub, GitLab, Bitbucket)
- Auto-commit on save, git history per workflow

---

## Shell

### Activity bar (48px, leftmost)
Icon-only vertical bar. No text labels — tooltips on hover.
Traffic lights at top (macOS desktop only).

Icons top to bottom:
- Requests (bolt icon)
- Workflows (topology icon)
- Overlays (stack icon)
- History (clock icon)
- [separator]
- Environments (layers icon)
- [spacer]
- Settings (gear icon, bottom)

Active icon: blue left-edge accent bar (3px).
Clicking an active icon collapses the side panel.
Clicking a different icon switches the side panel content.

### Side panel (220px, slides in/out)
Opens when an activity bar icon is clicked. Closes on re-click or × button.
When closed, full width goes to the editor — no wasted chrome.

Each panel has:
- Header: panel title (uppercase, muted) + × close button
- Scrollable content area
- Primary action at top (New request / New workflow / etc.)

Content per panel:
- Requests: Pinned section, Recent section. Each item: icon + name + method badge + hover ··· menu
- Workflows: New workflow + Open file actions. Grouped: traCtl native | OpenAPI. Groups collapsible.
- Overlays: overlay file list
- History: shortcut to Run History screen
- Environments: environment list with active indicator and variable preview

··· context menu per item: Open, Pin/Unpin, Add to workflow (requests only),
Duplicate, Export, Show in Finder/Explorer, Copy file path, Delete

### Topbar (42px)
Left: command palette search bar (flex:1, max 260px)
  - Placeholder: "Search and commands…"
  - Keyboard hint: ⌘K
  - Clicking opens command palette overlay

Centre: environment selector pill
  - Visible always regardless of current screen
  - Blue bordered pill for non-production environments
  - Red bordered pill for production — visual danger signal
  - Clicking cycles environments or opens environment picker
  - No environment: grey pill, "No environment"

Right: theme toggle (moon/sun icon)

### Tab bar (34px, below topbar)
Browser-style tabs. Each open request or workflow gets a tab.
Tabs persist across navigation — multiple items open simultaneously.

Each tab: method badge (coloured) + name (truncated) + × close button (hover-reveal)
Active tab: white background + blue 2px bottom border
New tab: + button rightmost

Method badge colours in tabs:
- GET: blue (#E6F1FB bg / #0C447C text)
- POST: green (#EAF3DE bg / #27500A text)
- PUT: amber (#FAEEDA bg / #633806 text)
- DELETE: red (#FCEBEB bg / #791F1F text)
- PATCH: purple (#EEEDFE bg / #3C3489 text)
- WF (workflow): green, same as POST

### URL bar (below tab bar)
Three elements only: [ method selector ] [ url input ] [ autosave hint ] [ Send button ]
Send button: blue background, "Send" label (not "Run" — aligns with HTTP mental model).
Autosave hint: "✓ Saved" in muted text, updates on change.

### Bottom toolbar (26px, above status bar)
Thin utility bar. Houses secondary controls out of primary chrome.

Left group:
- Layout toggle: Stacked ↕ (default) | Side by side ↔
  Stacked: config panel above, response below — works on any screen size
  Side by side: config left, response right — better on wide monitors
  Keyboard shortcut: ⌘⇧L

- Word wrap toggle: wraps long lines in code blocks and response body

Right:
- Shortcuts button: opens keyboard shortcuts reference popup
  Keyboard shortcut: ⌘/

### Status bar (22px, bottom)
Three zones:

Left: engine state dot + text
  - Green + "Engine ready"
  - Amber + "Executing…"
  - Green + "[status] · [duration] · [assertions summary]"
  - Red + "Error: [short message]"

Right: git indicator | surface + version
  - git: clean (green) | git: N uncommitted (amber) | git: untracked (grey)
  - "Desktop · v0.1.0-alpha" or "Web · v0.1.0-alpha"

---

## Keyboard shortcuts reference popup

Opened by: Shortcuts button in bottom toolbar, or ⌘/
Closed by: Escape or clicking outside

Grouped sections: General | Request | Response | Navigation

General:
- ⌘K — Search and commands
- ⌘/ — Keyboard shortcuts
- ⌘N — New request
- ⌘⇧N — New workflow
- ⌘E — Switch environment
- ⌘, — Settings

Request:
- ⌘↵ — Send request
- ⌘W — Close tab
- ⌘] / ⌘[ — Next / previous tab
- ⌥G — Select GET method
- ⌥P — Select POST method
- ⌥X — Select DELETE method

Response:
- ⌘. — Copy response body
- ⌘⇧L — Toggle layout

Navigation:
- ⌘H — Open run history
- ⌥W — Go to workflows

---

## Command palette

Opened by: clicking topbar search bar or ⌘K
Closed by: Escape or clicking outside

Input: full-text search across all requests, workflows, overlays, history, commands.
Results grouped: Recent | Commands
Each result: icon + name + type label or keyboard shortcut hint
Arrow keys navigate, Enter selects.

---

## S1: Fast Start screen

Shown on launch when no previous session, or when Home icon (activity bar top) is clicked.

Layout:
- Greeting line (contextual time of day)
- Subtitle: "What would you like to validate today?"
- 2×2 action card grid:
  - New request → opens in new tab, switches side panel to Requests
  - New workflow → opens in new tab, switches side panel to Workflows
  - Open file → file picker → opens as workflow
  - Open OpenAPI spec → file picker → opens in Workflow Canvas
- Recent list (unified: requests + workflows + OpenAPI)
  - Columns: type icon, name, type badge, time ago, hover-reveal Send/Run button
  - "View all" → Run History screen

---

## S2: Request Editor

### Config panel tabs
Order: Params | Headers | Body | Auth | Pre-script | Post-script | Assertions | Extracts | Settings
Pre-script and Post-script: tinted blue to signal code-entry context.
Count badges on Headers, Assertions, Extracts when populated.

### Body tab — encoding switcher
When Body tab is selected, the encoding bar replaces the panel body:

Encoding bar: "Encoding" label + segmented tab control
Encoding options: JSON | Form data | Multipart | Raw | Binary | None

JSON:
- Code editor (Monaco in production)
- Supports ${vars.x} and ${secrets.env.name} expression syntax

Form data (application/x-www-form-urlencoded):
- Key-value table: checkbox | key input | value/file cell | type selector (Text/File) | delete
- File rows: "Choose file" button + filename display instead of text input
- Drag file onto form data area → auto-creates a File row
- "Add field" row at bottom

Multipart (multipart/form-data):
- Same table as Form data
- Info note: "Each field sent as a separate part with its own Content-Type"
- Use when uploading files alongside text fields

Raw:
- Content-Type selector: text/plain | text/xml | application/xml | text/html
- Textarea below (monospace)

Binary:
- Drag-and-drop zone with upload icon
- "Drop a file here, or click to choose"
- Content-Type auto-detected from file extension

None:
- Empty state: "No body sent with this request"

### Other panels (unchanged from previous spec)
Params, Headers: key-value tables with checkbox, inline edit, delete
Auth: type selector revealing appropriate fields. Secret references via ${secrets.env.name}
Pre-script: hint bar + code editor + frozen context chips. Maps to hooks.beforeStep
Post-script: hint bar + code editor + context chips. Maps to hooks.afterStep
Assertions: kind | operator | expected | severity toggle | delete. Scale-aware grouping.
Extracts: source | path | → | variable name | scope | delete. Downstream ref shown below.
Settings: timeout (number + unit), retry (selector + fields), failure policy

### Results panel (stacked by default, below config)
Default layout: stacked (config above, results below, drag handle between).
Side by side: toggled from bottom toolbar.
If Send triggered while results hidden: panel auto-shows.

Header: "Response" label + outcome badge + duration (right-aligned)
Timing bar (shown post-run): DNS | TCP | TLS | TTFB | Transfer segments with legend
Result tabs: Body | Headers | Assertions | Extracts
- Body: status chip + duration chip + content-type chip + response body (code block)
- Headers: key-value table, sensitive headers shown as [masked]
- Assertions: pass/fail rows with received value and severity
- Extracts: variable reference + resolved value + scope note

Idle state: play icon + "Hit Send to execute" + ⌘↵ hint

---

## S3: Workflow Canvas

### Layout
Activity bar (48px) + side panel (workflows) + canvas area + optional result bar

### Topbar, tab bar, URL bar
Same shell as S2. Workflow opens as a tab with "WF" method badge.
URL bar replaced by workflow bar: name + step count + topology summary + autosave + Run workflow button.

### Filter bar (below workflow bar)
Pill chips for filtering by workflow name. "All workflows" default.
When workflow count > 5: chips collapse into dropdown selector.
Active chip: info colour accent.

### Graph view (default)

DAG layout rules:
- Single root: vertical chain, fan-out at branch points
- Multiple roots: side-by-side columns separated by "independent" vertical divider
- Fan-out: horizontal branch bar with dot at junction, parallel label, branch legs below
- Merge point: multiple connectors converging on a single downstream node

Plus buttons (step insertion):
- Vertical connector hover: + at midpoint → inserts sequential step between
- Junction bar hover: + on bar → adds parallel step at same level (same dependsOn)
- Leaf card hover: + below card → adds sequential child (dependsOn set to this card)
All + buttons: 16px circle, info colour on hover

Step card anatomy:
- Step number circle + step ID + method badge + outcome badge
- URL (monospace, truncated)
- Pills: Auth, Assertions count (count only, never per-assertion detail), Pre-script, extract pills
- Outcome badge: idle | running | passed | failed | skipped
- Card border: default grey → green on pass → red (1.5px) on fail
- Click: opens step detail popup
  - Pre-run: opens on Request tab
  - Post-run: opens on Result tab automatically

Code view (toggle in workflow bar):
- Full generated YAML, read-only, copy button
- Exactly matches auto-saved file — no gap between views

### Workflow config strip
Collapsed by default: read-only chips (Concurrency, Timeout, failure policy, hooks, overlay)
Clicking expands full config panel: concurrency, timeout + unit, failure policy, hooks, overlay selector
Collapse/expand chevron rightmost.

### Step detail popup
Tabs: Request | Dependencies | Pre-script | Post-script | Assertions | Extracts | Result

Result tab:
- Hidden pre-run. Shown post-run, auto-selected on card click after run.
- Tab label: green "Result ✓" or red "Result — N failed"
- Content: response meta, timing bar, assertions (failed-first grouping), extracts, response body

Assertions scale pattern (applies everywhere):
- Summary line: "N failed · M passed"
- Failed: shown first, always expanded, red background, expected vs received detail
- Passed: collapsed under toggle by default, green background
- Cards and result bar chips show count only — never per-assertion detail

Dependencies tab:
- "Depends on" section: upstream step chips or "Root step"
- "Required by" section: downstream step chips
- Parallel eligibility note

### Result bar (full width, above bottom toolbar)
Collapses/expands via chevron.
Header: title + outcome badge + total time + "View full results" button (post-run)
Expanded: horizontal scrollable row of step chips
  Each chip: outcome icon + step name + assertion summary (N/M) + duration
  Failed chip: red background, "N/M failed"
  Clicking chip: opens step detail popup on Result tab

Full results popup:
- Summary metrics: outcome, total time, steps ran, assertions total
- Dependency waterfall (ASCII, Phase 7 engine output)
- Per-step outcome rows with exact failure detail inline

---

## S7: Overlay Editor

Three-panel layout: patch list (220px) | patch editor (flex) | preview panel (260px)

### Patch list panel
Header: "Patches" label + Add button
Each row: action badge (replace/add/remove/merge) + name + target path + hover delete
Selected: surface bg + border accent

### Patch editor panel
- Patch name field
- Action picker: 4 cards (replace/add/remove/merge), coloured by semantic meaning
  Selecting "remove" hides patch data and merge strategy
- Target section: Path mode (JSONPath input) | Match mode (field + operator + value)
- Patch data: value input (hidden for remove)
- Merge strategy: replace | deep-merge | array-append | array-prepend
- Scope: all workflows | selected workflows | named spec only

### Preview panel
Diff mode: red strikethrough (removed) + green (added) + grey (context) + @@ section headers
Result mode: full resolved document, changed lines in green, read-only

Status bar left zone: "N patches · N conflicts" — amber when conflicts > 0

---

## S9: Run History

### Page bar
Title + run count subtitle + search input + "Clear all" button
Clear all: confirmation prompt "Clear all N runs? This cannot be undone."

### Filter row
Type: All | Workflows | Requests
Outcome: All | Passed | Failed
Environment: All | [named environments]
Sort: Newest first | Oldest first | Slowest first | Most failures first

### History list
Grouped by day. Each row:
- Outcome circle (green check / red X)
- Run name + type badge + step/assertion summary + environment badge
- Assertion chip: "N/M passed" (green) or "N/M failed" (red)
- Duration + time
- Hover-reveal: Re-run button + Open button

Re-run: uses environment stored in original run record, not currently active environment.

### Run detail popup
Header: name + timestamp + env · Re-run button · Open button · ×
Metrics: outcome | duration | steps | assertions
Dependency waterfall (ASCII)
Per-step outcomes with exact failure detail inline

### Data model
Desktop retention default: 1000 runs (SQLite, auto-prune)
Web retention default: 50 runs (IndexedDB, auto-prune)
Configurable in Settings.
Local-first. Cloud sync deferred to paid tier.

Each run record stores: run ID, timestamp, source name/type, environment snapshot,
outcome, total duration, per-step outcomes, provenance trace (Phase 7 ExecutionRecord).

---

## S10: Settings

Reached via Settings icon at bottom of activity bar.

### Settings nav (left, 168px)
Sections: General | Appearance | History | Engine | Git | About

### General
Workspace: app directory (Change button), default export format, local API port (requires restart)
Execution defaults: timeout, failure policy, concurrency

### Appearance
Theme: System | Light | Dark
Font size: Small (11px) | Default (13px) | Large (15px)
Code editor font: JetBrains Mono | Fira Code | Menlo | System mono

### History
Desktop retention: 50 | 200 | 1000 | Unlimited
Web retention: 10 | 25 | 50
Clear all history (danger, confirmation required)
Export history as JSON

### Engine
Script timeout + unit selector
Script log output toggle (show tractl.log() in results)
Capture TLS details toggle (Phase 7.5, off by default)
Capture DNS details toggle (off by default)

### Git
Show git status toggle
Warn when workspace is not a git repo toggle
Git executable path + auto-detect button

### About
App card: logo + name + alpha tag + version string
Build info: engine version, Wails runtime, Go version, schema version
App directory path + Open button
Reset all settings (danger, confirmation required, does not delete files)

### Settings persistence
Stored as settings.yaml in app directory.
Changes apply immediately except local API port (requires restart).

---

## Design system tokens

Method badge colours (used in tabs, cards, sidebar):
- GET:    #E6F1FB bg / #0C447C text
- POST:   #EAF3DE bg / #27500A text
- PUT:    #FAEEDA bg / #633806 text
- DELETE: #FCEBEB bg / #791F1F text
- PATCH:  #EEEDFE bg / #3C3489 text

Environment pill:
- Non-production: 1.5px #185FA5 border, #E6F1FB bg, #0C447C text
- Production: 1.5px #A32D2D border, #FCEBEB bg, #791F1F text

Semantic colours (light / dark):
- Pass:    #EAF3DE bg / #3B6D11 text  |  #173404 bg / #C0DD97 text
- Fail:    #FCEBEB bg / #A32D2D text  |  #501313 bg / #F7C1C1 text
- Warning: #FAEEDA bg / #854F0B text  |  #412402 bg / #FAC775 text
- Info:    #E6F1FB bg / #185FA5 text  |  #042C53 bg / #B5D4F4 text

Send button: #185FA5 bg / #E6F1FB text — same in light and dark
Script tabs: info colour ramp

Typography:
- UI labels: var(--font-sans), 11–13px, weight 400/500 only
- Code/expressions: var(--font-mono), 11–12px
- Sentence case always

Borders: 0.5px throughout. 1.5px for environment pill. 2px only for featured card accent.
Radius: 7–8px for inputs/buttons, 10–12px for cards.
No shadows, no gradients, no decorative effects.
