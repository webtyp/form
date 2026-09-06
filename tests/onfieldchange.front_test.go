//go:build wasm

package form_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/dom"
	"webtyp.com/input"
	"webtyp.com/model"

	"webtyp.com/form"
)

type ofcRecord struct {
	Name string
}

func (r *ofcRecord) Schema() []model.Field {
	return []model.Field{{Name: "name", Type: input.Text(), NotNull: true}}
}
func (r *ofcRecord) Values() []any    { return []any{r.Name} }
func (r *ofcRecord) Pointers() []any  { return []any{&r.Name} }
func (r *ofcRecord) FormName() string { return "ofc" }

// TestOnFieldChange_FiresOnBlur guards the auto-save hook crudview relies on:
// committing a field (focus, edit, blur — exactly what a real user does tabbing
// away) must invoke the callback registered via Form.OnFieldChange, and the
// form's own value must already be updated by the time it fires.
func TestOnFieldChange_FiresOnBlur(t *testing.T) {
	doc := js.Global().Get("document")
	mount := doc.Call("createElement", "div")
	mount.Set("id", "ofc-mount")
	doc.Get("body").Call("appendChild", mount)

	rec := &ofcRecord{Name: "initial"}
	f, err := form.New("ofc-mount", rec, &testIDGen{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}

	var fired int
	var valueAtCommit string
	f.OnFieldChange(func() {
		fired++
		// SyncValues reads the live value SIGNAL (the WASM source of truth) — the
		// same path crudview's saveAction uses. GetValue() would read a stale
		// SSR-only mirror and must NOT be used to verify a live commit.
		snap := &ofcRecord{}
		_ = f.SyncValues(snap)
		valueAtCommit = snap.Name
	})

	if err := dom.Render("ofc-mount", f); err != nil {
		t.Fatalf("dom.Render: %v", err)
	}

	// id = parentID + "." + structName (FormName()) + "." + fieldName.
	el := doc.Call("getElementById", "ofc-mount.ofc.name")
	if el.IsNull() || el.IsUndefined() {
		t.Fatal("input element #ofc-mount.ofc.name not found after render")
	}

	// Simulate real user interaction: edit (input event — this is what updates
	// the bound value signal), then blur (the commit point). A dispatched Event
	// invokes listeners regardless of real window/tab focus state, matching the
	// pattern webtyp/dom's own TestTwoWayInput uses for "input".
	el.Set("value", "changed")
	el.Call("dispatchEvent", js.Global().Get("Event").New("input"))
	el.Call("dispatchEvent", js.Global().Get("Event").New("blur"))

	if fired != 1 {
		t.Fatalf("OnFieldChange fired %d times, want 1", fired)
	}
	if valueAtCommit != "changed" {
		t.Errorf("value at commit = %q, want %q (callback must run AFTER the value syncs)", valueAtCommit, "changed")
	}
}
