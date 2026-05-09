// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"encoding/json"

	"github.com/ctx42/xrr/pkg/xrr"
)

// Marker types for the package's error domain.
type (
	edError    struct{}
	edInternal struct{}
)

// Compile checks.
var (
	_ error            = (*Error)(nil)
	_ xrr.Coder        = (*Error)(nil)
	_ json.Marshaler   = (*Error)(nil)
	_ json.Unmarshaler = (*Error)(nil)

	_ error            = (*InternalError)(nil)
	_ xrr.Coder        = (*InternalError)(nil)
	_ json.Marshaler   = (*InternalError)(nil)
	_ json.Unmarshaler = (*InternalError)(nil)

	_ error            = (*FieldErrors)(nil)
	_ xrr.Fielder      = (*FieldErrors)(nil)
	_ json.Marshaler   = (*FieldErrors)(nil)
	_ json.Unmarshaler = (*FieldErrors)(nil)
)

// Error constructor functions for the package's error domain.
var (
	newError          = xrr.ErrorFunc[edError]()
	newErrorf         = xrr.ErrorfFunc[edError]()
	newInternalError  = xrr.ErrorFunc[edInternal]()
	newInternalErrorf = xrr.ErrorfFunc[edInternal]()
	newFieldsError    = xrr.FieldsFunc[edError]()
)

// Error represents an error in the package's error domain.
type Error = xrr.GenericError[edError]

// NewError creates a new [Error] with the given message. The args may contain
// an optional string error code and any number of [xrr.Option] values in any
// order.
//
// Examples:
//
//	NewError("message")
//	NewError("message", "ECode")
//	NewError("message", "ECode", xrr.Option())
//
// When [xrr.WithCause] is provided:
//   - If msg is empty, Error() returns the cause's message directly.
//   - If msg is non-empty, Error() returns "msg: cause message".
//   - If no code is given and [xrr.WithCode] is not provided, the cause's
//     code is inherited via [xrr.GetCode]. Pass a code string or
//     [xrr.WithCode] to override it.
//
// For wrapping without a new message, prefer [xrr.Wrap], which makes the
// intent clearer.
func NewError(msg string, args ...any) error {
	return newError(msg, args...)
}

// InternalError represents an internal error (library misuse) in the package's
// error domain.
type InternalError = xrr.GenericError[edInternal]

// NewInternalError creates a new [InternalError] with the given message. The
// args may contain an optional string error code and any number of [xrr.Option]
// values in any order.
//
// Examples:
//
//	NewInternalError("message")
//	NewInternalError("message", "ECode")
//	NewInternalError("message", "ECode", xrr.Option())
//
// When [xrr.WithCause] is provided:
//   - If msg is empty, Error() returns the cause's message directly.
//   - If msg is non-empty, Error() returns "msg: cause message".
//   - If no code is given and [xrr.WithCode] is not provided, the cause's
//     code is inherited via [xrr.GetCode]. Pass a code string or
//     [xrr.WithCode] to override it.
//
// For wrapping without a new message, prefer [xrr.Wrap], which makes the
// intent clearer.
func NewInternalError(msg string, args ...any) error {
	return newInternalError(msg, args...)
}

// NewInternalErrorf creates a new [InternalError] using a format string. It is
// the format-style counterpart of [NewInternalError]: non-[xrr.Option] args are
// passed to the format string, while [xrr.Option] values are applied to the
// error. Unlike [NewInternalError], a bare string argument is treated as a
// format argument, not an error code — pass [xrr.WithCode] to set the code.
//
// When the format string contains %w, the error is created via [fmt.Errorf]
// and stored as the cause; [xrr.GenericError.Error] delegates to it. Without
// %w, the message is set to fmt.Sprintf(format, args...).
//
// Examples:
//
//	NewInternalErrorf("user %d not found", userID)
//	NewInternalErrorf("user %d not found", userID, xrr.WithCode("ECode"))
//	NewInternalErrorf("connect failed: %w", err)
//	NewInternalErrorf("connect failed: %w", err, xrr.WithCode("ECode"))
func NewInternalErrorf(format string, args ...any) error {
	return newInternalErrorf(format, args...)
}

// NewErrorf creates a new [Error] using a format string. It is the
// format-style counterpart of [NewError]: non-[xrr.Option] args are passed to
// the format string, while [xrr.Option] values are applied to the error.
// Unlike [NewError], a bare string argument is treated as a format argument,
// not an error code — pass [xrr.WithCode] to set the code.
//
// When the format string contains %w, the error is created via [fmt.Errorf]
// and stored as the cause; [xrr.GenericError.Error] delegates to it. Without
// %w, the message is set to fmt.Sprintf(format, args...).
//
// Examples:
//
//	NewErrorf("user %d not found", userID)
//	NewErrorf("user %d not found", userID, xrr.WithCode("ECode"))
//	NewErrorf("connect failed: %w", err)
//	NewErrorf("connect failed: %w", err, xrr.WithCode("ECode"))
func NewErrorf(format string, args ...any) error {
	return newErrorf(format, args...)
}

// FieldErrors represents a field error in the package's error domain.
type FieldErrors = xrr.GenericFields[edError]

// NewFieldError returns a new field error in the package's error domain.
func NewFieldError(field string, err error) *FieldErrors {
	return newFieldsError(field, err)
}

// NewFieldErrors creates a new [FieldErrors] from the given map. The map is
// stored directly without copying.
func NewFieldErrors(fields map[string]error) *FieldErrors {
	return xrr.NewFields[edError](fields)
}

// IsVeraxError reports whether the error is non-nil [Error], [FieldErrors] or
// [InternalError].
func IsVeraxError(err error) bool {
	return IsValidationError(err) || IsInternalError(err)
}

// IsValidationError reports whether the error is non-nil [Error] or
// [FieldErrors].
func IsValidationError(err error) bool {
	return err != nil && xrr.IsDomain[edError](err)
}

// IsInternalError reports whether the error is non-nil [InternalError].
func IsInternalError(err error) bool {
	return err != nil && xrr.IsDomain[edInternal](err)
}
