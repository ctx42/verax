// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"reflect"
)

// msgInvType represents an error message for an unexpected type.
var msgInvType = "not expected value type"

// Compile time conditions.
var (
	_ customizer[TypeRule]  = TypeRule{}
	_ conditioner[TypeRule] = TypeRule{}
	_ Rule                  = TypeRule{}
)

// Type is a rule validating a value is of the given [reflect.Type].
func Type(typ reflect.Type) TypeRule {
	return TypeRule{typ: typ, condition: true, msg: msgInvType, code: ECInvType}
}

// TypeOf is a rule validating a value is of the same type as the argument.
func TypeOf(val any) TypeRule { return Type(reflect.TypeOf(val)) }

// TypeRule is a rule validating a value is of the expected type.
type TypeRule struct {
	typ       reflect.Type // Expected type.
	condition bool         // Run validation only when true.
	msg       string       // Validation error message.
	code      string       // Validation error code.
	flags     uint8        // Customizations.
}

func (r TypeRule) Validate(have any) error {
	if !r.condition {
		return nil
	}
	if isNil := IsNil(have); isNil {
		return nil
	}
	if r.typ != reflect.TypeOf(have) {
		return NewError(r.msg, r.code)
	}
	return nil
}

func (r TypeRule) When(condition bool) TypeRule {
	r.condition = condition
	return r
}

func (r TypeRule) Message(msg string) TypeRule {
	if msg != "" {
		r.msg = msg
		r.flags |= flgCustomMsg
	}
	return r
}

func (r TypeRule) Code(code string) TypeRule {
	if code != "" {
		r.code = code
		r.flags |= flgCustomCode
	}
	return r
}
