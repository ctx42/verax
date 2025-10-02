// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"reflect"

	"github.com/ctx42/xrr/pkg/xrr"
)

// ECInvIn represents error code for invalid in rule.
const ECInvIn = "ECInvIn"

// ErrNotIn is the error that returns in case of an invalid value for "In" rule.
var ErrNotIn = xrr.New("must be in the list", ECInvIn)

// ErrIn is the error that returns in case of an invalid value for "NotIn" rule.
var ErrIn = xrr.New("must not be in the list", ECInvIn)

// In returns a validation rule that checks if a value can be found in the
// given list of values. Note that the value being checked and the possible
// range of values must be of the same type. The reflect.DeepEqual() will be
// used to determine if two values are equal. For more details please refer to
// https://golang.org/pkg/reflect/#DeepEqual. An empty value is considered
// valid. Use the Required rule to make sure a value is not empty.
func In(values ...any) InRule {
	return InRule{elements: values, condition: true, in: true, err: ErrNotIn}
}

// NotIn returns a validation rule that checks if a value cannot be found in
// the given list of values. Note that the value being checked and the possible
// range of values must be of the same type. The reflect.DeepEqual() will be
// used to determine if two values are equal. For more details please refer to
// https://golang.org/pkg/reflect/#DeepEqual. An empty value is considered
// valid. Use the Required rule to make sure a value is not empty.
func NotIn(values ...any) InRule {
	return InRule{elements: values, condition: true, in: false, err: ErrIn}
}

// Compile time checks.
var (
	_ Customizer[InRule]  = InRule{}
	_ Conditioner[InRule] = InRule{}
)

// InRule is a validation rule that validates if a value can be found in the
// given list of values.
type InRule struct {
	elements  []any // List of valid values.
	condition bool  // Run validation only when true.
	in        bool  // Value must (true) or must not (false) be on the list.
	err       error // Validation error.
}

// Validate checks if the given value is valid or not.
func (r InRule) Validate(v any) error {
	if !r.condition {
		return nil
	}
	if isNil, _ := IsNil(v); isNil {
		return nil
	}
	if IsEmpty(v) {
		return nil
	}
	val := Indirect(v)
	if r.in {
		return r.inRule(val)
	}
	return r.notInRule(val)
}

// inRule returns an error if v is not on the list of elements or its type
// doesn't match.
func (r InRule) inRule(v any) error {
	vt := reflect.TypeOf(v)
	for _, e := range r.elements {
		if vt != reflect.TypeOf(e) {
			return setCode(ErrInvType, xrr.GetCode(r.err))
		}
		if reflect.DeepEqual(e, v) {
			return nil
		}
	}
	return r.err
}

// notInRule returns an error if v is on the list of elements or its type
// doesn't match.
func (r InRule) notInRule(v any) error {
	vt := reflect.TypeOf(v)
	for _, e := range r.elements {
		if vt != reflect.TypeOf(e) {
			return setCode(ErrInvType, xrr.GetCode(r.err))
		}
		if reflect.DeepEqual(e, v) {
			return r.err
		}
	}
	return nil
}

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (r InRule) When(condition bool) InRule {
	r.condition = condition
	return r
}

// Code sets the error code for the rule.
func (r InRule) Code(code string) InRule {
	r.err = setCode(r.err, code)
	return r
}

// Error sets custom error for the rule.
func (r InRule) Error(err error) InRule {
	r.err = err
	return r
}
