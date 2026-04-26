// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"fmt"

	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
)

// FailRuleName represents [FailRule] name.
const FailRuleName = "fail-rule"

// Fail returns the rule which fails with a given error when the condition is
// true. By default, the condition is always true.
func Fail(msg, code string) FailRule {
	r := FailRule{condition: true, code: xrr.ECGeneric}
	return r.Message(msg).Code(code)
}

// Compile time checks.
var (
	_ conditioner[FailRule] = FailRule{}
	_ Rule                  = FailRule{}
)

// FailRule returns an error when its condition evaluates to true.
//
// By default, the condition is always true (fails unconditionally).
// Use [FailRule.When] to specify a custom condition.
type FailRule struct {
	condition bool   // Run validation only when true.
	msg       string // Validation error message.
	code      string // Validation error code.
}

func (r FailRule) Validate(_ any) error {
	if !r.condition {
		return nil
	}
	return NewError(r.msg, r.code)
}

func (r FailRule) When(condition bool) FailRule {
	r.condition = condition
	return r
}

func (r FailRule) Message(msg string) FailRule {
	if msg != "" {
		r.msg = msg
	}
	return r
}

func (r FailRule) Code(code string) FailRule {
	if code != "" {
		r.code = code
	}
	return r
}

func (r FailRule) Spec() (*spec.Spec, error) {
	if r.msg == "" {
		format := "%s: error cannot have an empty message"
		msg := fmt.Sprintf(format, FailRuleName)
		return nil, NewInternalError(msg, ECInternal)
	}

	spc := spec.NewSpec(FailRuleName).SetArg(ArgErrMsg, r.msg)
	if r.code != "" && r.code != xrr.ECGeneric {
		spc.SetArg(ArgErrCode, r.code)
	}
	return spc, nil
}

// FailRuleFromSpec creates a new instance of [FailRule] from the [spec.Spec].
func FailRuleFromSpec(spc *spec.Spec) (FailRule, error) {
	if spc.Name != FailRuleName {
		msg := fmt.Sprintf("%s: invalid spec name: %q", FailRuleName, spc.Name)
		return FailRule{}, NewInternalError(msg, spec.ECInvSpec)
	}

	msg, err := getArg[string](spc.Args, ArgErrMsg, FailRuleName)
	if err != nil {
		return FailRule{}, err
	}

	rule := Fail(msg, "")
	if spc.ArgExist(ArgErrCode) {
		code, err := getArg[string](spc.Args, ArgErrCode, FailRuleName)
		if err != nil {
			return FailRule{}, err
		}
		rule = rule.Code(code)
	}

	return rule, nil
}
