// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"bytes"
	"cmp"
	"fmt"
	"reflect"
	"text/template"
	"time"

	"github.com/ctx42/xrr/pkg/xrr"
)

// Length rule error codes.
const (
	// ECInvThreshold represents error code for the invalid threshold.
	ECInvThreshold = "ECInvThreshold"
)

// Message templates for Min and Max rules.
var (
	// tplMinGreaterEqualThan is an error message template for no less than.
	tplMinGreaterEqualThan = emtpl("must be no less than {{.threshold}}")

	// tplMinGreaterThan is an error message template for greater than.
	tplMinGreaterThan = emtpl("must be greater than {{.threshold}}")

	// tplMaxLessEqualThan is an error message template for no greater than.
	tplMaxLessEqualThan = emtpl("must be no greater than {{.threshold}}")

	// tplMaxLessThan is an error message template for less than.
	tplMaxLessThan = emtpl("must be less than {{.threshold}}")
)

// CompareFunc is a function that compares two values and returns the result of
// the comparison. The integer result is the same as in the [cmp.Compare]
// function.
//
//	-1 if "want" is less than "have",
//	 0 if "want" equals "have",
//	+1 if "want" is greater than "have".
//
// The error result is used to return an error in the case of unexpected types.
type CompareFunc func(want, have any) (int, error)

// Min creates a validation rule that checks if a value is greater than or
// equal to the specified threshold. Use [ThresholdRule.Exclusive] to enforce a
// strict greater-than check. The value being checked and the threshold must be
// of the same type, supporting only int, uint, float, and time.Time types.
// Empty values are considered valid; use the [Required] rule to ensure a value
// is not empty.
//
// Example:
//
//	rule := Min(10)             // Value must be >= 10
//	rule := Min(10).Exclusive() // Value must be > 10
func Min(minimum any) ThresholdRule {
	return ThresholdRule{
		threshold: minimum,
		operator:  greaterEqualThan,
		with:      compareFor(minimum),
		condition: true,
		errTpl:    tplMinGreaterEqualThan,
		code:      ECInvThreshold,
	}
}

// Max creates a validation rule that checks if a value is less than or equal
// to the specified threshold. Use [ThresholdRule.Exclusive] to enforce a
// strict less-than check. The value being checked and the threshold must be of
// the same type, supporting only int, uint, float, and time.Time types. Empty
// values are considered valid; use the [Required] rule to ensure a value is
// not empty.
//
// Example:
//
//	rule := Max(100)             // Value must be <= 100
//	rule := Max(100).Exclusive() // Value must be < 100
func Max(maximum any) ThresholdRule {
	return ThresholdRule{
		threshold: maximum,
		operator:  lessEqualThan,
		with:      compareFor(maximum),
		condition: true,
		errTpl:    tplMaxLessEqualThan,
		code:      ECInvThreshold,
	}
}

// Compile time checks.
var (
	_ Customizer[ThresholdRule]  = ThresholdRule{}
	_ Conditioner[ThresholdRule] = ThresholdRule{}
)

// ThresholdRule is a rule validating a value satisfies a given threshold.
type ThresholdRule struct {
	threshold any                // The threshold value.
	operator  int                // The comparison operator.
	with      CompareFunc        // Comparison function.
	condition bool               // Run validation only when true.
	errTpl    *template.Template // Error template.
	err       error              // Custom error.
	code      string             // Error code.
}

// Comparison operation definitions.
const (
	greaterThan      = iota // Threshold must be greater than a value.
	greaterEqualThan        // Threshold must be greater or equal than a value.
	lessThan                // Threshold must be less than a value.
	lessEqualThan           // Threshold must be less or equal than a value.
)

// Exclusive modifies a [ThresholdRule] to exclude the boundary value,
// enforcing a strict comparison. For example, when used with [Min], it checks
// if a value is strictly greater than the threshold, and with [Max], it checks
// if a value is strictly less than the threshold. The value and threshold must
// be of the same type.
//
// Example:
//
//	ruleMin := Min(10).Exclusive()  // Value must be > 10
//	ruleMax := Max(100).Exclusive() // Value must be < 100
func (r ThresholdRule) Exclusive() ThresholdRule {
	if r.operator == greaterEqualThan {
		r.operator = greaterThan
		r.errTpl = tplMinGreaterThan
	} else if r.operator == lessEqualThan {
		r.operator = lessThan
		r.errTpl = tplMaxLessThan
	}
	return r
}

// With sets a custom comparison function for a [ThresholdRule], overriding the
// default comparison behavior. The function, of type [CompareFunc], defines
// how the value is compared to the threshold. This is useful for custom
// validation logic.
//
// When this function is used, the configuration set by [ThresholdRule.Code]
// and [ThresholdRule.Error] is ignored.
//
// The function must return an error if the type is not supported.
//
// Example:
//
//	cmpMyType := func(a, b MyType) int { ... }
//	rule := Min(myTypeValue).With(cmpMyType)
func (r ThresholdRule) With(cmp CompareFunc) ThresholdRule {
	r.with = cmp
	return r
}

// Validate checks if the given value is valid or not.
func (r ThresholdRule) Validate(v any) error {
	if !r.condition {
		return nil
	}
	if r.with == nil {
		msg := fmt.Sprintf("type is not supported: %T", r.threshold)
		return xrr.New(msg, ECInvType)
	}

	if isNil, _ := IsNil(v); isNil {
		return nil
	}

	if IsEmpty(v) {
		return nil
	}

	res, err := r.with(r.threshold, v)
	if err != nil {
		return err
	}
	if !thresholdOutcome(r.operator, res) {
		if r.err != nil {
			return r.err
		}
		return thresholdError(r.threshold, r.errTpl, r.code)
	}
	return nil
}

// thresholdError constructs threshold error.
func thresholdError(th any, tpl *template.Template, code string) error {
	buf := bytes.Buffer{}
	data := map[string]any{"threshold": format(th)}
	_ = tpl.Execute(&buf, data)
	return xrr.New(buf.String(), code)
}

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (r ThresholdRule) When(condition bool) ThresholdRule {
	r.condition = condition
	return r
}

// Code sets the error code for the rule.
func (r ThresholdRule) Code(code string) ThresholdRule {
	r.code = code
	r.err = setCode(r.err, code)
	return r
}

// Error sets custom error for the rule.
func (r ThresholdRule) Error(err error) ThresholdRule {
	r.err = err
	return r
}

// thresholdOutcome returns true if the result of the comparison for given the
// operator valid, false otherwise.
func thresholdOutcome(operator, result int) bool {
	switch operator {
	case greaterThan:
		return result == -1
	case greaterEqualThan:
		return result <= 0
	case lessThan:
		return result == 1
	case lessEqualThan:
		return result >= 0
	}
	return false
}

// compareInt matches [CompareFunc] signature and compares two signed integers.
func compareInt(want, have any) (int, error) {
	w, err := ToInt(want)
	if err != nil {
		return 0, err
	}
	h, err := ToInt(have)
	if err != nil {
		return 0, err
	}
	return cmp.Compare(w, h), nil
}

// compareUint matches [CompareFunc] signature and compares two unsigned
// integers.
func compareUint(want, have any) (int, error) {
	w, err := ToUint(want)
	if err != nil {
		return 0, err
	}
	h, err := ToUint(have)
	if err != nil {
		return 0, err
	}
	return cmp.Compare(w, h), nil
}

// compareFloat matches [CompareFunc] signature and compares two float numbers.
func compareFloat(want, have any) (int, error) {
	w, err := ToFloat(want)
	if err != nil {
		return 0, err
	}
	h, err := ToFloat(have)
	if err != nil {
		return 0, err
	}
	return cmp.Compare(w, h), nil
}

// compareTime matches [CompareFunc] signature and compares two [time.Time]
// instances.
func compareTime(want, have any) (int, error) {
	w, ok := want.(time.Time)
	if !ok {
		msg := fmt.Sprintf("cannot convert %T to time.Time", want)
		return 0, xrr.New(msg, ECInvType)
	}
	h, ok := have.(time.Time)
	if !ok {
		msg := fmt.Sprintf("cannot convert %T to time.Time", have)
		return 0, xrr.New(msg, ECInvType)
	}
	return cmp.Compare(w.UnixNano(), h.UnixNano()), nil
}

// compareFor returns a [CompareFunc] function for the given value. The
// function will return an error if the type is not supported.
func compareFor(val any) CompareFunc {
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return compareInt

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		return compareUint

	case reflect.Float32, reflect.Float64:
		return compareFloat

	default:
		if _, ok := val.(time.Time); ok {
			return compareTime
		}
		return nil
	}
}
