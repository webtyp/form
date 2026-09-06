package form_test

import (
	"testing"

	"webtyp.com/dom"
	"webtyp.com/form"
	"webtyp.com/model"
)

type mockFielder struct {
}

func (m *mockFielder) Schema() []model.Field {
	return []model.Field{}
}

func (m *mockFielder) Pointers() []any {
	return []any{}
}

func (m *mockFielder) Values() []any {
	return []any{}
}

func TestForm_DomComponent(t *testing.T) {
	f, _ := form.New("parent", &mockFielder{}, &testIDGen{})

	// This assignment will fail to compile if *Form does not implement dom.Component
	var _ dom.Component = f
}
