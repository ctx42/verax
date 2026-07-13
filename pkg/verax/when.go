// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"github.com/ctx42/xrr/pkg/xrr"
)

// When returns a validation rule that executes the given list of rules when
// the condition is true. Use [WhenRule.Else] to specify rules for the false
// case. Empty values are not short-circuited (use [Skip] or [When] with
// [IsEmpty] for that).
func When(condition bool, rules ...Rule) WhenRule {
	return WhenRule{
		condition: condition,
		rules:     rules,
		elseRules: []Rule{},
	}
}

// Compile time conditions.
var (
	_ customizer[WhenRule] = WhenRule{}
	_ Rule                 = WhenRule{}
)

// WhenRule is a validation rule that applies rules from [When] if the
// condition is met, or rules from [WhenRule.Else] otherwise.
type WhenRule struct {
	condition bool   // Run validation only when true.
	rules     []Rule // Rules applied when condition is true.
	elseRules []Rule // Rules applied when condition is false.
	msg       string // Validation error message.
	code      string // Validation error code.
	flags     uint8  // Customizations.
}

func (r WhenRule) Validate(have any) error {
	var err error
	if r.condition {
		err = Validate(have, r.rules...)
	} else {
		err = Validate(have, r.elseRules...)
	}
	if err != nil {
		customMsg := r.flags&flgCustomMsg != 0
		customCode := r.flags&flgCustomCode != 0

		// Both a custom message and code are set.
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

// Else returns a validation rule that executes the given list of rules when
// the condition passed to [When] constructor function is false.
func (r WhenRule) Else(rules ...Rule) WhenRule {
	r.elseRules = rules
	return r
}

func (r WhenRule) Message(msg string) WhenRule {
	if msg != "" {
		r.msg = msg
		r.flags |= flgCustomMsg
	}
	return r
}

func (r WhenRule) Code(code string) WhenRule {
	if code != "" {
		r.code = code
		r.flags |= flgCustomCode
	}
	return r
}
