// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
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

// TstSpec is a test type implementing the [Specable] interface.
type TstSpec struct {
	name string
	err  error
	args map[string]any
}

func (tst TstSpec) Spec() (*Spec, error) {
	if tst.err != nil {
		return nil, tst.err
	}
	spc := NewSpec(tst.name)
	if tst.args != nil {
		spc.Args = maps.Clone(tst.args)
	}
	return spc, nil
}

// TstFn0 is a test function instance.
func TstFn0() {}
