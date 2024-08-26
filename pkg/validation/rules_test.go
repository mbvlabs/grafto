package validation_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mbvlabs/grafto/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestPasswordMatchConfirmRule(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		inputStruct struct {
			Password string
		}
		validations   map[string][]validation.Rule
		expectedError validation.ValidationErrors
	}{
		"should return no errors because password match": {
			inputStruct: struct{ Password string }{
				Password: "password",
			},
			validations: map[string][]validation.Rule{
				"Password": {
					validation.PasswordMatchConfirmRule("password"),
				},
			},
			expectedError: nil,
		},
		"should return an error because password don't match": {
			inputStruct: struct{ Password string }{
				Password: "password",
			},
			validations: map[string][]validation.Rule{
				"Password": {
					validation.PasswordMatchConfirmRule("PASSWORD"),
				},
			},
			expectedError: []validation.ValidationError{
				validation.Error{
					Value:               "password",
					FieldName:           "Password",
					Violations:          []error{validation.ErrPasswordDontMatchConfirm},
					ViolationsForHumans: []string{"password and confirm password must match"},
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualErrors := validation.ValidateStruct(test.inputStruct, test.validations)

			if test.expectedError == nil {
				assert.Equal(t, nil, actualErrors,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			}

			if test.expectedError != nil {
				var valiErrs validation.ValidationErrors
				if ok := errors.As(actualErrors, &valiErrs); !ok {
					t.Fail()
				}

				assert.EqualValues(
					t,
					test.expectedError,
					valiErrs,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			}
		})
	}
}

func TestRequiredRule(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		inputStruct struct {
			Data any
		}
		validations   map[string][]validation.Rule
		expectedError validation.ValidationErrors
	}{
		"should return no errors because required string field is provided": {
			inputStruct: struct{ Data any }{
				Data: "some information here",
			},
			validations: map[string][]validation.Rule{
				"Data": {
					validation.RequiredRule,
				},
			},
			expectedError: nil,
		},
		"should return an error because string field is not provided": {
			inputStruct: struct{ Data any }{
				Data: "",
			},
			validations: map[string][]validation.Rule{
				"Data": {
					validation.RequiredRule,
				},
			},
			expectedError: []validation.ValidationError{
				validation.Error{
					Value:               "TypeDefault",
					FieldName:           "Data",
					Violations:          []error{validation.ErrIsRequired},
					ViolationsForHumans: []string{"must be provided"},
				},
			},
		},
		"should return no errors because required int field is provided": {
			inputStruct: struct{ Data any }{
				Data: 1,
			},
			validations: map[string][]validation.Rule{
				"Data": {
					validation.RequiredRule,
				},
			},
			expectedError: nil,
		},
		"should return an error because int field is not provided": {
			inputStruct: struct{ Data any }{
				Data: 0,
			},
			validations: map[string][]validation.Rule{
				"Data": {
					validation.RequiredRule,
				},
			},
			expectedError: []validation.ValidationError{
				validation.Error{
					Value:               "TypeDefault",
					FieldName:           "Data",
					Violations:          []error{validation.ErrIsRequired},
					ViolationsForHumans: []string{"must be provided"},
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualErrors := validation.ValidateStruct(test.inputStruct, test.validations)

			if test.expectedError == nil {
				assert.Equal(t, nil, actualErrors,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			}

			if test.expectedError != nil {
				var valiErrs validation.ValidationErrors
				if ok := errors.As(actualErrors, &valiErrs); !ok {
					t.Fail()
				}

				assert.EqualValues(
					t,
					test.expectedError,
					valiErrs,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			}
		})
	}
}

func TestMinLengthRule(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		inputStruct struct {
			Data string
		}
		validations   map[string][]validation.Rule
		expectedError validation.ValidationErrors
	}{
		"should return no errors because value was longer than the minimum": {
			inputStruct: struct{ Data string }{
				Data: "suuuuuper long string",
			},
			validations: map[string][]validation.Rule{
				"Data": {
					validation.MinLengthRule(5),
				},
			},
			expectedError: nil,
		},
		"should return an error because value was shorter than the minimum": {
			inputStruct: struct{ Data string }{
				Data: "short string",
			},
			validations: map[string][]validation.Rule{
				"Data": {
					validation.MinLengthRule(50000),
				},
			},
			expectedError: []validation.ValidationError{
				validation.Error{
					Value:      "short string",
					FieldName:  "Data",
					Violations: []error{validation.ErrValueTooShort},
					ViolationsForHumans: []string{
						"needs to be longer than: '50000' characters",
					},
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualErrors := validation.ValidateStruct(test.inputStruct, test.validations)

			if test.expectedError == nil {
				assert.Equal(t, nil, actualErrors,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			}

			if test.expectedError != nil {
				var valiErrs validation.ValidationErrors
				if ok := errors.As(actualErrors, &valiErrs); !ok {
					t.Fail()
				}

				assert.EqualValues(
					t,
					test.expectedError,
					valiErrs,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			}
		})
	}
}

func TestMaxLengthRule(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		inputStruct struct {
			Data string
		}
		validations   map[string][]validation.Rule
		expectedError validation.ValidationErrors
	}{
		"should return no errors because value was shorter than the maximum": {
			inputStruct: struct{ Data string }{
				Data: "suuuuuper long string",
			},
			validations: map[string][]validation.Rule{
				"Data": {
					validation.MaxLengthRule(500),
				},
			},
			expectedError: nil,
		},
		"should return an error because value was longer than the maxium": {
			inputStruct: struct{ Data string }{
				Data: "short string",
			},
			validations: map[string][]validation.Rule{
				"Data": {
					validation.MaxLengthRule(5),
				},
			},
			expectedError: []validation.ValidationError{
				validation.Error{
					Value:      "short string",
					FieldName:  "Data",
					Violations: []error{validation.ErrValueTooLong},
					ViolationsForHumans: []string{
						"can max be: '5' characters",
					},
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualErrors := validation.ValidateStruct(test.inputStruct, test.validations)

			if test.expectedError == nil {
				assert.Equal(t, nil, actualErrors,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			}

			if test.expectedError != nil {
				var valiErrs validation.ValidationErrors
				if ok := errors.As(actualErrors, &valiErrs); !ok {
					t.Fail()
				}

				assert.EqualValues(
					t,
					test.expectedError,
					valiErrs,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			}
		})
	}
}

func TestValidEmailRule(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		inputStruct struct {
			Data string
		}
		validations   map[string][]validation.Rule
		expectedError validation.ValidationErrors
	}{
		"should return no errors because valid email was provided": {
			inputStruct: struct{ Data string }{
				Data: "test@gmail.com",
			},
			validations: map[string][]validation.Rule{
				"Data": {
					validation.ValidEmailRule,
				},
			},
			expectedError: nil,
		},
		"should return an error because invalid email was provided": {
			inputStruct: struct{ Data string }{
				Data: "testgmail.com",
			},
			validations: map[string][]validation.Rule{
				"Data": {
					validation.ValidEmailRule,
				},
			},
			expectedError: []validation.ValidationError{
				validation.Error{
					Value:      "testgmail.com",
					FieldName:  "Data",
					Violations: []error{validation.ErrInvalidEmail},
					ViolationsForHumans: []string{
						"the provided email: 'testgmail.com' is not valid",
					},
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualErrors := validation.ValidateStruct(test.inputStruct, test.validations)

			if test.expectedError == nil {
				assert.Equal(t, nil, actualErrors,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			}

			if test.expectedError != nil {
				var valiErrs validation.ValidationErrors
				if ok := errors.As(actualErrors, &valiErrs); !ok {
					t.Fail()
				}

				assert.EqualValues(
					t,
					test.expectedError,
					valiErrs,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			}
		})
	}
}
