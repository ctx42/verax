// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"regexp"

	"github.com/ctx42/xrr/pkg/xrr"
)

// ECInvMatch represents error code for an invalid regexp match.
const ECInvMatch = "ECInvMatch"

// ErrInvMatch is the error that returns in case of invalid format.
var ErrInvMatch = xrr.New("must be in a valid format", ECInvMatch)

// Match returns a validation rule that checks if a value matches the specified
// regular expression. This rule should only be used for validating strings and
// byte slices, or a validation error will be reported. An empty value is
// considered valid. Use the Required rule to make sure a value is not empty.
func Match(re *regexp.Regexp) MatchRule {
	return MatchRule{
		rx:        re,
		condition: true,
		err:       ErrInvMatch,
	}
}

// Compile time checks.
var (
	_ Customizer[MatchRule]  = MatchRule{}
	_ Conditioner[MatchRule] = MatchRule{}
)

// MatchRule is a validation rule that checks if a value matches the specified
// regular expression.
type MatchRule struct {
	rx        *regexp.Regexp
	condition bool
	err       error
}

// Validate checks if the given value is valid or not.
func (r MatchRule) Validate(v any) error {
	if !r.condition {
		return nil
	}
	if isNil, _ := IsNil(v); isNil {
		return nil
	}

	if IsEmpty(v) {
		return nil
	}

	if r.rx == nil {
		return ErrInvSetup
	}

	val := Indirect(v)
	isString, str, isBytes, bs := StringOrBytes(val)
	if isString && (str == "" || r.rx.MatchString(str)) {
		return nil
	} else if isBytes && (len(bs) == 0 || r.rx.Match(bs)) {
		return nil
	}
	return r.err
}

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (r MatchRule) When(condition bool) MatchRule {
	r.condition = condition
	return r
}

// Code sets the error code for the rule.
func (r MatchRule) Code(code string) MatchRule {
	r.err = setCode(r.err, code)
	return r
}

// Error sets custom error for the rule.
func (r MatchRule) Error(err error) MatchRule {
	r.err = err
	return r
}
