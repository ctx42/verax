// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_When(t *testing.T) {
	t.Run("condition is true", func(t *testing.T) {
		// --- Given ---
		have := When(true)

		// --- Then ---
		want := WhenRule{
			condition: true,
			rules:     nil,
			elseRules: nil,
			msg:       "",
			code:      "",
			flags:     0,
		}
		assert.Equal(t, want, have)
	})

	t.Run("condition false", func(t *testing.T) {
		// --- Given ---
		have := When(false)

		// --- Then ---
		want := WhenRule{
			condition: false,
			rules:     nil,
			elseRules: nil,
			msg:       "",
			code:      "",
			flags:     0,
		}
		assert.Equal(t, want, have)
	})
}

func Test_WhenRule_Validate_valid_tabular(t *testing.T) {
	abcRule := Equal("abc").Message("err_abc").Code("ECAbc")
	xyzRule := Equal("xyz").Message("err_xyz").Code("ECXyz")

	tt := []struct {
		testN string

		check     bool
		value     any
		rules     []Rule
		elseRules []Rule
	}{
		// True condition.
		{
			"condition is true - nil - no rules provided",
			true,
			nil,
			[]Rule{},
			[]Rule{},
		},
		{
			"condition is true - empty string - no rules provided",
			true,
			"",
			[]Rule{},
			[]Rule{},
		},
		{
			"condition is true - only when rules are evaluated",
			true,
			"abc",
			[]Rule{abcRule},
			[]Rule{xyzRule},
		},
		{
			"condition is true - multiple when rules pass",
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
			"condition is false - empty string - no rules provided",
			false,
			"",
			[]Rule{},
			[]Rule{},
		},
		{
			"condition is false - only else rules are evaluated",
			false,
			"xyz",
			[]Rule{abcRule},
			[]Rule{xyzRule},
		},
		{
			"condition is false - multiple else rules are evaluated",
			false,
			"xyz",
			[]Rule{abcRule},
			[]Rule{xyzRule, Length(3, 3)},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := When(tc.check, tc.rules...).Else(tc.elseRules...)

			// --- When ---
			err := r.Validate(tc.value)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_WhenRule_Validate_invalid_tabular(t *testing.T) {
	abcRule := Equal("abc").Message("err_abc").Code("ECAbc")
	xyzRule := Equal("xyz").Message("err_xyz").Code("ECXyz")

	tt := []struct {
		testN string

		check     bool
		have      any
		rules     []Rule
		elseRules []Rule
		err       string
		code      string
	}{
		// True condition.
		{
			"condition is true - nil - when rule error",
			true,
			nil,
			[]Rule{Required},
			[]Rule{},
			"cannot be blank",
			ECReq,
		},
		{
			"condition is true - empty string - when rule error",
			true,
			"",
			[]Rule{Required},
			[]Rule{},
			"cannot be blank",
			ECReq,
		},
		{
			"condition is true - when rule error",
			true,
			"xyz",
			[]Rule{abcRule},
			[]Rule{},
			"err_abc",
			"ECAbc",
		},
		{
			"condition is true - one of when rules error",
			true,
			"abc",
			[]Rule{abcRule, xyzRule},
			[]Rule{},
			"err_xyz",
			"ECXyz",
		},

		// False condition.
		{
			"condition is false - only else rule is evaluated",
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
			r := When(tc.check, tc.rules...).Else(tc.elseRules...)

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.SameType(t, &Error{}, err)
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_WhenRule_Message(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := WhenRule{}

		// --- When ---
		have := r.Message("test err")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Equal(t, flgCustomMsg, have.flags)
	})

	t.Run("an empty string is a noop", func(t *testing.T) {
		// --- Given ---
		r := WhenRule{msg: "test err"}

		// --- When ---
		have := r.Message("")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Zero(t, have.flags)
	})
}

func Test_WhenRule_Code(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := WhenRule{}

		// --- When ---
		have := r.Code("ECTst")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Equal(t, flgCustomCode, have.flags)
	})

	t.Run("an empty string is noop", func(t *testing.T) {
		// --- Given ---
		r := WhenRule{code: "ECTst"}

		// --- When ---
		have := r.Code("")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Zero(t, have.flags)
	})
}
