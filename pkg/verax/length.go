// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"bytes"
	"text/template"
	"unicode/utf8"

	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
)

// LengthRuleName represents [LengthRule] name.
const LengthRuleName = "length-rule"

// ECInvLength represents error code for invalid length.
const ECInvLength = "ECInvLength"

// [LengthRule] rule error messages.
var (
	// msgLengthTooLong is the error message template for length too long.
	msgLengthTooLong = "the length must be no more than {{.max}}"

	// tplLengthTooLong is a parsed msgLengthTooLong template.
	tplLengthTooLong = mustTpl(LengthRuleName, msgLengthTooLong)

	// msgLengthTooShort is the error message template for length too short.
	msgLengthTooShort = "the length must be no less than {{.min}}"

	// tplLengthTooShort is a parsed msgLengthTooShort template.
	tplLengthTooShort = mustTpl(LengthRuleName, msgLengthTooShort)

	// msgLengthInvalid is the error message template for an invalid length.
	msgLengthInvalid = "the length must be exactly {{.min}}"

	// tplLengthInvalid is a parsed msgLengthInvalid template.
	tplLengthInvalid = mustTpl(LengthRuleName, msgLengthInvalid)

	// msgLengthOutOfRange is the error message template for out-of-range
	// length.
	msgLengthOutOfRange = "the length must be between {{.min}} and {{.max}}"

	// tplLengthOutOfRange is a parsed msgLengthOutOfRange template.
	tplLengthOutOfRange = mustTpl(LengthRuleName, msgLengthOutOfRange)

	// msgLengthReqEmpty is the error message for non-empty value when both min
	// and max lengths are zero.
	msgLengthReqEmpty = "the value must be empty"

	// tplLengthReqEmpty is a parsed msgLengthReqEmpty template.
	tplLengthReqEmpty = mustTpl(LengthRuleName, msgLengthReqEmpty)
)

// Length returns a validation rule that conditions if a value's length is within
// the specified range. If max is 0, it means there is no upper bound for the
// length. This rule should only be used for validating strings, slices, maps,
// and arrays. An empty value is considered valid. Use the [Required] rule to
// make sure a value is not empty.
func Length(minimum, maximum int) LengthRule {
	r := LengthRule{
		mode:      "length",
		min:       minimum,
		max:       maximum,
		condition: true,
		code:      ECInvLength,
	}
	r.msg, r.tpl, r.sticky = buildLengthRuleMsg(minimum, maximum, r.mode)
	return r
}

// RuneLength returns a validation rule that conditions if a string's rune length
// is within the specified range. If max is 0, it means there is no upper bound
// for the length. This rule should only be used for validating strings, slices,
// maps, and arrays. An empty value is considered valid. Use the [Required]
// rule to make sure a value is not empty. If the value being validated is not
// a string, the rule works the same as Length.
func RuneLength(minimum, maximum int) LengthRule {
	r := LengthRule{
		mode:      "rune-length",
		min:       minimum,
		max:       maximum,
		condition: true,
		code:      ECInvLength,
	}
	r.msg, r.tpl, r.sticky = buildLengthRuleMsg(minimum, maximum, r.mode)
	return r
}

// Compile time conditions.
var (
	_ customizer[LengthRule]  = LengthRule{}
	_ conditioner[LengthRule] = LengthRule{}
	_ Rule                    = LengthRule{}
)

// LengthRule is a validation rule that conditions if a value's length is within
// the specified range.
type LengthRule struct {
	mode      string // Mode of operation.
	min       int    // Minimum length.
	max       int    // Maximum length.
	condition bool   // Run validation only when true.
	tpl       string // The original error message template.
	msg       string // Validation error message rendered from template.
	code      string // Validation error code.
	sticky    error  // Sticky error.
	flags     uint8  // Customizations.
}

// nolint: cyclop
func (r LengthRule) Validate(have any) error {
	if r.sticky != nil {
		return r.sticky
	}
	if !r.condition {
		return nil
	}
	if IsEmpty(have) {
		return nil
	}

	var l int
	var err error
	val := Indirect(have)
	if s, ok := val.(string); ok && r.mode == "rune-length" {
		l = utf8.RuneCountInString(s)
	} else if l, err = LengthOfValue(val); err != nil {
		return err
	}

	if r.min > 0 && l < r.min || r.max > 0 && l > r.max ||
		r.min == 0 && r.max == 0 && l > 0 {
		return NewError(r.msg, r.code)
	}
	return nil
}

func (r LengthRule) When(condition bool) LengthRule {
	r.condition = condition
	return r
}

func (r LengthRule) Message(tpl string) LengthRule {
	if r.sticky != nil || tpl == "" {
		return r
	}
	var parsed *template.Template
	parsed, r.sticky = template.New(LengthRuleName).
		Option("missingkey=error").
		Parse(tpl)
	if r.sticky != nil {
		r.sticky = NewInternalErrorf(
			"%s(%s): custom template parse error",
			LengthRuleName,
			r.mode,
			xrr.WithCode(ECInternal),
		)
		return r
	}

	buf := &bytes.Buffer{}
	data := map[string]any{ArgMin: r.min, ArgMax: r.max}
	r.sticky = parsed.Execute(buf, data)
	if r.sticky != nil {
		r.sticky = NewInternalErrorf(
			"%s(%s): custom template render error",
			LengthRuleName,
			r.mode,
			xrr.WithCode(ECInternal),
		)
		return r
	}
	r.tpl = tpl
	r.msg = buf.String()
	r.flags |= flgCustomMsg
	return r
}

func (r LengthRule) Code(code string) LengthRule {
	if code != "" {
		r.code = code
		r.flags |= flgCustomCode
	}
	return r
}

func (r LengthRule) Spec() (*spec.Spec, error) {
	if r.sticky != nil {
		return nil, r.sticky
	}

	spc := spec.NewSpec(LengthRuleName)
	switch r.mode {
	case "length", "rune-length":
		spc.SetArg(ArgMode, r.mode)
	default:
		return nil, NewInternalErrorf(
			"%s: invalid rule mode: %q",
			LengthRuleName,
			r.mode,
			xrr.WithCode(ECInvRuleMode),
		)
	}

	spc.SetArg(ArgMin, r.min)
	spc.SetArg(ArgMax, r.max)

	if r.flags&flgCustomMsg != 0 {
		spc.SetArg(ArgErrMsg, r.tpl)
	}
	if r.flags&flgCustomCode != 0 {
		spc.SetArg(ArgErrCode, r.code)
	}
	return spc, nil
}

// LengthRuleFromSpec creates a new instance of [LengthRule] from the
// [spec.Spec].
func LengthRuleFromSpec(spc *spec.Spec) (LengthRule, error) {
	if spc.Name != LengthRuleName {
		return LengthRule{}, NewInternalErrorf(
			"%s: invalid spec name: %q",
			LengthRuleName,
			spc.Name,
			xrr.WithCode(spec.ECInvSpec),
		)
	}

	mode, err := getArg[string](spc.Args, ArgMode, LengthRuleName)
	if err != nil {
		return LengthRule{}, err
	}

	iMin, err := getArg[int](spc.Args, ArgMin, LengthRuleName)
	if err != nil {
		return LengthRule{}, err
	}

	iMax, err := getArg[int](spc.Args, ArgMax, LengthRuleName)
	if err != nil {
		return LengthRule{}, err
	}

	var rule LengthRule
	switch mode {
	case "length":
		rule = Length(iMin, iMax)
	case "rune-length":
		rule = RuneLength(iMin, iMax)
	default:
		return LengthRule{}, NewInternalErrorf(
			"%s: invalid spec rule mode: %q",
			LengthRuleName,
			mode,
			xrr.WithCode(spec.ECInvSpec),
		)
	}

	if spc.ArgExist(ArgErrMsg) {
		msg, err := getArg[string](spc.Args, ArgErrMsg, LengthRuleName)
		if err != nil {
			return LengthRule{}, err
		}
		if msg != "" {
			rule = rule.Message(msg)
		}
	}

	if spc.ArgExist(ArgErrCode) {
		code, err := getArg[string](spc.Args, ArgErrCode, LengthRuleName)
		if err != nil {
			return LengthRule{}, err
		}
		if code != "" {
			rule = rule.Code(code)
		}
	}

	return rule, nil
}

// pickLengthRuleMsg picks the error message template and corresponding parsed
// template based on values of min and max. For invalid values returns empty
// string and nil.
//
// Valid values:
//   - min >= 0 and max >= 0
func pickLengthRuleMsg(minimum, maximum int) (string, *template.Template) {
	var tpl string
	var parsed *template.Template

	switch {
	case minimum == 0 && maximum > 0:
		tpl = msgLengthTooLong
		parsed = tplLengthTooLong

	case minimum > 0 && maximum == 0:
		tpl = msgLengthTooShort
		parsed = tplLengthTooShort

	case minimum > 0 && maximum > 0:
		if minimum == maximum {
			tpl = msgLengthInvalid
			parsed = tplLengthInvalid
		} else if minimum < maximum {
			tpl = msgLengthOutOfRange
			parsed = tplLengthOutOfRange
		}

	case minimum == 0 && maximum == 0:
		tpl = msgLengthReqEmpty
		parsed = tplLengthReqEmpty
	}
	return tpl, parsed
}

// buildLengthRuleMsg constructs the [LengthRule] error message based on wanted
// minimum nad maximum values. Returns the error message and its corresponding
// template.
func buildLengthRuleMsg(
	minimum, maximum int,
	mode string,
) (string, string, error) {

	tpl, parsed := pickLengthRuleMsg(minimum, maximum)
	if tpl == "" || parsed == nil {
		return "", "", NewInternalErrorf(
			"%s(%s): custom template parse error",
			LengthRuleName,
			mode,
			xrr.WithCode(ECInternal),
		)
	}

	buf := bytes.Buffer{}
	err := parsed.Execute(&buf, map[string]any{"min": minimum, "max": maximum})
	if err != nil {
		return "", "", NewInternalErrorf(
			"%s(%s): custom template render error",
			LengthRuleName,
			mode,
			xrr.WithCode(ECInternal),
		)
	}
	return buf.String(), tpl, nil
}
