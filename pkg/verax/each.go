// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"

	"github.com/ctx42/verax/pkg/spec"
)

// EachRuleName represents [EachRule] name.
const EachRuleName = "each-rule"

// Each returns a validation rule that loops through an iterable (map, slice,
// or array) and validates each value inside with the provided rules. Empty
// iterable is considered valid. Use the [Required] rule to make sure the
// iterable is not empty.
func Each(rules ...Rule) EachRule {
	return EachRule{
		condition: true,
		rules:     rules,
	}
}

// Compile time conditions.
var (
	_ conditioner[EachRule] = EachRule{}
	_ Rule                  = EachRule{}
)

// EachRule is a validation rule that validates elements in a map/slice/array
// using the specified list of rules.
type EachRule struct {
	condition bool   // Run validation only when true.
	rules     []Rule // Rules to apply to every iterable element.
}

func (r EachRule) Validate(have any) error {
	if !r.condition {
		return nil
	}
	var ers *FieldsError
	vo := reflect.ValueOf(have)
	switch vo.Kind() {
	case reflect.Map:
		for _, k := range vo.MapKeys() {
			val := getInterface(vo.MapIndex(k))
			if err := Validate(val, r.rules...); err != nil {
				if ers == nil {
					ers = NewFieldsErrors()
				}
				ers.Set(mapErrKey(k), err)
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < vo.Len(); i++ {
			val := getInterface(vo.Index(i))
			if err := Validate(val, r.rules...); err != nil {
				if ers == nil {
					ers = NewFieldsErrors()
				}
				ers.Set(strconv.Itoa(i), err)
			}
		}

	default:
		return NewInternalError("must be iterable", ECInvType)
	}

	if ers != nil {
		return ers
	}
	return nil
}

func (r EachRule) When(condition bool) EachRule {
	r.condition = condition
	return r
}

func (r EachRule) Spec() (*spec.Spec, error) {
	spc := spec.NewSpec(EachRuleName)
	if len(r.rules) > 0 {
		spc.SetArg(spec.ArgTypes, slices.Clone(r.rules))
	}
	return spc, nil
}

// EachRuleFromSpec creates a new instance of [EachRule] from the [spec.Spec].
func EachRuleFromSpec(spc *spec.Spec) (EachRule, error) {
	if spc.Name != EachRuleName {
		msg := fmt.Sprintf("%s: invalid spec name: %q", EachRuleName, spc.Name)
		return EachRule{}, NewInternalError(msg, spec.ECInvSpec)
	}
	rs, err := getArg[[]Rule](spc.Args, spec.ArgTypes, EachRuleName)
	if err != nil {
		return EachRule{}, err
	}
	return Each(rs...), nil
}
