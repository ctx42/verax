// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"errors"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_EachRule_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val   any
		rules []Rule
	}{
		{
			"slice empty",
			[]string{},
			[]Rule{},
		},
		{
			"slice with values",
			[]string{"abc", "def"},
			[]Rule{Required},
		},
		{
			"slice with validators",
			[]ModelVal{{"abc"}, {"abc"}},
			[]Rule{Required},
		},
		{
			"map empty",
			map[string]string{},
			[]Rule{},
		},
		{
			"map with keys",
			map[string]string{"key0": "val0", "key1": "val1"},
			[]Rule{Required},
		},
		{
			"map with validators keys",
			map[string]ModelVal{"key0": {"abc"}, "key1": {"abc"}},
			[]Rule{Required},
		},
		{
			"array empty",
			[...]string{},
			[]Rule{},
		},
		{
			"array with values",
			[...]string{"abc", "def"},
			[]Rule{Required},
		},
		{
			"array with validators",
			[...]ModelVal{{"abc"}, {"abc"}},
			[]Rule{Required},
		},
		{
			"channel instances",
			[]chan int{iChan},
			[]Rule{Required},
		},
		{
			"function instances",
			[]func(any) bool{iFunc},
			[]Rule{Required},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Each(tc.rules...).Validate(tc.val)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_EachRule_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val   any
		rules []Rule
		err   string
	}{
		{
			"not iterable",
			nil,
			[]Rule{},
			"must be an iterable (ECInvType)",
		},
		{
			"slice with values",
			[]string{"def", ""},
			[]Rule{Required},
			"1: cannot be blank (ECRequired)",
		},
		{
			"slice with nils",
			[]*string{pString, pStringNil},
			[]Rule{Required},
			"1: cannot be blank (ECRequired)",
		},
		{
			"slice with validators",
			[]ModelVal{{"abc"}, {"def"}},
			[]Rule{Required},
			"1.FStr: must be 'abc' (ECMustAbc)",
		},
		{
			"map with keys",
			map[string]string{"key0": "val0", "key1": ""},
			[]Rule{Required},
			"key1: cannot be blank (ECRequired)",
		},
		{
			"map with nils",
			map[string]*string{"key0": pString, "key1": pStringNil},
			[]Rule{Required},
			"key1: cannot be blank (ECRequired)",
		},
		{
			"map with validators keys",
			map[string]ModelVal{"key0": {"abc"}, "key1": {"def"}},
			[]Rule{Required},
			"key1.FStr: must be 'abc' (ECMustAbc)",
		},
		{
			"array with values",
			[...]string{"abc", ""},
			[]Rule{Required},
			"1: cannot be blank (ECRequired)",
		},
		{
			"array with validators",
			[...]ModelVal{{"abc"}, {"def"}},
			[]Rule{Required},
			"1.FStr: must be 'abc' (ECMustAbc)",
		},
		{
			"array with nils",
			[...]*string{pString, nil},
			[]Rule{Required},
			"1: cannot be blank (ECRequired)",
		},
		{
			"channel declared",
			[]any{dChan},
			[]Rule{Required},
			"0: cannot be blank (ECRequired)",
		},
		{
			"function declared",
			[]any{dFunc},
			[]Rule{Required},
			"0: cannot be blank (ECRequired)",
		},
		{
			"pointers",
			[]*int{pIntNil, nil},
			[]Rule{Required},
			"0: cannot be blank (ECRequired); 1: cannot be blank (ECRequired)",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Each(tc.rules...).Validate(tc.val)

			// --- Then ---
			xrrtest.AssertEqual(t, tc.err, err)
		})
	}
}

func Test_Each(t *testing.T) {
	t.Run("slice of pointers", func(t *testing.T) {
		// --- Given ---
		var s []*TStruct
		ts0 := NewTStruct()
		ts1 := NewTStruct()
		ts1.FStr = "wrong"
		s = append(s, &ts0, &ts1)

		fn := func(v any) error {
			sp := v.(*TStruct) // The test checks we get pointer here.
			if sp.FStr != "FStr" {
				return errors.New("error")
			}
			return nil
		}

		// --- When ---
		err := Each(By(fn)).Validate(s)

		// --- Then ---
		xrrtest.AssertEqual(t, "1: error (ECGeneric)", err)
	})

	t.Run("slice of values", func(t *testing.T) {
		// --- Given ---
		var s []TStruct
		ts0 := NewTStruct()
		ts1 := NewTStruct()
		ts1.FStr = "wrong"
		s = append(s, ts0, ts1)

		fn := func(v any) error {
			sp := v.(TStruct) // The test checks we get value here.
			if sp.FStr != "FStr" {
				return errors.New("error")
			}
			return nil
		}

		// --- When ---
		err := Each(By(fn)).Validate(s)

		// --- Then ---
		xrrtest.AssertEqual(t, "1: error (ECGeneric)", err)
	})
}
