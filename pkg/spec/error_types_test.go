// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package spec

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
		have := NewError("msg", "ECTst")

		// --- Then ---
		e, _ := assert.SameType(t, &Error{}, have)
		assert.ErrorEqual(t, "msg", e)
		xrrtest.AssertCode(t, "ECTst", e)
	})

	t.Run("with a metadata option", func(t *testing.T) {
		// --- Given ---
		meta := xrr.Meta().Str("key", "val").Option()

		// --- When ---
		have := NewError("msg", "ECTst", meta)

		// --- Then ---
		e, _ := assert.SameType(t, &Error{}, have)
		assert.ErrorEqual(t, "msg", e)
		xrrtest.AssertCode(t, "ECTst", e)
		assert.Equal(t, map[string]any{"key": "val"}, e.MetaAll())
	})

	t.Run("marshals to JSON", func(t *testing.T) {
		// --- Given ---
		meta := xrr.Meta().Str("k", "v").Option()
		e := NewError("msg", "ECTst", meta)

		// --- When ---
		have, err := json.Marshal(e)

		// --- Then ---
		assert.NoError(t, err)
		want := `{"error":"msg", "code":"ECTst", "meta":{"k":"v"}}`
		assert.JSON(t, want, string(have))
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

func Test_NewErrorf(t *testing.T) {
	t.Run("plain format", func(t *testing.T) {
		// --- When ---
		have := NewErrorf("msg")

		// --- Then ---
		e, _ := assert.SameType(t, &Error{}, have)
		assert.ErrorEqual(t, "msg", have)
		xrrtest.AssertCode(t, xrr.ECGeneric, e)
		assert.Nil(t, e.MetaAll())
	})

	t.Run("format with args", func(t *testing.T) {
		// --- When ---
		have := NewErrorf("value: %d", 42)

		// --- Then ---
		e, _ := assert.SameType(t, &Error{}, have)
		assert.ErrorEqual(t, "value: 42", have)
		xrrtest.AssertCode(t, xrr.ECGeneric, e)
		assert.Nil(t, e.MetaAll())
	})

	t.Run("with code", func(t *testing.T) {
		// --- When ---
		have := NewErrorf("msg %d", 42, xrr.WithCode("ECTst"))

		// --- Then ---
		e, _ := assert.SameType(t, &Error{}, have)
		assert.ErrorEqual(t, "msg 42", have)
		xrrtest.AssertCode(t, "ECTst", e)
		assert.Nil(t, e.MetaAll())
	})

	t.Run("wraps error", func(t *testing.T) {
		// --- Given ---
		cause := errors.New("original")

		// --- When ---
		have := NewErrorf("connect failed: %w", cause)

		// --- Then ---
		e, _ := assert.SameType(t, &Error{}, have)
		assert.ErrorEqual(t, "connect failed: original", have)
		assert.True(t, errors.Is(have, cause))
		xrrtest.AssertCode(t, xrr.ECGeneric, e)
		assert.Nil(t, e.MetaAll())
	})

	t.Run("wraps error with code", func(t *testing.T) {
		// --- Given ---
		cause := errors.New("original")

		// --- When ---
		have := NewErrorf("connect failed: %w", cause, xrr.WithCode("ECTst"))

		// --- Then ---
		e, _ := assert.SameType(t, &Error{}, have)
		assert.ErrorEqual(t, "connect failed: original", have)
		assert.True(t, errors.Is(have, cause))
		xrrtest.AssertCode(t, "ECTst", e)
		assert.Nil(t, e.MetaAll())
	})
}

func Test_NewFieldError(t *testing.T) {
	t.Run("message has field", func(t *testing.T) {
		// --- Given ---
		e := errors.New("msg")

		// --- When ---
		have := NewFieldError("field0", e)

		// --- Then ---
		assert.ErrorEqual(t, "field0: msg", have)
		xrrtest.AssertHasField(t, "field0", have)
	})

	t.Run("nil error", func(t *testing.T) {
		// --- When ---
		have := NewFieldError("field0", nil)

		// --- Then ---
		assert.True(t, have == nil)
	})

	t.Run("marshals to JSON", func(t *testing.T) {
		// --- Given ---
		e := NewFieldError("field0", NewError("inner msg", "ECInner"))

		// --- When ---
		have, err := json.Marshal(e)

		// --- Then ---
		assert.NoError(t, err)
		want := `{"field0":{"error":"inner msg","code":"ECInner"}}`
		assert.JSON(t, want, string(have))
	})
}

func Test_NewFieldErrors(t *testing.T) {
	t.Run("message has all fields", func(t *testing.T) {
		// --- Given ---
		fields := map[string]error{
			"field0": errors.New("msg0"),
			"field1": errors.New("msg1"),
		}

		// --- When ---
		have := NewFieldErrors(fields)

		// --- Then ---
		assert.ErrorEqual(t, "field0: msg0; field1: msg1", have)
		xrrtest.AssertHasField(t, "field0", have)
		xrrtest.AssertHasField(t, "field1", have)
	})

	t.Run("map not copied", func(t *testing.T) {
		// --- Given ---
		fields := map[string]error{"field0": errors.New("msg0")}
		err := NewFieldErrors(fields)

		// --- When ---
		fields["field1"] = errors.New("msg1")

		// --- Then ---
		xrrtest.AssertHasField(t, "field1", err)
	})

	t.Run("marshals to JSON", func(t *testing.T) {
		// --- Given ---
		e := NewFieldErrors(map[string]error{
			"field0": NewError("inner msg", "ECInner"),
		})

		// --- When ---
		have, err := json.Marshal(e)

		// --- Then ---
		assert.NoError(t, err)
		want := `{"field0":{"error":"inner msg","code":"ECInner"}}`
		assert.JSON(t, want, string(have))
	})
}

func Test_IsSpecError(t *testing.T) {
	t.Run("true for Error", func(t *testing.T) {
		// --- Given ---
		err := NewError("msg", "ECTst")

		// --- When ---
		have := IsSpecError(err)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("true for FieldsError", func(t *testing.T) {
		// --- Given ---
		err := NewFieldError("field0", NewError("msg", "ECTst"))

		// --- When ---
		have := IsSpecError(err)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("false for other domain", func(t *testing.T) {
		// --- Given ---
		err := errors.New("test message")

		// --- When ---
		have := IsSpecError(err)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("false for nil", func(t *testing.T) {
		// --- When ---
		have := IsSpecError(nil)

		// --- Then ---
		assert.False(t, have)
	})
}
