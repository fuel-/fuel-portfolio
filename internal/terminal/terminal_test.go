package terminal

import (
	"strings"
	"testing"

	"portfolio/internal/content"
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
	res = r.Execute("open <script>alert(1)</script>")
	if strings.Contains(res.HTML, "<script>") {
		t.Error("raw <script> leaked via command argument")
	}
	if !strings.Contains(res.HTML, "&lt;script&gt;") {
		t.Errorf("HTML = %q, want escaped argument", res.HTML)
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

func TestContactAction(t *testing.T) {
	if res := New().Execute("contact"); res.Action != "goto:#contact" {
		t.Errorf("Action = %q, want goto:#contact", res.Action)
	}
}

func TestSkillsBarAlwaysTenCells(t *testing.T) {
	res := New().Execute("skills")
	want := 10 * len(content.Skills)
	if got := strings.Count(res.HTML, "█") + strings.Count(res.HTML, "░"); got != want {
		t.Errorf("bar cells = %d, want %d", got, want)
	}
}
