package form_test

import (
	"testing"

	"webtyp.com/form"
)

// TestIsDirty covers the consumer-shaped use case a host (e.g. crudview's
// auto-save gate) actually relies on: "did the user change anything since
// the record was loaded?" WidgetsModel is the same model.Fielder load_test.go
// already declares in this package.
func TestIsDirty(t *testing.T) {
	f, err := form.New("parent-id", &WidgetsModel{Name: "Original", Price: 100}, &testIDGen{})
	if err != nil {
		t.Fatalf("unexpected error creating form: %v", err)
	}

	// 1. A freshly built form is pristine.
	if f.IsDirty() {
		t.Error("expected a freshly built form to not be dirty")
	}

	// 2. Editing a field makes it dirty.
	f.SetValues("Name", "Changed")
	if !f.IsDirty() {
		t.Error("expected the form to be dirty after SetValues changed a field")
	}

	// 3. Reverting to the original value clears dirty — the check compares
	// current vs baseline, not "has SetValues ever been called".
	f.SetValues("Name", "Original")
	if f.IsDirty() {
		t.Error("expected the form to not be dirty after reverting to the original value")
	}

	// 4. LoadValues re-baselines: loading a DIFFERENT record is not "dirty",
	// it's a fresh pristine state — only edits AFTER the load count.
	if err := f.LoadValues(&WidgetsModel{Name: "Loaded", Price: 200}); err != nil {
		t.Fatalf("unexpected error loading values: %v", err)
	}
	if f.IsDirty() {
		t.Error("expected the form to be pristine immediately after LoadValues")
	}
	f.SetValues("Name", "Edited")
	if !f.IsDirty() {
		t.Error("expected the form to be dirty after editing a freshly loaded record")
	}

	// 5. MarkPristine re-snapshots to the CURRENT values — the host calls this
	// right after a successful save so a later untouched field commit is not
	// considered dirty again.
	f.MarkPristine()
	if f.IsDirty() {
		t.Error("expected the form to be pristine right after MarkPristine")
	}
	f.SetValues("Price", "999")
	if !f.IsDirty() {
		t.Error("expected the form to be dirty after editing again post-MarkPristine")
	}

	// 6. Reset() clears the form AND its baseline together — an empty form is
	// pristine, not "dirty against the last loaded record".
	f.Reset()
	if f.IsDirty() {
		t.Error("expected a Reset() form to not be dirty")
	}
}
