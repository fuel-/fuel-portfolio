# Portfolio Site — Design Spec

**Date:** 2026-06-12
**Status:** Approved
**Owner:** Jasen Nicely

## Purpose

A personal website for Jasen Nicely with two audiences:

1. **Potential employers** — show what he's capable of (senior full-stack, Go/TS/React, 15 years).
2. **Freelance/contract clients** — act as a contact portal for inquiries.

Tone: terminal-inspired and a little nerdy, but polished and approachable — must never put off a non-technical recruiter or SMB client. The site should be eye-catching ("pop") with motion graphics. This build is also a capability showcase; it is unrelated to the older gideon.gg portfolio codebase (fresh start, no constraints from it).

## Approach (decided)

**Single-page home + htmx case-study expansion.** One high-momentum scroll page; each project card expands into a full case study fetched via htmx with `hx-push-url`, and the same case-study routes render as standalone pages on direct visit (deep-linkable, SEO-real). A global terminal overlay is the nerd layer; recruiters never have to touch it.

Rejected alternatives:
- *Pure single-page:* case studies squeezed into cards; the 1-week→12-seconds story deserves depth.
- *Multi-page site:* more surface to polish, loses scroll momentum, risks feeling templated.

## Visual language

- Dark theme matching the resume PDF: near-black zinc background, generous whitespace, big confident type.
- **Fonts:** IBM Plex Sans (body) + IBM Plex Mono (accents) — matches the resume PDF so site and resume read as one brand.
- **Accent:** phosphor green, used sparingly (sharp, not Matrix cosplay).
- All motion respects `prefers-reduced-motion`.

## Page flow (single scroll)

1. **Hero** — stylized terminal window; `$ whoami` types itself out → name, title, one-line pitch; blinking cursor; subtle animated phosphor-grid canvas behind. CTAs: "View work" (scroll to projects), "Hire me" (scroll to contact). Corner hint: `` press ` to open terminal ``.
2. **By the numbers** — 15 yrs / 99.9% / 30+ PRs / 9 yrs as a strip with count-up animation on scroll-into-view.
3. **About** — short, in the voice of the cover note: curiosity-driven, self-taught, thrives on hard problems and small teams.
4. **Projects** — three case-study cards, each expanding via htmx into a full case study (problem → approach → outcome with numbers), URL updated to `/projects/<slug>`:
   - **Property Tax Pipeline** — Go + SQLite contract work; 1 week → 12 seconds (99.9%).
   - **Announcement System** — SAP Concur; 30+ coordinated PRs across two codebases, full-stack ownership (described generically, no employer internals).
   - **Redline** — live SaaS at https://redline.gideon.gg; client-facing design review tool; Go + htmx + Alpine + Postgres — same stack as this site (say so).
5. **Resume** — vertical timeline: SAP Concur Implementations (2011–2013) → RAD (2013–2017) → STAT (2017–present, Senior Technical Consultant 2023) + Property Tax Processor contract work; tech-stack proficiency bars echoing the PDF; education (SAGU, 1998); "Download PDF" button.
6. **Contact** — htmx form: name, email, company (optional), project type (hiring / contract / other), message. Server-validated, saved to SQLite. Terminal-flavored success state (`✓ message queued — I'll get back to you within 24h`). Footer: GitHub, email, location (McKinney, TX).

**Privacy rule:** phone number stays off the site. Email and GitHub only. (Phone is on the downloadable PDF, which is an explicit user action.)

## Terminal easter egg

- Opened with `` ` `` keypress or clicking the hint; slide-up overlay.
- Commands executed **server-side** via `POST /terminal` against a command registry in Go — single source of truth with page content.
- Commands: `help`, `whoami`, `projects`, `open <project>`, `resume`, `skills`, `contact`, `clear`, `exit`, plus toys (`sudo hire-me` → wink + jump to contact form).
- Unknown command → `command not found — try 'help'`. The 404 page reuses the same gag.

## Architecture

**Stack:** Go (stdlib `net/http`, Go 1.22+ method routing — no framework), `html/template`, htmx, Alpine.js, Tailwind CSS via standalone CLI (no Node). htmx/Alpine vendored locally, no CDNs. Everything embedded via `embed.FS` → single static binary.

```
portfolio-fable/
├── cmd/server/main.go        # wiring, flags (port, db path)
├── internal/
│   ├── content/              # typed structs: Profile, Project, ResumeEntry, Skill
│   ├── handler/              # http handlers: pages, fragments, contact, terminal
│   ├── terminal/             # command registry: name → handler → response
│   └── store/                # SQLite inquiry persistence (modernc.org/sqlite, no cgo)
├── templates/                # layout, page, partials (project card/detail, form states)
├── static/                   # built CSS, vendored htmx/alpine, canvas JS, resume PDF
└── tailwind.css              # source; built to static/css via Makefile
```

**Content as data:** all copy — projects, resume timeline, skills, terminal output — lives in `internal/content` as typed Go values. Templates and the terminal render from the same structs so they cannot drift.

## Routes

| Route | Behavior |
|---|---|
| `GET /` | Full page render |
| `GET /projects/{slug}` | Fragment (case-study panel) when `HX-Request` header present; full standalone page otherwise |
| `POST /contact` | Validate → insert SQLite → success fragment; errors re-render form fragment inline, values preserved |
| `POST /terminal` | Command string in → HTML response from command registry |
| `GET /static/*` | Embedded assets with cache headers |

## Error handling

- Server-side form validation only (htmx swaps error states in; no client validation logic to maintain).
- SQLite write failure → "something broke, email me directly at …" fragment. A bug must never silently eat a lead.
- Unknown route → terminal-styled 404 (`command not found`).
- Panic-recovery middleware → styled 500.

## Testing

- Table-driven unit tests: contact validation, terminal command registry.
- `httptest` integration tests: every route, HX-Request fragment-vs-page split, contact happy/sad paths.
- Startup smoke test: all embedded templates parse.
- Manual browser verification of animations/motion via Playwright at the end.

## Dev & deploy

- `make dev` — Tailwind watch + server; `make build` — single binary.
- Two-stage scratch-based Dockerfile included.
- **Deployment itself is out of scope**; the deliverable runs locally on `:8080`.

## Decisions log

| Decision | Choice |
|---|---|
| Relationship to old gideon.gg build | Fresh start; capability showcase |
| Contact handling | htmx form → SQLite locally; email delivery deferred |
| Terminal literalness | Hybrid: polished layout + live terminal overlay |
| Projects shown | Property Tax Pipeline, Announcement System, Redline |
| UI framework | Tailwind (standalone CLI) |
| Fonts | IBM Plex Sans / Mono |
| Phone number | Off the site; PDF only |
