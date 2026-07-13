// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package rule

import (
	"regexp"

	"github.com/ctx42/verax/pkg/verax"
)

// semVerRx represents valid semantic version regular expression.
const semVerRx string = `` +
	`^v?(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)` +
	`(-(0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)` +
	`(\.(0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*)?` +
	`(\+[0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*)?$`

// semVerRxc represents semantic version compiled regular expression.
var semVerRxc = regexp.MustCompile(semVerRx)

// ECSemVer is an error code for an invalid semantic version.
const ECSemVer = "ECSemVer"

// msgSemVer is the error message for an invalid semantic version.
var msgSemVer = "must be a valid semantic version"

// IsSemVer checks whether a string is a valid semantic version.
func IsSemVer(str string) bool {
	return semVerRxc.MatchString(str)
}

// CheckSemVer is [verax.RuleFunc] that checks a string is a valid semantic
// version.
var CheckSemVer = verax.Check(IsSemVer, msgSemVer, ECSemVer)

// SemVer validates if a string is a valid semantic version.
var SemVer = verax.By(CheckSemVer)
