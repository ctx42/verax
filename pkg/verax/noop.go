// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
)

// NoopRuleName represents [NoopRule] name.
const NoopRuleName = "noop-rule"

// Compile time checks.
var (
	_ customizer[NoopRule]  = NoopRule{}
	_ conditioner[NoopRule] = NoopRule{}
	_ Rule                  = NoopRule{}
)

// Noop is an always passing no-op rule.
var Noop = NoopRule{}

// NoopRule is a special validation rule that always passes.
type NoopRule struct{}

func (nop NoopRule) Validate(_ any) error      { return nil }
func (nop NoopRule) When(_ bool) NoopRule      { return nop }
func (nop NoopRule) Code(_ string) NoopRule    { return nop }
func (nop NoopRule) Message(_ string) NoopRule { return nop }

func (nop NoopRule) Spec() (*spec.Spec, error) {
	return spec.NewSpec(NoopRuleName), nil
}

// NoopRuleFromSpec creates a new instance of [NoopRule] from the [spec.Spec].
func NoopRuleFromSpec(spc *spec.Spec) (NoopRule, error) {
	if spc.Name != NoopRuleName {
		return NoopRule{}, NewInternalErrorf(
			"%s: invalid spec name: %q",
			NoopRuleName,
			spc.Name,
			xrr.WithCode(spec.ECInvSpec),
		)
	}
	return Noop, nil
}
