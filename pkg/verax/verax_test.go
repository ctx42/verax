// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_Validate(t *testing.T) {
	t.Run("valid nil no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(iInt)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid pointer no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(pString)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid nil pointer no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(pStringNil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid nil slice no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(dSlice)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid nil array no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(dArray)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid nil map no rules", func(t *testing.T) {
		// --- When ---
		err := Validate(dMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid one rule", func(t *testing.T) {
		// --- When ---
		err := Validate("abc", StrRule("abc"))

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid both ok", func(t *testing.T) {
		// --- When ---
		err := Validate("abcxyz", StrContainRule("abc"), StrContainRule("xyz"))

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid first ok skip second", func(t *testing.T) {
		// --- When ---
		err := Validate("abc", StrRule("abc"), Skip, ErrRule("error"))

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid first ok skip when true second", func(t *testing.T) {
		// --- When ---
		err := Validate("abc", StrRule("abc"), Skip.When(true), StrRule("xyz"))

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid struct with validator", func(t *testing.T) {
		// --- Given ---
		s := &ModelPtr{"abc"}

		// --- When ---
		err := Validate(s)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("invalid struct with validator", func(t *testing.T) {
		// --- Given ---
		s := &ModelPtr{"xyz"}

		// --- When ---
		err := Validate(s)

		// --- Then ---
		xrrtest.AssertEqual(t, "FStr: must be 'abc' (ECMustAbc)", err)
	})

	t.Run("invalid first fails", func(t *testing.T) {
		// --- When ---
		err := Validate("123", StrRule("abc"), StrRule("xyz"))

		// --- Then ---
		xrrtest.AssertEqual(t, "must be 'abc' (ECMustAbc)", err)
	})

	t.Run("invalid second fails", func(t *testing.T) {
		// --- When ---
		err := Validate("abc", StrRule("abc"), StrRule("xyz"))

		// --- Then ---
		xrrtest.AssertEqual(t, "must be 'xyz' (ECMustXyz)", err)
	})

	t.Run("invalid first fails skip second", func(t *testing.T) {
		// --- When ---
		err := Validate("123", StrRule("abc"), Skip, ErrRule("error"))

		// --- Then ---
		xrrtest.AssertEqual(t, "must be 'abc' (ECMustAbc)", err)
	})

	t.Run("invalid first fails skip when true second", func(t *testing.T) {
		// --- When ---
		err := Validate(
			"123",
			StrRule("abc"),
			Skip.When(true),
			ErrRule("error"),
		)

		// --- Then ---
		xrrtest.AssertEqual(t, "must be 'abc' (ECMustAbc)", err)
	})

	t.Run("invalid first fail skip when false second", func(t *testing.T) {
		// --- When ---
		err := Validate(
			"123",
			StrRule("abc"),
			Skip.When(false),
			StrRule("xyz"),
		)

		// --- Then ---
		xrrtest.AssertEqual(t, "must be 'abc' (ECMustAbc)", err)
	})

	t.Run("invalid first ok skip when false second", func(t *testing.T) {
		// --- When ---
		err := Validate(
			"abc",
			StrRule("abc"),
			Skip.When(false),
			ErrRule("error"),
		)

		// --- Then ---
		xrrtest.AssertEqual(t, "error (ECGeneric)", err)
	})

	t.Run("invalid many in slice", func(t *testing.T) {
		// --- Given ---
		s := []*ModelPtr{{"xyz"}, {"xyz"}}

		// --- When ---
		err := Validate(s)

		// --- Then ---
		exp := "" +
			"0.FStr: must be 'abc' (ECMustAbc); " +
			"1.FStr: must be 'abc' (ECMustAbc)"
		xrrtest.AssertFieldsEqual(t, exp, err)
	})

	t.Run("valid all in slice", func(t *testing.T) {
		// --- Given ---
		s := []*ModelPtr{{"abc"}, {"abc"}}

		// --- When ---
		err := Validate(s)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("invalid many in map", func(t *testing.T) {
		// --- Given ---
		m := map[int]*ModelPtr{0: {"xyz"}, 4: {"xyz"}}

		// --- When ---
		err := Validate(m)

		// --- Then ---
		exp := "" +
			"0.FStr: must be 'abc' (ECMustAbc); " +
			"4.FStr: must be 'abc' (ECMustAbc)"
		xrrtest.AssertFieldsEqual(t, exp, err)
	})

	t.Run("valid all in map", func(t *testing.T) {
		// --- Given ---
		m := map[int]*ModelPtr{0: {"abc"}, 4: {"abc"}}

		// --- When ---
		err := Validate(m)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("ValidateWith success", func(t *testing.T) {
		// --- Given ---
		m := &ModelVW{"111"}

		// --- When ---
		err := Validate(m, StrRule("111"), NotNil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("ValidateWith runs provided rule", func(t *testing.T) {
		// --- Given ---
		m := &ModelVW{}

		// --- When ---
		err := Validate(m, StrRule("111"))

		// --- Then ---
		assert.Error(t, err)
		assert.Equal(t, "must be '111'", err.Error())
	})

	t.Run("ValidateWith additional validation in the type", func(t *testing.T) {
		// --- Given ---
		m := &ModelVW{"too_long"}

		// --- When ---
		err := Validate(m, StrRule("wrong_value"))

		// --- Then ---
		assert.ErrorIs(t, ErrTst, err)
	})

	t.Run("ValidateWith multiple rules", func(t *testing.T) {
		// --- Given ---
		m := &ModelVW{"too_long"}

		// --- When ---
		err := Validate(m, StrRule("wrong_value"))

		// --- Then ---
		assert.ErrorIs(t, ErrTst, err)
	})
}

func Test_ValidateName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		err := ValidateNamed("field", 42, Equal(42))

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		// --- When ---
		err := ValidateNamed("field", 43, Equal(42))

		// --- Then ---
		wMsg := "field: must be equal to '42' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})
}

func Test_Set(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		// --- Given ---
		r := Set{
			Min(40),
			Max(45),
		}

		// --- When ---
		err := Validate(42, r)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		// --- Given ---
		r := Set{
			Min(40),
			Max(45),
		}

		// --- When ---
		err := Validate(39, r)

		// --- Then ---
		assert.ErrorEqual(t, "must be no less than 40", err)
		xrrtest.AssertCode(t, ECInvThreshold, err)
	})
}

func Test_Named_Set_Get_GetOrNoop(t *testing.T) {
	t.Run("get set", func(t *testing.T) {
		// --- Given ---
		r1 := In(1)
		r2 := In(2)

		// --- When ---
		nr := NewNamed()
		nr.Set("r1", r1).Set("r2", r2)

		// --- Then ---
		g1 := nr.Get("r1")
		assert.NoError(t, g1.Validate(1))

		g2 := nr.Get("r2")
		assert.NoError(t, g2.Validate(2))
	})

	t.Run("get not present rule", func(t *testing.T) {
		// --- Given ---
		r1 := In(1)
		r2 := In(2)

		// --- When ---
		nr := NewNamed()
		nr.Set("r1", r1).Set("r2", r2)

		// --- Then ---
		assert.Nil(t, nr.Get("r3"))
	})

	t.Run("get safe existing", func(t *testing.T) {
		// --- Given ---
		r1 := In(1)
		r2 := In(2)

		// --- When ---
		nr := NewNamed()
		nr.Set("r1", r1).Set("r2", r2)

		// --- Then ---
		g2 := nr.GetOrNoop("r2")
		assert.NotNil(t, g2)
		assert.NoError(t, g2.Validate(2))
	})

	t.Run("get safe not existing", func(t *testing.T) {
		// --- Given ---
		r1 := In(1)
		r2 := In(2)

		// --- When ---
		nr := NewNamed()
		nr.Set("r1", r1).Set("r2", r2)

		// --- Then ---
		g3 := nr.GetOrNoop("r3")
		assert.NotNil(t, g3)
		assert.NoError(t, g3.Validate(3))
	})

	t.Run("set overrides existing", func(t *testing.T) {
		// --- Given ---
		r1 := In(1)
		r2 := In(2)

		// --- When ---
		nr := NewNamed()
		nr.Set("r1", r1).Set("r1", r2)

		// --- Then ---
		g1 := nr.Get("r1")
		assert.Error(t, g1.Validate(1))
		assert.NoError(t, g1.Validate(2))
	})
}

func Test_Named_GetOrError(t *testing.T) {
	t.Run("rule exists", func(t *testing.T) {
		// --- Given ---
		r1 := In(1, 11)
		r2 := In(2, 22)

		nr := NewNamed().Set("r1", r1).Set("r2", r2)

		// --- When ---
		have := nr.GetOrError("r1")

		// --- Then ---
		assert.NoError(t, have.Validate(1))
		assert.ErrorIs(t, ErrNotIn, have.Validate(2))
	})

	t.Run("rule does not exist", func(t *testing.T) {
		// --- Given ---
		r1 := In(1, 11)
		r2 := In(2, 22)

		nr := NewNamed().Set("r1", r1).Set("r2", r2)

		// --- When ---
		have := nr.GetOrError("r3")

		// --- Then ---
		assert.ErrorIs(t, ErrUnkRule, have.Validate(1))
	})
}
