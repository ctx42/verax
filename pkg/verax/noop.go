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

func (r NoopRule) Validate(_ any) error      { return nil }
func (r NoopRule) When(_ bool) NoopRule      { return r }
func (r NoopRule) Code(_ string) NoopRule    { return r }
func (r NoopRule) Message(_ string) NoopRule { return r }

func (r NoopRule) Spec() (*spec.Spec, error) {
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
