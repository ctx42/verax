// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"bytes"
	"fmt"
	"reflect"
	"text/template"

	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/spec"
)

// EqualRuleName represents [EqualRule] name.
const EqualRuleName = "equal-rule"

// [EqualRule] error codes.
const (
	// ECNotEqual indicates that a value does not match the expected one.
	ECNotEqual = "ECNotEqual"

	// ECEqual indicates that a value unexpectedly matches the expected one.
	ECEqual = "ECEqual"
)

// [EqualRule] rule error messages.
var (
	// msgEqual is the error message template for not matching values.
	msgEqual = "must be equal to '{{.value}}'"

	// tplEqual is parsed msgEqual template.
	tplEqual = mustTpl(EqualRuleName, msgEqual)

	// msgNotEqual is the error message template for matching values.
	msgNotEqual = "must not be equal to '{{.value}}'"

	// tplNotEqual is parsed msgNotEqual template.
	tplNotEqual = mustTpl(EqualRuleName, msgNotEqual)
)

// Equal constructs rule conditioning a validated value is equal to "want".
func Equal(want any) EqualRule {
	r := EqualRule{
		mode:      "equal",
		want:      want,
		condition: true,
		fn:        equal,
		tpl:       msgEqual,
		code:      ECNotEqual,
	}
	prefix := EqualRuleName + "(equal)"
	r.msg, r.sticky = renderTpl(tplEqual, r.want, prefix)
	return r
}

// NotEqual constructs rule conditioning a validated value is not equal to "want".
func NotEqual(want any) EqualRule {
	r := EqualRule{
		mode:      "not-equal",
		want:      want,
		condition: true,
		fn:        notEqual,
		tpl:       msgNotEqual,
		code:      ECEqual,
	}
	prefix := EqualRuleName + "(not-equal)"
	r.msg, r.sticky = renderTpl(tplNotEqual, r.want, prefix)
	return r
}

// EqualField constructs rule conditioning a validated value is equal to "want".
// When it isn't, the error message will say the value must be equal to "field".
func EqualField(want any, field string) EqualRule {
	msg := fmt.Sprintf("must be equal to '%s'", field)
	return Equal(want).Message(msg)
}

// NotEqualField constructs rule conditioning a validated value is not equal to
// "want". When it is the error message will say the value must not be equal to
// "field".
func NotEqualField(want any, field string) EqualRule {
	msg := fmt.Sprintf("must not be equal to '%s'", field)
	return NotEqual(want).Message(msg)
}

// equal matches [EqualFunc] signature and conditions both arguments are equal
// using [reflect.DeepEqual] function.
func equal(want, have any) error {
	if reflect.DeepEqual(want, have) {
		return nil
	}
	// This error is ignored inside [EqualRule.Validate] because it is not
	// treated as coming from a custom function.
	return NewError("equal error", ECNotEqual)
}

// notEqual matches [EqualFunc] signature and conditions both arguments are not
// equal using [reflect.DeepEqual] function.
func notEqual(want, have any) error {
	if !reflect.DeepEqual(want, have) {
		return nil
	}
	// This error is ignored inside [EqualRule.Validate] because it is not
	// treated as coming from a custom function.
	return NewError("not equal error", ECEqual)
}

// Compile time conditions.
var (
	_ customizer[EqualRule]  = EqualRule{}
	_ conditioner[EqualRule] = EqualRule{}
	_ Rule                   = EqualRule{}
)

// EqualRule is a rule that conditions a value matches the expected value using
// the provided comparison function.
type EqualRule struct {
	mode      string    // Mode of operation.
	want      any       // Wanted value.
	condition bool      // Run validation only when true.
	fn        EqualFunc // Equality testing function.
	tpl       string    // The original error message template.
	msg       string    // The original error template string.
	code      string    // Validation error code.
	sticky    error     // Sticky error.
	flags     uint8     // Customizations.
}

func (r EqualRule) Validate(have any) error {
	if r.sticky != nil {
		return r.sticky
	}
	if !r.condition {
		return nil
	}
	if IsEmpty(have) {
		return nil
	}
	if err := r.fn(r.want, have); err != nil {
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
	return nil
}

func (r EqualRule) When(condition bool) EqualRule {
	r.condition = condition
	return r
}

// With sets a custom equality function for an [EqualRule], overriding the
// default comparison behavior. The function, of type [EqualFunc], defines
// how the value is compared. This is useful for types not supported natively.
func (r EqualRule) With(fn EqualFunc) EqualRule {
	if r.sticky != nil {
		return r
	}
	switch r.mode {
	case "equal", "equal-by":
		r.mode = "equal-by"
	case "not-equal", "not-equal-by":
		r.mode = "not-equal-by"
	default:
		msg := fmt.Sprintf("%s: invalid rule mode: %q", EqualRuleName, r.mode)
		r.sticky = NewInternalError(msg, ECInvRuleMode)
		return r
	}
	r.fn = fn
	r.flags |= flgCustomFn
	return r
}

func (r EqualRule) Message(tpl string) EqualRule {
	if r.sticky != nil || tpl == "" {
		return r
	}
	var parsed *template.Template
	parsed, r.sticky = template.New(EqualRuleName).
		Option("missingkey=error").
		Parse(tpl)
	if r.sticky != nil {
		format := "%s(%s): custom template parse error"
		msg := fmt.Sprintf(format, EqualRuleName, r.mode)
		r.sticky = NewInternalError(msg, ECInternal)
		return r
	}

	buf := &bytes.Buffer{}
	r.sticky = parsed.Execute(buf, map[string]any{spec.ArgValue: r.want})
	if r.sticky != nil {
		format := "%s(%s): custom template render error"
		msg := fmt.Sprintf(format, EqualRuleName, r.mode)
		r.sticky = NewInternalError(msg, ECInternal)
		return r
	}

	r.tpl = tpl
	r.msg = buf.String()
	r.flags |= flgCustomMsg
	return r
}

func (r EqualRule) Code(code string) EqualRule {
	if code != "" {
		r.code = code
		r.flags |= flgCustomCode
	}
	return r
}

func (r EqualRule) Spec() (*spec.Spec, error) {
	return r.spec(EqualRuleName)
}
func (r EqualRule) spec(name string) (*spec.Spec, error) {
	if r.sticky != nil {
		return nil, r.sticky
	}

	spc := spec.NewSpec(name)
	switch r.mode {
	case "equal", "not-equal":
		// Nothing to do.
	case "equal-by", "not-equal-by":
		spc.SetArg(spec.ArgSrc, r.fn)
	default:
		msg := fmt.Sprintf("%s: invalid rule mode: %q", name, r.mode)
		return nil, NewInternalError(msg, ECInvRuleMode)
	}
	spc.SetArg(ArgMode, r.mode)
	spc.SetArg(spec.ArgValue, r.want)

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

// EqualRuleFromSpec creates a new instance of [EqualRule] from the [spec.Spec].
func EqualRuleFromSpec(spc *spec.Spec) (EqualRule, error) {
	return equalRuleFromSpec(spc, EqualRuleName)
}
func equalRuleFromSpec(spc *spec.Spec, name string) (EqualRule, error) {
	if spc.Name != name {
		msg := fmt.Sprintf("%s: invalid spec name: %q", name, spc.Name)
		return EqualRule{}, NewInternalError(msg, spec.ECInvSpec)
	}
	mode, err := getArg[string](spc.Args, ArgMode, name)
	if err != nil {
		return EqualRule{}, err
	}
	val, err := getArg[any](spc.Args, spec.ArgValue, name)
	if err != nil {
		return EqualRule{}, err
	}

	var rule EqualRule
	switch mode {
	case "equal":
		rule = Equal(val)
	case "not-equal":
		rule = NotEqual(val)
	case "equal-by":
		fn, err := getArg[EqualFunc](spc.Args, spec.ArgSrc, name)
		if err != nil {
			return EqualRule{}, err
		}
		rule = Equal(val).With(fn)
	case "not-equal-by":
		fn, err := getArg[EqualFunc](spc.Args, spec.ArgSrc, name)
		if err != nil {
			return EqualRule{}, err
		}
		rule = NotEqual(val).With(fn)
	default:
		format := "%s: invalid spec rule mode: %q"
		msg := fmt.Sprintf(format, name, mode)
		return EqualRule{}, NewInternalError(msg, spec.ECInvSpec)
	}

	if spc.ArgExist(ArgErrMsg) {
		msg, err := getArg[string](spc.Args, ArgErrMsg, name)
		if err != nil {
			return EqualRule{}, err
		}
		if msg != "" {
			rule = rule.Message(msg)
		}
	}

	if spc.ArgExist(ArgErrCode) {
		code, err := getArg[string](spc.Args, ArgErrCode, name)
		if err != nil {
			return EqualRule{}, err
		}
		if code != "" {
			rule = rule.Code(code)
		}
	}

	return rule, nil
}
