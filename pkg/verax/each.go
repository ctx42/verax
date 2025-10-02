// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"reflect"
	"strconv"

	"github.com/ctx42/xrr/pkg/xrr"
)

// Each returns a validation rule that loops through an iterable (map, slice or
// array) and validates each value inside with the provided rules. Empty
// iterable is considered valid. Use the [Required] rule to make sure the
// iterable is not empty.
func Each(rules ...Rule) EachRule { return EachRule{rules: rules} }

// EachRule is a validation rule that validates elements in a map/slice/array
// using the specified list of rules.
type EachRule struct {
	rules []Rule
}

// Validate loops through the given iterable and calls the Validate() method
// for each value.
func (r EachRule) Validate(v any) error {
	var ers xrr.Fields

	vo := reflect.ValueOf(v)
	switch vo.Kind() {
	case reflect.Map:
		for _, k := range vo.MapKeys() {
			val := getInterface(vo.MapIndex(k))
			if err := Validate(val, r.rules...); err != nil {
				if ers == nil {
					ers = xrr.Fields{}
				}
				ers[mapErrKey(k)] = err
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < vo.Len(); i++ {
			val := getInterface(vo.Index(i))
			if err := Validate(val, r.rules...); err != nil {
				if ers == nil {
					ers = xrr.Fields{}
				}
				ers[strconv.Itoa(i)] = err
			}
		}

	default:
		return xrr.New("must be an iterable", ECInvType)
	}

	if len(ers) > 0 {
		return ers
	}
	return nil
}
