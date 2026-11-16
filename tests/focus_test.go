package form_test

import (
	"testing"

	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

type focusRecord struct {
	Gender string
	Name   string
}

func (r *focusRecord) Schema() []model.Field {
	return []model.Field{
		{Name: "Gender", Type: input.Gender()},
		{Name: "Name", Type: input.Text()},
	}
}
func (r *focusRecord) Values() []any    { return []any{r.Gender, r.Name} }
func (r *focusRecord) Pointers() []any  { return []any{&r.Gender, &r.Name} }
func (r *focusRecord) FormName() string { return "focus" }

// A radio group's own div is not focusable: Focus must target an option of it.
func TestFocusTargetsAnOptionOfAFirstRadioGroup(t *testing.T) {
	f, err := form.New("parent", &focusRecord{}, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	html := f.String()
	id := f.Focus().FocusedFieldID()
	if id == "" {
		t.Fatal("Focus left no target")
	}
	groupID := f.Inputs[0].GetID()
	if id == groupID {
		t.Fatalf("Focus targeted the radio group div %q, not an option", id)
	}
	if !contains(html, "id='"+id+"'") || !contains(html, "type='radio'") {
		t.Fatalf("target %q is not a rendered control:\n%s", id, html)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
