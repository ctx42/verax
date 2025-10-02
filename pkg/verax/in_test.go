// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_In(t *testing.T) {
	// --- When ---
	have := In("a", "b", "c")

	// --- Then ---
	assert.Equal(t, []any{"a", "b", "c"}, have.elements)
	assert.True(t, have.condition)
	assert.True(t, have.in)
	assert.Same(t, ErrNotIn, have.err)
}

func Test_NotIn(t *testing.T) {
	// --- When ---
	have := NotIn("a", "b", "c")

	// --- Then ---
	assert.Equal(t, []any{"a", "b", "c"}, have.elements)
	assert.True(t, have.condition)
	assert.False(t, have.in)
	assert.Same(t, ErrIn, have.err)
}

func Test_In_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		values []any
		value  any
	}{
		{"int slice - zero value", []any{1, 2}, 0},
		{"int slice - nil pointer to int", []any{1, 2}, pIntNil},
		{"int slice - zero index", []any{1, 2}, 1},
		{"int slice - non-zero index", []any{1, 2}, 2},
		{"int slice - pointer to value", []any{1, 123}, &iInt},
		{"string slice", []any{"a", "b"}, "b"},
		{"string slice - zero value", []any{"a", "b"}, ""},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := In(tc.values...)

			// --- When ---
			err := r.Validate(tc.value)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_In_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		values []any
		value  any
		err    error
		code   string
	}{
		{"int slice", []any{1, 2}, 3, ErrNotIn, ECInvIn},
		{"empty int slice", []any{}, 3, ErrNotIn, ECInvIn},
		{"invalid type", []any{1, 2}, "1", ErrInvType, ECInvIn},
		{
			"mixed type slice",
			[]any{1, []byte{1}, 2},
			[]byte{1},
			ErrInvType,
			ECInvIn,
		},
		{"string slice", []any{"a", "b"}, "c", ErrNotIn, ECInvIn},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := In(tc.values...)

			// --- When ---
			err := r.Validate(tc.value)

			// --- Then ---
			assert.ErrorIs(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_NotIn_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		values []any
		value  any
	}{
		{"any slice - zero value", []any{1, 2}, 0},
		{"any slice - value does not exist", []any{1, 2}, 3},
		{"any slice - pointer to value", []any{1, 2}, &iInt},
		{"any slice - nil pointer to value", []any{1, 2}, pIntNil},
		{"empty any slice", []any{}, 2},
		{"string slice", []any{"a", "b"}, "c"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := NotIn(tc.values...)

			// --- When ---
			err := r.Validate(tc.value)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_NotIn_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		values []any
		value  any
		err    error
		code   string
	}{
		{"any slice", []any{1, 2}, 2, ErrIn, ECInvIn},
		{"wrong type", []any{1, 2}, "1", ErrInvType, ECInvIn},
		{
			"mixed type slice",
			[]any{[]byte{1}, 1, 2},
			[]byte{3},
			ErrInvType,
			ECInvIn,
		},
		{"string slice", []any{"a", "b"}, "a", ErrIn, ECInvIn},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := NotIn(tc.values...)

			// --- When ---
			err := r.Validate(tc.value)

			// --- Then ---
			assert.ErrorIs(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_InRule_When(t *testing.T) {
	t.Run("condition true", func(t *testing.T) {
		// --- Given ---
		r := In(41, 42, 43).When(true)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.ErrorIs(t, ErrNotIn, err)
		xrrtest.AssertCode(t, ECInvIn, err)
	})

	t.Run("condition false", func(t *testing.T) {
		// --- Given ---
		r := In(41, 42, 43).When(false)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.NoError(t, err)
	})
}

func Test_InRule_Code(t *testing.T) {
	t.Run("set custom code", func(t *testing.T) {
		// --- Given ---
		r := In(41, 42, 43)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate(44)
		assert.ErrorIs(t, ErrNotIn, err)
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("custom error code for custom error", func(t *testing.T) {
		// --- Given ---
		r := In(41, 42, 43).Error(ErrTst)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate(44)
		assert.ErrorIs(t, ErrTst, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})
}

func Test_InRule_Error(t *testing.T) {
	t.Run("set custom error", func(t *testing.T) {
		// --- Given ---
		r := In(41, 42, 43)

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate(44)
		assert.Same(t, ErrTst, err)
		assert.ErrorEqual(t, "tst msg", err)
		xrrtest.AssertCode(t, "ETstCode", ErrTst)
	})

	t.Run("clears custom error code", func(t *testing.T) {
		// --- Given ---
		r := In(41, 42, 43).Code("MyCode")

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate(44)
		assert.Same(t, ErrTst, err)
		assert.ErrorEqual(t, "tst msg", err)
		xrrtest.AssertCode(t, "ETstCode", ErrTst)
	})
}
