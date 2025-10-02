// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_ContainRule_Validate_valid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val  any
		rule EqualRule
	}{
		{"slice of int", []int{1, 2, 3}, Equal(2)},
		{"slice of string", []string{"a", "b", "c"}, Equal("c")},
		{"array of int", [...]int{1, 2, 3}, Equal(2)},
		{"array of string", [...]string{"a", "b", "c"}, Equal("c")},

		{"map string:int", map[string]int{"A": 1, "B": 2, "C": 3}, Equal(2)},
		{"map int:string", map[int]string{1: "A", 2: "B", 3: "C"}, Equal("C")},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Contain(tc.rule).Validate(tc.val)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_ContainRule_Validate_invalid_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val  any
		rule EqualRule
		err  string
		code string
	}{
		{
			"slice does not contain",
			[]int{1, 2, 3},
			Equal(0),
			"must contain at least one '0' value",
			ECNotEqual,
		},
		{
			"empty slice",
			[]int{},
			Equal(0),
			"must contain at least one '0' value",
			ECNotEqual,
		},
		{
			"nil slice",
			[]int(nil),
			Equal(0),
			"must contain at least one '0' value",
			ECNotEqual,
		},
		{
			"array does not contain",
			[...]int{1, 2, 3},
			Equal(4),
			"must contain at least one '4' value",
			ECNotEqual,
		},
		{
			"empty array",
			[...]int{},
			Equal(4),
			"must contain at least one '4' value",
			ECNotEqual,
		},
		{
			"map does not contain",
			map[string]int{"A": 1, "B": 2, "C": 3},
			Equal("D"),
			"must contain at least one 'D' value",
			ECNotEqual,
		},
		{
			"empty map",
			map[string]int{},
			Equal("D"),
			"must contain at least one 'D' value", ECNotEqual,
		},
		{
			"nil map",
			map[string]int(nil),
			Equal("D"),
			"must contain at least one 'D' value",
			ECNotEqual,
		},
		{
			"must be iterable",
			"ABC",
			Equal("C"),
			"must be an iterable",
			ECInvType,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Contain(tc.rule).Validate(tc.val)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}
