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
