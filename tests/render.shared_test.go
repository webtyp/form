package form_test

import "webtyp.com/model"

import (
	"strings"
	"testing"
	"webtyp.com/form"
	"webtyp.com/input"
)

type renderStruct struct {
	model.Fielder
	Nombre string
}

func (s *renderStruct) Schema() []model.Field {
	return []model.Field{
		{Name: "nombre", NotNull: true, Type: input.Text()},
	}
}

func (s *renderStruct) Pointers() []any { return []any{&s.Nombre} }
func (s *renderStruct) Values() []any   { return []any{s.Nombre} }

func runRenderTests(t *testing.T) {
	t.Run("TestRenderInput_EmitsErrorSpan", func(t *testing.T) {
		s := &renderStruct{}
		f, _ := form.New("app", s, &testIDGen{})
		html := f.String()

		// Note: html.Span().String() uses single quotes for attributes
		expectedSpan := `id='app.form.nombre.error' class='` + clsFieldError + `' aria-live='polite'`
		if !strings.Contains(html, expectedSpan) {
			t.Errorf("Expected error span not found in HTML: %s", html)
		}
	})

	t.Run("TestRender_SubmitButtonRendered", func(t *testing.T) {
		s := &renderStruct{}
		f, _ := form.New("app", s, &testIDGen{})
		html := f.String()

		if !strings.Contains(html, clsFieldSubmit) || !strings.Contains(html, "type='submit'") {
			t.Errorf("Expected submit button not found in HTML: %s", html)
		}
	})

	t.Run("TestRender_ErrorIDMethod", func(t *testing.T) {
		s := &renderStruct{}
		f, _ := form.New("app", s, &testIDGen{})
		inp := f.Input("nombre")

		expectedErrorID := "app.form.nombre.error"
		if getter, ok := inp.(interface{ ErrorID() string }); ok {
			if getter.ErrorID() != expectedErrorID {
				t.Errorf("Expected ErrorID %s, got %s", expectedErrorID, getter.ErrorID())
			}
		} else {
			t.Errorf("Input does not implement ErrorID()")
		}
	})

	t.Run("TestLabelFollowsControlID", func(t *testing.T) {
		s := &renderStruct{}
		f, _ := form.New("app", s, &testIDGen{})
		html := f.String()

		idxInputID := strings.Index(html, "<input id='")
		if idxInputID == -1 {
			t.Fatalf("<input id=' not found in HTML: %s", html)
		}
		startID := idxInputID + len("<input id='")
		endID := strings.Index(html[startID:], "'")
		controlID := html[startID : startID+endID]

		idxLabelFor := strings.Index(html, "for='")
		if idxLabelFor == -1 {
			t.Fatalf("for=' not found in HTML: %s", html)
		}
		startFor := idxLabelFor + len("for='")
		endFor := strings.Index(html[startFor:], "'")
		labelFor := html[startFor : startFor+endFor]

		if labelFor != controlID {
			t.Errorf("Expected label for='%s' to match control id='%s'", labelFor, controlID)
		}
	})
}
