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

var semVerRxc = regexp.MustCompile(semVerRx)

// ECSemVer is an error code for an invalid semantic version.
const ECSemVer = "ECSemVer"

const msgSemVer = "must be a valid semantic version"

// IsSemVer checks whether a string is a valid semantic version.
func IsSemVer(str string) bool {
	return semVerRxc.MatchString(str)
}

// CheckSemVer is a [verax.RuleFunc] that checks that a string is a valid
// semantic version.
var CheckSemVer = verax.Check(IsSemVer, msgSemVer, ECSemVer)

// SemVer validates if a string is a valid semantic version.
var SemVer = verax.By(CheckSemVer)
