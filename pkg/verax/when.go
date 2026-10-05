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

func (whn WhenRule) Validate(have any) error {
	var err error
	if whn.condition {
		err = Validate(have, whn.rules...)
	} else {
		err = Validate(have, whn.elseRules...)
	}
	if err != nil {
		customMsg := whn.flags&flgCustomMsg != 0
		customCode := whn.flags&flgCustomCode != 0

		// Both a custom message and code are set.
		if customMsg && customCode {
			return NewError(whn.msg, whn.code)
		}

		if customMsg {
			return NewError(whn.msg, xrr.GetCode(err))
		}
		if customCode {
			return xrr.SetCode[edError](err, whn.code)
		}
		return err
	}
	return nil
}

// Else returns a validation rule that executes the given list of rules when
// the condition passed to [When] constructor function is false.
func (whn WhenRule) Else(rules ...Rule) WhenRule {
	whn.elseRules = rules
	return whn
}

func (whn WhenRule) Message(msg string) WhenRule {
	if msg != "" {
		whn.msg = msg
		whn.flags |= flgCustomMsg
	}
	return whn
}

func (whn WhenRule) Code(code string) WhenRule {
	if code != "" {
		whn.code = code
		whn.flags |= flgCustomCode
	}
	return whn
}
