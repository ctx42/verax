// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_NotNil(t *testing.T) {
	t.Run("construct", func(t *testing.T) {
		// --- When ---
		r := NotNil

		// --- Then ---
		assert.True(t, r.condition)
		assert.Same(t, r.err, ErrReqNotNil)
	})
}

func Test_notNilRule_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val any
	}{
		{"empty string", ""},
		{"zero int", 0},
		{"pointer to string", pString},
		{"pointer to empty string", pStringEmpty},
		{"pointer to int", pInt},
		{"pointer to zero value int", pIntZero},
		{"pointer to time", pTime},
		{"pointer to zero value time", pTimeZero},
		{"pointer to empty struct", pStructEmpty},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := NotNil.Validate(tc.val)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_notNilRule_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val  any
		err  error
		code string
	}{
		{"nil slice", dSlice, ErrReqNotNil, ECReqNotNil},
		{"zero value array", dMap, ErrReqNotNil, ECReqNotNil},
		{"nil pointer to string", pStringNil, ErrReqNotNil, ECReqNotNil},
		{"nil pointer to int", pIntNil, ErrReqNotNil, ECReqNotNil},
		{"nil empty interface", dInterface, ErrReqNotNil, ECReqNotNil},
		{"nil interface", dValidate, ErrReqNotNil, ECReqNotNil}}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := NotNil.Validate(tc.val)

			// --- Then ---
			assert.ErrorIs(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_notNilRule_When(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		// --- When ---
		r := NotNil.When(false)

		// --- Then ---
		err := Validate(nil, r)
		assert.Nil(t, err)
	})

	t.Run("true", func(t *testing.T) {

		// --- When ---
		r := NotNil.When(true)

		// --- Then ---
		err := Validate(nil, r)
		assert.ErrorIs(t, ErrReqNotNil, err)
	})
}

func Test_notNilRule_Code(t *testing.T) {
	t.Run("set custom error code", func(t *testing.T) {
		// --- Given ---
		r := NotNil

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate(nil)
		assert.Same(t, ErrReqNotNil, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("custom error code for custom error", func(t *testing.T) {
		// --- Given ---
		r := NotNil.Error(ErrTst)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate(nil)
		assert.Same(t, ErrTst, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})
}

func Test_notNilRule_Error(t *testing.T) {
	t.Run("set custom error", func(t *testing.T) {
		// --- Given ---

		// --- When ---
		r := NotNil.Error(ErrTst)

		// --- Then ---
		err := r.Validate(nil)
		assert.Same(t, ErrTst, err)
		xrrtest.AssertCode(t, "ETstCode", ErrTst)
	})

	t.Run("clears custom code", func(t *testing.T) {
		// --- Given ---
		r := NotNil.Code("MyCode")

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate(nil)
		assert.ErrorIs(t, ErrTst, err)
		xrrtest.AssertCode(t, "ETstCode", err)
	})
}
