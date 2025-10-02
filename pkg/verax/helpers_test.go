// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"testing"

	"github.com/ctx42/testing/pkg/assert"

	"github.com/ctx42/xrr/pkg/xrr"
)

func Test_ToAnySlice(t *testing.T) {
	assert.Equal(t, []any{}, ToAnySlice[int]())
	assert.Equal(t, []any{}, ToAnySlice[string]())
	assert.Equal(t, []any{"a", "b", "c"}, ToAnySlice("a", "b", "c"))
	assert.Equal(t, []any{1, 2, 3}, ToAnySlice(1, 2, 3))
}

func Test_EncloseError(t *testing.T) {
	t.Run("enclose", func(t *testing.T) {
		// --- Given ---
		e := errors.New("error")

		// --- When ---
		err := EncloseError(e)

		// --- Then ---
		var xe xrr.Envelope
		assert.Type(t, &xe, err)
		assert.Same(t, e, xe.Unwrap())
		assert.ErrorIs(t, ErrValidation, err)
	})

	t.Run("returns nil when nil error", func(t *testing.T) {
		// --- When ---
		err := EncloseError(nil)

		// --- Then ---
		assert.Nil(t, err)
	})
}

func Test_setCode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		e := errors.New("error")

		// --- When ---
		err := setCode(e, "ECode")

		// --- Then ---
		var xe *xrr.Error
		assert.Type(t, &xe, err)
		assert.Same(t, e, xe.Unwrap())
		assert.Equal(t, "ECode", xe.ErrorCode())
	})

	t.Run("nil error", func(t *testing.T) {
		// --- When ---
		err := setCode(nil, "ECode")

		// --- Then ---
		assert.Nil(t, err)
	})

	t.Run("it does not wrap when the code is the same", func(t *testing.T) {
		// --- Given ---
		e := xrr.New("error", "ECode")

		// --- When ---
		err := setCode(e, "ECode")

		// --- Then ---
		assert.Same(t, e, err)
	})

	t.Run("returns the same instance when code is empty", func(t *testing.T) {
		// --- Given ---
		e := errors.New("error")

		// --- When ---
		err := setCode(e, "")

		// --- Then ---
		assert.Same(t, e, err)
	})
}
