// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package rule

import (
	"regexp"

	"github.com/ctx42/verax/pkg/verax"
)

// base64Rx matches standard, padded base64.
const base64Rx = `` +
	`^(?:[A-Za-z0-9+\/]{4})*` +
	`(?:[A-Za-z0-9+\/]{2}==|[A-Za-z0-9+\/]{3}=|[A-Za-z0-9+\/]{4})$`

var base64Rxc = regexp.MustCompile(base64Rx)

// ECBase64 is the error code for an invalid base64 value.
const ECBase64 = "ECBase64"

const msgBase64 = "must be a valid base64"

// IsBase64 checks if a string is valid base64.
func IsBase64(str string) bool {
	if str == "" {
		return false
	}
	return base64Rxc.MatchString(str)
}

// CheckBase64 is a [verax.RuleFunc] that checks that a string is valid
// base64.
var CheckBase64 = verax.Check(IsBase64, msgBase64, ECBase64)

// Base64 validates if a string is a valid base64.
var Base64 = verax.By(CheckBase64)
