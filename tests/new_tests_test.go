package form_test

import "webtyp.com/model"

import (
	"testing"

	"webtyp.com/form"
	"webtyp.com/input"
)

type testUser struct {
	id    string
	name  string
	email string
}

func (u *testUser) Schema() []model.Field {
	return []model.Field{
		{Name: "id", Type: input.Text(), DB: &model.FieldDB{PK: true}},
		{Name: "name", Type: input.Text(), NotNull: true},
		{Name: "email", Type: input.Email(), NotNull: true},
	}
}
func (u *testUser) Values() []any    { return []any{u.id, u.name, u.email} }
func (u *testUser) Pointers() []any  { return []any{&u.id, &u.name, &u.email} }
func (u *testUser) FormName() string { return "user" }

func TestNewWithFielder(t *testing.T) {
	u := &testUser{id: "1", name: "John", email: "john@example.com"}
	f, err := form.New("parent", u, &testIDGen{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if f.GetID() != "parent.user" {
		t.Errorf("Expected form ID 'parent.user', got '%s'", f.GetID())
	}

	// 2, not 3: "id" is a PK (DB.PK: true) and New hides every PK by
	// default now, auto-increment or not — see New's ShowField comment.
	if len(f.Inputs) != 2 {
		t.Errorf("Expected 2 inputs (PK hidden by default), got %d", len(f.Inputs))
	}
	if f.Input("id") != nil {
		t.Errorf("Expected 'id' input to be nil (hidden PK), but got non-nil")
	}
}

func TestNewShowFieldRevealsPK(t *testing.T) {
	u := &testUser{id: "1", name: "John", email: "john@example.com"}
	f, err := form.New("parent", u, &testIDGen{}, form.ShowField("id"))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if len(f.Inputs) != 3 {
		t.Errorf("Expected 3 inputs (id shown via ShowField), got %d", len(f.Inputs))
	}
	if f.Input("id") == nil {
		t.Errorf("Expected 'id' input to be non-nil after ShowField(\"id\")")
	}
}

// TestSyncValuesAssignsHiddenPK: a brand-new record's hidden text PK is
// still its zero value when SyncValues runs (nothing rendered it), and the
// backend (model.ValidateFields via orm.Create) rejects an empty PK
// outright — SyncValues must fill it before the record ever gets there.
func TestSyncValuesAssignsHiddenPK(t *testing.T) {
	u := &testUser{name: "Old"} // id left zero-value, like a fresh composing record
	f, err := form.New("p", u, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}

	if err := f.SyncValues(u); err != nil {
		t.Fatal(err)
	}

	if u.id == "" {
		t.Errorf("Expected a generated id, got empty string")
	}
}

// TestSyncValuesPreservesExistingHiddenPK: an EXISTING record's id must
// survive SyncValues untouched — only an EMPTY hidden PK gets a new one.
func TestSyncValuesPreservesExistingHiddenPK(t *testing.T) {
	u := &testUser{id: "existing-id", name: "Old"}
	f, err := form.New("p", u, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}

	if err := f.SyncValues(u); err != nil {
		t.Fatal(err)
	}

	if u.id != "existing-id" {
		t.Errorf("Expected id to stay 'existing-id', got %q", u.id)
	}
}

type autoUser struct {
	id   int64
	name string
}

func (u *autoUser) Schema() []model.Field {
	return []model.Field{
		{Name: "id", Type: model.Int(), DB: &model.FieldDB{PK: true, AutoInc: true}},
		{Name: "name", Type: input.Text()},
	}
}
func (u *autoUser) Values() []any    { return []any{u.id, u.name} }
func (u *autoUser) Pointers() []any  { return []any{&u.id, &u.name} }
func (u *autoUser) FormName() string { return "auto" }

func TestNewAutoIncPKExcludedReal(t *testing.T) {
	u := &autoUser{id: 1, name: "test"}
	f, err := form.New("parent", u, &testIDGen{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if len(f.Inputs) != 1 {
		t.Errorf("Expected 1 input (PK+AutoInc should be skipped), got %d", len(f.Inputs))
	}
	if f.Input("id") != nil {
		t.Errorf("Expected 'id' input to be nil (skipped), but got non-nil")
	}
}

func TestSyncValuesText(t *testing.T) {
	u := &testUser{id: "1", name: "Old"}
	f, err := form.New("p", u, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}

	f.SetValues("name", "New")
	err = f.SyncValues(u)
	if err != nil {
		t.Fatal(err)
	}

	if u.name != "New" {
		t.Errorf("Expected 'New', got '%s'", u.name)
	}
}

func TestValidateDataValid(t *testing.T) {
	u := &testUser{id: "100", name: "John Doe", email: "john@example.com"}
	f, err := form.New("p", u, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}

	err = f.ValidateData('u', u)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}
