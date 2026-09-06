package form_test

import (
	"testing"

	"webtyp.com/fmt"
	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

// kindFixture is a one-field Fielder used to render a single input kind
// through the real form pipeline.
type kindFixture struct {
	inp     input.Input
	valText string
	valInt  int64
	valBool bool
}

func (k *kindFixture) Schema() []model.Field {
	return []model.Field{{Name: "tfield", Type: k.inp}}
}

func (k *kindFixture) Pointers() []any {
	switch k.inp.Storage() {
	case model.FieldInt:
		return []any{&k.valInt}
	case model.FieldBool:
		return []any{&k.valBool}
	default:
		return []any{&k.valText}
	}
}

func (k *kindFixture) Values() []any {
	switch k.inp.Storage() {
	case model.FieldInt:
		return []any{k.valInt}
	case model.FieldBool:
		return []any{k.valBool}
	default:
		return []any{k.valText}
	}
}

// Test_Render_BlackBox verifies String output for inputs using the public API.
func Test_Render_BlackBox(t *testing.T) {
	opts12 := []fmt.KeyValue{{Key: "1", Value: "Admin"}, {Key: "2", Value: "Editor"}}
	optsGender := []fmt.KeyValue{{Key: "m", Value: "Male"}, {Key: "f", Value: "Female"}}

	cases := []rc{
		{
			t: "Checkbox", name: "renders checkbox input",
			contain: `type='checkbox'`,
		},
		{
			t: "Datalist", name: "contains datalist element",
			opts:    opts12,
			contain: `<datalist`,
		},
		{
			t: "Datalist", name: "contains option values",
			opts:    opts12,
			contain: `value='1'`,
		},
		{
			t: "Datalist", name: "links input to datalist via list attribute",
			opts:    opts12,
			contain: `list='`,
		},
		{
			t: "Radio", name: "renders label per option",
			opts:    optsGender,
			contain: `<label>`,
		},
		{
			t: "Radio", name: "renders male option value",
			opts:    optsGender,
			contain: `value='m'`,
		},
		{
			t: "Radio", name: "renders checked when value matches",
			opts:    optsGender,
			values:  []string{"f"},
			contain: `checked`,
		},
		{
			t: "Select", name: "renders select element",
			opts:    opts12,
			contain: `<select`,
		},
		{
			t: "Select", name: "renders option element",
			opts:    opts12,
			contain: `<option`,
		},
		{
			t: "Select", name: "renders selected when value matches",
			opts:    opts12,
			values:  []string{"2"},
			contain: `selected`,
		},
		{
			t: "Textarea", name: "uses textarea tag",
			contain: `<textarea`,
		},
		{
			t: "Textarea", name: "renders value inside tag",
			values:  []string{"hello world"},
			contain: `hello world`,
		},
		{
			t: "Search", name: "renders search input",
			contain: `type='search'`,
		},
		{
			t: "Text", name: "renders input tag",
			contain: `<input`,
		},
		{
			t: "Text", name: "has type text",
			contain: `type='text'`,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.t+"/"+c.name, func(t *testing.T) {
			var inp input.Input
			switch c.t {
			case "Checkbox":
				inp = input.Checkbox()
			case "Datalist":
				inp = input.Datalist()
			case "Radio":
				inp = input.Radio()
			case "Select":
				inp = input.Select()
			case "Textarea":
				inp = input.Textarea()
			case "Search":
				inp = input.Search()
			case "Text":
				inp = input.Text()
			default:
				t.Fatalf("unsupported test input type: %s", c.t)
			}

			fx := &kindFixture{inp: inp}
			f, err := form.New("app", fx, &testIDGen{})
			if err != nil {
				t.Fatalf("failed to create form: %v", err)
			}

			if len(c.opts) > 0 {
				f.SetOptions("tfield", c.opts...)
			}
			if len(c.values) > 0 {
				f.SetValues("tfield", c.values...)
			}

			html := f.String()
			if !fmt.Contains(html, c.contain) {
				t.Errorf("f.String() missing %q\ngot: %s", c.contain, html)
			}
		})
	}
}

type rc struct {
	t       string
	name    string
	values  []string
	opts    []fmt.KeyValue
	contain string
}
