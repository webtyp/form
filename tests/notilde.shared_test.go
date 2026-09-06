package form_test

import (
	"testing"
	"webtyp.com/input"
)

func runNotildeTests(t *testing.T) {
	t.Run("TestNotilde_RejectsAccent", func(t *testing.T) {
		inp := input.SetTilde(input.Text(), false)
		err := inp.Validate("á")
		if err == nil {
			t.Errorf("Expected error for accented character when tilde is disabled")
		}
	})

	t.Run("TestNotilde_AllowsNormal", func(t *testing.T) {
		inp := input.SetTilde(input.Text(), false)
		err := inp.Validate("ab")
		if err != nil {
			t.Errorf("Unexpected error for normal character: %v", err)
		}
	})

	t.Run("TestText_AllowsAccentByDefault", func(t *testing.T) {
		inp := input.Text()
		err := inp.Validate("áb")
		if err != nil {
			t.Errorf("Unexpected error for accented character by default: %v", err)
		}
	})
}
