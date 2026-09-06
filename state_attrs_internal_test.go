package form

import (
	"strings"
	"testing"

	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/model"
)

type stateAttrsStruct struct {
	model.Fielder
	Nombre string
}

func (s *stateAttrsStruct) Schema() []model.Field {
	return []model.Field{
		{Name: "nombre", NotNull: true, Type: input.Text()},
	}
}

func (s *stateAttrsStruct) Pointers() []any { return []any{&s.Nombre} }
func (s *stateAttrsStruct) Values() []any   { return []any{s.Nombre} }

// testIDGen is the internal-package double for model.IDGenerator — form never
// constructs its own generator, so even internal tests inject one.
type testIDGen struct{ n int }

func (g *testIDGen) NewID() string {
	g.n++
	return "test-id-" + fmt.Convert(g.n).String()
}

func TestForm_StateAttrs(t *testing.T) {
	s := &stateAttrsStruct{}
	f, _ := New("app", s, &testIDGen{})

	// Case 1: Valid and unlocked field -> should NOT have data-invalid='true' or data-locked='true'
	html := f.String()
	if strings.Contains(html, "data-invalid='true'") {
		t.Errorf("Expected valid form NOT to contain data-invalid='true', got: %s", html)
	}
	if strings.Contains(html, "data-locked='true'") {
		t.Errorf("Expected unlocked form NOT to contain data-locked='true', got: %s", html)
	}

	// Case 2: Field with error -> should contain data-invalid='true'
	f.errorSignals[0].Set("nombre is required")
	htmlErr := f.String()
	if !strings.Contains(htmlErr, "data-invalid='true'") {
		t.Errorf("Expected invalid form to contain data-invalid='true', got: %s", htmlErr)
	}

	// Case 3: Form locked -> should contain data-locked='true'
	f.SetLocked(true)
	htmlLocked := f.String()
	if !strings.Contains(htmlLocked, "data-locked='true'") {
		t.Errorf("Expected locked form to contain data-locked='true', got: %s", htmlLocked)
	}

	// Case 4: No tw-field-error--visible should be present in any of these states
	if strings.Contains(html, "tw-field-error--visible") ||
		strings.Contains(htmlErr, "tw-field-error--visible") ||
		strings.Contains(htmlLocked, "tw-field-error--visible") {
		t.Error("HTML contains obsolete class 'tw-field-error--visible'")
	}
}
