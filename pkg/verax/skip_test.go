// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Skip(t *testing.T) {
	// --- Given ---
	r := Skip

	// --- Then ---
	assert.True(t, bool(r))
}

func Test_skipRule_Validate_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val any
	}{
		{"nil", nil},
		{"int", 100},
		{"string", "str"},
		{"float", 1.1},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			assert.NoError(t, Skip.When(true).Validate(tc.val))
			assert.NoError(t, Skip.When(false).Validate(tc.val))
		})
	}
}

func Test_skipRule_When(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		// --- Given ---
		r := Skip.When(true)

		// --- Then ---
		assert.True(t, bool(r))
	})

	t.Run("false", func(t *testing.T) {
		// --- Given ---
		r := Skip.When(false)

		// --- Then ---
		assert.False(t, bool(r))
	})
}
