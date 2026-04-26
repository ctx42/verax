// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package spec provides tools to create serializable type specifications and
// encoding/decoding functionality for them.
package spec

// Package level error codes.
const (
	// ECInvSpec is the error code used when a [Spec] is invalid or incomplete.
	ECInvSpec = "ECInvSpec"

	// ECUnkSource is the error code used when a [Source] doesn't exist.
	ECUnkSource = "ECUnkSource"

	// ECUnkBuilder is the error code used when a [Builder] doesn't exist.
	ECUnkBuilder = "ECUnkBuilder"

	// ECInvSource is the error code used when a [Source] is invalid or
	// unsupported.
	ECInvSource = "ECInvSource"

	// ECNoGoSource is the error code used when a [Spec] lacks the required
	// [Source] definition for Go language.
	ECNoGoSource = "ECNoGoSource"

	// ECInvSpecArg is the error code used when a [Spec] argument is invalid.
	ECInvSpecArg = "ECECInvSpecArg"

	// ECInvSpecArgType is the error code used when a [Spec] argument has an
	// invalid type.
	ECInvSpecArgType = "ECInvSpecArgType"

	// ECNotSpecable is the error code used when a type doesn't implement
	// [Specable] interface.
	ECNotSpecable = "ECNotSpecable"
)

// Package level sentinel errors.
var (
	// ErrInvSpec is returned when a [Spec] is either invalid or incomplete.
	ErrInvSpec = NewError("invalid spec", ECInvSpec)

	// ErrNoGoSource is returned when a [Spec] lacks the required Go definition.
	ErrNoGoSource = NewError("spec has no Go source defined", ECNoGoSource)

	// ErrUnkSource is returned when a [Source] does not exist.
	ErrUnkSource = NewError("unknown source", ECUnkSource)

	// ErrUnkBuilder is returned when a [Builder] does not exist.
	ErrUnkBuilder = NewError("unknown builder", ECUnkBuilder)

	// ErrInvSource is returned when a [Source] definition is invalid.
	ErrInvSource = NewError("invalid source", ECInvSource)

	// ErrInvArg is returned when a [Spec] argument is invalid.
	ErrInvArg = NewError("invalid spec argument", ECInvSpecArg)

	// ErrInvArgType is returned when a [Spec] argument is of an invalid type.
	ErrInvArgType = NewError("invalid spec argument type", ECInvSpecArgType)

	// ErrNotSpecable is returned when a type doesn't implement [Specable].
	ErrNotSpecable = NewError("type not specable", ECNotSpecable)
)

// [Spec] argument names.
const (
	// ArgValue is a reserved [Spec] argument name holding a value.
	//
	// The value must be representable by [jsontype.Value].
	//
	// When encoding, this value should be replaced with [jsontype.Value]. When
	// decoding, this value must be replaced with the concrete value encoded by
	// [jsontype.Value].
	ArgValue = "value"

	// ArgValues is a reserved [Spec] argument name holding a list of values.
	//
	// It is represented by `[]any` slice, and each value must be representable
	// as [jsontype.Value].
	//
	// When encoding, this value should be replaced with a slice of
	// [jsontype.Value] representations. When decoding, this value must be
	// replaced with the `[]any` slice.
	ArgValues = "values"

	// ArgSrc is a reserved [Spec] argument name holding a value like functions
	// which cannot be serialized.
	//
	// When encoding, this value must be replaced with an instance of [Source]
	// representing the value. When decoding, this value must be replaced with
	// the concrete value.
	ArgSrc = "src_go"

	// ArgSpecs is a reserved [Spec] argument name holding a list of [Spec]
	// instances.
	ArgSpecs = "specs"

	// ArgTypes is a reserved [Spec] argument name holding a list of typed
	// values. The [Registry] encodes and decodes this argument specially: on
	// marshaling each element is turned into a [Spec]; on unmarshaling each
	// element spec is rebuilt via a registered [Builder].
	ArgTypes = "types"
)

// Specable types implementing this interface can be encoded as [Spec] instance.
type Specable interface {
	// Spec returns the [Spec] describing the given structure. The returned
	// spec must include all information needed to instantiate the structure at
	// a later time.
	Spec() (*Spec, error)
}

// Builder creates a type instance from the given [Spec].
type Builder[T any] func(*Spec) (T, error)

// Spec represents information about a type or process that can be serialized.
type Spec struct {
	// The unique name of the specification. Use application-wide unique names.
	Name string `json:"name"`

	// The map of additional arguments for the specification.
	Args map[string]any `json:"args,omitempty"`
}

// NewSpec creates a new [Spec] instance with the given name.
//
// Use application-wide unique names.
func NewSpec(name string) *Spec { return &Spec{Name: name} }

// SetArg sets [Source] argument; when the argument with the same name exists,
// it will be overwritten.
func (spc *Spec) SetArg(name string, val any) *Spec {
	if spc.Args == nil {
		spc.Args = make(map[string]any)
	}
	spc.Args[name] = val
	return spc
}

// ArgExist checks if the argument with the given name exists in the spec.
func (spc *Spec) ArgExist(name string) bool {
	_, exist := spc.Args[name]
	return exist
}
