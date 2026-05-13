// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"database/sql"
	"testing"
	"text/template"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_Length(t *testing.T) {
	t.Run("between", func(t *testing.T) {
		// --- When ---
		have := Length(2, 3)

		// --- Then ---
		want := LengthRule{
			mode:      "length",
			min:       2,
			max:       3,
			condition: true,
			tpl:       msgLengthOutOfRange,
			msg:       "the length must be between 2 and 3",
			code:      ECInvLength,
			sticky:    nil,
			flags:     0,
		}
		assert.Equal(t, want, have)
	})

	t.Run("exactly", func(t *testing.T) {
		// --- When ---
		have := Length(2, 2)

		// --- Then ---
		want := LengthRule{
			mode:      "length",
			min:       2,
			max:       2,
			condition: true,
			tpl:       msgLengthInvalid,
			msg:       "the length must be exactly 2",
			code:      ECInvLength,
			sticky:    nil,
			flags:     0,
		}
		assert.Equal(t, want, have)
	})

	t.Run("empty", func(t *testing.T) {
		// --- When ---
		have := Length(0, 0)

		// --- Then ---
		want := LengthRule{
			mode:      "length",
			min:       0,
			max:       0,
			condition: true,
			tpl:       msgLengthReqEmpty,
			msg:       msgLengthReqEmpty,
			code:      ECInvLength,
			sticky:    nil,
			flags:     0,
		}
		assert.Equal(t, want, have)
	})
}

func Test_RuneLength(t *testing.T) {
	// --- When ---
	have := RuneLength(2, 3)

	// --- Then ---
	want := LengthRule{
		mode:      "rune-length",
		min:       2,
		max:       3,
		condition: true,
		tpl:       msgLengthOutOfRange,
		msg:       "the length must be between 2 and 3",
		code:      ECInvLength,
		sticky:    nil,
		flags:     0,
	}
	assert.Equal(t, want, have)
}

func Test_LengthRule_Validate(t *testing.T) {
	t.Run("error - sticky", func(t *testing.T) {
		// --- Given ---
		r := LengthRule{sticky: ErrTst}

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.Same(t, ErrTst, err)
	})

	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		r := Length(42, 44).When(false)

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("nil ok", func(t *testing.T) {
		// --- Given ---
		r := Length(42, 44)

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty ok", func(t *testing.T) {
		// --- Given ---
		r := Length(42, 44)

		// --- When ---
		err := r.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("custom error message and code", func(t *testing.T) {
		// --- Given ---
		r := Length(42, 44).Message("{{.min}} - {{.max}}").Code("ECTst")

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "42 - 44", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- Given ---
		r := Length(42, 44)

		// --- When ---
		err := r.Validate(123)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "cannot get the length of int", err)
		xrrtest.AssertCode(t, ECInvType, err)
	})
}

func Test_LengthRule_Validate_Length_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rule LengthRule
		have any
	}{
		// Length.
		{"zero length", Length(0, 0), ""},
		{"exact length", Length(2, 2), "ab"},
		{"zero value string", Length(2, 4), ""},
		{"within range", Length(2, 4), "abc"},
		{"zero to length", Length(0, 4), "ab"},
		{"nil pointer to string", Length(0, 2), pStringNil},
		{"min bigger than max", Length(2, 0), "ab"},
		{
			"Valuer within range",
			Length(2, 4),
			sql.NullString{String: "abc", Valid: true},
		},
		{"Valuer empty", Length(2, 4), sql.NullString{String: "", Valid: true}},
		{
			"pointer to Valuer within range",
			Length(2, 4),
			&sql.NullString{String: "abc", Valid: true},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			rule := tc.rule.Message("test tpl").Code("ECTst")

			// --- When ---
			err := rule.Validate(tc.have)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_LengthRule_Validate_RuneLength_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rule LengthRule
		have any
	}{
		{"zero value string", RuneLength(2, 4), ""},
		{"within range", RuneLength(2, 4), "abc"},
		{"within range emoji", RuneLength(2, 3), "💥💥"},
		{"within range emoji", RuneLength(2, 3), "💥💥💥"},
		{"within - range min is zero", RuneLength(0, 4), "ab"},
		{"within - range max is zero", RuneLength(2, 0), "ab"},
		{"nil pointer to string", RuneLength(0, 2), pStringNil},
		{
			"Valuer within range",
			RuneLength(2, 4),
			sql.NullString{String: "abc", Valid: true},
		},
		{
			"Valuer zero value",
			RuneLength(2, 4),
			sql.NullString{String: "", Valid: true},
		},
		{
			"pointer to Valuer within range",
			RuneLength(2, 4),
			&sql.NullString{String: "abc", Valid: true},
		},
		{
			"pointer to Valuer within range - emoji",
			RuneLength(2, 3),
			&sql.NullString{String: "💥💥", Valid: true},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			rule := tc.rule.Message("test tpl").Code("ECTst")

			// --- When ---
			err := rule.Validate(tc.have)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_LengthRule_Validate_Length_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rule LengthRule
		have any
	}{
		// Length.
		{"too long", Length(2, 4), "abcdf"},
		{"too long - min is zero", Length(0, 4), "abcde"},
		{"too short", Length(2, 4), "a"},
		{"too short - max is zero", Length(2, 0), "a"},
		{"not of the exact length", Length(2, 2), "abcdf"},
		{"must be empty", Length(0, 0), "ab"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			rule := tc.rule.Message("test tpl").Code("ECTst")

			// --- When ---
			err := rule.Validate(tc.have)

			// --- Then ---
			assert.ErrorEqual(t, "test tpl", err)
			xrrtest.AssertCode(t, "ECTst", err)
		})
	}
}

func Test_LengthRule_Validate_RuneLength_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rule LengthRule
		have any
	}{
		{"runes too short", RuneLength(2, 3), "💥"},
		{"runes too long", RuneLength(2, 3), "💥💥💥💥"},
		{"runes string too long", RuneLength(2, 4), "abcdf"},
		{"runes string too long - min zero", RuneLength(0, 4), "abcde"},
		{"runes string too short - max iz zero", RuneLength(2, 0), "a"},
		{
			"pointer to Valuer too short ",
			RuneLength(2, 3),
			&sql.NullString{String: "💥", Valid: true},
		},
		{"runes must be empty", RuneLength(0, 0), "ab"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			rule := tc.rule.Message("test tpl").Code("ECTst")

			// --- When ---
			err := rule.Validate(tc.have)

			// --- Then ---
			assert.ErrorEqual(t, "test tpl", err)
			xrrtest.AssertCode(t, "ECTst", err)
		})
	}
}

func Test_LengthRule_When(t *testing.T) {
	// --- Given ---
	r := LengthRule{}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_LengthRule_Message(t *testing.T) {
	t.Run("an empty string is a noop", func(t *testing.T) {
		// --- Given ---
		r := Length(1, 2)

		// --- When ---
		have := r.Message("")

		// --- Then ---
		assert.Equal(t, msgLengthOutOfRange, have.tpl)
		assert.Equal(t, "the length must be between 1 and 2", have.msg)
		assert.Zero(t, have.flags)
	})

	t.Run("when the sticky error is not nil", func(t *testing.T) {
		// --- Given ---
		r := Length(1, 2)
		r.sticky = ErrTst

		// --- When ---
		have := r.Message("{{.min}} - {{.max}}")

		// --- Then ---
		assert.Equal(t, msgLengthOutOfRange, have.tpl)
		assert.Equal(t, "the length must be between 1 and 2", have.msg)
		assert.Zero(t, have.flags)
	})

	t.Run("valid template", func(t *testing.T) {
		// --- Given ---
		r := Length(1, 2)

		// --- When ---
		have := r.Message("{{.min}} - {{.max}}")

		// --- Then ---
		assert.NoError(t, have.sticky)
		assert.Equal(t, "{{.min}} - {{.max}}", have.tpl)
		assert.Equal(t, "1 - 2", have.msg)
		assert.Equal(t, flgCustomMsg, have.flags)
	})

	t.Run("error - template parse error", func(t *testing.T) {
		// --- Given ---
		r := Length(1, 2)
		tpl := r.tpl
		msg := r.msg

		// --- When ---
		have := r.Message("{{.")

		// --- Then ---
		assert.Equal(t, tpl, have.tpl)
		assert.Equal(t, msg, have.msg)
		assert.Zero(t, have.flags)

		assert.SameType(t, &InternalError{}, have.sticky)
		wMsg := "length-rule(length): custom template parse error"
		assert.ErrorEqual(t, wMsg, have.sticky)
		xrrtest.AssertCode(t, ECInternal, have.sticky)
	})

	t.Run("error - with not supported custom placeholder", func(t *testing.T) {
		// --- Given ---
		r := Length(1, 2)
		tpl := r.tpl
		msg := r.msg

		// --- When ---
		have := r.Message("{{.custom}}")

		// --- Then ---
		assert.Equal(t, tpl, have.tpl)
		assert.Equal(t, msg, have.msg)
		assert.Equal(t, msgLengthOutOfRange, have.tpl)
		assert.Equal(t, "the length must be between 1 and 2", have.msg)
		assert.Zero(t, have.flags)

		assert.SameType(t, &InternalError{}, have.sticky)
		wMsg := "length-rule(length): custom template render error"
		assert.ErrorEqual(t, wMsg, have.sticky)
		xrrtest.AssertCode(t, ECInternal, have.sticky)
	})
}

func Test_LengthRule_Code(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := LengthRule{}

		// --- When ---
		have := r.Code("ECTst")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Equal(t, flgCustomCode, have.flags)
	})

	t.Run("an empty string is noop", func(t *testing.T) {
		// --- Given ---
		r := LengthRule{code: "ECTst"}

		// --- When ---
		have := r.Code("")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Zero(t, have.flags)
	})
}

func Test_LengthRule_Spec(t *testing.T) {
	t.Run("error - sticky", func(t *testing.T) {
		// --- Given ---
		r := LengthRule{sticky: ErrTst}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.Same(t, ErrTst, err)
		assert.Nil(t, have)
	})

	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		r := LengthRule{mode: "unknown-mode"}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `length-rule: invalid rule mode: "unknown-mode"`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvRuleMode, err)
		assert.Nil(t, have)
	})

	t.Run("mode length", func(t *testing.T) {
		// --- Given ---
		r := Length(42, 44)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, LengthRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode: "length",
			ArgMin:  42,
			ArgMax:  44,
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("mode rune-length", func(t *testing.T) {
		// --- Given ---
		r := RuneLength(42, 44)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, LengthRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode: "rune-length",
			ArgMin:  42,
			ArgMax:  44,
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		r := Length(42, 44).Message("{{.min}} - {{.max}}")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, LengthRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:   "length",
			ArgMin:    42,
			ArgMax:    44,
			ArgErrMsg: "{{.min}} - {{.max}}",
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		r := RuneLength(42, 44).Code("ECTst")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, LengthRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:    "rune-length",
			ArgMin:     42,
			ArgMax:     44,
			ArgErrCode: "ECTst",
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("Length - JSON representation", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()

		// --- When ---
		spc, err := Length(3, 10).Spec()

		// --- Then ---
		assert.NoError(t, err)
		data := must.Value(reg.EncodeSpec(spc))
		want := `{
			"name": "length-rule",
			"args": {
				"mode": {"type": "string", "value": "length"},
				"min": {"type": "int", "value": 3},
				"max": {"type": "int", "value": 10}
			}
		}`
		assert.JSON(t, want, data)
	})
}

func Test_LengthRuleFromSpec(t *testing.T) {
	t.Run("error - not length rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `length-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - mode argument is required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(LengthRuleName)

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "length-rule: spec missing required argument: mode"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - invalid min type", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(LengthRuleName).
			SetArg(ArgMode, "length").
			SetArg(ArgMin, "str").
			SetArg(ArgMax, 44).
			SetArg(ArgErrCode, ECInvLength)

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `length-rule: spec argument "min" must be int, got string`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - invalid max type", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(LengthRuleName).
			SetArg(ArgMode, "length").
			SetArg(ArgMin, 44).
			SetArg(ArgMax, "str").
			SetArg(ArgErrCode, ECInvLength)

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `length-rule: spec argument "max" must be int, got string`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(LengthRuleName).
			SetArg(ArgMode, "bad-mode").
			SetArg(ArgMin, 42).
			SetArg(ArgMax, 44).
			SetArg(ArgErrCode, ECInvLength)

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `length-rule: invalid spec rule mode: "bad-mode"`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("mode length", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(LengthRuleName).
			SetArg(ArgMode, "length").
			SetArg(ArgMin, 42).
			SetArg(ArgMax, 44)

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Length(42, 44)
		assert.Equal(t, wRule, have)
	})

	t.Run("mode rune-length", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(LengthRuleName).
			SetArg(ArgMode, "rune-length").
			SetArg(ArgMin, 42).
			SetArg(ArgMax, 44)

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := RuneLength(42, 44)
		assert.Equal(t, wRule, have)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(LengthRuleName).
			SetArg(ArgMode, "length").
			SetArg(ArgMin, 42).
			SetArg(ArgMax, 44).
			SetArg(ArgErrMsg, "test {{.min}} {{.max}}")

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Length(42, 44).Message("test {{.min}} {{.max}}")
		assert.Equal(t, wRule, have)
	})

	t.Run("error - custom error message not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(LengthRuleName).
			SetArg(ArgMode, "rune-length").
			SetArg(ArgMin, 42).
			SetArg(ArgMax, 44).
			SetArg(ArgErrMsg, true)

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `length-rule: spec argument "err_msg" must be string, got bool`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("an empty custom error message is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(LengthRuleName).
			SetArg(ArgMode, "length").
			SetArg(ArgMin, 42).
			SetArg(ArgMax, 44).
			SetArg(ArgErrMsg, "")

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Length(42, 44)
		assert.Equal(t, wRule, have)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(LengthRuleName).
			SetArg(ArgMode, "length").
			SetArg(ArgMin, 42).
			SetArg(ArgMax, 44).
			SetArg(ArgErrCode, "ECTst")

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Length(42, 44).Code("ECTst")
		assert.Equal(t, wRule, have)
	})

	t.Run("an empty custom error code is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(LengthRuleName).
			SetArg(ArgMode, "rune-length").
			SetArg(ArgMin, 42).
			SetArg(ArgMax, 44).
			SetArg(ArgErrCode, "")

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)

		assert.Equal(t, msgLengthOutOfRange, have.tpl)
		assert.Equal(t, "the length must be between 42 and 44", have.msg)
		assert.Equal(t, ECInvLength, have.code)
	})

	t.Run("error - error code not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(LengthRuleName).
			SetArg(ArgMode, "rune-length").
			SetArg(ArgMin, 42).
			SetArg(ArgMax, 44).
			SetArg(ArgErrCode, 42)

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `length-rule: spec argument "err_code" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})
}

func Test_pickLengthRuleMsg_tabular(t *testing.T) {
	tt := []struct {
		testN string

		min int
		max int
		msg string
		tpl *template.Template
	}{
		{"not longer than", 0, 1, msgLengthTooLong, tplLengthTooLong},
		{"not shorted than", 1, 0, msgLengthTooShort, tplLengthTooShort},
		{"range", 1, 2, msgLengthOutOfRange, tplLengthOutOfRange},
		{"non zero exact length", 2, 2, msgLengthInvalid, tplLengthInvalid},
		{"invalid negative min", -1, 2, "", nil},
		{"invalid negative max", 1, -2, "", nil},
		{"invalid both max", -2, -1, "", nil},
		{"invalid min is bigger than max", 4, 1, "", nil},
		{"min and max equal zero", 0, 0, msgLengthReqEmpty, tplLengthReqEmpty},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			hMsg, hTpl := pickLengthRuleMsg(tc.min, tc.max)

			// --- Then ---
			assert.Equal(t, tc.msg, hMsg)
			assert.Same(t, tc.tpl, hTpl)
		})
	}
}

func Test_buildLengthRuleMsg(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		hMsg, hTpl, err := buildLengthRuleMsg(1, 2, "mode")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, msgLengthOutOfRange, hTpl)
		assert.Equal(t, "the length must be between 1 and 2", hMsg)
	})

	t.Run("error - invalid min and max", func(t *testing.T) {
		// --- When ---
		hMsg, hTpl, err := buildLengthRuleMsg(-1, 0, "mode")

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "length-rule(mode): custom template parse error"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInternal, err)
		assert.Empty(t, hTpl)
		assert.Empty(t, hMsg)
	})

	t.Run("error - execute template", func(t *testing.T) {
		// --- Given ---
		orig := tplLengthOutOfRange
		tplLengthOutOfRange = mustTpl("test tpl", "tpl {{.other}}")
		t.Cleanup(func() { tplLengthOutOfRange = orig })

		// --- When ---
		hMsg, hTpl, err := buildLengthRuleMsg(1, 2, "mode")

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "length-rule(mode): custom template render error"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInternal, err)
		assert.Empty(t, hTpl)
		assert.Empty(t, hMsg)
	})
}

func Test_LengthRule_Spec_LengthRuleFromSpec_round_trip(t *testing.T) {
	t.Run("Length - with message and code", func(t *testing.T) {
		// --- Given ---
		want := Length(42, 44).Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("RuneLengrth - with message and code", func(t *testing.T) {
		// --- Given ---
		want := RuneLength(42, 44).Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := LengthRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}
