// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package spec

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sync"

	"github.com/ctx42/jsontype/pkg/jsontype"
)

// Registry manages a collection of [Source] and [Builder] instances and
// provides methods to encode and decode [Spec] instances to generic type T. It
// is safe for concurrent use.
type Registry[T any] struct {
	// Preserve Go types during encoding to / decoding from JSON round trip.
	jtr *jsontype.Registry

	// Sources needed during decoding.
	sources []Source

	// Maps specification names [Spec.Name] to their [Builder].
	builders map[string]Builder[T]

	mx sync.RWMutex // Guards struct fields.
}

// NewRegistry returns a new instance of [Registry].
func NewRegistry[T any]() *Registry[T] {
	return &Registry[T]{
		jtr:      jsontype.DefaultRegistry(),
		sources:  make([]Source, 0, 20),
		builders: make(map[string]Builder[T], 20),
	}
}

// RegisterSource registers a source so it can be later used during decoding.
func (reg *Registry[T]) RegisterSource(src Source) Source {
	reg.mx.Lock()
	defer reg.mx.Unlock()

	var old Source
	for i, have := range reg.sources {
		if have.Name == src.Name {
			old = have
			reg.sources[i] = src
			return old
		}
	}
	reg.sources = append(reg.sources, src)
	return old
}

// SourceByName retrieves an instance of [Source] by its name.
func (reg *Registry[T]) SourceByName(name string) Source {
	reg.mx.RLock()
	defer reg.mx.RUnlock()

	for _, src := range reg.sources {
		if src.Name == name {
			return src
		}
	}
	return Source{}
}

// SourceByValue retrieves an instance of [Source] representing a given value.
func (reg *Registry[T]) SourceByValue(value any) Source {
	reg.mx.RLock()
	defer reg.mx.RUnlock()

	ptr := GetSrcPointer(reflect.ValueOf(value))
	if ptr == 0 {
		return Source{}
	}
	for _, src := range reg.sources {
		if src.Ptr() == ptr {
			return src
		}
	}
	return Source{}
}

// RegisterBuilder registers or replaces a [Builder] for the given spec name.
//
// If a builder with the same name already exists, it is replaced and the
// previous builder is returned.
//
// If bld is nil and a builder with the given name exists, it is removed and
// the removed builder is returned.
//
// Returns the previous builder (or nil if none existed).
func (reg *Registry[T]) RegisterBuilder(
	name string,
	bld Builder[T],
) Builder[T] {

	reg.mx.Lock()
	defer reg.mx.Unlock()

	old := reg.builders[name]
	if bld == nil {
		delete(reg.builders, name)
	} else {
		reg.builders[name] = bld
	}
	return old
}

// RegisterBuilders registers multiple [Builder] instances in a single call.
//
// It invokes [Registry.RegisterBuilder] for each entry in bls and returns a
// map of name → previous builder (or nil if none existed) for each
// registration.
func (reg *Registry[T]) RegisterBuilders(
	bls map[string]Builder[T],
) map[string]Builder[T] {

	old := make(map[string]Builder[T], len(bls))
	for name, bld := range bls {
		old[name] = reg.RegisterBuilder(name, bld)
	}
	return old
}

// BuilderFor returns the [Builder] registered for the given name, or nil if
// none exists.
func (reg *Registry[T]) BuilderFor(name string) Builder[T] {
	reg.mx.RLock()
	defer reg.mx.RUnlock()

	return reg.builders[name]
}

// Build creates an instance of T from the given [Spec] using a registered
// [Builder]. Returns [ErrInvSpec] if spc is nil, or [ErrUnkBuilder] if no
// [Builder] is registered for [Spec.Name].
func (reg *Registry[T]) Build(spc *Spec) (T, error) {
	var zero T
	if spc == nil {
		return zero, ErrInvSpec
	}
	bld := reg.BuilderFor(spc.Name)
	if bld == nil {
		return zero, NewErrorf("%w %s", ErrUnkBuilder, spc.Name)
	}
	return bld(spc)
}

// EncodeSpec encodes the given [Spec] to JSON.
//
// NOTE: The input spc is never mutated. EncodeSpec works on an internal
// copy so callers can safely reuse the same *Spec across multiple
// Encode / Build / roundtrip operations.
func (reg *Registry[T]) EncodeSpec(spc *Spec) ([]byte, error) {
	if spc == nil {
		return nil, ErrInvSpec
	}
	// Work on a copy so the caller's Spec is never mutated.
	work := &Spec{
		Name: spc.Name,
		Args: make(map[string]any, len(spc.Args)),
	}
	for k, v := range spc.Args {
		work.Args[k] = v
	}

	for name, value := range work.Args {
		switch name {
		case ArgSpecs:
			specs, err := reg.encodeSpecs(value)
			if err != nil {
				format := "spec to JSON: spec %s, argument %s: %w"
				return nil, NewErrorf(format, spc.Name, name, err)
			}
			work.Args[name] = specs

		case ArgTypes:
			tps, err := reg.encodeTypes(value)
			if err != nil {
				format := "spec to JSON: spec %s, argument %s: %w"
				return nil, NewErrorf(format, spc.Name, name, err)
			}
			work.Args[name] = tps

		case ArgSrc:
			src, err := reg.encodeSource(value)
			if err != nil {
				format := "spec to JSON: spec %s, argument %s: %w"
				return nil, NewErrorf(format, spc.Name, name, err)
			}
			work.Args[name] = src

		case ArgValues:
			values, err := reg.encodeValues(value)
			if err != nil {
				format := "spec to JSON: spec %s, argument %s: %w"
				return nil, NewErrorf(format, spc.Name, name, err)
			}
			work.Args[name] = values

		default:
			val, err := jsontype.NewValue(value)
			if err != nil {
				format := "spec %s to JSON: %w"
				return nil, NewErrorf(format, spc.Name, err)
			}
			work.Args[name] = val
		}
	}

	data, err := json.Marshal(work)
	if err != nil {
		return nil, NewErrorf("spec to JSON: spec %s: %w", spc.Name, err)
	}
	return data, nil
}

// jsNull represents the JSON null value.
var jsNull = json.RawMessage(`null`)

// DecodeSpec decodes JSON representation of [Spec].
func (reg *Registry[T]) DecodeSpec(data []byte, spc *Spec) error {
	tmp := struct {
		*Spec
		Args map[string]json.RawMessage `json:"args"`
	}{
		Spec: spc,
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return NewErrorf("JSON to spec: %w", ErrInvSpec)
	}

	for name, value := range tmp.Args {
		if bytes.Equal(value, jsNull) {
			continue
		}

		switch name {
		case ArgSpecs:
			sps, err := reg.decodeSpecs(value)
			if err != nil {
				format := "JSON to spec: spec %s, argument %s: %w"
				return NewErrorf(format, spc.Name, name, err)
			}
			spc.SetArg(ArgSpecs, sps)

		case ArgTypes:
			if err := reg.decodeTypes(value, spc); err != nil {
				return err
			}

		case ArgSrc:
			if err := reg.decodeSource(value, spc); err != nil {
				return err
			}

		case ArgValues:
			if err := reg.decodeValues(value, spc); err != nil {
				return err
			}

		default:
			if err := reg.decodeValue(name, value, spc); err != nil {
				return err
			}
		}
	}
	return nil
}

// DecodeAndBuild decodes the JSON representation of a [Spec] and builds an
// instance of T from it. It is a convenience wrapper around
// [Registry.DecodeSpec] followed by [Registry.Build].
func (reg *Registry[T]) DecodeAndBuild(data []byte) (T, error) {
	var zero T

	spec := &Spec{}
	if err := reg.DecodeSpec(data, spec); err != nil {
		return zero, err
	}

	value, err := reg.Build(spec)
	if err != nil {
		return zero, err
	}

	return value, nil
}

// encodeSpecs expects the provided value to be a slice of [Spec] instances and
// encodes as into a JSON array as a slice of [json.RawMessage].
func (reg *Registry[T]) encodeSpecs(value any) (any, error) {
	sps, ok := value.([]*Spec)
	if !ok {
		return nil, ErrInvArgType
	}

	var subs []json.RawMessage
	for idx, spc := range sps {
		data, err := reg.EncodeSpec(spc)
		if err != nil {
			return nil, NewErrorf("index %d: %w", idx, err)
		}
		subs = append(subs, data)
	}
	return subs, nil
}

// decodeSpecs expects the input data to be a JSON array of [Spec]
// representations and decodes it into a slice of [Spec] instances.
func (reg *Registry[T]) decodeSpecs(data []byte) ([]*Spec, error) {
	var subs []json.RawMessage
	if err := json.Unmarshal(data, &subs); err != nil {
		return nil, ErrInvArg
	}

	var sps []*Spec
	for idx, sub := range subs {
		s := &Spec{}
		if err := reg.DecodeSpec(sub, s); err != nil {
			return nil, NewErrorf("index %d: %w", idx, err)
		}
		sps = append(sps, s)
	}
	return sps, nil
}

// encodeTypes expects the provided value to be a slice of generic T instances
// and encodes them as a JSON array as a slice of [json.RawMessage].
func (reg *Registry[T]) encodeTypes(data any) (any, error) {
	tps, ok := data.([]T)
	if !ok {
		return nil, ErrInvArgType
	}

	var subs []json.RawMessage
	for idx, typ := range tps {
		spt, ok := any(typ).(Specable)
		if !ok {
			return nil, NewErrorf("index %d: %w", idx, ErrNotSpecable)
		}
		spc, err := spt.Spec()
		if err != nil {
			return nil, NewErrorf("index %d: %w", idx, err)
		}
		sub, err := reg.EncodeSpec(spc)
		if err != nil {
			return nil, NewErrorf("index %d: %w", idx, err)
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

// decodeTypes expects the input data to be a JSON array of [Spec]
// representations and decodes them into a slice of the generic type T using
// registered [Builder] functions. On success, it sets it as [ArgTypes] key
// in the [Spec.Args] map.
func (reg *Registry[T]) decodeTypes(data []byte, spc *Spec) error {
	sps, err := reg.decodeSpecs(data)
	if err != nil {
		format := "JSON to spec: spec %s, argument %s: %w"
		return NewErrorf(format, spc.Name, ArgTypes, err)
	}
	var tps []T
	for idx, s := range sps {
		bld := reg.BuilderFor(s.Name)
		if bld == nil {
			return NewErrorf(
				"JSON to spec: spec %s, argument %s[%d]: %w %s",
				spc.Name,
				ArgTypes,
				idx,
				ErrUnkBuilder,
				s.Name,
			)
		}
		typ, err := bld(s)
		if err != nil {
			return NewErrorf(
				"JSON to spec: spec %s, argument %s[%d]: %w",
				spc.Name,
				ArgTypes,
				idx,
				err,
			)
		}
		tps = append(tps, typ)
	}
	spc.SetArg(ArgTypes, tps)
	return nil
}

// encodeSource encodes the given source value. The source must be registered
// using the [Registry.SourceByValue] method beforehand.
func (reg *Registry[T]) encodeSource(value any) (any, error) {
	src := reg.SourceByValue(value)
	if src.IsZero() {
		return nil, ErrUnkSource
	}
	return src, nil
}

// decodeSource decodes a JSON representation of a [Source]. On success, it
// sets it as [ArgSrc] key in the [Spec.Args] map.
func (reg *Registry[T]) decodeSource(data []byte, spc *Spec) error {
	src := Source{}
	if err := json.Unmarshal(data, &src); err != nil {
		format := "JSON to spec: spec %s, argument %s: %w"
		return NewErrorf(format, spc.Name, ArgSrc, ErrInvArg)
	}
	if src.Name == "" {
		format := "JSON to spec: spec %s, argument %s: %w"
		return NewErrorf(format, spc.Name, ArgSrc, ErrInvSource)
	}
	if src.Lang != "go" {
		format := "JSON to spec: spec %s, argument %s: %w"
		return NewErrorf(format, spc.Name, ArgSrc, ErrInvSource)
	}
	src = reg.SourceByName(src.Name)
	if src.IsZero() {
		format := "JSON to spec: spec %s, argument %s: %w"
		return NewErrorf(format, spc.Name, ArgSrc, ErrUnkSource)
	}
	spc.SetArg(ArgSrc, src.Val())
	return nil
}

// encodeValues expects the given value to be a `[]any` and encodes them as a
// slice of [jsontype.Value] instances.
func (reg *Registry[T]) encodeValues(value any) (any, error) {
	vs, ok := value.([]any)
	if !ok {
		return nil, ErrInvArgType
	}
	var values []any
	for idx, v := range vs {
		jv, err := jsontype.NewValue(v, jsontype.WithRegistry(reg.jtr))
		if err != nil {
			return nil, NewErrorf("index %d: %w", idx, err)
		}
		values = append(values, jv)
	}
	return values, nil
}

// decodeValues decodes a JSON array of [jsontype.Value] representations
// into a `[]any` slice and sets it as [ArgValues] key in the [Spec.Args] map.
func (reg *Registry[T]) decodeValues(data []byte, spc *Spec) error {
	var rv []json.RawMessage
	if err := json.Unmarshal(data, &rv); err != nil {
		format := "JSON to spec: spec %s, argument %s: %w"
		return NewErrorf(format, spc.Name, ArgValues, ErrInvArg)
	}
	var vs []any
	for idx, v := range rv {
		val := jsontype.Value{}
		err := jsontype.Unmarshal(reg.jtr, v, &val)
		if err != nil {
			return NewErrorf(
				"JSON to spec: spec %s, argument %s: index %d: %w",
				spc.Name,
				ArgValues,
				idx,
				ErrInvArg,
			)
		}
		vs = append(vs, val.GoValue())
	}
	spc.SetArg(ArgValues, vs)
	return nil
}

// decodeValue decodes a single JSON representation of [jsontype.Value] and
// sets it with the given name in the [Spec.Args] map.
func (reg *Registry[T]) decodeValue(
	name string,
	data []byte,
	spc *Spec,
) error {

	val := jsontype.Value{}
	err := jsontype.Unmarshal(reg.jtr, data, &val)
	if err != nil {
		format := "JSON to spec: spec %s, argument %s: %w"
		return NewErrorf(format, spc.Name, ArgValue, ErrInvArg)
	}
	spc.SetArg(name, val.GoValue())
	return nil
}
