// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
)

// InRuleName represents [InRule] name.
const InRuleName = "in-rule"

// ECInvIn represents error code for invalid in rule.
const ECInvIn = "ECInvIn"

// [InRule] rule error messages.
var (
	// msgIn is the error message when a value is not on the list of valid
	// values.
	msgIn = "must be in the list"

	// msgNotIn is the error message when a value is on the list of invalid
	// values.
	msgNotIn = "must not be in the list"
)

// In returns a validation rule that conditions if a value can be found in the
// given list of values. Note that the value being conditioned and the possible
// range of values must be of the same type. Values are compared using
// [reflect.DeepEqual]. An empty value is considered valid. Use the [Required]
// rule to make sure a value is not empty.
func In(values ...any) InRule {
	var eqs []EqualRule
	for _, v := range values {
		eqs = append(eqs, Equal(v))
	}
	return InRule{
		mode:      "in",
		want:      eqs,
		condition: true,
		msg:       msgIn,
		code:      ECInvIn,
	}
}

// NotIn returns a validation rule that conditions if a value cannot be found
// in the given list of values. Note that the value being conditioned and the
// possible range of values must be of the same type. Values are compared using
// [reflect.DeepEqual]. An empty value is considered valid. Use the [Required]
// rule to make sure a value is not empty.
func NotIn(values ...any) InRule {
	var eqs []EqualRule
	for _, v := range values {
		eqs = append(eqs, Equal(v))
	}
	return InRule{
		mode:      "not-in",
		want:      eqs,
		condition: true,
		msg:       msgNotIn,
		code:      ECInvIn,
	}
}

// Compile time conditions.
var (
	_ customizer[InRule]  = InRule{}
	_ conditioner[InRule] = InRule{}
	_ Rule                = InRule{}
)

// InRule is a validation rule that validates if a value can be found in the
// given list of values.
type InRule struct {
	mode      string      // Mode of operation.
	want      []EqualRule // List of valid values.
	condition bool        // Run validation only when true.
	msg       string      // Validation error message.
	code      string      // Validation error code.
	flags     uint8       // Customizations.
}

func (r InRule) Validate(have any) error {
	if !r.condition {
		return nil
	}
	if IsEmpty(have) {
		return nil
	}

	v := Indirect(have)
	for _, e := range r.want {
		err := e.Validate(v)
		if r.mode == "in" && err == nil {
			// We've found a value in the set so we can return with success.
			return nil
		}
		if r.mode == "not-in" && err == nil {
			// We've found value that should not be valid.
			return NewError(r.msg, r.code)
		}
	}
	if r.mode == "in" {
		// None of the conditioned values are on the list of valid values.
		return NewError(r.msg, r.code)
	}

	// None of the values were on the list of invalid values.
	return nil
}

func (r InRule) When(condition bool) InRule {
	r.condition = condition
	return r
}

func (r InRule) Message(msg string) InRule {
	if msg != "" {
		r.msg = msg
		r.flags |= flgCustomMsg
	}
	return r
}

func (r InRule) Code(code string) InRule {
	if code != "" {
		r.code = code
		r.flags |= flgCustomCode
	}
	return r
}

func (r InRule) Spec() (*spec.Spec, error) {
	spc := spec.NewSpec(InRuleName)
	switch r.mode {
	case "in", "not-in":
		spc.SetArg(ArgMode, r.mode)
	default:
		return nil, NewInternalErrorf(
			"%s: invalid rule mode: %q",
			InRuleName,
			r.mode,
			xrr.WithCode(ECInvRuleMode),
		)
	}

	var vls []any
	for _, eq := range r.want {
		vls = append(vls, eq.want)
	}
	spc.SetArg(spec.ArgValues, vls)
	if r.flags&flgCustomMsg != 0 {
		spc.SetArg(ArgErrMsg, r.msg)
	}
	if r.flags&flgCustomCode != 0 {
		spc.SetArg(ArgErrCode, r.code)
	}
	return spc, nil
}

// InRuleFromSpec creates a new instance of [InRule] from the [spec.Spec].
func InRuleFromSpec(spc *spec.Spec) (InRule, error) {
	if spc.Name != InRuleName {
		return InRule{}, NewInternalErrorf(
			"%s: invalid spec name: %q",
			InRuleName,
			spc.Name,
			xrr.WithCode(spec.ECInvSpec),
		)
	}
	mode, err := getArg[string](spc.Args, ArgMode, InRuleName)
	if err != nil {
		return InRule{}, err
	}
	vls, err := getArg[[]any](spc.Args, spec.ArgValues, InRuleName)
	if err != nil {
		return InRule{}, err
	}

	var rule InRule
	switch mode {
	case "in":
		rule = In(vls...)
	case "not-in":
		rule = NotIn(vls...)
	default:
		return InRule{}, NewInternalErrorf(
			"%s: invalid spec rule mode: %q",
			InRuleName,
			mode,
			xrr.WithCode(spec.ECInvSpec),
		)
	}

	if spc.ArgExist(ArgErrMsg) {
		msg, err := getArg[string](spc.Args, ArgErrMsg, InRuleName)
		if err != nil {
			return InRule{}, err
		}
		if msg != "" {
			rule = rule.Message(msg)
		}
	}

	if spc.ArgExist(ArgErrCode) {
		code, err := getArg[string](spc.Args, ArgErrCode, InRuleName)
		if err != nil {
			return InRule{}, err
		}
		if code != "" {
			rule = rule.Code(code)
		}
	}

	return rule, nil
}
