package form_test

import "webtyp.com/model"

import (
	"testing"

	"webtyp.com/fmt"
	"webtyp.com/form"
	"webtyp.com/input"
)

func TestForm_NewAndBinding_Shared(t *testing.T) {
	f := createTestForm()

	// Helper to get value via type assertion
	getValue := func(inp interface{}) string {
		if g, ok := inp.(interface{ GetValue() string }); ok {
			return g.GetValue()
		}
		return ""
	}

	// Check if values were bound correctly
	nameInp := f.Input("Name")
	if nameInp == nil {
		t.Fatal("Input 'Name' not found")
	}
	if getValue(nameInp) != "John Doe" {
		t.Errorf("Expected 'John Doe', got '%s'", getValue(nameInp))
	}

	emailInp := f.Input("Email")
	if getValue(emailInp) != "john@example.com" {
		t.Errorf("Expected 'john@example.com', got '%s'", getValue(emailInp))
	}

	genderInp := f.Input("Gender")
	if getValue(genderInp) != "m" {
		t.Errorf("Expected 'm', got '%s'", getValue(genderInp))
	}
}

func TestForm_String_SSR_Shared(t *testing.T) {
	f := createTestForm()
	f.SetSSR(true)

	html := f.String()

	// Verify SSR specific attributes
	if !fmt.Contains(html, `method='POST'`) {
		t.Error("SSR form should contain method='POST'")
	}

	// Verify inputs are present
	if !fmt.Contains(html, `name='Name'`) {
		t.Error("SSR form should contain input name='Name'")
	}
	if !fmt.Contains(html, `value='John Doe'`) {
		t.Error("SSR form should contain value='John Doe'")
	}
}

func TestForm_AutoDefaults_Shared(t *testing.T) {
	f := createTestForm()

	// Title defaults to the raw field name. Placeholder does NOT — a host UI
	// already shows the field name as a label chip (components/fieldset), so
	// defaulting the placeholder to the same text would just repeat it.
	nameInp := f.Input("Name")
	if p, ok := nameInp.(interface{ GetPlaceholder() string }); ok {
		if p.GetPlaceholder() != "" {
			t.Errorf("Expected no default placeholder, got %q", p.GetPlaceholder())
		}
	}

	if p, ok := nameInp.(interface{ GetTitle() string }); ok {
		if p.GetTitle() != "Name" {
			t.Errorf("Expected 'Name', got '%s'", p.GetTitle())
		}
	}
}

func TestForm_Validate_Shared(t *testing.T) {
	f := createTestForm()

	// Valid form
	if err := f.Validate(); err != nil {
		t.Errorf("Expected valid form, got error: %v", err)
	}

	// Invalid form
	f.SetValues("Name", "") // Too short (min 2)
	if err := f.Validate(); err == nil {
		t.Error("Expected validation error for empty name, got nil")
	}

	// Reset Name to valid value
	f.SetValues("Name", "John Doe")
}

type CustomUser struct {
	Special string
}

func (u *CustomUser) Schema() []model.Field {
	return []model.Field{{Name: "Special", Type: input.Text()}}
}
func (u *CustomUser) Values() []any    { return []any{u.Special} }
func (u *CustomUser) Pointers() []any  { return []any{&u.Special} }
func (u *CustomUser) FormName() string { return "customuser" }

func TestForm_CustomInput_Shared(t *testing.T) {
	// Struct with a field that has a Widget
	f, err := form.New("parent", &CustomUser{}, &testIDGen{})
	if err != nil {
		t.Fatalf("Failed to create form with widget: %v", err)
	}

	if f.Input("Special") == nil {
		t.Error("Input 'Special' not found")
	}
}

func TestForm_ValidateData_Shared(t *testing.T) {
	f := createTestForm()

	// Valid struct: all fields filled with valid values — should pass
	valid := &User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "secret123",
		Gender:   "m",
		Role:     "admin",
		Address:  "123 Main St",
	}
	if err := f.ValidateData('c', valid); err != nil {
		t.Errorf("Expected valid data to pass, got error: %v", err)
	}

	// Invalid email — should fail
	invalid := &User{
		Name:     "John",
		Email:    "not-an-email",
		Password: "secret123",
	}
	if err := f.ValidateData('c', invalid); err == nil {
		t.Error("Expected invalid email to fail ValidateData, got nil")
	}
}

// runSharedTests executes all test cases common to both WASM and Standard Lib.
func runSharedTests(t *testing.T) {
	t.Run("NewAndBinding", TestForm_NewAndBinding_Shared)
	t.Run("AutoDefaults", TestForm_AutoDefaults_Shared)
	t.Run("CustomInput", TestForm_CustomInput_Shared)
	t.Run("String_SSR", TestForm_String_SSR_Shared)
	t.Run("Validate", TestForm_Validate_Shared)
	t.Run("ValidateData", TestForm_ValidateData_Shared)
}
