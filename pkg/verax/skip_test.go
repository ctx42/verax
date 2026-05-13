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
	// --- Given ---
	r := Skip

	// --- Then ---
	//goland:noinspection ALL
	assert.True(t, bool(r))
}

func Test_SkipRule_Validate_tabular(t *testing.T) {
	tt := []struct {
		testN string

		have any
	}{
		{"nil", nil},
		{"int", 100},
		{"string", "str"},
		{"float", 1.1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			rT := SkipRule(true)
			rF := SkipRule(false)

			// --- When ---
			errT := rT.Validate(tc.have)
			errF := rF.Validate(tc.have)

			// --- Then ---
			assert.NoError(t, errT)
			assert.NoError(t, errF)
		})
	}
}

func Test_SkipRule_When(t *testing.T) {
	// --- Given ---
	r := Skip

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, bool(have))
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

	t.Run("Skip - JSON representation", func(t *testing.T) {
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
	// --- Given ---
	want := Skip
	spc := must.Value(want.Spec())

	// --- When ---
	have, err := SkipRuleFromSpec(spc)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, want, have)
}
