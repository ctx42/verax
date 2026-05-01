// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package vcfg

import (
	"encoding/json"

	"github.com/ctx42/xrr/pkg/xrr"
)

// edError is the marker type for the package's error domain.
type edError struct{}

// Compile checks.
var (
	_ error            = (*Error)(nil)
	_ xrr.Coder        = (*Error)(nil)
	_ json.Marshaler   = (*Error)(nil)
	_ json.Unmarshaler = (*Error)(nil)

	_ error            = (*FieldErrors)(nil)
	_ xrr.Fielder      = (*FieldErrors)(nil)
	_ json.Marshaler   = (*FieldErrors)(nil)
	_ json.Unmarshaler = (*FieldErrors)(nil)
)

// Error constructor functions for the verax package domains.
var (
	newError       = xrr.ErrorFunc[edError]()
	newFieldsError = xrr.FieldsFunc[edError]()
)

// Error represents an error in the verax package error domain.
type Error = xrr.GenericError[edError]

// NewError returns a new error in the verax package error domain.
func NewError(msg, code string, opts ...xrr.Option) error {
	return newError(msg, code, opts...)
}

// FieldErrors represents a field error in the verax error domain.
type FieldErrors = xrr.GenericFields[edError]

// NewFieldError returns a new field error in the verax package error domain.
func NewFieldError(field string, err error) *FieldErrors {
	return newFieldsError(field, err)
}

// NewFieldErrors creates a new [FieldErrors] from the given map. The map is
// stored directly without copying.
func NewFieldErrors(fields map[string]error) *FieldErrors {
	return xrr.NewFields[edError](fields)
}

// IsConfigError reports whether the error belongs to the vcfg package error
// domain.
func IsConfigError(err error) bool {
	// TODO(rz): test this. err != nil
	return err != nil && xrr.IsDomain[edError](err)
}
