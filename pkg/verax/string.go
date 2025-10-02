// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

// ValidStringFunc is a function type for validating string values. Returns
// true if the string is valid, false otherwise.
type ValidStringFunc func(string) bool

// Compile time checks.
var (
	_ Customizer[StringRule]  = StringRule{}
	_ Conditioner[StringRule] = StringRule{}
)

// String creates a new string validation rule using the [ValidStringFunc]
// function. An empty string is considered to be valid. Please use the
// [Required] rule to make sure a value is not empty.
func String(fn ValidStringFunc) StringRule {
	return StringRule{
		fn:        fn,
		condition: true,
		err:       ErrNotEqual,
	}
}

// StringRule is a rule that checks a string passes the [ValidStringFunc].
type StringRule struct {
	fn        ValidStringFunc // Validation function.
	condition bool            // Run validation only when true.
	err       error           // Validation error.
}

// Validate checks if the given value is valid or not.
func (r StringRule) Validate(v any) error {
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
	str, err := EnsureString(val)
	if err != nil {
		return err
	}
	if r.fn(str) {
		return nil
	}
	return r.err
}

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (r StringRule) When(condition bool) StringRule {
	r.condition = condition
	return r
}

// Code sets the error code for the rule.
func (r StringRule) Code(code string) StringRule {
	r.err = setCode(r.err, code)
	return r
}

// Error sets custom error for the rule.
func (r StringRule) Error(err error) StringRule {
	r.err = err
	return r
}
