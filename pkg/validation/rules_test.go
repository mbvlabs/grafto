package validation_test

import (
	"fmt"
	"testing"

	"github.com/mbv-labs/grafto/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestPasswordMatchConfirmRule(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		inputStruct struct {
			Password string
		}
		validations   map[string][]validation.Rule
		expectedError error
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
			expectedError: fmt.Errorf(
				validation.BaseErrMsg,
				"Password",
				"password",
				validation.ErrPasswordDontMatchConfirm,
			),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualErrors := validation.ValidateStruct(test.inputStruct, test.validations)

			if test.expectedError == nil {
				assert.Equal(t, test.expectedError, actualErrors,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			} else {
				assert.EqualError(
					t,
					test.expectedError,
					actualErrors.Error(),
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
		expectedError error
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
			expectedError: fmt.Errorf(
				validation.BaseErrMsg,
				"Data",
				"TypeDefault",
				validation.ErrIsRequired,
			),
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
			expectedError: fmt.Errorf(
				validation.BaseErrMsg,
				"Data",
				"TypeDefault",
				validation.ErrIsRequired,
			),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualErrors := validation.ValidateStruct(test.inputStruct, test.validations)

			if test.expectedError == nil {
				assert.Equal(t, test.expectedError, actualErrors,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			} else {
				assert.EqualError(
					t,
					actualErrors,
					test.expectedError.Error(),
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
		expectedError error
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
			expectedError: fmt.Errorf(
				validation.BaseErrMsg,
				"Data",
				"short string",
				validation.ErrValueTooShort,
			),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualErrors := validation.ValidateStruct(test.inputStruct, test.validations)

			if test.expectedError == nil {
				assert.Equal(t, test.expectedError, actualErrors,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			} else {
				assert.EqualError(
					t,
					test.expectedError,
					actualErrors.Error(),
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
		expectedError error
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
			expectedError: fmt.Errorf(
				validation.BaseErrMsg,
				"Data",
				"short string",
				validation.ErrValueTooLong,
			),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualErrors := validation.ValidateStruct(test.inputStruct, test.validations)

			if test.expectedError == nil {
				assert.Equal(t, test.expectedError, actualErrors,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			} else {
				assert.EqualError(
					t,
					test.expectedError,
					actualErrors.Error(),
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
		expectedError error
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
			expectedError: fmt.Errorf(
				validation.BaseErrMsg,
				"Data",
				"testgmail.com",
				validation.ErrInvalidEmail,
			),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualErrors := validation.ValidateStruct(test.inputStruct, test.validations)

			if test.expectedError == nil {
				assert.Equal(t, test.expectedError, actualErrors,
					fmt.Sprintf(
						"test failed, expected: %v but got: %v",
						test.expectedError,
						actualErrors,
					),
				)
			} else {
				assert.EqualError(
					t,
					test.expectedError,
					actualErrors.Error(),
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
