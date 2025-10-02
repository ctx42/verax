// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_Error(t *testing.T) {
	// --- When ---
	r := Error(ErrTst)

	// --- Then ---
	assert.Same(t, ErrTst, r.err)
}

func Test_ErrorRule_Validate(t *testing.T) {
	t.Run("returns given error", func(t *testing.T) {
		// --- Given ---
		r := Error(ErrTst)

		// --- When ---
		err := r.Validate("any")

		// --- Then ---
		assert.Same(t, ErrTst, err)
	})

	t.Run("no error when condition false", func(t *testing.T) {
		// --- Given ---
		r := Error(ErrTst).When(false)

		// --- When ---
		err := r.Validate("any")

		// --- Then ---
		assert.NoError(t, err)
	})
}

func Test_ErrorRule_When(t *testing.T) {
	// --- Given ---
	r := Error(ErrTst)

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_ErrorRule_Code(t *testing.T) {
	t.Run("set custom error code", func(t *testing.T) {
		// --- Given ---
		r := Error(ErrTst)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate("any")
		assert.Error(t, err)
		assert.Same(t, ErrTst, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})
}
