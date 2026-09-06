package form

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
)

// SyncValues copies all input values back into the bound struct
// via the Fielder's Pointers() method.
func (f *Form) SyncValues(data model.Fielder) error {
	pointers := data.Pointers()
	schema := data.Schema()

	for i, inp := range f.Inputs {
		idx := f.fieldIndices[i]
		if idx < 0 || idx >= len(pointers) {
			continue
		}

		// Signal is the source of truth in WASM mode.
		val := f.valueSignals[i].Get()

		// Fallback only if we are somehow in SSR mode where signals might be empty
		// but input state is populated (though f.Render() should handle this).
		if val == "" && f.ssrMode {
			if getter, ok := inp.(interface{ GetValues() []string }); ok {
				vals := getter.GetValues()
				if len(vals) > 0 {
					val = vals[0]
				}
			}
		}

		values := []string{val}

		ptr := pointers[idx]
		field := schema[idx]

		if val == "" {
			// Zero the field
			zeroField(ptr, field.Type.Storage())
			continue
		}

		writeField(ptr, field.Type.Storage(), values)
	}

	// A hidden PK (New skipped it — see that function's comment) has no
	// Input and so is untouched by the loop above. An EXISTING record
	// already carries its real id here (Presenter.Select loaded it before
	// the form ever rendered); only a brand-new record's id is still the
	// zero value at this point, and it must not reach the backend that
	// way — model.ValidateFields (deeper in orm.Create) rejects an empty
	// PK outright. Only text PKs are assigned here: an int PK is normally
	// auto-increment, the DB's job (see orm.Create), not this form's.
	for _, idx := range f.hiddenPKIndices {
		field := schema[idx]
		if field.Type.Storage() != model.FieldText {
			continue
		}
		ptr := pointers[idx]
		if val, _ := model.ReadStringPtr(ptr); val != "" {
			continue
		}
		writeField(ptr, model.FieldText, []string{f.idGen.NewID()})
	}

	return nil
}

// zeroField sets a field to its zero value via its pointer.
func zeroField(ptr any, ft model.FieldType) {
	switch ft {
	case model.FieldText:
		if p, ok := ptr.(*string); ok {
			*p = ""
		}
	case model.FieldInt:
		if p, ok := ptr.(*int64); ok {
			*p = 0
		}
	case model.FieldFloat:
		if p, ok := ptr.(*float64); ok {
			*p = 0
		}
	case model.FieldBool:
		if p, ok := ptr.(*bool); ok {
			*p = false
		}
	}
}

// writeField writes string values into a field via its pointer.
func writeField(ptr any, ft model.FieldType, values []string) {
	switch ft {
	case model.FieldText:
		if p, ok := ptr.(*string); ok {
			*p = values[0]
		}
	case model.FieldInt:
		if p, ok := ptr.(*int64); ok {
			val, _ := fmt.Convert(values[0]).Int64()
			*p = val
		}
	case model.FieldFloat:
		if p, ok := ptr.(*float64); ok {
			val, _ := fmt.Convert(values[0]).Float64()
			*p = val
		}
	case model.FieldBool:
		if p, ok := ptr.(*bool); ok {
			val, _ := fmt.Convert(values[0]).Bool()
			*p = val
		}
	}
}
