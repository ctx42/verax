// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"reflect"

	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
)

// ContainRuleName represents [ContainRule] name.
const ContainRuleName = "contain-rule"

// Contain returns a validation rule that loops through iterable (map, slice,
// or array) and validates it contains at least one given value.
func Contain(rule EqualRule) ContainRule {
	return ContainRule{rule: rule, condition: true}
}

// Compile time checks.
var (
	_ customizer[ContainRule]  = ContainRule{}
	_ conditioner[ContainRule] = ContainRule{}
	_ Rule                     = ContainRule{}
)

// ContainRule is a validation rule that validates there is at least one
// element in a map/slice/array using the specified [EqualRule].
type ContainRule struct {
	rule      EqualRule // Equality rule used to match elements.
	condition bool      // Run validation only when true.
}

func (con ContainRule) Validate(have any) error {
	if con.rule.sticky != nil {
		return con.rule.sticky
	}
	if !con.condition {
		return nil
	}
	vo := reflect.ValueOf(have)

	switch vo.Kind() {
	case reflect.Invalid:
		// An untyped nil contains nothing.

	case reflect.Map:
		for _, k := range vo.MapKeys() {
			val := getInterface(vo.MapIndex(k))
			if con.rule.fn(con.rule.want, val) == nil {
				return nil
			}
		}

	case reflect.Slice, reflect.Array:
		for i := range vo.Len() {
			val := getInterface(vo.Index(i))
			if con.rule.fn(con.rule.want, val) == nil {
				return nil
			}
		}

	default:
		return NewInternalErrorf("must be iterable", xrr.WithCode(ECInvType))
	}

	code := ECNotEqual
	if con.rule.flags&flgCustomCode != 0 {
		code = con.rule.code
	}
	if con.rule.flags&flgCustomMsg != 0 {
		return NewError(con.rule.msg, code)
	}
	format := "must contain at least one '%v' value"
	return NewErrorf(format, con.rule.want, xrr.WithCode(code))
}

func (con ContainRule) When(condition bool) ContainRule {
	con.condition = condition
	con.rule = con.rule.When(condition)
	return con
}

func (con ContainRule) Code(code string) ContainRule {
	con.rule = con.rule.Code(code)
	return con
}

func (con ContainRule) Message(msg string) ContainRule {
	con.rule = con.rule.Message(msg)
	return con
}

func (con ContainRule) Spec() (*spec.Spec, error) {
	spc, err := con.rule.spec(ContainRuleName)
	if err != nil {
		return nil, err
	}
	return spc, nil
}

// ContainRuleFromSpec creates a new instance of [ContainRule] from the
// [spec.Spec].
func ContainRuleFromSpec(spc *spec.Spec) (ContainRule, error) {
	rule, err := equalRuleFromSpec(spc, ContainRuleName)
	if err != nil {
		return ContainRule{}, err
	}
	return Contain(rule), nil
}
