// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"fmt"

	"github.com/ctx42/xrr/pkg/xrr"
)

// ECInvDynamic represents error code for not implemented RuleFunc.
const ECInvDynamic = "ECInvDynamic"

// ErrInvDynamic is the error returned in the case of not implemented RuleFunc.
var ErrInvDynamic = xrr.New("dynamic function must be set", ECInvDynamic)

// errFn is default dynamic validation function.
var errFn = func(v any) error { return ErrInvDynamic }

// Dynamic wraps a packet and function represented by string.
func Dynamic(pkt, fn string) DynamicRule {
	return DynamicRule{
		pkg:       pkt,
		fnName:    fn,
		by:        errFn,
		condition: true,
	}
}

// Compile time checks.
var (
	_ Customizer[DynamicRule]  = DynamicRule{}
	_ Conditioner[DynamicRule] = DynamicRule{}
)

// DynamicRule is a validation rule that checks a value using a validation
// function that must be provided during execution. If the validate function is
// not implemented, it will return the [ErrInvDynamic] error.
type DynamicRule struct {
	pkg       string   // Package name.
	fnName    string   // Function name.
	by        RuleFunc // Validation function.
	condition bool     // Run validation only when true.
	err       error    // Custom rule error.
	code      string   // Custom error code.
}

// Validate checks if the given value is valid or not.
func (r DynamicRule) Validate(v any) error {
	if !r.condition {
		return nil
	}
	if r.by == nil {
		return ErrInvSetup
	}

	if err := r.by(v); err != nil {
		if r.err != nil {
			err = r.err
		}
		return setCode(err, r.code)
	}
	return nil
}

// Reference returns dynamic rule reference.
func (r DynamicRule) Reference() string {
	return fmt.Sprintf("%s.%s", r.pkg, r.fnName)
}

// RuleFunc sets the error code for the rule.
func (r DynamicRule) RuleFunc(fn RuleFunc) DynamicRule {
	r.by = fn
	return r
}

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (r DynamicRule) When(condition bool) DynamicRule {
	r.condition = condition
	return r
}

// Code sets the error code for the rule.
func (r DynamicRule) Code(code string) DynamicRule {
	r.code = code
	return r
}

// Error sets custom error for the rule.
func (r DynamicRule) Error(err error) DynamicRule {
	r.err = err
	return r
}
