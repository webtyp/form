package form_test

import (
	"strings"
	"testing"

	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

type anatomyStruct struct {
	model.Fielder
	Nombre string
	Gender string
}

func (s *anatomyStruct) Schema() []model.Field {
	return []model.Field{
		{Name: "nombre", NotNull: true, Type: input.Text()},
		{Name: "gender", Type: input.Radio()},
	}
}

func (s *anatomyStruct) Pointers() []any { return []any{&s.Nombre, &s.Gender} }
func (s *anatomyStruct) Values() []any   { return []any{s.Nombre, s.Gender} }

func TestForm_Anatomy(t *testing.T) {
	s := &anatomyStruct{}
	f, _ := form.New("app", s, &testIDGen{})
	html := f.String()

	// 1. Root container class should be the field root class
	if !strings.Contains(html, "class='"+clsField+"'") {
		t.Errorf("Expected html to contain root class the field root class, got: %s", html)
	}

	// 2. Label class should be the field label class
	if !strings.Contains(html, "class='"+clsFieldLabel+"'") {
		t.Errorf("Expected html to contain label class the field label class, got: %s", html)
	}

	// 3. Error span class should be the field error class
	if !strings.Contains(html, "class='"+clsFieldError+"'") {
		t.Errorf("Expected html to contain error class the field error class, got: %s", html)
	}

	// 4. Radio group should contain the field radio-group class
	if !strings.Contains(html, "class='"+clsFieldRadios+"'") {
		t.Errorf("Expected html to contain radio-group class the field radio-group class, got: %s", html)
	}
}
