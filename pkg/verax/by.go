// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
)

// ByRuleName represents [ByRule] name.
const ByRuleName = "by-rule"

// By validates a value using the provided [RuleFunc]. This is the primary
// escape hatch for custom validation logic. The function receives the value
// and should return nil on success or a verax domain error on failure
// (see [Rule]). Use [Check] for the simple func(bool) case.
func By(fn RuleFunc) ByRule { return ByRule{fn: fn, condition: true} }

// Compile time conditions.
var (
	_ customizer[ByRule]  = ByRule{}
	_ conditioner[ByRule] = ByRule{}
	_ Rule                = ByRule{}
)

// ByRule conditions if a value passed to a [RuleFunc] function is valid.
type ByRule struct {
	fn        RuleFunc // Validation function.
	condition bool     // Run validation only when true.
	msg       string   // Validation error message.
	code      string   // Validation error code.
	flags     uint8    // Customizations.
}

func (r ByRule) Validate(have any) error {
	if !r.condition {
		return nil
	}
	if IsEmpty(have) {
		return nil
	}
	if err := r.fn(have); err != nil {
		customMsg := r.flags&flgCustomMsg != 0
		customCode := r.flags&flgCustomCode != 0

		// Both custom message and code are set.
		if customMsg && customCode {
			return NewError(r.msg, r.code)
		}

		if customMsg {
			return NewError(r.msg, xrr.GetCode(err))
		}
		if customCode {
			return xrr.SetCode[edError](err, r.code)
		}
		return err
	}
	return nil
}

func (r ByRule) When(condition bool) ByRule {
	r.condition = condition
	return r
}

func (r ByRule) Message(msg string) ByRule {
	if msg != "" {
		r.msg = msg
		r.flags |= flgCustomMsg
	}
	return r
}

func (r ByRule) Code(code string) ByRule {
	if code != "" {
		r.code = code
		r.flags |= flgCustomCode
	}
	return r
}

func (r ByRule) Spec() (*spec.Spec, error) {
	spc := spec.NewSpec(ByRuleName).SetArg(spec.ArgSrc, r.fn)

	if r.flags&flgCustomMsg != 0 {
		spc.SetArg(ArgErrMsg, r.msg)
	}
	if r.flags&flgCustomCode != 0 {
		spc.SetArg(ArgErrCode, r.code)
	}
	return spc, nil
}

// ByRuleFromSpec creates a new instance of [ByRule] from the [spec.Spec].
func ByRuleFromSpec(spc *spec.Spec) (ByRule, error) {
	if spc.Name != ByRuleName {
		return ByRule{}, NewInternalErrorf(
			"%s: invalid spec name: %q",
			ByRuleName,
			spc.Name,
			xrr.WithCode(spec.ECInvSpec),
		)
	}
	fn, err := getArg[RuleFunc](spc.Args, spec.ArgSrc, ByRuleName)
	if err != nil {
		return ByRule{}, err
	}

	rule := By(fn)
	if spc.ArgExist(ArgErrMsg) {
		msg, err := getArg[string](spc.Args, ArgErrMsg, ByRuleName)
		if err != nil {
			return ByRule{}, err
		}
		if msg != "" {
			rule = rule.Message(msg)
		}
	}

	if spc.ArgExist(ArgErrCode) {
		code, err := getArg[string](spc.Args, ArgErrCode, ByRuleName)
		if err != nil {
			return ByRule{}, err
		}
		if code != "" {
			rule = rule.Code(code)
		}
	}

	return rule, nil
}
