// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

// By wraps a [RuleFunc].
func By(fn RuleFunc) ByRule { return ByRule{fn: fn, condition: true} }

// Compile time checks.
var (
	_ Customizer[ByRule]  = ByRule{}
	_ Conditioner[ByRule] = ByRule{}
)

// ByRule is a validation rule that checks if a value passed to a validation
// function.
type ByRule struct {
	fn        RuleFunc // Validation function.
	condition bool     // Run validation only when true.
	err       error    // Custom rule error.
	code      string   // Custom error code.
}

// Validate checks if the given value is valid or not.
func (r ByRule) Validate(v any) error {
	if !r.condition {
		return nil
	}
	if err := r.fn(v); err != nil {
		if r.err != nil {
			err = r.err
		}
		return setCode(err, r.code)
	}
	return nil
}

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (r ByRule) When(condition bool) ByRule {
	r.condition = condition
	return r
}

// Code sets the error code for the rule.
func (r ByRule) Code(code string) ByRule {
	r.code = code
	return r
}

// Error sets custom error for the rule.
func (r ByRule) Error(err error) ByRule {
	r.err = err
	return r
}
