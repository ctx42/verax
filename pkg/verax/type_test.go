// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"reflect"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_Type(t *testing.T) {
	t.Run("type", func(t *testing.T) {
		// --- Given ---
		typ := reflect.TypeOf(42)

		// --- When ---
		have := Type(typ)

		// --- Then ---
		assert.Equal(t, typ, have.typ)
		assert.True(t, have.condition)
		assert.Same(t, ErrExpType, have.err)
	})

	t.Run("nil", func(t *testing.T) {
		// --- Given ---
		typ := reflect.TypeOf(nil)

		// --- When ---
		have := Type(typ)

		// --- Then ---
		assert.Equal(t, typ, have.typ)
		assert.True(t, have.condition)
		assert.Same(t, ErrExpType, have.err)
	})
}

func Test_TypeOf(t *testing.T) {
	t.Run("type", func(t *testing.T) {
		// --- When ---
		have := TypeOf(42)

		// --- Then ---
		assert.Equal(t, reflect.TypeOf(42), have.typ)
		assert.True(t, have.condition)
		assert.Same(t, ErrExpType, have.err)
	})

	t.Run("nil", func(t *testing.T) {
		// --- When ---
		have := TypeOf(nil)

		// --- Then ---
		assert.Equal(t, reflect.TypeOf(nil), have.typ)
		assert.True(t, have.condition)
		assert.Same(t, ErrExpType, have.err)
	})
}

func Test_TypeRule_Validate(t *testing.T) {
	t.Run("nil is ok", func(t *testing.T) {
		// --- Given ---
		r := TypeOf(42)

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("same types", func(t *testing.T) {
		// --- Given ---
		r := TypeOf(42)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("type and value nil", func(t *testing.T) {
		// --- Given ---
		r := TypeOf(nil)

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("want untyped nil", func(t *testing.T) {
		// --- Given ---
		r := TypeOf(nil)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.Error(t, ErrExpType, err)
		xrrtest.AssertCode(t, ECInvType, err)
	})

	t.Run("returns nil when the condition is false", func(t *testing.T) {
		// --- Given ---
		r := TypeOf(4.2).When(false)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - no the same types", func(t *testing.T) {
		// --- Given ---
		r := TypeOf(4.2)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.Error(t, ErrExpType, err)
		xrrtest.AssertCode(t, ECInvType, err)
	})
}

func Test_TypeRule_When(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		// --- When ---
		r := TypeOf(42).When(false)

		// --- Then ---
		err := Validate(44, r)
		assert.Nil(t, err)
	})

	t.Run("true", func(t *testing.T) {
		// --- When ---
		r := TypeOf(42).When(true)

		// --- Then ---
		err := Validate(4.4, r)
		assert.Error(t, ErrExpType, err)
	})
}

func Test_TypeRule_Code(t *testing.T) {
	t.Run("set custom error code", func(t *testing.T) {
		// --- Given ---
		r := TypeOf(42)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate(4.4)
		assert.Error(t, err)
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("custom error code for custom error", func(t *testing.T) {
		// --- Given ---
		r := TypeOf(42).Error(ErrTst)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate(4.4)
		assert.Same(t, ErrTst, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})
}

func Test_TypeRule_Error(t *testing.T) {
	t.Run("set custom error", func(t *testing.T) {
		// --- Given ---
		r := TypeOf(42)

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate(4.4)
		assert.Same(t, ErrTst, err)
	})

	t.Run("clears custom error code", func(t *testing.T) {
		// --- Given ---
		r := TypeOf(42).Code("MyCode")

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate(4.4)
		assert.Same(t, ErrTst, err)
		xrrtest.AssertCode(t, "ETstCode", err)
	})
}
