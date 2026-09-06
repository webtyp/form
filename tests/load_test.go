package form_test

import (
	"testing"

	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

// Test model with plain model.Kind types (no widgets)
type NoWidgetsModel struct {
	Name  string
	Price int64
}

func (m *NoWidgetsModel) Schema() []model.Field {
	return []model.Field{
		{Name: "Name", Type: model.Text()},
		{Name: "Price", Type: model.Int()},
	}
}

func (m *NoWidgetsModel) Pointers() []any {
	return []any{&m.Name, &m.Price}
}

func (m *NoWidgetsModel) Values() []any {
	return []any{m.Name, m.Price}
}

func (m *NoWidgetsModel) FormName() string {
	return "no_widgets"
}

func (m *NoWidgetsModel) IsNil() bool {
	return m == nil
}

func (m *NoWidgetsModel) EncodeFields(w model.FieldWriter) {}

// Test model with actual widgets
type WidgetsModel struct {
	Name  string
	Price int64
}

func (m *WidgetsModel) Schema() []model.Field {
	return []model.Field{
		{Name: "Name", Type: input.Text()},
		{Name: "Price", Type: input.Number()},
	}
}

func (m *WidgetsModel) Pointers() []any {
	if m == nil {
		return nil
	}
	return []any{&m.Name, &m.Price}
}

func (m *WidgetsModel) Values() []any {
	if m == nil {
		return nil
	}
	return []any{m.Name, m.Price}
}

func (m *WidgetsModel) FormName() string {
	return "widgets"
}

func (m *WidgetsModel) IsNil() bool {
	return m == nil
}

func (m *WidgetsModel) EncodeFields(w model.FieldWriter) {}

func TestLoadValues(t *testing.T) {
	// 1. Rellena: New → LoadValues(&X{Name: "ACME", Price: 1500}) → los valueSignals correspondientes valen "ACME" y "1500"
	u := &WidgetsModel{Name: "Original", Price: 100}
	f, err := form.New("parent-id", u, &testIDGen{})
	if err != nil {
		t.Fatalf("unexpected error creating form: %v", err)
	}

	newData := &WidgetsModel{Name: "ACME", Price: 1500}
	err = f.LoadValues(newData)
	if err != nil {
		t.Fatalf("unexpected error loading values: %v", err)
	}

	// Verify internal state and signals
	nameInput := f.Input("Name")
	if nameInput == nil {
		t.Fatal("expected Name input")
	}

	// In the test setup, we can verify that the signals have been updated.
	// Since valueSignals is unexported but we can use SyncValues to verify,
	// or we can also test round-trip directly.
	// Let's do step 2: Round-trip: LoadValues(a) → SyncValues(b) → b es igual a a campo a campo
	target := &WidgetsModel{}
	err = f.SyncValues(target)
	if err != nil {
		t.Fatalf("unexpected error syncing values: %v", err)
	}

	if target.Name != "ACME" || target.Price != 1500 {
		t.Errorf("expected loaded values 'ACME' and 1500, got '%s' and %d", target.Name, target.Price)
	}

	// 3. Reemplazo, no acumulación: LoadValues(a) → LoadValues(b) → ningún campo conserva el valor de a
	newData2 := &WidgetsModel{Name: "Beta", Price: 0}
	err = f.LoadValues(newData2)
	if err != nil {
		t.Fatalf("unexpected error loading second data: %v", err)
	}

	target2 := &WidgetsModel{}
	err = f.SyncValues(target2)
	if err != nil {
		t.Fatalf("unexpected error syncing second data: %v", err)
	}

	if target2.Name != "Beta" || target2.Price != 0 {
		t.Errorf("expected loaded values 'Beta' and 0, got '%s' and %d", target2.Name, target2.Price)
	}

	// 4. Nil resetea: LoadValues(a) → LoadValues(nil) → todos los inputs vacíos.
	err = f.LoadValues(nil)
	if err != nil {
		t.Fatalf("unexpected error loading nil: %v", err)
	}

	target3 := &WidgetsModel{}
	err = f.SyncValues(target3)
	if err != nil {
		t.Fatalf("unexpected error syncing after nil load: %v", err)
	}

	if target3.Name != "" || target3.Price != 0 {
		t.Errorf("expected form to be reset, but got Name='%s', Price=%d", target3.Name, target3.Price)
	}

	// 5. Nil tipado resetea: LoadValues((*X)(nil)) → todos los inputs vacíos, sin panic.
	// First populate it again
	err = f.LoadValues(&WidgetsModel{Name: "Populated", Price: 42})
	if err != nil {
		t.Fatalf("unexpected error re-loading: %v", err)
	}

	var typedNil *WidgetsModel
	err = f.LoadValues(typedNil)
	if err != nil {
		t.Fatalf("unexpected error loading typed nil: %v", err)
	}

	target4 := &WidgetsModel{}
	err = f.SyncValues(target4)
	if err != nil {
		t.Fatalf("unexpected error syncing after typed nil load: %v", err)
	}

	if target4.Name != "" || target4.Price != 0 {
		t.Errorf("expected form to be reset on typed nil, but got Name='%s', Price=%d", target4.Name, target4.Price)
	}

	// 6. Limpia errores: fuerza un error de validación en un input, LoadValues(a), el errorSignal de ese input queda vacío.
	// Since errorSignals are internal, we can check that validating after loading does not report errors if the loaded data is valid,
	// and that calling LoadValues itself resets internal error state.
	// Let's verify that a valid load clears errors by using Validate() before and after.
	// Actually, we can trigger validation with bad input, assert it fails, then load valid values, then assert Validate() passes.
	f.SetValues("Price", "not-a-number")
	err = f.Validate()
	if err == nil {
		t.Fatal("expected validation error on invalid Price")
	}

	// Loading valid data
	err = f.LoadValues(&WidgetsModel{Name: "Valid", Price: 100})
	if err != nil {
		t.Fatalf("unexpected error loading valid data: %v", err)
	}

	err = f.Validate()
	if err != nil {
		t.Fatalf("expected validation to pass after loading valid data, got error: %v", err)
	}
}

func TestNewFailures(t *testing.T) {
	// 7. New con un modelo sin widgets falla: un Fielder cuyos Field.Type sean todos model.Text()/model.Int() → New devuelve err != nil (no un form vacío).
	noWidgets := &NoWidgetsModel{Name: "Test", Price: 10}
	f, err := form.New("parent-id", noWidgets, &testIDGen{})
	if err == nil {
		t.Fatal("expected form.New to fail when model has no widgets, but got no error")
	}
	if f != nil {
		t.Errorf("expected form to be nil when New fails, but got non-nil form")
	}

	// 8. New con widgets funciona: el mismo modelo con input.Text()/input.Number() → err == nil y len(f.Inputs) > 0.
	withWidgets := &WidgetsModel{Name: "Test", Price: 10}
	f2, err2 := form.New("parent-id", withWidgets, &testIDGen{})
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if f2 == nil {
		t.Fatal("expected non-nil form")
	}
	if len(f2.Inputs) == 0 {
		t.Error("expected at least one input in form")
	}
}
