# Portfolio Site Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build Jasen Nicely's personal portfolio site — a single-page Go + htmx + Alpine.js site with htmx case-study expansion, a live terminal easter egg, and an SQLite-backed contact form.

**Architecture:** Single Go binary (stdlib `net/http`, Go 1.22+ method routing) with all templates and static assets embedded via `embed.FS`. All site copy lives as typed Go values in `internal/content`; templates and the terminal command registry both render from those structs. Tailwind CSS built with the standalone CLI (no Node).

**Tech Stack:** Go (stdlib), `html/template`, htmx 2 (vendored), Alpine.js 3 (vendored), Tailwind CSS 4 standalone CLI, modernc.org/sqlite (pure Go, no cgo).

**Spec:** `docs/superpowers/specs/2026-06-12-portfolio-site-design.md`

---

## File map

| Path | Responsibility |
|---|---|
| `embed.go` | Root package `portfolio`; embeds `templates/` and `static/` |
| `cmd/server/main.go` | Flag parsing, wiring, ListenAndServe |
| `internal/content/content.go` | All site copy as typed Go values + `ProjectBySlug` |
| `internal/store/store.go` | SQLite inquiry persistence |
| `internal/terminal/terminal.go` | Terminal command registry |
| `internal/handler/handler.go` | `Server`, `New`, render helpers, `data()` |
| `internal/handler/pages.go` | home, project, projectCard, notFound handlers |
| `internal/handler/contact.go` | contact form handler + validation |
| `internal/handler/terminal.go` | `POST /terminal` handler |
| `internal/handler/middleware.go` | panic-recovery middleware |
| `templates/layout.html` | HTML skeleton; defines `layout` |
| `templates/pages/{home,project_page,404,500}.html` | Each defines `main` |
| `templates/partials/*.html` | `project_card`, `project_detail`, `contact_form`, `contact_success`, `contact_error`, `terminal` |
| `static/js/{htmx,alpine}.min.js` | Vendored libraries |
| `static/js/grid.js` | Phosphor-grid canvas background |
| `static/js/site.js` | Typing, count-up, scroll-reveal, terminal glue |
| `tailwind.css` → `static/css/site.css` | Theme tokens, fonts, custom CSS (built artifact is committed) |
| `static/fonts/*.woff2` | Self-hosted IBM Plex |
| `static/resume/jasen-nicely-resume.pdf` | Downloadable resume |

All shell commands below are Git Bash syntax, run from the repo root `D:\code\portfolio-fable`.

---

### Task 1: Scaffold, vendored assets, fonts, resume PDF

**Files:**
- Create: `.gitignore`, `go.mod` (via `go mod init`), `static/favicon.svg`, directory tree, vendored JS/fonts/Tailwind binary, resume PDF copy

- [ ] **Step 1: Verify Go ≥ 1.22 is installed**

Run: `go version`
Expected: `go version go1.2x ...` (1.22 or newer). If missing, stop and report.

- [ ] **Step 2: Create directory tree and module**

```bash
mkdir -p cmd/server internal/content internal/handler internal/store internal/terminal \
  templates/pages templates/partials static/js static/css static/fonts static/resume bin
go mod init portfolio
```

- [ ] **Step 3: Create `.gitignore`**

```gitignore
bin/
*.db
```

- [ ] **Step 4: Download vendored frontend assets**

```bash
curl -fsSL -o static/js/htmx.min.js   https://unpkg.com/htmx.org@2/dist/htmx.min.js
curl -fsSL -o static/js/alpine.min.js https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js
curl -fsSL -o static/fonts/plex-sans-400.woff2 https://cdn.jsdelivr.net/fontsource/fonts/ibm-plex-sans@latest/latin-400-normal.woff2
curl -fsSL -o static/fonts/plex-sans-600.woff2 https://cdn.jsdelivr.net/fontsource/fonts/ibm-plex-sans@latest/latin-600-normal.woff2
curl -fsSL -o static/fonts/plex-sans-700.woff2 https://cdn.jsdelivr.net/fontsource/fonts/ibm-plex-sans@latest/latin-700-normal.woff2
curl -fsSL -o static/fonts/plex-mono-400.woff2 https://cdn.jsdelivr.net/fontsource/fonts/ibm-plex-mono@latest/latin-400-normal.woff2
curl -fsSL -o static/fonts/plex-mono-500.woff2 https://cdn.jsdelivr.net/fontsource/fonts/ibm-plex-mono@latest/latin-500-normal.woff2
curl -fsSL -o bin/tailwindcss.exe https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-windows-x64.exe
```

Verify: `ls -la static/js static/fonts bin` — every file present and non-zero size.

- [ ] **Step 5: Copy the resume PDF**

```bash
cp "/d/obsidian/fuel-forge/Gideon/Career/jasen-cover-resume.pdf" static/resume/jasen-nicely-resume.pdf
```

- [ ] **Step 6: Create `static/favicon.svg`**

```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><rect width="32" height="32" rx="6" fill="#09090b"/><text x="6" y="22" font-family="monospace" font-size="13" fill="#4ade80">&gt;_</text></svg>
```

- [ ] **Step 7: Commit**

```bash
git add .gitignore go.mod static
git commit -m "chore: scaffold project, vendor htmx/alpine/fonts, add resume pdf"
```

---

### Task 2: Content package

**Files:**
- Create: `internal/content/content.go`
- Test: `internal/content/content_test.go`

- [ ] **Step 1: Write the failing tests** (`internal/content/content_test.go`)

```go
package content

import (
	"strings"
	"testing"
)

func TestProjectBySlug(t *testing.T) {
	p, ok := ProjectBySlug("redline")
	if !ok {
		t.Fatal("redline not found")
	}
	if p.Name != "Redline" {
		t.Errorf("Name = %q, want Redline", p.Name)
	}
	if _, ok := ProjectBySlug("nope"); ok {
		t.Error("ProjectBySlug(nope) = found, want not found")
	}
}

func TestProjectsComplete(t *testing.T) {
	if len(Projects) != 3 {
		t.Fatalf("len(Projects) = %d, want 3", len(Projects))
	}
	seen := map[string]bool{}
	for _, p := range Projects {
		if p.Slug == "" || seen[p.Slug] {
			t.Errorf("project %q: empty or duplicate slug", p.Name)
		}
		seen[p.Slug] = true
		if p.Name == "" || p.Tagline == "" || p.Summary == "" {
			t.Errorf("project %q: missing name/tagline/summary", p.Slug)
		}
		if len(p.Problem) == 0 || len(p.Approach) == 0 || len(p.Outcome) == 0 {
			t.Errorf("project %q: case study sections incomplete", p.Slug)
		}
		if len(p.Tech) == 0 || len(p.Metrics) == 0 {
			t.Errorf("project %q: missing tech or metrics", p.Slug)
		}
	}
}

func TestProfileComplete(t *testing.T) {
	if Me.Name == "" || Me.Title == "" || Me.Pitch == "" || Me.GitHub == "" {
		t.Error("profile has empty required fields")
	}
	if !strings.Contains(Me.Email, "@") {
		t.Errorf("Email = %q, not an email", Me.Email)
	}
	if len(Me.About) == 0 {
		t.Error("About is empty")
	}
}

func TestStatsAndSkills(t *testing.T) {
	if len(Stats) != 4 {
		t.Errorf("len(Stats) = %d, want 4", len(Stats))
	}
	for _, s := range Skills {
		if s.Level <= 0 || s.Level > 100 {
			t.Errorf("skill %q level %d out of range", s.Name, s.Level)
		}
	}
	if len(Resume) == 0 {
		t.Error("Resume is empty")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/content/`
Expected: FAIL — compile error, `undefined: ProjectBySlug` etc.

- [ ] **Step 3: Write the implementation** (`internal/content/content.go`)

```go
// Package content holds every piece of site copy as typed Go values.
// Templates and the terminal command registry both render from these
// structs, so the page and the easter egg can never drift apart.
package content

type Profile struct {
	Name        string
	Title       string
	Location    string
	Email       string
	GitHub      string // full URL
	GitHubLabel string // display text
	Pitch       string
	About       []string // paragraphs
}

type Stat struct {
	Value string // rendered figure, e.g. "99.9%" — site.js animates the numeric prefix
	Label string
}

type Metric struct {
	Value string
	Label string
}

type Project struct {
	Slug      string
	Name      string
	Tagline   string
	Summary   string   // card copy
	Problem   []string // case-study paragraphs
	Approach  []string
	Outcome   []string
	Tech      []string
	Link      string // optional live URL
	LinkLabel string
	Metrics   []Metric
}

type ResumeEntry struct {
	Org     string
	Period  string
	Note    string // e.g. promotion callout
	Bullets []string
}

type Skill struct {
	Name  string
	Level int // 0–100, drives the proficiency bar width
}

type Education struct {
	School string
	Place  string
	Year   string
	Detail string
}

var Me = Profile{
	Name:        "Jasen Nicely",
	Title:       "Senior Full-Stack Software Developer",
	Location:    "McKinney, Texas",
	Email:       "jasen.nicely@proton.me",
	GitHub:      "https://github.com/fuel-",
	GitHubLabel: "github.com/fuel-",
	Pitch:       "15 years of building software that changes how people work — from week-long batch jobs cut to 12 seconds, to production SaaS shipped solo.",
	About: []string{
		"I've been writing software professionally for 15 years, and I still get genuinely excited when a hard problem lands on my desk. That curiosity is probably the thing that defines me most as a developer — more than any language or framework on my resume.",
		"For the past several years I've been building full-stack systems at SAP Concur, working primarily in Go and TypeScript/React — greenfield features, security hardening, legacy migrations, and everything in between.",
		"I'm drawn to smaller teams because I like being close to the work and the people. I'm a self-taught developer at heart: I'm used to figuring things out, and I never really stop learning.",
	},
}

var Stats = []Stat{
	{Value: "15 yrs", Label: "Professional software experience"},
	{Value: "99.9%", Label: "Speedup on a tax-data pipeline (1 week → 12 s)"},
	{Value: "30+", Label: "PRs shipped on the Announcement System"},
	{Value: "9 yrs", Label: "Full-stack focus across Go, TS, React"},
}

var Projects = []Project{
	{
		Slug:    "property-tax-pipeline",
		Name:    "Property Tax Pipeline",
		Tagline: "One week of processing, down to twelve seconds.",
		Summary: "A law firm's property-tax data process took a full week per run. I rewrote it as a Go + SQLite pipeline that finishes in 12 seconds — a 99.9% improvement.",
		Problem: []string{
			"A law firm client processed fixed-width property tax files with county-specific calculations. Their existing process took a full week of machine time per run, and a single bad input file meant starting over.",
		},
		Approach: []string{
			"I rewrote the pipeline in Go with a SQLite backend: a streaming parser for the fixed-width county files, per-county calculation rules expressed as data instead of scattered branching, and the workload restructured around set-based SQL operations instead of row-at-a-time processing.",
		},
		Outcome: []string{
			"The same workload now runs in twelve seconds — a 99.9% reduction. The firm went from scheduling around a week-long batch job to re-running it on a whim whenever the data changed.",
		},
		Tech:    []string{"Go", "SQLite", "fixed-width parsing"},
		Metrics: []Metric{{Value: "1 wk → 12 s", Label: "runtime"}, {Value: "99.9%", Label: "faster"}},
	},
	{
		Slug:    "announcement-system",
		Name:    "Announcement System",
		Tagline: "Full-stack ownership across two codebases.",
		Summary: "Led the end-to-end build of a platform-wide announcement system at SAP Concur — 30+ coordinated PRs across two codebases, from data model to UI.",
		Problem: []string{
			"The platform needed a way to publish targeted announcements to users — spanning a Go/MSSQL backend and a TypeScript/React frontend that lived in separate codebases with separate release trains.",
		},
		Approach: []string{
			"I owned the feature end to end: schema and API design on the backend, the React UI on the frontend, and the release choreography to land 30+ coordinated PRs across both codebases without blocking either team's deploys.",
		},
		Outcome: []string{
			"The system shipped and became the platform's standard channel for user-facing announcements. Along the way I hardened adjacent platform security — token invalidation, XSS fixes, SHA-512 hashing, MIME whitelisting.",
		},
		Tech:    []string{"Go", "TypeScript", "React", "MSSQL"},
		Metrics: []Metric{{Value: "30+", Label: "coordinated PRs"}, {Value: "2", Label: "codebases, one owner"}},
	},
	{
		Slug:    "redline",
		Name:    "Redline",
		Tagline: "A live SaaS for design review — built on this site's exact stack.",
		Summary: "A client-facing design review tool, live in production: clients view their redesigned site and click any element to pin feedback. Go, htmx, Alpine.js, PostgreSQL.",
		Problem: []string{
			"Website redesign feedback over email is chaos: screenshots, vague descriptions, lost threads. Clients needed a way to point at the thing they meant.",
		},
		Approach: []string{
			"Redline serves each redesign in an iframe with a companion script injected into the HTML. Clients click any element to leave a pinned note — captured as a CSS selector plus page path and stored in PostgreSQL. Sessions, CSRF protection, rate limiting, and zip-slip-safe bundle uploads round out the production hardening.",
		},
		Outcome: []string{
			"Live at redline.gideon.gg and used for real client review cycles. It's also built on exactly the stack this website uses — Go, htmx, Alpine.js — so consider both of them a demo.",
		},
		Tech:      []string{"Go", "htmx", "Alpine.js", "PostgreSQL", "Docker", "Caddy"},
		Link:      "https://redline.gideon.gg",
		LinkLabel: "redline.gideon.gg",
		Metrics:   []Metric{{Value: "Live", Label: "in production"}, {Value: "100%", Label: "server-rendered"}},
	},
}

var Resume = []ResumeEntry{
	{
		Org:    "SAP Concur — STAT Team",
		Period: "2017 — Present",
		Note:   "Promoted to Senior Technical Consultant, 2023",
		Bullets: []string{
			"Architected full-stack platform features in Go, TypeScript, React, and MSSQL.",
			"Led the end-to-end Announcement System across two codebases — 30+ coordinated PRs.",
			"Hardened platform security: token invalidation, XSS fixes, SHA-512 hashing, MIME whitelisting.",
		},
	},
	{
		Org:    "SAP Concur — RAD Team",
		Period: "2013 — 2017",
		Bullets: []string{
			"Diagnosed and resolved complex client issues in travel automation systems.",
			"Built custom solutions to client-specific requirements on proprietary mid-office software.",
		},
	},
	{
		Org:    "SAP Concur — Implementations",
		Period: "2011 — 2013",
		Bullets: []string{
			"Configured and shipped automation for travel-industry clients.",
			"Built business logic for ticketing, invoicing, refunds, and quality control — translating business requirements into working code.",
		},
	},
	{
		Org:    "Property Tax Processor",
		Period: "Contract · side project",
		Bullets: []string{
			"Designed a Go + SQLite pipeline for fixed-width property tax files with county-specific calculations — one week of processing down to twelve seconds.",
		},
	},
}

var Skills = []Skill{
	{Name: "Go", Level: 95},
	{Name: "JavaScript / TypeScript", Level: 90},
	{Name: "REST API Design", Level: 90},
	{Name: "React", Level: 85},
	{Name: "MSSQL / PostgreSQL / SQLite", Level: 85},
	{Name: "HTMX", Level: 80},
	{Name: "C# / .NET MVC", Level: 75},
}

var School = Education{
	School: "Southwestern Assemblies of God University",
	Place:  "Waxahachie, TX",
	Year:   "1998",
	Detail: "Major: Youth Ministries · Minor: Counseling",
}

// ProjectBySlug returns the project with the given slug.
func ProjectBySlug(slug string) (Project, bool) {
	for _, p := range Projects {
		if p.Slug == slug {
			return p, true
		}
	}
	return Project{}, false
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/content/ -v`
Expected: PASS (4 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/content
git commit -m "feat: content package — profile, projects, resume, skills as typed data"
```

---

### Task 3: SQLite inquiry store

**Files:**
- Create: `internal/store/store.go`
- Test: `internal/store/store_test.go`

- [ ] **Step 1: Add the SQLite dependency**

```bash
go get modernc.org/sqlite
```

- [ ] **Step 2: Write the failing tests** (`internal/store/store_test.go`)

```go
package store

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndListInquiry(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer st.Close()

	id, err := st.SaveInquiry(Inquiry{
		Name: "Ada Lovelace", Email: "ada@example.com",
		Company: "Analytical Engines", Kind: "contract", Message: "Need a compiler.",
	})
	if err != nil {
		t.Fatalf("SaveInquiry: %v", err)
	}
	if id <= 0 {
		t.Errorf("id = %d, want > 0", id)
	}

	got, err := st.ListInquiries()
	if err != nil {
		t.Fatalf("ListInquiries: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	q := got[0]
	if q.Name != "Ada Lovelace" || q.Email != "ada@example.com" ||
		q.Company != "Analytical Engines" || q.Kind != "contract" || q.Message != "Need a compiler." {
		t.Errorf("round-trip mismatch: %+v", q)
	}
	if time.Since(q.CreatedAt) > time.Minute {
		t.Errorf("CreatedAt = %v, not recent", q.CreatedAt)
	}
}

func TestPersistsAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	st, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := st.SaveInquiry(Inquiry{Name: "n", Email: "e@x.com", Kind: "other", Message: "m"}); err != nil {
		t.Fatalf("SaveInquiry: %v", err)
	}
	st.Close()

	st2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer st2.Close()
	got, err := st2.ListInquiries()
	if err != nil {
		t.Fatalf("ListInquiries: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d after reopen, want 1", len(got))
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/store/`
Expected: FAIL — compile error, `undefined: Open`

- [ ] **Step 4: Write the implementation** (`internal/store/store.go`)

```go
// Package store persists contact-form inquiries to SQLite so a bug or
// missing email config can never silently eat a lead.
package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Inquiry struct {
	ID        int64
	Name      string
	Email     string
	Company   string
	Kind      string // "hiring" | "contract" | "other"
	Message   string
	CreatedAt time.Time
}

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	const schema = `CREATE TABLE IF NOT EXISTS inquiries (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		name       TEXT NOT NULL,
		email      TEXT NOT NULL,
		company    TEXT NOT NULL DEFAULT '',
		kind       TEXT NOT NULL,
		message    TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) SaveInquiry(q Inquiry) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO inquiries (name, email, company, kind, message, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		q.Name, q.Email, q.Company, q.Kind, q.Message,
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("insert inquiry: %w", err)
	}
	return res.LastInsertId()
}

func (s *Store) ListInquiries() ([]Inquiry, error) {
	rows, err := s.db.Query(
		`SELECT id, name, email, company, kind, message, created_at
		 FROM inquiries ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("query inquiries: %w", err)
	}
	defer rows.Close()

	var out []Inquiry
	for rows.Next() {
		var q Inquiry
		var created string
		if err := rows.Scan(&q.ID, &q.Name, &q.Email, &q.Company, &q.Kind, &q.Message, &created); err != nil {
			return nil, fmt.Errorf("scan inquiry: %w", err)
		}
		q.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, q)
	}
	return out, rows.Err()
}

func (s *Store) Close() error { return s.db.Close() }
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/store/ -v`
Expected: PASS (2 tests)

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum internal/store
git commit -m "feat: sqlite inquiry store (modernc.org/sqlite, no cgo)"
```

---

### Task 4: Terminal command registry

**Files:**
- Create: `internal/terminal/terminal.go`
- Test: `internal/terminal/terminal_test.go`

- [ ] **Step 1: Write the failing tests** (`internal/terminal/terminal_test.go`)

```go
package terminal

import (
	"strings"
	"testing"
)

func TestUnknownCommand(t *testing.T) {
	r := New()
	res := r.Execute("nope")
	if !strings.Contains(res.HTML, "command not found: nope") {
		t.Errorf("HTML = %q, want command-not-found", res.HTML)
	}
	if res.Action != "" {
		t.Errorf("Action = %q, want empty", res.Action)
	}
}

func TestInputIsEscaped(t *testing.T) {
	r := New()
	res := r.Execute("<script>alert(1)</script>")
	if strings.Contains(res.HTML, "<script>") {
		t.Error("raw <script> leaked into output")
	}
	if !strings.Contains(res.HTML, "&lt;script&gt;") {
		t.Errorf("HTML = %q, want escaped input", res.HTML)
	}
}

func TestEmptyInput(t *testing.T) {
	r := New()
	if res := r.Execute("   "); res.HTML != "" || res.Action != "" {
		t.Errorf("blank input → %+v, want zero Result", res)
	}
}

func TestHelpListsAllCommands(t *testing.T) {
	r := New()
	res := r.Execute("help")
	for _, name := range []string{"help", "whoami", "projects", "open", "resume", "skills", "contact", "clear", "exit", "sudo"} {
		if !strings.Contains(res.HTML, name) {
			t.Errorf("help output missing %q", name)
		}
	}
}

func TestWhoami(t *testing.T) {
	r := New()
	if res := r.Execute("whoami"); !strings.Contains(res.HTML, "Jasen Nicely") {
		t.Errorf("whoami = %q, want name", res.HTML)
	}
}

func TestOpenProject(t *testing.T) {
	r := New()
	res := r.Execute("open redline")
	if res.Action != "open:/projects/redline" {
		t.Errorf("Action = %q, want open:/projects/redline", res.Action)
	}
	if res := r.Execute("open bogus"); !strings.Contains(res.HTML, "no such project") {
		t.Errorf("open bogus = %q", res.HTML)
	}
	if res := r.Execute("open"); !strings.Contains(res.HTML, "usage: open") {
		t.Errorf("open (no args) = %q", res.HTML)
	}
}

func TestClearAndExit(t *testing.T) {
	r := New()
	if res := r.Execute("clear"); res.Action != "clear" {
		t.Errorf("clear Action = %q", res.Action)
	}
	if res := r.Execute("exit"); res.Action != "exit" {
		t.Errorf("exit Action = %q", res.Action)
	}
}

func TestSudoHireMe(t *testing.T) {
	r := New()
	res := r.Execute("sudo hire-me")
	if res.Action != "goto:#contact" {
		t.Errorf("Action = %q, want goto:#contact", res.Action)
	}
	if res := r.Execute("sudo rm -rf /"); !strings.Contains(res.HTML, "sudoers") {
		t.Errorf("sudo other = %q, want sudoers joke", res.HTML)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/terminal/`
Expected: FAIL — compile error, `undefined: New`

- [ ] **Step 3: Write the implementation** (`internal/terminal/terminal.go`)

```go
// Package terminal implements the easter-egg command registry. Commands
// are executed server-side and render from internal/content, so terminal
// output and page content share one source of truth.
package terminal

import (
	"fmt"
	"html"
	"strings"

	"portfolio/internal/content"
)

// Result is the outcome of executing one command line.
type Result struct {
	HTML   string // safe HTML appended to the terminal output
	Action string // optional client directive: "clear", "exit", "goto:#contact", "open:/projects/<slug>"
}

type command struct {
	name string
	desc string
	run  func(args []string) Result
}

type Registry struct {
	cmds  map[string]*command
	order []string
}

func New() *Registry {
	r := &Registry{cmds: map[string]*command{}}
	r.add("help", "list available commands", func([]string) Result { return r.help() })
	r.add("whoami", "who is this guy?", whoami)
	r.add("projects", "list case studies", projects)
	r.add("open", "open <project> — jump to a case study", openProject)
	r.add("resume", "career timeline at a glance", resumeCmd)
	r.add("skills", "tech stack and proficiency", skillsCmd)
	r.add("contact", "how to reach me", contactCmd)
	r.add("clear", "clear the screen", func([]string) Result { return Result{Action: "clear"} })
	r.add("exit", "close the terminal", func([]string) Result { return Result{Action: "exit"} })
	r.add("sudo", "with great power…", sudoCmd)
	return r
}

func (r *Registry) add(name, desc string, run func([]string) Result) {
	r.cmds[name] = &command{name: name, desc: desc, run: run}
	r.order = append(r.order, name)
}

// Execute parses one input line and runs the matching command.
// User input is always HTML-escaped before echoing back.
func (r *Registry) Execute(input string) Result {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return Result{}
	}
	name := strings.ToLower(fields[0])
	c, ok := r.cmds[name]
	if !ok {
		return Result{HTML: errLine(fmt.Sprintf("command not found: %s — try 'help'", html.EscapeString(name)))}
	}
	return c.run(fields[1:])
}

func line(s string) string    { return `<div class="term-line">` + s + `</div>` }
func errLine(s string) string { return `<div class="term-line term-err">` + s + `</div>` }
func accent(s string) string  { return `<span class="term-accent">` + s + `</span>` }

func (r *Registry) help() Result {
	var b strings.Builder
	for _, name := range r.order {
		c := r.cmds[name]
		b.WriteString(line(accent(fmt.Sprintf("%-10s", c.name)) + " " + c.desc))
	}
	return Result{HTML: b.String()}
}

func whoami([]string) Result {
	p := content.Me
	return Result{HTML: line(accent(p.Name)+" — "+p.Title) + line(p.Pitch)}
}

func projects([]string) Result {
	var b strings.Builder
	for _, p := range content.Projects {
		b.WriteString(line(accent(p.Slug) + " — " + p.Tagline))
	}
	b.WriteString(line("run 'open <name>' to read a case study"))
	return Result{HTML: b.String()}
}

func openProject(args []string) Result {
	if len(args) == 0 {
		return Result{HTML: errLine("usage: open <project> — try 'projects' for the list")}
	}
	p, ok := content.ProjectBySlug(strings.ToLower(args[0]))
	if !ok {
		return Result{HTML: errLine("no such project: " + html.EscapeString(args[0]))}
	}
	return Result{HTML: line("opening " + accent(p.Slug) + "…"), Action: "open:/projects/" + p.Slug}
}

func resumeCmd([]string) Result {
	var b strings.Builder
	for _, e := range content.Resume {
		b.WriteString(line(accent(e.Period) + "  " + e.Org))
	}
	b.WriteString(line(`full pdf: <a class="term-accent underline" href="/static/resume/jasen-nicely-resume.pdf" download>resume.pdf</a>`))
	return Result{HTML: b.String()}
}

func skillsCmd([]string) Result {
	var b strings.Builder
	for _, s := range content.Skills {
		bar := strings.Repeat("█", s.Level/10) + strings.Repeat("░", 10-s.Level/10)
		b.WriteString(line(fmt.Sprintf("%-28s %s %d", s.Name, accent(bar), s.Level)))
	}
	return Result{HTML: b.String()}
}

func contactCmd([]string) Result {
	p := content.Me
	return Result{
		HTML: line(`email: <a class="term-accent underline" href="mailto:`+p.Email+`">`+p.Email+`</a>`) +
			line(`github: <a class="term-accent underline" href="`+p.GitHub+`">`+p.GitHubLabel+`</a>`) +
			line("or just use the form — taking you there now."),
		Action: "goto:#contact",
	}
}

func sudoCmd(args []string) Result {
	if len(args) > 0 && args[0] == "hire-me" {
		return Result{HTML: line("permission granted. routing you to the contact form…"), Action: "goto:#contact"}
	}
	return Result{HTML: errLine("visitor is not in the sudoers file. this incident will be reported.")}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/terminal/ -v`
Expected: PASS (8 tests)

- [ ] **Step 5: Commit**

```bash
git add internal/terminal
git commit -m "feat: terminal command registry with server-side execution"
```

---

### Task 5: Templates, embed, handler core (home + static)

**Files:**
- Create: `embed.go`, `templates/layout.html`, `templates/pages/home.html`, `templates/pages/project_page.html`, `templates/pages/404.html`, `templates/pages/500.html`, `templates/partials/project_card.html`, `templates/partials/project_detail.html`, `templates/partials/contact_form.html`, `templates/partials/contact_success.html`, `templates/partials/contact_error.html`, `templates/partials/terminal.html`, `internal/handler/handler.go`, `internal/handler/pages.go`
- Test: `internal/handler/handler_test.go`, `internal/handler/pages_test.go`

Note: this task creates stub handlers `contact` and `terminal` returning 501 so the mux compiles; Tasks 7–8 replace them. The 404/500 pages are created here (the page-parse loop needs them); their handlers are tested in Task 9.

- [ ] **Step 1: Write the failing tests**

`internal/handler/handler_test.go`:

```go
package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"portfolio"
	"portfolio/internal/store"
	"portfolio/internal/terminal"
)

// newTestServer builds a Server against the real embedded templates and
// static assets, with a throwaway SQLite store.
func newTestServer(t *testing.T) (*Server, *store.Store) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := New(portfolio.Templates, portfolio.Static, st, terminal.New(), log)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return srv, st
}

func get(t *testing.T, h http.Handler, path string, hdr map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHomePage(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Jasen Nicely", `id="projects"`, `id="contact"`, `id="resume"`, "Property Tax Pipeline", "Announcement System", "Redline"} {
		if !strings.Contains(body, want) {
			t.Errorf("home page missing %q", want)
		}
	}
}

func TestStaticAssetServed(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/static/js/htmx.min.js", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.Contains(cc, "max-age") {
		t.Errorf("Cache-Control = %q, want max-age", cc)
	}
}

func TestResumePDFServed(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/static/resume/jasen-nicely-resume.pdf", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/handler/`
Expected: FAIL — compile error, `undefined: New` (and package `portfolio` missing until embed.go exists)

- [ ] **Step 3: Create `embed.go`** (repo root)

```go
// Package portfolio embeds the site's templates and static assets so the
// build produces a single self-contained binary.
package portfolio

import "embed"

//go:embed templates
var Templates embed.FS

//go:embed static
var Static embed.FS
```

- [ ] **Step 4: Create `templates/layout.html`**

```html
{{define "layout"}}<!doctype html>
<html lang="en" class="scroll-smooth">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{if .Title}}{{.Title}} · Jasen Nicely{{else}}Jasen Nicely — Senior Full-Stack Software Developer{{end}}</title>
  <meta name="description" content="{{.Profile.Pitch}}">
  <link rel="icon" href="/static/favicon.svg" type="image/svg+xml">
  <link rel="stylesheet" href="/static/css/site.css">
  <script defer src="/static/js/htmx.min.js"></script>
  <script defer src="/static/js/alpine.min.js"></script>
  <script defer src="/static/js/grid.js"></script>
  <script defer src="/static/js/site.js"></script>
</head>
<body class="bg-ink-950 text-fog-200 font-sans antialiased">
{{template "main" .}}
{{template "terminal" .}}
</body>
</html>{{end}}
```

- [ ] **Step 5: Create `templates/pages/home.html`**

```html
{{define "main"}}
<header class="fixed inset-x-0 top-0 z-40 border-b border-ink-800/80 bg-ink-950/80 backdrop-blur">
  <nav class="mx-auto flex max-w-5xl items-center justify-between px-6 py-3">
    <a href="/" class="font-mono text-sm text-phosphor-400">jasen@gideon:~$</a>
    <div class="flex items-center gap-6 font-mono text-xs text-fog-400">
      <a href="#projects" class="transition hover:text-phosphor-300">./work</a>
      <a href="#resume" class="transition hover:text-phosphor-300">./resume</a>
      <a href="#contact" class="transition hover:text-phosphor-300">./contact</a>
      <a href="{{.Profile.GitHub}}" class="transition hover:text-phosphor-300">github ↗</a>
    </div>
  </nav>
</header>

<section class="relative overflow-hidden pt-32 pb-20">
  <canvas id="grid-bg" class="pointer-events-none absolute inset-0 h-full w-full"></canvas>
  <div class="relative mx-auto max-w-5xl px-6">
    <div class="rounded-xl border border-ink-700 bg-ink-900/90 shadow-2xl shadow-phosphor-500/5">
      <div class="flex items-center gap-2 border-b border-ink-800 px-4 py-2.5">
        <span class="h-3 w-3 rounded-full bg-red-500/70"></span>
        <span class="h-3 w-3 rounded-full bg-amber-glow/70"></span>
        <span class="h-3 w-3 rounded-full bg-phosphor-500/70"></span>
        <span class="ml-3 font-mono text-xs text-fog-400">jasen@gideon: ~</span>
      </div>
      <div class="px-6 py-8 sm:px-10">
        <p class="font-mono text-sm text-fog-400"><span class="text-phosphor-400">$</span> <span id="type-cmd" class="cursor"></span></p>
        <div data-hero-output class="mt-6 opacity-0 transition-opacity duration-700">
          <h1 class="text-4xl font-bold tracking-tight text-white sm:text-6xl">{{.Profile.Name}}</h1>
          <p class="mt-3 font-mono text-base text-phosphor-300 sm:text-lg">{{.Profile.Title}}</p>
          <p class="mt-5 max-w-2xl text-lg text-fog-400">{{.Profile.Pitch}}</p>
          <div class="mt-8 flex flex-wrap gap-4">
            <a href="#projects" class="rounded-md bg-phosphor-500 px-5 py-2.5 font-mono text-sm font-semibold text-ink-950 transition hover:bg-phosphor-400">./view-work</a>
            <a href="#contact" class="rounded-md border border-ink-700 px-5 py-2.5 font-mono text-sm text-fog-200 transition hover:border-phosphor-500 hover:text-phosphor-300">./hire-me</a>
          </div>
        </div>
      </div>
    </div>
    <button x-data @click="$dispatch('terminal-open')" class="mt-4 font-mono text-xs text-fog-400 transition hover:text-phosphor-300">press <kbd class="rounded border border-ink-700 px-1.5 py-0.5 text-phosphor-400">`</kbd> to open terminal</button>
  </div>
</section>

<section class="border-y border-ink-800 bg-ink-900/50">
  <div class="mx-auto grid max-w-5xl grid-cols-2 sm:grid-cols-4">
    {{range .Stats}}
    <div class="px-6 py-8" data-reveal>
      <p class="font-mono text-3xl font-semibold text-phosphor-400" data-countup>{{.Value}}</p>
      <p class="mt-2 text-sm text-fog-400">{{.Label}}</p>
    </div>
    {{end}}
  </div>
</section>

<section id="about" class="mx-auto max-w-5xl px-6 py-24">
  <h2 class="font-mono text-sm text-phosphor-400" data-reveal>$ cat about.md</h2>
  <div class="mt-6 max-w-3xl space-y-5 text-lg leading-relaxed" data-reveal>
    {{range .Profile.About}}<p>{{.}}</p>{{end}}
  </div>
</section>

<section id="projects" class="mx-auto max-w-5xl px-6 py-24">
  <h2 class="font-mono text-sm text-phosphor-400" data-reveal>$ ls ./work --sort=impact</h2>
  <div class="mt-8 space-y-6">
    {{range .Projects}}{{template "project_card" .}}{{end}}
  </div>
</section>

<section id="resume" class="border-t border-ink-800 bg-ink-900/40">
  <div class="mx-auto max-w-5xl px-6 py-24">
    <div class="flex flex-wrap items-center justify-between gap-4" data-reveal>
      <h2 class="font-mono text-sm text-phosphor-400">$ git log --career</h2>
      <a href="/static/resume/jasen-nicely-resume.pdf" download class="rounded-md border border-ink-700 px-4 py-2 font-mono text-xs transition hover:border-phosphor-500 hover:text-phosphor-300">⤓ download resume.pdf</a>
    </div>
    <div class="mt-10 grid gap-16 lg:grid-cols-[1fr_280px]">
      <ol class="relative space-y-10 border-l border-ink-700 pl-8">
        {{range .Resume}}
        <li class="relative" data-reveal>
          <span class="absolute -left-[37px] mt-2 h-2.5 w-2.5 rounded-full bg-phosphor-500"></span>
          <p class="font-mono text-xs text-fog-400">{{.Period}}</p>
          <h3 class="mt-1 text-lg font-semibold text-white">{{.Org}}</h3>
          {{with .Note}}<p class="font-mono text-xs text-amber-glow">{{.}}</p>{{end}}
          <ul class="mt-3 space-y-1.5 text-sm text-fog-400">
            {{range .Bullets}}<li class="flex gap-2"><span class="text-phosphor-500">▸</span><span>{{.}}</span></li>{{end}}
          </ul>
        </li>
        {{end}}
      </ol>
      <div data-reveal>
        <h3 class="font-mono text-xs uppercase tracking-widest text-fog-400">tech stack</h3>
        <ul class="mt-4 space-y-3">
          {{range .Skills}}
          <li>
            <div class="flex justify-between font-mono text-xs"><span>{{.Name}}</span><span class="text-fog-400">{{.Level}}</span></div>
            <div class="mt-1 h-1 rounded bg-ink-700"><div class="h-1 rounded bg-phosphor-500" style="width: {{.Level}}%"></div></div>
          </li>
          {{end}}
        </ul>
        <h3 class="mt-10 font-mono text-xs uppercase tracking-widest text-fog-400">education</h3>
        <p class="mt-3 text-sm">{{.School.School}}</p>
        <p class="text-xs text-fog-400">{{.School.Place}} · {{.School.Year}}</p>
        <p class="text-xs text-fog-400">{{.School.Detail}}</p>
      </div>
    </div>
  </div>
</section>

<section id="contact" class="mx-auto max-w-5xl px-6 py-24">
  <h2 class="font-mono text-sm text-phosphor-400" data-reveal>$ mail -s "let's talk"</h2>
  <div class="mt-8 grid gap-12 lg:grid-cols-[1fr_320px]">
    <div data-reveal>{{template "contact_form" .Form}}</div>
    <div class="space-y-4 font-mono text-sm" data-reveal>
      <p class="text-fog-400">Prefer it direct?</p>
      <a class="block text-phosphor-300 transition hover:text-phosphor-400" href="mailto:{{.Profile.Email}}">{{.Profile.Email}}</a>
      <a class="block text-phosphor-300 transition hover:text-phosphor-400" href="{{.Profile.GitHub}}">{{.Profile.GitHubLabel}} ↗</a>
      <p class="text-fog-400">{{.Profile.Location}}</p>
    </div>
  </div>
</section>

<footer class="border-t border-ink-800 py-8">
  <p class="mx-auto max-w-5xl px-6 font-mono text-xs text-fog-400">© {{.Year}} {{.Profile.Name}} · built with Go, htmx, Alpine.js — <span class="text-phosphor-400">same stack as Redline</span></p>
</footer>
{{end}}
```

- [ ] **Step 6: Create the project partials**

`templates/partials/project_card.html`:

```html
{{define "project_card"}}
<article id="proj-{{.Slug}}" class="group rounded-xl border border-ink-700 bg-ink-900/60 p-6 transition hover:border-phosphor-500/50 sm:p-8" data-reveal>
  <div class="flex flex-wrap items-start justify-between gap-4">
    <div>
      <h3 class="text-xl font-semibold text-white">{{.Name}}</h3>
      <p class="mt-1 font-mono text-sm text-phosphor-300">{{.Tagline}}</p>
    </div>
    <div class="flex gap-6">
      {{range .Metrics}}<div class="text-right"><p class="font-mono text-lg font-semibold text-phosphor-400">{{.Value}}</p><p class="text-xs text-fog-400">{{.Label}}</p></div>{{end}}
    </div>
  </div>
  <p class="mt-4 max-w-3xl text-fog-400">{{.Summary}}</p>
  <div class="mt-5 flex flex-wrap gap-2">
    {{range .Tech}}<span class="rounded border border-ink-700 px-2 py-0.5 font-mono text-xs text-fog-400">{{.}}</span>{{end}}
  </div>
  <div class="mt-6 flex items-center gap-4 font-mono text-sm">
    <button data-project-link="{{.Slug}}" hx-get="/projects/{{.Slug}}" hx-target="#proj-{{.Slug}}" hx-swap="outerHTML show:top" hx-push-url="/projects/{{.Slug}}" class="text-phosphor-400 transition hover:text-phosphor-300">cat case-study.md →</button>
    {{if .Link}}<a href="{{.Link}}" class="text-fog-400 transition hover:text-phosphor-300">{{.LinkLabel}} ↗</a>{{end}}
  </div>
</article>
{{end}}
```

`templates/partials/project_detail.html`:

```html
{{define "project_detail"}}
<article id="proj-{{.Slug}}" class="rounded-xl border border-phosphor-500/40 bg-ink-900/80 p-6 sm:p-8">
  <div class="flex items-start justify-between gap-4">
    <div>
      <p class="font-mono text-xs text-fog-400">$ cat ./work/{{.Slug}}/case-study.md</p>
      <h3 class="mt-2 text-2xl font-semibold text-white">{{.Name}}</h3>
      <p class="mt-1 font-mono text-sm text-phosphor-300">{{.Tagline}}</p>
    </div>
    <button hx-get="/projects/{{.Slug}}/card" hx-target="#proj-{{.Slug}}" hx-swap="outerHTML" hx-push-url="/" class="font-mono text-sm text-fog-400 transition hover:text-phosphor-300" aria-label="collapse case study">[x] close</button>
  </div>
  <div class="mt-6 flex flex-wrap gap-8">
    {{range .Metrics}}<div><p class="font-mono text-2xl font-semibold text-phosphor-400">{{.Value}}</p><p class="text-xs text-fog-400">{{.Label}}</p></div>{{end}}
  </div>
  <div class="mt-8 max-w-3xl space-y-8">
    <section><h4 class="font-mono text-xs uppercase tracking-widest text-amber-glow">problem</h4>{{range .Problem}}<p class="mt-3 leading-relaxed">{{.}}</p>{{end}}</section>
    <section><h4 class="font-mono text-xs uppercase tracking-widest text-amber-glow">approach</h4>{{range .Approach}}<p class="mt-3 leading-relaxed">{{.}}</p>{{end}}</section>
    <section><h4 class="font-mono text-xs uppercase tracking-widest text-amber-glow">outcome</h4>{{range .Outcome}}<p class="mt-3 leading-relaxed">{{.}}</p>{{end}}</section>
  </div>
  <div class="mt-8 flex flex-wrap items-center gap-2">
    {{range .Tech}}<span class="rounded border border-ink-700 px-2 py-0.5 font-mono text-xs text-fog-400">{{.}}</span>{{end}}
    {{if .Link}}<a href="{{.Link}}" class="ml-auto font-mono text-sm text-phosphor-400 transition hover:text-phosphor-300">{{.LinkLabel}} ↗</a>{{end}}
  </div>
</article>
{{end}}
```

- [ ] **Step 7: Create the contact partials**

`templates/partials/contact_form.html` (context: `contactForm` value):

```html
{{define "contact_form"}}
<form hx-post="/contact" hx-swap="outerHTML" novalidate class="space-y-5">
  <div class="grid gap-5 sm:grid-cols-2">
    <label class="block">
      <span class="font-mono text-xs text-fog-400">name *</span>
      <input name="name" value="{{.Name}}" class="mt-1.5 w-full rounded-md border border-ink-700 bg-ink-900 px-3 py-2 outline-none transition focus:border-phosphor-500">
      {{with .Errors.name}}<p class="mt-1 font-mono text-xs text-red-400">! {{.}}</p>{{end}}
    </label>
    <label class="block">
      <span class="font-mono text-xs text-fog-400">email *</span>
      <input name="email" type="email" value="{{.Email}}" class="mt-1.5 w-full rounded-md border border-ink-700 bg-ink-900 px-3 py-2 outline-none transition focus:border-phosphor-500">
      {{with .Errors.email}}<p class="mt-1 font-mono text-xs text-red-400">! {{.}}</p>{{end}}
    </label>
  </div>
  <div class="grid gap-5 sm:grid-cols-2">
    <label class="block">
      <span class="font-mono text-xs text-fog-400">company</span>
      <input name="company" value="{{.Company}}" class="mt-1.5 w-full rounded-md border border-ink-700 bg-ink-900 px-3 py-2 outline-none transition focus:border-phosphor-500">
    </label>
    <label class="block">
      <span class="font-mono text-xs text-fog-400">this is about *</span>
      <select name="kind" class="mt-1.5 w-full rounded-md border border-ink-700 bg-ink-900 px-3 py-2 outline-none transition focus:border-phosphor-500">
        <option value="hiring" {{if eq .Kind "hiring"}}selected{{end}}>hiring me full-time</option>
        <option value="contract" {{if eq .Kind "contract"}}selected{{end}}>contract / freelance work</option>
        <option value="other" {{if eq .Kind "other"}}selected{{end}}>something else</option>
      </select>
      {{with .Errors.kind}}<p class="mt-1 font-mono text-xs text-red-400">! {{.}}</p>{{end}}
    </label>
  </div>
  <label class="block">
    <span class="font-mono text-xs text-fog-400">message *</span>
    <textarea name="message" rows="5" class="mt-1.5 w-full rounded-md border border-ink-700 bg-ink-900 px-3 py-2 outline-none transition focus:border-phosphor-500">{{.Message}}</textarea>
    {{with .Errors.message}}<p class="mt-1 font-mono text-xs text-red-400">! {{.}}</p>{{end}}
  </label>
  <button type="submit" class="rounded-md bg-phosphor-500 px-6 py-2.5 font-mono text-sm font-semibold text-ink-950 transition hover:bg-phosphor-400">./send-message</button>
</form>
{{end}}
```

`templates/partials/contact_success.html` (context: `contactForm`):

```html
{{define "contact_success"}}
<div class="rounded-md border border-phosphor-500/40 bg-ink-900 p-6 font-mono text-sm">
  <p class="text-phosphor-400">✓ message queued</p>
  <p class="mt-2 text-fog-400">Thanks, {{.Name}} — I'll get back to you within 24 hours.</p>
</div>
{{end}}
```

`templates/partials/contact_error.html` (context: `content.Profile`):

```html
{{define "contact_error"}}
<div class="rounded-md border border-red-400/40 bg-ink-900 p-6 font-mono text-sm">
  <p class="text-red-400">✗ something broke on my end</p>
  <p class="mt-2 text-fog-400">Your message wasn't saved. Email me directly at <a class="text-phosphor-300 underline" href="mailto:{{.Email}}">{{.Email}}</a> — I'd hate to miss you.</p>
</div>
{{end}}
```

- [ ] **Step 8: Create `templates/partials/terminal.html`**

```html
{{define "terminal"}}
<div x-data="{ open: false }"
     @keydown.window="if ($event.key === '`' && !['INPUT','TEXTAREA','SELECT'].includes($event.target.tagName)) { $event.preventDefault(); open = !open; if (open) $nextTick(() => $refs.cmd.focus()) }"
     @terminal-open.window="open = true; $nextTick(() => $refs.cmd.focus())"
     @terminal-close.window="open = false"
     @keydown.escape.window="open = false">
  <div x-show="open" x-cloak
       x-transition:enter="transition ease-out duration-200"
       x-transition:enter-start="translate-y-full"
       x-transition:enter-end="translate-y-0"
       x-transition:leave="transition ease-in duration-150"
       x-transition:leave-start="translate-y-0"
       x-transition:leave-end="translate-y-full"
       class="fixed inset-x-0 bottom-0 z-50 mx-auto max-w-4xl px-4 pb-4" role="dialog" aria-label="interactive terminal">
    <div class="overflow-hidden rounded-xl border border-phosphor-500/30 bg-ink-950/95 shadow-2xl shadow-phosphor-500/10 backdrop-blur">
      <div class="flex items-center justify-between border-b border-ink-800 px-4 py-2">
        <span class="font-mono text-xs text-fog-400">jasen@gideon: ~ — visitor session</span>
        <button @click="open = false" class="font-mono text-xs text-fog-400 transition hover:text-phosphor-300" aria-label="close terminal">[x]</button>
      </div>
      <div id="term-output" class="h-72 overflow-y-auto px-4 py-3 font-mono text-sm leading-relaxed">
        <div class="term-line text-fog-400">jasen.sh v1.0 — type <span class="term-accent">help</span> to get started</div>
      </div>
      <form id="terminal-form" hx-post="/terminal" hx-target="#term-output" hx-swap="beforeend"
            class="flex items-center gap-2 border-t border-ink-800 px-4 py-3">
        <span class="font-mono text-sm text-phosphor-400">$</span>
        <input x-ref="cmd" name="cmd" autocomplete="off" autocapitalize="off" spellcheck="false" maxlength="200"
               class="w-full bg-transparent font-mono text-sm outline-none" aria-label="terminal command input">
      </form>
    </div>
  </div>
</div>
{{end}}
```

- [ ] **Step 9: Create the standalone and error pages**

`templates/pages/project_page.html`:

```html
{{define "main"}}
<div class="mx-auto max-w-5xl px-6 pt-16 pb-24">
  <a href="/#projects" class="font-mono text-sm text-fog-400 transition hover:text-phosphor-300">← cd ~/work</a>
  <div class="mt-6">{{template "project_detail" .Project}}</div>
</div>
{{end}}
```

`templates/pages/404.html`:

```html
{{define "main"}}
<div class="mx-auto flex min-h-screen max-w-3xl flex-col justify-center px-6">
  <div class="rounded-xl border border-ink-700 bg-ink-900/90 p-8 font-mono text-sm">
    <p class="text-fog-400">$ GET {{.Path}}</p>
    <p class="mt-3 text-red-400">command not found: 404</p>
    <p class="mt-3 text-fog-400">try one of these instead:</p>
    <p class="mt-2 space-x-4"><a class="text-phosphor-400 transition hover:text-phosphor-300" href="/">cd ~</a><a class="text-phosphor-400 transition hover:text-phosphor-300" href="/#projects">ls ./work</a><a class="text-phosphor-400 transition hover:text-phosphor-300" href="/#contact">mail jasen</a></p>
  </div>
</div>
{{end}}
```

`templates/pages/500.html`:

```html
{{define "main"}}
<div class="mx-auto flex min-h-screen max-w-3xl flex-col justify-center px-6">
  <div class="rounded-xl border border-red-400/40 bg-ink-900/90 p-8 font-mono text-sm">
    <p class="text-red-400">kernel panic — something broke on my end</p>
    <p class="mt-3 text-fog-400">It's not you. Email me at <a class="text-phosphor-300 underline" href="mailto:{{.Profile.Email}}">{{.Profile.Email}}</a> and I'll look into it.</p>
  </div>
</div>
{{end}}
```

- [ ] **Step 10: Write `internal/handler/handler.go`**

```go
// Package handler wires HTTP routes to templates, the content package,
// the inquiry store, and the terminal registry.
package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"portfolio/internal/content"
	"portfolio/internal/store"
	"portfolio/internal/terminal"
)

type Server struct {
	pages    map[string]*template.Template
	partials *template.Template
	st       *store.Store
	reg      *terminal.Registry
	log      *slog.Logger
	h        http.Handler
}

// pageData is the single context type passed to full-page templates.
type pageData struct {
	Title    string
	Path     string
	Year     int
	Profile  content.Profile
	Stats    []content.Stat
	Projects []content.Project
	Resume   []content.ResumeEntry
	Skills   []content.Skill
	School   content.Education
	Project  *content.Project // set on the standalone project page
	Form     contactForm
}

func New(templatesFS, staticFS fs.FS, st *store.Store, reg *terminal.Registry, log *slog.Logger) (*Server, error) {
	s := &Server{pages: map[string]*template.Template{}, st: st, reg: reg, log: log}

	for _, page := range []string{"home.html", "project_page.html", "404.html", "500.html"} {
		t, err := template.ParseFS(templatesFS,
			"templates/layout.html", "templates/partials/*.html", "templates/pages/"+page)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", page, err)
		}
		s.pages[page] = t
	}
	partials, err := template.ParseFS(templatesFS, "templates/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse partials: %w", err)
	}
	s.partials = partials

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.home)
	mux.HandleFunc("GET /projects/{slug}", s.project)
	mux.HandleFunc("GET /projects/{slug}/card", s.projectCard)
	mux.HandleFunc("POST /contact", s.contact)
	mux.HandleFunc("POST /terminal", s.terminal)
	mux.Handle("GET /static/", staticHandler(staticFS))
	mux.HandleFunc("/", s.notFound)

	s.h = recoverPanics(log, s.serverError, mux)
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.h.ServeHTTP(w, r) }

func (s *Server) data() pageData {
	return pageData{
		Year:     time.Now().Year(),
		Profile:  content.Me,
		Stats:    content.Stats,
		Projects: content.Projects,
		Resume:   content.Resume,
		Skills:   content.Skills,
		School:   content.School,
	}
}

// renderPage buffers template execution so a render error can still
// produce a clean 500 instead of a half-written page.
func (s *Server) renderPage(w http.ResponseWriter, status int, page string, d pageData) {
	t, ok := s.pages[page]
	if !ok {
		s.log.Error("unknown page template", "page", page)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", d); err != nil {
		s.log.Error("render page", "page", page, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	buf.WriteTo(w)
}

func (s *Server) renderPartial(w http.ResponseWriter, name string, d any) {
	var buf bytes.Buffer
	if err := s.partials.ExecuteTemplate(&buf, name, d); err != nil {
		s.log.Error("render partial", "partial", name, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func staticHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		fileServer.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 11: Write `internal/handler/pages.go`** (with temporary stubs for contact/terminal)

```go
package handler

import (
	"net/http"

	"portfolio/internal/content"
)

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	s.renderPage(w, http.StatusOK, "home.html", s.data())
}

func (s *Server) project(w http.ResponseWriter, r *http.Request) {
	p, ok := content.ProjectBySlug(r.PathValue("slug"))
	if !ok {
		s.notFound(w, r)
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		s.renderPartial(w, "project_detail", p)
		return
	}
	d := s.data()
	d.Title = p.Name
	d.Project = &p
	s.renderPage(w, http.StatusOK, "project_page.html", d)
}

func (s *Server) projectCard(w http.ResponseWriter, r *http.Request) {
	p, ok := content.ProjectBySlug(r.PathValue("slug"))
	if !ok {
		s.notFound(w, r)
		return
	}
	s.renderPartial(w, "project_card", p)
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	d := s.data()
	d.Title = "404"
	d.Path = r.URL.Path
	s.renderPage(w, http.StatusNotFound, "404.html", d)
}

func (s *Server) serverError(w http.ResponseWriter, r *http.Request) {
	d := s.data()
	d.Title = "500"
	s.renderPage(w, http.StatusInternalServerError, "500.html", d)
}
```

Temporary stubs so the mux compiles (replaced in Tasks 7–8). Put them in `pages.go` for now; Tasks 7–8 move them to their own files:

```go
// TEMPORARY until Task 7/8.
func (s *Server) contact(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (s *Server) terminal(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
```

And `internal/handler/middleware.go` (tested in Task 9, needed now to compile):

```go
package handler

import (
	"log/slog"
	"net/http"
)

// recoverPanics converts a handler panic into a logged, styled 500
// instead of a dropped connection.
func recoverPanics(log *slog.Logger, on500 http.HandlerFunc, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic recovered", "path", r.URL.Path, "panic", rec)
				on500(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
```

Also create the `contactForm` type now (in `internal/handler/contact.go`, validation added in Task 7):

```go
package handler

// contactForm carries submitted values and per-field validation errors
// back into the form partial.
type contactForm struct {
	Name    string
	Email   string
	Company string
	Kind    string
	Message string
	Errors  map[string]string
}
```

- [ ] **Step 12: Run tests to verify they pass**

Run: `go test ./internal/handler/ -v`
Expected: PASS — TestHomePage, TestStaticAssetServed, TestResumePDFServed

- [ ] **Step 13: Commit**

```bash
git add embed.go templates internal/handler
git commit -m "feat: templates, embedded assets, handler core with home page and static serving"
```

---

### Task 6: Project routes (fragment vs standalone page)

**Files:**
- Test: `internal/handler/pages_test.go` (handlers already written in Task 5 — this task locks behavior in with tests)

- [ ] **Step 1: Write the tests** (`internal/handler/pages_test.go`)

```go
package handler

import (
	"net/http"
	"strings"
	"testing"
)

func TestProjectFragmentForHTMX(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/projects/redline", map[string]string{"HX-Request": "true"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `id="proj-redline"`) {
		t.Error("fragment missing project article")
	}
	if strings.Contains(body, "<!doctype") {
		t.Error("fragment contains full page doctype")
	}
}

func TestProjectStandalonePage(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/projects/property-tax-pipeline", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<!doctype") {
		t.Error("standalone page missing doctype")
	}
	if !strings.Contains(body, "Property Tax Pipeline") {
		t.Error("standalone page missing project name")
	}
}

func TestProjectUnknownSlugIs404(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/projects/nope", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestProjectCardFragment(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/projects/redline/card", map[string]string{"HX-Request": "true"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `data-project-link="redline"`) {
		t.Error("card fragment missing expand button")
	}
	if strings.Contains(body, "<!doctype") {
		t.Error("card fragment contains full page doctype")
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/handler/ -v`
Expected: PASS — implementation already exists from Task 5; if any test fails, fix `pages.go` until green.

- [ ] **Step 3: Commit**

```bash
git add internal/handler/pages_test.go
git commit -m "test: project fragment/page split and card collapse route"
```

---

### Task 7: Contact form handler

**Files:**
- Modify: `internal/handler/contact.go` (replace stub from Task 5; remove stub from `pages.go`)
- Test: `internal/handler/contact_test.go`

- [ ] **Step 1: Write the failing tests** (`internal/handler/contact_test.go`)

```go
package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func postForm(t *testing.T, h http.Handler, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestContactValidSubmission(t *testing.T) {
	srv, st := newTestServer(t)
	rec := postForm(t, srv, "/contact", url.Values{
		"name": {"Ada Lovelace"}, "email": {"ada@example.com"},
		"company": {"Analytical Engines"}, "kind": {"contract"}, "message": {"Need a compiler."},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "message queued") {
		t.Error("success fragment missing confirmation")
	}
	got, err := st.ListInquiries()
	if err != nil || len(got) != 1 {
		t.Fatalf("inquiries = %d (%v), want 1", len(got), err)
	}
	if got[0].Name != "Ada Lovelace" || got[0].Kind != "contract" {
		t.Errorf("saved inquiry mismatch: %+v", got[0])
	}
}

func TestContactValidationErrors(t *testing.T) {
	srv, st := newTestServer(t)
	rec := postForm(t, srv, "/contact", url.Values{"kind": {"hiring"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (form re-render)", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"tell me who you are", "an email is required", "say something"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing validation message %q", want)
		}
	}
	if got, _ := st.ListInquiries(); len(got) != 0 {
		t.Errorf("invalid submission was saved: %d rows", len(got))
	}
}

func TestContactPreservesValuesOnError(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := postForm(t, srv, "/contact", url.Values{
		"name": {"Ada Lovelace"}, "email": {"not-an-email"}, "kind": {"other"}, "message": {"hi"},
	})
	body := rec.Body.String()
	if !strings.Contains(body, `value="Ada Lovelace"`) {
		t.Error("name value not preserved on re-render")
	}
	if !strings.Contains(body, "valid email") {
		t.Error("missing email format error")
	}
}

func TestValidate(t *testing.T) {
	valid := contactForm{Name: "n", Email: "e@x.com", Kind: "hiring", Message: "m"}
	tests := []struct {
		name   string
		mutate func(*contactForm)
		field  string
	}{
		{"missing name", func(f *contactForm) { f.Name = "" }, "name"},
		{"missing email", func(f *contactForm) { f.Email = "" }, "email"},
		{"bad email", func(f *contactForm) { f.Email = "nope" }, "email"},
		{"bad kind", func(f *contactForm) { f.Kind = "spam" }, "kind"},
		{"missing message", func(f *contactForm) { f.Message = "" }, "message"},
		{"huge message", func(f *contactForm) { f.Message = strings.Repeat("x", 5001) }, "message"},
	}
	if errs := validate(valid); len(errs) != 0 {
		t.Errorf("valid form has errors: %v", errs)
	}
	for _, tc := range tests {
		f := valid
		tc.mutate(&f)
		if errs := validate(f); errs[tc.field] == "" {
			t.Errorf("%s: no error on field %q (got %v)", tc.name, tc.field, errs)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/handler/ -run TestContact -v`
Expected: FAIL — stub returns 501, `validate` undefined (compile error)

- [ ] **Step 3: Implement** — replace `internal/handler/contact.go` entirely and delete the `contact` stub from `pages.go`:

```go
package handler

import (
	"net/http"
	"net/mail"
	"strings"

	"portfolio/internal/content"
	"portfolio/internal/store"
)

// contactForm carries submitted values and per-field validation errors
// back into the form partial.
type contactForm struct {
	Name    string
	Email   string
	Company string
	Kind    string
	Message string
	Errors  map[string]string
}

func (s *Server) contact(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	f := contactForm{
		Name:    strings.TrimSpace(r.PostFormValue("name")),
		Email:   strings.TrimSpace(r.PostFormValue("email")),
		Company: strings.TrimSpace(r.PostFormValue("company")),
		Kind:    r.PostFormValue("kind"),
		Message: strings.TrimSpace(r.PostFormValue("message")),
	}
	f.Errors = validate(f)
	if len(f.Errors) > 0 {
		// 200, not 422: this is a UX state and htmx should swap it in.
		s.renderPartial(w, "contact_form", f)
		return
	}
	_, err := s.st.SaveInquiry(store.Inquiry{
		Name: f.Name, Email: f.Email, Company: f.Company, Kind: f.Kind, Message: f.Message,
	})
	if err != nil {
		s.log.Error("save inquiry", "err", err)
		s.renderPartial(w, "contact_error", content.Me)
		return
	}
	s.renderPartial(w, "contact_success", f)
}

func validate(f contactForm) map[string]string {
	errs := map[string]string{}
	if f.Name == "" {
		errs["name"] = "tell me who you are"
	}
	switch {
	case f.Email == "":
		errs["email"] = "an email is required — it's how I reply"
	default:
		if _, err := mail.ParseAddress(f.Email); err != nil {
			errs["email"] = "that doesn't look like a valid email"
		}
	}
	switch f.Kind {
	case "hiring", "contract", "other":
	default:
		errs["kind"] = "pick a topic"
	}
	switch {
	case f.Message == "":
		errs["message"] = "say something — even just hi"
	case len(f.Message) > 5000:
		errs["message"] = "keep it under 5000 characters"
	}
	return errs
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/handler/ -v`
Expected: PASS (all handler tests)

- [ ] **Step 5: Commit**

```bash
git add internal/handler
git commit -m "feat: contact form handler with server-side validation and sqlite persistence"
```

---

### Task 8: Terminal HTTP handler

**Files:**
- Create: `internal/handler/terminal.go` (and delete the `terminal` stub from `pages.go`)
- Test: `internal/handler/terminal_test.go`

- [ ] **Step 1: Write the failing tests** (`internal/handler/terminal_test.go`)

```go
package handler

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestTerminalWhoami(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := postForm(t, srv, "/terminal", url.Values{"cmd": {"whoami"}})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "term-echo") || !strings.Contains(body, "whoami") {
		t.Error("response missing echo line")
	}
	if !strings.Contains(body, "Jasen Nicely") {
		t.Error("response missing whoami output")
	}
}

func TestTerminalActionHeader(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := postForm(t, srv, "/terminal", url.Values{"cmd": {"clear"}})
	trigger := rec.Header().Get("HX-Trigger")
	if !strings.Contains(trigger, "term-action") || !strings.Contains(trigger, "clear") {
		t.Errorf("HX-Trigger = %q, want term-action clear", trigger)
	}
}

func TestTerminalEchoEscapesInput(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := postForm(t, srv, "/terminal", url.Values{"cmd": {"<img src=x onerror=alert(1)>"}})
	if strings.Contains(rec.Body.String(), "<img") {
		t.Error("unescaped input echoed back")
	}
}

func TestTerminalUnknownCommand(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := postForm(t, srv, "/terminal", url.Values{"cmd": {"frobnicate"}})
	if !strings.Contains(rec.Body.String(), "command not found") {
		t.Error("missing command-not-found output")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/handler/ -run TestTerminal -v`
Expected: FAIL — stub returns 501

- [ ] **Step 3: Implement `internal/handler/terminal.go`** (delete the stub in `pages.go`)

```go
package handler

import (
	"fmt"
	"html"
	"net/http"
)

func (s *Server) terminal(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	cmd := r.PostFormValue("cmd")
	if len(cmd) > 200 {
		cmd = cmd[:200]
	}
	res := s.reg.Execute(cmd)
	if res.Action != "" {
		w.Header().Set("HX-Trigger", fmt.Sprintf(`{"term-action": %q}`, res.Action))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<div class="term-line term-echo"><span class="term-prompt">$</span> %s</div>%s`,
		html.EscapeString(cmd), res.HTML)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/handler/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/handler
git commit -m "feat: terminal http endpoint with HX-Trigger actions"
```

---

### Task 9: 404 page and panic recovery

**Files:**
- Test: `internal/handler/middleware_test.go` (middleware itself written in Task 5)

- [ ] **Step 1: Write the tests** (`internal/handler/middleware_test.go`)

```go
package handler

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNotFoundPage(t *testing.T) {
	srv, _ := newTestServer(t)
	rec := get(t, srv, "/no/such/path", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "command not found: 404") {
		t.Error("404 page missing terminal gag")
	}
	if !strings.Contains(body, "/no/such/path") {
		t.Error("404 page missing requested path")
	}
}

func TestRecoverPanics(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") })
	on500 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "kernel panic")
	}
	h := recoverPanics(log, on500, next)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "kernel panic") {
		t.Error("500 body not rendered")
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/handler/ -v`
Expected: PASS — implementation exists from Task 5; fix `pages.go`/`middleware.go` if not.

- [ ] **Step 3: Commit**

```bash
git add internal/handler/middleware_test.go
git commit -m "test: 404 page and panic recovery middleware"
```

---

### Task 10: Frontend — Tailwind theme, canvas grid, motion JS

**Files:**
- Create: `tailwind.css`, `static/js/grid.js`, `static/js/site.js`
- Build artifact: `static/css/site.css` (committed)

- [ ] **Step 1: Create `tailwind.css`** (repo root)

```css
@import "tailwindcss";

@source "./templates";

@theme {
  --font-sans: "IBM Plex Sans", ui-sans-serif, system-ui, sans-serif;
  --font-mono: "IBM Plex Mono", ui-monospace, "Cascadia Mono", monospace;

  --color-ink-950: #09090b;
  --color-ink-900: #111113;
  --color-ink-800: #1b1b1f;
  --color-ink-700: #2a2a30;
  --color-fog-400: #9d9da7;
  --color-fog-200: #d4d4dc;
  --color-phosphor-300: #86efac;
  --color-phosphor-400: #4ade80;
  --color-phosphor-500: #22c55e;
  --color-amber-glow: #fbbf24;
}

@font-face { font-family: "IBM Plex Sans"; src: url("/static/fonts/plex-sans-400.woff2") format("woff2"); font-weight: 400; font-display: swap; }
@font-face { font-family: "IBM Plex Sans"; src: url("/static/fonts/plex-sans-600.woff2") format("woff2"); font-weight: 600; font-display: swap; }
@font-face { font-family: "IBM Plex Sans"; src: url("/static/fonts/plex-sans-700.woff2") format("woff2"); font-weight: 700; font-display: swap; }
@font-face { font-family: "IBM Plex Mono"; src: url("/static/fonts/plex-mono-400.woff2") format("woff2"); font-weight: 400; font-display: swap; }
@font-face { font-family: "IBM Plex Mono"; src: url("/static/fonts/plex-mono-500.woff2") format("woff2"); font-weight: 500; font-display: swap; }

/* Alpine: hide x-cloak elements until Alpine initializes */
[x-cloak] { display: none !important; }

/* Terminal output */
.term-line { white-space: pre-wrap; }
.term-echo { color: var(--color-fog-400); }
.term-err { color: #f87171; }
.term-prompt, .term-accent { color: var(--color-phosphor-400); }

/* Blinking block cursor for the hero typing animation */
.cursor::after {
  content: "▋";
  color: var(--color-phosphor-400);
  animation: blink 1.1s steps(1) infinite;
}
@keyframes blink { 50% { opacity: 0; } }

/* Scroll-reveal: elements start hidden, site.js adds .revealed */
[data-reveal] { opacity: 0; transform: translateY(14px); transition: opacity .6s ease, transform .6s ease; }
[data-reveal].revealed { opacity: 1; transform: none; }

@media (prefers-reduced-motion: reduce) {
  html { scroll-behavior: auto; }
  [data-reveal] { opacity: 1; transform: none; transition: none; }
  .cursor::after { animation: none; }
}
```

- [ ] **Step 2: Create `static/js/grid.js`**

```js
// Phosphor dot-grid background for the hero. A soft green glow drifts
// across a grid of dots. Renders a single static frame when the user
// prefers reduced motion.
(function () {
  var canvas = document.getElementById("grid-bg");
  if (!canvas) return;
  var ctx = canvas.getContext("2d");
  var reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  var dpr = window.devicePixelRatio || 1;
  var w, h, t = 0;

  function resize() {
    w = canvas.width = canvas.offsetWidth * dpr;
    h = canvas.height = canvas.offsetHeight * dpr;
  }
  window.addEventListener("resize", resize);
  resize();

  var GAP = 28 * dpr;

  function frame() {
    ctx.clearRect(0, 0, w, h);
    var cx = w / 2 + Math.cos(t / 600) * w * 0.25;
    var cy = h * 0.4 + Math.sin(t / 800) * h * 0.15;
    for (var x = GAP / 2; x < w; x += GAP) {
      for (var y = GAP / 2; y < h; y += GAP) {
        var d = Math.hypot(x - cx, y - cy);
        var a = Math.max(0, 1 - d / (w * 0.45));
        if (a <= 0.05) continue;
        ctx.fillStyle = "rgba(74, 222, 128, " + (a * 0.35).toFixed(3) + ")";
        ctx.beginPath();
        ctx.arc(x, y, 1.1 * dpr, 0, Math.PI * 2);
        ctx.fill();
      }
    }
    t++;
    if (!reduced) requestAnimationFrame(frame);
  }
  frame();
})();
```

- [ ] **Step 3: Create `static/js/site.js`**

```js
// Hero typing, stat count-up, scroll reveal, and terminal overlay glue.
(function () {
  var reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  // --- hero typing -------------------------------------------------------
  var cmd = document.getElementById("type-cmd");
  var output = document.querySelector("[data-hero-output]");
  function revealHero() {
    if (output) output.classList.remove("opacity-0");
  }
  if (cmd && output) {
    var text = "whoami";
    if (reduced) {
      cmd.textContent = text;
      revealHero();
    } else {
      var i = 0;
      setTimeout(function tick() {
        cmd.textContent = text.slice(0, ++i);
        if (i < text.length) setTimeout(tick, 90 + Math.random() * 70);
        else setTimeout(revealHero, 250);
      }, 500);
    }
  }

  // --- stat count-up -----------------------------------------------------
  // Animates the numeric prefix of strings like "99.9%", "30+", "15 yrs".
  var counters = document.querySelectorAll("[data-countup]");
  function parseStat(s) {
    var m = s.match(/^([\d.]+)(.*)$/);
    if (!m) return null;
    return { n: parseFloat(m[1]), suffix: m[2], decimals: (m[1].split(".")[1] || "").length };
  }
  if (!reduced && "IntersectionObserver" in window && counters.length) {
    var io = new IntersectionObserver(function (entries) {
      entries.forEach(function (entry) {
        if (!entry.isIntersecting) return;
        io.unobserve(entry.target);
        var p = parseStat(entry.target.textContent.trim());
        if (!p) return;
        var start = performance.now();
        (function step(now) {
          var t = Math.min(1, (now - start) / 900);
          var eased = 1 - Math.pow(1 - t, 3);
          entry.target.textContent = (p.n * eased).toFixed(p.decimals) + p.suffix;
          if (t < 1) requestAnimationFrame(step);
        })(start);
      });
    }, { threshold: 0.4 });
    counters.forEach(function (el) { io.observe(el); });
  }

  // --- scroll reveal -----------------------------------------------------
  var revealEls = document.querySelectorAll("[data-reveal]");
  if (reduced || !("IntersectionObserver" in window)) {
    revealEls.forEach(function (el) { el.classList.add("revealed"); });
  } else {
    var io2 = new IntersectionObserver(function (entries) {
      entries.forEach(function (e) {
        if (e.isIntersecting) {
          e.target.classList.add("revealed");
          io2.unobserve(e.target);
        }
      });
    }, { threshold: 0.15 });
    revealEls.forEach(function (el) { io2.observe(el); });
  }

  // --- terminal glue -----------------------------------------------------
  // Server commands signal client behavior via the HX-Trigger header,
  // which htmx re-fires as a "term-action" DOM event.
  document.body.addEventListener("term-action", function (e) {
    var action = e.detail && e.detail.value;
    if (!action) return;
    if (action === "clear") {
      var out = document.getElementById("term-output");
      if (out) out.innerHTML = "";
    } else if (action === "exit") {
      window.dispatchEvent(new CustomEvent("terminal-close"));
    } else if (action.indexOf("goto:") === 0) {
      window.dispatchEvent(new CustomEvent("terminal-close"));
      var el = document.querySelector(action.slice(5));
      if (el) el.scrollIntoView({ behavior: reduced ? "auto" : "smooth" });
    } else if (action.indexOf("open:") === 0) {
      window.dispatchEvent(new CustomEvent("terminal-close"));
      var url = action.slice(5);
      var slug = url.split("/").pop();
      var btn = document.querySelector('[data-project-link="' + slug + '"]');
      if (btn) {
        btn.scrollIntoView({ behavior: reduced ? "auto" : "smooth", block: "center" });
        btn.click();
      } else {
        window.location.assign(url);
      }
    }
  });

  // Clear the input after each command and keep output pinned to bottom.
  document.body.addEventListener("htmx:afterRequest", function (e) {
    var form = document.getElementById("terminal-form");
    if (form && e.target === form) form.querySelector("input[name=cmd]").value = "";
  });
  document.body.addEventListener("htmx:afterSwap", function (e) {
    var target = e.detail && e.detail.target;
    if (target && target.id === "term-output") target.scrollTop = target.scrollHeight;
  });
})();
```

- [ ] **Step 4: Build the CSS**

```bash
./bin/tailwindcss.exe -i tailwind.css -o static/css/site.css --minify
```

Expected: `static/css/site.css` created, non-trivial size (> 10 KB). If the binary reports utility classes missing, confirm `@source "./templates"` resolves (run from repo root).

- [ ] **Step 5: Run the full test suite** (templates unchanged, but confirm nothing broke)

Run: `go test ./...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add tailwind.css static/js/grid.js static/js/site.js static/css/site.css
git commit -m "feat: tailwind theme, phosphor grid canvas, typing/count-up/reveal/terminal js"
```

---

### Task 11: Server wiring, Makefile, manual smoke run

**Files:**
- Create: `cmd/server/main.go`, `Makefile`

- [ ] **Step 1: Write `cmd/server/main.go`**

```go
// Command server runs the portfolio site: one binary, assets embedded.
package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"portfolio"
	"portfolio/internal/handler"
	"portfolio/internal/store"
	"portfolio/internal/terminal"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dbPath := flag.String("db", "inquiries.db", "path to the SQLite inquiry database")
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	st, err := store.Open(*dbPath)
	if err != nil {
		log.Error("open store", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	srv, err := handler.New(portfolio.Templates, portfolio.Static, st, terminal.New(), log)
	if err != nil {
		log.Error("init server", "err", err)
		os.Exit(1)
	}

	log.Info("listening", "addr", *addr)
	if err := http.ListenAndServe(*addr, srv); err != nil {
		log.Error("server", "err", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Write `Makefile`** (recipes must be indented with tabs)

```makefile
TAILWIND := ./bin/tailwindcss.exe

.PHONY: css dev test build

css:
	$(TAILWIND) -i tailwind.css -o static/css/site.css --minify

dev: css
	go run ./cmd/server -db dev.db

test:
	go test ./...

build: css
	go build -ldflags="-s -w" -o bin/portfolio.exe ./cmd/server
```

(If `make` is unavailable on the machine, run the recipe commands directly — they are plain shell.)

- [ ] **Step 3: Build and smoke-run**

```bash
go build -o bin/portfolio.exe ./cmd/server
./bin/portfolio.exe -addr :8080 -db dev.db &
sleep 1
curl -s http://localhost:8080/ | grep -c "Jasen Nicely"   # expect >= 1
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/projects/redline   # expect 200
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/static/css/site.css  # expect 200
kill %1
```

- [ ] **Step 4: Commit**

```bash
git add cmd Makefile
git commit -m "feat: server entrypoint and makefile"
```

---

### Task 12: Dockerfile, README, final verification

**Files:**
- Create: `Dockerfile`, `README.md`

- [ ] **Step 1: Create `Dockerfile`**

```dockerfile
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /portfolio ./cmd/server

FROM scratch
COPY --from=build /portfolio /portfolio
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/portfolio", "-addr", ":8080", "-db", "/data/inquiries.db"]
```

(CSS must be built before `docker build` — `static/css/site.css` is committed, so a clean checkout works.)

- [ ] **Step 2: Create `README.md`**

```markdown
# jasen-nicely.dev — portfolio

Personal site for Jasen Nicely: terminal-inspired single-page portfolio with
htmx-expanded case studies, a live terminal easter egg, and an SQLite-backed
contact form. One Go binary, everything embedded.

## Stack

Go (stdlib net/http) · html/template · htmx 2 · Alpine.js 3 · Tailwind CSS 4
(standalone CLI, no Node) · modernc.org/sqlite (no cgo)

## Develop

    make css     # rebuild static/css/site.css from tailwind.css
    make dev     # run on :8080 with dev.db
    make test    # go test ./...
    make build   # bin/portfolio.exe

Press ` on the site to open the terminal. Try `help`, `projects`,
`sudo hire-me`.

## Deploy

    docker build -t portfolio .
    docker run -p 8080:8080 -v portfolio-data:/data portfolio

Inquiries land in the `inquiries` table of the SQLite db at `/data/inquiries.db`.

## Content

All copy lives in `internal/content/content.go` — projects, resume, skills,
and terminal output render from the same structs.
```

- [ ] **Step 3: Full verification gate**

```bash
gofmt -l .          # expect: no output
go vet ./...        # expect: no findings
go test ./...       # expect: all PASS
make build          # expect: bin/portfolio.exe produced
```

- [ ] **Step 4: Manual browser verification (Playwright MCP)**

Run the server (`./bin/portfolio.exe -db dev.db`), then walk this checklist in the browser at `http://localhost:8080`:

1. Hero: `whoami` types itself out, output fades in, grid glow drifts behind the terminal window.
2. Stats count up when scrolled into view.
3. Click "cat case-study.md →" on Redline — card expands in place, URL becomes `/projects/redline`; "[x] close" collapses it and URL returns to `/`.
4. Direct-visit `http://localhost:8080/projects/redline` — full standalone page renders.
5. Press `` ` `` — terminal slides up. Type `help`, `projects`, `open redline` (closes terminal, expands card), `sudo hire-me` (scrolls to contact), `clear`, `exit`.
6. Submit the contact form empty — inline errors appear, values preserved on partial fill. Submit valid — "message queued" confirmation; verify a row exists: `sqlite3 dev.db "select * from inquiries;"` (or re-run the store test).
7. Visit `/nope` — terminal-styled 404 with the requested path.
8. Resize to 375 px width — nav, cards, form, and terminal overlay all usable; take screenshots at desktop and mobile widths.
9. Emulate `prefers-reduced-motion` — no typing animation, no count-up, content fully visible.

Fix anything that fails, re-run the gate in Step 3, then:

- [ ] **Step 5: Final commit**

```bash
git add Dockerfile README.md
git commit -m "chore: dockerfile, readme, final verification"
```

---

## Self-review notes

- **Spec coverage:** hero/typing/canvas (T5/T10), stats count-up (T5/T10), about (T2/T5), three projects with htmx expand + deep links (T2/T5/T6), resume timeline + skills + education + PDF download (T1/T2/T5), contact form → SQLite with inline errors (T3/T7), terminal overlay with server-side registry + all listed commands incl. `sudo hire-me` (T4/T8/T5), 404 gag + panic recovery (T5/T9), single binary via embed (T5/T11), Tailwind standalone + vendored htmx/Alpine + IBM Plex (T1/T10), `prefers-reduced-motion` (T10), Dockerfile + Makefile (T11/T12). Phone number appears nowhere except inside the PDF.
- **Type consistency:** `contactForm` defined once (T5 creates the type in `contact.go`; T7 replaces the file wholesale, keeping the type). `pageData` fields match every template reference (`.Title`, `.Path`, `.Year`, `.Profile`, `.Stats`, `.Projects`, `.Resume`, `.Skills`, `.School`, `.Project`, `.Form`). Terminal `Result{HTML, Action}` consistent across T4/T8/T10 glue.
- **Known judgment call:** validation errors return HTTP 200 (htmx default swap behavior); documented in code comment.
