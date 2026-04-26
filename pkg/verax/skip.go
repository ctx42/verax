// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"fmt"

	"github.com/ctx42/verax/pkg/spec"
)

// SkipRuleName represents [SkipRule] name.
const SkipRuleName = "skip-rule"

// Skip is a special validation rule that indicates all rules following it
// should be skipped.
var Skip = SkipRule(true)

// Compile time checks.
var (
	_ conditioner[SkipRule] = SkipRule(false)
	_ Rule                  = SkipRule(false)
)

// SkipRule represents a validation rule that skips later rules.
type SkipRule bool

func (_ SkipRule) Validate(_ any) error         { return nil }
func (_ SkipRule) When(condition bool) SkipRule { return SkipRule(condition) }

func (_ SkipRule) Spec() (*spec.Spec, error) {
	return spec.NewSpec(SkipRuleName), nil
}

// SkipRuleFromSpec creates a new instance of [SkipRule] from the [spec.Spec].
func SkipRuleFromSpec(spc *spec.Spec) (SkipRule, error) {
	if spc.Name != SkipRuleName {
		msg := fmt.Sprintf("%s: invalid spec name: %q", SkipRuleName, spc.Name)
		return false, NewInternalError(msg, spec.ECInvSpec)
	}
	return Skip, nil
}
