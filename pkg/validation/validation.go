package validation

import (
	"errors"
	"fmt"
	"reflect"
)

var BaseErrMsg = "Field: '%s' with Value: '%v' has Error(s): validation failed due to '%v'"

type ValidationError interface {
	Error() string
	ErrorForHumans() string
	Field() string
	Value() string
	Causes() []error
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var errMsg string
	if len(ve) == 1 {
		return ve[0].Error()
	}

	for _, err := range ve {
		errMsg += err.Error() + "; "
	}
	return errMsg
}

func (ve ValidationErrors) UnwrapViolations() []error {
	var errs []error
	for _, errValidation := range ve {
		errs = append(errs, errValidation.Causes()...)
	}

	return errs
}

type Error struct {
	FieldValue         string
	FieldName          string
	Violations         []error
	ViolationsForHuman []error
}

func (e Error) Field() string {
	return e.FieldName
}

func (e Error) Value() string {
	return e.FieldValue
}

func (e Error) Causes() []error {
	return e.Violations
}

func (e Error) ErrorForHumans() string {
	var causes string
	for i, violation := range e.ViolationsForHuman {
		var validationErr ValidationError
		if errors.As(violation, &validationErr) {
			if i == 0 {
				causes = validationErr.ErrorForHumans()
			}

			if i != 0 {
				causes = causes + ", " + validationErr.ErrorForHumans()
			}
		}

	}

	return fmt.Sprintf(
		BaseErrMsg,
		e.FieldName,
		e.FieldValue,
		causes,
	)
}

func (e Error) Error() string {
	var causes string
	for i, violation := range e.Violations {
		if i == 0 {
			causes = violation.Error()
		}

		if i != 0 {
			causes = causes + ", " + violation.Error()
		}
	}

	return fmt.Sprintf(
		BaseErrMsg,
		e.FieldName,
		e.FieldValue,
		causes,
	)
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
			FieldValue: fieldVal,
			FieldName:  name,
		}

		for _, rule := range validationMap[name] {
			if rule.IsViolated(getFieldValue(value)) {
				errVal.Violations = append(
					errVal.Violations,
					rule.Violation(),
				)
				errVal.ViolationsForHuman = append(
					errVal.ViolationsForHuman,
					rule.ViolationForHumans(name),
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
