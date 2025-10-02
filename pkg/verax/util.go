// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"text/template"
	"time"

	"github.com/ctx42/xrr/pkg/xrr"
)

var bytesType = reflect.TypeOf([]byte(nil))

// EnsureString ensures the given value is a string.
// If the value is a byte slice, it will be typecast into a string.
// An error is returned otherwise.
func EnsureString(value any) (string, error) {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.String {
		return v.String(), nil
	}
	if v.Type() == bytesType {
		return string(v.Interface().([]byte)), nil // nolint: forcetypeassert
	}
	return "", xrr.New("must be either a string or byte slice", ECInvType)
}

// StringOrBytes typecasts a value into a string or byte slice.
// Boolean flags are returned to indicate if the typecasting succeeds or not.
func StringOrBytes(value any) (isString bool, str string, isBytes bool, bs []byte) {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.String {
		str = v.String()
		isString = true
	} else if v.Kind() == reflect.Slice && v.Type() == bytesType {
		bs = v.Interface().([]byte) // nolint: forcetypeassert
		isBytes = true
	}
	return
}

// LengthOfValue returns the length of a value that is a string, slice, map,
// or array. An error is returned for all other types.
func LengthOfValue(value any) (int, error) {
	v := reflect.ValueOf(value)
	switch v.Kind() { // nolint: exhaustive
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return v.Len(), nil

	default:
		msg := fmt.Sprintf("cannot get the length of %T", value)
		return 0, xrr.New(msg, ECInvType)
	}
}

// ToInt converts the given value to an int64.
// An error is returned for all incompatible types.
func ToInt(value any) (int64, error) {
	v := reflect.ValueOf(value)
	switch v.Kind() { // nolint: exhaustive
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil

	default:
		msg := fmt.Sprintf("cannot convert %T to int64", value)
		return 0, xrr.New(msg, ECInvType)
	}
}

// ToUint converts the given value to an uint64.
// An error is returned for all incompatible types.
func ToUint(value any) (uint64, error) {
	v := reflect.ValueOf(value)
	switch v.Kind() { // nolint: exhaustive
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return v.Uint(), nil

	case reflect.Uint64, reflect.Uintptr:
		return v.Uint(), nil

	default:
		msg := fmt.Sprintf("cannot convert %T to uint64", value)
		return 0, xrr.New(msg, ECInvType)
	}
}

// ToFloat converts the given value to a float64.
// An error is returned for all incompatible types.
func ToFloat(value any) (float64, error) {
	v := reflect.ValueOf(value)
	switch v.Kind() { // nolint: exhaustive
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil

	default:
		msg := fmt.Sprintf("cannot convert %T to float64", value)
		return 0, xrr.New(msg, ECInvType)
	}
}

// IsEmpty checks if a value is empty.
//
// A value is considered empty if:
//   - integer, float: zero
//   - bool: false
//   - string, array, slice, map: nil or len() == 0
//   - interface, pointer: nil or the referenced value is empty
//   - struct: all fields are empty
//   - other: IsZero() == true
//
// If the value implements [driver.Valuer], it returns the result of calling
// its Value method. If the input is nil, it turns true.
func IsEmpty(v any) bool {
	if isNil, _ := IsNil(v); isNil {
		return true
	}

	if z, ok := v.(interface{ IsZero() bool }); ok {
		return z.IsZero()
	}

	if val, ok := v.(driver.Valuer); ok {
		if vr, err := val.Value(); err == nil {
			return IsEmpty(vr)
		}
	}

	val := reflect.ValueOf(v)
	switch knd := val.Kind(); knd {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return val.Len() == 0

	case reflect.Interface, reflect.Ptr:
		if val.IsNil() {
			return true
		}
		return IsEmpty(val.Elem().Interface())

	case reflect.Struct:
		return val.Type().NumField() == 0

	default:
		return val.IsZero()
	}
}

// IsNil checks whether the provided value is actual nil or wrapped nil.
// Actual nil means the interface itself has no type or value (have == nil).
// Wrapped nil means the interface holds a nil value of a concrete type (e.g.,
// a nil pointer or slice). It returns two booleans:
//
//   - isNil: true if the interface is actual nil.
//   - isWrapped: true if the interface holds a nil value of a type.
func IsNil(v any) (isNil, isWrapped bool) {
	if v == nil {
		return true, false
	}
	val := reflect.ValueOf(v)
	kind := val.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice {
		return val.IsNil(), true
	}
	return false, false
}

// Indirect dereferences the given value if it is a pointer or interface,
// returning the underlying value. If the value implements [driver.Valuer], it
// returns the result of calling its Value method. If the input is nil or not a
// pointer/interface, it is returned unchanged.
//
// Examples:
//
//   - For a *int pointing to 42, it returns 42.
//   - For an interface{} containing a *string, it returns the string.
//   - For a driver.Valuer, it returns the result of Value().
func Indirect(v any) any {
	if isNil, _ := IsNil(v); isNil {
		return nil
	}

	if val, ok := v.(driver.Valuer); ok {
		if vr, err := val.Value(); err == nil {
			return vr
		}
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		return Indirect(rv.Elem().Interface())
	default:
		return v
	}
}

// getInterface returns interface for given reflection value.
func getInterface(value reflect.Value) any {
	//goland:noinspection GoSwitchMissingCasesForIotaConsts
	switch value.Kind() { // nolint: exhaustive
	case reflect.Ptr, reflect.Interface:
		if value.IsNil() {
			return nil
		}
	}
	return value.Interface()
}

// mapErrKey returns a string to use as a key for a map error.
func mapErrKey(value reflect.Value) string {
	switch value.Kind() {
	case reflect.Ptr, reflect.Interface:
		if value.IsNil() {
			return ""
		}
		return value.Elem().String()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64:

		return strconv.FormatInt(value.Int(), 10)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:

		return strconv.FormatUint(value.Uint(), 10)

	default:
		return value.String()
	}
}

// emtpl returns parsed error message template. Panics on error.
func emtpl(tpl string) *template.Template {
	return template.Must(template.New("").Parse(tpl))
}

// format formats some value types in the more readable way.
//
// - time.Time: formatted as RFC3339.
// - other: returned as is.
func format(v any) any {
	if tim, ok := v.(time.Time); ok {
		return tim.Format(time.RFC3339)
	}
	return v
}
