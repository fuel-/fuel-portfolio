# portfolio

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

Press `` ` `` on the site to open the terminal. Try `help`, `projects`,
`sudo hire-me`.

## Deploy

    docker build -t portfolio .
    docker run -p 8080:8080 -v portfolio-data:/data portfolio

Inquiries land in the `inquiries` table of the SQLite db at `/data/inquiries.db`.

To also get emailed on each inquiry, `cp .env.example .env` and fill in the
SMTP creds (Resend). If `.env` is absent or incomplete, the app just saves to
the db and skips email — the db row is always the source of truth. The email's
Reply-To is set to the visitor, so replying from your inbox reaches them.

## Content

All copy lives in `internal/content/content.go` — projects, resume, skills,
and terminal output render from the same structs.
