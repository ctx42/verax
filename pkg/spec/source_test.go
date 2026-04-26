// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package spec

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/check"
)

func Test_NewSource(t *testing.T) {
	t.Run("anonymous func", func(t *testing.T) {
		// --- Given ---
		fn := func() {}

		// --- When ---
		have, err := NewSource("fn", fn)

		// --- Then ---
		assert.NoError(t, err)
		want := Source{
			Lang: "go",
			Name: "fn",
			Src:  "github.com/ctx42/verax/pkg/spec.func1.1",
			Desc: "",
			val:  fn,
		}
		assert.Equal(t, want, have, check.WithSkipTrail("Source.ptr"))
		assert.True(t, have.ptr > 0)
	})

	t.Run("func", func(t *testing.T) {
		// --- When ---
		have, err := NewSource("fn", GetSrcPointer)

		// --- Then ---
		assert.NoError(t, err)
		want := Source{
			Lang: "go",
			Name: "fn",
			Src:  "github.com/ctx42/verax/pkg/spec.GetSrcPointer",
			Desc: "",
			val:  GetSrcPointer,
		}
		assert.Equal(t, want, have, check.WithSkipTrail("Source.ptr"))
		assert.True(t, have.ptr > 0)
	})

	t.Run("func std", func(t *testing.T) {
		// --- When ---
		have, err := NewSource("fn", math.Max)

		// --- Then ---
		assert.NoError(t, err)
		want := Source{
			Lang: "go",
			Name: "fn",
			Src:  "math.Max",
			Desc: "",
			val:  math.Max,
		}
		assert.Equal(t, want, have, check.WithSkipTrail("Source.ptr"))
		assert.True(t, have.ptr > 0)
	})

	t.Run("error - nil", func(t *testing.T) {
		// --- When ---
		have, err := NewSource("fn", nil)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSource, err)
		assert.Zero(t, have)
	})

	t.Run("error - unsupported type", func(t *testing.T) {
		// --- When ---
		have, err := NewSource("fn", errors.New("test"))

		// --- Then ---
		assert.ErrorIs(t, ErrInvSource, err)
		assert.Zero(t, have)
	})
}

func Test_Source_SetLang(t *testing.T) {
	// --- Given ---
	src := Source{}

	// --- When ---
	have := src.SetLang("go")

	// --- Then ---
	assert.Equal(t, "go", have.Lang)
	assert.Equal(t, "", src.Lang)
}

func Test_Source_SetDesc(t *testing.T) {
	// --- Given ---
	src := Source{}

	// --- When ---
	have := src.SetDesc("desc")

	// --- Then ---
	assert.Equal(t, "desc", have.Desc)
	assert.Equal(t, "", src.Desc)
}

func Test_Source_Val(t *testing.T) {
	// --- Given ---
	src := Source{val: math.Max}

	// --- When ---
	have := src.Val()

	// --- Then ---
	assert.Same(t, math.Max, have)
}

func Test_Source_Ptr(t *testing.T) {
	// --- Given ---
	src := Source{ptr: 123}

	// --- When ---
	have := src.Ptr()

	// --- Then ---
	assert.Equal(t, uintptr(123), have)
}

func Test_Source_IsZero(t *testing.T) {
	t.Run("zero", func(t *testing.T) {
		// --- Given ---
		src := Source{}

		// --- When ---
		have := src.IsZero()

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("not zero", func(t *testing.T) {
		// --- Given ---
		src := Source{val: 123}

		// --- When ---
		have := src.IsZero()

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_GetSrcPointer_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val     any
		success bool
	}{
		{"anonymous func", func() {}, true},
		{"func", GetSrcPointer, true},
		{"func std", math.Max, true},
		{"nil", nil, false},
		{"not supported", errors.New("test"), false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			rv := reflect.ValueOf(tc.val)

			// --- When ---
			have := GetSrcPointer(rv)

			// --- Then ---
			assert.Equal(t, tc.success, have > 0)
		})
	}
}
