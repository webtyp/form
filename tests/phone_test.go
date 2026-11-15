package form_test

import (
	"testing"

	"webtyp.com/country"
	"webtyp.com/fmt"
	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

type phoneFixture struct{ phone string }

func (p *phoneFixture) Schema() []model.Field {
	return []model.Field{{Name: "phone", Type: input.Phone(), Label: "Phone"}}
}
func (p *phoneFixture) Pointers() []any { return []any{&p.phone} }
func (p *phoneFixture) Values() []any   { return []any{p.phone} }

// With the app's countries a phone is a prefix select (default first) and the national number.
func TestPhone_DrawsPrefixSelectAndNumber(t *testing.T) {
	fx := &phoneFixture{phone: "+584121234567"}
	f, err := form.New("app", fx, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	f.SetOptions("phone", country.Catalog{Default: "CL", First: []string{"VE"}}.PrefixOptions()...)
	html := f.String()
	for _, want := range []string{"vy-field__prefix", "vy-field__number", "+56", "+58"} {
		if !fmt.Contains(html, want) {
			t.Errorf("phone field must contain %q:\n%s", want, html)
		}
	}
}

// With no countries it is one box where the number is typed with its "+".
func TestPhone_WithoutCountriesIsOneBox(t *testing.T) {
	f, err := form.New("app", &phoneFixture{}, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Contains(f.String(), "vy-field__prefix") {
		t.Error("with no countries the phone must not draw a prefix select")
	}
}
