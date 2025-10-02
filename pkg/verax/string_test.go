// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_String(t *testing.T) {
	// --- Given ---
	check := checkString(iString)

	// --- When ---
	r := String(check)

	// --- Then ---
	assert.Same(t, check, r.fn)
	assert.True(t, r.condition)
	assert.Same(t, ErrNotEqual, r.err)
}

func Test_StringRule_Validate(t *testing.T) {
	t.Run("error - not equal", func(t *testing.T) {
		// --- Given ---
		r := String(checkString(iString))

		// --- When ---
		err := r.Validate("invalid")

		// --- Then ---
		assert.Same(t, ErrNotEqual, err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- Given ---
		check := checkString(iString)
		r := String(check)

		// --- When ---
		err := r.Validate(100)

		// --- Then ---
		assert.ErrorEqual(t, "must be either a string or byte slice", err)
		xrrtest.AssertCode(t, ECInvType, err)
	})

	t.Run("error - invalid type with custom code", func(t *testing.T) {
		// --- Given ---
		check := checkString(iString)
		r := String(check).Code("MyCode")

		// --- When ---
		err := r.Validate(100)

		// --- Then ---
		assert.ErrorEqual(t, "must be either a string or byte slice", err)
		xrrtest.AssertCode(t, ECInvType, err)
	})
}

func Test_StringRule_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val any
	}{
		{"nil", nil},
		{"empty", ""},
		{"string", iString},
		{"string pointer", pString},
		{"string pointer to nil", pStringNil},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			check := checkString(iString)
			r := String(check)

			// --- When ---
			err := r.Validate(tc.val)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_StringRule_When(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		// --- When ---
		r := String(checkString("abc")).When(false)

		// --- Then ---
		err := Validate("xyz", r)
		assert.Nil(t, err)
	})

	t.Run("true", func(t *testing.T) {
		// --- When ---
		r := String(checkString("abc")).When(true)

		// --- Then ---
		err := Validate("xyz", r)
		assert.ErrorIs(t, ErrNotEqual, err)
	})
}

func Test_StringRule_Code(t *testing.T) {
	t.Run("set custom error code", func(t *testing.T) {
		// --- Given ---
		r := String(checkString(iString))

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("invalid")
		assert.ErrorIs(t, ErrNotEqual, err)
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("the last error code is used", func(t *testing.T) {
		// --- Given ---
		r := String(checkString(iString)).Code("MyCode")

		// --- When ---
		have := r.Code("MyCode1")

		// --- Then ---
		err := have.Validate("invalid")
		assert.ErrorIs(t, ErrNotEqual, err)
		xrrtest.AssertCode(t, "MyCode1", err)
	})

	t.Run("custom error code for custom error", func(t *testing.T) {
		// --- Given ---
		check := checkString(iString)
		r := String(check).Error(ErrTst)

		// --- When ---
		have := r.Code("MyCode1")

		// --- Then ---
		err := have.Validate("invalid")
		assert.Same(t, ErrTst, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode1", err)
	})
}

func Test_StringRule_Error(t *testing.T) {
	t.Run("set custom error", func(t *testing.T) {
		// --- Given ---
		check := checkString(iString)
		r := String(check)

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate("invalid")
		assert.Same(t, ErrTst, err)
	})

	t.Run("clears custom error code", func(t *testing.T) {
		// --- Given ---
		check := checkString(iString)
		r := String(check).Code("MyCode")

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate("invalid")
		assert.Same(t, ErrTst, err)
	})
}
