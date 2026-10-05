// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_ToAnySlice(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		// --- When ---
		have := ToAnySlice[int]()

		// --- Then ---
		assert.Equal(t, []any{}, have)
	})

	t.Run("strings", func(t *testing.T) {
		// --- When ---
		have := ToAnySlice("a", "b", "c")

		// --- Then ---
		assert.Equal(t, []any{"a", "b", "c"}, have)
	})

	t.Run("ints", func(t *testing.T) {
		// --- When ---
		have := ToAnySlice(1, 2, 3)

		// --- Then ---
		assert.Equal(t, []any{1, 2, 3}, have)
	})
}

func Test_convertTo_tabular(t *testing.T) {
	type hostname string
	str := "abc"
	hst := hostname("abc")
	var nilStr *string

	tt := []struct {
		testN string

		val  any
		want string
		ok   bool
	}{
		{"string", "abc", "abc", true},
		{"named string", hst, "abc", true},
		{"pointer to string", &str, "abc", true},
		{"pointer to named string", &hst, "abc", true},
		{"nil pointer", nilStr, "", false},
		{"nil", nil, "", false},
		{"int", 65, "", false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, ok := convertTo[string](tc.val)

			// --- Then ---
			assert.Equal(t, tc.want, have)
			assert.Equal(t, tc.ok, ok)
		})
	}
}

func Test_AsRuleBuilder(t *testing.T) {
	t.Run("T not a Rule", func(t *testing.T) {
		// --- When ---
		have := AsRuleBuilder(func(_ *spec.Spec) (int, error) { return 0, nil })

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("typed constructor", func(t *testing.T) {
		// --- Given ---
		fn := func(_ *spec.Spec) (NoopRule, error) { return Noop, nil }

		// --- When ---
		bld := AsRuleBuilder(fn)

		// --- Then ---
		assert.NotNil(t, bld)
		have, err := bld(spec.NewSpec(NoopRuleName))
		assert.NoError(t, err)
		assert.Equal(t, Rule(Noop), have)
	})

	t.Run("constructor returning an interface", func(t *testing.T) {
		// --- Given ---
		fn := func(_ *spec.Spec) (Rule, error) { return Noop, nil }

		// --- When ---
		bld := AsRuleBuilder(fn)

		// --- Then ---
		assert.NotNil(t, bld)
		have, err := bld(spec.NewSpec(NoopRuleName))
		assert.NoError(t, err)
		assert.Equal(t, Rule(Noop), have)
	})

	t.Run("error - constructor", func(t *testing.T) {
		// --- Given ---
		wErr := NewInternalError("spec error", ECInternal)
		fn := func(_ *spec.Spec) (NoopRule, error) { return NoopRule{}, wErr }

		// --- When ---
		bld := AsRuleBuilder(fn)
		have, err := bld(spec.NewSpec(NoopRuleName))

		// --- Then ---
		assert.ErrorIs(t, wErr, err)
		assert.Nil(t, have)
	})
}

func Test_mustTpl(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tplS := " {{.value}} "

		// --- When ---
		tpl := mustTpl("tpl-name", tplS)

		// --- Then ---
		data := map[string]string{spec.ArgValue: "abc"}
		w := &strings.Builder{}
		assert.NoError(t, tpl.Execute(w, data))
		assert.Equal(t, " abc ", w.String())
	})

	t.Run("error - panics", func(t *testing.T) {
		// --- When ---
		msg := assert.PanicMsg(t, func() { mustTpl("tpl-name", " {{.value} ") })

		// --- Then ---
		assert.Equal(t, "template: tpl-name:1: bad character U+007D '}'", *msg)
	})

	t.Run("missing key set to error", func(t *testing.T) {
		// --- When ---
		have := mustTpl("name", "custom tpl {{.not_supported}}")

		// --- Then ---
		buf := &bytes.Buffer{}
		err := have.Execute(buf, map[string]any{spec.ArgValue: 42})
		wRx := `^template: name:.*map has no entry for key "not_supported"$`
		assert.ErrorRegexp(t, wRx, err)
	})
}

func Test_tplText_tabular(t *testing.T) {
	tt := []struct {
		testN string

		txt  string
		want string
	}{
		{"plain", "abc", "abc"},
		{"with delimiter", "a{{b", `{{"a{{b"}}`},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := tplText(tc.txt)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_formatValue_tabular(t *testing.T) {
	tt := []struct {
		testN string

		value any
		want  any
	}{
		{"nil", nil, "nil"},
		{"int", 42, 42},
		{
			"time",
			time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC),
			"2000-01-02T03:04:05Z",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := formatValue(tc.value)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_renderTpl(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tpl := template.Must(template.New("name").Parse("tpl {{.value}}"))

		// --- When ---
		have, err := renderTpl(tpl, "abc", "prefix")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "tpl abc", have)
	})

	t.Run("template without a range", func(t *testing.T) {
		// --- Given ---
		tpl := template.Must(template.New("name").Parse("tpl"))

		// --- When ---
		have, err := renderTpl(tpl, "abc", "prefix")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "tpl", have)
	})

	t.Run("error - render", func(t *testing.T) {
		// --- Given ---
		tpl := template.Must(template.New("name").Parse("tpl {{.value}}"))

		// --- When ---
		have, err := renderTpl(tpl, func() {}, "prefix")

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "" +
			`prefix: template render error: template: name:1:6: executing ` +
			`"name" at <{{.value}}>: can't print {{.value}} of type func()`
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInternal, err)
		assert.NotNil(t, errors.Unwrap(err))
		assert.Empty(t, have)
	})
}

func Test_getArg(t *testing.T) {
	t.Run("error - key does not exist", func(t *testing.T) {
		// --- Given ---
		args := map[string]any{"name": uint(42)}

		// --- When ---
		have, err := getArg[uint](args, "other", "role_name")

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "role_name: spec missing required argument: other"
		assert.ErrorEqual(t, wMsg, err)
		assert.Equal(t, uint(0), have)
	})

	t.Run("success", func(t *testing.T) {
		// --- Given ---
		args := map[string]any{"name": uint(42)}

		// --- When ---
		have, err := getArg[uint](args, "name", "role_name")

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, uint(42), have)
	})

	t.Run("unnamed function for a named function type", func(t *testing.T) {
		// --- Given ---
		var called bool
		fn := func(any) error { called = true; return nil }
		args := map[string]any{"name": fn}

		// --- When ---
		have, err := getArg[RuleFunc](args, "name", "role_name")

		// --- Then ---
		assert.NoError(t, err)
		assert.NoError(t, have(1))
		assert.True(t, called)
	})

	t.Run("error - argument is of a wrong type", func(t *testing.T) {
		// --- Given ---
		args := map[string]any{"name": uint(42)}

		// --- When ---
		have, err := getArg[int](args, "name", "role_name")

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := `role_name: spec argument "name" must be int, got uint`
		assert.ErrorEqual(t, wMsg, err)
		assert.Equal(t, 0, have)
	})
}

func Test_errConvert(t *testing.T) {
	t.Run("internal error", func(t *testing.T) {
		// --- When ---
		err := errConvert("my-rule", 42, int64(0))

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "my-rule: cannot convert int to int64"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
	})

	t.Run("from and to types", func(t *testing.T) {
		// --- When ---
		err := errConvert("my-rule", "hello", 0.0)

		// --- Then ---
		wMsg := "my-rule: cannot convert string to float64"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("rule name", func(t *testing.T) {
		// --- When ---
		err := errConvert(RangeRuleName, true, int64(0))

		// --- Then ---
		assert.ErrorEqual(t, "range-rule: cannot convert bool to int64", err)
	})
}
