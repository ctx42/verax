// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"reflect"
)

// msgInvType represents an error message for an unexpected type.
const msgInvType = "not expected value type"

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

func (tpr TypeRule) Validate(have any) error {
	if !tpr.condition {
		return nil
	}
	if have == nil {
		return nil
	}
	if tpr.typ != reflect.TypeOf(have) {
		return NewError(tpr.msg, tpr.code)
	}
	return nil
}

func (tpr TypeRule) When(condition bool) TypeRule {
	tpr.condition = condition
	return tpr
}

func (tpr TypeRule) Message(msg string) TypeRule {
	if msg != "" {
		tpr.msg = msg
		tpr.flags |= flgCustomMsg
	}
	return tpr
}

func (tpr TypeRule) Code(code string) TypeRule {
	if code != "" {
		tpr.code = code
		tpr.flags |= flgCustomCode
	}
	return tpr
}
