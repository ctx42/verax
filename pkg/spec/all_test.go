// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package spec

import (
	"maps"
)

// ErrTst is an error instance used in tests.
var ErrTst = NewError("test msg", "ECTst")

// TstType is a test type.
type TstType struct{ name string }

// TstBuilder is a builder type for TstType.
type TstBuilder = Builder[TstType]

// TstSpec is a test type implementing [Specable] interface.
type TstSpec struct {
	name string
	err  error
	args map[string]any
}

func (s TstSpec) Spec() (*Spec, error) {
	if s.err != nil {
		return nil, s.err
	}
	spc := NewSpec(s.name)
	if s.args != nil {
		spc.Args = maps.Clone(s.args)
	}
	return spc, nil
}

func TstFn0() {} // Test function instance.
