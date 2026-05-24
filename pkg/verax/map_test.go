// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_Map(t *testing.T) {
	// --- Given ---
	mks := []MapKey{
		Key("A", Max(1)),
		Key("B", Max(2)),
	}

	// --- When ---
	have := Map(mks...)

	// --- Then ---
	want := MapRule{
		condition:    true,
		allowUnknown: false,
		keys: []MapKey{
			Key("A", Max(1)),
			Key("B", Max(2)),
		},
	}
	assert.Equal(t, want, have)
}

func Test_MapRule_AllowUnknown(t *testing.T) {
	// --- Given ---
	r := MapRule{}

	// --- When ---
	have := r.AllowUnknown()

	// --- Then ---
	assert.True(t, have.allowUnknown)
}

func Test_MapRule_IsOptional(t *testing.T) {
	t.Run("existing keys", func(t *testing.T) {
		// --- Given ---
		mks := []MapKey{
			Key("A", Min(42)).Optional(),
			Key("C", Max(42)),
		}
		r := Map(mks...)

		// --- Then ---
		assert.True(t, r.IsOptional("A"))
		assert.False(t, r.IsOptional("C"))
	})

	t.Run("not existing key", func(t *testing.T) {
		// --- Given ---
		r := MapRule{}

		// --- When ---
		have := r.IsOptional(42)

		// --- Then ---
		assert.True(t, have)
	})
}

func Test_MapRule_IsDefined(t *testing.T) {
	t.Run("defined", func(t *testing.T) {
		// --- Given ---
		r := Map(Key(1, Min(42)))

		// --- Then ---
		assert.True(t, r.IsDefined(1))
	})

	t.Run("not defined", func(t *testing.T) {
		// --- Given ---
		r := Map(Key(1, Min(42)))

		// --- Then ---
		assert.False(t, r.IsDefined(3))
	})
}

func Test_MapRule_validate_valid(t *testing.T) {
	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		rs := Map(Key("abc", Max(42))).When(false)

		// --- When ---
		err := rs.Validate(map[string]any{"abc": 44})

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("nil map", func(t *testing.T) {
		// --- When ---
		err := Map().Validate(dMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("pointer to map", func(t *testing.T) {
		// --- Given ---
		m := map[string]int{"A": 1}

		// --- When ---
		err := Map().AllowUnknown().Validate(&m)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("nil key", func(t *testing.T) {
		// --- Given ---
		kr := Key("KpStrNil")

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty rules", func(t *testing.T) {
		// --- Given ---
		mks := make([]MapKey, 0)

		// --- When ---
		err := Map(mks...).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty key rules", func(t *testing.T) {
		// --- Given ---
		kr := Key("KStrAbc")

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty string rule", func(t *testing.T) {
		// --- Given ---
		kr := Key("KStrEmpty", Length(1, 5))

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("nil pointer rule", func(t *testing.T) {
		// --- Given ---
		kr := Key("KpStrNil", Length(1, 5))

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("multi key rules", func(t *testing.T) {
		// --- Given ---
		kr0 := Key("KStrAbc", Equal("abc"))
		kr1 := Key("KStrXyz", Equal("xyz"))

		// --- When ---
		err := Map(kr0, kr1).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("map key with slice", func(t *testing.T) {
		// --- Given ---
		kr := Key("KsString", Each(Equal("abc")))

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("map of map", func(t *testing.T) {
		// --- Given ---
		kr := Key("KmStringString", Map(Key("foo", Equal("abc"))))

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("optional key", func(t *testing.T) {
		// --- Given ---
		kr := Key("X").Optional()

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("skip key value validators", func(t *testing.T) {
		// --- Given ---
		kr := Key("KStructInvalid", Skip)

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("skip true key value validators", func(t *testing.T) {
		// --- Given ---
		kr := Key("KStructInvalid", Skip.When(true))

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("skip required", func(t *testing.T) {
		// --- Given ---
		kr := Key("KpStrNil", Skip, Required)

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("skip nil", func(t *testing.T) {
		// --- Given ---
		kr := Key("KpStructNil", Skip, NotNil)

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("int keys", func(t *testing.T) {
		// --- Given ---
		kr0 := Key(1, Equal("abc"))
		kr1 := Key(3, Equal("xyz"))

		// --- When ---
		err := Map(kr0, kr1).AllowUnknown().Validate(TMapInt)

		// --- Then ---
		assert.NoError(t, err)
	})
}

func Test_MapRule_Validate_invalid(t *testing.T) {
	t.Run("not map", func(t *testing.T) {
		// --- When ---
		err := Map().Validate(123)

		// --- Then ---
		assert.ErrorIs(t, ErrNotMapPtr, err)
	})

	t.Run("two rules", func(t *testing.T) {
		// --- Given ---
		kr0 := Key("KStrAbc", Equal("xyz"))
		kr1 := Key("KpStr", Equal("abc"))

		// --- When ---
		err := Map(kr0, kr1).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		want := "" +
			"KStrAbc: must be equal to 'xyz' (ECNotEqual); " +
			"KpStr: must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, want, err)
	})

	t.Run("not matching key type", func(t *testing.T) {
		// --- Given ---
		kr := Key(123)

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "123: the key type does not match the map (ECInternal)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("missing required key", func(t *testing.T) {
		// --- Given ---
		kr := Key("X")

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "X: missing key (ECMapKeyMissing)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("run key value validators", func(t *testing.T) {
		// --- Given ---
		kr := Key("KStructInvalid")

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "KStructInvalid.FStr: must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("skip false key value validators", func(t *testing.T) {
		// --- Given ---
		kr := Key("KStructInvalid", Skip.When(false))

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "KStructInvalid.FStr: must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("required key", func(t *testing.T) {
		// --- Given ---
		kr := Key("KStrEmpty", Required)

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		xrrtest.AssertEqual(t, "KStrEmpty: cannot be blank (ECRequired)", err)
	})

	t.Run("not nil key", func(t *testing.T) {
		// --- Given ---
		kr := Key("KpStrNil", NotNil)

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		xrrtest.AssertEqual(t, "KpStrNil: is required (ECReqNotNil)", err)
	})

	t.Run("int keys", func(t *testing.T) {
		// --- Given ---
		kr0 := Key(1, Equal("xyz"))
		kr1 := Key(3, Equal("abc"))

		// --- When ---
		err := Map(kr0, kr1).AllowUnknown().Validate(TMapInt)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		want := "" +
			"1: must be equal to 'xyz' (ECNotEqual); " +
			"3: must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, want, err)
	})

	t.Run("error", func(t *testing.T) {
		// --- Given ---
		kr := Key("KStrAbc", Fail("test err", "ECTst"))

		// --- When ---
		err := Map(kr).AllowUnknown().Validate(TMap)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		xrrtest.AssertEqual(t, "KStrAbc: test err (ECTst)", err)
	})

	t.Run("do not allow unknown keys", func(t *testing.T) {
		// --- Given ---
		kr := Key("KStrAbc", Equal("abc"))

		// --- When ---
		err := Map(kr).Validate(TMap)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		want := "" +
			"KStrEmpty: unexpected key (ECMapKeyUnexpected); " +
			"KStrXyz: unexpected key (ECMapKeyUnexpected); " +
			"KStructInvalid: unexpected key (ECMapKeyUnexpected); " +
			"KStructValid: unexpected key (ECMapKeyUnexpected); " +
			"KmStringString: unexpected key (ECMapKeyUnexpected); " +
			"KpStr: unexpected key (ECMapKeyUnexpected); " +
			"KpStrNil: unexpected key (ECMapKeyUnexpected); " +
			"KpStructNil: unexpected key (ECMapKeyUnexpected); " +
			"KsString: unexpected key (ECMapKeyUnexpected)"
		xrrtest.AssertEqual(t, want, err)
	})
}

func Test_MapRule_When(t *testing.T) {
	// --- Given ---
	r := MapRule{}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_MapRule_Spec(t *testing.T) {
	t.Run("without rules nor allowUnknown", func(t *testing.T) {
		// --- Given ---
		r := Map()

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, MapRuleName, have.Name)
		assert.Nil(t, have.Args)
	})

	t.Run("with rules", func(t *testing.T) {
		// --- Given ---
		r := Map(
			Key(1, Min(42)),
			Key(3, Max(44)),
		)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, MapRuleName, have.Name)
		assert.Len(t, 1, have.Args)
		wRules := []*spec.Spec{
			must.Value(Key(1, Min(42)).Spec()),
			must.Value(Key(3, Max(44)).Spec()),
		}
		hRules, _ := assert.HasKey(t, spec.ArgSpecs, have.Args)
		assert.Equal(t, wRules, hRules)
	})

	t.Run("without rules", func(t *testing.T) {
		// --- Given ---
		r := Map().AllowUnknown()

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, MapRuleName, have.Name)
		assert.Len(t, 1, have.Args)
		assert.HasKeyValue(t, ArgAllowUnk, true, have.Args)
	})

	t.Run("with rules and AllowUnknown", func(t *testing.T) {
		// --- Given ---
		r := Map(Key(1, Min(42))).AllowUnknown()

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, MapRuleName, have.Name)
		assert.Len(t, 2, have.Args)
		assert.HasKeyValue(t, ArgAllowUnk, true, have.Args)
		wRules := []*spec.Spec{must.Value(Key(1, Min(42)).Spec())}
		hRules, _ := assert.HasKey(t, spec.ArgSpecs, have.Args)
		assert.Equal(t, wRules, hRules)
	})

	t.Run("Map - JSON encode", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()
		spc := must.Value(Map(Key("name", Required)).Spec())

		// --- When ---
		data := must.Value(reg.EncodeSpec(spc))

		// --- Then ---
		want := `{
			"name": "map-rule",
			"args": {
				"specs": [
					{
						"name": "map-key",
						"args": {
							"value": "name",
							"types": [
								{
									"name": "required-rule",
									"args": {"mode": "required"}
								}
							]
						}
					}
				]
			}
		}`
		assert.JSON(t, want, data)
	})

	t.Run("Map - JSON decode", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()
		reg.RegisterBuilders(Builders())
		data := []byte(`{
			"name": "map-rule",
			"args": {
				"specs": [
					{
						"name": "map-key",
						"args": {
							"value": "name",
							"types": [
								{
									"name": "required-rule",
									"args": {"mode": "required"}
								}
							]
						}
					}
				]
			}
		}`)

		// --- When ---
		have, err := reg.DecodeAndBuild(data)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, Map(Key("name", Required)), have)
	})
}

func Test_MapRuleFromSpec(t *testing.T) {
	t.Run("error - not map rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := MapRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `map-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("rule without keys", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MapRuleName)

		// --- When ---
		have, err := MapRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.SameType(t, MapRule{}, have)
		wRule := Map()
		assert.Equal(t, have, wRule)
	})

	t.Run("error - specs argument not spec instances", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MapRuleName).SetArg(spec.ArgSpecs, "bad-type")

		// --- When ---
		have, err := MapRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `map-rule: spec argument "specs" must be []*spec.Spec, got string`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - key spec error", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MapRuleName).
			SetArg(spec.ArgSpecs, []*spec.Spec{spec.NewSpec("bad-key-spec")})

		// --- When ---
		have, err := MapRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `map-rule: key-spec[0]: map-key: invalid spec name: "bad-key-spec"`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("rule with keys", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MapRuleName).
			SetArg(
				spec.ArgSpecs,
				[]*spec.Spec{
					spec.NewSpec(MapKeyName).
						SetArg(spec.ArgValue, 1).
						SetArg(spec.ArgTypes, []Rule{Min(1), Max(11)}),
					spec.NewSpec(MapKeyName).
						SetArg(spec.ArgValue, 3).
						SetArg(spec.ArgTypes, []Rule{Min(3), Max(33)}),
				},
			)

		// --- When ---
		have, err := MapRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.SameType(t, MapRule{}, have)
		wRule := Map(
			Key(1, Min(1), Max(11)),
			Key(3, Min(3), Max(33)),
		)
		assert.Equal(t, have, wRule)
	})

	t.Run("error - the argument allowing unknown not bool", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MapRuleName).
			SetArg(ArgAllowUnk, "bad-value")

		// --- When ---
		have, err := MapRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `map-rule: spec argument "allow_unknown" must be bool, got string`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("rule with argument allow unknown", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MapRuleName).SetArg(ArgAllowUnk, true)

		// --- When ---
		have, err := MapRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.SameType(t, MapRule{}, have)
		wRule := Map().AllowUnknown()
		assert.Equal(t, have, wRule)
	})
}

func Test_Key(t *testing.T) {
	t.Run("with rules", func(t *testing.T) {
		// --- When ---
		have := Key(1, Noop)

		// --- Then ---
		want := MapKey{
			key:      1,
			optional: false,
			rules:    []Rule{Noop},
		}
		assert.Equal(t, want, have)
	})

	t.Run("without rules", func(t *testing.T) {
		// --- When ---
		have := Key(1)

		// --- Then ---
		want := MapKey{
			key:      1,
			optional: false,
			rules:    nil,
		}
		assert.Equal(t, want, have)
	})
}

func Test_MapKey_Optional(t *testing.T) {
	// --- Given ---
	r := MapKey{}

	// --- When ---
	have := r.Optional()

	// --- Then ---
	assert.True(t, have.optional)
}

func Test_MapKey_When(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		// --- Given ---
		r := MapKey{}

		// --- When ---
		have := r.When(true)

		// --- Then ---
		assert.False(t, have.optional)
	})

	t.Run("false", func(t *testing.T) {
		// --- Given ---
		r := MapKey{}

		// --- When ---
		have := r.When(false)

		// --- Then ---
		assert.True(t, have.optional)
	})
}

func Test_MapKey_KeyString(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		// --- Given ---
		r := Key(nil)

		// --- When ---
		have := r.KeyString()

		// --- Then ---
		assert.Equal(t, "<nil>", have)
	})

	t.Run("string", func(t *testing.T) {
		// --- Given ---
		r := Key("abc")

		// --- When ---
		have := r.KeyString()

		// --- Then ---
		assert.Equal(t, "abc", have)
	})

	t.Run("int", func(t *testing.T) {
		// --- Given ---
		r := Key(42)

		// --- When ---
		have := r.KeyString()

		// --- Then ---
		assert.Equal(t, "42", have)
	})
}

func Test_MapKey_Spec(t *testing.T) {
	t.Run("with rules", func(t *testing.T) {
		// --- Given ---
		r := Key(1, Min(42), Noop, Min(44))

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, MapKeyName, have.Name)
		assert.Len(t, 2, have.Args)
		assert.HasKeyValue(t, spec.ArgValue, 1, have.Args)
		wSpecs := []Rule{Min(42), Noop, Min(44)}
		hSpecs, _ := assert.HasKey(t, spec.ArgTypes, have.Args)
		assert.Equal(t, wSpecs, hSpecs)
	})

	t.Run("without rules", func(t *testing.T) {
		// --- Given ---
		r := Key(1)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, MapKeyName, have.Name)
		assert.Len(t, 1, have.Args)
		assert.HasKeyValue(t, spec.ArgValue, 1, have.Args)
	})

	t.Run("optional", func(t *testing.T) {
		// --- Given ---
		r := Key(1).Optional()

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, MapKeyName, have.Name)
		assert.Len(t, 2, have.Args)
		assert.HasKeyValue(t, spec.ArgValue, 1, have.Args)
		assert.HasKeyValue(t, ArgOptional, true, have.Args)
	})

	t.Run("MapKey - JSON encode", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()
		spc := must.Value(Key("name", Required).Spec())

		// --- When ---
		data := must.Value(reg.EncodeSpec(spc))

		// --- Then ---
		want := `{
			"name": "map-key",
			"args": {
				"value": "name",
				"types": [
					{
						"name": "required-rule",
						"args": {
							"mode": "required"
						}
					}
				]
			}
		}`
		assert.JSON(t, want, data)
	})

	t.Run("MapKey - JSON decode", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()
		reg.RegisterBuilders(Builders())
		data := []byte(`{
			"name": "map-key",
			"args": {
				"value": "name",
				"types": [
					{
						"name": "required-rule",
						"args": {
							"mode": "required"
						}
					}
				]
			}
		}`)

		// --- When ---
		spc := &spec.Spec{}
		err := reg.DecodeSpec(data, spc)

		// --- Then ---
		assert.NoError(t, err)
		have, err := MapKeyFromSpec(spc)
		assert.NoError(t, err)
		assert.Equal(t, Key("name", Required), have)
	})
}

func Test_MapKeyFromSpec(t *testing.T) {
	t.Run("error - not key rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := MapKeyFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `map-key: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - want argument is required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MapKeyName)

		// --- When ---
		have, err := MapKeyFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "map-key: spec missing required argument: value"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - types argument is required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MapKeyName).SetArg(spec.ArgValue, 1)

		// --- When ---
		have, err := MapKeyFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "map-key: spec missing required argument: types"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("success - int key", func(t *testing.T) {
		// --- Given ---
		rules := []Rule{Min(42), Max(44)}
		spc := spec.NewSpec(MapKeyName).
			SetArg(spec.ArgValue, 1).
			SetArg(spec.ArgTypes, rules)

		// --- When ---
		have, err := MapKeyFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		rule := Key(1, Min(42), Max(44))
		assert.Equal(t, rule, have)
	})

	t.Run("success - uint32 key", func(t *testing.T) {
		// --- Given ---
		rules := []Rule{Min(uint8(42)), Max(uint8(44))}
		spc := spec.NewSpec(MapKeyName).
			SetArg(spec.ArgValue, uint32(1)).
			SetArg(spec.ArgTypes, rules)

		// --- When ---
		have, err := MapKeyFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		rule := Key(uint32(1), Min(uint8(42)), Max(uint8(44)))
		assert.Equal(t, rule, have)
	})

	t.Run("success - optional", func(t *testing.T) {
		// --- Given ---
		rules := []Rule{Min(42), Max(44)}
		spc := spec.NewSpec(MapKeyName).
			SetArg(spec.ArgValue, 1).
			SetArg(spec.ArgTypes, rules).
			SetArg(ArgOptional, true)

		// --- When ---
		have, err := MapKeyFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		rule := Key(1, Min(42), Max(44)).Optional()
		assert.Equal(t, rule, have)
	})

	t.Run("error - argument optional is not bool", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MapKeyName).
			SetArg(spec.ArgValue, 1).
			SetArg(spec.ArgTypes, []Rule{Min(42), Max(44)}).
			SetArg(ArgOptional, "true")

		// --- When ---
		have, err := MapKeyFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `map-key: spec argument "optional" must be bool, got string`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})
}

func Test_MapRule_Spec_MapRuleFromSpec_round_trip(t *testing.T) {
	// --- Given ---
	fn := RuleFunc(func(v any) error { return nil })
	want := Map(
		Key("A", Min(42), Required, By(fn)),
	)
	spc := must.Value(want.Spec())

	// --- When ---
	have, err := MapRuleFromSpec(spc)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, want, have)
}
