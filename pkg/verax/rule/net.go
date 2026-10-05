// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package rule

import (
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/ctx42/verax/pkg/verax"
)

// Regexp rules.
const (
	// dnsNameRx matches DNS names; labels may contain underscores.
	dnsNameRx string = `^[a-zA-Z0-9_][a-zA-Z0-9_-]{0,62}` +
		`(\.[a-zA-Z0-9_][a-zA-Z0-9_-]{0,62})*\.?$`

	// domainRx represents the regex source: https://stackoverflow.com/a/7933253
	// Slightly modified: Removed 255 max length validation since Go regex does
	// not support lookarounds. More info: https://stackoverflow.com/a/38935027
	domainRx = `^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+` +
		`(?:[a-zA-Z]{1,63}|` +
		`(?i:xn--)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,57}[a-zA-Z0-9])?)$`
)

// Compiled regexp rules.
var (
	dnsNameRxc = regexp.MustCompile(dnsNameRx)
	domainRxc  = regexp.MustCompile(domainRx)
)

// Net error codes.
const (
	ECIP      = "ECIP"      // Error code for an invalid IP address.
	ECIPv4    = "ECIPv4"    // Error code for an invalid IPv4 address.
	ECIPv6    = "ECIPv6"    // Error code for an invalid IPv6 address.
	ECPort    = "ECPort"    // Error code for an invalid network port.
	ECDNSName = "ECDNSName" // Error code for an invalid DNS name.
	ECDomain  = "ECDomain"  // Error code for an invalid domain name.
	ECHost    = "ECHost"    // Error code for an invalid network hostname.
)

// Validation error messages.
const (
	msgIP      = "must be a valid IP address"
	msgIPv4    = "must be a valid IPv4 address"
	msgIPv6    = "must be a valid IPv6 address"
	msgPort    = "must be a valid network port"
	msgDNSName = "must be a valid DNS name"
	msgDomain  = "must be a valid domain"
	msgHost    = "must be a valid network hostname"
)

// IsIP checks if a string is either IPv4 or IPv6.
func IsIP(str string) bool { return net.ParseIP(str) != nil }

// CheckIP is a [verax.RuleFunc] that checks that a string is a valid IPv4 or
// IPv6 address.
var CheckIP = verax.Check(IsIP, msgIP, ECIP)

// IP validates if a string is a valid IPv4 or IPv6 address.
var IP = verax.By(CheckIP)

// IsIPv4 checks if the string is IP version 4 in dotted-decimal notation;
// IPv4-mapped IPv6 addresses are rejected.
func IsIPv4(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && !strings.Contains(str, ":")
}

// CheckIPv4 is a [verax.RuleFunc] that checks that a string is a valid IPv4
// address.
var CheckIPv4 = verax.Check(IsIPv4, msgIPv4, ECIPv4)

// IPv4 validates if a string is a valid IPv4 address.
var IPv4 = verax.By(CheckIPv4)

// IsIPv6 checks if the string is IP version 6.
func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}

// CheckIPv6 is a [verax.RuleFunc] that checks that a string is a valid IPv6
// address.
var CheckIPv6 = verax.Check(IsIPv6, msgIPv6, ECIPv6)

// IPv6 validates if a string is a valid IPv6 address.
var IPv6 = verax.By(CheckIPv6)

// IsPort checks if a string represents a valid network port: a decimal
// number from 1 to 65535 without a sign or leading zeros.
func IsPort(str string) bool {
	if str == "" || str[0] < '1' || str[0] > '9' {
		return false
	}
	i, err := strconv.Atoi(str)
	return err == nil && i < 65536
}

// CheckPort is a [verax.RuleFunc] that checks that a string is a valid network
// port.
var CheckPort = verax.Check(IsPort, msgPort, ECPort)

// Port validates if a string is a valid network port number.
var Port = verax.By(CheckPort)

// IsDNSName checks if a string represents a valid DNS name of at most 253
// characters, not counting an optional trailing dot. The last label must not
// be all-numeric, which also rules out IPv4 addresses.
func IsDNSName(str string) bool {
	name := strings.TrimSuffix(str, ".")
	if name == "" || len(name) > 253 {
		return false
	}
	tld := name[strings.LastIndexByte(name, '.')+1:]
	if strings.Trim(tld, "0123456789") == "" {
		return false
	}
	return dnsNameRxc.MatchString(str)
}

// CheckDNSName is a [verax.RuleFunc] that checks that a string is a valid DNS
// name.
var CheckDNSName = verax.Check(IsDNSName, msgDNSName, ECDNSName)

// DNSName validates if a string is a valid DNS name.
var DNSName = verax.By(CheckDNSName)

// IsDomain checks if a string represents a valid domain name of at most 253
// characters.
func IsDomain(str string) bool {
	if str == "" || len(str) > 253 {
		return false
	}
	return domainRxc.MatchString(str)
}

// CheckDomain is a [verax.RuleFunc] that checks that a string is a valid
// domain name.
var CheckDomain = verax.Check(IsDomain, msgDomain, ECDomain)

// Domain validates if a string is a valid domain name.
var Domain = verax.By(CheckDomain)

// IsHost checks if the string is a valid IPv4, IPv6, or valid DNS name.
func IsHost(str string) bool { return IsIP(str) || IsDNSName(str) }

// CheckHost is a [verax.RuleFunc] that checks that a string is a valid
// network hostname.
var CheckHost = verax.Check(IsHost, msgHost, ECHost)

// Host validates if a string is a valid network hostname.
var Host = verax.By(CheckHost)
