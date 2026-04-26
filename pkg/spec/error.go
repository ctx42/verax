// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package spec

import (
	"encoding/json"

	"github.com/ctx42/xrr/pkg/xrr"
)

// Unexported zero-size marker type used as the domain type parameter.
type edSpec struct{}

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
)

// Error constructor functions for the spec package domain.
var (
	newError   = xrr.ErrorFactory[edSpec]()
	fieldError = xrr.FieldsFactory[edSpec]()
)

// Error represents an error in the spec package error domain.
type Error = xrr.GenericError[edSpec]

// FieldsError represents a field error in the spec error domain.
type FieldsError = xrr.GenericFields[edSpec]

// NewError returns a new error in the spec package error domain.
func NewError(msg, code string, opts ...xrr.Option) error {
	return newError(msg, code, opts...)
}

// FieldError returns a new field error in the spec package error domain.
func FieldError(field string, err error) error { return fieldError(field, err) }
