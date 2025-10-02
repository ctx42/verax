// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ctx42/mirror/pkg/mirror"
	"github.com/ctx42/xrr/pkg/xrr"
)

// Struct validation errors.
var (
	// ErrNotStructPtr is the error that a struct being validated is not
	// specified as a pointer.
	ErrNotStructPtr = xrr.New(
		"only a pointer to a struct can be validated",
		ECInternal,
	)
)

// ErrFieldNotFound is the error that a field cannot be found in the struct.
type ErrFieldNotFound int

// Error returns the error string of ErrFieldNotFound.
func (e ErrFieldNotFound) Error() string {
	return fmt.Sprintf("the field #%v cannot be found in the struct", int(e))
}

// ErrorCode always returns ECInternal error code.
func (e ErrFieldNotFound) ErrorCode() string { return ECInternal }

// ErrFieldPointer is the error that a field is not specified as a pointer.
type ErrFieldPointer int

// Error returns the error string of ErrFieldPointer.
func (e ErrFieldPointer) Error() string {
	return fmt.Sprintf("field #%v must be specified as a pointer", int(e))
}

// ErrorCode always returns ECInternal error code.
func (e ErrFieldPointer) ErrorCode() string { return ECInternal }

// FieldRules represents a rule set associated with a struct field.
type FieldRules struct {
	fieldPtr any
	tag      string
	rules    []Rule
}

// ValidateStruct validates a struct by checking the specified struct fields
// against the corresponding validation rules. Note that the struct being
// validated must be specified as a pointer to it. If the pointer is nil, it is
// considered valid. Use Field() to specify struct fields that need to be
// validated. Each Field() call specifies a single field which should be
// specified as a pointer to the field. A field can be associated with multiple
// rules.
//
// For example,
//
//	value := struct {
//	    Name  string
//	    Value string
//	}{"name", "demo"}
//	err := validation.ValidateStruct(&value,
//	    vrule.Field(&a.Name, vrule.Required),
//	    vrule.Field(&a.Value, vrule.Required, vrule.Length(5, 10)),
//	)
//	fmt.Println(err)
//	// Value: the length must be between 5 and 10.
//
// Returns error with ECInternal code on unexpected errors, otherwise it
// returns xrr.Fields error.
//
// nolint: cyclop
func ValidateStruct(v any, fields ...*FieldRules) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || !val.IsNil() &&
		val.Elem().Kind() != reflect.Struct {

		return ErrNotStructPtr // Must be a pointer to a struct.
	}
	if val.IsNil() {
		return nil // Treat a nil struct pointer as valid.
	}
	val = val.Elem()

	var ers xrr.Fields
	for i, fr := range fields {
		fv := reflect.ValueOf(fr.fieldPtr)
		if fv.Kind() != reflect.Ptr {
			return ErrFieldPointer(i)
		}

		sf := findStructField(val, fv)
		if sf == nil {
			return ErrFieldNotFound(i)
		}

		v = fv.Elem().Interface()
		if err := Validate(v, fr.rules...); err != nil {
			if xrr.GetCode(err) == ECInternal {
				msg := fmt.Sprintf("%s: %s", getErrorFieldName(fr.tag, sf), err)
				return xrr.New(msg, ECInternal)
			}
			if ers == nil {
				ers = xrr.Fields{}
			}
			if sf.Anonymous {
				// Merge errors from the anonymous struct field.
				if es, ok := err.(xrr.Fielder); ok { // nolint: errorlint
					for name, value := range es.ErrorFields() {
						ers[name] = value
					}
					continue
				}
			}
			ers[getErrorFieldName(fr.tag, sf)] = err
		}
	}
	return ers.Filter()
}

// Field specifies a struct field and the corresponding validation rules.
// The struct field must be specified as a pointer to it.
func Field(fieldPtr any, rules ...Rule) *FieldRules {
	return &FieldRules{
		fieldPtr: fieldPtr,
		rules:    rules,
	}
}

// Tag sets a tag to use for the error field name.
func (fr *FieldRules) Tag(tag string) *FieldRules {
	fr.tag = tag
	return fr
}

// findStructField looks for a field in the given struct.
// The field being looked for should be a pointer to the actual struct field.
// If found, the field info will be returned. Otherwise, nil will be returned.
func findStructField(s, f reflect.Value) *reflect.StructField {
	ptr := f.Pointer()
	for i := s.NumField() - 1; i >= 0; i-- {
		sf := mirror.ReflectType(s.Type()).FieldByIndex(i)
		if ptr == s.Field(i).UnsafeAddr() {
			// Do additional type comparison because it's possible that
			// the address of an embedded struct is the same as the first
			// field of the embedded struct.
			if sf.Type() == f.Elem().Type() {
				sf := sf.StructField()
				return &sf
			}
		}
		if sf.IsAnonymous() {
			// Dive into the anonymous struct to look for the field.
			fi := s.Field(i)
			if sf.Kind() == reflect.Ptr {
				fi = fi.Elem()
			}
			if fi.Kind() == reflect.Struct {
				if f := findStructField(fi, f); f != nil {
					return f
				}
			}
		}
	}
	return nil
}

// getErrorFieldName returns the name that should be used to represent the
// validation error of a struct field.
func getErrorFieldName(fieldTag string, f *reflect.StructField) string {
	if fieldTag == "" {
		fieldTag = ErrorTag
	}
	if tag := f.Tag.Get(fieldTag); tag != "" && tag != "-" {
		if cps := strings.SplitN(tag, ",", 2); cps[0] != "" {
			return cps[0]
		}
	}
	return f.Name
}
