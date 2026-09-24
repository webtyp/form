package form

// Validate validates all inputs and returns the first error found.
//
// A failing field with a value paints its own error signal, so a record
// loaded with an invalid value (never typed, so live validation never ran
// on it) is still visible — a block validate that returns an error without
// painting the culprit is a silent failure. An empty failing field stays
// unpainted: in a fresh draft the user has not reached it yet, and the live
// path (fieldComponent.validate) paints it the moment they type.
func (f *Form) Validate() error {
	var first error
	for i, inp := range f.Inputs {
		// Skip validation if requested via tag
		if skipper, ok := inp.(interface{ GetSkipValidation() bool }); ok && skipper.GetSkipValidation() {
			continue
		}

		// Signal is the source of truth in WASM mode.
		val := f.valueSignals[i].Get()

		// Fallback only if we are somehow in SSR mode where signals might be empty
		if val == "" && f.ssrMode {
			if valuer, ok := inp.(interface{ GetSelectedValue() string }); ok {
				val = valuer.GetSelectedValue()
			}
		}

		if err := inp.Validate(val); err != nil {
			if first == nil {
				first = err
			}
			if val != "" {
				f.errorSignals[i].Set(err.Error())
			}
			continue
		}
		f.errorSignals[i].Set("")
	}
	return first
}
