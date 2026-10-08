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

type solRecordMulti struct {
	A string
	B string
}

func (r *solRecordMulti) Schema() []model.Field {
	return []model.Field{
		{Name: "a", Type: input.Select(), NotNull: true},
		{Name: "b", Type: input.Select(), NotNull: true},
	}
}
func (r *solRecordMulti) Values() []any    { return []any{r.A, r.B} }
func (r *solRecordMulti) Pointers() []any  { return []any{&r.A, &r.B} }
func (r *solRecordMulti) FormName() string { return "sol_multi" }

func TestSetOptions_OnlyUpdatesTargetField(t *testing.T) {
	doc := js.Global().Get("document")
	mount := doc.Call("createElement", "div")
	mount.Set("id", "sol-mount-multi")
	doc.Get("body").Call("appendChild", mount)

	f, err := form.New("sol-mount-multi", &solRecordMulti{}, &testIDGen{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}

	f.SetOptions("a", fmt.KeyValue{Key: "a1", Value: "A 1"})

	if err := dom.Render("sol-mount-multi", f); err != nil {
		t.Fatalf("dom.Render: %v", err)
	}

	selA := doc.Call("querySelector", "#sol-mount-multi select[name='a']")
	if selA.IsNull() || selA.IsUndefined() {
		t.Fatal("select a not rendered")
	}
	optsA := selA.Get("options")
	if n := optsA.Get("length").Int(); n != 1 {
		t.Fatalf("before SetOptions(b): A has %d options, want 1", n)
	}

	selB := doc.Call("querySelector", "#sol-mount-multi select[name='b']")
	optsB := selB.Get("options")
	if n := optsB.Get("length").Int(); n != 0 {
		t.Fatalf("before SetOptions(b): B has %d options, want 0", n)
	}

	f.SetOptions("b",
		fmt.KeyValue{Key: "b1", Value: "B 1"},
	)

	selA = doc.Call("querySelector", "#sol-mount-multi select[name='a']")
	optsA = selA.Get("options")
	if n := optsA.Get("length").Int(); n != 1 {
		t.Fatalf("after SetOptions(b): A has %d options, want 1 (it should not have been updated)", n)
	}

	selB = doc.Call("querySelector", "#sol-mount-multi select[name='b']")
	optsB = selB.Get("options")
	if n := optsB.Get("length").Int(); n != 1 {
		t.Fatalf("after SetOptions(b): B has %d options, want 1", n)
	}
}
