// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"bytes"
	"text/template"
	"unicode/utf8"

	"github.com/ctx42/xrr/pkg/xrr"
)

// ECInvLength represents error code for invalid length.
const ECInvLength = "ECInvLength"

// Length rule error message templates.
var (
	// tplLengthTooLong is the error message template for length too long.
	tplLengthTooLong = emtpl("the length must be no more than {{.max}}")

	// tplLengthTooShort is the error message template for length too short.
	tplLengthTooShort = emtpl("the length must be no less than {{.min}}")

	// tplLengthInvalid is the error message template for an invalid length.
	tplLengthInvalid = emtpl("the length must be exactly {{.min}}")

	// tplLengthOutOfRange is the error message template for out of range length.
	tplLengthOutOfRange = emtpl("the length must be between {{.min}} and {{.max}}")
)

// ErrReqLengthEmpty is the error that returns in the case of non-empty value.
var ErrReqLengthEmpty = xrr.New("the value must be empty", ECReqEmpty)

// Length returns a validation rule that checks if a value's length is within
// the specified range. If max is 0, it means there is no upper bound for the
// length. This rule should only be used for validating strings, slices, maps,
// and arrays. An empty value is considered valid. Use the [Required] rule to
// make sure a value is not empty.
func Length(minimum, maximum int) LengthRule {
	return LengthRule{
		min:       minimum,
		max:       maximum,
		condition: true,
		err:       buildLengthRuleError(minimum, maximum, ECInvLength),
	}
}

// RuneLength returns a validation rule that checks if a string's rune length
// is within the specified range. If max is 0, it means there is no upper bound
// for the length. This rule should only be used for validating strings, slices,
// maps, and arrays. An empty value is considered valid. Use the [Required]
// rule to make sure a value is not empty. If the value being validated is not
// a string, the rule works the same as Length.
func RuneLength(minimum, maximum int) LengthRule {
	r := Length(minimum, maximum)
	r.rune = true
	return r
}

// Compile time checks.
var (
	_ Customizer[LengthRule]  = LengthRule{}
	_ Conditioner[LengthRule] = LengthRule{}
)

// LengthRule is a validation rule that checks if a value's length is within
// the specified range.
type LengthRule struct {
	min       int   // Minimum length.
	max       int   // Maximum length.
	condition bool  // Run validation only when true.
	rune      bool  // Check rune length.
	err       error // Default validation error.
}

// Validate checks if the given value is valid or not.
func (r LengthRule) Validate(v any) error {
	if !r.condition {
		return nil
	}
	if isNil, _ := IsNil(v); isNil {
		return nil
	}

	if IsEmpty(v) {
		return nil
	}

	var l int
	var err error
	val := Indirect(v)
	if s, ok := val.(string); ok && r.rune {
		l = utf8.RuneCountInString(s)
	} else if l, err = LengthOfValue(val); err != nil {
		return err
	}

	if r.min > 0 && l < r.min || r.max > 0 && l > r.max ||
		r.min == 0 && r.max == 0 && l > 0 {
		return r.err
	}
	return nil
}

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (r LengthRule) When(condition bool) LengthRule {
	r.condition = condition
	return r
}

// Code sets the error code for the rule.
func (r LengthRule) Code(code string) LengthRule {
	r.err = setCode(r.err, code)
	return r
}

// Error sets custom error for the rule.
func (r LengthRule) Error(err error) LengthRule {
	r.err = err
	return r
}

// buildLengthRuleError constructs a length rule error.
func buildLengthRuleError(minimum, maximum int, code string) error {
	var tpl *template.Template

	switch {
	case minimum == 0 && maximum > 0:
		tpl = tplLengthTooLong

	case minimum > 0 && maximum == 0:
		tpl = tplLengthTooShort

	case minimum > 0 && maximum > 0:
		if minimum == maximum {
			tpl = tplLengthInvalid
		} else {
			tpl = tplLengthOutOfRange
		}

	default:
		return ErrReqLengthEmpty
	}

	buf := bytes.Buffer{}
	_ = tpl.Execute(&buf, map[string]any{"min": minimum, "max": maximum})
	return xrr.New(buf.String(), code)
}
