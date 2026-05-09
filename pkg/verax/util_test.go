// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/testcases"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

// tErrValuer implements driver.Valuer but always returns an error,
// so isEmptyValue falls through to the reflect-based branch.
type tErrValuer struct{ n int }

func (tErrValuer) Value() (driver.Value, error) { return nil, errors.New("forced error") }

func Test_EnsureString_ok_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val  any
		want string
	}{
		{"string", "abc", "abc"},
		{"byte slice", []byte("abc"), "abc"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have, err := EnsureString(tc.val)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_EnsureString_error_tabular(t *testing.T) {
	str := "abc"
	bytes := []byte("abc")

	tt := []struct {
		testN string

		val  any
		want string
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
			have, err := EnsureString(tc.val)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
			assert.Empty(t, have)
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
			have, err := LengthOfValue(tc.val)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.len, have)
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
			have, err := LengthOfValue(tc.val)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
			assert.Equal(t, 0, have)
		})
	}
}

func Test_IsEmpty_tabular_ZENValues(t *testing.T) {
	for _, tc := range testcases.ZENValues() {
		if tc.Desc == "false boolean" {
			continue // In verax we treat false as not empty.
		}
		t.Run("IsEmpty "+tc.Desc, func(t *testing.T) {
			// --- When ---
			have := IsEmpty(tc.Val)

			// --- Then ---
			assert.Equal(t, tc.IsEmpty, have)
		})

		t.Run("checkNilAndEmpty IsEmpty "+tc.Desc, func(t *testing.T) {
			// --- When ---
			_, have := checkNilAndEmpty(tc.Val)

			// --- Then ---
			assert.Equal(t, tc.IsEmpty, have)
		})
	}
}

func Test_IsEmpty_tabular(t *testing.T) {
	type MyString string

	var str0 string
	var str1 = "a"
	var str2 *string
	var str3 = ""
	var nilChan chan int
	emptyChan := make(chan int)
	nonEmptyChan := make(chan int, 1)
	nonEmptyChan <- 1

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
		{"bool false is not empty", false, false},
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

		// chan
		{"chan empty", emptyChan, true},
		{"chan not empty", nonEmptyChan, false},
		{"chan nil", nilChan, true},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsEmpty(tc.val)

			// --- Then ---
			assert.Equal(t, tc.empty, have)
		})

		t.Run("checkNilAndEmpty "+tc.testN, func(t *testing.T) {
			// --- When ---
			_, have := checkNilAndEmpty(tc.val)

			// --- Then ---
			assert.Equal(t, tc.empty, have)
		})
	}
}

func Test_isEmptyValue_Valuer_error(t *testing.T) {
	t.Run("falls through to reflect branch when Value returns error", func(t *testing.T) {
		// tErrValuer.Value() returns an error, so isEmptyValue skips the
		// Valuer branch and falls through to reflect; n=1 makes it non-zero.

		// --- When ---
		have := isEmptyValue(tErrValuer{n: 1})

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_IsNil_tabular_ZENValues(t *testing.T) {
	for _, tc := range testcases.ZENValues() {
		t.Run("Nil "+tc.Desc, func(t *testing.T) {
			// --- When ---
			have := IsNil(tc.Val)

			// --- Then ---
			assert.Equal(t, tc.IsNil, have)
		})

		t.Run("checkNilAndEmpty Nil "+tc.Desc, func(t *testing.T) {
			// --- When ---
			have, _ := checkNilAndEmpty(tc.Val)

			// --- Then ---
			assert.Equal(t, tc.IsNil, have)
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

		val  reflect.Value
		want string
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
			assert.Equal(t, tc.want, mapErrKey(tc.val))
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
