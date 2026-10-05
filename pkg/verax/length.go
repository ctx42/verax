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
const (
	// msgLengthTooLong is the error message template for length too long.
	msgLengthTooLong = "the length must be no more than {{.max}}"

	// msgLengthTooShort is the error message template for length too short.
	msgLengthTooShort = "the length must be no less than {{.min}}"

	// msgLengthInvalid is the error message template for an invalid length.
	msgLengthInvalid = "the length must be exactly {{.min}}"

	// msgLengthOutOfRange is the error message template for out-of-range
	// length.
	msgLengthOutOfRange = "the length must be between {{.min}} and {{.max}}"

	// msgLengthReqEmpty is the error message for non-empty value when both min
	// and max lengths are zero.
	msgLengthReqEmpty = "the value must be empty"
)

// Parsed message templates.
var (
	// tplLengthTooLong is a parsed msgLengthTooLong template.
	tplLengthTooLong = mustTpl(LengthRuleName, msgLengthTooLong)

	// tplLengthTooShort is a parsed msgLengthTooShort template.
	tplLengthTooShort = mustTpl(LengthRuleName, msgLengthTooShort)

	// tplLengthInvalid is a parsed msgLengthInvalid template.
	tplLengthInvalid = mustTpl(LengthRuleName, msgLengthInvalid)

	// tplLengthOutOfRange is a parsed msgLengthOutOfRange template.
	tplLengthOutOfRange = mustTpl(LengthRuleName, msgLengthOutOfRange)

	// tplLengthReqEmpty is a parsed msgLengthReqEmpty template.
	tplLengthReqEmpty = mustTpl(LengthRuleName, msgLengthReqEmpty)
)

// Length returns a validation rule that conditions if a value's length is
// within the specified range. If max is 0, it means there is no upper bound
// for the length. This rule should only be used for validating strings,
// slices, maps, and arrays. An empty value is considered valid. Use the
// [Required] rule to make sure a value is not empty.
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

// RuneLength returns a validation rule that conditions if a string's rune
// length is within the specified range. If max is 0, it means there is no
// upper bound for the length. This rule should only be used for validating
// strings, slices, maps, and arrays. An empty value is considered valid. Use
// the [Required] rule to make sure a value is not empty. If the value being
// validated is not a string, the rule works the same as Length.
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

//nolint:cyclop
func (lng LengthRule) Validate(have any) error {
	if lng.sticky != nil {
		return lng.sticky
	}
	if !lng.condition {
		return nil
	}
	if IsEmpty(have) {
		return nil
	}

	var l int
	var err error
	val := Indirect(have)
	if s, ok := val.(string); ok && lng.mode == "rune-length" {
		l = utf8.RuneCountInString(s)
	} else if l, err = LengthOfValue(val); err != nil {
		return err
	}

	if lng.min > 0 && l < lng.min || lng.max > 0 && l > lng.max ||
		lng.min == 0 && lng.max == 0 && l > 0 {
		return NewError(lng.msg, lng.code)
	}
	return nil
}

func (lng LengthRule) When(condition bool) LengthRule {
	lng.condition = condition
	return lng
}

// Message sets a custom error message. The message is a text/template
// rendered with the bounds as {{.min}} and {{.max}}. An empty template is
// ignored. A template that fails to parse or render sets a sticky
// [InternalError] returned by [LengthRule.Validate] and [LengthRule.Spec].
func (lng LengthRule) Message(tpl string) LengthRule {
	if lng.sticky != nil || tpl == "" {
		return lng
	}
	parsed, err := template.New(LengthRuleName).
		Option("missingkey=error").
		Parse(tpl)
	if err != nil {
		lng.sticky = NewInternalErrorf(
			"%s(%s): custom template parse error",
			LengthRuleName,
			lng.mode,
			xrr.WithCode(ECInternal),
			xrr.WithCause(err),
		)
		return lng
	}

	buf := &bytes.Buffer{}
	data := map[string]any{ArgMin: lng.min, ArgMax: lng.max}
	if err = parsed.Execute(buf, data); err != nil {
		lng.sticky = NewInternalErrorf(
			"%s(%s): custom template render error",
			LengthRuleName,
			lng.mode,
			xrr.WithCode(ECInternal),
			xrr.WithCause(err),
		)
		return lng
	}
	lng.tpl = tpl
	lng.msg = buf.String()
	lng.flags |= flgCustomMsg
	return lng
}

func (lng LengthRule) Code(code string) LengthRule {
	if code != "" {
		lng.code = code
		lng.flags |= flgCustomCode
	}
	return lng
}

func (lng LengthRule) Spec() (*spec.Spec, error) {
	if lng.sticky != nil {
		return nil, lng.sticky
	}

	spc := spec.NewSpec(LengthRuleName)
	switch lng.mode {
	case "length", "rune-length":
		spc.SetArg(ArgMode, lng.mode)

	default:
		return nil, NewInternalErrorf(
			"%s: invalid rule mode: %q",
			LengthRuleName,
			lng.mode,
			xrr.WithCode(ECInvRuleMode),
		)
	}

	spc.SetArg(ArgMin, lng.min)
	spc.SetArg(ArgMax, lng.max)

	if lng.flags&flgCustomMsg != 0 {
		spc.SetArg(ArgErrMsg, lng.tpl)
	}
	if lng.flags&flgCustomCode != 0 {
		spc.SetArg(ArgErrCode, lng.code)
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
// minimum and maximum values. Returns the error message and its corresponding
// template.
func buildLengthRuleMsg(
	minimum int,
	maximum int,
	mode string,
) (string, string, error) {

	tpl, parsed := pickLengthRuleMsg(minimum, maximum)
	if tpl == "" || parsed == nil {
		return "", "", NewInternalErrorf(
			"%s(%s): invalid length range: min %d, max %d",
			LengthRuleName,
			mode,
			minimum,
			maximum,
			xrr.WithCode(ECInternal),
		)
	}
	msg, err := renderLengthMsg(parsed, minimum, maximum, mode)
	if err != nil {
		return "", "", err
	}
	return msg, tpl, nil
}

// renderLengthMsg renders the parsed [LengthRule] message template with the
// minimum and maximum values.
func renderLengthMsg(
	parsed *template.Template,
	minimum int,
	maximum int,
	mode string,
) (string, error) {

	buf := bytes.Buffer{}
	err := parsed.Execute(&buf, map[string]any{"min": minimum, "max": maximum})
	if err != nil {
		return "", NewInternalErrorf(
			"%s(%s): custom template render error",
			LengthRuleName,
			mode,
			xrr.WithCode(ECInternal),
			xrr.WithCause(err),
		)
	}
	return buf.String(), nil
}
