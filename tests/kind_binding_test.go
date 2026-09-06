package form_test

import (
	"testing"

	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

// mixedKindRecord mixes input.* (form-bound) kinds with model.* base kinds
// (validation-only, no UI). Only the input.* fields must produce a bound
// form input, in schema order (Kind unification phase B, Stage 4).
type mixedKindRecord struct {
	Name string
	Note string
	Age  int
}

func (m *mixedKindRecord) Schema() []model.Field {
	return []model.Field{
		{Name: "Name", Type: input.Text(), NotNull: true},
		{Name: "Note", Type: model.Text()}, // base kind: validation only, no input
		{Name: "Age", Type: model.Int()},   // base kind: validation only, no input
	}
}

func (m *mixedKindRecord) Values() []any {
	return []any{m.Name, m.Note, m.Age}
}

func (m *mixedKindRecord) Pointers() []any {
	return []any{&m.Name, &m.Note, &m.Age}
}

func (m *mixedKindRecord) FormName() string {
	return "mixedKindRecord"
}

func TestForm_New_BindsOnlyInputKinds(t *testing.T) {
	rec := &mixedKindRecord{Name: "Jane", Note: "internal note", Age: 30}
	f, err := form.New("test-parent", rec, &testIDGen{})
	if err != nil {
		t.Fatalf("form.New() error = %v", err)
	}

	if len(f.Inputs) != 1 {
		t.Fatalf("len(f.Inputs) = %d, want 1 (only the input.Text field should bind)", len(f.Inputs))
	}
	if f.Inputs[0].Name() != "text" {
		t.Errorf("f.Inputs[0].Name() = %q, want %q (the Name field's kind)", f.Inputs[0].Name(), "text")
	}
}

// TestBaseKind_StillValidates proves the old "Widget: nil => no validation"
// hole is closed: a field using a base model.Kind (no input.Input UI) must
// still run through Field.Validate and reject disallowed input.
func TestBaseKind_StillValidates(t *testing.T) {
	field := model.Field{Name: "Note", Type: model.Text(), NotNull: true}

	if err := field.Validate(""); err == nil {
		t.Error("Validate(\"\") on a NotNull base-kind field = nil error, want required error")
	}

	if err := field.Validate("<script>alert(1)</script>"); err == nil {
		t.Error("Validate() on base Text kind accepted HTML-dangerous chars, want error")
	}

	if err := field.Validate("a normal note"); err != nil {
		t.Errorf("Validate() on valid text = %v, want nil", err)
	}
}
