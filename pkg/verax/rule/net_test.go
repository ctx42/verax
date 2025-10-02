package rule

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"

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
		err := verax.Validate("1.2.3.4", IP)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("", IP)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("256.0.0.0", IP)

		// --- Then ---
		assert.ErrorIs(t, ErrIP, err)
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
		err := verax.Validate("1.2.3.4", IPv4)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("", IPv4)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("256.0.0.0", IPv4)

		// --- Then ---
		assert.ErrorIs(t, ErrIPv4, err)
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
		err := verax.Validate("1ce:c01d:bee2:15:a5:900d:a5:11fe", IPv6)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("", IPv6)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("256.0.0.0", IPv6)

		// --- Then ---
		assert.ErrorIs(t, ErrIPv6, err)
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
		err := verax.Validate("42", Port)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("", Port)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("abc", Port)

		// --- Then ---
		assert.ErrorIs(t, ErrPort, err)
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
		{"1", "a.bc", true},
		{"2", "a.b.", true},
		{"3", "a.b..", false},
		{"4", "localhost.local", true},
		{"5", "localhost.localdomain.intern", true},
		{"6", "l.local.intern", true},
		{"7", "ru.link.n.svpncloud.com", true},
		{"8", "-localhost", false},
		{"9", "localhost.-localdomain", false},
		{"10", "localhost.localdomain.-int", false},
		{"11", "_localhost", true},
		{"12", "localhost._localdomain", true},
		{"13", "localhost.localdomain._int", true},
		{"14", "lÖcalhost", false},
		{"15", "localhost.lÖcaldomain", false},
		{"16", "localhost.localdomain.üntern", false},
		{"17", "__", true},
		{"18", "localhost/", false},
		{"19", "127.0.0.1", false},
		{"20", "[::1]", false},
		{"21", "50.50.50.50", false},
		{"22", "localhost.localdomain.intern:65535", false},
		{"23", "漢字汉字", false},
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
		err := verax.Validate("localhost", DNSName)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("", DNSName)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("localhost/", DNSName)

		// --- Then ---
		assert.ErrorIs(t, ErrDNSName, err)
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
		{"1", "a.bc", true},
		{"2", "a.b.", false},
		{"3", "a.b..", false},
		{"4", "localhost.local", true},
		{"5", "localhost.localdomain.intern", true},
		{"6", "l.local.intern", true},
		{"7", "ru.link.n.svpncloud.com", true},
		{"8", "-localhost", false},
		{"9", "localhost.-localdomain", false},
		{"10", "localhost.localdomain.-int", false},
		{"11", "_localhost", false},
		{"12", "localhost._localdomain", false},
		{"13", "localhost.localdomain._int", false},
		{"14", "lÖcalhost", false},
		{"15", "localhost.lÖcaldomain", false},
		{"16", "localhost.localdomain.üntern", false},
		{"17", "__", false},
		{"18", "localhost/", false},
		{"19", "127.0.0.1", false},
		{"20", "[::1]", false},
		{"21", "50.50.50.50", false},
		{"22", "localhost.localdomain.intern:65535", false},
		{"23", "漢字汉字", false},
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
		err := verax.Validate("example.com", Domain)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("", Domain)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("a.b..", Domain)

		// --- Then ---
		assert.ErrorIs(t, ErrDomain, err)
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
		{"domain invalid 1", "-localhost", false},
		{"domain invalid 2", ".localhost", false},
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
		err := verax.Validate("localhost", Host)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("success when empty", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("", Host)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		// --- When ---
		err := verax.Validate("localhost/", Host)

		// --- Then ---
		assert.ErrorIs(t, ErrHost, err)
	})
}
