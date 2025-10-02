// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import "github.com/ctx42/xrr/pkg/xrr"

// Absence error codes.
const (
	// ECReqNil represents error code for not nil value.
	ECReqNil = "ECReqNil"

	// ECReqEmpty represents error code for not empty value.
	ECReqEmpty = "ECReqEmpty"
)

// Absence errors.
var (
	// ErrReqNil is the error returned when a value is not nil.
	ErrReqNil = xrr.New("must be blank", ECReqNil)

	// ErrReqEmpty is returned when a not nil value is not empty.
	ErrReqEmpty = xrr.New("must be blank", ECReqEmpty)
)

// Absence rules.
var (
	// Nil checks if a value is nil.
	Nil = absentRule{condition: true, skipNil: false, err: ErrReqNil}

	// Empty checks if a not nil value is empty.
	Empty = absentRule{condition: true, skipNil: true, err: ErrReqEmpty}
)

// Compile time checks.
var (
	_ Customizer[absentRule]  = absentRule{}
	_ Conditioner[absentRule] = absentRule{}
)

type absentRule struct {
	condition bool  // Run validation only when true.
	skipNil   bool  // When true, the nil value is considered valid.
	err       error // Validation error.
}

// Validate checks if the given value is valid or not.
func (r absentRule) Validate(v any) error {
	if !r.condition {
		return nil
	}
	isNil, _ := IsNil(v)
	if !isNil && (!r.skipNil || !IsEmpty(v)) {
		return r.err
	}
	return nil
}

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (r absentRule) When(condition bool) absentRule {
	r.condition = condition
	return r
}

// Code sets the error code for the rule.
func (r absentRule) Code(code string) absentRule {
	r.err = setCode(r.err, code)
	return r
}

// Error sets custom error for the rule.
func (r absentRule) Error(err error) absentRule {
	r.err = err
	return r
}
