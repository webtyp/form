package form_test

import (
	"strings"
	"testing"

	"webtyp.com/dom"
	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

type noLabelStruct struct {
	model.Fielder
	HiddenField string
}

func (s *noLabelStruct) Schema() []model.Field {
	return []model.Field{
		{Name: "hidden_field", Type: input.Text()}, // no label
	}
}
func (s *noLabelStruct) Pointers() []any { return []any{&s.HiddenField} }
func (s *noLabelStruct) Values() []any   { return []any{s.HiddenField} }

func TestLabelMissingLogsDiagnostic(t *testing.T) {
	var logged []string
	dom.SetLog(func(args ...any) {
		for _, arg := range args {
			if s, ok := arg.(string); ok {
				logged = append(logged, s)
			}
		}
	})
	defer dom.SetLog(nil)

	s := &noLabelStruct{}
	_, err := form.New("app", s, &testIDGen{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}

	joined := strings.Join(logged, " ")
	if !strings.Contains(joined, "webtyp/form: field has a widget but no Label") {
		t.Errorf("expected missing label diagnostic, got %q", joined)
	}
	if !strings.Contains(joined, "hidden_field") {
		t.Errorf("expected diagnostic to include field name, got %q", joined)
	}
}
