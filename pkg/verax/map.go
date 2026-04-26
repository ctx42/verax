// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
)

// Rule names.
const (
	// MapRuleName represents [MapRule] name.
	MapRuleName = "map-rule"

	// MapKeyName represents [MapKey] name.
	MapKeyName = "map-key"
)

// [MapRule] and [MapKey] error codes.
const (
	// ECMapKeyMissing represents error code for the missing required key.
	ECMapKeyMissing = "ECMapKeyMissing"

	// ECMapKeyUnexpected represents error code for the not expected key.
	ECMapKeyUnexpected = "ECMapKeyUnexpected"
)

// [MapRule] sentinel errors.
var (
	// ErrNotMapPtr is the error returned when the validated value is not a map.
	ErrNotMapPtr = NewInternalError("only a map can be validated", ECInternal)

	// ErrInvKeyType is the error returned when a map key type is incorrect.
	ErrInvKeyType = NewInternalError("the key type does not match the map", ECInternal)

	// ErrKeyMissing is the error returned when a required map key is missing.
	ErrKeyMissing = NewError("missing key", ECMapKeyMissing)

	// ErrKeyUnexpected is the error returned when a map key is unexpected.
	ErrKeyUnexpected = NewError("unexpected key", ECMapKeyUnexpected)
)

// Compile time conditions.
var (
	_ conditioner[MapRule] = MapRule{}
	_ Rule                 = MapRule{}
)

// MapRule represents a rule set associated with map types.
type MapRule struct {
	condition    bool     // Run validation only when true.
	allowUnknown bool     // Skip validation of keys not listed in keys.
	keys         []MapKey // Rules keyed by map key.
}

// MapKey represents a rule set associated with a map key.
type MapKey struct {
	key      any    // Map key.
	optional bool   // Is the key optional?
	rules    []Rule // Rules for the key value.
}

// Map returns a validation rule that conditions the keys and values of a map.
// This rule should only be used for validating maps, or a validation error
// will be reported. Use Key() to specify map keys that need to be validated.
// Each Key() call specifies a single key which can be associated with multiple
// rules.
//
// For example,
//
//	verax.Map(
//	    verax.Key("Name", verax.Required),
//	    verax.Key("Value", verax.Required, verax.Length(5, 10)),
//	)
//
// A nil value is considered valid. Use the [Required] rule to make sure a map
// value is present.
func Map(keys ...MapKey) MapRule { return MapRule{condition: true, keys: keys} }

// AllowUnknown configures the rule to ignore unknown keys.
func (r MapRule) AllowUnknown() MapRule {
	r.allowUnknown = true
	return r
}

// IsOptional returns true if the given map key is optional. It will return
// true for keys that are not defined in the map.
func (r MapRule) IsOptional(key any) bool {
	for _, mk := range r.keys {
		if mk.key == key {
			return mk.optional
		}
	}
	return true
}

// IsDefined returns true if the given map key is defined.
func (r MapRule) IsDefined(key any) bool {
	for _, mk := range r.keys {
		if mk.key == key {
			return true
		}
	}
	return false
}

// Validate conditions if the given value is valid or not.
//
// nolint: cyclop, gocognit
func (r MapRule) Validate(have any) error {
	val := reflect.ValueOf(have)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Map {
		return ErrNotMapPtr
	}
	if !r.condition {
		return nil
	}
	if val.IsNil() {
		return nil
	}

	var ers *FieldsError
	kt := val.Type().Key()

	var extraKeys map[any]bool
	if !r.allowUnknown {
		extraKeys = make(map[any]bool, val.Len())
		iter := val.MapRange()
		for iter.Next() {
			extraKeys[iter.Key().Interface()] = true
		}
	}

	for _, kr := range r.keys {
		var err error
		if kv := reflect.ValueOf(kr.key); !kt.AssignableTo(kv.Type()) {
			err = ErrInvKeyType
		} else if vv := val.MapIndex(kv); !vv.IsValid() {
			if !kr.optional {
				err = ErrKeyMissing
			}
		} else {
			err = Validate(vv.Interface(), kr.rules...)
		}

		if err != nil {
			if xrr.GetCode(err) == ECInternal {
				return FieldError(kr.KeyString(), err)
			}
			if ers == nil {
				ers = &FieldsError{}
			}
			ers.Set(kr.KeyString(), err)
		}
		if !r.allowUnknown {
			delete(extraKeys, kr.key)
		}
	}

	if !r.allowUnknown {
		for key := range extraKeys {
			if ers == nil {
				ers = &FieldsError{}
			}
			ers.Set(getErrorKeyName(key), ErrKeyUnexpected)
		}
	}

	if ers != nil {
		return ers
	}
	return nil
}

func (r MapRule) When(condition bool) MapRule {
	r.condition = condition
	return r
}

func (r MapRule) Spec() (*spec.Spec, error) {
	var sps []*spec.Spec
	for _, key := range r.keys {
		spc, err := key.Spec()
		if err != nil {
			msg := fmt.Sprintf("%s[%s]: %s", MapRuleName, key.KeyString(), err)
			return nil, NewInternalError(msg, ECInternal)
		}
		sps = append(sps, spc)
	}

	spc := spec.NewSpec(MapRuleName)
	if r.allowUnknown {
		spc.SetArg(ArgAllowUnk, true)
	}
	if len(sps) > 0 {
		spc.SetArg(spec.ArgSpecs, sps)
	}
	return spc, nil
}

// MapRuleFromSpec creates a new instance of [MapRule] from the [spec.Spec].
func MapRuleFromSpec(spc *spec.Spec) (MapRule, error) {
	if spc.Name != MapRuleName {
		msg := fmt.Sprintf("%s: invalid spec name: %q", MapRuleName, spc.Name)
		return MapRule{}, NewInternalError(msg, spec.ECInvSpec)
	}

	var mks []MapKey
	if spc.ArgExist(spec.ArgSpecs) {
		sps, err := getArg[[]*spec.Spec](spc.Args, spec.ArgSpecs, MapRuleName)
		if err != nil {
			return MapRule{}, err
		}
		for idx, keySpc := range sps {
			key, err := MapKeyFromSpec(keySpc)
			if err != nil {
				format := "%s: key-spec[%d]: %s"
				msg := fmt.Sprintf(format, MapRuleName, idx, err)
				return MapRule{}, NewInternalError(msg, spec.ECInvSpec)
			}
			mks = append(mks, key)
		}
	}

	rule := Map(mks...)

	if spc.ArgExist(ArgAllowUnk) {
		val, err := getArg[bool](spc.Args, ArgAllowUnk, MapRuleName)
		if err != nil {
			return MapRule{}, err
		}
		rule.allowUnknown = val
	}
	return rule, nil
}

// Key specifies a map key and the corresponding validation rules.
func Key(key any, rules ...Rule) MapKey {
	return MapKey{
		key:   key,
		rules: rules,
	}
}

// Optional configures the rule to ignore the key if missing.
func (mk MapKey) Optional() MapKey {
	mk.optional = true
	return mk
}

// When marks the key as required when condition is true, optional otherwise.
func (mk MapKey) When(condition bool) MapKey {
	mk.optional = !condition
	return mk
}

// KeyString returns the string representation of the key.
func (mk MapKey) KeyString() string { return fmt.Sprintf("%v", mk.key) }

func (mk MapKey) Spec() (*spec.Spec, error) {
	spc := spec.NewSpec(MapKeyName).SetArg(spec.ArgValue, mk.key)
	if len(mk.rules) > 0 {
		spc.SetArg(spec.ArgTypes, slices.Clone(mk.rules))
	}
	if mk.optional {
		spc.SetArg(ArgOptional, mk.optional)
	}
	return spc, nil
}

// MapKeyFromSpec creates a new instance of [MapKey] from the [spec.Spec].
func MapKeyFromSpec(spc *spec.Spec) (MapKey, error) {
	if spc.Name != MapKeyName {
		msg := fmt.Sprintf("%s: invalid spec name: %q", MapKeyName, spc.Name)
		return MapKey{}, NewInternalError(msg, spec.ECInvSpec)
	}
	key, err := getArg[any](spc.Args, spec.ArgValue, MapKeyName)
	if err != nil {
		return MapKey{}, err
	}
	rs, err := getArg[[]Rule](spc.Args, spec.ArgTypes, MapKeyName)
	if err != nil {
		return MapKey{}, err
	}
	rule := Key(key, rs...)

	if spc.ArgExist(ArgOptional) {
		opt, err := getArg[bool](spc.Args, ArgOptional, MapKeyName)
		if err != nil {
			return MapKey{}, err
		}
		if opt {
			rule = rule.Optional()
		}
	}
	return rule, nil
}

// getErrorKeyName returns the name that should be used to represent
// the validation error of a map key.
func getErrorKeyName(key any) string { return fmt.Sprintf("%v", key) }
