package form_test

import (
	"strings"
	"testing"

	"webtyp.com/form"
	"webtyp.com/model"
)

// rpcMessage is what ormc generates from a model.Definition{Wire: true}.
type rpcMessage struct{ Method string }

func (m *rpcMessage) Schema() []model.Field { return []model.Field{{Name: "method", Type: model.Text()}} }
func (m *rpcMessage) Pointers() []any       { return []any{&m.Method} }
func (m *rpcMessage) Wire() bool            { return true }

// TestNew_RefusesWireModel: a transport format (JSON-RPC message) is never a
// form — its field names are protocol, not translated screen text.
func TestNew_RefusesWireModel(t *testing.T) {
	_, err := form.New("parent", &rpcMessage{}, &testIDGen{})
	if err == nil || !strings.Contains(err.Error(), "transport format") {
		t.Fatalf("form.New must refuse a Wire model, got err = %v", err)
	}
}
