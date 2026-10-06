//go:build wasm

package form_test

import "testing"

func TestLabelAndHelpWASM(t *testing.T) {
	runLabelTests(t)
}
