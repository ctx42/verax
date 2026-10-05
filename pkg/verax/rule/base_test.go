// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package rule

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/verax"
)

func Test_IsBase64(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		// --- When ---
		have := IsBase64("")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("standard text", func(t *testing.T) {
		// --- When ---
		have := IsBase64("dGVzdA==")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("standard binary", func(t *testing.T) {
		// --- When ---
		have := IsBase64("ACAwQFA/Mw==")

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("url binary", func(t *testing.T) {
		// --- When ---
		have := IsBase64("ACAwQFA_Mw==")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("standard no padding", func(t *testing.T) {
		// --- When ---
		have := IsBase64("MgD/Og")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("url no padding", func(t *testing.T) {
		// --- When ---
		have := IsBase64("MgD_Og")

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("invalid", func(t *testing.T) {
		// --- When ---
		have := IsBase64("aa")

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_Base64(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		err := Base64.Validate("dGVzdA==")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := Base64.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - validation", func(t *testing.T) {
		// --- When ---
		err := Base64.Validate("abc")

		// --- Then ---
		assert.ErrorEqual(t, msgBase64, err)
		xrrtest.AssertCode(t, ECBase64, err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- When ---
		err := Base64.Validate(42)

		// --- Then ---
		assert.SameType(t, &verax.InternalError{}, err)
		want := "must be a valid base64: expected string, got int"
		assert.ErrorEqual(t, want, err)
		xrrtest.AssertCode(t, verax.ECInvType, err)
	})
}
