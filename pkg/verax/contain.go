// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"fmt"
	"reflect"

	"github.com/ctx42/verax/pkg/spec"
)

// ContainRuleName represents [ContainRule] name.
const ContainRuleName = "contain-rule"

// Contain returns a validation rule that loops through iterable (map, slice,
// or array) and validates it contains at least one given value.
func Contain(rule EqualRule) ContainRule { return ContainRule{rule: rule} }

// Compile time checks.
var (
	_ customizer[ContainRule]  = ContainRule{}
	_ conditioner[ContainRule] = ContainRule{}
	_ Rule                     = ContainRule{}
)

// ContainRule is a validation rule that validates there is at least one
// element in a map/slice/array using the specified [EqualRule].
type ContainRule struct {
	rule EqualRule // Equality rule used to match elements.
}

func (r ContainRule) Validate(have any) error {
	vo := reflect.ValueOf(have)

	var success bool
	switch vo.Kind() {
	case reflect.Map:
		for _, k := range vo.MapKeys() {
			val := getInterface(vo.MapIndex(k))
			if err := Validate(val, r.rule); err == nil {
				success = true
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < vo.Len(); i++ {
			val := getInterface(vo.Index(i))
			if err := Validate(val, r.rule); err == nil {
				success = true
			}
		}

	default:
		return NewInternalError("must be iterable", ECInvType)
	}

	if success {
		return nil
	}

	msg := fmt.Sprintf("must contain at least one '%v' value", r.rule.want)
	return NewError(msg, ECNotEqual)
}

func (r ContainRule) When(condition bool) ContainRule {
	r.rule = r.rule.When(condition)
	return r
}

func (r ContainRule) Code(code string) ContainRule {
	r.rule = r.rule.Code(code)
	return r
}

func (r ContainRule) Message(msg string) ContainRule {
	r.rule = r.rule.Message(msg)
	return r
}

func (r ContainRule) Spec() (*spec.Spec, error) {
	spc, err := r.rule.spec(ContainRuleName)
	if err != nil {
		return nil, err
	}
	spc.Name = ContainRuleName
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
