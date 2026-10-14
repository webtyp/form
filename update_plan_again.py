import re

with open('docs/PLAN.md', 'r') as f:
    content = f.read()

tests_replacement = """
## Stage 5 — tests (`tests/`)

Test behaviour only, never the look.
1. DELETE existing tests whose only assertion is the look (emitted CSS declarations, chip/legend classes, label truncation, child order: `styling_test.go`, `css_test.go` and look-only cases in `label.*_test.go`, `render_input_test.go`, `anatomy*_test.go`).
2. Keep every behaviour test (values, validation, submit, dirty, locked, options).
3. Add `tests/validate_on_blur.front_test.go` (wasm, like `onfieldchange.front_test.go`): typing an invalid value shows no error; blur shows it; typing a valid value then clears it without a blur.
4. Add `tests/label_missing_test.go`: `form.New` over a model whose widget field has no `Label` calls the logger set with `dom.SetLog` once with `errFieldWithoutLabel` text and the field name; the label still renders the humanized name.
"""

content = re.sub(
    r'## Stage 5 — tests \(`tests/`\).*?## Stage 6 — docs',
    tests_replacement.strip() + '\n\n## Stage 6 — docs',
    content,
    flags=re.DOTALL
)

with open('docs/PLAN.md', 'w') as f:
    f.write(content)
