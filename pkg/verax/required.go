// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import "github.com/ctx42/xrr/pkg/xrr"

// Required rule error codes.
const (
	// ECRequired represents error code for missing required value.
	ECRequired = "ECRequired"

	// ECReqNotEmpty represents error code for not nil but empty value.
	ECReqNotEmpty = "ECReqNotEmpty"
)

// Required rule errors.
var (
	// ErrReq is the error returned when a value is required.
	// See [Required] for more details.
	ErrReq = xrr.New("cannot be blank", ECRequired)

	// ErrReqNotEmpty is the error returned for not nil empty values.
	// See [NotEmpty] for more details.
	ErrReqNotEmpty = xrr.New("cannot be blank", ECReqNotEmpty)
)

// Required rules.
var (
	// Required is a validation rule that checks if a value is not empty.
	//
	// A value is considered not empty if
	//  - integer, float: not zero
	//  - bool: true
	//  - string, array, slice, map: len() > 0
	//  - interface, pointer: not nil and the referenced value is not empty
	//  - any other types
	Required = requiredRule{condition: true, skipNil: false, err: ErrReq}

	// NotEmpty checks if a value is a nil pointer or a not empty. It
	// differs from [Required] in that it treats a nil pointer as valid.
	NotEmpty = requiredRule{condition: true, skipNil: true, err: ErrReqNotEmpty}
)

// Compile time checks.
var (
	_ Customizer[requiredRule]  = requiredRule{}
	_ Conditioner[requiredRule] = requiredRule{}
)

// requiredRule is a rule that checks if a value is not empty.
type requiredRule struct {
	condition bool  // Run validation only when true.
	skipNil   bool  // When true, the nil value is considered valid.
	err       error // Validation error.
}

// Validate checks if the given value is valid or not.
func (r requiredRule) Validate(v any) error {
	if !r.condition {
		return nil
	}

	isNil, _ := IsNil(v)
	if r.skipNil && isNil {
		return nil
	}

	if IsEmpty(v) {
		return r.err
	}
	return nil
}

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (r requiredRule) When(condition bool) requiredRule {
	r.condition = condition
	return r
}

// Code sets custom error code for the rule.
func (r requiredRule) Code(code string) requiredRule {
	r.err = setCode(r.err, code)
	return r
}

// Error sets custom error for the rule.
func (r requiredRule) Error(err error) requiredRule {
	r.err = err
	return r
}
