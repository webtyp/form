
## Field Formats

The layout is a one-column stack. Labels render fully (no truncation) above the control. Errors (if any) replace the help text below the control. If a field lacks a `Label` in its `model.Field`, the engine logs a diagnostic (`webtyp/form: field has a widget but no Label — set model.Field.Label:`) and falls back to a humanized struct field name.
