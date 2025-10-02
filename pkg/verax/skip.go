// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

// Skip is a special validation rule that indicates all rules following it
// should be skipped.
var Skip = skipRule(true)

type skipRule bool

func (_ skipRule) Validate(_ any) error { return nil }

// When specifies a condition that determines whether validation should be
// performed. If the condition is false, validation is skipped, and no errors
// are reported.
func (_ skipRule) When(condition bool) skipRule { return skipRule(condition) }
