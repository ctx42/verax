// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"fmt"
	"reflect"

	"github.com/ctx42/xrr/pkg/xrr"
)

// Equality error codes.
const (
	// ECNotEqual represents error code for values which must not be equal
	// to some other field.
	ECNotEqual = "ECNotEqual"

	// ECEqual represents error code for values which must be equal
	// to some other field.
	ECEqual = "ECEqual"
)

// ErrNotEqual is generic error for not equal values.
var ErrNotEqual = xrr.New("not equal", ECNotEqual)

// Equal constructs rule checking a validated value is equal to "want".
func Equal(want any) EqualRule {
	return EqualRule{
		want:      want,
		condition: true,
		compare:   reflect.DeepEqual,
		err:       equalToError(want, ECNotEqual),
	}
}

// NotEqual constructs rule checking a validated value is not equal to "want".
func NotEqual(want any) EqualRule {
	return EqualRule{
		want:      want,
		condition: true,
		compare:   notEqual,
		err:       notEqualToError(want, ECEqual),
	}
}

// EqualField constructs rule checking a validated value is equal to "want".
// When it isn't, the error message will say the value must be equal to "field".
func EqualField(want any, field string) EqualRule {
	r := Equal(want)
	msg := fmt.Sprintf("must be equal to '%s'", field)
	r.err = xrr.New(msg, ECNotEqual)
	return r
}

// NotEqualField constructs rule checking a validated value is not equal to
// "want". When it is the error message will say the value must not be equal to
// "field".
func NotEqualField(want any, field string) EqualRule {
	r := NotEqual(want)
	msg := fmt.Sprintf("must not be equal to '%s'", field)
	r.err = xrr.New(msg, ECEqual)
	return r
}

// EqualBy constructs rule checking a validated value is equal to "want" using
// the given comparison function.
func EqualBy(want any, fn func(want, have any) bool) EqualRule {
	return EqualRule{
		want:      want,
		condition: true,
		compare:   fn,
		err:       equalToError(want, ECEqual),
	}
}

// notEqual returns true when are not equal.
func notEqual(want, have any) bool { return !reflect.DeepEqual(want, have) }

// Compile time checks.
var (
	_ Customizer[EqualRule]  = EqualRule{}
	_ Conditioner[EqualRule] = EqualRule{}
)

// EqualRule is a rule that checks a value matches the expected value.
// The [reflect.DeepEqual] is used to make comparisons.
type EqualRule struct {
	want      any                 // Wanted value.
	condition bool                // Run validation only when true.
	compare   func(x, y any) bool // Comparison function.
	err       error               // Validation error.
}

// Validate checks if the given value is valid or not.
func (r EqualRule) Validate(v any) error {
	if !r.condition {
		return nil
	}
	if !r.compare(r.want, v) {
		return r.err
	}
	return nil
}

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (r EqualRule) When(condition bool) EqualRule {
	r.condition = condition
	return r
}

// Code sets the error code for the rule.
func (r EqualRule) Code(code string) EqualRule {
	r.err = setCode(r.err, code)
	return r
}

// Error sets custom error for the rule.
func (r EqualRule) Error(err error) EqualRule {
	r.err = err
	return r
}

// equalToError is a helper function generating must be equal to v error.
func equalToError(v any, code string) error {
	msg := fmt.Sprintf("must be equal to '%v'", format(v))
	return xrr.New(msg, code)
}

// notEqualToError is a helper function generating must not be equal to v error.
func notEqualToError(v any, code string) error {
	msg := fmt.Sprintf("must not be equal to '%v'", format(v))
	return xrr.New(msg, code)
}
