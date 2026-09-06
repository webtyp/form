package form

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
)

// ValidateData validates a Fielder instance using this form's input rules.
// Satisfies the updated crudp.DataValidator interface (with model.Fielder).
func (f *Form) ValidateData(action byte, data model.Fielder) error {
	values := model.ReadValues(data.Schema(), data.Pointers())
	for i, inp := range f.Inputs {
		idx := f.fieldIndices[i]
		if idx < 0 || idx >= len(values) {
			continue
		}
		if skipper, ok := inp.(interface{ GetSkipValidation() bool }); ok && skipper.GetSkipValidation() {
			continue
		}
		val := fmt.Convert(values[idx]).String()
		if err := inp.Validate(val); err != nil {
			return err
		}
	}
	return nil
}
