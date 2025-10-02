// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_Required(t *testing.T) {
	t.Run("construct", func(t *testing.T) {
		// --- Given ---
		r := Required

		// --- Then ---
		assert.True(t, r.condition)
		assert.False(t, r.skipNil)
		assert.ErrorIs(t, ErrReq, r.err)
	})
}

func Test_NotEmpty(t *testing.T) {
	t.Run("construct", func(t *testing.T) {
		// --- Given ---
		r := NotEmpty

		// --- Then ---
		assert.True(t, r.condition)
		assert.True(t, r.skipNil)
		assert.ErrorIs(t, ErrReqNotEmpty, r.err)
	})
}

func Test_Required_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		value any
	}{
		{"int", 123},
		{"string", iString},
		{"time", iTime},
		{"chan", iChan},
		{"func", iFunc},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Required.Validate(tc.value)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_Required_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		value any
		err   string
		code  string
	}{
		{"nil", nil, "cannot be blank", ECRequired},
		{"empty string", iStringEmpty, "cannot be blank", ECRequired},
		{"declared zero value time", dTime, "cannot be blank", ECRequired},
		{"zero value time", iTimeZero, "cannot be blank", ECRequired},
		{"empty struct", iStructEmpty, "cannot be blank", ECRequired},
		{"pointer to empty struct", pStructEmpty, "cannot be blank", ECRequired},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Required.Validate(tc.value)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_NotEmpty_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		value any
	}{
		{"nil", nil},
		{"string pointer", pString},
		{"string nil pointer", pStringNil},
		{"int pointer", pInt},
		{"int nil pointer", pIntNil},
		{"time pointer", pTime},
		{"time nil pointer", pTimeNil},

		{"empty struct nil pointer", pStructEmptyNil},
		{"int", 123},
		{"struct with fields", iValidate},
		{"any(123)", iInterface},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := NotEmpty.Validate(tc.value)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_NotEmpty_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		value any
		err   error
		code  string
	}{
		{
			"empty string",
			iStringEmpty,
			ErrReqNotEmpty,
			ECReqNotEmpty,
		},
		{
			"pointer to empty string",
			pStringEmpty,
			ErrReqNotEmpty,
			ECReqNotEmpty,
		},
		{
			"zero value int",
			iIntZero,
			ErrReqNotEmpty,
			ECReqNotEmpty,
		},
		{
			"pointer to zero value int",
			pIntZero,
			ErrReqNotEmpty,
			ECReqNotEmpty,
		},
		{
			"zero value time",
			iTimeZero,
			ErrReqNotEmpty,
			ECReqNotEmpty,
		},
		{
			"pointer to zero value time",
			pTimeZero,
			ErrReqNotEmpty,
			ECReqNotEmpty,
		},
		{
			"any(0)",
			iInterfaceZero,
			ErrReqNotEmpty,
			ECReqNotEmpty,
		},
		{
			"pointer to empty struct",
			pStructEmpty,
			ErrReqNotEmpty,
			ECReqNotEmpty,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := NotEmpty.Validate(tc.value)

			// --- Then ---
			assert.ErrorIs(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_requiredRule_When(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		// --- Given ---
		r := NotEmpty.When(false)

		// --- When ---
		err := Validate("", r)

		// --- Then ---
		assert.Nil(t, err)
	})

	t.Run("true", func(t *testing.T) {
		// --- Given ---
		r := NotEmpty.When(true)

		// --- When ---
		err := Validate("", r)

		// --- Then ---
		assert.ErrorIs(t, ErrReqNotEmpty, err)
	})
}

func Test_requiredRule_Code(t *testing.T) {
	t.Run("set custom error code", func(t *testing.T) {
		// --- Given ---
		r := Required

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("")
		assert.ErrorIs(t, ErrReq, err)
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("custom error code for custom error", func(t *testing.T) {
		// --- Given ---
		r := Required.Error(ErrTst)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("")
		assert.Same(t, ErrTst, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})
}

func Test_requiredRule_Error(t *testing.T) {
	t.Run("set custom error", func(t *testing.T) {
		// --- Given ---
		r := Required

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate("")
		assert.Same(t, ErrTst, err)
		assert.ErrorEqual(t, "tst msg", err)
		xrrtest.AssertCode(t, "ETstCode", ErrTst)
	})

	t.Run("clears custom code", func(t *testing.T) {
		// --- Given ---
		r := Required.Code("MyCode")

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate("")
		assert.Same(t, ErrTst, err)
		assert.ErrorEqual(t, "tst msg", err)
		xrrtest.AssertCode(t, "ETstCode", ErrTst)
	})
}
