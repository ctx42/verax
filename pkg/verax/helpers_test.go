// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/spec"
)

func Test_ToAnySlice(t *testing.T) {
	assert.Equal(t, []any{}, ToAnySlice[int]())
	assert.Equal(t, []any{}, ToAnySlice[string]())
	assert.Equal(t, []any{"a", "b", "c"}, ToAnySlice("a", "b", "c"))
	assert.Equal(t, []any{1, 2, 3}, ToAnySlice(1, 2, 3))
}

func Test_AsRuleBuilder(t *testing.T) {
	t.Run("returns nil when T does not implement Rule", func(t *testing.T) {
		// --- When ---
		have := AsRuleBuilder(func(_ *spec.Spec) (int, error) { return 0, nil })

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("wraps the typed constructor", func(t *testing.T) {
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

	t.Run("propagates constructor error", func(t *testing.T) {
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

func Test_hasGoSource(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		srcJs := &spec.Source{Name: "name_js", Lang: "js"}
		srcGo := &spec.Source{Name: "name_go", Lang: "go"}

		// --- When ---
		err := hasGoSource("name", srcJs, srcGo)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - empty list", func(t *testing.T) {
		// --- When ---
		err := hasGoSource("name")

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		xrrtest.AssertCode(t, ECInternal, err)
		assert.ErrorContain(t, "name: ", err)
	})

	t.Run("error - no Go source", func(t *testing.T) {
		// --- When ---
		err := hasGoSource("name", &spec.Source{Name: "name", Lang: "js"})

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		xrrtest.AssertCode(t, spec.ECNoGoSource, err)
		assert.ErrorContain(t, "name: ", err)
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

	t.Run("panics", func(t *testing.T) {
		// --- When ---
		msg := assert.PanicMsg(t, func() { mustTpl("tpl-name", " {{.value} ") })

		// --- Then ---
		wMsg := "template: tpl-name:1: bad character U+007D '}'"
		assert.Equal(t, wMsg, *msg)
	})

	t.Run("missing key set to error", func(t *testing.T) {
		// --- When ---
		have := mustTpl("name", "custom tpl {{.not_supported}}")

		// --- Then ---
		buf := &bytes.Buffer{}
		err := have.Execute(buf, map[string]any{spec.ArgValue: 42})
		assert.ErrorContain(t, "template: name:", err)
		assert.ErrorContain(t, `map has no entry for key "not_supported"`, err)
	})
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
		assert.ErrorEqual(t, "prefix: template render error", err)
		xrrtest.AssertCode(t, ECInternal, err)
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
	t.Run("returns an internal error with ECInvType code", func(t *testing.T) {
		// --- When ---
		err := errConvert("my-rule", 42, int64(0))

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		wMsg := "my-rule: cannot convert int to int64"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, ECInvType, err)
	})

	t.Run("message contains from and to types", func(t *testing.T) {
		// --- When ---
		err := errConvert("my-rule", "hello", 0.0)

		// --- Then ---
		wMsg := "my-rule: cannot convert string to float64"
		assert.ErrorEqual(t, wMsg, err)
	})

	t.Run("message contains rule name", func(t *testing.T) {
		// --- When ---
		err := errConvert(RangeRuleName, true, int64(0))

		// --- Then ---
		wMsg := "range-rule: cannot convert bool to int64"
		assert.ErrorEqual(t, wMsg, err)
	})
}
