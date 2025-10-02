package rule

import (
	"regexp"

	"github.com/ctx42/xrr/pkg/xrr"

	. "github.com/ctx42/verax/pkg/verax"
)

// semVerRx represents valid semantic version regular expression.
const semVerRx string = `` +
	`^v?(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)` +
	`(-(0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)` +
	`(\.(0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*)?` +
	`(\+[0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*)?$`

// semVerRxc represents semantic version compiled regular expression.
var semVerRxc = regexp.MustCompile(semVerRx)

// ErrSemVer is the error that returns in case of an invalid semver.
var ErrSemVer = xrr.New("must be a valid semantic version", "ECSemVer")

// IsSemver checks if string is valid semantic version.
func IsSemver(str string) bool {
	return semVerRxc.MatchString(str)
}

// SemVer validates if a string is a valid semantic version.
var SemVer = String(IsSemver).Error(ErrSemVer)
