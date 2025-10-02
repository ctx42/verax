// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_Length(t *testing.T) {
	// --- When ---
	r := Length(2, 3)

	// --- Then ---
	assert.SameType(t, LengthRule{}, r)
	assert.Equal(t, 2, r.min)
	assert.Equal(t, 3, r.max)
	assert.True(t, r.condition)
	assert.False(t, r.rune)
	assert.ErrorEqual(t, "the length must be between 2 and 3", r.err)
	xrrtest.AssertCode(t, ECInvLength, r.err)
}

func Test_RuneLength(t *testing.T) {
	// --- When ---
	r := RuneLength(2, 3)

	// --- Then ---
	assert.SameType(t, LengthRule{}, r)
	assert.Equal(t, 2, r.min)
	assert.Equal(t, 3, r.max)
	assert.True(t, r.condition)
	assert.True(t, r.rune)
	assert.ErrorEqual(t, "the length must be between 2 and 3", r.err)
	xrrtest.AssertCode(t, ECInvLength, r.err)
}

func Test_Length_Validate_valid_tabular(t *testing.T) {
	var v *string

	tt := []struct {
		testN string

		min, max int
		val      any
	}{
		{"zero length", 0, 0, ""},
		{"exact length", 2, 2, "ab"},
		{"zero value string", 2, 4, ""},
		{"within range", 2, 4, "abc"},
		{"3", 0, 4, "ab"},
		{"4", 2, 0, "ab"},
		{"5", 2, 0, v},
		{
			"Valuer within range",
			2,
			4,
			sql.NullString{String: "abc", Valid: true},
		},
		{"Valuer empty", 2, 4, sql.NullString{String: "", Valid: true}},
		{
			"pointer to Valuer within range",
			2,
			4,
			&sql.NullString{String: "abc", Valid: true},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Length(tc.min, tc.max).Validate(tc.val)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_Length_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		min, max int
		val      any
		err      string
		code     string
	}{
		{
			"too long",
			2,
			4,
			"abcdf",
			"the length must be between 2 and 4",
			ECInvLength,
		},
		{
			"too long - min is zero",
			0,
			4,
			"abcde",
			"the length must be no more than 4",
			ECInvLength,
		},
		{
			"too short",
			2,
			4,
			"a",
			"the length must be between 2 and 4",
			ECInvLength,
		},
		{
			"too short - max is zero",
			2,
			0,
			"a",
			"the length must be no less than 2",
			ECInvLength,
		},
		{
			"invalid type",
			2,
			0,
			123,
			"cannot get the length of int",
			ECInvType,
		},
		{
			"not of the exact length",
			2,
			2,
			"abcdf",
			"the length must be exactly 2",
			ECInvLength,
		},
		{
			"must be empty",
			0,
			0,
			"ab",
			"the value must be empty",
			ECReqEmpty,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Length(tc.min, tc.max).Validate(tc.val)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_RuneLength_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		min, max int
		value    any
	}{
		{"zero value string", 2, 4, ""},
		{"within range", 2, 4, "abc"},
		{"within range emoji", 2, 3, "💥💥"},
		{"within range emoji", 2, 3, "💥💥💥"},
		{"within - range min is zero", 0, 4, "ab"},
		{"within - range max is zero", 2, 0, "ab"},
		{"nil pointer to string", 2, 0, pStringNil},
		{"Valuer within range", 2, 4, sql.NullString{String: "abc", Valid: true}},
		{"Valuer zero value", 2, 4, sql.NullString{String: "", Valid: true}},
		{
			"pointer to Valuer within range",
			2,
			4,
			&sql.NullString{String: "abc", Valid: true},
		},
		{
			"pointer to Valuer within range - emoji",
			2,
			3,
			&sql.NullString{String: "💥💥", Valid: true},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := RuneLength(tc.min, tc.max).Validate(tc.value)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_RuneLength_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		min, max int
		value    any
		err      string
		code     string
	}{
		{"1.1", 2, 3, "💥", "the length must be between 2 and 3", ECInvLength},
		{"1.2", 2, 3, "💥💥💥💥", "the length must be between 2 and 3", ECInvLength},
		{"2", 2, 4, "abcdf", "the length must be between 2 and 4", ECInvLength},
		{"3", 0, 4, "abcde", "the length must be no more than 4", ECInvLength},
		{"4", 2, 0, "a", "the length must be no less than 2", ECInvLength},
		{"5", 2, 0, 123, "cannot get the length of int", ECInvType},
		{
			"6",
			2,
			3,
			&sql.NullString{String: "💥", Valid: true},
			"the length must be between 2 and 3",
			ECInvLength,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := RuneLength(tc.min, tc.max).Validate(tc.value)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_LengthRule_When(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		// --- When ---
		r := Length(1, 1).When(false)

		// --- Then ---
		err := Validate("abc", r)
		assert.Nil(t, err)
	})

	t.Run("true", func(t *testing.T) {
		// --- When ---
		r := Length(1, 1).When(true)

		// --- Then ---
		err := Validate("abc", r)
		assert.ErrorEqual(t, "the length must be exactly 1", err)
	})
}

func Test_LengthRule_Code(t *testing.T) {
	t.Run("set custom error code", func(t *testing.T) {
		// --- Given ---
		r := Length(2, 3)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("too_long")
		assert.ErrorEqual(t, "the length must be between 2 and 3", err)
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("custom error code for custom error", func(t *testing.T) {
		// --- Given ---
		r := Length(2, 3).Error(ErrTst)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("too_long")
		assert.Same(t, ErrTst, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})
}

func Test_LengthRule_Error(t *testing.T) {
	t.Run("set custom error", func(t *testing.T) {
		// --- Given ---
		r := Length(2, 3)

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate("too_long")
		assert.Same(t, ErrTst, err)
		xrrtest.AssertCode(t, "ETstCode", err)
	})

	t.Run("clears custom error code", func(t *testing.T) {
		// --- Given ---
		r := Length(2, 3).Code("MyCode")

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate("too_long")
		assert.Same(t, ErrTst, err)
		xrrtest.AssertCode(t, "ETstCode", err)
	})
}
