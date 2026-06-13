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

// A Registry is immutable after New, so Execute is safe for concurrent use.
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
// User input is HTML-escaped at the boundary so no command can ever echo
// raw user input into HTML.
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
	// Escape every argument at the boundary so no command can ever echo
	// raw user input into HTML. Legitimate inputs (slugs, command names)
	// contain no HTML metacharacters, so escaping never alters them.
	args := fields[1:]
	for i, a := range args {
		args[i] = html.EscapeString(a)
	}
	return c.run(args)
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
		return Result{HTML: errLine("no such project: " + args[0])}
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
		level := min(max(s.Level, 0), 100)
		bar := strings.Repeat("█", level/10) + strings.Repeat("░", 10-level/10)
		b.WriteString(line(fmt.Sprintf("%-28s %s %d", s.Name, accent(bar), s.Level)))
	}
	return Result{HTML: b.String()}
}

func contactCmd([]string) Result {
	// content.Me is owner-authored, trusted copy — not user input.
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
