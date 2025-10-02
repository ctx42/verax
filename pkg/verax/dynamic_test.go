// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_DynamicRule_Validate(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("pkt", "Fn")

		// --- When ---
		err := Validate("abc", r)

		// --- Then ---
		assert.ErrorIs(t, ErrInvDynamic, err)
	})

	t.Run("valid", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("pkt", "Fn").RuleFunc(StrRuleFunc("abc"))

		// --- When ---
		err := Validate("abc", r)

		// --- Then ---
		assert.Nil(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("pkt", "Fn").RuleFunc(StrRuleFunc("abc"))

		// --- When ---
		err := Validate("xyz", r)

		// --- Then ---
		assert.ErrorEqual(t, "must be 'abc'", err)
		xrrtest.AssertCode(t, "ECMustAbc", err)
	})

	t.Run("custom error", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("", "").RuleFunc(StrRuleFunc("abc"))

		// --- When ---
		have := r.Error(xrr.New("custom error", "ECOther"))

		// --- Then ---
		err := have.Validate("xyz")
		assert.ErrorEqual(t, "custom error", err)
		xrrtest.AssertCode(t, "ECOther", err)
	})

	t.Run("reference", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("pkt", "Fn").RuleFunc(StrRuleFunc("abc"))

		// --- When ---
		have := r.Reference()

		// --- Then ---
		assert.Equal(t, "pkt.Fn", have)
	})

	t.Run("empty reference", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("", "").RuleFunc(StrRuleFunc("abc"))

		// --- When ---
		have := r.Reference()

		// --- Then ---
		assert.Equal(t, ".", have)
	})

	t.Run("nil RuleFunc", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("pkt", "Fn").RuleFunc(nil)

		// --- When ---
		err := Validate("abc", r)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSetup, err)
		xrrtest.AssertCode(t, ECInternal, err)
	})

	t.Run("nil RuleFunc with custom code", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("pkt", "Fn").Code("ECMyCode").RuleFunc(nil)

		// --- When ---
		err := Validate("abc", r)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSetup, err)
		xrrtest.AssertCode(t, ECInternal, err)
	})
}

func Test_DynamicRule_When(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("", "").RuleFunc(StrRuleFunc("abc"))

		// --- When ---
		r = r.When(false)

		// --- Then ---
		err := Validate("xyz", r)
		assert.Nil(t, err)
	})

	t.Run("true", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("", "").RuleFunc(StrRuleFunc("abc"))

		// --- When ---
		r = r.When(true)

		// --- Then ---
		err := Validate("xyz", r)
		assert.ErrorEqual(t, "must be 'abc'", err)
	})
}

func Test_DynamicRule_Code(t *testing.T) {
	t.Run("set custom", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("", "").RuleFunc(StrRuleFunc("abc"))

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("xyz")
		assert.ErrorEqual(t, "must be 'abc'", err)
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("with custom error core error", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("", "").
			RuleFunc(StrRuleFunc("abc")).
			Error(errors.New("custom error"))

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("xyz")
		assert.ErrorEqual(t, "custom error", err)
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("with custom error vrr error", func(t *testing.T) {
		// --- Given ---
		r := Dynamic("", "").
			RuleFunc(StrRuleFunc("abc")).
			Error(xrr.New("custom error", "ECOther"))

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("xyz")
		assert.ErrorEqual(t, "custom error", err)
		xrrtest.AssertCode(t, "MyCode", err)
	})
}
