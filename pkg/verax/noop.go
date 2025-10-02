// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

// Compile time checks.
var (
	_ Customizer[ByRule]  = Noop
	_ Conditioner[ByRule] = Noop
)

// Noop is a special validation rule that always passes.
var Noop = By(func(v any) error { return nil })
