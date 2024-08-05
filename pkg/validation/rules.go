package validation

import (
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
)

type Rule interface {
	IsViolated(val any) bool
	Violation() error
	// TODO: should maybe be named HumanExplanation/HumanDescription/Description?
	ViolationForHumans(val string) string
}

func PasswordMatchConfirmRule(confirm string) passwordMatchConfirm {
	return passwordMatchConfirm{
		confirm,
	}
}

type passwordMatchConfirm struct {
	confirm string
}

// IsViolated implements Rule.
func (p passwordMatchConfirm) IsViolated(val any) bool {
	valString := fmt.Sprintf("%v", val)

	return valString != p.confirm
}

// Violation implements Rule.
func (p passwordMatchConfirm) Violation() error {
	return ErrPasswordDontMatchConfirm
}

// ViolationForHumans implements Rule.
func (p passwordMatchConfirm) ViolationForHumans(val string) string {
	return "password and confirm password must match"
}

var RequiredRule required

type required struct{}

// ViolationForHumans implements Rule.
func (r required) ViolationForHumans(val string) string {
	return "must be provided"
}

// IsViolated implements Rule.
func (r required) IsViolated(v any) bool {
	return isEmpty(v)
}

// Violation implements Rule.
func (r required) Violation() error {
	return ErrIsRequired
}

func MinLengthRule(length int) minLength {
	return minLength{length}
}

type minLength struct {
	minimum int
}

// IsViolated implements Rule.
func (m minLength) IsViolated(val any) bool {
	v := reflect.ValueOf(val)
	valLen, err := lengthOfValue(&v)
	if err != nil {
		slog.Error("could not get length of value for MinLenRule", "error", err, "val", val)
		return true
	}

	return valLen < m.minimum
}

// Violation implements Rule.
func (m minLength) Violation() error {
	return ErrValueTooShort
}

// ViolationForHumans implements Rule.
func (m minLength) ViolationForHumans(val string) string {
	return fmt.Sprintf(
		"needs to be longer than: '%v' characters",
		m.minimum,
	)
}

func MaxLengthRule(length int) maxLength {
	return maxLength{length}
}

type maxLength struct {
	maximum int
}

// IsViolated implements Rule.
func (m maxLength) IsViolated(val any) bool {
	v := reflect.ValueOf(val)
	valLen, err := lengthOfValue(&v)
	if err != nil {
		slog.Error("could not get length of value for MaxLenRule", "error", err, "val", val)
		return true
	}

	return valLen > m.maximum
}

// Violation implements Rule.
func (m maxLength) Violation() error {
	return ErrValueTooLong
}

// ViolationForHumans implements Rule.
func (m maxLength) ViolationForHumans(val string) string {
	return fmt.Sprintf(
		"can max be: '%v' characters",
		m.maximum,
	)
}

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

func isEmailValid(e string) bool {
	return emailRegex.MatchString(e)
}

var ValidEmailRule validEmail

type validEmail struct{}

// IsViolated implements Rule.
func (v validEmail) IsViolated(val any) bool {
	stringVal, err := toString(val)
	if err != nil {
		return true
	}

	return !isEmailValid(stringVal)
}

// Violation implements Rule.
func (v validEmail) Violation() error {
	return ErrInvalidEmail
}

// ViolationForHumans implements Rule.
func (v validEmail) ViolationForHumans(val string) string {
	return fmt.Sprintf("the provided email: '%v' is not valid", val)
}

var (
	_ Rule = new(passwordMatchConfirm)
	_ Rule = new(required)
	_ Rule = new(minLength)
	_ Rule = new(maxLength)
	_ Rule = new(validEmail)
)
