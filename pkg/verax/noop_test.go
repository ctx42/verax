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

func Test_Noop(t *testing.T) {
	// --- Given ---
	r := Noop.When(true).Code("ECTst").Message("test err")

	// --- Then ---
	assert.NoError(t, r.Validate("abc"))
}

func Test_NoopRule_Validate(t *testing.T) {
	// --- Given ---
	r := NoopRule{}

	// --- When ---
	err := r.Validate("abc")

	// --- Then ---
	assert.NoError(t, err)
}

func Test_Noop_Spec(t *testing.T) {
	t.Run("Noop", func(t *testing.T) {
		// --- Given ---
		r := NoopRule{}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, NoopRuleName, have.Name)
		assert.Nil(t, have.Args)
	})

	t.Run("Noop - JSON representation", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()

		// --- When ---
		spc, err := Noop.Spec()

		// --- Then ---
		assert.NoError(t, err)
		data := must.Value(reg.EncodeSpec(spc))
		want := `{"name": "noop-rule"}`
		assert.JSON(t, want, data)
	})
}

func Test_NoopRuleFromSpec(t *testing.T) {
	t.Run("error - not noop rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := NoopRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `noop-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("Noop", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(NoopRuleName)

		// --- When ---
		have, err := NoopRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Noop, have)
	})
}

func Test_NoopRule_Spec_NoopRuleFromSpec_round_trip(t *testing.T) {
	// --- Given ---
	want := Noop
	spc := must.Value(want.Spec())

	// --- When ---
	have, err := NoopRuleFromSpec(spc)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, want, have)
}
