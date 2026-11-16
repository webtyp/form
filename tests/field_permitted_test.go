package form_test

import (
	"testing"

	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

type permittedRecord struct{ Number string }

func (r *permittedRecord) Schema() []model.Field {
	return []model.Field{{
		Name: "Number", Type: input.Text(),
		Permitted: model.Permitted{Letters: true, Numbers: true, Extra: []rune{'-', '.'}, Maximum: 30},
	}}
}
func (r *permittedRecord) Values() []any    { return []any{r.Number} }
func (r *permittedRecord) Pointers() []any  { return []any{&r.Number} }
func (r *permittedRecord) FormName() string { return "permitted" }

// The charset a field declares is the one the form checks, the same one the
// server applies (model.Field.Validate): a value the server accepts must not be
// rejected on screen.
func TestFormValidatesWithTheFieldsOwnPermitted(t *testing.T) {
	f, err := form.New("parent", &permittedRecord{Number: "12345678-9"}, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Validate(); err != nil {
		t.Fatalf("form rejects what the field declares allowed: %v", err)
	}
}
