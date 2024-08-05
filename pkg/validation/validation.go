package validation

import (
	"fmt"
	"reflect"
)

type ValidationError interface {
	Error() string
	GetFieldName() string
	GetFieldValue() string
	GetViolations() []error
	// TODO: rename this
	GetHumanExplanations() []string
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var msgs string
	for _, err := range ve {
		if len(ve) == 1 {
			msgs = fmt.Sprintf("%v has the following errors: %v", err.GetFieldName(), err.Error())
		}

		msgs = fmt.Sprintf(
			"%v; %v has the following errors: %v",
			msgs,
			err.GetFieldName(),
			err.Error(),
		)
	}

	return msgs
}

// Unwrap() is used to make testing easier for now
func (ve ValidationErrors) Unwrap() []error {
	var errs []error
	for _, err := range ve {
		errs = append(errs, err.GetViolations()...)
	}

	return errs
}

//
// func (ve ValidationErrors) Error() string {
// 	var errMsg string
// 	if len(ve) == 1 {
// 		return ve[0].Field()
// 	}
//
// 	for _, err := range ve {
// 		errMsg += err.Field() + "; "
// 	}
//
// 	return errMsg
// }

//func (ve ValidationErrors) UnwrapMe() map[string][]error {
//	errs := make(map[string][]error, len(ve))
//	for _, errValidation := range ve {
//		errs[errValidation.Field()] = errValidation.causes()
//	}
//
//	return errs
//}

type Error struct {
	Value               string
	FieldName           string
	Violations          []error
	ViolationsForHumans []string
}

func (e Error) GetFieldName() string {
	return e.FieldName
}

func (e Error) GetFieldValue() string {
	return e.Value
}

func (e Error) GetViolations() []error {
	return e.Violations
}

func (e Error) GetHumanExplanations() []string {
	return e.ViolationsForHumans
}

func (e Error) Error() string {
	causes := fmt.Sprintf("field: %v has error(s): ", e.GetFieldName())
	for i, violation := range e.Violations {
		if i == 0 {
			causes = violation.Error()
		}

		if i != 0 {
			causes = causes + "; " + violation.Error()
		}
	}

	return causes
}

func getFieldValue(fieldValue reflect.Value) any {
	if !fieldValue.CanInterface() {
		return nil
	}
	return fieldValue.Interface()
}

func ValidateStruct(structToValidate any, validationMap map[string][]Rule) error {
	val := reflect.ValueOf(structToValidate)
	typ := reflect.TypeOf(structToValidate)

	var errors ValidationErrors
	for i := 0; i < val.NumField(); i++ {
		value := val.Field(i)
		name := typ.Field(i).Name

		fieldVal := value.String()
		if fieldVal == "<interface {} Value>" {
			fieldVal = "TypeDefault"
		}

		errVal := Error{
			Value:     fieldVal,
			FieldName: name,
		}

		for _, rule := range validationMap[name] {
			if rule.IsViolated(getFieldValue(value)) {
				errVal.Violations = append(
					errVal.Violations,
					rule.Violation(),
				)
				errVal.ViolationsForHumans = append(
					errVal.ViolationsForHumans,
					rule.ViolationForHumans(fieldVal),
				)
			}
		}

		if len(errVal.Violations) > 0 {
			errors = append(errors, errVal)
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
