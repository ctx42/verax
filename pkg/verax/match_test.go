// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"regexp"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_Match(t *testing.T) {
	// --- Given ---
	re := regexp.MustCompile(`\d+`)

	// --- When ---
	r := Match(re)

	// --- Then ---
	assert.Same(t, re, r.rx)
	assert.True(t, r.condition)
	assert.Same(t, ErrInvMatch, r.err)
}

func Test_MatchRule_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		re    string
		value any
	}{
		{"nil", "[a-z]+", nil},
		{"zero value string", "[a-z]+", ""},
		{"string", "[a-z]+", "abc"},
		{"nil pointer to string", "[a-z]+", pStringNil},
		{"byte slice", "[a-z]+", []byte("abc")},
		{"empty byte slice", "[a-z]+", []byte{}},
		{"byte slice from empty string", "[a-z]+", []byte("")},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := Match(regexp.MustCompile(tc.re))

			// --- When ---
			err := r.Validate(tc.value)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_MatchRule_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		re    string
		value any
		err   error
		code  string
	}{
		{"string", "[a-z]+", "123", ErrInvMatch, ECInvMatch},
		{"byte slice", "[a-z]+", []byte("123"), ErrInvMatch, ECInvMatch},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := Match(regexp.MustCompile(tc.re))

			// --- When ---
			err := r.Validate(tc.value)

			// --- Then ---
			assert.ErrorIs(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_MatchRule_When(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		// --- When ---
		r := Match(regexp.MustCompile(`\d+`)).When(false)

		// --- Then ---
		err := Validate("abc", r)
		assert.Nil(t, err)
	})

	t.Run("true", func(t *testing.T) {
		// --- Given ---
		r := Match(regexp.MustCompile(`\d+`)).When(true)

		// --- When ---
		err := Validate("abc", r)

		// --- Then ---
		assert.ErrorIs(t, ErrInvMatch, err)
	})
}

func Test_MatchRule_Code(t *testing.T) {
	t.Run("set custom error code", func(t *testing.T) {
		// --- Given ---
		r := Match(regexp.MustCompile(`\d+`))

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("abc")
		assert.Same(t, ErrInvMatch, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("custom error code for custom error", func(t *testing.T) {
		// --- Given ---
		r := Match(regexp.MustCompile(`\d+`)).Error(ErrTst)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("abc")
		assert.Same(t, ErrTst, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})
}

func Test_MatchRule_Error(t *testing.T) {
	t.Run("set custom error", func(t *testing.T) {
		// --- Given ---
		r := Match(regexp.MustCompile(`\d+`))

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate("abc")
		assert.Same(t, ErrTst, err)
		xrrtest.AssertCode(t, "ETstCode", ErrTst)
	})

	t.Run("clears custom error code", func(t *testing.T) {
		// --- Given ---
		r := Match(regexp.MustCompile(`\d+`)).Code("MyCode")

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate("abc")
		assert.Same(t, ErrTst, err)
		xrrtest.AssertCode(t, "ETstCode", ErrTst)
	})

	t.Run("sys error nil given", func(t *testing.T) {
		// --- Given ---
		r := Match(nil)

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.ErrorIs(t, ErrInvSetup, err)
		xrrtest.AssertCode(t, ECInternal, err)
	})

	t.Run("sys error nil given with custom code", func(t *testing.T) {
		// --- Given ---
		r := Match(nil).Code("ECode")

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.ErrorIs(t, ErrInvSetup, err)
		xrrtest.AssertCode(t, ECInternal, err)
	})
}
