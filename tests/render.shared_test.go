package form_test

import "webtyp.com/model"

import (
	"webtyp.com/form"
	"webtyp.com/input"
	"strings"
	"testing"
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

	t.Run("TestRender_SubmitButtonHasID", func(t *testing.T) {
		s := &renderStruct{}
		f, _ := form.New("app", s, &testIDGen{})
		html := f.String()

		expectedID := `id='app.form.submit'`
		if !strings.Contains(html, expectedID) {
			t.Errorf("Expected submit button ID not found in HTML: %s", html)
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
}
