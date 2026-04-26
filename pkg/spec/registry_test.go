// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package spec

import (
	"encoding/json"
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
	assert.NotNil(t, reg.jtr.Converter(jsontype.Uint))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Uint16))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Uint32))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Uint64))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Uint8))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Float32))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Float64))
	assert.NotNil(t, reg.jtr.Converter(jsontype.Byte))
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
		src0 := must.Value(NewSource("fn", fn0))
		src1 := must.Value(NewSource("fn", fn1))

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
		src0 := must.Value(NewSource("fn", fn0))
		src1 := must.Value(NewSource("fn", fn1))

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

func Test_Registry_EncodeSpec(t *testing.T) {
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
						{"type": "float64", "value": 2.6},
						{"type": "nil", "value": null}
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
		wMsg := "spec my-spec to JSON: unsupported type: func()"
		assert.ErrorEqual(t, wMsg, err)
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
					"float": {"type": "float64", "value": 3},
					"bool": {"type": "bool", "value": true},
					"time": {"type": "time.Time", "value": "2000-01-02T03:04:05Z"}
				}
			}`
		assert.JSON(t, want, string(have))
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
		assert.ErrorEqual(t, "JSON to spec: invalid spec", err)
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

	t.Run("JSON null arguments are ignored", func(t *testing.T) {
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
		want := &Spec{Name: "my-spec", Args: nil}
		assert.Equal(t, want, have)
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
			"invalid spec argument"
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
			"invalid spec argument"
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
			"invalid spec argument"
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
			"invalid spec argument"
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
				ArgValues: []any{42, uint(44)},
			},
		}
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
		wMsg := "JSON to spec: spec my-spec, argument value: " +
			"invalid spec argument"
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

func Test_Registry_encodeSpecs(t *testing.T) {
	t.Run("error - invalid argument type", func(t *testing.T) {
		// --- Given ---
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.encodeSpecs(true)

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
		have, err := reg.encodeSpecs(sps)

		// --- Then ---
		assert.ErrorIs(t, convert.ErrUnsType, err)
		wMsg := "index 1: spec my-spec1 to JSON: unsupported type: func()"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})

	t.Run("specs", func(t *testing.T) {
		// --- Given ---
		sps := []*Spec{{Name: "s0"}, {Name: "s1"}}
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.encodeSpecs(sps)

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
	t.Run("error - invalid JOSN type", func(t *testing.T) {
		// --- Given ---
		data := `42`
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.decodeSpecs([]byte(data))

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		assert.ErrorEqual(t, "invalid spec argument", err)
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
		assert.ErrorEqual(t, "invalid spec argument", err)
		assert.Nil(t, have)
	})

	t.Run("error - invalid specs type", func(t *testing.T) {
		// --- Given ---
		data := `[ "wrong0", "wrong1" ]`
		reg := NewRegistry[TstType]()

		// --- When ---
		have, err := reg.decodeSpecs([]byte(data))

		// --- Then ---
		assert.ErrorIs(t, ErrInvSpec, err)
		assert.ErrorEqual(t, "index 0: JSON to spec: invalid spec", err)
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
		have, err := reg.encodeTypes(data)

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
		have, err := reg.encodeTypes(data)

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
		have, err := reg.encodeTypes(data)

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
		have, err := reg.encodeTypes(data)

		// --- Then ---
		assert.ErrorIs(t, convert.ErrUnsType, err)
		wMsg := "index 0: spec name to JSON: unsupported type: func()"
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
		have, err := reg.encodeTypes(data)

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
	t.Run("error - invalid JOSN type", func(t *testing.T) {
		// --- Given ---
		data := `42`
		reg := NewRegistry[TstType]()
		spc := &Spec{Name: "my-spec"}

		// --- When ---
		err := reg.decodeTypes([]byte(data), spc)

		// --- Then ---
		assert.ErrorIs(t, ErrInvArg, err)
		wMsg := "JSON to spec: spec my-spec, argument types: " +
			"invalid spec argument"
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
			"invalid spec argument"
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
		wMsg := "index 1: unsupported type: func()"
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
			"invalid spec argument"
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
			"invalid spec argument"
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
		wMsg := "JSON to spec: spec my-spec, argument value: " +
			"invalid spec argument"
		assert.ErrorEqual(t, wMsg, err)
	})
}
