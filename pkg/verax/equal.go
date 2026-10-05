// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
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
const (
	// msgEqual is the error message template for not matching values.
	msgEqual = "must be equal to '{{.value}}'"

	// msgNotEqual is the error message template for matching values.
	msgNotEqual = "must not be equal to '{{.value}}'"
)

// Parsed message templates.
var (
	// tplEqual is parsed msgEqual template.
	tplEqual = mustTpl(EqualRuleName, msgEqual)

	// tplNotEqual is parsed msgNotEqual template.
	tplNotEqual = mustTpl(EqualRuleName, msgNotEqual)
)

// Equal constructs a rule conditioning a validated value is equal to "want".
// Values are compared using [reflect.DeepEqual]. An empty value is considered
// valid. Use the [Required] rule to make sure a value is not empty.
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

// NotEqual constructs a rule conditioning a validated value is not equal to
// "want". Values are compared using [reflect.DeepEqual]. An empty value is
// considered valid. Use the [Required] rule to make sure a value is not empty.
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
	msg := fmt.Sprintf("must be equal to '%s'", tplText(field))
	return Equal(want).Message(msg)
}

// NotEqualField constructs rule conditioning a validated value is not equal to
// "want". When it is the error message will say the value must not be equal to
// "field".
func NotEqualField(want any, field string) EqualRule {
	msg := fmt.Sprintf("must not be equal to '%s'", tplText(field))
	return NotEqual(want).Message(msg)
}

// errNotEqual signals inequality from the built-in equal function. It is
// a package-level singleton to avoid per-call allocation; EqualRule.Validate
// discards it (flgCustomFn == 0 branch) and returns its own pre-rendered error.
var errNotEqual = NewError("equal error", ECNotEqual)

// errEqual signals equality from the built-in notEqual function. It is
// a package-level singleton to avoid per-call allocation; EqualRule.Validate
// discards it (flgCustomFn == 0 branch) and returns its own pre-rendered error.
var errEqual = NewError("not equal error", ECEqual)

// equal matches [EqualFunc] signature and conditions both arguments are equal
// using [reflect.DeepEqual] function.
func equal(want, have any) error {
	if reflect.DeepEqual(want, have) {
		return nil
	}
	return errNotEqual
}

// notEqual matches [EqualFunc] signature and conditions both arguments are not
// equal using [reflect.DeepEqual] function.
func notEqual(want, have any) error {
	if !reflect.DeepEqual(want, have) {
		return nil
	}
	return errEqual
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
	msg       string    // Validation error message rendered from template.
	code      string    // Validation error code.
	sticky    error     // Sticky error.
	flags     uint8     // Customizations.
}

func (eql EqualRule) Validate(have any) error {
	if eql.sticky != nil {
		return eql.sticky
	}
	if !eql.condition {
		return nil
	}
	if IsEmpty(have) {
		return nil
	}
	if err := eql.fn(eql.want, have); err != nil {
		customMsg := eql.flags&flgCustomMsg != 0
		customCode := eql.flags&flgCustomCode != 0

		// If no custom function is set, uses the current message and code.
		// Also applies when both the custom message and code are provided.
		if eql.flags&flgCustomFn == 0 || (customMsg && customCode) {
			return NewError(eql.msg, eql.code)
		}

		if customMsg {
			return NewError(eql.msg, xrr.GetCode(err))
		}
		if customCode {
			return xrr.SetCode[edError](err, eql.code)
		}
		return err
	}
	return nil
}

func (eql EqualRule) When(condition bool) EqualRule {
	eql.condition = condition
	return eql
}

// With sets a custom equality function for an [EqualRule], overriding the
// default comparison behavior. The function, of type [EqualFunc], defines
// how the value is compared. This is useful for types not supported natively.
func (eql EqualRule) With(fn EqualFunc) EqualRule {
	if eql.sticky != nil {
		return eql
	}
	switch eql.mode {
	case "equal", "equal-by":
		eql.mode = "equal-by"

	case "not-equal", "not-equal-by":
		eql.mode = "not-equal-by"

	default:
		eql.sticky = NewInternalErrorf(
			"%s: invalid rule mode: %q",
			EqualRuleName,
			eql.mode,
			xrr.WithCode(ECInvRuleMode),
		)
		return eql
	}
	eql.fn = fn
	eql.flags |= flgCustomFn
	return eql
}

// Message sets a custom error message. The message is a text/template
// rendered with the wanted value as {{.value}}. An empty template is ignored.
// A template that fails to parse or render sets a sticky [InternalError]
// returned by [EqualRule.Validate] and [EqualRule.Spec].
func (eql EqualRule) Message(tpl string) EqualRule {
	if eql.sticky != nil || tpl == "" {
		return eql
	}
	parsed, err := template.New(EqualRuleName).
		Option("missingkey=error").
		Parse(tpl)
	if err != nil {
		eql.sticky = NewInternalErrorf(
			"%s(%s): custom template parse error",
			EqualRuleName,
			eql.mode,
			xrr.WithCode(ECInternal),
			xrr.WithCause(err),
		)
		return eql
	}

	buf := &bytes.Buffer{}
	data := map[string]any{spec.ArgValue: formatValue(eql.want)}
	if err = parsed.Execute(buf, data); err != nil {
		eql.sticky = NewInternalErrorf(
			"%s(%s): custom template render error",
			EqualRuleName,
			eql.mode,
			xrr.WithCode(ECInternal),
			xrr.WithCause(err),
		)
		return eql
	}

	eql.tpl = tpl
	eql.msg = buf.String()
	eql.flags |= flgCustomMsg
	return eql
}

func (eql EqualRule) Code(code string) EqualRule {
	if code != "" {
		eql.code = code
		eql.flags |= flgCustomCode
	}
	return eql
}

func (eql EqualRule) Spec() (*spec.Spec, error) {
	return eql.spec(EqualRuleName)
}
func (eql EqualRule) spec(name string) (*spec.Spec, error) {
	if eql.sticky != nil {
		return nil, eql.sticky
	}

	spc := spec.NewSpec(name)
	switch eql.mode {
	case "equal", "not-equal", "equal-by", "not-equal-by":
		// Valid mode.

	default:
		return nil, NewInternalErrorf(
			"%s: invalid rule mode: %q",
			name,
			eql.mode,
			xrr.WithCode(ECInvRuleMode),
		)
	}
	spc.SetArg(ArgMode, eql.mode)
	spc.SetArg(spec.ArgValue, eql.want)

	if eql.flags&flgCustomMsg != 0 {
		spc.SetArg(ArgErrMsg, eql.tpl)
	}
	if eql.flags&flgCustomCode != 0 {
		spc.SetArg(ArgErrCode, eql.code)
	}
	if eql.flags&flgCustomFn != 0 {
		spc.SetArg(spec.ArgSrc, eql.fn)
	}

	return spc, nil
}

// EqualRuleFromSpec creates a new instance of [EqualRule] from the [spec.Spec].
func EqualRuleFromSpec(spc *spec.Spec) (EqualRule, error) {
	return equalRuleFromSpec(spc, EqualRuleName)
}
func equalRuleFromSpec(spc *spec.Spec, name string) (EqualRule, error) {
	if spc.Name != name {
		return EqualRule{}, NewInternalErrorf(
			"%s: invalid spec name: %q",
			name,
			spc.Name,
			xrr.WithCode(spec.ECInvSpec),
		)
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
		return EqualRule{}, NewInternalErrorf(
			"%s: invalid spec rule mode: %q",
			name,
			mode,
			xrr.WithCode(spec.ECInvSpec),
		)
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
