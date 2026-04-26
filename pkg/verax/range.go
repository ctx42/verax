// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"text/template"
	"time"

	"github.com/ctx42/convert/pkg/convert"
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
var (
	msgGreaterOrEqual = "must be greater or equal to {{.value}}"
	tplGreaterOrEqual = mustTpl(RangeRuleName, msgGreaterOrEqual)

	msgGreaterThan = "must be greater than {{.value}}"
	tplGreaterThan = mustTpl(RangeRuleName, msgGreaterThan)

	msgLessOrEqual = "must be less or equal to {{.value}}"
	tplLessOrEqual = mustTpl(RangeRuleName, msgLessOrEqual)

	msgLessThan = "must be less than {{.value}}"
	tplLessThan = mustTpl(RangeRuleName, msgLessThan)
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
// It must be called before [RangeRule.Code] or [RangeRule.Message].
//
// Example:
//
//	ruleMin := Min(10).Exclusive()  // Value must be > 10
//	ruleMax := Max(100).Exclusive() // Value must be < 100
func (r RangeRule) Exclusive() RangeRule {
	if r.sticky != nil {
		return r
	}
	switch r.mode {
	case "max":
		r.mode = "max-exclusive"
		r.tpl = msgLessThan
		prefix := fmt.Sprintf("%s(max-exclusive)", RangeRuleName)
		r.msg, r.sticky = renderTpl(tplLessThan, r.threshold, prefix)
	case "min":
		r.mode = "min-exclusive"
		r.tpl = msgGreaterThan
		prefix := fmt.Sprintf("%s(min-exclusive)", RangeRuleName)
		r.msg, r.sticky = renderTpl(tplGreaterThan, r.threshold, prefix)
	default:
		// NOOP: already exclusive.
	}
	return r
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
func (r RangeRule) With(fn CompareFunc) RangeRule {
	// Allow overriding when sticky is ECInvType: the custom function may
	// support a type that the built-in comparison does not.
	if r.sticky != nil && xrr.GetCode(r.sticky) != ECInvType {
		return r
	}
	r.fn = fn
	r.flags |= flgCustomFn
	return r
}

func (r RangeRule) Validate(have any) error {
	if r.sticky != nil {
		return r.sticky
	}
	if !r.condition {
		return nil
	}
	if IsEmpty(have) {
		return nil
	}

	res, err := r.fn(r.threshold, have)
	if err != nil {
		var e *InternalError
		if errors.As(err, &e) {
			return e
		}
		customMsg := r.flags&flgCustomMsg != 0
		customCode := r.flags&flgCustomCode != 0

		// If no custom function is set, uses the current message and code.
		// Also applies when both the custom message and code are provided.
		if r.flags&flgCustomFn == 0 || (customMsg && customCode) {
			return NewError(r.msg, r.code)
		}

		if customMsg {
			return NewError(r.msg, xrr.GetCode(err))
		}
		if customCode {
			return xrr.SetCode[edError](err, r.code)
		}
		return err
	}

	if !rangeOutcome(r.mode, res) {
		return NewError(r.msg, r.code)
	}
	return nil
}

func (r RangeRule) When(condition bool) RangeRule {
	r.condition = condition
	return r
}

func (r RangeRule) Message(tpl string) RangeRule {
	if r.sticky != nil || tpl == "" {
		return r
	}
	var parsed *template.Template
	parsed, r.sticky = template.New(RangeRuleName).
		Option("missingkey=error").
		Parse(tpl)
	if r.sticky != nil {
		format := "%s(%s): custom template parse error"
		msg := fmt.Sprintf(format, RangeRuleName, r.mode)
		r.sticky = NewInternalError(msg, ECInternal)
		return r
	}

	buf := &bytes.Buffer{}
	r.sticky = parsed.Execute(buf, map[string]any{spec.ArgValue: r.threshold})
	if r.sticky != nil {
		format := "%s(%s): custom template render error"
		msg := fmt.Sprintf(format, RangeRuleName, r.mode)
		r.sticky = NewInternalError(msg, ECInternal)
		return r
	}
	r.tpl = tpl
	r.msg = buf.String()
	r.flags |= flgCustomMsg
	return r
}

func (r RangeRule) Code(code string) RangeRule {
	if code != "" {
		r.code = code
		r.flags |= flgCustomCode
	}
	return r
}

func (r RangeRule) Spec() (*spec.Spec, error) {
	if r.sticky != nil {
		return nil, r.sticky
	}

	spc := spec.NewSpec(RangeRuleName)
	switch r.mode {
	case "min":
		spc.SetArg(ArgMode, "min")
	case "min-exclusive":
		spc.SetArg(ArgMode, "min-exclusive")
	case "max":
		spc.SetArg(ArgMode, "max")
	case "max-exclusive":
		spc.SetArg(ArgMode, "max-exclusive")
	default:
		format := "%s: invalid rule mode: %q"
		msg := fmt.Sprintf(format, RangeRuleName, r.mode)
		return nil, NewInternalError(msg, ECInvRuleMode)
	}
	spc.SetArg(spec.ArgValue, r.threshold)

	if r.flags&flgCustomMsg != 0 {
		spc.SetArg(ArgErrMsg, r.tpl)
	}
	if r.flags&flgCustomCode != 0 {
		spc.SetArg(ArgErrCode, r.code)
	}

	if r.flags&flgCustomFn != 0 {
		spc.SetArg(spec.ArgSrc, r.fn)
	}

	return spc, nil
}

// RangeRuleFromSpec creates a new instance of [RangeRule] from the [spec.Spec].
func RangeRuleFromSpec(spc *spec.Spec) (RangeRule, error) {
	if spc.Name != RangeRuleName {
		format := "%s: invalid spec name: %q"
		msg := fmt.Sprintf(format, RangeRuleName, spc.Name)
		return RangeRule{}, NewInternalError(msg, spec.ECInvSpec)
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
		format := "%s: invalid spec rule mode: %q"
		msg := fmt.Sprintf(format, RangeRuleName, mode)
		return RangeRule{}, NewInternalError(msg, spec.ECInvSpec)
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
		return result == -1
	case "max":
		return result >= 0
	case "max-exclusive":
		return result == 1
	}
	return false
}

// compareInt matches [CompareFunc] signature and compares two signed integers.
func compareInt(want, have any) (int, error) {
	w, err := convert.AnyToInt64(want)
	if err != nil {
		return 0, errConvert(RangeRuleName, want, int64(0))
	}
	h, err := convert.AnyToInt64(have)
	if err != nil {
		return 0, errConvert(RangeRuleName, have, int64(0))
	}
	return cmp.Compare(w, h), nil
}

// compareUInt matches [CompareFunc] signature and compares two unsigned
// integers.
func compareUInt(want, have any) (int, error) {
	w, err := convert.AnyToUint64(want)
	if err != nil {
		return 0, errConvert(RangeRuleName, want, uint64(0))
	}
	h, err := convert.AnyToUint64(have)
	if err != nil {
		return 0, errConvert(RangeRuleName, have, uint64(0))
	}
	return cmp.Compare(w, h), nil
}

// compareFloat matches [CompareFunc] signature and compares two float numbers.
func compareFloat(want, have any) (int, error) {
	w, err := convert.AnyToFloat64(want)
	if err != nil {
		return 0, errConvert(RangeRuleName, want, 0.0)
	}
	h, err := convert.AnyToFloat64(have)
	if err != nil {
		return 0, errConvert(RangeRuleName, have, 0.0)
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
	return cmp.Compare(w.UnixNano(), h.UnixNano()), nil
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
		format := "%s: unsupported type comparison: %T"
		msg := fmt.Sprintf(format, RangeRuleName, val)
		return nil, NewInternalError(msg, ECInvType)
	}
}
