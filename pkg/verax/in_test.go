// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_In(t *testing.T) {
	// --- When ---
	have := In("a", "b", "c")

	// --- Then ---
	want := InRule{
		mode:      "in",
		want:      []EqualRule{Equal("a"), Equal("b"), Equal("c")},
		condition: true,
		msg:       msgIn,
		code:      ECInvIn,
		flags:     0,
	}
	assert.Equal(t, want, have)
}

func Test_NotIn(t *testing.T) {
	// --- When ---
	have := NotIn("a", "b", "c")

	// --- Then ---
	want := InRule{
		mode:      "not-in",
		want:      []EqualRule{Equal("a"), Equal("b"), Equal("c")},
		condition: true,
		msg:       msgNotIn,
		code:      ECInvIn,
		flags:     0,
	}
	assert.Equal(t, want, have)
}

func Test_InRule_Validate(t *testing.T) {
	t.Run("nil ok", func(t *testing.T) {
		// --- Given ---
		r := InRule{condition: true}

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("empty ok", func(t *testing.T) {
		// --- Given ---
		r := InRule{condition: true}

		// --- When ---
		err := r.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - in - custom error message and code", func(t *testing.T) {
		// --- Given ---
		r := In(42).Message("test err").Code("ECTst")

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "test err", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})

	t.Run("error - not in - custom error message and code", func(t *testing.T) {
		// --- Given ---
		r := NotIn(42).Message("test err").Code("ECTst")

		// --- When ---
		err := r.Validate(42)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "test err", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})

	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		r := In(42).When(false)

		// --- When ---
		err := r.Validate(44)

		// --- Then ---
		assert.NoError(t, err)
	})
}

func Test_InRule_Validate_in_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want []any
		have any
	}{
		{"int slice - zero value", []any{1, 2}, 0},
		{"int slice - nil pointer to int", []any{1, 2}, pIntNil},
		{"int slice - zero index", []any{1, 2}, 1},
		{"int slice - non-zero index", []any{1, 2}, 2},
		{"int slice - pointer to value", []any{1, 123}, &iInt},
		{"string slice - in", []any{"a", "b"}, "b"},
		{"string slice - zero value", []any{"a", "b"}, ""},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := In(tc.want...)

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_InRule_Validate_not_in_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want []any
		have any
	}{
		{"any slice - zero value", []any{1, 2}, 0},
		{"any slice - value does not exist", []any{1, 2}, 3},
		{"any slice - pointer to value", []any{1, 2}, &iInt},
		{"any slice - nil pointer to value", []any{1, 2}, pIntNil},
		{"empty any slice", []any{}, 2},
		{"string slice - not in", []any{"a", "b"}, "c"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := NotIn(tc.want...)

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_InRule_Validate_in_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want []any
		have any
	}{
		{"int slice", []any{1, 2}, 3},
		{"empty int slice", []any{}, 3},
		{"string slice", []any{"a", "b"}, "c"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := In(tc.want...)

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.SameType(t, &Error{}, err)
			assert.ErrorEqual(t, msgIn, err)
			xrrtest.AssertCode(t, ECInvIn, err)
		})
	}
}

func Test_InRule_Validate_not_in_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		want []any
		have any
	}{
		{"any slice", []any{1, 2}, 2},
		{"string slice", []any{"a", "b"}, "a"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := NotIn(tc.want...)

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.SameType(t, &Error{}, err)
			assert.ErrorEqual(t, msgNotIn, err)
			xrrtest.AssertCode(t, ECInvIn, err)
		})
	}
}

func Test_InRule_When(t *testing.T) {
	// --- Given ---
	r := InRule{}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_InRule_Message(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := InRule{}

		// --- When ---
		have := r.Message("test err")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Equal(t, flgCustomMsg, have.flags)
	})

	t.Run("an empty string is a noop", func(t *testing.T) {
		// --- Given ---
		r := InRule{msg: "test err"}

		// --- When ---
		have := r.Message("")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Zero(t, have.flags)
	})
}

func Test_InRule_Code(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := InRule{}

		// --- When ---
		have := r.Code("ECTst")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Equal(t, flgCustomCode, have.flags)
	})

	t.Run("an empty string is noop", func(t *testing.T) {
		// --- Given ---
		r := InRule{code: "ECTst"}

		// --- When ---
		have := r.Code("")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Zero(t, have.flags)
	})
}

func Test_InRule_Spec(t *testing.T) {
	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		r := InRule{mode: "bad-mode"}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `in-rule: invalid rule mode: "bad-mode"`, err)
		xrrtest.AssertCode(t, ECInvRuleMode, err)
		assert.Nil(t, have)
	})

	t.Run("mode in", func(t *testing.T) {
		// --- Given ---
		r := In(42, 44)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, InRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:        "in",
			spec.ArgValues: []any{42, 44},
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("mode not-in", func(t *testing.T) {
		// --- Given ---
		r := NotIn(42, 44)

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, InRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:        "not-in",
			spec.ArgValues: []any{42, 44},
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		r := In(42).Code("ECTst")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, InRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:        "in",
			spec.ArgValues: []any{42},
			ArgErrCode:     "ECTst",
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		r := In(42).Message("test err")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, InRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:        "in",
			spec.ArgValues: []any{42},
			ArgErrMsg:      "test err",
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("In - JSON encode", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()

		// --- When ---
		spc, err := In("a", "b", "c").Spec()

		// --- Then ---
		assert.NoError(t, err)
		data := must.Value(reg.EncodeSpec(spc))
		want := `{
			"name": "in-rule",
			"args": {
				"mode": "in",
				"values": ["a", "b", "c"]
			}
		}`
		assert.JSON(t, want, data)
	})

	t.Run("In - JSON decode", func(t *testing.T) {
		// --- Given ---
		reg := spec.NewRegistry[Rule]()
		reg.RegisterBuilders(Builders())
		data := []byte(`{
			"name": "in-rule",
			"args": {
				"mode": "in",
				"values": ["a", "b", "c"]
			}
		}`)

		// --- When ---
		have, err := reg.DecodeAndBuild(data)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, In("a", "b", "c"), have)
	})
}

func Test_InRuleFromSpec(t *testing.T) {
	t.Run("error - not in rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `in-rule: invalid spec name: "bad-name"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - mode argument is required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(InRuleName)

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "in-rule: spec missing required argument: mode"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - want argument is missing", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(InRuleName).SetArg(ArgMode, "in")

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "in-rule: spec missing required argument: values"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(InRuleName).
			SetArg(ArgMode, "bad-mode").
			SetArg(spec.ArgValues, []any{42, 44})

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, `in-rule: invalid spec rule mode: "bad-mode"`, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("mode in", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(InRuleName).
			SetArg(ArgMode, "in").
			SetArg(spec.ArgValues, []any{42, 44})

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := In(42, 44)
		assert.Equal(t, wRule, have)
	})

	t.Run("mode not-in", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(InRuleName).
			SetArg(ArgMode, "not-in").
			SetArg(spec.ArgValues, []any{42, 44})

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := NotIn(42, 44)
		assert.Equal(t, wRule, have)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(InRuleName).
			SetArg(ArgMode, "in").
			SetArg(spec.ArgValues, []any{42}).
			SetArg(ArgErrMsg, "test err")

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := In(42).Message("test err")
		assert.Equal(t, wRule, have)
	})

	t.Run("error - custom error message not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(InRuleName).
			SetArg(ArgMode, "in").
			SetArg(spec.ArgValues, []any{42}).
			SetArg(ArgErrMsg, 42)

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `in-rule: spec argument "err_msg" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("an empty custom error message is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(InRuleName).
			SetArg(ArgMode, "in").
			SetArg(spec.ArgValues, []any{42}).
			SetArg(ArgErrMsg, "")

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := In(42)
		assert.Equal(t, wRule, have)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(InRuleName).
			SetArg(ArgMode, "in").
			SetArg(spec.ArgValues, []any{42}).
			SetArg(ArgErrCode, "ECTst")

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := In(42).Code("ECTst")
		assert.Equal(t, wRule, have)
	})

	t.Run("error - custom error code not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(InRuleName).
			SetArg(ArgMode, "in").
			SetArg(spec.ArgValues, []any{42}).
			SetArg(ArgErrCode, 42)

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `in-rule: spec argument "err_code" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("an empty custom error code is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(InRuleName).
			SetArg(ArgMode, "in").
			SetArg(spec.ArgValues, []any{42}).
			SetArg(ArgErrCode, "")

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := In(42)
		assert.Equal(t, wRule, have)
	})
}

func Test_InRule_Spec_InRuleFromSpec_round_trip(t *testing.T) {
	t.Run("In - with message and code", func(t *testing.T) {
		// --- Given ---
		want := In(1, 2, 3).Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("NotIn - with message and code", func(t *testing.T) {
		// --- Given ---
		want := NotIn(1, 2, 3).Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := InRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}
