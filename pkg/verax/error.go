// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

// Error returns the rule which fails with a given error when the condition is
// true. By default, the condition is always true.
func Error(err error) ErrorRule {
	return ErrorRule{
		condition: true,
		err:       err,
	}
}

var _ Conditioner[ErrorRule] = ErrorRule{} // Compile time check.

// ErrorRule is a rule that returns an error if the condition is true.
// By default, the condition is always true.
type ErrorRule struct {
	condition bool  // Run validation only when true.
	err       error // Validation error.
}

func (r ErrorRule) Validate(_ any) error {
	if !r.condition {
		return nil
	}
	return r.err
}

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (r ErrorRule) When(condition bool) ErrorRule {
	r.condition = condition
	return r
}

// Code sets the error code for the rule.
func (r ErrorRule) Code(code string) ErrorRule {
	r.err = setCode(r.err, code)
	return r
}
