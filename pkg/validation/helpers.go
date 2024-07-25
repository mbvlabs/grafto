package validation

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

func isEmpty(value interface{}) bool {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		if v.Type().String() == "uuid.UUID" {
			uid, ok := v.Interface().(uuid.UUID)
			if !ok {
				return true
			}

			return uid == uuid.UUID{}
		}
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Invalid:
		return true
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return isEmpty(v.Elem().Interface())
	case reflect.Struct:
		v, ok := value.(time.Time)
		if ok && v.IsZero() {
			return true
		}
	}

	return false
}

func lengthOfValue(v *reflect.Value) (int, error) {
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return v.Len(), nil
	case reflect.Invalid:
		return 0, fmt.Errorf("cannot get the length of invalid kind")
	default:
		return 0, fmt.Errorf("provided value: '%v' did not match any kind", v.Kind())
	}
}

func toString(value interface{}) (string, error) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	}

	return "", fmt.Errorf("cannot convert %v to string", v.Kind())
}
