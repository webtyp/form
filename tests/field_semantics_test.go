//go:build !wasm

package form_test

import (
	"strings"
	"testing"

	"webtyp.com/fmt"
	"webtyp.com/form"
	"webtyp.com/input"
)

func renderKind(t *testing.T, inp input.Input, opts []fmt.KeyValue, values ...string) string {
	t.Helper()
	f, err := form.New("app", &kindFixture{inp: inp}, &testIDGen{})
	if err != nil {
		t.Fatalf("form.New: %v", err)
	}
	if len(opts) > 0 {
		f.SetOptions("tfield", opts...)
	}
	if len(values) > 0 {
		f.SetValues("tfield", values...)
	}
	return f.String()
}

// A lone checkbox says its text once, inside the option; a second label above
// it would repeat it.
func TestCheckboxHasNoLabelAboveIt(t *testing.T) {
	html := renderKind(t, input.Checkbox(), nil)
	if fmt.Contains(html, "<label class='vy-field__label'") || fmt.Contains(html, `<label class="vy-field__label"`) {
		t.Errorf("a checkbox must not render a label above its option\ngot: %s", html)
	}
}

// A select with nothing chosen shows a prompt instead of presenting its first
// option as if it had been picked; once a value exists the prompt is not selected.
func TestSelectPromptsUntilChosen(t *testing.T) {
	opts := []fmt.KeyValue{{Key: "1", Value: "Admin"}, {Key: "2", Value: "Editor"}}

	// promptTag returns the opening <option …> tag of the prompt, or "".
	promptTag := func(html string) string {
		i := strings.Index(html, "Select an option")
		if i < 0 {
			return ""
		}
		start := strings.LastIndex(html[:i], "<option")
		return html[start:i]
	}

	empty := promptTag(renderKind(t, input.Select(), opts))
	if empty == "" {
		t.Fatal("an unchosen select must offer a prompt option")
	}
	if !strings.Contains(empty, "selected") {
		t.Errorf("the prompt must be the selected option while nothing is chosen, got %q", empty)
	}

	chosen := promptTag(renderKind(t, input.Select(), opts, "2"))
	if strings.Contains(chosen, "selected") {
		t.Errorf("the prompt must not stay selected once a value exists, got %q", chosen)
	}
}

// The generated default title is the field's technical name; it must not
// become a hover tooltip or the control's accessible name.
func TestNoTooltipFromTheFieldName(t *testing.T) {
	html := renderKind(t, input.Text(), nil) // SetName defaults the title to the field name
	if fmt.Contains(html, "title='tfield'") || fmt.Contains(html, `title="tfield"`) {
		t.Errorf("a title equal to the field name must not be rendered\ngot: %s", html)
	}
}
