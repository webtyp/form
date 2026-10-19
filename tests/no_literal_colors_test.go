//go:build !wasm

package form_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every colour a form paints comes from its stylesheet (css.go), through the
// style recipes, so the theme decides gradient or solid for all of them. A
// colour literal in the markup builders is a second, hand-written way to paint
// the primary — the inconsistency this guards against (a check drawn in a solid
// colour next to a gradient ring).
func TestRenderersWriteNoColors(t *testing.T) {
	files, err := filepath.Glob("../render*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("no renderer sources found: %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(src), "--color-") {
			t.Errorf("%s writes a colour token by hand: paint it from css.go with a style recipe", filepath.Base(f))
		}
	}
}
