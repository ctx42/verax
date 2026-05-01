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

func Test_NewFieldErrors(t *testing.T) {
	t.Run("the error message includes all field names", func(t *testing.T) {
		// --- Given ---
		fields := map[string]error{
			"field0": errors.New("msg0"),
			"field1": errors.New("msg1"),
		}

		// --- When ---
		err := NewFieldErrors(fields)

		// --- Then ---
		assert.ErrorEqual(t, "field0: msg0; field1: msg1", err)
		xrrtest.AssertHasField(t, "field0", err)
		xrrtest.AssertHasField(t, "field1", err)
	})

	t.Run("stores the map directly without copying", func(t *testing.T) {
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
		data, err := json.Marshal(e)

		// --- Then ---
		assert.NoError(t, err)
		wData := `{"field0":{"error":"inner msg","code":"ECInner"}}`
		assert.JSON(t, wData, string(data))
	})
}

func Test_IsVeraxError(t *testing.T) {
	t.Run("true for Error", func(t *testing.T) {
		// --- Given ---
		err := NewError("msg", "ECTst")

		// --- When ---
		have := IsVeraxError(err)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("true for InternalError", func(t *testing.T) {
		// --- Given ---
		err := NewInternalError("msg", "ECTst")

		// --- When ---
		have := IsVeraxError(err)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("true for FieldsError", func(t *testing.T) {
		// --- Given ---
		err := NewFieldError("field0", NewError("msg", "ECTst"))

		// --- When ---
		have := IsVeraxError(err)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("false for errors not from this domain", func(t *testing.T) {
		// --- Given ---
		err := errors.New("test message")

		// --- When ---
		have := IsVeraxError(err)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("false for nil", func(t *testing.T) {
		// --- When ---
		have := IsVeraxError(nil)

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_IsValidationError(t *testing.T) {
	t.Run("true for Error", func(t *testing.T) {
		// --- Given ---
		err := NewError("msg", "ECTst")

		// --- When ---
		have := IsValidationError(err)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("true for FieldsError", func(t *testing.T) {
		// --- Given ---
		err := NewFieldError("field0", NewError("msg", "ECTst"))

		// --- When ---
		have := IsValidationError(err)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("false for InternalError", func(t *testing.T) {
		// --- Given ---
		err := NewInternalError("msg", "ECTst")

		// --- When ---
		have := IsValidationError(err)

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("false for errors not from this domain", func(t *testing.T) {
		// --- Given ---
		err := errors.New("test message")

		// --- When ---
		have := IsValidationError(err)

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

	t.Run("false for errors not from this domain", func(t *testing.T) {
		// --- Given ---
		err := errors.New("test message")

		// --- When ---
		have := IsInternalError(err)

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
