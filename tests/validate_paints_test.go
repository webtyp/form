package form_test

import (
	"strings"
	"testing"

	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

// paintRecord is a two-field form: code arrives with a present-but-invalid
// value (never typed, so live validation never ran on it), name arrives
// empty while required.
type paintRecord struct {
	Code string
	Name string
}

func (r *paintRecord) Schema() []model.Field {
	return []model.Field{
		{Name: "code", Type: input.Text()},
		{Name: "name", Type: input.Text(), NotNull: true},
	}
}
func (r *paintRecord) Pointers() []any {
	if r == nil {
		return nil
	}
	return []any{&r.Code, &r.Name}
}
func (r *paintRecord) FormName() string { return "paint" }
func (r *paintRecord) IsNil() bool      { return r == nil }

// errorSpanText extracts the text inside a field's error span from the
// rendered form, or "" when the span is empty/absent.
func errorSpanText(t *testing.T, f *form.Form, field string) string {
	t.Helper()
	html := f.Render().String()
	marker := "id='p.paint." + field + ".error'"
	idx := strings.Index(html, marker)
	if idx == -1 {
		t.Fatalf("error span for %q not found in:\n%s", field, html)
	}
	rest := html[idx:]
	endAttrs := strings.Index(rest, ">")
	if endAttrs == -1 {
		t.Fatalf("malformed error span for %q", field)
	}
	content := rest[endAttrs+1:]
	end := strings.Index(content, "</span>")
	if end == -1 {
		t.Fatalf("unclosed error span for %q", field)
	}
	return content[:end]
}

// A Validate over a loaded-but-never-typed invalid value must paint that
// field's error: returning the error while the form shows nothing is a
// silent failure. An empty field that fails stays unpainted — in a fresh
// draft the user has not reached it yet.
func TestValidatePaintsLoadedInvalidValue(t *testing.T) {
	f, err := form.New("p", &paintRecord{}, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	// "-" is outside input.Text()'s charset: present and invalid, never typed.
	if err := f.LoadValues(&paintRecord{Code: "12345678-9", Name: ""}); err != nil {
		t.Fatal(err)
	}
	if err := f.Validate(); err == nil {
		t.Fatal("expected a validation error for the loaded invalid code, got nil")
	}
	if got := errorSpanText(t, f, "code"); got == "" {
		t.Error("expected the loaded invalid 'code' field to be painted, got an empty error span")
	}
	if got := errorSpanText(t, f, "name"); got != "" {
		t.Errorf("expected the empty 'name' field to stay unpainted, got %q", got)
	}
}

// Fixing the value clears the painted error on the next Validate.
func TestValidateClearsPaintedErrorOnceFixed(t *testing.T) {
	f, err := form.New("p", &paintRecord{}, &testIDGen{})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.LoadValues(&paintRecord{Code: "12345678-9", Name: ""}); err != nil {
		t.Fatal(err)
	}
	if err := f.Validate(); err == nil {
		t.Fatal("expected a validation error, got nil")
	}
	if got := errorSpanText(t, f, "code"); got == "" {
		t.Fatal("sanity check failed: 'code' was not painted")
	}
	f.SetValues("code", "Valid Code")
	f.SetValues("name", "Valid Name")
	if err := f.Validate(); err != nil {
		t.Fatalf("expected validation to pass after fixing both fields, got: %v", err)
	}
	if got := errorSpanText(t, f, "code"); got != "" {
		t.Errorf("expected the fixed 'code' error to be cleared, got %q", got)
	}
}
