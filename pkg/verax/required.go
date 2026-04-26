// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"fmt"

	"github.com/ctx42/verax/pkg/spec"
)

// RequiredRuleName represents [RequiredRule] name.
const RequiredRuleName = "required-rule"

// [RequiredRule] rule error codes.
const (
	// ECReq represents error code for missing required value.
	ECReq = "ECRequired"

	// ECReqNotEmpty represents error code for not nil but empty value.
	ECReqNotEmpty = "ECReqNotEmpty"

	// ECReqNotNil represents error code for nil value.
	ECReqNotNil = "ECReqNotNil"
)

// [RequiredRule] rule errors.
var (
	// msgMissing is the error message when a required value was not provided.
	// See [Required] for more details.
	msgMissing = "cannot be blank"

	// msgReqNotEmpty is the error message when a required value is empty.
	msgReqNotEmpty = "cannot be blank"

	// msgReqNotNil is the error message when a value has not been provided.
	msgReqNotNil = "is required"
)

// Rules based on [RequiredRule].
var (
	// Required is a validation rule that conditions if a value is not empty.
	//
	// A value is considered not empty if:
	//  - integer, float: not zero
	//  - bool: always not empty
	//  - string, array, slice, map: len() > 0
	//  - interface, pointer: not nil, and the referenced value is not empty
	//  - other: IsZero() returns false
	Required = RequiredRule{
		mode:      "required",
		condition: true,
		msg:       msgMissing,
		code:      ECReq,
	}

	// NotEmpty conditions if a value is either nil or not empty. It differs
	// from [Required] in that it treats a nil pointer or nil interface as valid.
	NotEmpty = RequiredRule{
		mode:      "not-empty",
		condition: true,
		msg:       msgReqNotEmpty,
		code:      ECReqNotEmpty,
	}

	// NotNil is a validation rule that conditions if a value is not nil. It
	// applies to interface, pointer, slice, and map types. All other types are
	// considered valid.
	NotNil = RequiredRule{
		mode:      "not-nil",
		condition: true,
		msg:       msgReqNotNil,
		code:      ECReqNotNil,
	}
)

// Compile time conditions.
var (
	_ customizer[RequiredRule]  = RequiredRule{}
	_ conditioner[RequiredRule] = RequiredRule{}
	_ Rule                      = RequiredRule{}
)

// RequiredRule is a rule that conditions if a value is not empty.
type RequiredRule struct {
	mode      string // Mode of operation.
	condition bool   // Run validation only when true.
	msg       string // Validation error message.
	code      string // Validation error code.
	flags     uint8  // Customizations.
}

func (r RequiredRule) Validate(have any) error {
	if !r.condition {
		return nil
	}
	isNil, isEmpty := checkNilAndEmpty(have)
	if isNil && r.mode == "not-empty" {
		return nil
	}
	if isNil && (r.mode == "required" || r.mode == "not-nil") {
		return NewError(r.msg, r.code)
	}
	if !isEmpty {
		return nil
	}
	if r.mode == "not-nil" {
		return nil
	}
	return NewError(r.msg, r.code)
}

func (r RequiredRule) When(condition bool) RequiredRule {
	r.condition = condition
	return r
}

func (r RequiredRule) Message(msg string) RequiredRule {
	if msg == "" {
		return r
	}
	r.msg = msg
	r.flags |= flgCustomMsg
	return r
}

func (r RequiredRule) Code(code string) RequiredRule {
	if code != "" {
		r.code = code
		r.flags |= flgCustomCode
	}
	return r
}

func (r RequiredRule) Spec() (*spec.Spec, error) {
	spc := spec.NewSpec(RequiredRuleName)
	if r.flags&flgCustomMsg != 0 {
		spc.SetArg(ArgErrMsg, r.msg)
	}
	if r.flags&flgCustomCode != 0 {
		spc.SetArg(ArgErrCode, r.code)
	}
	switch r.mode {
	case "required", "not-empty", "not-nil":
		spc.SetArg(ArgMode, r.mode)
	default:
		format := "%s: invalid rule mode: %q"
		msg := fmt.Sprintf(format, RequiredRuleName, r.mode)
		return nil, NewInternalError(msg, ECInvRuleMode)
	}
	return spc, nil
}

// RequiredRuleFromSpec creates a new instance of [RequiredRule] from the
// [spec.Spec].
func RequiredRuleFromSpec(spc *spec.Spec) (RequiredRule, error) {
	if spc.Name != RequiredRuleName {
		format := "%s: invalid spec name: %q"
		msg := fmt.Sprintf(format, RequiredRuleName, spc.Name)
		return RequiredRule{}, NewInternalError(msg, spec.ECInvSpec)
	}
	mode, err := getArg[string](spc.Args, ArgMode, RequiredRuleName)
	if err != nil {
		return RequiredRule{}, err
	}

	var rule RequiredRule
	switch mode {
	case "required":
		rule = Required
	case "not-empty":
		rule = NotEmpty
	case "not-nil":
		rule = NotNil
	default:
		format := "%s: invalid spec rule mode: %q"
		msg := fmt.Sprintf(format, RequiredRuleName, mode)
		return RequiredRule{}, NewInternalError(msg, spec.ECInvSpec)
	}

	if spc.ArgExist(ArgErrMsg) {
		msg, err := getArg[string](spc.Args, ArgErrMsg, RequiredRuleName)
		if err != nil {
			return RequiredRule{}, err
		}
		if msg != "" {
			rule = rule.Message(msg)
		}
	}

	if spc.ArgExist(ArgErrCode) {
		code, err := getArg[string](spc.Args, ArgErrCode, RequiredRuleName)
		if err != nil {
			return RequiredRule{}, err
		}
		if code != "" {
			rule = rule.Code(code)
		}
	}

	return rule, nil
}
