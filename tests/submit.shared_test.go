package form_test

import (
	"testing"
	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

type submitStruct struct {
	model.Fielder
	Nombre string
}

func (s *submitStruct) Schema() []model.Field {
	return []model.Field{
		{Name: "nombre", NotNull: true, Type: input.Text()},
	}
}

func (s *submitStruct) Pointers() []any { return []any{&s.Nombre} }
func (s *submitStruct) Values() []any   { return []any{s.Nombre} }

func runSubmitTests(t *testing.T) {
	t.Run("TestSubmit_CallbackReceivesDataAndDone", func(t *testing.T) {
		s := &submitStruct{Nombre: "Jules"}
		f, _ := form.New("app", s, &testIDGen{})

		called := false
		var receivedData model.Fielder
		f.OnSubmit(func(data model.Fielder, done func(error)) {
			called = true
			receivedData = data
			done(nil)
		})

		f.SetValues("nombre", "New Name")
		err := f.Submit()
		if err != nil {
			t.Fatalf("Submit returned unexpected error: %v", err)
		}

		if !called {
			t.Error("OnSubmit callback was not called")
		}

		if receivedData != s {
			t.Error("Callback received wrong data pointer")
		}

		if s.Nombre != "New Name" {
			t.Errorf("Expected struct field synced to 'New Name', got %q", s.Nombre)
		}
	})

	t.Run("TestSubmit_ValidationFailureReturnsError", func(t *testing.T) {
		s := &submitStruct{Nombre: ""}
		f, _ := form.New("app", s, &testIDGen{})

		called := false
		f.OnSubmit(func(data model.Fielder, done func(error)) {
			called = true
			done(nil)
		})

		// "nombre" is NotNull: true, so empty should fail
		f.SetValues("nombre", "")
		err := f.Submit()
		if err == nil {
			t.Fatal("Expected validation error from Submit(), got nil")
		}

		if called {
			t.Error("OnSubmit callback should NOT have been called on validation failure")
		}
	})

	t.Run("TestSubmit_NoResetOnSuccess", func(t *testing.T) {
		s := &submitStruct{Nombre: "Jules"}
		f, _ := form.New("app", s, &testIDGen{})
		f.NoResetOnSuccess()
		f.SetValues("nombre", "Valor Original")

		f.OnSubmit(func(data model.Fielder, done func(error)) {
			done(nil)
		})

		err := f.Submit()
		if err != nil {
			t.Fatalf("Submit failed: %v", err)
		}

		val := f.Input("nombre").(interface{ GetValue() string }).GetValue()
		if val != "Valor Original" {
			t.Errorf("Expected form to retain value %q, got %q", "Valor Original", val)
		}
	})

	t.Run("TestSubmit_DefaultResetOnSuccess", func(t *testing.T) {
		s := &submitStruct{Nombre: "Jules"}
		f, _ := form.New("app", s, &testIDGen{})
		f.SetValues("nombre", "To Be Reset")

		f.OnSubmit(func(data model.Fielder, done func(error)) {
			done(nil)
		})

		err := f.Submit()
		if err != nil {
			t.Fatalf("Submit failed: %v", err)
		}

		val := f.Input("nombre").(interface{ GetValue() string }).GetValue()
		if val != "" {
			t.Errorf("Expected value to be reset (empty), got %q", val)
		}
	})

	t.Run("TestSubmit_ResetClearsValues", func(t *testing.T) {
		s := &submitStruct{Nombre: "Jules"}
		f, _ := form.New("app", s, &testIDGen{})
		f.SetValues("nombre", "New Value")

		f.Reset()

		val := f.Input("nombre").(interface{ GetValue() string }).GetValue()
		if val != "" {
			t.Errorf("Expected empty value after reset, got %q", val)
		}
	})
}
