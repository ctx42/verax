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

func Test_Equal(t *testing.T) {
	t.Run("error - not equal", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.ErrorEqual(t, "must be equal to '42'", err)
		xrrtest.AssertCode(t, ECNotEqual, err)
	})

	t.Run("error - not equal time", func(t *testing.T) {
		// --- Given ---
		r := Equal(time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC))

		// --- When ---
		err := r.Validate(time.Date(2001, 1, 2, 3, 4, 5, 0, time.UTC))

		// --- Then ---
		assert.ErrorEqual(t, "must be equal to '2000-01-02T03:04:05Z'", err)
		xrrtest.AssertCode(t, ECNotEqual, err)
	})
}

func Test_NotEqual(t *testing.T) {
	t.Run("error - not equal", func(t *testing.T) {
		// --- Given ---
		r := NotEqual(42)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.ErrorEqual(t, "must not be equal to '42'", err)
		xrrtest.AssertCode(t, ECEqual, err)
	})

	t.Run("error - equal time", func(t *testing.T) {
		// --- Given ---
		r := NotEqual(time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC))

		// --- When ---
		err := r.Validate(time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC))

		// --- Then ---
		assert.ErrorEqual(t, "must not be equal to '2000-01-02T03:04:05Z'", err)
		xrrtest.AssertCode(t, ECEqual, err)
	})
}

func Test_EqualField(t *testing.T) {
	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		r := EqualField(1, "field_name").Code("ECode")

		// --- When ---
		err := r.Validate(2)

		// --- Then ---
		assert.ErrorEqual(t, "must be equal to 'field_name'", err)
		xrrtest.AssertCode(t, "ECode", err)
	})
}

func Test_NotEqualField(t *testing.T) {
	t.Run("custom code", func(t *testing.T) {
		// --- Given ---
		r := NotEqualField(1, "field_name").Code("ECode")

		// --- When ---
		err := r.Validate(1)

		// --- Then ---
		assert.ErrorEqual(t, "must not be equal to 'field_name'", err)
		xrrtest.AssertCode(t, "ECode", err)
	})
}

func Test_EqualBy(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		// --- Given ---
		fn := func(want, have any) bool { return want.(int) == have.(int) }
		r := EqualBy(42, fn)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("not equal", func(t *testing.T) {
		// --- Given ---
		fn := func(want, have any) bool { return want.(int) == have.(int) }
		r := EqualBy(42, fn)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.ErrorEqual(t, "must be equal to '42'", err)
		xrrtest.AssertCode(t, ECEqual, err)
	})
}

func Test_EqualRule_Validate(t *testing.T) {
	t.Run("no error", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("validation error", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.ErrorEqual(t, "must be equal to '42'", err)
		xrrtest.AssertCode(t, ECNotEqual, err)
	})

	t.Run("no error when condition false", func(t *testing.T) {
		// --- Given ---
		r := Equal(42).When(false)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.NoError(t, err)
	})
}

func Test_Equal_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		exp any
		got any
	}{
		{"nil", nil, nil},
		{"int", 1, 1},
		{"string", "abc", "abc"},
		{"empty string", "", ""},
		{"float", 1.23, 1.23},
		{
			"time",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := Equal(tc.exp)

			// --- When ---
			err := r.Validate(tc.got)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_Equal_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		exp     any
		got     any
		setCode string
		err     string
	}{
		{"int", 1, 2, "ECode1", "must be equal to '1'"},
		{"string", "abc", "abb", "ECode2", "must be equal to 'abc'"},
		{"float", 1.23, 1.24, "ECode3", "must be equal to '1.23'"},
		{
			"time",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			time.Date(2001, 1, 2, 3, 4, 5, 0, time.UTC),
			"ECode4",
			"must be equal to '2000-01-02T03:04:05Z'",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := Equal(tc.exp).Code(tc.setCode)

			// --- When ---
			err := r.Validate(tc.got)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.setCode, err)
		})
	}
}

func Test_EqualField_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		exp any
		got any
	}{
		{
			"time",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
		},
		{"int", 1, 1},
		{"bool both true", true, true},
		{"bool both false", false, false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := EqualField(tc.exp, "field_name")

			// --- When ---
			err := r.Validate(tc.got)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_EqualField_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		exp  any
		got  any
		err  string
		code string
	}{
		{
			"time",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			time.Date(2001, 1, 2, 3, 4, 5, 0, time.UTC),
			"must be equal to 'field_name'",
			ECNotEqual,
		},
		{
			"int",
			1,
			2,
			"must be equal to 'field_name'",
			ECNotEqual,
		},
		{
			"bool true",
			true,
			false,
			"must be equal to 'field_name'",
			ECNotEqual,
		},
		{
			"bool false",
			false,
			true,
			"must be equal to 'field_name'",
			ECNotEqual,
		},
		{
			"string",
			"",
			"a",
			"must be equal to 'field_name'",
			ECNotEqual,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := EqualField(tc.exp, "field_name")

			// --- When ---
			err := r.Validate(tc.got)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_NotEqualField_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		not any
		got any
	}{
		{
			"time",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			time.Date(2001, 1, 2, 3, 4, 5, 0, time.UTC),
		},
		{"int", 1, 2},
		{"bool", true, false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := NotEqualField(tc.not, "field_name")

			// --- When ---
			err := r.Validate(tc.got)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_NotEqualField_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		not  any
		got  any
		err  string
		code string
	}{
		{
			"time",
			time.Date(2001, 1, 2, 3, 4, 5, 0, time.UTC),
			time.Date(2001, 1, 2, 3, 4, 5, 0, time.UTC),
			"must not be equal to 'field_name'",
			ECEqual,
		},
		{"int", 1, 1, "must not be equal to 'field_name'", ECEqual},
		{"bool", true, true, "must not be equal to 'field_name'", ECEqual},
		{"empty string", "", "", "must not be equal to 'field_name'", ECEqual},
		{"nil", nil, nil, "must not be equal to 'field_name'", ECEqual},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := NotEqualField(tc.not, "field_name")

			// --- When ---
			err := r.Validate(tc.got)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_EqualRule_When(t *testing.T) {
	// --- Given ---
	r := Equal(42)

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_EqualRule_Code(t *testing.T) {
	t.Run("set custom error code", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate(44)
		assert.ErrorEqual(t, "must be equal to '42'", err)
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("custom error code for custom error", func(t *testing.T) {
		// --- Given ---
		r := Equal(42).Error(ErrTst)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate(44)
		assert.Same(t, ErrTst, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})
}

func Test_EqualRule_Error(t *testing.T) {
	t.Run("set custom error", func(t *testing.T) {
		// --- Given ---
		r := Equal(42)

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate(44)
		assert.Same(t, ErrTst, err)
		xrrtest.AssertCode(t, "ETstCode", err)
	})

	t.Run("clears custom error code", func(t *testing.T) {
		// --- Given ---
		r := Equal(42).Code("ECCustom")

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate(44)
		assert.Same(t, ErrTst, err)
		xrrtest.AssertCode(t, "ETstCode", err)
	})
}
