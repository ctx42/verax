package rule

import (
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/ctx42/xrr/pkg/xrr"

	"github.com/ctx42/verax/pkg/verax"
)

// Regexp rules.
const (
	// dnsNameRx represents valid DNS name regular expression.
	dnsNameRx string = `^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}` +
		`(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`

	// domainRx represents regex source: https://stackoverflow.com/a/7933253
	// Slightly modified: Removed 255 max length validation since Go regex does
	// not support lookarounds. More info: https://stackoverflow.com/a/38935027
	domainRx = `^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-z0-9])?\.)+` +
		`(?:[a-zA-Z]{1,63}| xn--[a-z0-9]{1,59})$`
)

// Compiled regexp rules.
var (
	// dnsNameRxc represents compiled valid DNS name regular expression.
	dnsNameRxc = regexp.MustCompile(dnsNameRx)

	// domainRxc represents compiled valid domain name regular expression.
	domainRxc = regexp.MustCompile(domainRx)
)

// Validation errors.
var (
	// ErrIP is the error that returns in case of an invalid IPv4 or IPv6
	// address.
	ErrIP = xrr.New("must be a valid IP address", "ECIP")

	// ErrIPv4 is the error that returns in case of an invalid IPv4 address.
	ErrIPv4 = xrr.New("must be a valid IPv4 address", "ECIPv4")

	// ErrIPv6 is the error that returns in case of an invalid IPv6 address.
	ErrIPv6 = xrr.New("must be a valid IPv6 address", "ECIPv6")

	// ErrPort is the error that returns in case of an invalid IP port.
	ErrPort = xrr.New("must be a valid network port", "ECPort")

	// ErrDNSName is the error that returns in case of an invalid DNS name.
	ErrDNSName = xrr.New("must be a valid DNS name", "ECDNSName")

	// ErrDomain is the error that returns in case of an invalid domain name.
	ErrDomain = xrr.New("must be a valid domain", "ECDomain")

	// ErrHost is the error that returns in case of an invalid network hostname.
	ErrHost = xrr.New("must be a valid network hostname", "ECHost")
)

// IsIP checks if a string is either IPv4 or IPv6.
func IsIP(str string) bool { return net.ParseIP(str) != nil }

// IP validates if a string is a valid IPv4 or IPv6 address.
var IP = verax.String(IsIP).Error(ErrIP)

// IsIPv4 checks if the string is IP version 4.
func IsIPv4(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ".")
}

// IPv4 validates if a string is a valid IPv4 address.
var IPv4 = verax.String(IsIPv4).Error(ErrIPv4)

// IsIPv6 checks if the string is IP version 6.
func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}

// IPv6 validates if a string is a valid IPv6 address.
var IPv6 = verax.String(IsIPv6).Error(ErrIPv6)

// IsPort checks if a string represents a valid network port.
func IsPort(str string) bool {
	if i, err := strconv.Atoi(str); err == nil {
		return i > 0 && i < 65536
	}
	return false
}

// Port validates if a string is a valid network port number.
var Port = verax.String(IsPort).Error(ErrPort)

// IsDNSName checks if a string represents a valid DNS name.
func IsDNSName(str string) bool {
	if str == "" || len(strings.ReplaceAll(str, ".", "")) > 255 {
		return false
	}
	return !IsIP(str) && dnsNameRxc.MatchString(str)
}

// DNSName validates if a string is a valid DNS name.
var DNSName = verax.String(IsDNSName).Error(ErrDNSName)

// IsDomain checks if a string represents a valid domain name.
func IsDomain(str string) bool {
	if str == "" || len(str) > 255 {
		return false
	}
	return domainRxc.MatchString(str)
}

// Domain validates if a string is a valid domain name.
var Domain = verax.String(IsDomain).Error(ErrDomain)

// IsHost checks if the string is a valid IPv4, IPv6 or valid DNS name.
func IsHost(str string) bool { return IsIP(str) || IsDNSName(str) }

// Host validates if a string is a valid network hostname.
var Host = verax.String(IsHost).Error(ErrHost)
