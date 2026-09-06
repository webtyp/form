package form_test

import (
	"reflect"
	"testing"

	"webtyp.com/form"
)

func TestDirtyFieldsIsEmptyOnAFreshForm(t *testing.T) {
	f, err := form.New("parent-id", &WidgetsModel{Name: "Original", Price: 100}, &testIDGen{})
	if err != nil {
		t.Fatalf("unexpected error creating form: %v", err)
	}

	if f.IsDirty() {
		t.Error("expected fresh form to not be dirty")
	}
	fields := f.DirtyFields()
	if fields != nil {
		t.Errorf("expected DirtyFields() to be nil on fresh form, got %v", fields)
	}
}

func TestDirtyFieldsNamesOnlyTheChangedField(t *testing.T) {
	f, err := form.New("parent-id", &WidgetsModel{Name: "Original", Price: 100}, &testIDGen{})
	if err != nil {
		t.Fatalf("unexpected error creating form: %v", err)
	}

	f.SetValues("Name", "Changed")
	expected := []string{"Name"}
	fields := f.DirtyFields()
	if !reflect.DeepEqual(fields, expected) {
		t.Errorf("expected DirtyFields() = %v, got %v", expected, fields)
	}
}

func TestDirtyFieldsNamesEveryChangedField(t *testing.T) {
	f, err := form.New("parent-id", &WidgetsModel{Name: "Original", Price: 100}, &testIDGen{})
	if err != nil {
		t.Fatalf("unexpected error creating form: %v", err)
	}

	f.SetValues("Name", "Changed")
	f.SetValues("Price", "200")
	expected := []string{"Name", "Price"}
	fields := f.DirtyFields()
	if !reflect.DeepEqual(fields, expected) {
		t.Errorf("expected DirtyFields() = %v, got %v", expected, fields)
	}
}

func TestDirtyFieldsClearsAfterMarkPristine(t *testing.T) {
	f, err := form.New("parent-id", &WidgetsModel{Name: "Original", Price: 100}, &testIDGen{})
	if err != nil {
		t.Fatalf("unexpected error creating form: %v", err)
	}

	f.SetValues("Name", "Changed")
	if len(f.DirtyFields()) == 0 {
		t.Fatal("expected non-empty DirtyFields() after SetValues")
	}

	f.MarkPristine()
	if f.DirtyFields() != nil {
		t.Errorf("expected DirtyFields() to be nil after MarkPristine, got %v", f.DirtyFields())
	}
	if f.IsDirty() {
		t.Error("expected IsDirty() to be false after MarkPristine")
	}
}

func TestDirtyFieldsClearsAfterReset(t *testing.T) {
	f, err := form.New("parent-id", &WidgetsModel{Name: "Original", Price: 100}, &testIDGen{})
	if err != nil {
		t.Fatalf("unexpected error creating form: %v", err)
	}

	f.SetValues("Name", "Changed")
	if len(f.DirtyFields()) == 0 {
		t.Fatal("expected non-empty DirtyFields() after SetValues")
	}

	f.Reset()
	if f.DirtyFields() != nil {
		t.Errorf("expected DirtyFields() to be nil after Reset, got %v", f.DirtyFields())
	}
	if f.IsDirty() {
		t.Error("expected IsDirty() to be false after Reset")
	}
}

func TestDirtyFieldsAgreesWithIsDirty(t *testing.T) {
	f, err := form.New("parent-id", &WidgetsModel{Name: "Original", Price: 100}, &testIDGen{})
	if err != nil {
		t.Fatalf("unexpected error creating form: %v", err)
	}

	checkInvariant := func(scenario string) {
		t.Helper()
		dirty := f.IsDirty()
		fields := f.DirtyFields()
		hasDirtyFields := len(fields) > 0

		if dirty != hasDirtyFields {
			t.Errorf("[%s] mismatch: IsDirty() = %v, but len(DirtyFields()) > 0 is %v", scenario, dirty, hasDirtyFields)
		}
		if !dirty && fields != nil {
			t.Errorf("[%s] expected DirtyFields() to be nil when not dirty, got %v", scenario, fields)
		}
	}

	checkInvariant("1. Fresh form")

	f.SetValues("Name", "Changed")
	checkInvariant("2. One field changed")

	f.SetValues("Price", "999")
	checkInvariant("3. Two fields changed")

	f.SetValues("Name", "Original")
	checkInvariant("4. One field reverted to baseline")

	f.SetValues("Price", "100")
	checkInvariant("5. All fields reverted to baseline")

	f.SetValues("Name", "NewVal")
	f.MarkPristine()
	checkInvariant("6. After MarkPristine")

	if err := f.LoadValues(&WidgetsModel{Name: "Loaded", Price: 50}); err != nil {
		t.Fatalf("unexpected error loading values: %v", err)
	}
	checkInvariant("7. After LoadValues")

	f.SetValues("Price", "75")
	checkInvariant("8. Changed after LoadValues")

	f.Reset()
	checkInvariant("9. After Reset")
}

func TestDirtyFieldsIgnoresAValueSetBackToItsOriginal(t *testing.T) {
	f, err := form.New("parent-id", &WidgetsModel{Name: "Original", Price: 100}, &testIDGen{})
	if err != nil {
		t.Fatalf("unexpected error creating form: %v", err)
	}

	f.SetValues("Name", "Changed")
	if !reflect.DeepEqual(f.DirtyFields(), []string{"Name"}) {
		t.Fatalf("expected DirtyFields() to be ['Name'], got %v", f.DirtyFields())
	}

	f.SetValues("Name", "Original")
	if f.DirtyFields() != nil {
		t.Errorf("expected DirtyFields() to be nil after setting back to original, got %v", f.DirtyFields())
	}
	if f.IsDirty() {
		t.Error("expected IsDirty() to be false after setting back to original")
	}
}
