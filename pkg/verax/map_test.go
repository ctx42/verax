// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_MapRule_valid(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		// --- When ---
		err := Map().Validate(dMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("nil key", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KpStrNil"),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty rules", func(t *testing.T) {
		// --- Given ---
		rs := make([]*KeyRules, 0)

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty key rules", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KStrAbc"),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty string rule", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KStrEmpty", Length(1, 5)),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("nil pointer rule", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KpStrNil", Length(1, 5)),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("multi key rules", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KStrAbc", StrRule("abc")),
			Key("KStrXyz", StrRule("xyz")),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("map key with slice", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KsString", Each(StrRule("abc"))),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("map of map", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KmStringString", Map(Key("foo", StrRule("abc")))),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("optional key", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("X").Optional(),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("skip key value validators", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KStructInvalid", Skip),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("skip true key value validators", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KStructInvalid", Skip.When(true)),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("skip required", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KpStrNil", Skip, Required),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("skip nil", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KpStructNil", Skip, NotNil),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("int keys", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key(1, StrRule("abc")),
			Key(3, StrRule("xyz")),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMapInt)

		// --- Then ---
		assert.NoError(t, err)
	})
}

func Test_MapRule_invalid(t *testing.T) {
	t.Run("not map", func(t *testing.T) {
		// --- When ---
		err := Map().Validate(123)

		// --- Then ---
		assert.ErrorIs(t, ErrNotMapPtr, err)
	})

	t.Run("two rules", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KStrAbc", StrRule("xyz")),
			Key("KpStr", StrRule("abc")),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		exp := "" +
			"KStrAbc: must be 'xyz' (ECMustXyz); " +
			"KpStr: must be 'abc' (ECMustAbc)"
		xrrtest.AssertFieldsEqual(t, exp, err)
	})

	t.Run("not matching key type", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key(123),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		wMsg := "123: key not the correct type (ECInternal)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("missing required key", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("X"),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		wMsg := "X: required key is missing (ECMapKeyMissing)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("run key value validators", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KStructInvalid"),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		wMsg := "KStructInvalid.FStr: must be 'abc' (ECMustAbc)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("skip false key value validators", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KStructInvalid", Skip.When(false)),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		wMsg := "KStructInvalid.FStr: must be 'abc' (ECMustAbc)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("required key", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KStrEmpty", Required),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		xrrtest.AssertEqual(t, "KStrEmpty: cannot be blank (ECRequired)", err)
	})

	t.Run("not nil key", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KpStrNil", NotNil),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		xrrtest.AssertEqual(t, "KpStrNil: is required (ECReqNotNil)", err)
	})

	t.Run("int keys", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key(1, StrRule("xyz")),
			Key(3, StrRule("abc")),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMapInt)

		// --- Then ---
		exp := "" +
			"1: must be 'xyz' (ECMustXyz); " +
			"3: must be 'abc' (ECMustAbc)"
		xrrtest.AssertFieldsEqual(t, exp, err)
	})

	t.Run("internal error", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KStrAbc", InternalErrRule),
		}

		// --- When ---
		err := Map(rs...).AllowUnknown().Validate(TMap)

		// --- Then ---
		xrrtest.AssertEqual(t, "KStrAbc: internal error (ECInternal)", err)
	})

	t.Run("dont allow unknown keys", func(t *testing.T) {
		// --- Given ---
		rs := []*KeyRules{
			Key("KStrAbc", StrRule("abc")),
		}

		// --- When ---
		err := Map(rs...).Validate(TMap)

		// --- Then ---
		want := "" +
			"KStrEmpty: key not expected (ECMapKeyUnexpected); " +
			"KStrXyz: key not expected (ECMapKeyUnexpected); " +
			"KStructInvalid: key not expected (ECMapKeyUnexpected); " +
			"KStructValid: key not expected (ECMapKeyUnexpected); " +
			"KmStringString: key not expected (ECMapKeyUnexpected); " +
			"KpStr: key not expected (ECMapKeyUnexpected); " +
			"KpStrNil: key not expected (ECMapKeyUnexpected); " +
			"KpStructNil: key not expected (ECMapKeyUnexpected); " +
			"KsString: key not expected (ECMapKeyUnexpected)"
		xrrtest.AssertFieldsEqual(t, want, err)
	})
}

func Test_MapRule_IsOptional(t *testing.T) {
	// --- Given ---
	rs := []*KeyRules{
		Key(1, StrRule("xyz")).Optional(),
		Key(3, StrRule("abc")),
	}
	mr := Map(rs...)

	// --- Then ---
	assert.True(t, mr.IsOptional(1))
	assert.True(t, mr.IsOptional(2))
	assert.False(t, mr.IsOptional(3))
	assert.True(t, mr.IsOptional("abc"))
	assert.True(t, mr.IsOptional(nil))
}

func Test_MapRule_IsDefined(t *testing.T) {
	// --- Given ---
	rs := []*KeyRules{
		Key(1, StrRule("xyz")).Optional(),
		Key(3, StrRule("abc")),
	}
	mr := Map(rs...)

	// --- Then ---
	assert.True(t, mr.IsDefined(1))
	assert.False(t, mr.IsDefined(2))
	assert.False(t, mr.IsDefined("abc"))
	assert.False(t, mr.IsDefined(nil))
}

func Test_KeyRules(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		// --- When ---
		kr := Key(1, Noop)

		// --- Then ---
		assert.Equal(t, 1, kr.key)
		assert.False(t, kr.optional)
		assert.Len(t, 1, kr.rules)
	})

	t.Run("optional", func(t *testing.T) {
		// --- When ---
		kr := Key(1, Noop).Optional()

		// --- Then ---
		assert.True(t, kr.optional)
	})

	t.Run("required when true", func(t *testing.T) {
		// --- When ---
		kr := Key(1, Noop).RequiredWhen(true)

		// --- Then ---
		assert.False(t, kr.optional)
	})

	t.Run("required when false", func(t *testing.T) {
		// --- When ---
		kr := Key(1, Noop).RequiredWhen(false)

		// --- Then ---
		assert.True(t, kr.optional)
	})
}
