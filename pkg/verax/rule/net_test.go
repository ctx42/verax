// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package rule

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"

	"github.com/ctx42/verax/pkg/verax"
)

func Test_IsIP_tabular(t *testing.T) {
	tt := []struct {
		testN string

		ip   string
		want bool
	}{
		{"empty", "", false},
		{"IPv4", "1.2.3.4", true},
		{"IPv4 loopback", "127.0.0.1", true},
		{"IPv4 unspecified", "0.0.0.0", true},
		{"IPv4 broadcast", "255.255.255.255", true},
		{"IPv4 invalid", "256.0.0.0", false},
		{"IPv6 loopback", "::1", true},
		{"IPv6", "1ce:c01d:bee2:15:a5:900d:a5:11fe", true},
		{"IPv6 invalid", ":::1", false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsIP(tc.ip)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_IP(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		err := IP.Validate("1.2.3.4")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := IP.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - validation", func(t *testing.T) {
		// --- When ---
		err := IP.Validate("256.0.0.0")

		// --- Then ---
		assert.ErrorEqual(t, msgIP, err)
		xrrtest.AssertCode(t, ECIP, err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- When ---
		err := IP.Validate(42)

		// --- Then ---
		assert.SameType(t, &verax.InternalError{}, err)
		wMsg := "must be a valid IP address: expected string, got int"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, verax.ECInvType, err)
	})
}

func Test_IsIPv4_tabular(t *testing.T) {
	tt := []struct {
		testN string

		ip   string
		want bool
	}{
		{"empty", "", false},
		{"IPv4", "1.2.3.4", true},
		{"IPv4 loopback", "127.0.0.1", true},
		{"IPv4 unspecified", "0.0.0.0", true},
		{"IPv4 broadcast", "255.255.255.255", true},
		{"IPv4 invalid", "256.0.0.0", false},
		{"IPv6 loopback", "::1", false},
		{"IPv6", "1ce:c01d:bee2:15:a5:900d:a5:11fe", false},
		{"IPv6 invalid", ":::1", false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsIPv4(tc.ip)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_IPv4(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		err := IPv4.Validate("1.2.3.4")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := IPv4.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - validation", func(t *testing.T) {
		// --- When ---
		err := IPv4.Validate("256.0.0.0")

		// --- Then ---
		assert.ErrorEqual(t, msgIPv4, err)
		xrrtest.AssertCode(t, ECIPv4, err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- When ---
		err := IPv4.Validate(42)

		// --- Then ---
		assert.SameType(t, &verax.InternalError{}, err)
		wMsg := "must be a valid IPv4 address: expected string, got int"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, verax.ECInvType, err)
	})
}

func Test_IsIPv6_tabular(t *testing.T) {
	tt := []struct {
		testN string

		ip   string
		want bool
	}{
		{"empty", "", false},
		{"IPv4", "1.2.3.4", false},
		{"IPv4 loopback", "127.0.0.1", false},
		{"IPv4 unspecified", "0.0.0.0", false},
		{"IPv4 broadcast", "255.255.255.255", false},
		{"IPv4 invalid", "256.0.0.0", false},
		{"IPv6 loopback", "::1", true},
		{"IPv6", "1ce:c01d:bee2:15:a5:900d:a5:11fe", true},
		{"IPv6 invalid", ":::1", false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsIPv6(tc.ip)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_IPv6(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		err := IPv6.Validate("1ce:c01d:bee2:15:a5:900d:a5:11fe")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := IPv6.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - validation", func(t *testing.T) {
		// --- When ---
		err := IPv6.Validate("256.0.0.0")

		// --- Then ---
		assert.ErrorEqual(t, msgIPv6, err)
		xrrtest.AssertCode(t, ECIPv6, err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- When ---
		err := IPv6.Validate(42)

		// --- Then ---
		assert.SameType(t, &verax.InternalError{}, err)
		wMsg := "must be a valid IPv6 address: expected string, got int"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, verax.ECInvType, err)
	})
}

func Test_IsPort_tabular(t *testing.T) {
	tt := []struct {
		testN string

		port string
		want bool
	}{
		{"invalid empty", "", false},
		{"invalid negative", "-1", false},
		{"invalid zero", "0", false},
		{"first valid", "1", true},
		{"last valid", "65535", true},
		{"invalid too big", "65536", false},
		{"invalid value", "abc", false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsPort(tc.port)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_Port(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		err := Port.Validate("42")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := Port.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - validation", func(t *testing.T) {
		// --- When ---
		err := Port.Validate("abc")

		// --- Then ---
		assert.ErrorEqual(t, msgPort, err)
		xrrtest.AssertCode(t, ECPort, err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- When ---
		err := Port.Validate(42)

		// --- Then ---
		assert.SameType(t, &verax.InternalError{}, err)
		wMsg := "must be a valid network port: expected string, got int"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, verax.ECInvType, err)
	})
}

func Test_IsDNSName_tabular(t *testing.T) {
	tt := []struct {
		testN string

		name string
		want bool
	}{
		{"empty", "", false},
		{"localhost", "localhost", true},
		{"two labels", "a.bc", true},
		{"trailing dot", "a.b.", true},
		{"double trailing dot", "a.b..", false},
		{"two word labels", "localhost.local", true},
		{"three labels", "localhost.localdomain.intern", true},
		{"single-char first label", "l.local.intern", true},
		{"five labels", "ru.link.n.svpncloud.com", true},
		{"leading hyphen", "-localhost", false},
		{"label starts with hyphen", "localhost.-localdomain", false},
		{
			"the last label starts with hyphen",
			"localhost.localdomain.-int",
			false,
		},
		{"leading underscore", "_localhost", true},
		{"label starts with underscore", "localhost._localdomain", true},
		{
			"the last label starts with an underscore",
			"localhost.localdomain._int",
			true,
		},
		{"non-ascii first label", "lÖcalhost", false},
		{"non-ascii middle label", "localhost.lÖcaldomain", false},
		{"non-ascii last label", "localhost.localdomain.üntern", false},
		{"only underscores", "__", true},
		{"trailing slash", "localhost/", false},
		{"ipv4 address", "127.0.0.1", false},
		{"bracketed ipv6", "[::1]", false},
		{"all numeric labels", "50.50.50.50", false},
		{"with port", "localhost.localdomain.intern:65535", false},
		{"cjk characters", "漢字汉字", false},
		{
			"too long",
			"www.jubfvq1v3p38i51622y0dvmdk1mymowjyeu26gbtw9andgyn" +
				"j1gg8z3msb1kl5z6906k846pj3sulm4kiyk82ln5teqj9nsh" +
				"t59opr0cs5ssltx78lfyvml19lfq1wp4usbl0o36cmiykch1" +
				"vywbttcus1p9yu0669h8fj4ll7a6bmop505908s1m83q2ec2" +
				"qr9nbvql2589adma3xsq2o38os2z3dmfh2tth4is4ixyfasa" +
				"sasefqwe4t2ub2fz1rme.de",
			false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsDNSName(tc.name)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_DNSName(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		err := DNSName.Validate("localhost")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := DNSName.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - validation", func(t *testing.T) {
		// --- When ---
		err := DNSName.Validate("localhost/")

		// --- Then ---
		assert.ErrorEqual(t, msgDNSName, err)
		xrrtest.AssertCode(t, ECDNSName, err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- When ---
		err := DNSName.Validate(42)

		// --- Then ---
		assert.SameType(t, &verax.InternalError{}, err)
		wMsg := "must be a valid DNS name: expected string, got int"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, verax.ECInvType, err)
	})
}

func Test_IsDomain_tabular(t *testing.T) {
	tt := []struct {
		testN string

		name string
		want bool
	}{
		{"empty", "", false},
		{"localhost", "localhost", false},
		{"two labels", "a.bc", true},
		{"trailing dot", "a.b.", false},
		{"double trailing dot", "a.b..", false},
		{"two word labels", "localhost.local", true},
		{"three labels", "localhost.localdomain.intern", true},
		{"single-char first label", "l.local.intern", true},
		{"five labels", "ru.link.n.svpncloud.com", true},
		{"leading hyphen", "-localhost", false},
		{"label starts with hyphen", "localhost.-localdomain", false},
		{"last label starts with hyphen", "localhost.localdomain.-int", false},
		{"leading underscore", "_localhost", false},
		{"label starts with underscore", "localhost._localdomain", false},
		{"last label starts with underscore", "localhost.localdomain._int", false},
		{"non-ascii first label", "lÖcalhost", false},
		{"non-ascii middle label", "localhost.lÖcaldomain", false},
		{"non-ascii last label", "localhost.localdomain.üntern", false},
		{"only underscores", "__", false},
		{"trailing slash", "localhost/", false},
		{"ipv4 address", "127.0.0.1", false},
		{"bracketed ipv6", "[::1]", false},
		{"all numeric labels", "50.50.50.50", false},
		{"with port", "localhost.localdomain.intern:65535", false},
		{"cjk characters", "漢字汉字", false},
		{"punycode TLD", "example.xn--p1ai", true},
		{
			"too long",
			"www.jubfvq1v3p38i51622y0dvmdk1mymowjyeu26gbtw9andgyn" +
				"j1gg8z3msb1kl5z6906k846pj3sulm4kiyk82ln5teqj9nsh" +
				"t59opr0cs5ssltx78lfyvml19lfq1wp4usbl0o36cmiykch1" +
				"vywbttcus1p9yu0669h8fj4ll7a6bmop505908s1m83q2ec2" +
				"qr9nbvql2589adma3xsq2o38os2z3dmfh2tth4is4ixyfasa" +
				"sasefqwe4t2ub2fz1rme.de",
			false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsDomain(tc.name)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_Domain(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		err := Domain.Validate("example.com")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := Domain.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - validation", func(t *testing.T) {
		// --- When ---
		err := Domain.Validate("a.b..")

		// --- Then ---
		assert.ErrorEqual(t, msgDomain, err)
		xrrtest.AssertCode(t, ECDomain, err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- When ---
		err := Domain.Validate(42)

		// --- Then ---
		assert.SameType(t, &verax.InternalError{}, err)
		wMsg := "must be a valid domain: expected string, got int"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, verax.ECInvType, err)
	})
}

func Test_IsHost_tabular(t *testing.T) {
	tt := []struct {
		testN string

		host string
		want bool
	}{
		{"empty", "", false},
		{"local", "localhost", true},
		{"loopback hostname", "localhost.localdomain", true},
		{"IPv6", "1ce:c01d:bee2:15:a5:900d:a5:11fe", true},
		{"IPv6 loopback", "::1", true},
		{"IPv6 invalid", "-[::1]", false},
		{"domain", "example.com", true},
		{"domain with port", "localhost.localdomain:65535", false},
		{"domain leading hyphen", "-localhost", false},
		{"domain leading dot", ".localhost", false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := IsHost(tc.host)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_Host(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		err := Host.Validate("localhost")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := Host.Validate("")

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error - validation", func(t *testing.T) {
		// --- When ---
		err := Host.Validate("localhost/")

		// --- Then ---
		assert.ErrorEqual(t, msgHost, err)
		xrrtest.AssertCode(t, ECHost, err)
	})

	t.Run("error - invalid type", func(t *testing.T) {
		// --- When ---
		err := Host.Validate(42)

		// --- Then ---
		assert.SameType(t, &verax.InternalError{}, err)
		wMsg := "must be a valid network hostname: expected string, got int"
		assert.ErrorEqual(t, wMsg, err)
		xrrtest.AssertCode(t, verax.ECInvType, err)
	})
}
