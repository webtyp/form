package form_test

import (
	"strings"
	"testing"

	"webtyp.com/dom"
	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

// groupInput is a custom control: a wrapper div around several inputs.
type groupInput struct{ input.Input }

func (g groupInput) RenderInput(*dom.SignalString, func(string)) *dom.Element {
	return dom.NewElement("div")
}

func (g groupInput) Clone(parentID, name string) input.Input {
	return groupInput{g.Input.Clone(parentID, name)}
}

type a11yRecord struct{ Born, Mail, Name string }

func (r *a11yRecord) Schema() []model.Field {
	return []model.Field{
		{Name: "Born", Type: groupInput{input.Text()}},
		{Name: "Mail", Type: input.Email()},
		{Name: "Name", Type: input.Text()},
	}
}
func (r *a11yRecord) Values() []any    { return []any{r.Born, r.Mail, r.Name} }
func (r *a11yRecord) Pointers() []any  { return []any{&r.Born, &r.Mail, &r.Name} }
func (r *a11yRecord) FormName() string { return "a11y" }

// A label may not point at a wrapper div: the custom control names itself as a
// group, and every text control declares its autofill hint.
func TestCustomControlIsALabelledGroupAndInputsHaveAutocomplete(t *testing.T) {
	f, err := form.New("parent", &a11yRecord{}, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	html := f.String()
	if !strings.Contains(html, "role='group'") || !strings.Contains(html, "aria-labelledby=") {
		t.Errorf("custom control must be role=group + aria-labelledby:\n%s", html)
	}
	if !strings.Contains(html, "autocomplete='email'") || !strings.Contains(html, "autocomplete='off'") {
		t.Errorf("inputs must declare autocomplete (email, off):\n%s", html)
	}
}

// Chrome flags a <label> that points at no control: a group's title (radio
// group, custom control) is a plain element named by aria-labelledby.
func TestNoLabelWithoutAControl(t *testing.T) {
	for name, rec := range map[string]model.Fielder{"custom": &a11yRecord{}, "radio": &focusRecord{}} {
		f, err := form.New("parent", rec, &testIDGen{})
		if err != nil {
			t.Fatal(err)
		}
		html := f.String()
		for _, tag := range strings.Split(html, "<label")[1:] {
			head := tag[:strings.Index(tag, ">")]
			if !strings.Contains(head, "for=") {
				t.Errorf("%s: <label%s> has no for=", name, head)
			}
		}
	}
}
