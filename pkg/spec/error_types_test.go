// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package vcfg

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_NewError(t *testing.T) {
	t.Run("without options", func(t *testing.T) {
		// --- When ---
		err := NewError("msg", "ECTst")

		// --- Then ---
		e, _ := assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "msg", e)
		xrrtest.AssertCode(t, "ECTst", e)
	})

	t.Run("with a metadata option", func(t *testing.T) {
		// --- Given ---
		meta := xrr.Meta().Str("key", "val").Option()

		// --- When ---
		err := NewError("msg", "ECTst", meta)

		// --- Then ---
		e, _ := assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "msg", e)
		xrrtest.AssertCode(t, "ECTst", e)
		assert.Equal(t, map[string]any{"key": "val"}, e.MetaAll())
	})

	t.Run("marshals to JSON", func(t *testing.T) {
		// --- Given ---
		meta := xrr.Meta().Str("k", "v").Option()
		e := NewError("msg", "ECTst", meta)

		// --- When ---
		data, err := json.Marshal(e)

		// --- Then ---
		assert.NoError(t, err)
		wData := `{"error":"msg", "code":"ECTst", "meta":{"k":"v"}}`
		assert.JSON(t, wData, string(data))
	})

	t.Run("unmarshals from JSON", func(t *testing.T) {
		// --- Given ---
		data := []byte(`{"error":"msg","code":"ECTst","meta":{"k":"v"}}`)
		var e *Error

		// --- When ---
		err := json.Unmarshal(data, &e)

		// --- Then ---
		assert.NoError(t, err)
		assert.ErrorEqual(t, "msg", e)
		xrrtest.AssertCode(t, "ECTst", e)
		assert.Equal(t, map[string]any{"k": "v"}, e.MetaAll())
	})
}

func Test_NewFieldError(t *testing.T) {
	t.Run("the error message includes the field name", func(t *testing.T) {
		// --- Given ---
		e := errors.New("msg")

		// --- When ---
		err := NewFieldError("field0", e)

		// --- Then ---
		assert.ErrorEqual(t, "field0: msg", err)
		xrrtest.AssertHasField(t, "field0", err)
	})

	t.Run("marshals to JSON", func(t *testing.T) {
		// --- Given ---
		e := NewFieldError("field0", NewError("inner msg", "ECInner"))

		// --- When ---
		data, err := json.Marshal(e)

		// --- Then ---
		assert.NoError(t, err)
		wData := `{"field0":{"error":"inner msg","code":"ECInner"}}`
		assert.JSON(t, wData, string(data))
	})
}

func Test_IsCfgError(t *testing.T) {
	t.Run("true for Error", func(t *testing.T) {
		// --- Given ---
		err := NewError("msg", "ECTst")

		// --- When ---
		have := IsConfigError(err)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("true for FieldsError", func(t *testing.T) {
		// --- Given ---
		err := NewFieldError("field0", NewError("msg", "ECTst"))

		// --- When ---
		have := IsConfigError(err)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("false for nil", func(t *testing.T) {
		// --- When ---
		have := IsConfigError(nil)

		// --- Then ---
		assert.False(t, have)
	})
}
