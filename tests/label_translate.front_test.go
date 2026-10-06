//go:build wasm

package form_test

import (
	"strings"
	"testing"
	"syscall/js"

	"webtyp.com/fmt"
	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/lang"
	"webtyp.com/model"
)

type translateStruct struct {
	model.Fielder
	IP       string
	IsActive string
	Device   string
}

func (s *translateStruct) Schema() []model.Field {
	return []model.Field{
		{Name: "ip", Label: "IP address", Type: func() input.Input { i := input.Text(); i.(interface{ SetPlaceholder(...string) }).SetPlaceholder("example:", "192.168.1.1"); return i }()},
		{Name: "is_active", Type: input.Checkbox()},
		{Name: "device", Type: input.Radio(
			fmt.KeyValue{Key: "computer", Value: "Computer"},
		)},
	}
}

func (s *translateStruct) Pointers() []any {
	return []any{&s.IP, &s.IsActive, &s.Device}
}
func (s *translateStruct) Values() []any {
	return []any{s.IP, s.IsActive, s.Device}
}

func TestTranslateWASM(t *testing.T) {
	script := js.Global().Get("document").Call("createElement", "script")
	script.Set("type", "application/json")
	script.Set("id", lang.ScriptID)
	script.Set("textContent", `{"default":"es","languages":["es"],"keys":{"Computer":["Computador"],"IP address":["Dirección IP"],"Submit":["Guardar"],"example:":["ejemplo:"],"is active":["Activo"]}}`)
	js.Global().Get("document").Get("head").Call("appendChild", script)

	lang.OutLang(lang.ES)

	s := &translateStruct{}
	f, err := form.New("app", s, &testIDGen{})
	if err != nil {
		t.Fatalf("form.New failed: %v", err)
	}
	f.SubmitLabel("Submit")
    lang.Translate("Submit") // Make sure the dictionary sees it? Actually the render resolves it.

    html := f.String()

	if !strings.Contains(html, "Dirección IP") {
		t.Errorf("Expected 'Dirección IP' in HTML, got:\n%s", html)
	}
	if !strings.Contains(html, ">Activo</span>") && !strings.Contains(html, "title='Activo'") {
		t.Errorf("Expected 'Activo' in HTML, got:\n%s", html)
	}

	if !strings.Contains(html, "Computador") {
		t.Errorf("Expected 'Computador' in HTML, got:\n%s", html)
	}
	if !strings.Contains(html, "value='computer'") {
		t.Errorf("Expected value='computer' in HTML, got:\n%s", html)
	}

	if !strings.Contains(html, "placeholder='ejemplo: 192.168.1.1'") {
		t.Errorf("Expected placeholder 'ejemplo: 192.168.1.1' in HTML, got:\n%s", html)
	}
}
