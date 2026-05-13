// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_Match(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		re := regexp.MustCompile(`\d+`)

		// --- When ---
		have := Match(re)

		// --- Then ---
		want := MatchRule{
			want:      re,
			condition: true,
			msg:       msgInvMatch,
			code:      ECInvMatch,
			sticky:    nil,
			flags:     0,
		}
		assert.Equal(t, want, have)
	})

	t.Run("error - nil regexp", func(t *testing.T) {
		// --- When ---
		r := Match(nil)

		// --- Then ---
		assert.SameType(t, &InternalError{}, r.sticky)
		assert.ErrorEqual(t, "match-rule: nil regexp", r.sticky)
		xrrtest.AssertCode(t, ECInternal, r.sticky)
	})
}

func Test_MatchRule_Validate(t *testing.T) {
	t.Run("error - sticky", func(t *testing.T) {
		// --- Given ---
		r := MatchRule{sticky: ErrTst}

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.Same(t, ErrTst, err)
	})

	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		r := Match(regexp.MustCompile(`\d+`)).When(false)

		// --- When ---
		err := r.Validate("abc")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("nil ok", func(t *testing.T) {
		// --- Given ---
		r := Match(regexp.MustCompile(`\d+`))

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty ok", func(t *testing.T) {
		// --- Given ---
		r := Match(regexp.MustCompile(`\d+`))

		// --- When ---
		err := r.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})
}

func Test_MatchRule_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rx   string
		have any
	}{
		{"nil", "[a-z]+", nil},
		{"zero value string", "[a-z]+", ""},
		{"string", "[a-z]+", "abc"},
		{"nil pointer to string", "[a-z]+", pStringNil},
		{"byte slice", "[a-z]+", []byte("abc")},
		{"empty byte slice", "[a-z]+", []byte{}},
		{"byte slice from empty string", "[a-z]+", []byte("")},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := Match(regexp.MustCompile(tc.rx))

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_MatchRule_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rx   string
		have any
	}{
		{"string", "[a-z]+", "123"},
		{"byte slice", "[a-z]+", []byte("123")},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := Match(regexp.MustCompile(tc.rx)).
				Message("test err").
				Code("ECTst")

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.SameType(t, &Error{}, err)
			assert.ErrorEqual(t, "test err", err)
			xrrtest.AssertCode(t, "ECTst", err)
		})
	}
}

func Test_MatchRule_When(t *testing.T) {
	// --- Given ---
	r := MatchRule{}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_MatchRule_Message(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := MatchRule{}

		// --- When ---
		have := r.Message("test err")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Equal(t, flgCustomMsg, have.flags)
	})

	t.Run("an empty string is a noop", func(t *testing.T) {
		// --- Given ---
		r := MatchRule{msg: "test err"}

		// --- When ---
		have := r.Message("")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Zero(t, have.flags)
	})
}

func Test_MatchRule_Code(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := MatchRule{}

		// --- When ---
		have := r.Code("ECTst")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Equal(t, flgCustomCode, have.flags)
	})

	t.Run("an empty string is noop", func(t *testing.T) {
		// --- Given ---
		r := MatchRule{code: "ECTst"}

		// --- When ---
		have := r.Code("")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Zero(t, have.flags)
	})
}

func Test_MatchRule_Spec(t *testing.T) {
	t.Run("error - sticky", func(t *testing.T) {
		// --- Given ---
		r := MatchRule{sticky: ErrTst}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.Same(t, ErrTst, err)
		assert.Nil(t, have)
	})

	t.Run("Match", func(t *testing.T) {
		// --- Given ---
		r := Match(regexp.MustCompile(`\d+`))

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, MatchRuleName, have.Name)
		assert.Equal(t, map[string]any{spec.ArgValue: `\d+`}, have.Args)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		r := Match(regexp.MustCompile(`\d+`)).Message("test err")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, MatchRuleName, have.Name)
		wArgs := map[string]any{
			spec.ArgValue: `\d+`,
			ArgErrMsg:     "test err",
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		r := Match(regexp.MustCompile(`\d+`)).Code("ECTst")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, MatchRuleName, have.Name)
		wArgs := map[string]any{spec.ArgValue: `\d+`, ArgErrCode: "ECTst"}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("Match - JSON representation", func(t *testing.T) {
		// --- When ---
		have, err := Match(regexp.MustCompile(`\d+`)).Spec()

		// --- Then ---
		assert.NoError(t, err)
		data := must.Value(json.Marshal(have))
		want := `{
			"name": "match-rule",
			"args": {
				"value": "\\d+"
			}
		}`
		assert.JSON(t, want, data)
	})
}

func Test_MatchRuleFromSpec(t *testing.T) {
	t.Run("error - not match rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := MatchRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `match-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - want argument is missing", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MatchRuleName)

		// --- When ---
		have, err := MatchRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "match-rule: spec missing required argument: value"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - want argument is of an unexpected type", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MatchRuleName).SetArg(spec.ArgValue, true)

		// --- When ---
		have, err := MatchRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `match-rule: spec argument "value" must be string, got bool`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - rx argument is an invalid regexp", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MatchRuleName).SetArg(spec.ArgValue, "[")

		// --- When ---
		have, err := MatchRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, "match-rule: invalid regexp: `[`", err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("Match", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MatchRuleName).SetArg(spec.ArgValue, `\d+`)

		// --- When ---
		have, err := MatchRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Match(regexp.MustCompile(`\d+`))
		assert.Equal(t, wRule, have)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MatchRuleName).
			SetArg(spec.ArgValue, `\d+`).
			SetArg(ArgErrMsg, "test err")

		// --- When ---
		have, err := MatchRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Match(regexp.MustCompile(`\d+`)).Message("test err")
		assert.Equal(t, wRule, have)
	})

	t.Run("error - custom error message not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MatchRuleName).
			SetArg(spec.ArgValue, `\d+`).
			SetArg(ArgErrMsg, 42)

		// --- When ---
		have, err := MatchRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `match-rule: spec argument "err_msg" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("an empty custom error message is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MatchRuleName).
			SetArg(spec.ArgValue, `\d+`).
			SetArg(ArgErrMsg, "")

		// --- When ---
		have, err := MatchRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Match(regexp.MustCompile(`\d+`))
		assert.Equal(t, wRule, have)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MatchRuleName).
			SetArg(spec.ArgValue, `\d+`).
			SetArg(ArgErrCode, "ECTst")

		// --- When ---
		have, err := MatchRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Match(regexp.MustCompile(`\d+`)).Code("ECTst")
		assert.Equal(t, wRule, have)
	})

	t.Run("error - custom error code not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MatchRuleName).
			SetArg(spec.ArgValue, `\d+`).
			SetArg(ArgErrCode, 42)

		// --- When ---
		have, err := MatchRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `match-rule: spec argument "err_code" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("an empty custom error code is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(MatchRuleName).
			SetArg(spec.ArgValue, `\d+`).
			SetArg(ArgErrCode, "")

		// --- When ---
		have, err := MatchRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Match(regexp.MustCompile(`\d+`))
		assert.Equal(t, wRule, have)
	})
}

func Test_MatchRule_Spec_MatchRuleFromSpec_round_trip(t *testing.T) {
	// --- Given ---
	rx := regexp.MustCompile(`\d+`)
	want := Match(rx).Message("test msg").Code("ECTst")
	spc := must.Value(want.Spec())

	// --- When ---
	have, err := MatchRuleFromSpec(spc)

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, want, have)
}
