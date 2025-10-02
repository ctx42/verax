// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

// Package verax provides configurable and extensible rules for validating data
// of various types.
package verax

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/ctx42/xrr/pkg/xrr"
)

// Package level error codes.
const (
	// ECInternal represents error code for internal error (library misuse).
	ECInternal = "ECInternal"

	// ECInvType represents error code for an unexpected type.
	ECInvType = "ECInvType"

	// ECInvFormat represents error code for invalid value format.
	ECInvFormat = "ECInvFormat"

	// ECInvValue represents error code for invalid value.
	ECInvValue = "ECInvValue"

	// ECEmpty represents error code for empty value.
	ECEmpty = "ECEmpty"

	// ECMissing represents error code for missing value.
	ECMissing = "ECMissing"

	// ECFound represents error code for found value.
	ECFound = "ECFound"

	// ECNotFound represents error code for not found value.
	ECNotFound = "ECNotFound"

	// ECValidation represents validation error code.
	ECValidation = "ECValidation"

	// ECUnkRule represents unknown rule error code.
	ECUnkRule = "ECUnkRule"
)

var (
	// ErrorTag is the default struct tag name used to customize the error
	// field name for a struct field.
	ErrorTag = "json"

	// ErrInvSetup represents error for wrong rule setup.
	ErrInvSetup = xrr.New("invalid setup", ECInternal)

	// ErrInvType is returned in case of an unexpected value type.
	ErrInvType = xrr.New("unexpected value type", ECInvType)

	// ErrValidation represents validation error.
	ErrValidation = xrr.New("validation error", ECValidation)

	// ErrUnkRule represents unknown rule error.
	ErrUnkRule = xrr.New("unknown rule", ECUnkRule)
)

// Validator wraps Validate method.
type Validator interface {
	// Validate validates implementer and returns an error on failure.
	Validate() error
}

// WithValidator wraps ValidateWith method which validates implementor using
// the given [Rule].
type WithValidator interface {
	ValidateWith(rule Rule) error
}

// Rule represents a validation rule.
type Rule interface {
	// Validate validates a value and returns a value if validation fails.
	Validate(v any) error
}

// Customizer is a generic interface for rule modification.
type Customizer[T any] interface {
	// Code sets custom error code to set on the rule error when it fails.
	Code(code string) T

	// Error sets custom error to be returned when a rule fails.
	Error(err error) T
}

// Conditioner wraps the When method to control validation execution.
type Conditioner[T any] interface {
	// When sets a condition to determine if validation should run. If false,
	// validation is skipped, and no errors are returned.
	When(condition bool) T
}

// Set groups multiple validation rules and implements the [Rule] interface.
type Set []Rule

func (rg Set) Validate(value any) error { return Validate(value, rg...) }

// RuleFunc represents a validator function.
// You may wrap it as a [Rule] by calling By().
type RuleFunc func(v any) error

var validatableType = reflect.TypeOf((*Validator)(nil)).Elem()

// Validate checks the given value against the provided validation rules.
// Returns nil if all rules pass, or the first validation error encountered.
// Skips validation if one of the rules is [Skip]. Supports types implementing
// [Validator] or [WithValidator], and recursively validates maps, slices,
// arrays, pointers, or interfaces with validatable elements. Returns nil for
// nil pointers or interfaces.
//
// nolint: cyclop
func Validate(v any, rules ...Rule) error {
	for _, rule := range rules {
		if s, ok := rule.(skipRule); ok && bool(s) {
			return nil
		}
		if red, ok := v.(WithValidator); ok {
			if err := red.ValidateWith(rule); err != nil {
				return err
			}
			continue
		}
		if err := rule.Validate(v); err != nil {
			return err
		}
	}

	rv := reflect.ValueOf(v)
	if (rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface) &&
		rv.IsNil() {

		return nil
	}

	if vi, ok := v.(Validator); ok {
		return vi.Validate()
	}

	//goland:noinspection GoSwitchMissingCasesForIotaConsts
	switch rv.Kind() { // nolint: exhaustive
	case reflect.Map:
		if rv.Type().Elem().Implements(validatableType) {
			return validateMap(rv)
		}

	case reflect.Slice, reflect.Array:
		if rv.Type().Elem().Implements(validatableType) {
			return validateSlice(rv)
		}

	case reflect.Ptr, reflect.Interface:
		return Validate(rv.Elem().Interface())
	}

	return nil
}

// ValidateNamed validates v using the provided rules, wrapping any error in
// [xrr.Fields] with the specified field name.
func ValidateNamed(name string, v any, rules ...Rule) error {
	return xrr.Fields{name: Set(rules).Validate(v)}.Filter()
}

// validateMap validates a map of validatable elements.
func validateMap(rv reflect.Value) error {
	var ers xrr.Fields
	for _, key := range rv.MapKeys() {
		if mv := rv.MapIndex(key).Interface(); mv != nil {
			// nolint: forcetypeassert
			if err := mv.(Validator).Validate(); err != nil {
				if ers == nil {
					ers = xrr.Fields{}
				}
				ers[fmt.Sprintf("%v", key.Interface())] = err
			}
		}
	}
	if len(ers) > 0 {
		return ers
	}
	return nil
}

// validateSlice validates a slice/array of validatable elements.
func validateSlice(rv reflect.Value) error {
	var ers xrr.Fields
	l := rv.Len()
	for i := 0; i < l; i++ {
		if ev := rv.Index(i).Interface(); ev != nil {
			// nolint: forcetypeassert
			if err := ev.(Validator).Validate(); err != nil {
				if ers == nil {
					ers = xrr.Fields{}
				}
				ers[strconv.Itoa(i)] = err
			}
		}
	}
	if len(ers) > 0 {
		return ers
	}
	return nil
}

// Named represents a collection of named rules.
type Named map[string]Rule

// NewNamed returns new instance of Names.
func NewNamed() Named {
	return make(map[string]Rule)
}

// Set sets named rule.
func (n Named) Set(name string, rule Rule) Named {
	n[name] = rule
	return n
}

// Get returns the named rule or nil if it doesn't exist.
func (n Named) Get(name string) Rule {
	if rule, ok := n[name]; ok {
		return rule
	}
	return nil
}

// GetOrError returns the named rule; when it doesn't exist, it returns [Error]
// rule with [ErrUnkRule] error.
func (n Named) GetOrError(name string) Rule {
	if r := n.Get(name); r != nil {
		return r
	}
	return Error(ErrUnkRule)
}

// GetOrNoop returns the named rule or [Noop] rule if it doesn't exist.
func (n Named) GetOrNoop(name string) Rule {
	if rule, ok := n[name]; ok {
		return rule
	}
	return Noop
}
