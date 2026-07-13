// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package spec

import (
	"reflect"

	"github.com/ctx42/mirror/pkg/mirror"
)

// Source represents values like functions which cannot be serialized.
//
// Is used to associate a unique name for non-serializable values so we can
// find them later when they are unserialized.
type Source struct {
	Lang string  `json:"lang"`           // Source language: go, js, etc.
	Name string  `json:"name"`           // Unique name of the source.
	Src  string  `json:"src,omitempty"`  // Optional source code location.
	Desc string  `json:"desc,omitempty"` // Optional description.
	val  any     // Function instance.
	ptr  uintptr // Pointer to the implementation.
}

// NewSource returns a new instance of named [Source] for the value. By default,
// the source language is set to "go".
func NewSource(name string, val any) (Source, error) {
	if val == nil {
		return Source{}, ErrInvSource
	}
	rv := reflect.ValueOf(val)
	ptr := GetSrcPointer(rv)
	if ptr == 0 {
		return Source{}, ErrInvSource
	}

	md := mirror.NewValueMetadata(rv)
	src := Source{
		Lang: "go",
		Name: name,
		Src:  md.PackageAndName(),
		val:  val,
		ptr:  ptr,
	}
	return src, nil
}

// SetLang sets the source language.
func (src Source) SetLang(lang string) Source {
	src.Lang = lang
	return src
}

// SetDesc sets the source description.
func (src Source) SetDesc(desc string) Source {
	src.Desc = desc
	return src
}

// Val retrieves the value described by the source.
func (src Source) Val() any { return src.val }

// Ptr retrieves the pointer to the value described by the source.
func (src Source) Ptr() uintptr { return src.ptr }

// IsZero reports whether the source is the zero value.
func (src Source) IsZero() bool { return Source{} == src }

// GetSrcPointer returns a pointer to the value. Returns zero for not supported
// or invalid values. Currently only functions are supported.
func GetSrcPointer(rv reflect.Value) uintptr {
	if !rv.IsValid() || rv.Kind() != reflect.Func {
		return 0
	}
	return rv.Pointer()
}
