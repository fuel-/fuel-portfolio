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
