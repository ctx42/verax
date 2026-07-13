// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package rule

import (
	"regexp"

	"github.com/ctx42/verax/pkg/verax"
)

// Regexp rules.
const (
	// base64Rx represents valid base64 regular expression.
	base64Rx string = `^(?:[A-Za-z0-9+\/]{4})*(?:[A-Za-z0-9+\/]{2}==|[A-Za-z0-9+\/]{3}=|[A-Za-z0-9+\/]{4})$`
)

// Compiled regexp rules.
var (
	// base64Rxc represents compiled valid base64 regular expression.
	base64Rxc = regexp.MustCompile(base64Rx)
)

// ECBase64 is the error code for an invalid base64 value.
const ECBase64 = "ECBase64"

// Validation error messages.
var (
	// msgBase64 is the error message for an invalid base64 value.
	msgBase64 = "must be a valid base64"
)

// IsBase64 checks if a string is valid base64.
func IsBase64(str string) bool {
	if str == "" {
		return false
	}
	return base64Rxc.MatchString(str)
}

// CheckBase64 is [verax.RuleFunc] that checks a string is valid base64.
var CheckBase64 = verax.Check(IsBase64, msgBase64, ECBase64)

// Base64 validates if a string is a valid base64.
var Base64 = verax.By(CheckBase64)
