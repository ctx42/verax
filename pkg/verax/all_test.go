// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"time"
)

// errTst is an error instance used in tests.
var errTst = NewError("test msg", "ECTst")

// Types used in tests.
type tStructEmpty struct{}

// Declared variables used in tests.
var (
	dString      string
	dInt         int
	dTime        time.Time
	dStructEmpty tStructEmpty
	dSlice       []byte
	dArray       [2]byte
	dMap         map[string]struct{}
	dChan        chan int
	dFunc        func(v any) bool
	dValidate    Validator
	dInterface   any
)

// Pointers used in tests.
var (
	pStringNil      *string
	pIntNil         *int
	pTimeNil        *time.Time
	pStructEmptyNil *tStructEmpty

	pString      = &iString
	pStringEmpty = &iStringEmpty
	pInt         = &iInt
	pIntZero     = &iIntZero
	pTime        = &iTime
	pTimeZero    = &iTimeZero
	pStructEmpty = &iStructEmpty
)

// Initialized variables used in tests.
var (
	iString        = "test string"
	iStringEmpty   = ""
	iInt           = 123
	iIntZero       = 0
	iTime          = time.Date(2022, 2, 25, 21, 13, 0, 0, time.UTC)
	iTimeZero      = time.Time{}
	iStructEmpty   = tStructEmpty{}
	iChan          = make(chan int)
	iFunc          = func(v any) bool { return true }
	iInterface     = any(123)
	iInterfaceZero = any(0)
)

// twoStr is a struct not implementing [Validator] with two string fields.
type twoStr struct {
	FStr    string
	FStrPtr *string
}

// newTwoStr returns new instance of twoStr.
func newTwoStr() *twoStr {
	p := "FpStr"
	return &twoStr{FStr: "FStr", FStrPtr: &p}
}

func (two *twoStr) String() string { return two.FStr + " " + *two.FStrPtr }

// embeddedPtr is a struct with twoStr pointer embedded not implementing
// Validator interface.
type embeddedPtr struct {
	*twoStr
}

// newEmbeddedPtr returns a new instance of [embeddedPtr].
func newEmbeddedPtr() embeddedPtr {
	p := "emp.twoStr.FpStr"
	return embeddedPtr{
		twoStr: &twoStr{
			FStr:    "emp.twoStr.FStr",
			FStrPtr: &p,
		},
	}
}

// embedded is a struct with embedded twoStr struct not implementing
// [Validator] interface.
type embedded struct {
	twoStr
}

// newEmbedded returns a new instance of [embedded].
func newEmbedded() embedded {
	p := "emb.twoStr.FpStr"
	return embedded{
		twoStr: twoStr{
			FStr:    "emb.twoStr.FStr",
			FStrPtr: &p,
		},
	}
}

// tMap is a map used in tests.
var tMap = map[string]any{
	"KStrAbc":        "abc",
	"KStrXyz":        "xyz",
	"KStrEmpty":      "",
	"KpStr":          pString,
	"KpStrNil":       (*string)(nil),
	"KpStructNil":    (*modelPtr)(nil),
	"KsString":       []string{"abc", "abc"},
	"KmStringString": map[string]string{"foo": "abc"},
	"KStructValid":   modelVal{"abc"},
	"KStructInvalid": modelVal{"xyz"},
}

// tMapInt is a map used in tests.
var tMapInt = map[int]any{
	1: "abc",
	3: "xyz",
}

// tStruct is a struct with multiple fields used for tests.
type tStruct struct {
	FStr  string `json:"f_json"`
	fStr  string
	FpStr *string  `json:"-"`
	FsStr []string `custom:"custom" json:"fs_str"`
	FaStr [4]string
	FmStr map[int]string
	SPtr  *twoStr
	SVal  twoStr
	SNil  *twoStr
}

// newTStruct returns tStruct with default values.
func newTStruct() tStruct {
	FpStr := "tStruct.FpStr"
	PtrTwoStrFStrPtr := "ptr.twoStr.FpStr"
	ValTwoStrFStrPtr := "val.twoStr.FpStr"

	return tStruct{
		FStr:  "FStr",
		FpStr: &FpStr,
		FsStr: []string{"0", "1", "2"},
		FaStr: [4]string{"0", "1", "2", "3"},
		FmStr: map[int]string{1: "v1", 3: "vs"},
		fStr:  "fStr",
		SPtr: &twoStr{
			FStr:    "ptr.twoStr.FStr",
			FStrPtr: &PtrTwoStrFStrPtr,
		},
		SVal: twoStr{
			FStr:    "ptr.twoStr.FStr",
			FStrPtr: &ValTwoStrFStrPtr,
		},
		SNil: nil,
	}
}

// model is a struct with few sub structs as fields, not implementing
// [Validator] interface.
type model struct {
	modelVal           // embedded struct.
	SvSM1    modelVal  // Value struct.
	SpSM1    *modelVal // Pointer to struct (value receiver).
	SpSM2    *modelPtr // Pointer to struct (pointer receiver).
}

// modelVal implements [Validator] interface with value receiver.
type modelVal struct {
	FStr string
}

func (mdl modelVal) Validate() error {
	return ValidateStruct(&mdl, Field(&mdl.FStr, Required, Equal("abc")))
}

func (mdl modelVal) String() string { return mdl.FStr }

// modelPtr implements [Validator] interface with a pointer receiver.
type modelPtr struct {
	FStr string
}

func (mdl *modelPtr) Validate() error {
	return ValidateStruct(mdl, Field(&mdl.FStr, Required, Equal("abc")))
}

func (mdl *modelPtr) String() string { return mdl.FStr }

// modelVW implements [WithValidator] interface.
type modelVW struct {
	value string
}

func (mvw *modelVW) ValidateWith(rule Rule) error {
	if mvw.value == "too_long" {
		return errTst
	}
	return rule.Validate(mvw.value)
}

// tstRule is a test structure implementing [Rule] interface.
type tstRule struct{ want string }

// Validate returns error unless v is 42 or "abc".
func (tst tstRule) Validate(have any) error {
	switch val := have.(type) {
	case int:
		if val == 42 {
			return nil
		}
		return NewErrorf("invalid value '%v'", val)

	case string:
		if val == "abc" {
			return nil
		}
		return NewErrorf("invalid value '%v'", val)
	}
	return NewErrorf("invalid value type: %T", have)
}
