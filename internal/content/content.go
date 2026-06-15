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
