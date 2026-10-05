// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package rule provides specialized validation rules for use with
// [verax.Validate]: base64, network values (IP addresses, ports, DNS names,
// domains, and hosts), and semantic versions.
//
// Each rule comes in three forms: an IsXxx predicate, a CheckXxx
// [verax.RuleFunc], and a ready-to-use Xxx rule. The rules accept only
// strings and treat an empty string as valid; combine them with
// [verax.Required] to reject empty values.
//
//	import "github.com/ctx42/verax/pkg/verax/rule"
//
//	err := verax.Validate("example.com", verax.Required, rule.Domain)
package rule
