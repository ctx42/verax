// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_Nil(t *testing.T) {
	assert.ErrorIs(t, ErrReqNil, Nil.err)
	assert.True(t, Nil.condition)
	assert.False(t, Nil.skipNil)
}

func Test_Empty(t *testing.T) {
	assert.ErrorIs(t, ErrReqEmpty, Empty.err)
	assert.True(t, Empty.condition)
	assert.True(t, Empty.skipNil)
}

func Test_Nil_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val any
	}{
		{"nil", nil},
		{"nil pointer to string", pStringNil},
		{"nil pointer to int", pIntNil},
		{"nil pointer to time", pTimeNil},
		{"nil pointer to empty struct", pStructEmptyNil},
		{"nil declared slice", dSlice},
		{"nil declared map", dMap},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Nil.Validate(tc.val)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_Nil_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		value any
		err   error
		code  string
	}{
		{"int", 123, ErrReqNil, ECReqNil},
		{"string instance", iString, ErrReqNil, ECReqNil},
		{"empty string", "", ErrReqNil, ECReqNil},
		{"declared string", dString, ErrReqNil, ECReqNil},
		{"time", time.Now(), ErrReqNil, ECReqNil},

		{"time", iTime, ErrReqNil, ECReqNil},
		{"declared time", dTime, ErrReqNil, ECReqNil},
		{"zero value time", time.Time{}, ErrReqNil, ECReqNil},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Nil.Validate(tc.value)

			// --- Then ---
			assert.ErrorIs(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_Empty_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		value any
	}{
		{"nil", nil},
		{"empty string", ""},
		{"empty string instance", iStringEmpty},
		{"nil pointer to string", pStringNil},
		{"declared int", dInt},
		{"zero value time", iTimeZero},
		{"empty struct", iStructEmpty},
		{"declared empty struct", dStructEmpty},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			assert.NoError(t, Empty.Validate(tc.value))
		})
	}
}

func Test_Empty_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		value any
		err   error
		code  string
	}{
		{"string", "abc", ErrReqEmpty, ECReqEmpty},
		{"string instance", iString, ErrReqEmpty, ECReqEmpty},
		{"pointer to string instance", &iString, ErrReqEmpty, ECReqEmpty},
		{"int", 123, ErrReqEmpty, ECReqEmpty},
		{"time instance", time.Now(), ErrReqEmpty, ECReqEmpty},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Empty.Validate(tc.value)

			// --- Then ---
			assert.ErrorIs(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_absentRule_When(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		// --- Given ---
		r := Nil

		// --- When ---
		have := r.When(false)

		// --- Then ---
		err := Validate(nil, have)
		assert.Nil(t, err)
	})

	t.Run("true", func(t *testing.T) {
		// --- Given ---
		r := Nil

		// --- When ---
		have := r.When(true)

		// --- Then ---
		err := Validate(42, have)
		assert.ErrorIs(t, ErrReqNil, err)
	})
}

func Test_absentRule_Code(t *testing.T) {
	t.Run("set custom error code", func(t *testing.T) {
		// --- Given ---
		r := Empty

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("not_empty")
		assert.Error(t, err)
		assert.Same(t, ErrReqEmpty, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("custom error code for custom error", func(t *testing.T) {
		// --- Given ---
		r := Empty.Error(ErrTst)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("not_empty")
		assert.Same(t, ErrTst, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})
}

func Test_absentRule_Error(t *testing.T) {
	t.Run("set custom error", func(t *testing.T) {
		// --- Given ---
		r := Empty

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate("not_empty")
		assert.Same(t, ErrTst, err)
	})

	t.Run("clears custom error code", func(t *testing.T) {
		// --- Given ---
		r := Empty.Code("MyCode")

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate("not_empty")
		assert.Same(t, ErrTst, err)
	})
}
