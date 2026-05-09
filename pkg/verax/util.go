// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"database/sql/driver"
	"reflect"
	"strconv"

	"github.com/ctx42/xrr/pkg/xrr"
)

var bytesType = reflect.TypeFor[[]byte]()

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
	return "", NewError("must be either a string or byte slice", ECInvType)
}

// StringOrBytes typecasts a value into a string or byte slice.
// Boolean flags are returned to indicate if the typecasting succeeds or not.
func StringOrBytes(value any) (
	isString bool,
	str string,
	isBytes bool,
	bs []byte,
) {

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
	switch v.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return v.Len(), nil

	default:
		format := "cannot get the length of %T"
		return 0, NewErrorf(format, value, xrr.WithCode(ECInvType))
	}
}

// checkNilAndEmpty checks value is nil or empty.
func checkNilAndEmpty(v any) (isNil, isEmpty bool) {
	isNil = IsNil(v)
	if isNil {
		return true, true
	}
	isEmpty = isEmptyValue(v)
	return isNil, isEmpty
}

// IsEmpty checks if a value is empty.
//
// A value is considered empty if:
//   - integer, float: zero
//   - bool: never empty
//   - string, array, slice, map, chan: nil or len() == 0
//   - interface, pointer: nil or the referenced value is empty
//   - other: IsZero() == true
//
// If the value implements [driver.Valuer], it returns the result of calling
// its Value method. If the input is nil, it returns true.
func IsEmpty(v any) bool {
	if isNil := IsNil(v); isNil {
		return true
	}
	return isEmptyValue(v)
}
func isEmptyValue(v any) bool {
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
	case reflect.Bool:
		return false

	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return val.Len() == 0

	case reflect.Chan:
		return val.Len() == 0

	case reflect.Interface, reflect.Pointer:
		if val.IsNil() {
			return true
		}
		return IsEmpty(val.Elem().Interface())

	default:
		return val.IsZero()
	}
}

// IsNil checks whether the provided value is actual nil or wrapped nil.
// Actual nil means the interface itself has no type or value (have == nil).
//
//   - isNil: true if the interface is actual nil.
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	val := reflect.ValueOf(v)
	kind := val.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice {
		return val.IsNil()
	}
	return false
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
	if isNil := IsNil(v); isNil {
		return nil
	}

	if val, ok := v.(driver.Valuer); ok {
		if vr, err := val.Value(); err == nil {
			return vr
		}
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		return Indirect(rv.Elem().Interface())
	default:
		return v
	}
}

// getInterface returns interface for given reflection value.
func getInterface(value reflect.Value) any {
	//goland:noinspection GoSwitchMissingCasesForIotaConsts
	switch value.Kind() { // nolint: exhaustive
	case reflect.Pointer, reflect.Interface:
		if value.IsNil() {
			return nil
		}
	}
	return value.Interface()
}

// mapErrKey returns a string to use as a key for a map error.
func mapErrKey(value reflect.Value) string {
	switch value.Kind() {
	case reflect.Pointer, reflect.Interface:
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
