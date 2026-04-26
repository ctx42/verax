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

func Test_Required(t *testing.T) {
	// --- Given ---
	have := Required

	// --- Then ---
	want := RequiredRule{
		mode:      "required",
		condition: true,
		msg:       msgMissing,
		code:      ECReq,
		flags:     0,
	}
	assert.Equal(t, want, have)
}

func Test_NotEmpty(t *testing.T) {
	// --- Given ---
	have := NotEmpty

	// --- Then ---
	want := RequiredRule{
		mode:      "not-empty",
		condition: true,
		msg:       msgReqNotEmpty,
		code:      ECReqNotEmpty,
		flags:     0,
	}
	assert.Equal(t, want, have)
}

func Test_NotNil(t *testing.T) {
	// --- Given ---
	have := NotNil

	// --- Then ---
	want := RequiredRule{
		mode:      "not-nil",
		condition: true,
		msg:       msgReqNotNil,
		code:      ECReqNotNil,
		flags:     0,
	}
	assert.Equal(t, want, have)
}

func Test_RequiredRule_Validate(t *testing.T) {
	t.Run("skip validation when the condition is false", func(t *testing.T) {
		// --- Given ---
		r := RequiredRule{condition: false}

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - nil - uses message and code", func(t *testing.T) {
		// --- Given ---
		r := RequiredRule{condition: true, msg: "test err", code: "ECTst"}

		// --- When ---
		err := r.Validate(nil)

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "test err", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})

	t.Run("error - empty - uses message and code", func(t *testing.T) {
		// --- Given ---
		r := RequiredRule{condition: true, msg: "test err", code: "ECTst"}

		// --- When ---
		err := r.Validate("")

		// --- Then ---
		assert.SameType(t, &Error{}, err)
		assert.ErrorEqual(t, "test err", err)
		xrrtest.AssertCode(t, "ECTst", err)
	})
}

func Test_RequiredRule_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rule Rule
		have any
	}{
		// Required.
		{"Required int", Required, 123},
		{"Required string", Required, iString},
		{"Required time", Required, iTime},
		{"Required func", Required, iFunc},
		{"Required bool true", Required, true},
		{"Required bool false", Required, false},

		// NotEmpty.
		{"NotEmpty nil", NotEmpty, nil},
		{"NotEmpty string pointer", NotEmpty, pString},
		{"NotEmpty string nil pointer", NotEmpty, pStringNil},
		{"NotEmpty int pointer", NotEmpty, pInt},
		{"NotEmpty int nil pointer", NotEmpty, pIntNil},
		{"NotEmpty time pointer", NotEmpty, pTime},
		{"NotEmpty time nil pointer", NotEmpty, pTimeNil},
		{"NotEmpty bool true", NotEmpty, true},
		{"NotEmpty bool false", NotEmpty, false},

		{"NotEmpty empty struct nil pointer", NotEmpty, pStructEmptyNil},
		{"NotEmpty int", NotEmpty, 123},
		{"NotEmpty any(123)", NotEmpty, iInterface},

		// NotNil.
		{"NotNil empty string", NotNil, ""},
		{"NotNil zero int", NotNil, 0},
		{"NotNil pointer to string", NotNil, pString},
		{"NotNil pointer to empty string", NotNil, pStringEmpty},
		{"NotNil pointer to int", NotNil, pInt},
		{"NotNil pointer to zero value int", NotNil, pIntZero},
		{"NotNil pointer to time", NotNil, pTime},
		{"NotNil pointer to zero value time", NotNil, pTimeZero},
		{"NotNil pointer to empty struct", NotNil, pStructEmpty},
		{"NotNil bool true", NotNil, true},
		{"NotNil bool false", NotNil, false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := tc.rule

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_RequiredRule_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		rule Rule
		have any
	}{
		// Required.
		{"Required nil", Required, nil},
		{"Required empty string", Required, iStringEmpty},
		{"Required declared zero value time", Required, dTime},
		{"Required zero value time", Required, iTimeZero},
		{"Required empty struct", Required, iStructEmpty},
		{"Required pointer to empty struct", Required, pStructEmpty},
		{"Required chan", Required, iChan},

		// NotEmpty.
		{"NotEmpty empty string", NotEmpty, iStringEmpty},
		{"NotEmpty pointer to empty string", NotEmpty, pStringEmpty},
		{"NotEmpty zero value int", NotEmpty, iIntZero},
		{"NotEmpty pointer to zero value int", NotEmpty, pIntZero},
		{"NotEmpty zero value time", NotEmpty, iTimeZero},
		{"NotEmpty pointer to zero value time", NotEmpty, pTimeZero},
		{"NotEmpty any(0)", NotEmpty, iInterfaceZero},
		{"NotEmpty pointer to empty struct", NotEmpty, pStructEmpty},
		{"NotEmpty struct with empty fields", NotEmpty, ModelPtr{}},

		// NotNil
		{"NotNil nil slice", NotNil, dSlice},
		{"NotNil zero value array", NotNil, dMap},
		{"NotNil nil pointer to string", NotNil, pStringNil},
		{"NotNil nil pointer to int", NotNil, pIntNil},
		{"NotNil nil empty interface", NotNil, dInterface},
		{"NotNil nil interface", NotNil, dValidate},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			r := tc.rule

			// --- When ---
			err := r.Validate(tc.have)

			// --- Then ---
			assert.SameType(t, &Error{}, err)
			assert.Error(t, err)
		})
	}
}

func Test_RequiredRule_When(t *testing.T) {
	// --- Given ---
	r := RequiredRule{}

	// --- When ---
	have := r.When(true)

	// --- Then ---
	assert.True(t, have.condition)
}

func Test_RequiredRule_Message(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := RequiredRule{}

		// --- When ---
		have := r.Message("test err")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Equal(t, flgCustomMsg, have.flags)
	})

	t.Run("an empty string is a noop", func(t *testing.T) {
		// --- Given ---
		r := RequiredRule{msg: "test err"}

		// --- When ---
		have := r.Message("")

		// --- Then ---
		assert.Equal(t, "test err", have.msg)
		assert.Zero(t, have.flags)
	})
}

func Test_RequiredRule_Code(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		r := RequiredRule{}

		// --- When ---
		have := r.Code("ECTst")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Equal(t, flgCustomCode, have.flags)
	})

	t.Run("an empty string is noop", func(t *testing.T) {
		// --- Given ---
		r := RequiredRule{code: "ECTst"}

		// --- When ---
		have := r.Code("")

		// --- Then ---
		assert.Equal(t, "ECTst", have.code)
		assert.Zero(t, have.flags)
	})
}

func Test_RequiredRule_Spec(t *testing.T) {
	t.Run("mode required", func(t *testing.T) {
		// --- Given ---
		r := Required

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, RequiredRuleName, have.Name)
		wArgs := map[string]any{ArgMode: "required"}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("mode nor-empty", func(t *testing.T) {
		// --- Given ---
		r := NotEmpty

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, RequiredRuleName, have.Name)
		wArgs := map[string]any{ArgMode: "not-empty"}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("mode not-nil", func(t *testing.T) {
		// --- Given ---
		r := NotNil

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, RequiredRuleName, have.Name)
		wArgs := map[string]any{ArgMode: "not-nil"}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		r := RequiredRule{mode: "unknown"}

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `required-rule: invalid rule mode: "unknown"`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvRuleMode, err)
		assert.Nil(t, have)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		r := Required.Code("ECTst")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, RequiredRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:    "required",
			ArgErrCode: "ECTst",
		}
		assert.Equal(t, wArgs, have.Args)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		r := NotEmpty.Message("test msg")

		// --- When ---
		have, err := r.Spec()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, RequiredRuleName, have.Name)
		wArgs := map[string]any{
			ArgMode:   "not-empty",
			ArgErrMsg: "test msg",
		}
		assert.Equal(t, wArgs, have.Args)
	})
}

func Test_RequiredRuleFromSpec(t *testing.T) {
	t.Run("error - not required rule spec", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec("bad-name")

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `required-rule: invalid spec name: "bad-name"`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("error - mode argument is required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RequiredRuleName)

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "required-rule: spec missing required argument: mode"
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.ErrorEqual(t, wMsg, err)
		assert.Zero(t, have)
	})

	t.Run("error - invalid mode", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RequiredRuleName).SetArg(ArgMode, "bad-mode")

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `required-rule: invalid spec rule mode: "bad-mode"`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("required", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RequiredRuleName).SetArg(ArgMode, "required")

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Required
		assert.Equal(t, wRule, have)
	})

	t.Run("mode not-empty", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RequiredRuleName).SetArg(ArgMode, "not-empty")

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := NotEmpty
		assert.Equal(t, wRule, have)
	})

	t.Run("mode not-nil", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RequiredRuleName).SetArg(ArgMode, "not-nil")

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := NotNil
		assert.Equal(t, wRule, have)
	})

	t.Run("custom error message", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RequiredRuleName).
			SetArg(ArgMode, "required").
			SetArg(ArgErrMsg, "test err")

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Required.Message("test err")
		assert.Equal(t, wRule, have)
	})

	t.Run("an empty custom error message is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RequiredRuleName).
			SetArg(ArgMode, "required").
			SetArg(ArgErrMsg, "")

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Required
		assert.Equal(t, wRule, have)
	})

	t.Run("error - custom error message not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RequiredRuleName).
			SetArg(ArgMode, "required").
			SetArg(ArgErrMsg, 42)

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `required-rule: spec argument "err_msg" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})

	t.Run("custom error code", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RequiredRuleName).
			SetArg(ArgMode, "required").
			SetArg(ArgErrCode, "ECTst")

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Required.Code("ECTst")
		assert.Equal(t, wRule, have)
	})

	t.Run("an empty custom error code is ignored", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RequiredRuleName).
			SetArg(ArgMode, "required").
			SetArg(ArgErrCode, "")

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		wRule := Required
		assert.Equal(t, wRule, have)
	})

	t.Run("error - error code not string", func(t *testing.T) {
		// --- Given ---
		spc := spec.NewSpec(RequiredRuleName).
			SetArg(ArgMode, "required").
			SetArg(ArgErrCode, 42)

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `required-rule: spec argument "err_code" must be string, got int`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, spec.ECInvSpec, err)
		assert.Zero(t, have)
	})
}

func Test_RequiredRuleFromSpec_tabular(t *testing.T) {
	tt := []struct {
		testN string

		wMsg  string
		wCode string
	}{
		{"required", msgMissing, ECReq},
		{"not-empty", msgReqNotEmpty, ECReqNotEmpty},
		{"not-nil", msgReqNotNil, ECReqNotNil},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			spc := spec.NewSpec(RequiredRuleName).SetArg(ArgMode, tc.testN)

			// --- When ---
			have, err := RequiredRuleFromSpec(spc)

			// --- Then ---
			assert.NoError(t, err)

			assert.Equal(t, tc.testN, have.mode)
			assert.True(t, have.condition)
			assert.Equal(t, tc.wMsg, have.msg)
			assert.Equal(t, tc.wCode, have.code)
			assert.Equal(t, uint8(0), have.flags)
		})
	}
}

func Test_RequiredRule_Spec_RequiredRuleFromSpec_round_trip(t *testing.T) {
	t.Run("Required - with message and code", func(t *testing.T) {
		// --- Given ---
		want := Required.Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("NotEmpty - with message and code", func(t *testing.T) {
		// --- Given ---
		want := NotEmpty.Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})

	t.Run("NotNil - with message and code", func(t *testing.T) {
		// --- Given ---
		want := NotNil.Message("test msg").Code("ECTst")
		spc := must.Value(want.Spec())

		// --- When ---
		have, err := RequiredRuleFromSpec(spc)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, want, have)
	})
}
