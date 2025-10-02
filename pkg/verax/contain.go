// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"fmt"
	"reflect"

	"github.com/ctx42/xrr/pkg/xrr"
)

// Contain returns a validation rule that loops through iterable (map, slice
// or array) and validates it contains at least one given value.
func Contain(rule EqualRule) ContainRule { return ContainRule(rule) }

// ContainRule is a validation rule that validates there is at least one
// element in a map/slice/array using the specified [EqualRule].
type ContainRule EqualRule

// Validate loops through the given iterable and calls the Validate() method
// for each value with provided [EqualRule].
func (r ContainRule) Validate(v any) error {
	vo := reflect.ValueOf(v)

	var success bool
	switch vo.Kind() {
	case reflect.Map:
		for _, k := range vo.MapKeys() {
			val := getInterface(vo.MapIndex(k))
			if err := Validate(val, EqualRule(r)); err == nil {
				success = true
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < vo.Len(); i++ {
			val := getInterface(vo.Index(i))
			if err := Validate(val, EqualRule(r)); err == nil {
				success = true
			}
		}

	default:
		return xrr.New("must be an iterable", ECInvType)
	}

	if success {
		return nil
	}

	msg := fmt.Sprintf("must contain at least one '%v' value", r.want)
	return xrr.New(msg, ECNotEqual)
}
