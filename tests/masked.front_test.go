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

type maskedRecord struct {
	Code string
}

func (r *maskedRecord) Schema() []model.Field {
	return []model.Field{{Name: "code", Type: input.Text(), NotNull: true}}
}
func (r *maskedRecord) Values() []any    { return []any{r.Code} }
func (r *maskedRecord) Pointers() []any  { return []any{&r.Code} }
func (r *maskedRecord) FormName() string { return "masked" }

// TestBuildInput_HonorsMasked guards the render side of input.Base.Masked:
// a Text() input marked SetMasked(true) must draw as <input type="password">
// even though its own HTMLName() is still "text" — masking is presentation
// over whatever charset/validation the field already has, never a second
// field type to keep in sync (see docs/PLAN.md in veltylabs/mjosefa-cms for
// why the app used to fake this by overriding HTMLName() by hand).
func TestBuildInput_HonorsMasked(t *testing.T) {
	doc := js.Global().Get("document")
	mount := doc.Call("createElement", "div")
	mount.Set("id", "masked-mount")
	doc.Get("body").Call("appendChild", mount)

	f, err := form.New("masked-mount", &maskedRecord{}, &testIDGen{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}
	inp := f.Input("code")
	if inp == nil {
		t.Fatal("f.Input(\"code\") returned nil")
	}
	setter, ok := inp.(interface{ SetMasked(bool) })
	if !ok {
		t.Fatal("input does not implement SetMasked")
	}
	setter.SetMasked(true)

	if err := dom.Render("masked-mount", f); err != nil {
		t.Fatalf("dom.Render: %v", err)
	}

	el := doc.Call("getElementById", "masked-mount.masked.code")
	if el.IsNull() || el.IsUndefined() {
		t.Fatal("input element #masked-mount.masked.code not found after render")
	}
	if got := el.Get("type").String(); got != "password" {
		t.Errorf("type attribute = %q, want %q for a masked Text() input", got, "password")
	}

	// The mask must not touch the value binding: typing still updates the
	// field like any ordinary input.
	el.Set("value", "12345678-5")
	el.Call("dispatchEvent", js.Global().Get("Event").New("input"))
	snap := &maskedRecord{}
	if err := f.SyncValues(snap); err != nil {
		t.Fatalf("SyncValues: %v", err)
	}
	if snap.Code != "12345678-5" {
		t.Errorf("value after typing = %q, want %q — masking must not affect binding", snap.Code, "12345678-5")
	}
}

// TestRevealToggle_TogglesInputType guards the show/hide button
// wireMaskToggle draws next to a masked input: clicking it must flip the
// input between "password" and "text" and reflect the state on the button
// itself (data-selected), with no JS in the app — see docs/PLAN.md in
// veltylabs/mjosefa-cms for the syscall/js version this replaced.
func TestRevealToggle_TogglesInputType(t *testing.T) {
	doc := js.Global().Get("document")
	mount := doc.Call("createElement", "div")
	mount.Set("id", "reveal-mount")
	doc.Get("body").Call("appendChild", mount)

	f, err := form.New("reveal-mount", &maskedRecord{}, &testIDGen{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}
	f.Input("code").(interface{ SetMasked(bool) }).SetMasked(true)

	if err := dom.Render("reveal-mount", f); err != nil {
		t.Fatalf("dom.Render: %v", err)
	}

	input := doc.Call("getElementById", "reveal-mount.masked.code")
	if input.IsNull() || input.IsUndefined() {
		t.Fatal("input element not found after render")
	}
	btn := doc.Call("getElementById", "reveal-mount.masked.code.reveal")
	if btn.IsNull() || btn.IsUndefined() {
		t.Fatal("reveal button #reveal-mount.masked.code.reveal not found for a masked field")
	}
	if got := btn.Get("type").String(); got != "button" {
		t.Errorf("reveal button type = %q, want %q — a submit type would send the form on click", got, "button")
	}
	if svg := btn.Call("querySelector", "svg"); svg.IsNull() {
		t.Error("reveal button must contain a real <svg> icon, not an emoji or bare text")
	}
	if input.Get("type").String() != "password" {
		t.Fatalf("input must start as type=password")
	}

	btn.Call("dispatchEvent", js.Global().Get("Event").New("click"))
	if got := input.Get("type").String(); got != "text" {
		t.Errorf("after one click, input type = %q, want %q", got, "text")
	}
	if !btn.Call("hasAttribute", "data-selected").Bool() {
		t.Error("revealed button should carry data-selected=\"true\"")
	}

	btn.Call("dispatchEvent", js.Global().Get("Event").New("click"))
	if got := input.Get("type").String(); got != "password" {
		t.Errorf("after a second click, input type = %q, want %q", got, "password")
	}
	if btn.Call("hasAttribute", "data-selected").Bool() {
		t.Error("hidden button should not carry data-selected")
	}
}

// TestRevealToggle_AbsentWhenNotMasked: an ordinary field gets no reveal
// button — the button only ever exists for the class of field it protects.
func TestRevealToggle_AbsentWhenNotMasked(t *testing.T) {
	doc := js.Global().Get("document")
	mount := doc.Call("createElement", "div")
	mount.Set("id", "no-reveal-mount")
	doc.Get("body").Call("appendChild", mount)

	f, err := form.New("no-reveal-mount", &maskedRecord{}, &testIDGen{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}
	if err := dom.Render("no-reveal-mount", f); err != nil {
		t.Fatalf("dom.Render: %v", err)
	}

	btn := doc.Call("getElementById", "no-reveal-mount.masked.code.reveal")
	if !btn.IsNull() && !btn.IsUndefined() {
		t.Error("an unmasked field must not render a reveal button")
	}
}
