// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/xrr/pkg/xrr"
)

func Test_By(t *testing.T) {
	// --- Given ---
	fn := func(v any) error { return fmt.Errorf("error: %s", v) }

	// --- When ---
	r := By(fn)

	// --- Then ---
	assert.Same(t, fn, r.fn)
	assert.True(t, r.condition)
	assert.Nil(t, r.err)
	assert.Empty(t, r.code)
}

func Test_ByRule_Validate(t *testing.T) {
	t.Run("passes the argument to function", func(t *testing.T) {
		// --- Given ---
		var have any
		fn := func(v any) error { have = v; return nil }
		r := By(fn)

		// --- When ---
		err := Validate("abc", r)

		// --- Then ---
		assert.Nil(t, err)
		assert.Equal(t, "abc", have)
	})

	t.Run("returns function error", func(t *testing.T) {
		// --- Given ---
		e := errors.New("test error")
		fn := func(any) error { return e }
		r := By(fn)

		// --- When ---
		err := Validate("abc", r)

		// --- Then ---
		assert.Same(t, e, err)
	})
}

func Test_ByRule_When(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		// --- Given ---
		fn := func(v any) error { return fmt.Errorf("error: %s", v) }

		// --- When ---
		r := By(fn).When(false)

		// --- Then ---
		err := Validate("abc", r)
		assert.Nil(t, err)
	})

	t.Run("true", func(t *testing.T) {
		// --- Given ---
		fn := func(v any) error { return fmt.Errorf("error: %s", v) }

		// --- When ---
		r := By(fn).When(true)

		// --- Then ---
		err := Validate("abc", r)
		assert.ErrorEqual(t, "error: abc", err)
	})
}

func Test_ByRule_Code(t *testing.T) {
	t.Run("set custom", func(t *testing.T) {
		// --- Given ---
		r := By(StrRuleFunc("abc"))

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("xyz")
		assert.ErrorEqual(t, "must be 'abc'", err)
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("with custom error core error", func(t *testing.T) {
		// --- Given ---
		r := By(StrRuleFunc("abc")).Error(errors.New("custom error"))

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("xyz")
		assert.ErrorEqual(t, "custom error", err)
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("with custom error vrr error", func(t *testing.T) {
		// --- Given ---
		r := By(StrRuleFunc("abc")).Error(xrr.New("custom error", "ECOther"))

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("xyz")
		assert.ErrorEqual(t, "custom error", err)
		xrrtest.AssertCode(t, "MyCode", err)
	})
}

func Test_ByRule_Error(t *testing.T) {
	t.Run("set custom", func(t *testing.T) {
		// --- Given ---
		r := By(StrRuleFunc("abc"))

		// --- When ---
		have := r.Error(xrr.New("custom error", "ECOther"))

		// --- Then ---
		err := have.Validate("xyz")
		assert.ErrorEqual(t, "custom error", err)
		xrrtest.AssertCode(t, "ECOther", err)
	})
}
