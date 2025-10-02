// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

// When returns a validation rule that executes the given list of rules when
// the condition is true.
func When(condition bool, rules ...Rule) WhenRule {
	return WhenRule{
		condition: condition,
		rules:     rules,
		elseRules: []Rule{},
	}
}

// Compile time checks.
var (
	_ Customizer[WhenRule] = WhenRule{}
)

// WhenRule is a validation rule that applies rules from [When] if the
// condition is met, or rules from [WhenRule.Else] otherwise.
type WhenRule struct {
	condition bool   // Run validation only when true.
	rules     []Rule // When rules.
	elseRules []Rule // Else rules.
	err       error  // Custom rule error.
	code      string // Custom error code.
}

// Validate checks if the condition is true, and if so, it validates the value
// using the specified rules.
func (r WhenRule) Validate(value any) error {
	var err error
	if r.condition {
		err = Validate(value, r.rules...)
	} else {
		err = Validate(value, r.elseRules...)
	}
	if err != nil {
		if r.err != nil {
			return setCode(r.err, r.code)
		}
		return setCode(err, r.code)
	}
	return nil
}

// Else returns a validation rule that executes the given list of rules when
// the condition is false.
func (r WhenRule) Else(rules ...Rule) WhenRule {
	r.elseRules = rules
	return r
}

// Code sets the error code for the rule.
func (r WhenRule) Code(code string) WhenRule {
	r.code = code
	return r
}

// Error sets custom error for the rule.
func (r WhenRule) Error(err error) WhenRule {
	r.err = err
	return r
}
