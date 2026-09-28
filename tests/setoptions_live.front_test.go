//go:build wasm

package form_test

import (
	"syscall/js"
	"testing"

	"webtyp.com/dom"
	"webtyp.com/fmt"
	"webtyp.com/input"
	"webtyp.com/model"

	"webtyp.com/form"
)

type solRecord struct {
	FloorId string
}

func (r *solRecord) Schema() []model.Field {
	return []model.Field{{Name: "floor_id", Type: input.Select(), NotNull: true}}
}
func (r *solRecord) Values() []any    { return []any{r.FloorId} }
func (r *solRecord) Pointers() []any  { return []any{&r.FloorId} }
func (r *solRecord) FormName() string { return "sol" }

// TestSetOptions_AfterRender_RepaintsSelect guards the async-options case every
// consumer hits: the choices of a <select> come from a Caller whose answer
// arrives AFTER the form is on screen. Form.SetOptions at that moment must
// reach the live <select>, not only the input's slice that the next full
// render would read.
func TestSetOptions_AfterRender_RepaintsSelect(t *testing.T) {
	doc := js.Global().Get("document")
	mount := doc.Call("createElement", "div")
	mount.Set("id", "sol-mount")
	doc.Get("body").Call("appendChild", mount)

	f, err := form.New("sol-mount", &solRecord{}, &testIDGen{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}
	if err := dom.Render("sol-mount", f); err != nil {
		t.Fatalf("dom.Render: %v", err)
	}

	sel := doc.Call("querySelector", "#sol-mount select[name='floor_id']")
	if sel.IsNull() || sel.IsUndefined() {
		t.Fatal("select floor_id not rendered")
	}
	if n := sel.Get("options").Get("length").Int(); n != 0 {
		t.Fatalf("before SetOptions: %d options, want 0", n)
	}

	f.SetOptions("floor_id",
		fmt.KeyValue{Key: "f1", Value: "Piso 1"},
		fmt.KeyValue{Key: "f2", Value: "Piso 2"},
	)

	sel = doc.Call("querySelector", "#sol-mount select[name='floor_id']")
	opts := sel.Get("options")
	if n := opts.Get("length").Int(); n != 2 {
		t.Fatalf("after SetOptions: live <select> has %d options, want 2 — the choices never reach the screen", n)
	}
	if v := opts.Call("item", 1).Get("value").String(); v != "f2" {
		t.Errorf("second option value = %q, want f2", v)
	}
	if txt := opts.Call("item", 1).Get("textContent").String(); txt != "Piso 2" {
		t.Errorf("second option text = %q, want Piso 2", txt)
	}
}
