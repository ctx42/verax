// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/ctx42/xrr/pkg/xrr"

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

// convertTo returns val as T. It dereferences val with [Indirect] and converts
// values whose type is convertible to T and has the same kind as T. It
// returns false when val cannot be represented as T.
func convertTo[T any](val any) (T, bool) {
	if vt, ok := val.(T); ok {
		return vt, true
	}
	var zero T
	rv := reflect.ValueOf(Indirect(val))
	typ := reflect.TypeFor[T]()
	if !rv.IsValid() || rv.Kind() != typ.Kind() || !rv.CanConvert(typ) {
		return zero, false
	}
	vt, ok := rv.Convert(typ).Interface().(T)
	return vt, ok
}

// AsRuleBuilder wraps a typed spec constructor and returns a [spec.Builder]
// for it. It returns nil when T does not implement [Rule].
func AsRuleBuilder[T any](fn func(*spec.Spec) (T, error)) spec.Builder[Rule] {
	if !reflect.TypeFor[T]().Implements(reflect.TypeFor[Rule]()) {
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

// mustTpl returns a parsed text template with the given name. Panics on error.
func mustTpl(name, tpl string) *template.Template {
	return template.Must(template.New(name).
		Option("missingkey=error").
		Parse(tpl))
}

// tplText returns txt as text/template source rendering txt verbatim. Text
// containing an action delimiter is emitted as a quoted string action.
func tplText(txt string) string {
	if !strings.Contains(txt, "{{") {
		return txt
	}
	return fmt.Sprintf("{{%q}}", txt)
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
		return "", NewInternalErrorf(
			"%s: template render error",
			prefix,
			xrr.WithCode(ECInternal),
			xrr.WithCause(err),
		)
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
		return retVal, NewInternalErrorf(
			"%s: spec missing required argument: %s",
			rule,
			key,
			xrr.WithCode(spec.ECInvSpec),
		)
	}

	if retVal, ok = anyVal.(T); ok {
		return retVal, nil
	}
	// Accept a function of an unnamed type for a named function type.
	rv, typ := reflect.ValueOf(anyVal), reflect.TypeFor[T]()
	if rv.Kind() == reflect.Func && rv.Type().ConvertibleTo(typ) {
		if retVal, ok = rv.Convert(typ).Interface().(T); ok {
			return retVal, nil
		}
	}
	return retVal, NewInternalErrorf(
		"%s: spec argument %q must be %T, got %T",
		rule,
		key,
		retVal,
		anyVal,
		xrr.WithCode(spec.ECInvSpec),
	)
}

// errConvert returns an [InternalError] with [ECInvType] code indicating that
// the value "from" cannot be converted to the type of "to". The name argument
// identifies the rule that triggered the error.
func errConvert(name string, from, to any) error {
	return NewInternalErrorf(
		"%s: cannot convert %T to %T",
		name,
		from,
		to,
		xrr.WithCode(ECInvType),
	)
}
