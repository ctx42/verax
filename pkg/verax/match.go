// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"regexp"

	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
)

// MatchRuleName represents [MatchRule] name.
const MatchRuleName = "match-rule"

// ECInvMatch represents error code for an invalid regexp match.
const ECInvMatch = "ECInvMatch"

// [MatchRule] rule error messages.
var (
	// msgInvMatch is the error message when a value does not match the regexp.
	msgInvMatch = "must match a valid format"
)

// Match returns a validation rule that conditions if a value matches the specified
// regular expression. This rule should only be used for validating strings and
// byte slices, or a validation error will be reported. An empty value is
// considered valid. Use the [Required] rule to make sure a value is not empty.
func Match(want *regexp.Regexp) MatchRule {
	r := MatchRule{
		want:      want,
		condition: true,
		msg:       msgInvMatch,
		code:      ECInvMatch,
	}
	if r.want == nil {
		r.sticky = NewInternalErrorf(
			"%s: nil regexp",
			MatchRuleName,
			xrr.WithCode(ECInternal),
		)
	}
	return r
}

// Compile time conditions.
var (
	_ customizer[MatchRule]  = MatchRule{}
	_ conditioner[MatchRule] = MatchRule{}
	_ Rule                   = MatchRule{}
)

// MatchRule is a validation rule that conditions if a value matches the specified
// regular expression.
type MatchRule struct {
	want      *regexp.Regexp // Regexp a value must match.
	condition bool           // Run validation only when true.
	msg       string         // Validation error message rendered from template.
	code      string         // Validation error code.
	sticky    error          // Sticky error.
	flags     uint8          // Customizations.
}

func (r MatchRule) Validate(have any) error {
	if r.sticky != nil {
		return r.sticky
	}
	if !r.condition {
		return nil
	}
	if IsEmpty(have) {
		return nil
	}

	val := Indirect(have)
	isString, str, isBytes, bs := StringOrBytes(val)
	if isString && (str == "" || r.want.MatchString(str)) {
		return nil
	} else if isBytes && (len(bs) == 0 || r.want.Match(bs)) {
		return nil
	}
	return NewError(r.msg, r.code)
}

func (r MatchRule) When(condition bool) MatchRule {
	r.condition = condition
	return r
}

func (r MatchRule) Message(msg string) MatchRule {
	if msg != "" {
		r.msg = msg
		r.flags |= flgCustomMsg
	}
	return r
}

func (r MatchRule) Code(code string) MatchRule {
	if code != "" {
		r.code = code
		r.flags |= flgCustomCode
	}
	return r
}

func (r MatchRule) Spec() (*spec.Spec, error) {
	if r.sticky != nil {
		return nil, r.sticky
	}
	spc := spec.NewSpec(MatchRuleName).SetArg(spec.ArgValue, r.want.String())
	if r.flags&flgCustomMsg != 0 {
		spc.SetArg(ArgErrMsg, r.msg)
	}
	if r.flags&flgCustomCode != 0 {
		spc.SetArg(ArgErrCode, r.code)
	}
	return spc, nil
}

// MatchRuleFromSpec creates a new instance of [MatchRule] from the [spec.Spec].
func MatchRuleFromSpec(spc *spec.Spec) (MatchRule, error) {
	if spc.Name != MatchRuleName {
		return MatchRule{}, NewInternalErrorf(
			"%s: invalid spec name: %q",
			MatchRuleName,
			spc.Name,
			xrr.WithCode(spec.ECInvSpec),
		)
	}
	rxs, err := getArg[string](spc.Args, spec.ArgValue, MatchRuleName)
	if err != nil {
		return MatchRule{}, err
	}
	rx, err := regexp.Compile(rxs)
	if err != nil {
		return MatchRule{}, NewInternalErrorf(
			"%s: invalid regexp: %#q",
			MatchRuleName,
			rxs,
			xrr.WithCode(spec.ECInvSpec),
		)
	}
	rule := Match(rx)

	if spc.ArgExist(ArgErrMsg) {
		msg, err := getArg[string](spc.Args, ArgErrMsg, MatchRuleName)
		if err != nil {
			return MatchRule{}, err
		}
		if msg != "" {
			rule = rule.Message(msg)
		}
	}

	if spc.ArgExist(ArgErrCode) {
		code, err := getArg[string](spc.Args, ArgErrCode, MatchRuleName)
		if err != nil {
			return MatchRule{}, err
		}
		if code != "" {
			rule = rule.Code(code)
		}
	}

	return rule, nil
}
