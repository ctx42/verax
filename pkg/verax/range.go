// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"math"
	"text/template"
	"time"

	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
)

// RangeRuleName represents [RangeRule] name.
const RangeRuleName = "range-rule"

// [RangeRule] rule error codes.
const (
	// ECInvRange is the error code for values failing range conditions.
	ECInvRange = "ECInvRange"
)

// Error message templates for [RangeRule].
const (
	msgGreaterOrEqual = "must be greater or equal to {{.value}}"
	msgGreaterThan    = "must be greater than {{.value}}"
	msgLessOrEqual    = "must be less or equal to {{.value}}"
	msgLessThan       = "must be less than {{.value}}"
)

// Parsed error message templates for [RangeRule].
var (
	tplGreaterOrEqual = mustTpl(RangeRuleName, msgGreaterOrEqual)
	tplGreaterThan    = mustTpl(RangeRuleName, msgGreaterThan)
	tplLessOrEqual    = mustTpl(RangeRuleName, msgLessOrEqual)
	tplLessThan       = mustTpl(RangeRuleName, msgLessThan)
)

// Min creates a validation rule that conditions if a value is greater than or
// equal to the specified range. Use [RangeRule.Exclusive] to enforce a strict
// greater-than condition. The value being conditioned and the range must be of
// the same type, supporting only int, uint, float, and time.Time types. Empty
// values are considered valid; use the [Required] rule to ensure a value is
// not empty.
//
// Example:
//
//	rule := Min(10)             // Value must be >= 10
//	rule := Min(10).Exclusive() // Value must be > 10
func Min(minimum any) RangeRule {
	r := RangeRule{
		mode:      "min",
		threshold: minimum,
		condition: true,
		tpl:       msgGreaterOrEqual,
		code:      ECInvRange,
	}
	if r.fn, r.sticky = compareFor(minimum); r.sticky != nil {
		return r
	}
	prefix := fmt.Sprintf("%s(min)", RangeRuleName)
	r.msg, r.sticky = renderTpl(tplGreaterOrEqual, r.threshold, prefix)
	return r
}

// Max creates a validation rule that conditions if a value is less than or
// equal to the specified range. Use [RangeRule.Exclusive] to enforce a strict
// less-than condition. The value being conditioned and the range must be of
// the same type, supporting only int, uint, float, and time.Time types. Empty
// values are considered valid; use the [Required] rule to ensure a value is
// not empty.
//
// Example:
//
//	rule := Max(100)             // Value must be <= 100
//	rule := Max(100).Exclusive() // Value must be < 100
func Max(maximum any) RangeRule {
	r := RangeRule{
		mode:      "max",
		threshold: maximum,
		condition: true,
		tpl:       msgLessOrEqual,
		code:      ECInvRange,
	}
	if r.fn, r.sticky = compareFor(maximum); r.sticky != nil {
		return r
	}
	prefix := fmt.Sprintf("%s(max)", RangeRuleName)
	r.msg, r.sticky = renderTpl(tplLessOrEqual, r.threshold, prefix)
	return r
}

// Compile time conditions.
var (
	_ customizer[RangeRule]  = RangeRule{}
	_ conditioner[RangeRule] = RangeRule{}
	_ Rule                   = RangeRule{}
)

// RangeRule is a rule validating a value satisfies a given range.
type RangeRule struct {
	mode      string      // Mode of operation.
	threshold any         // The range value.
	fn        CompareFunc // Comparison function.
	condition bool        // Run validation only when true.
	tpl       string      // The original error message template.
	msg       string      // Validation error message rendered from template.
	code      string      // Validation error code.
	sticky    error       // Sticky error.
	flags     uint8       // Customizations.
}

// Exclusive modifies a [RangeRule] to exclude the boundary value, enforcing a
// strict comparison. For example, when used with [Min], it conditions if a
// value is strictly greater than the range, and with [Max], it conditions if a
// value is strictly less than the range. The value and range must be of the
// same type.
//
// A custom message set with [RangeRule.Message] is kept.
//
// Example:
//
//	ruleMin := Min(10).Exclusive()  // Value must be > 10
//	ruleMax := Max(100).Exclusive() // Value must be < 100
func (rng RangeRule) Exclusive() RangeRule {
	if rng.sticky != nil {
		return rng
	}
	var msg string
	var tpl *template.Template
	switch rng.mode {
	case "max":
		rng.mode = "max-exclusive"
		msg, tpl = msgLessThan, tplLessThan

	case "min":
		rng.mode = "min-exclusive"
		msg, tpl = msgGreaterThan, tplGreaterThan

	default:
		return rng // Already exclusive.
	}
	if rng.flags&flgCustomMsg != 0 {
		return rng
	}
	rng.tpl = msg
	prefix := fmt.Sprintf("%s(%s)", RangeRuleName, rng.mode)
	rng.msg, rng.sticky = renderTpl(tpl, rng.threshold, prefix)
	return rng
}

// With sets a custom comparison function for a [RangeRule], overriding the
// default comparison behavior. The function, of type [CompareFunc], defines
// how the value is compared to the range. This is useful for custom validation
// logic.
//
// The function must return an error if the type is not supported.
//
// Example:
//
//	cmpMyType := func(a, b any) int { ... }
//	rule := Min(myTypeValue).With(cmpMyType)
func (rng RangeRule) With(fn CompareFunc) RangeRule {
	if rng.sticky != nil {
		// Allow overriding when sticky is ECInvType: the custom function may
		// support a type that the built-in comparison does not.
		if xrr.GetCode(rng.sticky) != ECInvType {
			return rng
		}
		tpl := tplGreaterOrEqual
		if rng.mode == "max" {
			tpl = tplLessOrEqual
		}
		prefix := fmt.Sprintf("%s(%s)", RangeRuleName, rng.mode)
		rng.msg, rng.sticky = renderTpl(tpl, rng.threshold, prefix)
		if rng.sticky != nil {
			return rng
		}
	}
	rng.fn = fn
	rng.flags |= flgCustomFn
	return rng
}

func (rng RangeRule) Validate(have any) error {
	if rng.sticky != nil {
		return rng.sticky
	}
	if !rng.condition {
		return nil
	}
	if IsEmpty(have) {
		return nil
	}

	res, err := rng.fn(rng.threshold, have)
	if err != nil {
		if e, ok := errors.AsType[*InternalError](err); ok {
			return e
		}
		customMsg := rng.flags&flgCustomMsg != 0
		customCode := rng.flags&flgCustomCode != 0

		// If no custom function is set, uses the current message and code.
		// Also applies when both the custom message and code are provided.
		if rng.flags&flgCustomFn == 0 || (customMsg && customCode) {
			return NewError(rng.msg, rng.code)
		}

		if customMsg {
			return NewError(rng.msg, xrr.GetCode(err))
		}
		if customCode {
			return xrr.SetCode[edError](err, rng.code)
		}
		return err
	}

	if !rangeOutcome(rng.mode, res) {
		return NewError(rng.msg, rng.code)
	}
	return nil
}

func (rng RangeRule) When(condition bool) RangeRule {
	rng.condition = condition
	return rng
}

func (rng RangeRule) Message(tpl string) RangeRule {
	if rng.sticky != nil || tpl == "" {
		return rng
	}
	parsed, err := template.New(RangeRuleName).
		Option("missingkey=error").
		Parse(tpl)
	if err != nil {
		rng.sticky = NewInternalErrorf(
			"%s(%s): custom template parse error",
			RangeRuleName,
			rng.mode,
			xrr.WithCode(ECInternal),
			xrr.WithCause(err),
		)
		return rng
	}

	buf := &bytes.Buffer{}
	data := map[string]any{spec.ArgValue: formatValue(rng.threshold)}
	if err = parsed.Execute(buf, data); err != nil {
		rng.sticky = NewInternalErrorf(
			"%s(%s): custom template render error",
			RangeRuleName,
			rng.mode,
			xrr.WithCode(ECInternal),
			xrr.WithCause(err),
		)
		return rng
	}
	rng.tpl = tpl
	rng.msg = buf.String()
	rng.flags |= flgCustomMsg
	return rng
}

func (rng RangeRule) Code(code string) RangeRule {
	if code != "" {
		rng.code = code
		rng.flags |= flgCustomCode
	}
	return rng
}

func (rng RangeRule) Spec() (*spec.Spec, error) {
	if rng.sticky != nil {
		return nil, rng.sticky
	}

	spc := spec.NewSpec(RangeRuleName)
	switch rng.mode {
	case "min":
		spc.SetArg(ArgMode, "min")

	case "min-exclusive":
		spc.SetArg(ArgMode, "min-exclusive")

	case "max":
		spc.SetArg(ArgMode, "max")

	case "max-exclusive":
		spc.SetArg(ArgMode, "max-exclusive")

	default:
		return nil, NewInternalErrorf(
			"%s: invalid rule mode: %q",
			RangeRuleName,
			rng.mode,
			xrr.WithCode(ECInvRuleMode),
		)
	}
	spc.SetArg(spec.ArgValue, rng.threshold)

	if rng.flags&flgCustomMsg != 0 {
		spc.SetArg(ArgErrMsg, rng.tpl)
	}
	if rng.flags&flgCustomCode != 0 {
		spc.SetArg(ArgErrCode, rng.code)
	}

	if rng.flags&flgCustomFn != 0 {
		spc.SetArg(spec.ArgSrc, rng.fn)
	}

	return spc, nil
}

// RangeRuleFromSpec creates a new instance of [RangeRule] from the [spec.Spec].
func RangeRuleFromSpec(spc *spec.Spec) (RangeRule, error) {
	if spc.Name != RangeRuleName {
		return RangeRule{}, NewInternalErrorf(
			"%s: invalid spec name: %q",
			RangeRuleName,
			spc.Name,
			xrr.WithCode(spec.ECInvSpec),
		)
	}
	mode, err := getArg[string](spc.Args, ArgMode, RangeRuleName)
	if err != nil {
		return RangeRule{}, err
	}
	val, err := getArg[any](spc.Args, spec.ArgValue, RangeRuleName)
	if err != nil {
		return RangeRule{}, err
	}

	var rule RangeRule
	switch mode {
	case "min":
		rule = Min(val)

	case "min-exclusive":
		rule = Min(val).Exclusive()

	case "max":
		rule = Max(val)

	case "max-exclusive":
		rule = Max(val).Exclusive()

	default:
		return RangeRule{}, NewInternalErrorf(
			"%s: invalid spec rule mode: %q",
			RangeRuleName,
			mode,
			xrr.WithCode(spec.ECInvSpec),
		)
	}

	if rule.sticky != nil {
		return RangeRule{}, rule.sticky
	}

	if spc.ArgExist(ArgErrMsg) {
		msg, err := getArg[string](spc.Args, ArgErrMsg, RangeRuleName)
		if err != nil {
			return RangeRule{}, err
		}
		if msg != "" {
			rule = rule.Message(msg)
		}
	}

	if spc.ArgExist(ArgErrCode) {
		code, err := getArg[string](spc.Args, ArgErrCode, RangeRuleName)
		if err != nil {
			return RangeRule{}, err
		}
		if code != "" {
			rule = rule.Code(code)
		}
	}

	if spc.ArgExist(spec.ArgSrc) {
		fn, err := getArg[CompareFunc](spc.Args, spec.ArgSrc, RangeRuleName)
		if err != nil {
			return RangeRule{}, err
		}
		rule = rule.With(fn)
	}

	return rule, nil
}

// rangeOutcome returns true if the result of the comparison for the given
// operator is valid, false otherwise.
func rangeOutcome(mode string, result int) bool {
	switch mode {
	case "min":
		return result <= 0
	case "min-exclusive":
		return result < 0
	case "max":
		return result >= 0
	case "max-exclusive":
		return result > 0
	}
	return false
}

// compareInt matches [CompareFunc] signature and compares two signed integers.
func compareInt(want, have any) (int, error) {
	var w, h int64
	switch v := want.(type) {
	case int:
		w = int64(v)
	case int8:
		w = int64(v)
	case int16:
		w = int64(v)
	case int32:
		w = int64(v)
	case int64:
		w = v
	case time.Duration:
		w = int64(v)
	default:
		return 0, errConvert(RangeRuleName, want, int64(0))
	}
	switch v := have.(type) {
	case int:
		h = int64(v)
	case int8:
		h = int64(v)
	case int16:
		h = int64(v)
	case int32:
		h = int64(v)
	case int64:
		h = v
	case time.Duration:
		h = int64(v)
	default:
		return 0, errConvert(RangeRuleName, have, int64(0))
	}
	return cmp.Compare(w, h), nil
}

// compareUInt matches [CompareFunc] signature and compares two unsigned
// integers.
func compareUInt(want, have any) (int, error) {
	var w, h uint64
	switch v := want.(type) {
	case uint:
		w = uint64(v)
	case uint8:
		w = uint64(v)
	case uint16:
		w = uint64(v)
	case uint32:
		w = uint64(v)
	case uint64:
		w = v
	case uintptr:
		w = uint64(v)
	default:
		return 0, errConvert(RangeRuleName, want, uint64(0))
	}
	switch v := have.(type) {
	case uint:
		h = uint64(v)
	case uint8:
		h = uint64(v)
	case uint16:
		h = uint64(v)
	case uint32:
		h = uint64(v)
	case uint64:
		h = v
	case uintptr:
		h = uint64(v)
	default:
		return 0, errConvert(RangeRuleName, have, uint64(0))
	}
	return cmp.Compare(w, h), nil
}

// compareFloat matches [CompareFunc] signature and compares two float numbers.
func compareFloat(want, have any) (int, error) {
	var w, h float64
	switch v := want.(type) {
	case float32:
		w = float64(v)
	case float64:
		w = v
	default:
		return 0, errConvert(RangeRuleName, want, 0.0)
	}
	switch v := have.(type) {
	case float32:
		h = float64(v)
	case float64:
		h = v
	default:
		return 0, errConvert(RangeRuleName, have, 0.0)
	}
	if math.IsNaN(w) || math.IsNaN(h) {
		return 0, NewError("not a number", ECInvRange)
	}
	return cmp.Compare(w, h), nil
}

// compareTime matches [CompareFunc] signature and compares two [time.Time]
// instances.
func compareTime(want, have any) (int, error) {
	w, ok := want.(time.Time)
	if !ok {
		return 0, errConvert(RangeRuleName, want, time.Time{})
	}
	h, ok := have.(time.Time)
	if !ok {
		return 0, errConvert(RangeRuleName, have, time.Time{})
	}
	return w.Compare(h), nil
}

// compareFor returns a [CompareFunc] function for the given value. The
// function will return an error if the type is not supported.
func compareFor(val any) (CompareFunc, error) {
	switch val.(type) {
	case int, int8, int16, int32, int64, time.Duration:
		return compareInt, nil

	case uint, uint8, uint16, uint32, uint64, uintptr:
		return compareUInt, nil

	case float32, float64:
		return compareFloat, nil

	case time.Time:
		return compareTime, nil

	default:
		return nil, NewInternalErrorf(
			"%s: unsupported type comparison: %T",
			RangeRuleName,
			val,
			xrr.WithCode(ECInvType),
		)
	}
}
