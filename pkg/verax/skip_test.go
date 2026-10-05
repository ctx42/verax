// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_Skip(t *testing.T) {
	// --- When ---
	have := Skip

	// --- Then ---
	assert.True(t, bool(have))
}

func Test_SkipRule_Validate_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rule SkipRule
		have any
	}{
		{"true nil", SkipRule(true), nil},
		{"true int", SkipRule(true), 100},
		{"true string", SkipRule(true), "str"},
		{"true float", SkipRule(true), 1.1},
		{"false nil", SkipRule(false), nil},
		{"false int", SkipRule(false), 100},
		{"false string", SkipRule(false), "str"},
		{"false float", SkipRule(false), 1.1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := tc.rule.Validate(tc.have)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_SkipRule_When(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		// --- Given ---
		r := SkipRule(false)

		// --- When ---
		have := r.When(true)

		// --- Then ---
		assert.True(t, bool(have))
	})

	t.Run("false", func(t *testing.T) {
		// --- When ---
		have := Skip.When(false)

		// --- Then ---
		assert.False(t, bool(have))
	})
}

func Test_SkipRule_Spec(t *testing.T) {
	t.Run("Skip", func(t *testing.T) {
		// --- Given ---
		r := Skip

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, SkipRuleName, have.Name)
		assert.Nil(t, have.Args)
	})

	t.Run("Skip - JSON encode", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()

		// --- When ---
		spc, err := Skip.Spec()

		// --- Then ---
		assert.NoError(t, err)
		data := must.Value(reg.EncodeSpec(spc))
		want := `{"name": "skip-rule"}`
		assert.JSON(t, want, data)
	})

	t.Run("Skip - JSON decode", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()
		reg.RegisterBuilders(Builders())
		data := []byte(`{"name": "skip-rule"}`)

		// --- When ---
		have, err := reg.DecodeAndBuild(data)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Skip, have)
	})
}

func Test_SkipRuleFromSpec(t *testing.T) {
	t.Run("error - not skip rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := SkipRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `skip-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - value not bool", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(SkipRuleName).SetArg(spec.ArgValue, 1)

		// --- When ---
		have, err := SkipRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `skip-rule: spec argument "value" must be bool, got int`
		assert.ErrorEqual(t, wMsg, err)
		assert.False(t, bool(have))
	})

	t.Run("Skip", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(SkipRuleName)

		// --- When ---
		have, err := SkipRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Skip, have)
	})
}

func Test_SkipRule_Spec_SkipRuleFromSpec_round_trip(t *testing.T) {
	t.Run("Skip", func(t *testing.T) {
		// --- Given ---
		want := Skip
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := SkipRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("Skip when false", func(t *testing.T) {
		// --- Given ---
		want := Skip.When(false)
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := SkipRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}
