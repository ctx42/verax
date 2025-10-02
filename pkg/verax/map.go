// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"fmt"
	"reflect"

	"github.com/ctx42/xrr/pkg/xrr"
)

// ECMapKeyMissing represents error code for the missing required key.
const ECMapKeyMissing = "ECMapKeyMissing"

// ECMapKeyUnexpected represents error code for the not expected key.
const ECMapKeyUnexpected = "ECMapKeyUnexpected"

// MapRule sentinel errors.
var (
	// ErrNotMapPtr is the error that the value being validated is not a map.
	ErrNotMapPtr = xrr.New("only a map can be validated", ECInternal)

	// ErrInvKeyType is the error returned in case of an incorrect key type.
	ErrInvKeyType = xrr.New("key not the correct type", ECInternal)

	// ErrKeyMissing is the error returned in case of a missing key.
	ErrKeyMissing = xrr.New("required key is missing", ECMapKeyMissing)

	// ErrKeyUnexpected is the error returned in case of an unexpected key.
	ErrKeyUnexpected = xrr.New("key not expected", ECMapKeyUnexpected)
)

// MapRule represents a rule set associated with a map.
type MapRule struct {
	keys         map[any]*KeyRules
	allowUnknown bool
}

// KeyRules represents a rule set associated with a map key.
type KeyRules struct {
	key      any
	optional bool
	rules    []Rule
}

// Map returns a validation rule that checks the keys and values of a map.
// This rule should only be used for validating maps, or a validation error
// will be reported. Use Key() to specify map keys that need to be validated.
// Each Key() call specifies a single key which can be associated with multiple
// rules.
//
// For example,
//
//	vrule.Map(
//	    vrule.Key("Name", vrule.Required),
//	    vrule.Key("Value", vrule.Required, vrule.Length(5, 10)),
//	)
//
// A nil value is considered valid. Use the Required rule to make sure a map
// value is present.
func Map(keys ...*KeyRules) MapRule {
	kr := make(map[any]*KeyRules, len(keys))
	for _, k := range keys {
		kr[k.key] = k
	}
	return MapRule{keys: kr}
}

// AllowUnknown configures the rule to ignore unknown keys.
func (r MapRule) AllowUnknown() MapRule {
	r.allowUnknown = true
	return r
}

// IsOptional returns true if the given map key is optional. It will return
// true for keys that are not defined in the map.
func (r MapRule) IsOptional(key any) bool {
	if kr, ok := r.keys[key]; ok {
		return kr.optional
	}
	return true
}

// IsDefined returns true if the given map key is defined.
func (r MapRule) IsDefined(key any) bool {
	_, ok := r.keys[key]
	return ok
}

// Validate checks if the given value is valid or not.
//
// Returns error with ECInternal code on unexpected errors, otherwise it
// returns xrr.Fields error.
//
// nolint: cyclop, gocognit
func (r MapRule) Validate(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Map {
		return ErrNotMapPtr
	}
	if val.IsNil() {
		return nil
	}

	var ers xrr.Fields
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
				msg := fmt.Sprintf("%s: %s", getErrorKeyName(kr.key), err)
				return xrr.New(msg, ECInternal)
			}
			if ers == nil {
				ers = xrr.Fields{}
			}
			ers[getErrorKeyName(kr.key)] = err
		}
		if !r.allowUnknown {
			delete(extraKeys, kr.key)
		}
	}

	if !r.allowUnknown {
		if ers == nil {
			ers = xrr.Fields{}
		}
		for key := range extraKeys {
			ers[getErrorKeyName(key)] = ErrKeyUnexpected
		}
	}

	if len(ers) > 0 {
		return ers
	}
	return nil
}

// Key specifies a map key and the corresponding validation rules.
func Key(key any, rules ...Rule) *KeyRules {
	return &KeyRules{
		key:   key,
		rules: rules,
	}
}

// Optional configures the rule to ignore the key if missing.
func (r *KeyRules) Optional() *KeyRules {
	r.optional = true
	return r
}

// RequiredWhen sets key as required when the condition is true.
func (r *KeyRules) RequiredWhen(condition bool) *KeyRules {
	r.optional = !condition
	return r
}

// getErrorKeyName returns the name that should be used to represent
// the validation error of a map key.
func getErrorKeyName(key any) string { return fmt.Sprintf("%v", key) }
