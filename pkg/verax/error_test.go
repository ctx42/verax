// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

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

func Test_NewInternalError(t *testing.T) {
	t.Run("without options", func(t *testing.T) {
		// --- When ---
		err := NewInternalError("msg", "ECTst")

		// --- Then ---
		e, _ := assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, "msg", e)
		xrrtest.AssertCode(t, "ECTst", e)
	})

	t.Run("with a metadata option", func(t *testing.T) {
		// --- Given ---
		meta := xrr.Meta().Str("key", "val").Option()

		// --- When ---
		err := NewInternalError("msg", "ECTst", meta)

		// --- Then ---
		e, _ := assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, "msg", e)
		xrrtest.AssertCode(t, "ECTst", e)
		assert.Equal(t, map[string]any{"key": "val"}, e.MetaAll())
	})

	t.Run("marshals to JSON", func(t *testing.T) {
		// --- Given ---
		meta := xrr.Meta().Str("k", "v").Option()
		e := NewInternalError("msg", "ECTst", meta)

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
		var e *InternalError

		// --- When ---
		err := json.Unmarshal(data, &e)

		// --- Then ---
		assert.NoError(t, err)
		assert.ErrorEqual(t, "msg", e)
		xrrtest.AssertCode(t, "ECTst", e)
		assert.Equal(t, map[string]any{"k": "v"}, e.MetaAll())
	})
}

func Test_FieldError(t *testing.T) {
	t.Run("the error message includes the field name", func(t *testing.T) {
		// --- Given ---
		e := errors.New("msg")

		// --- When ---
		err := FieldError("field0", e)

		// --- Then ---
		assert.ErrorEqual(t, "field0: msg", err)
		xrrtest.AssertHasField(t, "field0", err)
	})

	t.Run("marshals to JSON", func(t *testing.T) {
		// --- Given ---
		e := FieldError("field0", NewError("inner msg", "ECInner"))

		// --- When ---
		data, err := json.Marshal(e)

		// --- Then ---
		assert.NoError(t, err)
		wData := `{"field0":{"error":"inner msg","code":"ECInner"}}`
		assert.JSON(t, wData, string(data))
	})
}

func Test_IsError(t *testing.T) {
	t.Run("true for Error", func(t *testing.T) {
		// --- When ---
		have := IsError(NewError("msg", "ECTst"))

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("true for InternalError", func(t *testing.T) {
		// --- When ---
		have := IsError(NewInternalError("msg", "ECTst"))

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("true for FieldsError", func(t *testing.T) {
		// --- When ---
		have := IsError(FieldError("field0", NewError("msg", "ECTst")))

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("false for nil", func(t *testing.T) {
		// --- When ---
		have := IsError(nil)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_IsValidationError(t *testing.T) {
	t.Run("true for Error", func(t *testing.T) {
		// --- When ---
		have := IsValidationError(NewError("msg", "ECTst"))

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("true for FieldsError", func(t *testing.T) {
		// --- When ---
		have := IsValidationError(FieldError("field0", NewError("msg", "ECTst")))

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("false for InternalError", func(t *testing.T) {
		// --- When ---
		have := IsValidationError(NewInternalError("msg", "ECTst"))

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("false for nil", func(t *testing.T) {
		// --- When ---
		have := IsValidationError(nil)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_IsInternalError(t *testing.T) {
	t.Run("true for InternalError", func(t *testing.T) {
		// --- When ---
		have := IsInternalError(NewInternalError("msg", "ECTst"))

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("false for Error", func(t *testing.T) {
		// --- When ---
		have := IsInternalError(NewError("msg", "ECTst"))

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("false for nil", func(t *testing.T) {
		// --- When ---
		have := IsInternalError(nil)

		// --- Then ---
		assert.False(t, have)
	})
}
