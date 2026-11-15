package form_test

import (
	"testing"

	"webtyp.com/fmt"
	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

// linkFixture is a record with a document type, an issuing country and an
// estimated flag, to exercise what one field declares about another.
type linkFixture struct {
	typ, country, rut string
	estimated         bool
	schema            []model.Field
}

func (l *linkFixture) Schema() []model.Field { return l.schema }
func (l *linkFixture) Pointers() []any       { return []any{&l.typ, &l.rut, &l.country, &l.estimated} }
func (l *linkFixture) Values() []any         { return []any{l.typ, l.rut, l.country, l.estimated} }

func showWhenSchema(extra func(*model.Field)) []model.Field {
	country := model.Field{Name: "country", Type: input.Text(), Label: "Country",
		ShownWhen: model.When{Field: "typ", In: []string{"passport"}}}
	if extra != nil {
		extra(&country)
	}
	return []model.Field{
		{Name: "typ", Type: input.Text(), Label: "Type"},
		{Name: "rut", Type: input.Text(), Label: "Number"},
		country,
		{Name: "estimated", Type: input.Checkbox(), Label: "Estimated"},
	}
}

// A field is shown only while the field it names holds a listed value.
func TestShownWhen_HidesAndShows(t *testing.T) {
	fx := &linkFixture{schema: showWhenSchema(nil), typ: "rut"}
	f, err := form.New("app", fx, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	if !fmt.Contains(f.String(), "data-hidden='true'") {
		t.Errorf("country must start hidden while typ is %q:\n%s", "rut", f.String())
	}
	f.SetValues("typ", "passport")
	if fmt.Contains(f.String(), "data-hidden='true'") {
		t.Errorf("country must show once typ is passport:\n%s", f.String())
	}
}

// A hidden field is not validated and travels empty.
func TestShownWhen_HiddenFieldIsEmptyAndNotValidated(t *testing.T) {
	fx := &linkFixture{schema: showWhenSchema(func(c *model.Field) {
		c.Type = input.Text()
		c.NotNull = true
	}), typ: "rut", country: "VE"}
	f, err := form.New("app", fx, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Validate(); err != nil {
		t.Errorf("a required field that is hidden must not fail validation: %v", err)
	}
	out := &linkFixture{schema: fx.schema}
	if err := f.SyncValues(out); err != nil {
		t.Fatal(err)
	}
	if out.country != "" {
		t.Errorf("a hidden field must travel empty, got %q", out.country)
	}
}

// New fails when ShownWhen names a field the form does not have.
func TestShownWhen_UnknownFieldFailsAtNew(t *testing.T) {
	fx := &linkFixture{schema: showWhenSchema(func(c *model.Field) {
		c.ShownWhen = model.When{Field: "nope", In: []string{"x"}}
	})}
	if _, err := form.New("app", fx, &testIDGen{}); err == nil {
		t.Error("a ShownWhen naming a missing field must fail at New")
	}
}

// alsoInput sets another field when it is told to.
type alsoInput struct {
	input.Base
	set func(string)
}

func (a *alsoInput) Clone(parentID, name string) input.Input {
	c := *a
	c.InitBase(parentID, name, "text")
	return &c
}
func (a *alsoInput) BindAlsoSets(set func(string)) { a.set = set }

func alsoSchema(inp input.Input, target string) []model.Field {
	return []model.Field{
		{Name: "typ", Type: inp, Label: "Type", AlsoSetsField: target},
		{Name: "rut", Type: input.Text(), Label: "Number"},
		{Name: "country", Type: input.Text(), Label: "Country"},
		{Name: "estimated", Type: input.Checkbox(), Label: "Estimated"},
	}
}

// The form hands the input a function that writes the field the model names.
func TestAlsoSetsField_InputWritesTheOtherField(t *testing.T) {
	probe := &alsoInput{}
	fx := &linkFixture{schema: alsoSchema(probe, "estimated")}
	f, err := form.New("app", fx, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	bound := f.Input("typ").(*alsoInput)
	if bound.set == nil {
		t.Fatal("form did not bind the setter")
	}
	bound.set("true")
	out := &linkFixture{schema: fx.schema}
	if err := f.SyncValues(out); err != nil {
		t.Fatal(err)
	}
	if !out.estimated {
		t.Error("the input's set(\"true\") did not reach the estimated field")
	}
}

// New fails when the input cannot set another field, or the target is missing.
func TestAlsoSetsField_FailsAtNew(t *testing.T) {
	plain := &linkFixture{schema: alsoSchema(input.Text(), "estimated")}
	if _, err := form.New("app", plain, &testIDGen{}); err == nil {
		t.Error("an input that does not implement input.AlsoSetter must fail at New")
	}
	missing := &linkFixture{schema: alsoSchema(&alsoInput{}, "nope")}
	if _, err := form.New("app", missing, &testIDGen{}); err == nil {
		t.Error("an AlsoSetsField naming a missing field must fail at New")
	}
}
