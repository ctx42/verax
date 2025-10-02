package rule

import (
	"regexp"

	"github.com/ctx42/xrr/pkg/xrr"

	. "github.com/ctx42/verax/pkg/verax"
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

// Validation errors.
var (
	// ErrBase64 is the error that returns in the case of an invalid base64
	// value.
	ErrBase64 = xrr.New("must be a valid base64", "ECBase64")
)

// IsBase64 checks if a string is valid base64.
func IsBase64(str string) bool {
	if str == "" {
		return false
	}
	return base64Rxc.MatchString(str)
}

// Base64 validates if a string is a valid base64.
var Base64 = String(IsBase64).Error(ErrBase64)
