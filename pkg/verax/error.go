// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"encoding/json"

	"github.com/ctx42/xrr/pkg/xrr"
)

// Unexported zero-size marker types used as domain type parameters.
type (
	edError    struct{}
	edInternal struct{}
	edFields   struct{}
)

// Compile checks.
var (
	_ error            = (*Error)(nil)
	_ xrr.Coder        = (*Error)(nil)
	_ json.Marshaler   = (*Error)(nil)
	_ json.Unmarshaler = (*Error)(nil)

	_ error            = (*FieldsError)(nil)
	_ xrr.Fielder      = (*FieldsError)(nil)
	_ json.Marshaler   = (*FieldsError)(nil)
	_ json.Unmarshaler = (*FieldsError)(nil)

	_ error            = (*InternalError)(nil)
	_ xrr.Coder        = (*InternalError)(nil)
	_ json.Marshaler   = (*InternalError)(nil)
	_ json.Unmarshaler = (*InternalError)(nil)
)

// Error constructor functions for the verax package domains.
var (
	newError    = xrr.ErrorFactory[edError]()
	newInternal = xrr.ErrorFactory[edInternal]()
	fieldError  = xrr.FieldsFactory[edFields]()
)

// Error represents an error in the verax package error domain.
type Error = xrr.GenericError[edError]

// InternalError represents an internal error (library misuse) in the verax
// package error domain.
type InternalError = xrr.GenericError[edInternal]

// FieldsError represents a field error in the verax error domain.
type FieldsError = xrr.GenericFields[edFields]

// NewError returns a new error in the verax package error domain.
func NewError(msg, code string, opts ...xrr.Option) error {
	return newError(msg, code, opts...)
}

// NewInternalError returns a new internal error in the verax package domain.
func NewInternalError(msg, code string, opts ...xrr.Option) error {
	return newInternal(msg, code, opts...)
}

// NewFieldsErrors returns a new empty [FieldsError].
func NewFieldsErrors() *FieldsError {
	return &FieldsError{}
}

// FieldError returns a new field error in the verax package error domain.
func FieldError(field string, err error) error {
	return fieldError(field, err)
}

// FieldsErrors creates a new [FieldsError] from the given map. The map is
// stored directly without copying.
func FieldsErrors(fields map[string]error) error {
	return xrr.NewDomainFields[edFields](fields)
}

// IsError reports whether err belongs to the verax error domain, i.e. it is
// one of [Error], [InternalError], or [FieldsError].
func IsError(err error) bool {
	return IsValidationError(err) || IsInternalError(err)
}

// IsValidationError reports whether an err is a user-facing validation error
// i.e. [Error] or [FieldsError].
func IsValidationError(err error) bool {
	//goland:noinspection GoTypeAssertionOnErrors
	_, isFields := err.(*FieldsError)
	return isFields || xrr.IsDomain[edError](err)
}

// IsInternalError reports whether an err is an [InternalError].
func IsInternalError(err error) bool { return xrr.IsDomain[edInternal](err) }
