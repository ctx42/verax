// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package rule

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
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
		// --- Given ---
		val := base64.StdEncoding.EncodeToString([]byte("test"))
		assert.Equal(t, "dGVzdA==", val)

		// --- When ---
		have := IsBase64(val)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("standard binary", func(t *testing.T) {
		// --- Given ---
		bin := must.Value(hex.DecodeString("00203040503f33"))
		val := base64.StdEncoding.EncodeToString(bin)
		assert.Equal(t, "ACAwQFA/Mw==", val)

		// --- When ---
		have := IsBase64(val)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("url binary", func(t *testing.T) {
		// --- Given ---
		bin := must.Value(hex.DecodeString("00203040503f33"))
		val := base64.URLEncoding.EncodeToString(bin)
		assert.Equal(t, "ACAwQFA_Mw==", val)

		// --- When ---
		have := IsBase64(val)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("standard no padding", func(t *testing.T) {
		// --- Given ---
		bin := must.Value(hex.DecodeString("3200ff3a"))
		val := base64.RawStdEncoding.EncodeToString(bin)
		assert.Equal(t, "MgD/Og", val)

		// --- When ---
		have := IsBase64(val)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("url no padding", func(t *testing.T) {
		// --- Given ---
		bin := must.Value(hex.DecodeString("3200ff3a"))
		val := base64.RawURLEncoding.EncodeToString(bin)
		assert.Equal(t, "MgD_Og", val)

		// --- When ---
		have := IsBase64(val)

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
		// --- Given ---
		val := base64.StdEncoding.EncodeToString([]byte("test"))

		// --- When ---
		err := Base64.Validate(val)

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
		assert.True(t, verax.IsInternalError(err))
		wMsg := "must be a valid base64: expected string, got int"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, verax.ECInvType, err)
	})
}
