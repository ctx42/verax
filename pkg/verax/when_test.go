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

func Test_When(t *testing.T) {
	t.Run("condition true", func(t *testing.T) {
		// --- Given ---
		r := When(true)

		// --- Then ---
		assert.True(t, r.condition)
		assert.Empty(t, r.rules)
		assert.Empty(t, r.elseRules)
	})

	t.Run("condition false", func(t *testing.T) {
		// --- Given ---
		r := When(false)

		// --- Then ---
		assert.False(t, r.condition)
		assert.Empty(t, r.rules)
		assert.Empty(t, r.elseRules)
	})
}

func Test_WhenRule_Validate_valid_tabular(t *testing.T) {
	abcValidation := func(val string) bool { return val == "abc" }
	abcRule := String(abcValidation)

	xyzValidation := func(val string) bool { return val == "xyz" }
	xyzRule := String(xyzValidation)

	tt := []struct {
		testN string

		condition bool
		value     any
		rules     []Rule
		elseRules []Rule
	}{
		// True condition.
		{
			"condition true - nil - no rules provided",
			true,
			nil,
			[]Rule{},
			[]Rule{},
		},
		{
			"condition true - empty string - no rules provided",
			true,
			"",
			[]Rule{},
			[]Rule{},
		},
		{
			"condition true - only when rules are evaluated",
			true,
			"abc",
			[]Rule{abcRule},
			[]Rule{xyzRule},
		},
		{
			"condition true - multiple when rules pass",
			true,
			"abc",
			[]Rule{abcRule, Length(3, 3)},
			[]Rule{},
		},

		// False condition.
		{
			"condition false - nil - no rules provided",
			false,
			nil,
			[]Rule{},
			[]Rule{},
		},
		{
			"condition false - empty string - no rules provided",
			false,
			"",
			[]Rule{},
			[]Rule{},
		},
		{
			"condition false - only else rules are evaluated",
			false,
			"xyz",
			[]Rule{abcRule},
			[]Rule{xyzRule},
		},
		{
			"condition false - multiple else rules are evaluated",
			false,
			"xyz",
			[]Rule{abcRule},
			[]Rule{xyzRule, Length(3, 3)},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := When(tc.condition, tc.rules...).Else(tc.elseRules...)

			// --- When ---
			err := Validate(tc.value, r)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_WhenRule_Validate_invalid_tabular(t *testing.T) {
	abcValidation := func(val string) bool { return val == "abc" }
	abcRule := String(abcValidation).Error(errors.New("err_abc")).Code("ECAbc")

	xyzValidation := func(val string) bool { return val == "xyz" }
	xyzRule := String(xyzValidation).Error(errors.New("err_xyz")).Code("ECXyz")

	tt := []struct {
		testN string

		condition bool
		value     any
		rules     []Rule
		elseRules []Rule
		err       string
		code      string
	}{
		// True condition.
		{
			"condition true - nil - when rule error",
			true,
			nil,
			[]Rule{Required},
			[]Rule{},
			"cannot be blank",
			ECRequired,
		},
		{
			"condition true - empty string - when rule error",
			true,
			"",
			[]Rule{Required},
			[]Rule{},
			"cannot be blank",
			ECRequired,
		},
		{
			"condition true - when rule error",
			true,
			"xyz",
			[]Rule{abcRule},
			[]Rule{},
			"err_abc",
			"ECAbc",
		},
		{
			"condition true - one of when rules error",
			true,
			"abc",
			[]Rule{abcRule, xyzRule},
			[]Rule{},
			"err_xyz",
			"ECXyz",
		},

		// False condition.
		{
			"condition false - only else rule is evaluated",
			false,
			"abc",
			[]Rule{abcRule},
			[]Rule{xyzRule},
			"err_xyz",
			"ECXyz",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := When(tc.condition, tc.rules...).Else(tc.elseRules...)

			// --- When ---
			err := Validate(tc.value, r)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_WhenRule_Code(t *testing.T) {
	t.Run("with custom code", func(t *testing.T) {
		// --- Given ---
		r := When(true, Required).Code("ECode")

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.ErrorIs(t, ErrReq, err)
		xrrtest.AssertCode(t, "ECode", err)
	})
}

func Test_WhenRule_Error(t *testing.T) {
	t.Run("when rules", func(t *testing.T) {
		// --- Given ---
		err := xrr.New("test msg", "ECode")

		// --- When ---
		have := When(true, In("abc")).
			Else(In("xyz")).
			Error(err).
			Validate("xyz")

		// --- Then ---
		assert.ErrorIs(t, err, have)
	})

	t.Run("else rules", func(t *testing.T) {
		// --- Given ---
		err := xrr.New("test msg", "ECode")

		// --- When ---
		have := When(false, In("abc")).
			Else(In("xyz")).
			Error(err).
			Validate("abc")

		// --- Then ---
		assert.ErrorIs(t, err, have)
	})
}
