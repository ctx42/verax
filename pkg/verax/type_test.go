// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"reflect"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_Type(t *testing.T) {
	// --- Given ---
	typ := reflect.TypeFor[int]()

	// --- When ---
	have := Type(typ)

	// --- Then ---
	want := TypeRule{
		typ:       typ,
		condition: true,
		msg:       msgInvType,
		code:      ECInvType,
		flags:     0,
	}
	assert.Equal(t, want, have)
}

func Test_TypeOf(t *testing.T) {
	t.Run("type", func(t *testing.T) {
		// --- When ---
		have := TypeOf(42)

		// --- Then ---
		want := TypeRule{
			typ:       reflect.TypeFor[int](),
			condition: true,
			msg:       msgInvType,
			code:      ECInvType,
			flags:     0,
		}
		assert.Equal(t, want, have)
	})

	t.Run("nil", func(t *testing.T) {
		// --- When ---
		have := TypeOf(nil)

		// --- Then ---
		want := TypeRule{
			typ:       reflect.Type(nil),
			condition: true,
			msg:       msgInvType,
			code:      ECInvType,
			flags:     0,
		}
		assert.Equal(t, want, have)
	})
}

func Test_TypeRule_Validate(t *testing.T) {
	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		r := TypeRule{typ: reflect.TypeFor[int](), condition: false}

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("nil ok", func(t *testing.T) {
		// --- Given ---
		r := TypeRule{typ: reflect.TypeFor[int](), condition: true}

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("same types", func(t *testing.T) {
		// --- Given ---
		r := TypeRule{typ: reflect.TypeFor[int](), condition: true}

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("type and value are nil", func(t *testing.T) {
		// --- Given ---
		r := TypeRule{typ: nil, condition: true}

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("want untyped nil", func(t *testing.T) {
		// --- Given ---
		r := TypeRule{
			typ:       reflect.Type(nil),
			condition: true,
			msg:       "test err",
			code:      "ECTst",
		}

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "test err", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})

	t.Run("error - no the same types", func(t *testing.T) {
		// --- Given ---
		r := TypeOf(4.2)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, msgInvType, err)
		xrrtest.AssertCode(t, ECInvType, err)
	})
}

func Test_TypeRule_When(t *testing.T) {
	// --- Given ---
	r := TypeRule{}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_TypeRule_Message(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := TypeRule{}

		// --- When ---
		have := r.Message("test err")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Equal(t, flgCustomMsg, have.flags)
	})

	t.Run("an empty string is a noop", func(t *testing.T) {
		// --- Given ---
		r := TypeRule{msg: "test err"}

		// --- When ---
		have := r.Message("")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Zero(t, have.flags)
	})
}

func Test_TypeRule_Code(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := TypeRule{}

		// --- When ---
		have := r.Code("ECTst")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Equal(t, flgCustomCode, have.flags)
	})

	t.Run("an empty string is noop", func(t *testing.T) {
		// --- Given ---
		r := TypeRule{code: "ECTst"}

		// --- When ---
		have := r.Code("")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Zero(t, have.flags)
	})
}
