// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package spec

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"sync"

	"github.com/ctx42/jsontype/pkg/jsontype"
)

// Error message formats wrapping an argument error with its spec and name.
const (
	fmtEncArg = "spec to JSON: spec %s, argument %s: %w"
	fmtDecArg = "JSON to spec: spec %s, argument %s: %w"
)

// Registry manages a collection of [Source] and [Builder] instances and
// provides methods to encode and decode [Spec] instances to generic type T. It
// is safe for concurrent use. The zero value is ready to use.
type Registry[T any] struct {
	// Preserve Go types during encoding to / decoding from JSON round trip.
	jtr *jsontype.Registry

	// Lazily sets jtr on a zero value [Registry].
	jtrOnce sync.Once

	// Sources needed during decoding.
	sources []Source

	// Maps specification names [Spec.Name] to their [Builder].
	builders map[string]Builder[T]

	mx sync.RWMutex // Guards struct fields.
}

// NewRegistry returns a new [Registry] preserving Go types in JSON with the
// default [jsontype] converters.
func NewRegistry[T any]() *Registry[T] {
	return &Registry[T]{
		jtr:      jsontype.DefaultRegistry(),
		sources:  make([]Source, 0, 20),
		builders: make(map[string]Builder[T], 20),
	}
}

// RegisterSource registers a source so it can be later used during encoding
// and decoding. It replaces a source with the same name and returns it, or
// returns the zero [Source] when none existed.
//
// A source with a zero pointer, such as one decoded from JSON, is never
// matched by [Registry.SourceByValue].
func (reg *Registry[T]) RegisterSource(src Source) Source {
	reg.mx.Lock()
	defer reg.mx.Unlock()

	for i, cur := range reg.sources {
		if cur.Name == src.Name {
			reg.sources[i] = src
			return cur
		}
	}
	reg.sources = append(reg.sources, src)
	return Source{}
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
//
// Sources are matched by the function's code pointer, which closures created
// by the same function literal, and method values of the same method, share.
// When several registered sources match, the first registered one is
// returned; encoding such a value returns [ErrInvSource].
func (reg *Registry[T]) SourceByValue(value any) Source {
	if srcs := reg.sourcesByValue(value); len(srcs) > 0 {
		return srcs[0]
	}
	return Source{}
}

// sourcesByValue returns all registered sources matching the value's
// function code pointer, in registration order.
func (reg *Registry[T]) sourcesByValue(value any) []Source {
	ptr := GetSrcPointer(reflect.ValueOf(value))
	if ptr == 0 {
		return nil
	}

	reg.mx.RLock()
	defer reg.mx.RUnlock()

	var srcs []Source
	for _, src := range reg.sources {
		if src.Ptr() == ptr {
			srcs = append(srcs, src)
		}
	}
	return srcs
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
		if reg.builders == nil {
			reg.builders = make(map[string]Builder[T])
		}
		reg.builders[name] = bld
	}
	return old
}

// RegisterBuilders registers multiple [Builder] instances in a single call.
//
// It invokes [Registry.RegisterBuilder] for each entry in bls and returns a
// map of name → previous builder (or nil if none existed) for each
// registration. The registrations are not atomic: concurrent readers may
// observe some builders registered before others.
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

// EncodeSpec encodes the given [Spec] to JSON. Returns [ErrInvSpec] if spc is
// nil or has an empty name.
//
// NOTE: The input spc is never mutated. EncodeSpec works on an internal
// copy so callers can safely reuse the same *Spec across multiple
// Encode / Build / roundtrip operations.
//
// A spec nested in itself through [ArgSpecs] returns [ErrInvSpec].
func (reg *Registry[T]) EncodeSpec(spc *Spec) ([]byte, error) {
	return reg.encodeSpec(spc, nil)
}

// encodeSpec encodes the given [Spec] to JSON. The path holds the specs being
// encoded above spc and is used to detect cycles.
func (reg *Registry[T]) encodeSpec(
	spc *Spec,
	path map[*Spec]struct{},
) ([]byte, error) {

	if spc == nil {
		return nil, ErrInvSpec
	}
	if spc.Name == "" {
		return nil, NewErrorf("spec to JSON: empty name: %w", ErrInvSpec)
	}
	if _, ok := path[spc]; ok {
		format := "spec to JSON: cyclic spec %s: %w"
		return nil, NewErrorf(format, spc.Name, ErrInvSpec)
	}
	if path == nil {
		path = make(map[*Spec]struct{})
	}
	path[spc] = struct{}{}
	defer delete(path, spc)

	// Work on a copy so the caller's Spec is never mutated.
	work := &Spec{
		Name: spc.Name,
		Args: make(map[string]any, len(spc.Args)),
	}
	maps.Copy(work.Args, spc.Args)

	for name, value := range work.Args {
		switch name {
		case ArgSpecs:
			specs, err := reg.encodeSpecs(value, path)
			if err != nil {
				return nil, NewErrorf(fmtEncArg, spc.Name, name, err)
			}
			work.Args[name] = specs

		case ArgTypes:
			tps, err := reg.encodeTypes(value, path)
			if err != nil {
				return nil, NewErrorf(fmtEncArg, spc.Name, name, err)
			}
			work.Args[name] = tps

		case ArgSrc:
			src, err := reg.encodeSource(value)
			if err != nil {
				return nil, NewErrorf(fmtEncArg, spc.Name, name, err)
			}
			work.Args[name] = src

		case ArgValues:
			values, err := reg.encodeValues(value)
			if err != nil {
				return nil, NewErrorf(fmtEncArg, spc.Name, name, err)
			}
			work.Args[name] = values

		default:
			jtr := jsontype.WithRegistry(reg.jsonTypes())
			val, err := jsontype.NewValue(value, jtr)
			if err != nil {
				return nil, NewErrorf(fmtEncArg, spc.Name, name, err)
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

// isReserved reports whether name is a reserved [Spec] argument name the
// [Registry] encodes and decodes especially.
func isReserved(name string) bool {
	switch name {
	case ArgSpecs, ArgTypes, ArgSrc, ArgValues:
		return true
	}
	return false
}

// withCause returns an error matching both the sentinel and its cause with
// [errors.Is], reading "sentinel: cause".
func withCause(sentinel, cause error) error {
	return fmt.Errorf("%w: %w", sentinel, cause)
}

// DecodeSpec decodes JSON representation of [Spec]. It resets spc before
// decoding, so no name or argument from a reused spc survives. Returns
// [ErrInvSpec] if spc is nil or the decoded name is empty.
//
// A JSON null argument is decoded as a nil value, except for the reserved
// [ArgSpecs], [ArgTypes], [ArgSrc], and [ArgValues] arguments, which are
// skipped.
func (reg *Registry[T]) DecodeSpec(data []byte, spc *Spec) error {
	if spc == nil {
		return NewErrorf("JSON to spec: nil spec: %w", ErrInvSpec)
	}
	*spc = Spec{}
	tmp := struct {
		*Spec
		Args map[string]json.RawMessage `json:"args"`
	}{
		Spec: spc,
	}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return NewErrorf("JSON to spec: %w", withCause(ErrInvSpec, err))
	}
	if spc.Name == "" {
		return NewErrorf("JSON to spec: empty name: %w", ErrInvSpec)
	}

	for name, value := range tmp.Args {
		if isReserved(name) && string(value) == "null" {
			continue
		}

		switch name {
		case ArgSpecs:
			sps, err := reg.decodeSpecs(value)
			if err != nil {
				return NewErrorf(fmtDecArg, spc.Name, name, err)
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

	spc := &Spec{}
	if err := reg.DecodeSpec(data, spc); err != nil {
		return zero, err
	}

	value, err := reg.Build(spc)
	if err != nil {
		return zero, err
	}

	return value, nil
}

// jsonTypes returns the registry preserving Go types in JSON, creating the
// default one on first use when the [Registry] is a zero value.
func (reg *Registry[T]) jsonTypes() *jsontype.Registry {
	reg.jtrOnce.Do(func() {
		if reg.jtr == nil {
			reg.jtr = jsontype.DefaultRegistry()
		}
	})
	return reg.jtr
}

// encodeSpecs expects the provided value to be a slice of [Spec] instances and
// encodes them into a slice of [json.RawMessage] forming a JSON array.
func (reg *Registry[T]) encodeSpecs(
	value any,
	path map[*Spec]struct{},
) (any, error) {

	sps, ok := value.([]*Spec)
	if !ok {
		return nil, ErrInvArgType
	}

	subs := make([]json.RawMessage, 0, len(sps))
	for idx, spc := range sps {
		data, err := reg.encodeSpec(spc, path)
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
		return nil, NewErrorf("%w", withCause(ErrInvArg, err))
	}

	sps := make([]*Spec, 0, len(subs))
	for idx, sub := range subs {
		spc := &Spec{}
		if err := reg.DecodeSpec(sub, spc); err != nil {
			return nil, NewErrorf("index %d: %w", idx, err)
		}
		sps = append(sps, spc)
	}
	return sps, nil
}

// encodeTypes expects the provided value to be a slice of generic T instances
// and encodes them into a slice of [json.RawMessage] forming a JSON array.
func (reg *Registry[T]) encodeTypes(
	data any,
	path map[*Spec]struct{},
) (any, error) {

	tps, ok := data.([]T)
	if !ok {
		return nil, ErrInvArgType
	}

	subs := make([]json.RawMessage, 0, len(tps))
	for idx, typ := range tps {
		spt, ok := any(typ).(Specable)
		if !ok {
			return nil, NewErrorf("index %d: %w", idx, ErrNotSpecable)
		}
		spc, err := spt.Spec()
		if err != nil {
			return nil, NewErrorf("index %d: %w", idx, err)
		}
		sub, err := reg.encodeSpec(spc, path)
		if err != nil {
			return nil, NewErrorf("index %d: %w", idx, err)
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

// decodeTypes expects the input data to be a JSON array of [Spec]
// representations and decodes them into a slice of the generic type T using
// [Registry.Build]. On success, it sets it as [ArgTypes] key
// in the [Spec.Args] map.
func (reg *Registry[T]) decodeTypes(data []byte, spc *Spec) error {
	sps, err := reg.decodeSpecs(data)
	if err != nil {
		return NewErrorf(fmtDecArg, spc.Name, ArgTypes, err)
	}
	tps := make([]T, 0, len(sps))
	for idx, sub := range sps {
		typ, err := reg.Build(sub)
		if err != nil {
			format := "JSON to spec: spec %s, argument %s[%d]: %w"
			return NewErrorf(format, spc.Name, ArgTypes, idx, err)
		}
		tps = append(tps, typ)
	}
	spc.SetArg(ArgTypes, tps)
	return nil
}

// encodeSource encodes the given source value. The source must be registered
// with [Registry.RegisterSource] beforehand. Returns [ErrInvSource] when the
// value matches several sources or its source language is not Go.
func (reg *Registry[T]) encodeSource(value any) (any, error) {
	srcs := reg.sourcesByValue(value)
	if len(srcs) == 0 {
		return nil, ErrUnkSource
	}
	if len(srcs) > 1 {
		format := "ambiguous source: value matches %s and %s: %w"
		return nil, NewErrorf(format, srcs[0].Name, srcs[1].Name, ErrInvSource)
	}
	src := srcs[0]
	if src.Lang != "go" {
		format := "source %s: lang %s: %w"
		return nil, NewErrorf(format, src.Name, src.Lang, ErrInvSource)
	}
	return src, nil
}

// decodeSource decodes a JSON representation of a [Source]. On success, it
// sets it as [ArgSrc] key in the [Spec.Args] map.
func (reg *Registry[T]) decodeSource(data []byte, spc *Spec) error {
	src := Source{}
	if err := json.Unmarshal(data, &src); err != nil {
		cause := withCause(ErrInvArg, err)
		return NewErrorf(fmtDecArg, spc.Name, ArgSrc, cause)
	}
	if src.Name == "" {
		return NewErrorf(fmtDecArg, spc.Name, ArgSrc, ErrInvSource)
	}
	if src.Lang != "go" {
		return NewErrorf(fmtDecArg, spc.Name, ArgSrc, ErrInvSource)
	}
	src = reg.SourceByName(src.Name)
	if src.IsZero() {
		return NewErrorf(fmtDecArg, spc.Name, ArgSrc, ErrUnkSource)
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
	values := make([]any, 0, len(vs))
	for idx, v := range vs {
		jv, err := jsontype.NewValue(v, jsontype.WithRegistry(reg.jsonTypes()))
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
		cause := withCause(ErrInvArg, err)
		return NewErrorf(fmtDecArg, spc.Name, ArgValues, cause)
	}
	vs := make([]any, 0, len(rv))
	for idx, v := range rv {
		val := jsontype.Value{}
		err := jsontype.Unmarshal(reg.jsonTypes(), v, &val)
		if err != nil {
			cause := withCause(ErrInvArg, err)
			format := "JSON to spec: spec %s, argument %s: index %d: %w"
			return NewErrorf(format, spc.Name, ArgValues, idx, cause)
		}
		vs = append(vs, val.GoValue())
	}
	spc.SetArg(ArgValues, vs)
	return nil
}

// decodeValue decodes a single JSON representation of [jsontype.Value] and
// sets it with the given name in the [Spec.Args] map.
func (reg *Registry[T]) decodeValue(name string, data []byte, spc *Spec) error {
	val := jsontype.Value{}
	err := jsontype.Unmarshal(reg.jsonTypes(), data, &val)
	if err != nil {
		cause := withCause(ErrInvArg, err)
		return NewErrorf(fmtDecArg, spc.Name, name, cause)
	}
	spc.SetArg(name, val.GoValue())
	return nil
}
