// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"time"
)

// ErrTst is an error instance used in tests.
var ErrTst = NewError("test msg", "ECTst")

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

// TwoStr is a struct not implementing [Validator] with two string fields.
type TwoStr struct {
	FStr    string
	FStrPtr *string
}

// NewTwoStr returns new instance of TwoStr.
func NewTwoStr() *TwoStr {
	p := "FpStr"
	return &TwoStr{FStr: "FStr", FStrPtr: &p}
}

func (t *TwoStr) String() string { return t.FStr + " " + *t.FStrPtr }

// EmbeddedPtr is a struct with TwoStr pointer embedded not implementing
// Validator interface.
type EmbeddedPtr struct {
	*TwoStr
}

// NewEmbeddedPtr returns a new instance of [EmbeddedPtr].
func NewEmbeddedPtr() EmbeddedPtr {
	p := "emp.TwoStr.FpStr"
	return EmbeddedPtr{
		TwoStr: &TwoStr{
			FStr:    "emp.TwoStr.FStr",
			FStrPtr: &p,
		},
	}
}

// Embedded is a struct with embedded TwoStr struct not implementing
// [Validator] interface.
type Embedded struct {
	TwoStr
}

// NewEmbedded returns a new instance of [Embedded].
func NewEmbedded() Embedded {
	p := "emb.TwoStr.FpStr"
	return Embedded{
		TwoStr: TwoStr{
			FStr:    "emb.TwoStr.FStr",
			FStrPtr: &p,
		},
	}
}

// TMap is a map used in tests.
var TMap = map[string]any{
	"KStrAbc":        "abc",
	"KStrXyz":        "xyz",
	"KStrEmpty":      "",
	"KpStr":          pString,
	"KpStrNil":       (*string)(nil),
	"KpStructNil":    (*ModelPtr)(nil),
	"KsString":       []string{"abc", "abc"},
	"KmStringString": map[string]string{"foo": "abc"},
	"KStructValid":   ModelVal{"abc"},
	"KStructInvalid": ModelVal{"xyz"},
}

// TMapInt is a map used in tests.
var TMapInt = map[int]any{
	1: "abc",
	3: "xyz",
}

// TStruct is a struct with multiple fields used for tests.
type TStruct struct {
	FStr  string `json:"f_json"`
	fStr  string
	FpStr *string  `json:"-"`
	FsStr []string `custom:"custom" json:"fs_str"`
	FaStr [4]string
	FmStr map[int]string
	SPtr  *TwoStr
	SVal  TwoStr
	SNil  *TwoStr
}

// NewTStruct returns TStruct with default values.
func NewTStruct() TStruct {
	FpStr := "TStruct.FpStr"
	PtrTwoStrFStrPtr := "ptr.TwoStr.FpStr"
	ValTwoStrFStrPtr := "val.TwoStr.FpStr"

	return TStruct{
		FStr:  "FStr",
		FpStr: &FpStr,
		FsStr: []string{"0", "1", "2"},
		FaStr: [4]string{"0", "1", "2", "3"},
		FmStr: map[int]string{1: "v1", 3: "vs"},
		fStr:  "fStr",
		SPtr: &TwoStr{
			FStr:    "ptr.TwoStr.FStr",
			FStrPtr: &PtrTwoStrFStrPtr,
		},
		SVal: TwoStr{
			FStr:    "ptr.TwoStr.FStr",
			FStrPtr: &ValTwoStrFStrPtr,
		},
		SNil: nil,
	}
}

// Model is a struct with few sub structs as fields, not implementing
// [Validator] interface.
type Model struct {
	ModelVal           // Embedded struct.
	SvSM1    ModelVal  // Value struct.
	SpSM1    *ModelVal // Pointer to struct (value receiver).
	SpSM2    *ModelPtr // Pointer to struct (pointer receiver).
}

// ModelVal implements [Validator] interface with value receiver.
type ModelVal struct {
	FStr string
}

func (m ModelVal) Validate() error {
	return ValidateStruct(&m, Field(&m.FStr, Required, Equal("abc")))
}

func (m ModelVal) String() string { return m.FStr }

// ModelPtr implements [Validator] interface with a pointer receiver.
type ModelPtr struct {
	FStr string
}

func (m *ModelPtr) Validate() error {
	return ValidateStruct(m, Field(&m.FStr, Required, Equal("abc")))
}

func (m *ModelPtr) String() string { return m.FStr }

// ModelVW implements [WithValidator] interface.
type ModelVW struct {
	value string
}

func (m *ModelVW) ValidateWith(rule Rule) error {
	if m.value == "too_long" {
		return ErrTst
	}
	return rule.Validate(m.value)
}

// TstRule is a test structure implementing [Rule] interface.
type TstRule struct{ want string }

// Validate returns error unless v is 42 or "abc".
func (t TstRule) Validate(have any) error {
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
