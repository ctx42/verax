// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package verax provides configurable and extensible rules for validating data
// of various types.
package verax

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"

	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
)

// Package level error codes.
const (
	// ECInternal is the error code for internal error (library misuse).
	ECInternal = "ECInternal"

	// ECInvType is the error code for an unexpected type.
	ECInvType = "ECInvType"

	// ECInvFormat is the error code for an invalid value format.
	ECInvFormat = "ECInvFormat"

	// ECInvValue is the error code for invalid value.
	ECInvValue = "ECInvValue"

	// ECUnkRule represents an unknown rule error code.
	ECUnkRule = "ECUnkRule"

	// ECInvRuleMode is error code used when a [spec.Spec] defines unsupported
	// rule mode.
	ECInvRuleMode = "ECInvRuleMode"
)

// Customization flags.
const (
	flgCustomMsg  uint8 = 0b0000_0001 // Rule uses a custom error message.
	flgCustomCode uint8 = 0b0000_0010 // Rule uses a custom error code.
	flgCustomFn   uint8 = 0b0000_0100 // Rule uses a custom check function.
)

// ErrorTag is the default struct tag name used to customize the error
// field name for a struct field.
const ErrorTag = "json"

// Package level sentinel errors.
var (
	// ErrInvType is returned in case of an unexpected value type.
	ErrInvType = NewInternalError("unexpected value type", ECInvType)
)

// Validator wraps Validate method.
type Validator interface {
	// Validate validates the implementer and returns an error on failure.
	Validate() error
}

// WithValidator wraps ValidateWith method which validates implementor using
// the given [Rule].
type WithValidator interface {
	ValidateWith(rule Rule) error
}

// Rule represents a validation rule.
type Rule interface {
	// Validate checks the value against the rule's condition(s) and returns
	// a validation error if it fails.
	//
	// Implementations should return [Error] instance (or nil on success) to
	// support error targeting with [errors.As] and [errors.Is].
	//
	// For rules that delegate to custom validation functions ([RuleFunc],
	// [EqualFunc], [CompareFunc], etc.):
	//
	// - If both message and code are overridden: return a new [Error] instance.
	// - If only the message is overridden: return a new [Error] instance with
	//   the original error's code and the custom message.
	// - If only the code is overridden: return the original error after
	//   calling [SetCode] on it.
	// - If neither is overridden: return the error from the validation
	//   function unchanged.
	Validate(have any) error
}

// SpecableRule is implemented by rules that can both validate values and
// describe themselves as a [spec.Spec]. Implement this interface to make a
// rule serializable, for example, when storing validation configuration in a
// database or exchanging it over an API.
type SpecableRule interface {
	spec.Specable
	Rule
}

// Builders returns a map of builder definitions for all built-in rule specs.
// The returned map is a new instance and can be safely modified by callers.
func Builders() map[string]spec.Builder[Rule] {
	return map[string]spec.Builder[Rule]{
		AbsentRuleName:   AsRuleBuilder(AbsentRuleFromSpec),
		ByRuleName:       AsRuleBuilder(ByRuleFromSpec),
		ContainRuleName:  AsRuleBuilder(ContainRuleFromSpec),
		EachRuleName:     AsRuleBuilder(EachRuleFromSpec),
		EqualRuleName:    AsRuleBuilder(EqualRuleFromSpec),
		FailRuleName:     AsRuleBuilder(FailRuleFromSpec),
		InRuleName:       AsRuleBuilder(InRuleFromSpec),
		LengthRuleName:   AsRuleBuilder(LengthRuleFromSpec),
		MapRuleName:      AsRuleBuilder(MapRuleFromSpec),
		MatchRuleName:    AsRuleBuilder(MatchRuleFromSpec),
		NoopRuleName:     AsRuleBuilder(NoopRuleFromSpec),
		RequiredRuleName: AsRuleBuilder(RequiredRuleFromSpec),
		SkipRuleName:     AsRuleBuilder(SkipRuleFromSpec),
		RangeRuleName:    AsRuleBuilder(RangeRuleFromSpec),

		// [TypeRule] is intentionally excluded: it holds a [reflect.Type]
		// which has no portable cross-language representation.
		SetRuleName: AsRuleBuilder(SetRuleFromSpec),
	}
}

// Spec argument names used in rule serialization.
const (
	ArgErrMsg   = "err_msg"       // Custom error message.
	ArgErrCode  = "err_code"      // Custom error code.
	ArgMode     = "mode"          // Rule mode (e.g., exclusive, inclusive).
	ArgMin      = "min"           // Minimum value.
	ArgMax      = "max"           // Maximum value.
	ArgOptional = "optional"      // Optional value flag.
	ArgAllowUnk = "allow_unknown" // Allow unknown map keys.
)

// SetRuleName represents [Set] name.
const SetRuleName = "set-rule"

// Set groups multiple validation rules and implements the [Rule] interface.
type Set []Rule

func (set Set) Validate(have any) error { return Validate(have, set...) }

func (set Set) Spec() (*spec.Spec, error) {
	spc := spec.NewSpec(SetRuleName)
	if len(set) > 0 {
		spc.SetArg(spec.ArgTypes, slices.Clone([]Rule(set)))
	}
	return spc, nil
}

// SetRuleFromSpec creates a new instance of [Set] from the [spec.Spec].
func SetRuleFromSpec(spc *spec.Spec) (Set, error) {
	if spc.Name != SetRuleName {
		return nil, NewInternalErrorf(
			"%s: invalid spec name: %q",
			SetRuleName,
			spc.Name,
			xrr.WithCode(spec.ECInvSpec),
		)
	}
	rs, err := getArg[[]Rule](spc.Args, spec.ArgTypes, SetRuleName)
	if err != nil {
		return nil, err
	}
	return Set(rs), nil
}

// RuleFunc represents a custom validation function.
//
// It receives a value and must return nil on success, an [Error] for
// validation failures, or an [InternalError] for misuse such as unsupported
// or mismatched types.
//
// Wrap it into a [Rule] using [By].
type RuleFunc func(v any) error

// IsFunc represents a validator function returning true only for valid values.
// You may wrap it as a [Rule] by calling [Check].
type IsFunc[T any] func(v T) bool

// EqualFunc is the signature for a custom comparison function. It receives the
// expected value (want) and the value under validation (have) and returns nil
// when the comparison criterion is satisfied. When the criterion is violated,
// it should return an [Error]; which condition counts as a violation depends
// on the context the function is used in. For example, in an equality context
// the function returns an error when "want" and "have" differ, whereas in an
// inequality context it returns an error when they are equal. It must return
// an [InternalError] for misuse such as unsupported or mismatched types.
type EqualFunc func(want, have any) error

// Check creates a [RuleFunc] from [IsFunc].
func Check[T any](fn IsFunc[T], msg, code string) RuleFunc {
	return func(have any) error {
		vt, ok := have.(T)
		if !ok {
			return NewInternalErrorf(
				msg+": expected %T, got %T",
				vt,
				have,
				xrr.WithCode(ECInvType),
			)
		}
		if !fn(vt) {
			return NewError(msg, code)
		}
		return nil
	}
}

// CompareFunc is a custom validation function that compares two values and
// returns the result of the comparison. The integer result is the same as in
// the [cmp.Compare] function.
//
//	-1 if "want" is less than "have",
//	 0 if "want" equals "have",
//	+1 if "want" is greater than "have".
//
// The error result must be an [InternalError] for misuse such as unsupported
// or mismatched types, and should be an [Error] for comparison failures.
type CompareFunc func(want, have any) (int, error)

var validatableType = reflect.TypeFor[Validator]()

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
		if s, ok := rule.(SkipRule); ok && bool(s) {
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

	// Fast path: primitive built-in types can never implement Validator or
	// require composite-type traversal.
	switch v.(type) {
	case string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, uintptr,
		float32, float64:
		return nil
	}

	rv := reflect.ValueOf(v)
	if (rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface) &&
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

	case reflect.Pointer, reflect.Interface:
		return Validate(rv.Elem().Interface())
	}

	return nil
}

// ValidateNamed validates v using the provided rules, wrapping any error in
// [FieldErrors] with the specified field name.
func ValidateNamed(name string, have any, rules ...Rule) error {
	if err := Set(rules).Validate(have); err != nil {
		return NewFieldError(name, err)
	}
	return nil
}

// validateMap validates a map of validatable elements.
func validateMap(rv reflect.Value) error {
	var ers *FieldErrors
	for _, key := range rv.MapKeys() {
		if mv := rv.MapIndex(key).Interface(); mv != nil {
			// nolint: forcetypeassert
			if err := mv.(Validator).Validate(); err != nil {
				if ers == nil {
					ers = &FieldErrors{}
				}
				ers.Set(fmt.Sprintf("%v", key.Interface()), err)
			}
		}
	}
	if ers != nil {
		return ers
	}
	return nil
}

// validateSlice validates a slice/array of validatable elements.
func validateSlice(rv reflect.Value) error {
	var ers *FieldErrors
	for i := range rv.Len() {
		if ev := rv.Index(i).Interface(); ev != nil {
			// nolint: forcetypeassert
			if err := ev.(Validator).Validate(); err != nil {
				if ers == nil {
					ers = &FieldErrors{}
				}
				ers.Set(strconv.Itoa(i), err)
			}
		}
	}
	if ers != nil {
		return ers
	}
	return nil
}

// Named represents a collection of named rules.
type Named map[string]Rule

// NewNamed creates an empty [Named].
func NewNamed() Named {
	return make(map[string]Rule)
}

// Set associates the rule with the name. Chainable.
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

// GetOrError returns the named rule; when it doesn't exist, it returns an
// instance of [Fail] with [ECUnkRule] error code.
func (n Named) GetOrError(name string) Rule {
	if r := n.Get(name); r != nil {
		return r
	}
	return Fail("unknown rule", ECUnkRule)
}

// GetOrNoop returns the named rule or [Noop] rule if it doesn't exist.
func (n Named) GetOrNoop(name string) Rule {
	if rule, ok := n[name]; ok {
		return rule
	}
	return Noop
}

// customizer is a generic interface for rule modification.
type customizer[T any] interface {
	// Code sets a custom error code for rule failure.
	//
	// Empty codes are ignored (NOOP).
	Code(code string) T

	// Message sets a custom error message for rule failure.
	//
	// Empty messages are ignored (NOOP).
	Message(msg string) T
}

// conditioner wraps the When method to control validation execution.
type conditioner[T any] interface {
	// When sets a condition to determine if validation should run. If false,
	// validation is skipped, and no errors are returned.
	When(condition bool) T
}
