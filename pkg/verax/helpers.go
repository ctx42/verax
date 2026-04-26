// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/ctx42/verax/pkg/spec"
)

// ToAnySlice returns a slice of T as a slice of any.
func ToAnySlice[T any](vls ...T) []any {
	tmp := make([]any, 0, len(vls))
	for _, val := range vls {
		tmp = append(tmp, val)
	}
	return tmp
}

// AsRuleBuilder wraps a typed spec constructor and returns a [Builder] for it.
func AsRuleBuilder[T any](fn func(*spec.Spec) (T, error)) spec.Builder[Rule] {
	var t T
	if _, ok := any(t).(Rule); !ok {
		return nil
	}
	return func(spc *spec.Spec) (Rule, error) {
		t, err := fn(spc)
		if err != nil {
			return nil, err
		}
		r, _ := any(t).(Rule)
		return r, nil
	}
}

// hasGoSource reports whether any of the provided sources contains Go source
// code. It returns an error (referencing the rule name) if none of the sources
// is Go.
func hasGoSource(name string, srs ...*spec.Source) error {
	if len(srs) == 0 {
		msg := fmt.Sprintf("%s: incomplete rule", name)
		return NewInternalError(msg, ECInternal)
	}
	var hasGo bool
	for _, src := range srs {
		if src.Lang == "go" {
			hasGo = true
			break
		}
	}
	if hasGo {
		return nil
	}
	msg := fmt.Sprintf("%s: %s", name, spec.ErrNoGoSource)
	return NewInternalError(msg, spec.ECNoGoSource)
}

// mustTpl returns parsed text template with the given name. Panics on error.
func mustTpl(name, tpl string) *template.Template {
	return template.Must(template.New(name).
		Option("missingkey=error").
		Parse(tpl))
}

// formatValue formats some value types in the more readable way.
//
// - time.Time: RFC3339.
// - nil: "nil"
// - other: returned as is.
func formatValue(v any) any {
	if v == nil {
		return "nil"
	}
	if tim, ok := v.(time.Time); ok {
		return tim.Format(time.RFC3339)
	}
	return v
}

// renderTpl executes the template with the value. On error, returns
// [InternalError] prefixed with prefix.
func renderTpl(tpl *template.Template, val any, prefix string) (string, error) {
	buf := &bytes.Buffer{}
	data := map[string]any{spec.ArgValue: formatValue(val)}
	if err := tpl.Execute(buf, data); err != nil {
		msg := fmt.Sprintf("%s: template render error", prefix)
		return "", NewInternalError(msg, ECInternal)
	}
	return buf.String(), nil
}

// getArg retrieves an argument of type T from the args map.
//
// Returns [InternalError] if the argument does not exist or if the value is
// not of type T.
func getArg[T any](args map[string]any, key, rule string) (T, error) {
	var ok bool
	var anyVal any
	var retVal T

	if anyVal, ok = args[key]; !ok {
		msg := fmt.Sprintf("%s: spec missing required argument: %s", rule, key)
		return retVal, NewInternalError(msg, spec.ECInvSpec)
	}

	if retVal, ok = anyVal.(T); !ok {
		format := "%s: spec argument %q must be %T, got %T"
		msg := fmt.Sprintf(format, rule, key, retVal, anyVal)
		return retVal, NewInternalError(msg, spec.ECInvSpec)
	}
	return retVal, nil
}

// errConvert returns an [InternalError] with [ECInvType] code indicating that
// the value "from" cannot be converted to the type of "to". The name argument
// identifies the rule that triggered the error.
func errConvert(name string, from, to any) error {
	format := "%s: cannot convert %T to %T"
	msg := fmt.Sprintf(format, name, from, to)
	return NewInternalError(msg, ECInvType)
}
