package rule

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/verax"
)

func Test_IsSemver_tabular(t *testing.T) {
	var tt = []struct {
		testN string

		param string
		exp   bool
	}{
		{"1", "v1.0.0", true},
		{"2", "1.0.0", true},
		{"3", "1.1.01", false},
		{"4", "1.01.0", false},
		{"5", "01.1.0", false},
		{"6", "v1.1.01", false},
		{"7", "v1.01.0", false},
		{"8", "v01.1.0", false},
		{"9", "1.0.0-alpha", true},
		{"10", "1.0.0-alpha.1", true},
		{"11", "1.0.0-0.3.7", true},
		{"12", "1.0.0-0.03.7", false},
		{"13", "1.0.0-00.3.7", false},
		{"14", "1.0.0-x.7.z.92", true},
		{"15", "1.0.0-alpha+001", true},
		{"16", "1.0.0+20130313144700", true},
		{"17", "1.0.0-beta+exp.sha.5114f85", true},
		{"18", "1.0.0-beta+exp.sha.05114f85", true},
		{"19", "1.0.0-+beta", false},
		{"20", "1.0.0-b+-9+eta", false},
		{"21", "v+1.8.0-b+-9+eta", false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsSemver(tc.param)

			// --- Then ---
			assert.Equal(t, tc.exp, have)
		})
	}
}

func Test_SemVer_tabular(t *testing.T) {
	tt := []struct {
		testN string

		semVer string
	}{
		{"1", "v1.0.0"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := verax.Validate(tc.semVer, SemVer)

			// --- Then ---
			assert.NoError(t, err)
		})
	}
}

func Test_SemVer_errors_tabular(t *testing.T) {
	tt := []struct {
		testN string

		semVer string
		err    string
		code   string
	}{
		{"1", "1.0.0-+beta", "must be a valid semantic version", "ECSemVer"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			err := verax.Validate(tc.semVer, SemVer)

			// --- Then ---
			assert.ErrorEqual(t, tc.err, err)
			xrrtest.AssertCode(t, tc.code, err)
		})
	}
}
