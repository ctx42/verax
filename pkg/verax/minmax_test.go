// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"testing"
	"text/template"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_Min(t *testing.T) {
	t.Run("supported type", func(t *testing.T) {
		// --- When ---
		have := Min(42)

		// --- Then ---
		assert.Equal(t, 42, have.threshold)
		assert.Equal(t, greaterEqualThan, have.operator)
		assert.Same(t, compareInt, have.with)
		assert.True(t, have.condition)
		assert.Same(t, tplMinGreaterEqualThan, have.errTpl)
		assert.Equal(t, ECInvThreshold, have.code)
	})

	t.Run("not supported type", func(t *testing.T) {
		// --- Given ---
		type my struct{ V int }

		// --- When ---
		have := Min(my{42})

		// --- Then ---
		assert.Equal(t, my{42}, have.threshold)
		assert.Equal(t, greaterEqualThan, have.operator)
		assert.Nil(t, nil, have.with)
		assert.True(t, have.condition)
		assert.Same(t, tplMinGreaterEqualThan, have.errTpl)
		assert.Equal(t, ECInvThreshold, have.code)
	})
}

func Test_Max(t *testing.T) {
	t.Run("supported type", func(t *testing.T) {
		// --- When ---
		have := Max(42)

		// --- Then ---
		assert.Equal(t, 42, have.threshold)
		assert.Equal(t, lessEqualThan, have.operator)
		assert.Same(t, compareInt, have.with)
		assert.True(t, have.condition)
		assert.Same(t, tplMaxLessEqualThan, have.errTpl)
		assert.Equal(t, ECInvThreshold, have.code)
	})

	t.Run("not supported type", func(t *testing.T) {
		// --- Given ---
		type my struct{ V int }

		// --- When ---
		have := Max(my{42})

		// --- Then ---
		assert.Equal(t, my{42}, have.threshold)
		assert.Equal(t, lessEqualThan, have.operator)
		assert.Nil(t, have.with)
		assert.True(t, have.condition)
		assert.Same(t, tplMaxLessEqualThan, have.errTpl)
		assert.Equal(t, ECInvThreshold, have.code)
	})
}

func Test_ThresholdRule_Exclusive_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rule     ThresholdRule
		operator int
		tpl      *template.Template
	}{
		{
			"greater or equal than to greater than",
			Min(42),
			greaterThan,
			tplMinGreaterThan,
		},
		{
			"already greater than",
			Min(42).Exclusive(),
			greaterThan,
			tplMinGreaterThan,
		},
		{
			"less or equal than to less than",
			Max(42),
			lessThan,
			tplMaxLessThan,
		},
		{
			"already less than",
			Max(42).Exclusive(),
			lessThan,
			tplMaxLessThan,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := tc.rule.Exclusive()

			// --- Then ---
			assert.Equal(t, tc.operator, have.operator)
			assert.Equal(t, tc.tpl, have.errTpl)
		})
	}
}

func Test_ThresholdRule_With(t *testing.T) {
	// --- Given ---
	cmp := func(want, have any) (int, error) { return 0, nil }
	r := Min(42)

	// --- When ---
	have := r.With(cmp)

	// --- Then ---
	assert.Same(t, cmp, have.with)
}

func Test_ThresholdRule_Validate(t *testing.T) {
	t.Run("nil is ok", func(t *testing.T) {
		// --- Given ---
		r := Min(42)

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty is ok", func(t *testing.T) {
		// --- Given ---
		r := Min(42)

		// --- When ---
		err := r.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("with is called with proper arguments", func(t *testing.T) {
		// --- Given ---
		var w, h any
		with := func(want, have any) (int, error) {
			w = want
			h = have
			return -1, nil
		}
		r := Max(42).With(with)

		// --- When ---
		_ = r.Validate(44)

		// --- Then ---
		assert.Equal(t, 42, w)
		assert.Equal(t, 44, h)
	})

	t.Run("with function error is handled", func(t *testing.T) {
		// --- Given ---
		with := func(want, have any) (int, error) { return 0, ErrTst }
		r := Max(42).With(with)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.Same(t, ErrTst, err)
	})

	t.Run("custom error code does not work with CompareFunc", func(t *testing.T) {
		// --- Given ---
		with := func(want, have any) (int, error) { return 0, ErrTst }
		r := Max(42).With(with).Code("MyCode")

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.ErrorIs(t, ErrTst, err)
		xrrtest.AssertCode(t, "ETstCode", err)
	})

	t.Run("the threshold outcome is checked", func(t *testing.T) {
		// --- Given ---
		r := Max(42)

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - the threshold outcome", func(t *testing.T) {
		// --- Given ---
		r := Max(42)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.ErrorEqual(t, "must be no greater than 42", err)
		xrrtest.AssertCode(t, ECInvThreshold, err)
	})

	t.Run("error - custom error code", func(t *testing.T) {
		// --- Given ---
		r := Max(42).Code("MyCode")

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.ErrorEqual(t, "must be no greater than 42", err)
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("error - custom error and error code", func(t *testing.T) {
		// --- Given ---
		r := Max(42).Error(ErrTst).Code("MyCode")

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.Same(t, ErrTst, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})
}

func Test_ThresholdRule_When(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		// --- When ---
		r := Max(42).When(false)

		// --- Then ---
		err := Validate(44, r)
		assert.Nil(t, err)
	})

	t.Run("true", func(t *testing.T) {
		// --- When ---
		r := Max(42).When(true)

		// --- Then ---
		err := Validate(44, r)
		assert.ErrorEqual(t, "must be no greater than 42", err)
	})
}

func Test_ThresholdRule_Code(t *testing.T) {
	t.Run("set custom error code", func(t *testing.T) {
		// --- Given ---
		r := Max(42)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate(44)
		assert.Error(t, err)
		xrrtest.AssertCode(t, "MyCode", err)
	})

	t.Run("custom error code for custom error", func(t *testing.T) {
		// --- Given ---
		r := Max(42).Error(ErrTst)

		// --- When ---
		have := r.Code("MyCode")

		// --- Then ---
		err := have.Validate(44)
		assert.Same(t, ErrTst, errors.Unwrap(err))
		xrrtest.AssertCode(t, "MyCode", err)
	})
}

func Test_ThresholdRule_Error(t *testing.T) {
	t.Run("set custom error", func(t *testing.T) {
		// --- Given ---
		r := Max(42)

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate(44)
		assert.Same(t, ErrTst, err)
	})

	t.Run("clears custom error code", func(t *testing.T) {
		// --- Given ---
		r := Max(42).Code("MyCode")

		// --- When ---
		have := r.Error(ErrTst)

		// --- Then ---
		err := have.Validate(44)
		assert.Same(t, ErrTst, err)
		xrrtest.AssertCode(t, "ETstCode", err)
	})
}

func Test_thresholdError(t *testing.T) {
	t.Run("threshold is a simple type", func(t *testing.T) {
		// --- When ---
		err := thresholdError(42, tplMaxLessEqualThan, "ECSimple")

		// --- Then ---
		assert.ErrorEqual(t, "must be no greater than 42", err)
		xrrtest.AssertCode(t, "ECSimple", err)
	})

	t.Run("threshold is a type implementing fmt.String", func(t *testing.T) {
		// --- Given ---
		s := NewTwoStr()

		// --- When ---
		err := thresholdError(s, tplMaxLessEqualThan, "ECFmt")

		// --- Then ---
		assert.ErrorEqual(t, "must be no greater than FStr FpStr", err)
		xrrtest.AssertCode(t, "ECFmt", err)
	})

	t.Run("threshold does not implement fmt.Stringer", func(t *testing.T) {
		// --- Given ---
		m := map[string]string{"a": "b"}

		// --- When ---
		err := thresholdError(m, tplMaxLessEqualThan, "ECode")

		// --- Then ---
		assert.ErrorEqual(t, "must be no greater than map[a:b]", err)
		xrrtest.AssertCode(t, "ECode", err)
	})
}

func Test_thresholdOutcome_tabular(t *testing.T) {
	tt := []struct {
		testN string

		operator int
		result   int
		want     bool
	}{
		{
			"a threshold must be greater than a value - value is less",
			greaterThan,
			-1,
			true,
		},
		{
			"a threshold must be greater than a value - value equal",
			greaterThan,
			0,
			false,
		},
		{
			"a threshold must be greater than a value - value greater",
			greaterThan,
			1,
			false,
		},

		// ---

		{
			"a threshold must be greater or equal than a value - value is less",
			greaterEqualThan,
			-1,
			true,
		},
		{
			"a threshold must be greater or equal than a value - value equal",
			greaterEqualThan,
			0,
			true,
		},
		{
			"a threshold must be greater or equal than a value - value greater",
			greaterEqualThan,
			1,
			false,
		},

		// ---

		{
			"a threshold must be less than a value - value is less",
			lessThan,
			-1,
			false,
		},
		{
			"a threshold must be less than a value - value equal",
			lessThan,
			0,
			false,
		},
		{
			"a threshold must be less than a value - value greater",
			lessThan,
			1,
			true,
		},

		// ---

		{
			"a threshold must be less or equal than a value - value is less",
			lessEqualThan,
			-1,
			false,
		},
		{
			"a threshold must be less or equal than a value - value equal",
			lessEqualThan,
			0,
			true,
		},
		{
			"a threshold must be less or equal than a value - value greater",
			lessEqualThan,
			1,
			true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := thresholdOutcome(tc.operator, tc.result)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_compareInt(t *testing.T) {
	t.Run("error - want is not integer", func(t *testing.T) {
		// --- When ---
		have, err := compareInt(1.0, 1)

		// --- Then ---
		assert.ErrorEqual(t, "cannot convert float64 to int64", err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})

	t.Run("error - have is not integer", func(t *testing.T) {
		// --- When ---
		have, err := compareInt(1, 1.0)

		// --- Then ---
		assert.ErrorEqual(t, "cannot convert float64 to int64", err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})
}

func Test_compareInt_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want    any
		have    any
		wantCmp int
	}{
		{"int - w is less than h", 1, 2, -1},
		{"int - w is equal to h", 1, 1, 0},
		{"int - w is greater than h", 1, 0, 1},

		{"int8 - w is less than h", int8(1), int8(2), -1},
		{"int8 - w is equal to h", int8(1), int8(1), 0},
		{"int8 - w is greater than h", int8(1), int8(0), 1},

		{"int16 - w is less than h", int16(1), int16(2), -1},
		{"int16 - w is equal to h", int16(1), int16(1), 0},
		{"int16 - w is greater than h", int16(1), int16(0), 1},

		{"int32 - w is less than h", int32(1), int32(2), -1},
		{"int32 - w is equal to h", int32(1), int32(1), 0},
		{"int32 - w is greater than h", int32(1), int32(0), 1},

		{"int64 - w is less than h", int64(1), int64(2), -1},
		{"int64 - w is equal to h", int64(1), int64(1), 0},
		{"int64 - w is greater than h", int64(1), int64(0), 1},

		{"duration - w is less than h", time.Second, time.Hour, -1},
		{"duration - w is equal to h", time.Second, time.Second, 0},
		{"duration - w is greater than h", time.Hour, time.Second, 1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := compareInt(tc.want, tc.have)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCmp, have)
		})
	}
}

func Test_compareUint(t *testing.T) {
	t.Run("error - want is not an unsigned integer", func(t *testing.T) {
		// --- When ---
		have, err := compareUint(1.0, uint(1))

		// --- Then ---
		assert.ErrorEqual(t, "cannot convert float64 to uint64", err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})

	t.Run("error - have is not an unsigned integer", func(t *testing.T) {
		// --- When ---
		have, err := compareUint(uint(1), 1.0)

		// --- Then ---
		assert.ErrorEqual(t, "cannot convert float64 to uint64", err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})
}

func Test_compareUint_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want    any
		have    any
		wantCmp int
	}{
		{"uint - w is less than h", uint(1), uint(2), -1},
		{"uint - w is equal to h", uint(1), uint(1), 0},
		{"uint - w is greater than h", uint(1), uint(0), 1},

		{"uint8 - w is less than h", uint8(1), uint8(2), -1},
		{"uint8 - w is equal to h", uint8(1), uint8(1), 0},
		{"uint8 - w is greater than h", uint8(1), uint8(0), 1},

		{"uint16 - w is less than h", uint16(1), uint16(2), -1},
		{"uint16 - w is equal to h", uint16(1), uint16(1), 0},
		{"uint16 - w is greater than h", uint16(1), uint16(0), 1},

		{"uint32 - w is less than h", uint32(1), uint32(2), -1},
		{"uint32 - w is equal to h", uint32(1), uint32(1), 0},
		{"uint32 - w is greater than h", uint32(1), uint32(0), 1},

		{"uint64 - w is less than h", uint64(1), uint64(2), -1},
		{"uint64 - w is equal to h", uint64(1), uint64(1), 0},
		{"uint64 - w is greater than h", uint64(1), uint64(0), 1},

		{"uintptr - w is less than h", uintptr(1), uintptr(2), -1},
		{"uintptr - w is equal to h", uintptr(1), uintptr(1), 0},
		{"uintptr - w is greater than h", uintptr(1), uintptr(0), 1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := compareUint(tc.want, tc.have)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCmp, have)
		})
	}
}

func Test_compareFloat(t *testing.T) {
	t.Run("error - want is not float", func(t *testing.T) {
		// --- When ---
		have, err := compareFloat(1, 1.0)

		// --- Then ---
		assert.ErrorEqual(t, "cannot convert int to float64", err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})

	t.Run("error - have is not a float", func(t *testing.T) {
		// --- When ---
		have, err := compareFloat(1.0, 1)

		// --- Then ---
		assert.ErrorEqual(t, "cannot convert int to float64", err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})
}

func Test_compareFloat_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want    any
		have    any
		wantCmp int
	}{
		{"float32 - w is less than h", float32(1), float32(2), -1},
		{"float32 - w is equal to h", float32(1), float32(1), 0},
		{"float32 - w is greater than h", float32(1), float32(0), 1},

		{"float64 - w is less than h", float64(1), float64(2), -1},
		{"float64 - w is equal to h", float64(1), float64(1), 0},
		{"float64 - w is greater than h", float64(1), float64(0), 1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := compareFloat(tc.want, tc.have)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCmp, have)
		})
	}
}

func Test_compareTime(t *testing.T) {
	t.Run("error - want is not time", func(t *testing.T) {
		// --- When ---
		have, err := compareTime(1, time.Now())

		// --- Then ---
		assert.ErrorEqual(t, "cannot convert int to time.Time", err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})

	t.Run("error - have is not time", func(t *testing.T) {
		// --- When ---
		have, err := compareTime(time.Now(), 1)

		// --- Then ---
		assert.ErrorEqual(t, "cannot convert int to time.Time", err)
		xrrtest.AssertCode(t, ECInvType, err)
		assert.Equal(t, 0, have)
	})
}

func Test_compareTime_tabular(t *testing.T) {
	tim0 := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
	tim1 := time.Date(2001, 1, 2, 3, 4, 5, 0, time.UTC)

	tt := []struct {
		testN string

		want    any
		have    any
		wantCmp int
	}{
		{"w is less than h", tim0, tim1, -1},
		{"w is equal to h", tim0, tim0, 0},
		{"w is greater than h", tim1, tim0, 1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := compareTime(tc.want, tc.have)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCmp, have)
		})
	}
}

func Test_compareFor_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val  any
		want CompareFunc
	}{
		{"int", 1, compareInt},
		{"int8", int8(1), compareInt},
		{"int16", int16(1), compareInt},
		{"int32", int32(1), compareInt},
		{"int64", int64(1), compareInt},

		{"uint", uint(1), compareUint},
		{"uint8", uint8(1), compareUint},
		{"uint16", uint16(1), compareUint},
		{"uint32", uint32(1), compareUint},
		{"uint64", uint64(1), compareUint},
		{"uintptr", uintptr(1), compareUint},

		{"float32", float32(1), compareFloat},
		{"float64", 1.0, compareFloat},

		{"time", time.Now(), compareTime},

		{"not supported", NewTwoStr(), nil},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := compareFor(tc.val)

			// --- Then ---
			assert.Same(t, tc.want, have)
		})
	}
}
