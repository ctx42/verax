// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import "github.com/ctx42/xrr/pkg/xrr"

// NotNil rule error codes.
const (
	// ECReqNotNil represents error code for nil value.
	ECReqNotNil = "ECReqNotNil"
)

// ErrReqNotNil is the error returned when a value is nil.
var ErrReqNotNil = xrr.New("is required", ECReqNotNil)

// NotNil is a validation rule that checks if a value is not nil.
// NotNil only handles types including interface, pointer, slice, and map.
// All other types are considered valid.
var NotNil = notNilRule{condition: true, err: ErrReqNotNil}

// Compile time checks.
var (
	_ Customizer[notNilRule]  = notNilRule{}
	_ Conditioner[notNilRule] = notNilRule{}
)

type notNilRule struct {
	condition bool  // Run validation only when true.
	err       error // Validation error.
}

// Validate checks if the given value is valid or not.
func (r notNilRule) Validate(v any) error {
	if !r.condition {
		return nil
	}
	if isNil, _ := IsNil(v); isNil {
		return r.err
	}
	return nil
}

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (r notNilRule) When(condition bool) notNilRule {
	r.condition = condition
	return r
}

// Code sets custom error code for the rule.
func (r notNilRule) Code(code string) notNilRule {
	r.err = setCode(r.err, code)
	return r
}

// Error sets custom error for the rule.
func (r notNilRule) Error(err error) notNilRule {
	r.err = err
	return r
}
