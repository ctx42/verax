// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"reflect"

	"github.com/ctx42/xrr/pkg/xrr"
)

// ErrExpType represents error code for unexpected type.
var ErrExpType = xrr.New("not expected value type", ECInvType)

// Compile time checks.
var (
	_ Customizer[TypeRule]  = TypeRule{}
	_ Conditioner[TypeRule] = TypeRule{}
)

// Type creates a validation rule that checks if a value is of the same type.
func Type(typ reflect.Type) *TypeRule {
	return &TypeRule{typ: typ, condition: true, err: ErrExpType}
}

// TypeOf creates a validation rule that checks if a value is of the same type.
func TypeOf(typ any) *TypeRule {
	return &TypeRule{typ: reflect.TypeOf(typ), condition: true, err: ErrExpType}
}

// TypeRule represents rule checking a validated value is of the expected type.
type TypeRule struct {
	typ       reflect.Type // Expected type.
	condition bool         // Run validation only when true.
	err       error        // Validation error.
}

func (r TypeRule) Validate(v any) error {
	if !r.condition {
		return nil
	}
	if isNil, _ := IsNil(v); isNil {
		return nil
	}
	if r.typ != reflect.TypeOf(v) {
		return r.err
	}
	return nil
}

func (r TypeRule) When(condition bool) TypeRule {
	r.condition = condition
	return r
}

func (r TypeRule) Code(code string) TypeRule {
	r.err = setCode(r.err, code)
	return r
}

func (r TypeRule) Error(err error) TypeRule {
	r.err = err
	return r
}
