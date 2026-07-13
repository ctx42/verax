// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package rule

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/verax"
)

func Test_IsSemVer_tabular(t *testing.T) {
	var tt = []struct {
		testN string

		param string
		want  bool
	}{
		{"v prefix", "v1.0.0", true},
		{"basic", "1.0.0", true},
		{"leading zero patch", "1.1.01", false},
		{"leading zero minor", "1.01.0", false},
		{"leading zero major", "01.1.0", false},
		{"v prefix leading zero patch", "v1.1.01", false},
		{"v prefix leading zero minor", "v1.01.0", false},
		{"v prefix leading zero major", "v01.1.0", false},
		{"prerelease", "1.0.0-alpha", true},
		{"prerelease dotted", "1.0.0-alpha.1", true},
		{"prerelease numeric", "1.0.0-0.3.7", true},
		{"prerelease leading zero", "1.0.0-0.03.7", false},
		{"prerelease double zero", "1.0.0-00.3.7", false},
		{"prerelease alphanumeric", "1.0.0-x.7.z.92", true},
		{"prerelease with build", "1.0.0-alpha+001", true},
		{"build metadata", "1.0.0+20130313144700", true},
		{"prerelease and build", "1.0.0-beta+exp.sha.5114f85", true},
		{
			"prerelease and build leading zero",
			"1.0.0-beta+exp.sha.05114f85",
			true,
		},
		{"empty prerelease", "1.0.0-+beta", false},
		{"invalid build chars", "1.0.0-b+-9+eta", false},
		{"invalid v prefix", "v+1.8.0-b+-9+eta", false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsSemVer(tc.param)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_SemVer(t *testing.T) {
	t.Run("error - validation", func(t *testing.T) {
		// --- When ---
		err := SemVer.Validate("1.0.0-+beta")

		// --- Then ---
		assert.ErrorEqual(t, "must be a valid semantic version", err)
		xrrtest.AssertCode(t, ECSemVer, err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- When ---
		err := SemVer.Validate(42)

		// --- Then ---
		assert.SameType(t, &verax.InternalError{}, err)
		wMsg := "must be a valid semantic version: expected string, got int"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, verax.ECInvType, err)
	})
}

func Test_SemVer_tabular(t *testing.T) {
	tt := []struct {
		testN string

		semVer string
	}{
		{"simple", "v1.0.0"},
		{"empty", ""},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := SemVer.Validate(tc.semVer)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}
