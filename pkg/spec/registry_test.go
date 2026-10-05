// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package spec

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/ctx42/convert/pkg/convert"
	"github.com/ctx42/jsontype/pkg/jsontype"
	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
)

func Test_NewRegistry(t *testing.T) {
	// --- When ---
	reg := NewRegistry[TstType]()

	// --- Then ---
	assert.NotNil(t, reg.jtr.Converter(jsontype.Byte))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Uint8))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Uint16))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Uint32))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Uint64))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Uint))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Int))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Int8))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Int16))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Int32))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Int64))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Float32))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Float64))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Rune))
	assert.NotNil(t, reg.jtr.Converter(jsontype.String))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Bool))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Time))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Duration))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Nil))
	assert.Len(t, 0, reg.sources)
	assert.NotNil(t, reg.sources)
	assert.Len(t, 0, reg.builders)
	assert.NotNil(t, reg.builders)
}

func Test_Registry_RegisterSource(t *testing.T) {
	t.Run("register new", func(t *testing.T) {
		// --- Given ---
		fn := func() {}
		src := must.Value(NewSource("fn", fn))
		reg := NewRegistry[TstType]()

		// --- When ---
		have := reg.RegisterSource(src)

		// --- Then ---
		assert.Zero(t, have)
	})

	t.Run("register already existing", func(t *testing.T) {
		// --- Given ---
		fnOld := func() {}
		fnNew := func() {}
		srcOld := must.Value(NewSource("fn", fnOld))
		srcNew := must.Value(NewSource("fn", fnNew))

		reg := NewRegistry[TstType]()
		reg.RegisterSource(srcOld)

		// --- When ---
		have := reg.RegisterSource(srcNew)

		// --- Then ---
		assert.Equal(t, srcOld, have)
		assert.Equal(t, []Source{srcNew}, reg.sources)
	})
}

func Test_Registry_SourceByName(t *testing.T) {
	t.Run("not existing", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()

		// --- When ---
		have := reg.SourceByName("name")

		// --- Then ---
		assert.Zero(t, have)
	})

	t.Run("existing source", func(t *testing.T) {
		// --- Given ---
		fn0 := func() {}
		fn1 := func() {}
		src0 := must.Value(NewSource("fn0", fn0))
		src1 := must.Value(NewSource("fn1", fn1))

		reg := NewRegistry[TstType]()
		reg.RegisterSource(src0)
		reg.RegisterSource(src1)

		// --- When ---
		have := reg.SourceByName(src1.Name)

		// --- Then ---
		assert.Equal(t, src1, have)
	})
}

func Test_Registry_SourceByValue(t *testing.T) {
	t.Run("not existing", func(t *testing.T) {
		// --- Given ---
		fn := func() {}
		reg := NewRegistry[TstType]()

		// --- When ---
		have := reg.SourceByValue(fn)

		// --- Then ---
		assert.Zero(t, have)
	})

	t.Run("non-pointer type", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()

		// --- When ---
		have := reg.SourceByValue(TstType{})

		// --- Then ---
		assert.Zero(t, have)
	})

	t.Run("existing source", func(t *testing.T) {
		// --- Given ---
		fn0 := func() {}
		fn1 := func() {}
		src0 := must.Value(NewSource("fn0", fn0))
		src1 := must.Value(NewSource("fn1", fn1))

		reg := NewRegistry[TstType]()
		reg.RegisterSource(src0)
		reg.RegisterSource(src1)

		// --- When ---
		have := reg.SourceByValue(fn1)

		// --- Then ---
		assert.Equal(t, src1, have)
	})
}

func Test_Registry_RegisterBuilder(t *testing.T) {
	t.Run("register", func(t *testing.T) {
		// --- Given ---
		fn := func(*Spec) (TstType, error) { return TstType{}, nil }
		reg := NewRegistry[TstType]()

		// --- When ---
		have := reg.RegisterBuilder("name", fn)

		// --- Then ---
		assert.Nil(t, have)
		assert.Equal(t, map[string]TstBuilder{"name": fn}, reg.builders)
	})

	t.Run("register already existing", func(t *testing.T) {
		// --- Given ---
		fnOld := func(*Spec) (TstType, error) { return TstType{}, nil }
		fnNew := func(*Spec) (TstType, error) { return TstType{}, nil }
		reg := NewRegistry[TstType]()
		reg.RegisterBuilder("name", fnOld)

		// --- When ---
		have := reg.RegisterBuilder("name", fnNew)

		// --- Then ---
		assert.Same(t, fnOld, have)
		assert.Equal(t, map[string]TstBuilder{"name": fnNew}, reg.builders)
	})

	t.Run("zero value registry", func(t *testing.T) {
		// --- Given ---
		fn := func(*Spec) (TstType, error) { return TstType{}, nil }
		reg := &Registry[TstType]{}

		// --- When ---
		have := reg.RegisterBuilder("name", fn)

		// --- Then ---
		assert.Nil(t, have)
		assert.Equal(t, map[string]TstBuilder{"name": fn}, reg.builders)
	})

	t.Run("remove builder", func(t *testing.T) {
		// --- Given ---
		fn := func(*Spec) (TstType, error) { return TstType{}, nil }
		reg := NewRegistry[TstType]()
		reg.RegisterBuilder("name", fn)

		// --- When ---
		have := reg.RegisterBuilder("name", nil)

		// --- Then ---
		assert.Same(t, fn, have)
		assert.Empty(t, reg.builders)
	})
}

func Test_Registry_RegisterBuilders(t *testing.T) {
	t.Run("register all new", func(t *testing.T) {
		// --- Given ---
		b0 := func(*Spec) (TstType, error) { return TstType{}, nil }
		b1 := func(*Spec) (TstType, error) { return TstType{}, nil }
		bls := map[string]TstBuilder{"b0": b0, "b1": b1}
		reg := NewRegistry[TstType]()

		// --- When ---
		have := reg.RegisterBuilders(bls)

		// --- Then ---
		assert.Equal(t, map[string]TstBuilder{"b0": b0, "b1": b1}, reg.builders)
		assert.Equal(t, map[string]TstBuilder{"b0": nil, "b1": nil}, have)
	})

	t.Run("register some existing", func(t *testing.T) {
		// --- Given ---
		b0 := func(*Spec) (TstType, error) { return TstType{}, nil }
		b1 := func(*Spec) (TstType, error) { return TstType{}, nil }
		bls := map[string]TstBuilder{"b0": b0, "b1": b1}
		reg := NewRegistry[TstType]()
		reg.RegisterBuilder("b0", b0)

		// --- When ---
		have := reg.RegisterBuilders(bls)

		// --- Then ---
		assert.Equal(t, map[string]TstBuilder{"b0": b0, "b1": b1}, reg.builders)
		assert.Equal(t, map[string]TstBuilder{"b0": b0, "b1": nil}, have)
	})

	t.Run("remove some builders", func(t *testing.T) {
		// --- Given ---
		b0 := func(*Spec) (TstType, error) { return TstType{}, nil }
		b1 := func(*Spec) (TstType, error) { return TstType{}, nil }
		bls := map[string]TstBuilder{"b0": nil, "b1": b1}
		reg := NewRegistry[TstType]()
		reg.RegisterBuilder("b0", b0)

		// --- When ---
		have := reg.RegisterBuilders(bls)

		// --- Then ---
		assert.Equal(t, map[string]TstBuilder{"b1": b1}, reg.builders)
		assert.Equal(t, map[string]TstBuilder{"b0": b0, "b1": nil}, have)
	})
}

func Test_Registry_BuilderFor(t *testing.T) {
	t.Run("not existing", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()

		// --- When ---
		have := reg.BuilderFor("name")

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("existing", func(t *testing.T) {
		// --- Given ---
		fn := func(*Spec) (TstType, error) { return TstType{}, nil }
		reg := NewRegistry[TstType]()
		reg.RegisterBuilder("name", fn)

		// --- When ---
		have := reg.BuilderFor("name")

		// --- Then ---
		assert.Same(t, fn, have)
	})
}

func Test_Registry_Build(t *testing.T) {
	t.Run("error - no builder registered", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.Build(NewSpec("my-spec"))

		// --- Then ---
		assert.ErrorIs(t, ErrUnkBuilder, err)
		assert.ErrorEqual(t, "unknown builder my-spec", err)
		assert.Zero(t, have)
	})

	t.Run("error - nil spec", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.Build(nil)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSpec, err)
		assert.ErrorEqual(t, "invalid spec", err)
		assert.Zero(t, have)
	})

	t.Run("error - builder returns error", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()
		reg.RegisterBuilder("my-spec", func(*Spec) (TstType, error) {
			return TstType{}, ErrTst
		})

		// --- When ---
		have, err := reg.Build(NewSpec("my-spec"))

		// --- Then ---
		assert.ErrorIs(t, ErrTst, err)
		assert.ErrorEqual(t, "test msg", err)
		assert.Zero(t, have)
	})

	t.Run("builds value", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()
		reg.RegisterBuilder("my-spec", func(*Spec) (TstType, error) {
			return TstType{"built"}, nil
		})

		// --- When ---
		have, err := reg.Build(NewSpec("my-spec"))

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, TstType{"built"}, have)
	})
}

func Test_Registry_EncodeSpec(t *testing.T) {
	t.Run("error - nil spec", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(nil)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSpec, err)
		assert.ErrorEqual(t, "invalid spec", err)
		assert.Nil(t, have)
	})

	t.Run("error - empty name", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(&Spec{})

		// --- Then ---
		assert.ErrorIs(t, ErrInvSpec, err)
		assert.ErrorEqual(t, "spec to JSON: empty name: invalid spec", err)
		assert.Nil(t, have)
	})

	t.Run("error - specs argument", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec").SetArg(ArgSpecs, true)
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArgType, err)
		wMsg := "spec to JSON: spec my-spec, argument specs: " +
			"invalid spec argument type"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("specs argument", func(t *testing.T) {
		// --- Given ---
		sub := []*Spec{NewSpec("s0"), NewSpec("s1")}
		spc := NewSpec("my-spec").SetArg(ArgSpecs, sub)
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		want := `{
			"name": "my-spec",
				"args": {
					"specs": [
						{"name": "s0"},
						{"name": "s1"}
					]
				}
			}`
		assert.JSON(t, want, string(have))
	})

	t.Run("error - types argument", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec").SetArg(ArgTypes, true)
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArgType, err)
		wMsg := "spec to JSON: spec my-spec, argument types: " +
			"invalid spec argument type"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("types argument", func(t *testing.T) {
		// --- Given ---
		sub := []TstSpec{
			{name: "name0"},
			{name: "name1"},
		}
		spc := NewSpec("my-spec").SetArg(ArgTypes, sub)
		reg := NewRegistry[TstSpec]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		want := `{
			"name": "my-spec",
			"args": {
				"types": [
					{"name": "name0"},
					{"name": "name1"}
				]
			}
		}`
		assert.JSON(t, want, string(have))
	})

	t.Run("error - source argument", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec").SetArg(ArgSrc, true)
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.ErrorIs(t, ErrUnkSource, err)
		wMsg := "spec to JSON: spec my-spec, argument src_go: unknown source"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("source argument", func(t *testing.T) {
		// --- Given ---
		src := must.Value(NewSource("my-fn", TstFn0))
		spc := NewSpec("my-spec").SetArg(ArgSrc, TstFn0)
		reg := NewRegistry[TstType]()
		reg.RegisterSource(src)

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		want := `{
			"name":"my-spec",
			"args": {
				"src_go": {
					"lang":"go",
					"name":"my-fn",
					"src":"github.com/ctx42/verax/pkg/spec.TstFn0"
				}
			}
		}`
		assert.JSON(t, want, string(have))
	})

	t.Run("error - values argument", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec").SetArg(ArgValues, true)
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArgType, err)
		wMsg := "spec to JSON: spec my-spec, argument values: " +
			"invalid spec argument type"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("values argument", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec").
			SetArg(ArgValues, []any{1, 2.6, nil})
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		want := `{
				"name":"my-spec",
				"args": {
					"values": [
						{"type": "int", "value": 1},
						2.6,
						null
					]
				}
			}`
		assert.JSON(t, want, string(have))
	})

	t.Run("error - with not reserved argument name", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec").SetArg("custom", TstFn0)
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.ErrorIs(t, convert.ErrUnsType, err)
		wMsg := "spec to JSON: spec my-spec, argument custom: " +
			"jsontype: unsupported type: func()"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("error - argument type not in registry", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec").SetArg("int", 1)
		reg := NewRegistry[TstType]()
		reg.jtr = jsontype.NewRegistry()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.ErrorIs(t, convert.ErrUnsType, err)
		assert.Nil(t, have)
	})

	t.Run("not reserved argument names", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec").
			SetArg("int", 1).
			SetArg("uint", uint(2)).
			SetArg("float", 3.0).
			SetArg("bool", true).
			SetArg("time", time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC))
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		want := `{
			"name":"my-spec",
			"args": {
				"int": {"type": "int", "value": 1},
				"uint": {"type": "uint", "value": 2},
				"float": 3,
				"bool": true,
				"time": {"type": "time.Time", "value": "2000-01-02T03:04:05Z"}
			}
		}`
		assert.JSON(t, want, string(have))
	})

	t.Run("empty list arguments", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec").
			SetArg(ArgSpecs, []*Spec{}).
			SetArg(ArgTypes, []TstSpec{}).
			SetArg(ArgValues, []any{})
		reg := NewRegistry[TstSpec]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		want := `{
			"name": "my-spec",
			"args": {"specs": [], "types": [], "values": []}
		}`
		assert.JSON(t, want, string(have))
	})

	t.Run("does not mutate input", func(t *testing.T) {
		// --- Given ---
		src := must.Value(NewSource("my-fn", TstFn0))
		spc := NewSpec("my-spec").
			SetArg("int", 1).
			SetArg(ArgSpecs, []*Spec{NewSpec("sub").SetArg("int", 2)}).
			SetArg(ArgSrc, TstFn0).
			SetArg(ArgValues, []any{3})
		reg := NewRegistry[TstType]()
		reg.RegisterSource(src)

		// --- When ---
		_, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{
			Name: "my-spec",
			Args: map[string]any{
				"int": 1,
				ArgSpecs: []*Spec{
					{Name: "sub", Args: map[string]any{"int": 2}},
				},
				ArgSrc:    TstFn0,
				ArgValues: []any{3},
			},
		}
		assert.Equal(t, want, spc)
	})

	t.Run("no args spec", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec")
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.JSON(t, `{"name":"my-spec"}`, string(have))
	})

	t.Run("error - nested sub-spec encode error", func(t *testing.T) {
		// --- Given ---
		// TstFn0 is func() — not serializable.
		sub := NewSpec("sub").SetArg("custom", TstFn0)
		spc := NewSpec("my-spec").SetArg(ArgSpecs, []*Spec{sub})
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.ErrorIs(t, convert.ErrUnsType, err)
		wMsg := "spec to JSON: spec my-spec, argument specs: " +
			"index 0: spec to JSON: spec sub, argument custom: " +
			"jsontype: unsupported type: func()"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("error - cyclic spec", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec")
		sub := NewSpec("sub").SetArg(ArgSpecs, []*Spec{spc})
		spc.SetArg(ArgSpecs, []*Spec{sub})
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSpec, err)
		wMsg := "spec to JSON: spec my-spec, argument specs: index 0: " +
			"spec to JSON: spec sub, argument specs: index 0: " +
			"spec to JSON: cyclic spec my-spec: invalid spec"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("shared sub-spec", func(t *testing.T) {
		// --- Given ---
		sub := NewSpec("sub")
		spc := NewSpec("my-spec").SetArg(ArgSpecs, []*Spec{sub, sub})
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		want := `{
			"name": "my-spec",
			"args": {"specs": [{"name": "sub"}, {"name": "sub"}]}
		}`
		assert.JSON(t, want, string(have))
	})

	t.Run("error - types contains non-Specable element", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec").SetArg(ArgTypes, []TstType{{name: "x"}})
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.ErrorIs(t, ErrNotSpecable, err)
		wMsg := "spec to JSON: spec my-spec, argument types: " +
			"index 0: type not specable"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("error - types element Spec returns error", func(t *testing.T) {
		// --- Given ---
		tps := []TstSpec{{name: "x", err: ErrTst}}
		spc := NewSpec("my-spec").SetArg(ArgTypes, tps)
		reg := NewRegistry[TstSpec]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.ErrorIs(t, ErrTst, err)
		wMsg := "spec to JSON: spec my-spec, argument types: " +
			"index 0: test msg"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("error - values element cannot be serialized", func(t *testing.T) {
		// --- Given ---
		spc := NewSpec("my-spec").SetArg(ArgValues, []any{func() {}})
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.EncodeSpec(spc)

		// --- Then ---
		assert.ErrorIs(t, convert.ErrUnsType, err)
		wMsg := "spec to JSON: spec my-spec, argument values: " +
			"index 0: jsontype: unsupported type: func()"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})
}

func Test_Registry_DecodeSpec(t *testing.T) {
	t.Run("error - invalid JSON", func(t *testing.T) {
		// --- Given ---
		data := `{!!!}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSpec, err)
		wMsg := "JSON to spec: invalid spec: " +
			"invalid character '!' looking for beginning of object key string"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("error - nil spec", func(t *testing.T) {
		// --- Given ---
		data := `{"name": "my-spec", "args": {"arg": 1}}`
		reg := NewRegistry[TstType]()

		// --- When ---
		err := reg.DecodeSpec([]byte(data), nil)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSpec, err)
		assert.ErrorEqual(t, "JSON to spec: nil spec: invalid spec", err)
	})

	t.Run("without sources nor arguments", func(t *testing.T) {
		// --- Given ---
		data := `{"name":"my-spec"}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{Name: "my-spec", Args: nil}
		assert.Equal(t, want, have)
	})

	t.Run("resets reused spec", func(t *testing.T) {
		// --- Given ---
		data := `{"name": "new", "args": {"new": true}}`
		reg := NewRegistry[TstType]()
		have := &Spec{Name: "old", Args: map[string]any{"old": 1}}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{Name: "new", Args: map[string]any{"new": true}}
		assert.Equal(t, want, have)
	})

	t.Run("JSON null arguments are kept", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name": "my-spec",
			"args": {"other": null}
		}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{Name: "my-spec", Args: map[string]any{"other": nil}}
		assert.Equal(t, want, have)
	})

	t.Run("JSON null reserved arguments are ignored", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name": "my-spec",
			"args": {
				"specs": null,
				"types": null,
				"src_go": null,
				"values": null
			}
		}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, &Spec{Name: "my-spec"}, have)
	})

	t.Run("round trip", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstSpec]()
		reg.RegisterSource(must.Value(NewSource("my-fn", TstFn0)))
		reg.RegisterBuilder("name0", func(spc *Spec) (TstSpec, error) {
			return TstSpec{name: spc.Name}, nil
		})

		spc := NewSpec("my-spec").
			SetArg("int", 1).
			SetArg("uint", uint(2)).
			SetArg("float", 3.5).
			SetArg("bool", true).
			SetArg("string", "str").
			SetArg("time", time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)).
			SetArg("nil", nil).
			SetArg(ArgSpecs, []*Spec{NewSpec("sub")}).
			SetArg(ArgTypes, []TstSpec{{name: "name0"}}).
			SetArg(ArgSrc, TstFn0).
			SetArg(ArgValues, []any{1, 2.6, nil, "str"})
		data := must.Value(reg.EncodeSpec(spc))
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec(data, have)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, spc, have)
	})

	t.Run("round trip empty lists", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstSpec]()
		spc := NewSpec("my-spec").
			SetArg(ArgSpecs, []*Spec{}).
			SetArg(ArgTypes, []TstSpec{}).
			SetArg(ArgValues, []any{})
		data := must.Value(reg.EncodeSpec(spc))
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec(data, have)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, spc, have)
	})

	t.Run("error - empty name", func(t *testing.T) {
		// --- Given ---
		data := `{"args": {"arg": true}}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSpec, err)
		assert.ErrorEqual(t, "JSON to spec: empty name: invalid spec", err)
	})

	t.Run("error - null nested spec", func(t *testing.T) {
		// --- Given ---
		data := `{"name": "my-spec", "args": {"specs": [null]}}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSpec, err)
		wMsg := "JSON to spec: spec my-spec, argument specs: " +
			"index 0: JSON to spec: empty name: invalid spec"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("error - specs argument", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name":"my-spec",
			"args": {"specs": 42}
		}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "JSON to spec: spec my-spec, argument specs: " +
			"invalid spec argument: " +
			"json: cannot unmarshal number " +
			"into Go value of type []json.RawMessage"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("specs", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name":"my-spec",
			"args": {
				"specs": [
					{"name": "sub-spec0"},
					{"name": "sub-spec1"}
				]
			}
		}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{
			Name: "my-spec",
			Args: map[string]any{
				ArgSpecs: []*Spec{
					{Name: "sub-spec0"},
					{Name: "sub-spec1"},
				},
			},
		}
		assert.Equal(t, want, have)
	})

	t.Run("error - types argument", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name":"my-spec",
			"args": {
				"types": 42
			}
		}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "JSON to spec: spec my-spec, argument types: " +
			"invalid spec argument: " +
			"json: cannot unmarshal number " +
			"into Go value of type []json.RawMessage"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("types", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name":"my-spec",
			"args": {
				"types": [
					{"name": "type0"},
					{"name": "type1"}
				]
			}
		}`
		reg := NewRegistry[TstType]()
		reg.RegisterBuilder("type0", func(s *Spec) (TstType, error) {
			return TstType{"type0"}, nil
		})
		reg.RegisterBuilder("type1", func(s *Spec) (TstType, error) {
			return TstType{"type1"}, nil
		})
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{
			Name: "my-spec",
			Args: map[string]any{
				ArgTypes: []TstType{
					{"type0"},
					{"type1"},
				},
			},
		}
		assert.Equal(t, want, have)
	})

	t.Run("error - source argument", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name":"my-spec",
			"args": {
				"src_go": 42
			}
		}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "JSON to spec: spec my-spec, argument src_go: " +
			"invalid spec argument: " +
			"json: cannot unmarshal number into Go value of type spec.Source"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("source", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name":"my-spec",
			"args": {
				"src_go": {"name": "src0", "lang": "go"}
			}
		}`
		src := must.Value(NewSource("src0", TstFn0))
		reg := NewRegistry[TstType]()
		reg.RegisterSource(src)
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{
			Name: "my-spec",
			Args: map[string]any{
				ArgSrc: TstFn0,
			},
		}
		assert.Equal(t, want, have)
	})

	t.Run("error - values argument", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name":"my-spec",
			"args": {
				"values": 42
			}
		}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "JSON to spec: spec my-spec, argument values: " +
			"invalid spec argument: " +
			"json: cannot unmarshal number " +
			"into Go value of type []json.RawMessage"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("values", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name":"my-spec",
			"args": {
				"values": [
					{"type": "int", "value": 42},
					{"type": "uint", "value": 44}
				]
			}
		}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{
			Name: "my-spec",
			Args: map[string]any{
				ArgValues: []any{42, uint(44)},
			},
		}
		assert.Equal(t, want, have)
	})

	t.Run("zero value registry", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name": "my-spec",
			"args": {"int": {"type": "int", "value": 1}}
		}`
		reg := &Registry[TstType]{}
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{Name: "my-spec", Args: map[string]any{"int": 1}}
		assert.Equal(t, want, have)
	})

	t.Run("error - with not reserved argument name", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name":"my-spec",
			"args": {
				"arg": {"type": "int", "value": "wrong"}
			}
		}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "JSON to spec: spec my-spec, argument arg: " +
			"invalid spec argument: jsontype: invalid type: " +
			"expected float64 got string"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("not reserved argument names", func(t *testing.T) {
		// --- Given ---
		data := `{
			"name":"my-spec",
			"args": {
				"int": {"type": "int", "value": 1},
				"uint": {"type": "uint", "value": 2},
				"float": {"type": "float64", "value": 3},
				"bool": {"type": "bool", "value": true},
				"time": {"type": "time.Time", "value": "2000-01-02T03:04:05Z"}
			}
		}`
		reg := NewRegistry[TstType]()
		have := &Spec{}

		// --- When ---
		err := reg.DecodeSpec([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{
			Name: "my-spec",
			Args: map[string]any{
				"int":   1,
				"uint":  uint(2),
				"float": 3.0,
				"bool":  true,
				"time":  time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			},
		}
		assert.Equal(t, want, have)
	})
}

func Test_Registry_DecodeAndBuild(t *testing.T) {
	t.Run("error - nil slice", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.DecodeAndBuild(nil)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSpec, err)
		wMsg := "JSON to spec: invalid spec: " +
			"unexpected end of JSON input"
		assert.ErrorEqual(t, wMsg, err)
		assert.Zero(t, have)
	})

	t.Run("error - decode spec", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.DecodeAndBuild([]byte(`{!!!}`))

		// --- Then ---
		assert.ErrorIs(t, ErrInvSpec, err)
		wMsg := "JSON to spec: invalid spec: " +
			"invalid character '!' looking for beginning of object key string"
		assert.ErrorEqual(t, wMsg, err)
		assert.Zero(t, have)
	})

	t.Run("error - no builder registered", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.DecodeAndBuild([]byte(`{"name": "my-spec"}`))

		// --- Then ---
		assert.ErrorIs(t, ErrUnkBuilder, err)
		assert.ErrorEqual(t, "unknown builder my-spec", err)
		assert.Zero(t, have)
	})

	t.Run("builds value", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()
		reg.RegisterBuilder("my-spec", func(*Spec) (TstType, error) {
			return TstType{"built"}, nil
		})

		// --- When ---
		have, err := reg.DecodeAndBuild([]byte(`{"name": "my-spec"}`))

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, TstType{"built"}, have)
	})
}

func Test_Registry_concurrent_use(t *testing.T) {
	// --- Given ---
	reg := NewRegistry[TstType]()
	bld := func(*Spec) (TstType, error) { return TstType{"built"}, nil }
	data := []byte(`{"name": "my-spec", "args": {"src_go": {"name": "fn"}}}`)

	// --- When ---
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			reg.RegisterBuilder("my-spec", bld)
			reg.RegisterSource(must.Value(NewSource("fn", TstFn0)))
			_, _ = reg.EncodeSpec(NewSpec("my-spec").SetArg(ArgSrc, TstFn0))
			_, _ = reg.DecodeAndBuild(data)
		})
	}
	wg.Wait()

	// --- Then ---
	have, err := reg.DecodeAndBuild([]byte(`{"name": "my-spec"}`))
	assert.NoError(t, err)
	assert.Equal(t, TstType{"built"}, have)
}

func Test_Registry_encodeSpecs(t *testing.T) {
	t.Run("error - invalid argument type", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.encodeSpecs(true, nil)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArgType, err)
		assert.ErrorEqual(t, "invalid spec argument type", err)
		assert.Nil(t, have)
	})

	t.Run("error - marshaling error", func(t *testing.T) {
		// --- Given ---
		sps := []*Spec{
			{Name: "my-spec0", Args: map[string]any{"arg": 42}},
			{Name: "my-spec1", Args: map[string]any{"arg": func() {}}},
		}
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.encodeSpecs(sps, nil)

		// --- Then ---
		assert.ErrorIs(t, convert.ErrUnsType, err)
		wMsg := "index 1: spec to JSON: spec my-spec1, argument arg: " +
			"jsontype: unsupported type: func()"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("specs", func(t *testing.T) {
		// --- Given ---
		sps := []*Spec{{Name: "s0"}, {Name: "s1"}}
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.encodeSpecs(sps, nil)

		// --- Then ---
		assert.NoError(t, err)
		assert.SameType(t, []json.RawMessage{}, have)
		want := `[
			{"name":"s0"},
			{"name":"s1"}
		]`
		hJSON := must.Value(json.Marshal(have))
		assert.JSON(t, want, string(hJSON))
	})
}

func Test_Registry_decodeSpecs(t *testing.T) {
	t.Run("error - invalid JSON type", func(t *testing.T) {
		// --- Given ---
		data := `42`
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.decodeSpecs([]byte(data))

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "invalid spec argument: " +
			"json: cannot unmarshal number " +
			"into Go value of type []json.RawMessage"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("error - invalid specs type", func(t *testing.T) {
		// --- Given ---
		data := `"wrong0"`
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.decodeSpecs([]byte(data))

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "invalid spec argument: json: cannot unmarshal string " +
			"into Go value of type []json.RawMessage"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("error - invalid spec element", func(t *testing.T) {
		// --- Given ---
		data := `[ "wrong0", "wrong1" ]`
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.decodeSpecs([]byte(data))

		// --- Then ---
		assert.ErrorIs(t, ErrInvSpec, err)
		wMsg := "^index 0: JSON to spec: invalid spec: " +
			"json: cannot unmarshal string into Go value"
		assert.ErrorRegexp(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("specs", func(t *testing.T) {
		// --- Given ---
		data := `[
			{"name": "sub-spec0"},
			{"name": "sub-spec1"}
		]`
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.decodeSpecs([]byte(data))

		// --- Then ---
		assert.NoError(t, err)
		want := []*Spec{
			{Name: "sub-spec0"},
			{Name: "sub-spec1"},
		}
		assert.Equal(t, want, have)
	})
}

func Test_Registry_encodeTypes(t *testing.T) {
	t.Run("error - invalid argument type", func(t *testing.T) {
		// --- Given ---
		data := `42`
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.encodeTypes(data, nil)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArgType, err)
		assert.ErrorEqual(t, "invalid spec argument type", err)
		assert.Nil(t, have)
	})

	t.Run("error - not Specable type", func(t *testing.T) {
		// --- Given ---
		data := []TstType{{"name"}}
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.encodeTypes(data, nil)

		// --- Then ---
		assert.ErrorIs(t, ErrNotSpecable, err)
		assert.ErrorEqual(t, "index 0: type not specable", err)
		assert.Nil(t, have)
	})

	t.Run("error - calling spec method", func(t *testing.T) {
		// --- Given ---
		data := []TstSpec{{name: "name", err: ErrTst}}
		reg := NewRegistry[TstSpec]()

		// --- When ---
		have, err := reg.encodeTypes(data, nil)

		// --- Then ---
		assert.ErrorIs(t, ErrTst, err)
		assert.ErrorEqual(t, "index 0: test msg", err)
		assert.Nil(t, have)
	})

	t.Run("error - encoding spec", func(t *testing.T) {
		// --- Given ---
		data := []TstSpec{{
			name: "name",
			args: map[string]any{"arg": func() {}},
		}}
		reg := NewRegistry[TstSpec]()

		// --- When ---
		have, err := reg.encodeTypes(data, nil)

		// --- Then ---
		assert.ErrorIs(t, convert.ErrUnsType, err)
		wMsg := "index 0: spec to JSON: spec name, argument arg: " +
			"jsontype: unsupported type: func()"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("types argument", func(t *testing.T) {
		// --- Given ---
		data := []TstSpec{
			{name: "name0"},
			{name: "name1"},
		}
		reg := NewRegistry[TstSpec]()

		// --- When ---
		have, err := reg.encodeTypes(data, nil)

		// --- Then ---
		assert.NoError(t, err)
		assert.SameType(t, []json.RawMessage{}, have)
		want := `[
			{"name": "name0"},
			{"name": "name1"}
		]`
		hJSON := must.Value(json.Marshal(have))
		assert.JSON(t, want, string(hJSON))
	})
}

func Test_Registry_decodeTypes(t *testing.T) {
	t.Run("error - invalid JSON type", func(t *testing.T) {
		// --- Given ---
		data := `42`
		reg := NewRegistry[TstType]()
		spc := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeTypes([]byte(data), spc)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "JSON to spec: spec my-spec, argument types: " +
			"invalid spec argument: " +
			"json: cannot unmarshal number " +
			"into Go value of type []json.RawMessage"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("error - unregistered type", func(t *testing.T) {
		// --- Given ---
		data := `[
			{"name": "type0"},
			{"name": "type1"}
		]`
		reg := NewRegistry[TstType]()
		spc := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeTypes([]byte(data), spc)

		// --- Then ---
		assert.ErrorIs(t, ErrUnkBuilder, err)
		wMsg := "JSON to spec: spec my-spec, argument types[0]: " +
			"unknown builder type0"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("error - building type", func(t *testing.T) {
		// --- Given ---
		data := `[
			{"name": "type0"}
		]`
		reg := NewRegistry[TstType]()
		reg.RegisterBuilder("type0", func(s *Spec) (TstType, error) {
			return TstType{}, ErrTst
		})
		have := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeTypes([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrTst, err)
		wMsg := "JSON to spec: spec my-spec, argument types[0]: test msg"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("types", func(t *testing.T) {
		// --- Given ---
		data := `[
			{"name": "type0"},
			{"name": "type1"}
		]`
		reg := NewRegistry[TstType]()
		reg.RegisterBuilder("type0", func(s *Spec) (TstType, error) {
			return TstType{"type0"}, nil
		})
		reg.RegisterBuilder("type1", func(s *Spec) (TstType, error) {
			return TstType{"type1"}, nil
		})
		have := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeTypes([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{
			Name: "my-spec",
			Args: map[string]any{
				ArgTypes: []TstType{{"type0"}, {"type1"}},
			},
		}
		assert.Equal(t, want, have)
	})
}

func Test_Registry_encodeSource(t *testing.T) {
	t.Run("error - invalid type", func(t *testing.T) {
		// --- Given ---
		data := 42
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.encodeSource(data)

		// --- Then ---
		assert.ErrorIs(t, ErrUnkSource, err)
		wMsg := "unknown source"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("error - ambiguous source", func(t *testing.T) {
		// --- Given ---
		var fns []func() int // Closures of one literal share a code pointer.
		for i := range 2 {
			fns = append(fns, func() int { return i })
		}
		reg := NewRegistry[TstType]()
		reg.RegisterSource(must.Value(NewSource("src0", fns[0])))
		reg.RegisterSource(must.Value(NewSource("src1", fns[1])))

		// --- When ---
		have, err := reg.encodeSource(fns[1])

		// --- Then ---
		assert.ErrorIs(t, ErrInvSource, err)
		wMsg := "ambiguous source: value matches src0 and src1: invalid source"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("error - not go source", func(t *testing.T) {
		// --- Given ---
		src := must.Value(NewSource("my-src", TstFn0)).SetLang("js")
		reg := NewRegistry[TstType]()
		reg.RegisterSource(src)

		// --- When ---
		have, err := reg.encodeSource(TstFn0)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSource, err)
		assert.ErrorEqual(t, "source my-src: lang js: invalid source", err)
		assert.Nil(t, have)
	})

	t.Run("source", func(t *testing.T) {
		// --- Given ---
		fn := func() {}
		src := must.Value(NewSource("my-src", fn))
		reg := NewRegistry[TstType]()
		reg.RegisterSource(src)

		// --- When ---
		have, err := reg.encodeSource(fn)

		// --- Then ---
		assert.NoError(t, err)
		want := must.Value(NewSource("my-src", fn))
		assert.Equal(t, want, have)
	})
}

func Test_Registry_decodeSource(t *testing.T) {
	t.Run("error - invalid JSON", func(t *testing.T) {
		// --- Given ---
		data := `{!!!}`
		reg := NewRegistry[TstType]()
		have := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeSource([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "JSON to spec: spec my-spec, argument src_go: " +
			"invalid spec argument: " +
			"invalid character '!' looking for beginning of object key string"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("error - name must not be empty", func(t *testing.T) {
		// --- Given ---
		data := `{}`
		reg := NewRegistry[TstType]()
		have := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeSource([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSource, err)
		wMsg := "JSON to spec: spec my-spec, argument src_go: invalid source"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("error - field lang not equal to go", func(t *testing.T) {
		// --- Given ---
		data := `{"name": "my-src", "lang": "python"}`
		reg := NewRegistry[TstType]()
		have := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeSource([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvSource, err)
		wMsg := "JSON to spec: spec my-spec, argument src_go: invalid source"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("error - not registered source", func(t *testing.T) {
		// --- Given ---
		data := `{"name": "my-src", "lang": "go"}`
		reg := NewRegistry[TstType]()
		have := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeSource([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrUnkSource, err)
		wMsg := "JSON to spec: spec my-spec, argument src_go: unknown source"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("source", func(t *testing.T) {
		// --- Given ---
		data := `{"name": "my-src", "lang": "go"}`
		reg := NewRegistry[TstType]()
		reg.RegisterSource(must.Value(NewSource("my-src", TstFn0)))
		have := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeSource([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{
			Name: "my-spec",
			Args: map[string]any{
				"src_go": TstFn0,
			},
		}
		assert.Equal(t, want, have)
	})
}

func Test_Registry_encodeValues(t *testing.T) {
	t.Run("error - invalid spec argument type", func(t *testing.T) {
		// --- Given ---
		data := 42
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.encodeValues(data)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArgType, err)
		wMsg := "invalid spec argument type"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("error - invalid argument type", func(t *testing.T) {
		// --- Given ---
		data := []any{42, func() {}}
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.encodeValues(data)

		// --- Then ---
		assert.ErrorIs(t, convert.ErrUnsType, err)
		wMsg := "index 1: jsontype: unsupported type: func()"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("values", func(t *testing.T) {
		// --- Given ---
		data := []any{42, 4.4}
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.encodeValues(data)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, []any{jsontype.New(42), jsontype.New(4.4)}, have)
	})
}

func Test_Registry_decodeValues(t *testing.T) {
	t.Run("error - invalid JSON", func(t *testing.T) {
		// --- Given ---
		data := `[!!!]`
		reg := NewRegistry[TstType]()
		have := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeValues([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "JSON to spec: spec my-spec, argument values: " +
			"invalid spec argument: " +
			"invalid character '!' looking for beginning of value"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("error - invalid jsontype object", func(t *testing.T) {
		// --- Given ---
		data := `[{}]`
		reg := NewRegistry[TstType]()
		have := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeValues([]byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "JSON to spec: spec my-spec, argument values: index 0: " +
			"invalid spec argument: jsontype: unsupported type: "
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("values", func(t *testing.T) {
		// --- Given ---
		data := `[
			{"type": "int", "value": 1},
			{"type": "float64", "value": 2.6},
			{"type": "nil", "value": null}
		]`
		reg := NewRegistry[TstType]()
		have := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeValues([]byte(data), have)

		// --- Then ---
		assert.NoError(t, err)
		want := &Spec{
			Name: "my-spec",
			Args: map[string]any{
				"values": []any{1, 2.6, nil},
			},
		}
		assert.Equal(t, want, have)
	})
}

func Test_Registry_decodeValue(t *testing.T) {
	t.Run("error - invalid JSON", func(t *testing.T) {
		// --- Given ---
		data := `{!!!}`
		reg := NewRegistry[TstType]()
		have := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeValue("arg-name", []byte(data), have)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "JSON to spec: spec my-spec, argument arg-name: " +
			"invalid spec argument: jsontype: " +
			"invalid character '!' looking for beginning of object key string"
		assert.ErrorEqual(t, wMsg, err)
	})
}
