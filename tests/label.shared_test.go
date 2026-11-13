package form_test

import (
	"strings"
	"testing"

	"webtyp.com/form"
	"webtyp.com/input"
	"webtyp.com/model"
)

type labelStruct struct {
	model.Fielder
	IP     string
	Active string
}

func (s *labelStruct) Schema() []model.Field {
	return []model.Field{
		{Name: "ip", Label: "IP address", Help: "Format: 192.168.1.1", Type: input.IP()},
		{Name: "is_active", Type: input.Checkbox()},
	}
}

func (s *labelStruct) Pointers() []any {
	return []any{&s.IP, &s.Active}
}

func (s *labelStruct) Values() []any {
	return []any{s.IP, s.Active}
}

func runLabelTests(t *testing.T) {
	s := &labelStruct{}
	f, _ := form.New("app", s, &testIDGen{})
	html := f.String()

	t.Run("Label rendering", func(t *testing.T) {
		if !strings.Contains(html, "IP address") {
			t.Errorf("Expected IP address label, got: %s", html)
		}
		if !strings.Contains(html, "is active") {
			t.Errorf("Expected 'is active' (humanized) label, got: %s", html)
		}
	})

	t.Run("Help text rendering", func(t *testing.T) {
		// The help text is in the markup, hidden until its (i) is tapped: the
		// button starts collapsed and names the panel it opens.
		if !strings.Contains(html, "Format: 192.168.1.1") {
			t.Errorf("Expected Format: 192.168.1.1, got: %s", html)
		}
		if !strings.Contains(html, "aria-expanded='false'") {
			t.Errorf("Expected the help button to start collapsed, got: %s", html)
		}
		if !strings.Contains(html, "aria-controls=") {
			t.Errorf("Expected the help button to name its panel, got: %s", html)
		}
		// The control is described by the help text
		if !strings.Contains(html, "aria-describedby=") {
			t.Errorf("Expected aria-describedby, got: %s", html)
		}
	})

	t.Run("Field without Help has no help button", func(t *testing.T) {
		if strings.Count(html, "aria-expanded=") != 1 {
			t.Errorf("Expected exactly one help button (the IP field), got: %s", html)
		}
	})
}
