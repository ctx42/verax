// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/testcases"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_EnsureString_ok_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val any
		exp string
	}{
		{"string", "abc", "abc"},
		{"byte slice", []byte("abc"), "abc"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			got, err := EnsureString(tc.val)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.exp, got)
		})
	}
}

func Test_EnsureString_error_tabular(t *testing.T) {
	str := "abc"
	bytes := []byte("abc")

	tt := []struct {
		testN string

		val  any
		exp  string
		err  string
		code string
	}{
		{"int", 100, "", "must be either a string or byte slice", ECInvType},
		{
			"pointer to string",
			&str,
			"",
			"must be either a string or byte slice",
			ECInvType,
		},
		{
			"pointer to byte slice",
			&bytes,
			"",
			"must be either a string or byte slice",
			ECInvType,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			got, err := EnsureString(tc.val)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
			assert.Empty(t, got)
		})
	}
}

func Test_StringOrBytes_tabular(t *testing.T) {
	type MyString string

	str0 := "abc"
	var str1 string
	var str2 MyString = "abc"
	var str3 *string

	bytes := []byte("abc")
	var bytes2 []byte

	tt := []struct {
		testN string

		value    any
		str      string
		bs       []byte
		isString bool
		isBytes  bool
	}{
		{"string", str0, "abc", nil, true, false},
		{"pointer to string", &str0, "", nil, false, false},
		{"byte slice", bytes, "", []byte("abc"), false, true},
		{"pointer to byte slice", &bytes, "", nil, false, false},
		{"int", 100, "", nil, false, false},
		{"zero value string", str1, "", nil, true, false},
		{"pointer to zero value string", &str1, "", nil, false, false},
		{"nil byte slice", bytes2, "", nil, false, true},
		{"pointer to nil byte slice", &bytes2, "", nil, false, false},
		{"not empty string base type", str2, "abc", nil, true, false},
		{"pointer to not empty string base type", &str2, "", nil, false, false},
		{"nil pointer to zero value string", str3, "", nil, false, false},
	}
	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			isString, str, isBytes, bs := StringOrBytes(tc.value)

			// --- Then ---
			assert.Equal(t, tc.str, str)
			assert.Equal(t, tc.bs, bs)
			assert.Equal(t, tc.isString, isString)
			assert.Equal(t, tc.isBytes, isBytes)
		})
	}
}

func Test_LengthOfValue_ok_tabular(t *testing.T) {
	var a [3]int

	tt := []struct {
		testN string

		val any
		len int
	}{
		{"string", "abc", 3},
		{"slice", []int{1, 2}, 2},
		{"map", map[string]int{"A": 1, "B": 2}, 2},
		{"array", a, 3},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			got, err := LengthOfValue(tc.val)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.len, got)
		})
	}
}

func Test_LengthOfValue_error_tabular(t *testing.T) {
	var a [3]int

	tt := []struct {
		testN string

		val  any
		err  string
		code string
	}{
		{"pointer to array", &a, "cannot get the length of *[3]int", ECInvType},
		{"int", 123, "cannot get the length of int", ECInvType},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			got, err := LengthOfValue(tc.val)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
			assert.Equal(t, 0, got)
		})
	}
}

func Test_ToInt_ok_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val any
		exp int64
	}{
		{"int", 1, 1},
		{"int8", int8(1), 1},
		{"int16", int16(1), 1},
		{"int32", int32(1), 1},
		{"int64", int64(1), 1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			got, err := ToInt(tc.val)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.exp, got)
		})
	}
}

func Test_ToInt_error_tabular(t *testing.T) {
	var i32 int

	tt := []struct {
		testN string

		val  any
		err  string
		code string
	}{
		{"pointer to int", &i32, "cannot convert *int to int64", ECInvType},
		{"pointer to uint", uint(1), "cannot convert uint to int64", ECInvType},
		{"float64", float64(1), "cannot convert float64 to int64", ECInvType},
		{"string", "abc", "cannot convert string to int64", ECInvType},
		{"slice", []int{1, 2}, "cannot convert []int to int64", ECInvType},
		{
			"map",
			map[string]int{"A": 1},
			"cannot convert map[string]int to int64",
			ECInvType,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			got, err := ToInt(tc.val)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			assert.Equal(t, int64(0), got)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_ToUint_ok_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val any
		exp uint64
	}{
		{"uint", uint(1), 1},
		{"uint8", uint8(1), 1},
		{"uint16", uint16(1), 1},
		{"uint32", uint32(1), 1},
		{"uint64", uint64(1), 1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			got, err := ToUint(tc.val)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.exp, got)
		})
	}
}

func Test_ToUint_error_tabular(t *testing.T) {
	var i32 int
	var u32 uint

	tt := []struct {
		testN string

		val  any
		err  string
		code string
	}{
		{"int", 1, "cannot convert int to uint64", ECInvType},
		{"pointer to int", &i32, "cannot convert *int to uint64", ECInvType},
		{"pointer to uint", &u32, "cannot convert *uint to uint64", ECInvType},
		{"float64", float64(1), "cannot convert float64 to uint64", ECInvType},
		{"string", "abc", "cannot convert string to uint64", ECInvType},
		{"slice", []int{1, 2}, "cannot convert []int to uint64", ECInvType},
		{
			"map",
			map[string]int{"A": 1},
			"cannot convert map[string]int to uint64", ECInvType,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			got, err := ToUint(tc.val)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			assert.Equal(t, uint64(0), got)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}

func Test_ToFloat_ok_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val any
		exp float64
	}{
		{"float32", float32(1), 1},
		{"float64", float64(1), 1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			got, err := ToFloat(tc.val)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.exp, got)
		})
	}
}

func Test_ToFloat_error_tabular(t *testing.T) {
	var i32 int
	var u32 uint

	tt := []struct {
		testN string

		val  any
		err  string
		code string
	}{
		{"int", 1, "cannot convert int to float64", ECInvType},
		{"uint", uint(1), "cannot convert uint to float64", ECInvType},
		{"pointer to int", &i32, "cannot convert *int to float64", ECInvType},
		{"pointer to uint", &u32, "cannot convert *uint to float64", ECInvType},
		{"string", "abc", "cannot convert string to float64", ECInvType},
		{"slice", []int{1, 2}, "cannot convert []int to float64", ECInvType},
		{
			"map",
			map[string]int{"A": 1},
			"cannot convert map[string]int to float64",
			ECInvType,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			got, err := ToFloat(tc.val)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
			assert.Equal(t, 0.0, got)
		})
	}
}

func Test_IsEmpty_tabular(t *testing.T) {
	type MyString string

	var str0 string
	var str1 = "a"
	var str2 *string
	var str3 = ""

	s0 := struct{}{}
	tim0 := time.Now()
	var tim1 time.Time

	tt := []struct {
		testN string

		val   any
		empty bool
	}{
		// nil
		{"nil", nil, true},

		// string
		{"zero value string", "", true},
		{"string", "1", false},
		{"empty string base type", MyString(""), true},
		{"string base type", MyString("1"), false},

		// slice
		{"empty byte slice", []byte{}, true},
		{"byte slice", []byte("1"), false},
		{"empty byte slice from string", []byte(""), true},

		// map
		{"empty map", map[string]int{}, true},
		{"map", map[string]int{"a": 1}, false},

		// bool
		{"bool false is empty", false, true},
		{"bool true is not empty", true, false},

		// int
		{"int empty", 0, true},
		{"int8 empty", int8(0), true},
		{"int16 empty", int16(0), true},
		{"int32 empty", int32(0), true},
		{"int64 empty", int64(0), true},
		{"int not empty", 1, false},
		{"int8 not empty", int8(1), false},
		{"int16 not empty", int16(1), false},
		{"int32 not empty", int32(1), false},
		{"int64 not empty", int64(1), false},

		// uint
		{"uint empty", uint(0), true},
		{"uint8 empty", uint8(0), true},
		{"uint16 empty", uint16(0), true},
		{"uint32 empty", uint32(0), true},
		{"uint64 empty", uint64(0), true},
		{"uint not empty", uint(1), false},
		{"uint8 not empty", uint8(1), false},
		{"uint16 not empty", uint16(1), false},
		{"uint32 not empty", uint32(1), false},
		{"uint64 not empty", uint64(1), false},

		// float
		{"float32 empty", float32(0), true},
		{"float64 empty", float64(0), true},
		{"float32 not empty", float32(1), false},
		{"float64 not empty", float64(1), false},

		// interface, ptr
		{"pointer to empty string is empty", &str0, true},
		{"pointer to empty string is empty", &str3, true},
		{"pointer to not empty string is not empty", &str1, false},
		{"nil pointer to string is empty", str2, true},

		// struct
		{"empty struct is empty", s0, true},
		{"pointer to empty struct is empty", &s0, true},

		// time.Time
		{"time instance is not empty", tim0, false},
		{"pointer to time instance is not empty", &tim0, false},
		{"zero value time is empty", tim1, true},
		{"pointer to zero value time is empty\"", &tim1, true},

		// driver.Valuer
		{"valuer invalid", sql.NullInt64{Int64: 0, Valid: false}, true},
		{"valuer zero value", sql.NullInt64{Int64: 0, Valid: true}, true},
		{"valuer value", sql.NullInt64{Int64: 1, Valid: true}, false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsEmpty(tc.val)

			// --- Then ---
			assert.Equal(t, tc.empty, have)
		})
	}
}

func Test_IsNil_tabular_ZENValues(t *testing.T) {
	for _, tc := range testcases.ZENValues() {
		t.Run("Nil "+tc.Desc, func(t *testing.T) {
			// --- When ---
			hNil, hWrapped := IsNil(tc.Val)

			// --- Then ---
			assert.Equal(t, tc.IsNil, hNil)
			assert.Equal(t, tc.IsWrappedNil, hWrapped)
		})
	}
}

func Test_Indirect_tabular(t *testing.T) {
	var ptr0 *int
	var ptr1 *sql.NullInt64

	tt := []struct {
		testN string

		val    any
		result any
	}{
		{"nil pointer to int", ptr0, nil},
		{"nil pointer to struct", ptr1, nil},
		{"nil", nil, nil},
		{"int", 100, 100},
		{"invalid sql.NullInt64", sql.NullInt64{Int64: 0, Valid: false}, nil},
		{
			"invalid sql.NullInt64 with value",
			sql.NullInt64{Int64: 1, Valid: false},
			nil,
		},
		{
			"invalid pointer to sql.NullInt64",
			&sql.NullInt64{Int64: 0, Valid: false},
			nil,
		},
		{
			"invalid pointer to sql.NullInt64 with value",
			&sql.NullInt64{Int64: 1, Valid: false},
			nil,
		},
		{
			"valid sql.NullInt64 with zero value",
			sql.NullInt64{Int64: 0, Valid: true},
			int64(0),
		},
		{
			"valid sql.NullInt64",
			sql.NullInt64{Int64: 1, Valid: true},
			int64(1),
		},
		{
			"valid pointer to sql.NullInt64 with zero value",
			&sql.NullInt64{Int64: 0, Valid: true},
			int64(0),
		},
		{
			"valid pointer to sql.NullInt64",
			&sql.NullInt64{Int64: 1, Valid: true},
			int64(1),
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := Indirect(tc.val)

			// --- Then ---
			assert.Equal(t, tc.result, have)
		})
	}
}

func Test_mapErrKey_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val reflect.Value
		exp string
	}{
		{"int", reflect.ValueOf(1), "1"},
		{"uint", reflect.ValueOf(uint(2)), "2"},
		{"int8", reflect.ValueOf(int8(3)), "3"},
		{"uint8", reflect.ValueOf(uint8(4)), "4"},
		{"int16", reflect.ValueOf(int16(5)), "5"},
		{"uint16", reflect.ValueOf(uint16(6)), "6"},
		{"int32", reflect.ValueOf(int32(7)), "7"},
		{"uint32", reflect.ValueOf(uint32(8)), "8"},
		{"int64", reflect.ValueOf(int64(9)), "9"},
		{"uint64", reflect.ValueOf(uint64(10)), "10"},

		{"string", reflect.ValueOf("abc"), "abc"},
		{"nil pointer to string", reflect.ValueOf(pStringNil), ""},
		{"pointer to string", reflect.ValueOf(pString), "test string"},
		{"struct 1", reflect.ValueOf(TStruct{}), "<verax.TStruct Value>"},
		{
			"struct 2",
			reflect.ValueOf(ModelVal{"abc"}),
			"<verax.ModelVal Value>",
		},
		{
			"pointer to struct",
			reflect.ValueOf(&ModelPtr{"abc"}),
			"<verax.ModelPtr Value>",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			assert.Equal(t, tc.exp, mapErrKey(tc.val))
		})
	}
}

func Test_emtpl(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tplS := " {{.value}} "

		// --- When ---
		tpl := emtpl(tplS)

		// --- Then ---
		data := map[string]string{
			"value": "abc",
		}
		w := &strings.Builder{}
		assert.NoError(t, tpl.Execute(w, data))
		assert.Equal(t, " abc ", w.String())
	})

	t.Run("panics", func(t *testing.T) {
		assert.Panic(t, func() { emtpl(" {{.value} ") })
	})
}
