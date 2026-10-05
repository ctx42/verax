// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
)

// AbsentRuleName represents [AbsentRule] name.
const AbsentRuleName = "absent-rule"

// [AbsentRule] error codes.
const (
	// ECReqNil represents error code when a value must be nil but is not.
	ECReqNil = "ECReqNil"

	// ECReqEmpty represents error code when a value must be empty but is not.
	ECReqEmpty = "ECReqEmpty"
)

// [AbsentRule] rule error messages.
const (
	// msgReqNil is the error message when a value is not nil.
	msgReqNil = "must be blank"

	// msgReqEmpty is the error message when a not nil value is not empty.
	msgReqEmpty = "must be blank"
)

// Rules based on [AbsentRule].
var (
	// Nil conditions if a value is nil.
	Nil = AbsentRule{
		condition: true,
		mode:      "nil",
		msg:       msgReqNil,
		code:      ECReqNil,
	}

	// Empty conditions if a not nil value is empty.
	Empty = AbsentRule{
		condition: true,
		mode:      "empty",
		msg:       msgReqEmpty,
		code:      ECReqEmpty,
	}
)

// Compile time conditions.
var (
	_ customizer[AbsentRule]  = AbsentRule{}
	_ conditioner[AbsentRule] = AbsentRule{}
	_ Rule                    = AbsentRule{}
)

// AbsentRule conditions if a value is absent.
type AbsentRule struct {
	mode      string // Mode of operation.
	condition bool   // Run validation only when true.
	msg       string // Validation error message.
	code      string // Validation error code.
	flags     uint8  // Customizations.
}

func (abs AbsentRule) Validate(have any) error {
	if !abs.condition {
		return nil
	}
	isNil := IsNil(have)
	if !isNil && (abs.mode == "nil" || !isEmptyValue(have)) {
		return NewError(abs.msg, abs.code)
	}
	return nil
}

func (abs AbsentRule) When(condition bool) AbsentRule {
	abs.condition = condition
	return abs
}

func (abs AbsentRule) Message(msg string) AbsentRule {
	if msg != "" {
		abs.msg = msg
		abs.flags |= flgCustomMsg
	}
	return abs
}

func (abs AbsentRule) Code(code string) AbsentRule {
	if code != "" {
		abs.code = code
		abs.flags |= flgCustomCode
	}
	return abs
}

func (abs AbsentRule) Spec() (*spec.Spec, error) {
	spc := spec.NewSpec(AbsentRuleName)
	switch abs.mode {
	case "nil", "empty":
		spc.SetArg(ArgMode, abs.mode)

	default:
		return nil, NewInternalErrorf(
			"%s: invalid rule mode: %q",
			AbsentRuleName,
			abs.mode,
			xrr.WithCode(ECInvRuleMode),
		)
	}

	if abs.flags&flgCustomMsg != 0 {
		spc.SetArg(ArgErrMsg, abs.msg)
	}
	if abs.flags&flgCustomCode != 0 {
		spc.SetArg(ArgErrCode, abs.code)
	}
	return spc, nil
}

// AbsentRuleFromSpec creates an instance of [AbsentRule] from the [spec.Spec].
func AbsentRuleFromSpec(spc *spec.Spec) (AbsentRule, error) {
	if spc.Name != AbsentRuleName {
		return AbsentRule{}, NewInternalErrorf(
			"%s: invalid spec name: %q",
			AbsentRuleName,
			spc.Name,
			xrr.WithCode(spec.ECInvSpec),
		)
	}
	mode, err := getArg[string](spc.Args, ArgMode, AbsentRuleName)
	if err != nil {
		return AbsentRule{}, err
	}

	var rule AbsentRule
	switch mode {
	case "empty":
		rule = Empty

	case "nil":
		rule = Nil

	default:
		return AbsentRule{}, NewInternalErrorf(
			"%s: invalid spec rule mode: %q",
			AbsentRuleName,
			mode,
			xrr.WithCode(spec.ECInvSpec),
		)
	}

	if spc.ArgExist(ArgErrMsg) {
		msg, err := getArg[string](spc.Args, ArgErrMsg, AbsentRuleName)
		if err != nil {
			return AbsentRule{}, err
		}
		if msg != "" {
			rule = rule.Message(msg)
		}
	}

	if spc.ArgExist(ArgErrCode) {
		code, err := getArg[string](spc.Args, ArgErrCode, AbsentRuleName)
		if err != nil {
			return AbsentRule{}, err
		}
		if code != "" {
			rule = rule.Code(code)
		}
	}

	return rule, nil
}
