// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"github.com/ctx42/xrr/pkg/xrr"
)

// ToAnySlice returns a slice of T as a slice of any.
func ToAnySlice[T any](vls ...T) []any {
	tmp := make([]any, 0, len(vls))
	for _, val := range vls {
		tmp = append(tmp, val)
	}
	return tmp
}

// EncloseError is a helper function to [xrr.Enclose] error with the
// [ErrValidation] error as the leading error. It returns nil if err is nil.
func EncloseError(err error) error {
	if err == nil {
		return nil
	}
	return xrr.Enclose(err, ErrValidation)
}

// setCode is a helper function to [xrr.Wrap] non-nil error with the given code.
// It returns nil if err is nil. Returns the same error if the code is empty.
func setCode(err error, code string) error {
	if err == nil {
		return nil
	}
	if code == "" {
		return err
	}
	if have := xrr.GetCode(err); have == code {
		return err
	}
	return xrr.Wrap(err, xrr.WithCode(code))
}
