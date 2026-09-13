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

type svRecord struct {
	Code string
}

func (r *svRecord) Schema() []model.Field {
	return []model.Field{{Name: "code", Type: input.Text(), NotNull: true}}
}
func (r *svRecord) Values() []any    { return []any{r.Code} }
func (r *svRecord) Pointers() []any  { return []any{&r.Code} }
func (r *svRecord) FormName() string { return "sv" }

// renderSVForm mounts a one-field form and returns the field element and its
// error span. skip decides whether the field declares SetSkipValidation(true).
func renderSVForm(t *testing.T, mountID string, skip bool) (js.Value, js.Value) {
	t.Helper()
	doc := js.Global().Get("document")
	mount := doc.Call("createElement", "div")
	mount.Set("id", mountID)
	doc.Get("body").Call("appendChild", mount)

	f, err := form.New(mountID, &svRecord{}, &testIDGen{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}
	if skip {
		inp := f.Input("code")
		if inp == nil {
			t.Fatal("f.Input(\"code\") returned nil")
		}
		setter, ok := inp.(interface{ SetSkipValidation(bool) })
		if !ok {
			t.Fatal("input does not implement SetSkipValidation")
		}
		setter.SetSkipValidation(true)
	}
	if err := dom.Render(mountID, f); err != nil {
		t.Fatalf("dom.Render: %v", err)
	}

	el := doc.Call("getElementById", mountID+".sv.code")
	if el.IsNull() || el.IsUndefined() {
		t.Fatalf("input element #%s.sv.code not found after render", mountID)
	}
	errSpan := doc.Call("getElementById", mountID+".sv.code.error")
	if errSpan.IsNull() || errSpan.IsUndefined() {
		t.Fatalf("error span #%s.sv.code.error not found after render", mountID)
	}
	return el, errSpan
}

// typeInto simulates a real edit: set the value and fire "input", which is the
// event fieldComponent binds its live validation to.
func typeInto(el js.Value, value string) {
	el.Set("value", value)
	el.Call("dispatchEvent", js.Global().Get("Event").New("input"))
}

// invalidForText is a value input.Text() rejects: "-" is outside its charset
// (Letters, Tilde, Numbers, Spaces, Extra '.', ',', '(', ')'). This is the very
// character a Chilean RUT needs, which is how the gap below was found.
const invalidForText = "12345678-9"

// TestLiveValidation_ShowsErrorByDefault is the control: without
// SetSkipValidation, typing a value the charset rejects must surface the error
// live, on the keystroke — the behavior every ordinary field depends on.
func TestLiveValidation_ShowsErrorByDefault(t *testing.T) {
	el, errSpan := renderSVForm(t, "sv-mount-default", false)

	typeInto(el, invalidForText)

	if got := errSpan.Get("textContent").String(); got == "" {
		t.Error("expected a live validation error for a rejected character, got none")
	}
}

// TestLiveValidation_SilentWhenSkipValidation guards the fix: Form.Validate()
// (which gates submission) already honored SetSkipValidation, but the live
// per-keystroke path did not — so a field explicitly opted out of validation
// still flashed a red error as the user typed, revealing a format it was
// supposed to keep to itself. Both paths must agree.
func TestLiveValidation_SilentWhenSkipValidation(t *testing.T) {
	el, errSpan := renderSVForm(t, "sv-mount-skip", true)

	typeInto(el, invalidForText)

	if got := errSpan.Get("textContent").String(); got != "" {
		t.Errorf("expected no live error when SetSkipValidation(true), got %q", got)
	}
}
