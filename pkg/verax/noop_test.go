// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Noop_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val any
	}{
		{"nil", nil},
		{"empty string", ""},
		{"zero value int", 0},
		{"pointer to string", pString},
		{"pointer to zero value string", pStringEmpty},
		{"pointer to int", pInt},
		{"pointer to zero value int", pIntZero},
		{"pointer to time", pTime},
		{"pointer to zero value time", pTimeZero},
		{"pointer to empty struct", pStructEmpty},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := Noop.Validate(tc.val)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}
