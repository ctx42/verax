// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"github.com/ctx42/xrr/pkg/xrr"

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

func (s SkipRule) Spec() (*spec.Spec, error) {
	spc := spec.NewSpec(SkipRuleName)
	if !s {
		spc.SetArg(spec.ArgValue, false)
	}
	return spc, nil
}

// SkipRuleFromSpec creates a new instance of [SkipRule] from the [spec.Spec].
func SkipRuleFromSpec(spc *spec.Spec) (SkipRule, error) {
	if spc.Name != SkipRuleName {
		return false, NewInternalErrorf(
			"%s: invalid spec name: %q",
			SkipRuleName,
			spc.Name,
			xrr.WithCode(spec.ECInvSpec),
		)
	}
	if spc.ArgExist(spec.ArgValue) {
		val, err := getArg[bool](spc.Args, spec.ArgValue, SkipRuleName)
		if err != nil {
			return false, err
		}
		return SkipRule(val), nil
	}
	return Skip, nil
}
