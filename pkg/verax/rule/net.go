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
	// dnsNameRx represents valid DNS name regular expression.
	dnsNameRx string = `^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}` +
		`(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`

	// domainRx represents the regex source: https://stackoverflow.com/a/7933253
	// Slightly modified: Removed 255 max length validation since Go regex does
	// not support lookarounds. More info: https://stackoverflow.com/a/38935027
	domainRx = `^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-z0-9])?\.)+` +
		`(?:[a-zA-Z]{1,63}|xn--[a-z0-9]{1,59})$`
)

// Compiled regexp rules.
var (
	// dnsNameRxc represents compiled valid DNS name regular expression.
	dnsNameRxc = regexp.MustCompile(dnsNameRx)

	// domainRxc represents compiled valid domain name regular expression.
	domainRxc = regexp.MustCompile(domainRx)
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

// Validation errors.
var (
	// msgIP is the error message for invalid IPv4 or IPv6 address.
	msgIP = "must be a valid IP address"

	// msgIPv4 is the error message for invalid IPv4 address.
	msgIPv4 = "must be a valid IPv4 address"

	// msgIPv6 is the error message for invalid IPv6 address.
	msgIPv6 = "must be a valid IPv6 address"

	// msgPort is the error message for an invalid network port.
	msgPort = "must be a valid network port"

	// msgDNSName is the error message for an invalid DNS name.
	msgDNSName = "must be a valid DNS name"

	// msgDomain is the error message for an invalid domain name.
	msgDomain = "must be a valid domain"

	// msgHost is the error message for an invalid network hostname.
	msgHost = "must be a valid network hostname"
)

// IsIP checks if a string is either IPv4 or IPv6.
func IsIP(str string) bool { return net.ParseIP(str) != nil }

// CheckIP is [verax.RuleFunc] that checks a string is valid IPv4 or IPv6.
var CheckIP = verax.Check(IsIP, msgIP, ECIP)

// IP validates if a string is a valid IPv4 or IPv6 address.
var IP = verax.By(CheckIP)

// IsIPv4 checks if the string is IP version 4.
func IsIPv4(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ".")
}

// CheckIPv4 is [verax.RuleFunc] that checks a string is a valid IPv4.
var CheckIPv4 = verax.Check(IsIPv4, msgIPv4, ECIPv4)

// IPv4 validates if a string is a valid IPv4 address.
var IPv4 = verax.By(CheckIPv4)

// IsIPv6 checks if the string is IP version 6.
func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}

// CheckIPv6 is [verax.RuleFunc] that checks a string is a valid IPv6.
var CheckIPv6 = verax.Check(IsIPv6, msgIPv6, ECIPv6)

// IPv6 validates if a string is a valid IPv6 address.
var IPv6 = verax.By(CheckIPv6)

// IsPort checks if a string represents a valid network port.
func IsPort(str string) bool {
	if i, err := strconv.Atoi(str); err == nil {
		return i > 0 && i < 65536
	}
	return false
}

// CheckPort is [verax.RuleFunc] that checks a string is a valid network port.
var CheckPort = verax.Check(IsPort, msgPort, ECPort)

// Port validates if a string is a valid network port number.
var Port = verax.By(CheckPort)

// IsDNSName checks if a string represents a valid DNS name.
func IsDNSName(str string) bool {
	if str == "" || len(strings.ReplaceAll(str, ".", "")) > 255 {
		return false
	}
	return !IsIP(str) && dnsNameRxc.MatchString(str)
}

// CheckDNSName is [verax.RuleFunc] that checks a string is a valid DNS name.
var CheckDNSName = verax.Check(IsDNSName, msgDNSName, ECDNSName)

// DNSName validates if a string is a valid DNS name.
var DNSName = verax.By(CheckDNSName)

// IsDomain checks if a string represents a valid domain name.
func IsDomain(str string) bool {
	if str == "" || len(str) > 255 {
		return false
	}
	return domainRxc.MatchString(str)
}

// CheckDomain is [verax.RuleFunc] that checks a string is a valid domain name.
var CheckDomain = verax.Check(IsDomain, msgDomain, ECDomain)

// Domain validates if a string is a valid domain name.
var Domain = verax.By(CheckDomain)

// IsHost checks if the string is a valid IPv4, IPv6, or valid DNS name.
func IsHost(str string) bool { return IsIP(str) || IsDNSName(str) }

// CheckHost is [verax.RuleFunc] that checks a string is a valid network
// hostname.
var CheckHost = verax.Check(IsHost, msgHost, ECHost)

// Host validates if a string is a valid network hostname.
var Host = verax.By(CheckHost)
