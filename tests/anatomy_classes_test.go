package form_test

import "webtyp.com/widget"

// The field anatomy classes, DERIVED from widget rather than written out.
//
// They used to be hardcoded as "tw-field…" literals across four test files.
// When widget renamed its prefix to "vy-", every one of those assertions
// became a lie — and nothing failed, because the package had stopped compiling
// for an unrelated reason and the breakage stayed hidden until it built again.
//
// The prefix is widget's to choose. A test in THIS package must assert that the
// form emits the field anatomy, not what that anatomy happens to be spelled
// like this month.
var (
	clsField       = widget.NameField.Root().String()
	clsFieldLabel  = widget.NameField.Class(widget.PartLabel).String()
	clsFieldInput  = widget.NameField.Class(widget.PartInput).String()
	clsFieldError  = widget.NameField.Class(widget.PartError).String()
	clsFieldSubmit = widget.NameField.Class(widget.PartSubmit).String()
	clsFieldRadios = widget.NameField.Class(widget.PartRadioGroup).String()
)
