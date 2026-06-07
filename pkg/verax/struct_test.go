// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"reflect"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_Field(t *testing.T) {
	t.Run("without rules", func(t *testing.T) {
		// --- Given ---
		var s1 TStruct

		// --- When ---
		have := Field(&s1.FStr)

		// --- Then ---
		assert.Same(t, &s1.FStr, have.fieldPtr)
		assert.Equal(t, "", have.tag)
		assert.Nil(t, have.rules)
	})

	t.Run("with rules", func(t *testing.T) {
		// --- Given ---
		var s1 TStruct

		// --- When ---
		have := Field(&s1.FStr, Equal("FStr"))

		// --- Then ---
		assert.Same(t, &s1.FStr, have.fieldPtr)
		assert.Equal(t, "", have.tag)
		assert.Equal(t, []Rule{Equal("FStr")}, have.rules)
	})
}

func Test_Field_Tag(t *testing.T) {
	t.Run("field tag set", func(t *testing.T) {
		// --- Given ---
		var s1 TStruct
		fld := Field(s1.FStr)

		// --- When ---
		have := fld.Tag("custom")

		// --- Then ---
		assert.Equal(t, "custom", have.tag)
		assert.Equal(t, "", fld.tag)
	})
}

func Test_ValidateStruct(t *testing.T) {
	t.Run("nil struct", func(t *testing.T) {
		// --- Given ---
		var s *TStruct

		// --- When ---
		err := ValidateStruct(s)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid when no rules", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()

		// --- When ---
		err := ValidateStruct(&mf)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid when no field rules", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		fr := []FieldRule{
			Field(&mf.FStr),
			Field(&mf.FpStr),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid field", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		fr := []FieldRule{
			Field(&mf.FStr, Equal("FStr")),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid not exported field", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		fr := []FieldRule{
			Field(&mf.FStr, Equal("FStr")),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid field pointer", func(t *testing.T) {
		// --- Given ---
		str := "TStruct.FpStr"
		mf := NewTStruct()
		fr := []FieldRule{
			Field(&mf.FpStr, Equal(&str)),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid slice field each rule", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		fr := []FieldRule{
			Field(&mf.FsStr, Each(Length(1, 1))),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid array field each rule", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		rs := []FieldRule{
			Field(&mf.FaStr, Each(Length(1, 1))),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid map field rule", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		rs := []FieldRule{
			Field(&mf.FmStr, Each(Length(2, 2))),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid sub structs", func(t *testing.T) {
		// --- Given ---
		s := Model{
			ModelVal: ModelVal{"abc"},
			SvSM1:    ModelVal{"abc"},
			SpSM1:    &ModelVal{"abc"},
			SpSM2:    &ModelPtr{"abc"},
		}

		rs := []FieldRule{
			Field(&s.ModelVal),
			Field(&s.SvSM1),
			Field(&s.SpSM1),
			Field(&s.SpSM2),
		}

		// --- When ---
		err := ValidateStruct(&s, rs...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("invalid field with JSON tag", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()

		fr := []FieldRule{
			Field(&mf.FStr, Equal("other")),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "f_json: must be equal to 'other' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid field without JSON tag", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()

		fr := []FieldRule{
			Field(&mf.fStr, Equal("other")),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "fStr: must be equal to 'other' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid field with ignored JSON tag", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()

		fr := []FieldRule{
			Field(&mf.FpStr, Equal("other")),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "FpStr: must be equal to 'other' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid field from embedded", func(t *testing.T) {
		// --- Given ---
		s := Model{
			ModelVal: ModelVal{"abc"},
			SvSM1:    ModelVal{"abc"},
			SpSM1:    &ModelVal{"abc"},
			SpSM2:    &ModelPtr{"abc"},
		}

		fr := []FieldRule{
			Field(&s.FStr, Equal("other")),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "FStr: must be equal to 'other' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid field with value struct", func(t *testing.T) {
		// --- Given ---
		s := Model{
			SvSM1: ModelVal{"abc"},
			SpSM1: &ModelVal{"abc"},
			SpSM2: &ModelPtr{"abc"},
		}

		fr := []FieldRule{
			Field(&s.ModelVal),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		xrrtest.AssertEqual(t, "FStr: cannot be blank (ECRequired)", err)
	})

	t.Run("invalid field with value struct", func(t *testing.T) {
		// --- Given ---
		s := Model{
			ModelVal: ModelVal{"abc"},
			SvSM1:    ModelVal{"abc"},
			SpSM1:    &ModelVal{"abc"},
			SpSM2:    &ModelPtr{"abc"},
		}

		fr := []FieldRule{
			Field(&s.FStr, Fail("test msg", "ECTst")),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		xrrtest.AssertEqual(t, "FStr: test msg (ECTst)", err)
	})

	t.Run("error uses JSON field name", func(t *testing.T) {
		// --- Given ---
		s := TStruct{
			FStr: "abc",
		}

		fr := []FieldRule{
			Field(&s.FStr, Fail("test msg", "ECTst")),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		xrrtest.AssertEqual(t, "f_json: test msg (ECTst)", err)
	})

	t.Run("invalid field from value struct", func(t *testing.T) {
		// --- Given ---
		s := Model{
			ModelVal: ModelVal{"abc"},
			SvSM1:    ModelVal{"invalid"},
			SpSM1:    &ModelVal{"abc"},
			SpSM2:    &ModelPtr{"abc"},
		}

		fr := []FieldRule{
			Field(&s.SvSM1),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "SvSM1.FStr: must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid field from pointer struct value receiver", func(t *testing.T) {
		// --- Given ---
		s := Model{
			ModelVal: ModelVal{"abc"},
			SvSM1:    ModelVal{"abc"},
			SpSM1:    &ModelVal{"invalid"},
			SpSM2:    &ModelPtr{"abc"},
		}

		fr := []FieldRule{
			Field(&s.SpSM1),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "SpSM1.FStr: must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid field from pointer struct pointer receiver", func(t *testing.T) {
		// --- Given ---
		s := Model{
			ModelVal: ModelVal{"abc"},
			SvSM1:    ModelVal{"abc"},
			SpSM1:    &ModelVal{"abc"},
			SpSM2:    &ModelPtr{"invalid"},
		}

		fr := []FieldRule{
			Field(&s.SpSM2),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "SpSM2.FStr: must be equal to 'abc' (ECNotEqual)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("invalid multiple errors in a slice", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		rs := []FieldRule{
			Field(&mf.FaStr, Each(Length(2, 2))),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		want := "" +
			"FaStr.0: the length must be exactly 2 (ECInvLength); " +
			"FaStr.1: the length must be exactly 2 (ECInvLength); " +
			"FaStr.2: the length must be exactly 2 (ECInvLength); " +
			"FaStr.3: the length must be exactly 2 (ECInvLength)"
		xrrtest.AssertEqual(t, want, err)
	})

	t.Run("invalid multiple field errors", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		rs := []FieldRule{
			Field(&mf.FpStr, Length(2, 2)),
			Field(&mf.FaStr, Length(2, 2)),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		want := "" +
			"FaStr: the length must be exactly 2 (ECInvLength); " +
			"FpStr: the length must be exactly 2 (ECInvLength)"
		xrrtest.AssertEqual(t, want, err)
	})

	t.Run("non-struct pointer", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()

		// --- When ---
		err := ValidateStruct(mf)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorIs(t, ErrNotStructPtr, err)
	})

	t.Run("field not found", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		rs := []FieldRule{
			Field(&mf),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, "the field #0 cannot be found in the struct", err)
		xrrtest.AssertCode(t, ECInternal, err)
	})

	t.Run("field not pointer", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		rs := []FieldRule{
			Field(mf.FStr),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		assert.ErrorEqual(t, "field #0 must be specified as a pointer", err)
		xrrtest.AssertCode(t, ECInternal, err)
	})

	t.Run("field must not be nil", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		mf.FpStr = nil

		rs := []FieldRule{
			Field(&mf.FpStr, Required),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		xrrtest.AssertEqual(t, "FpStr: cannot be blank (ECRequired)", err)
	})

	t.Run("field must not be empty", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		mf.FStr = ""

		rs := []FieldRule{
			Field(&mf.FStr, Required),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		xrrtest.AssertEqual(t, "f_json: cannot be blank (ECRequired)", err)
	})

	t.Run("valid inline", func(t *testing.T) {
		// --- Given ---
		obj := struct {
			Name  string
			Value string
		}{
			"name",
			"demo",
		}

		// --- When ---
		err := ValidateStruct(
			&obj,
			Field(&obj.Name, Required),
			Field(&obj.Value, Required, Length(4, 10)),
		)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("invalid inline", func(t *testing.T) {
		// --- Given ---
		obj := struct {
			Name  string
			Value string
		}{
			"name",
			"demo",
		}

		// --- When ---
		err := ValidateStruct(
			&obj,
			Field(&obj.Name, Required),
			Field(&obj.Value, Required, Length(5, 10)),
		)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		wMsg := "Value: the length must be between 5 and 10 (ECInvLength)"
		xrrtest.AssertEqual(t, wMsg, err)
	})

	t.Run("rule returning ECInternal propagates as InternalError", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		rs := []FieldRule{
			Field(&mf.FStr, By(func(v any) error {
				return NewInternalError("bad rule", ECInternal)
			})),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		assert.SameType(t, &InternalError{}, err)
		xrrtest.AssertCode(t, ECInternal, err)
		assert.ErrorContain(t, "f_json", err)
	})

	t.Run("anonymous field non-Fielder error stored under field name", func(t *testing.T) {
		// --- Given ---
		s := Model{
			ModelVal: ModelVal{"abc"},
			SvSM1:    ModelVal{"abc"},
			SpSM1:    &ModelVal{"abc"},
			SpSM2:    &ModelPtr{"abc"},
		}
		rs := []FieldRule{
			Field(&s.ModelVal, Fail("anon err", "ECTst")),
		}

		// --- When ---
		err := ValidateStruct(&s, rs...)

		// --- Then ---
		assert.SameType(t, &FieldErrors{}, err)
		xrrtest.AssertEqual(t, "ModelVal: anon err (ECTst)", err)
	})
}

func Test_findStructField_found_tabular(t *testing.T) {
	em := NewEmbedded()
	ep := NewEmbeddedPtr()
	mf := NewTStruct()

	tt := []struct {
		testN string

		sp    any // Pointer to struct.
		field any // Pointer to struct field.
	}{
		{"string", &mf, &mf.FStr},
		{"unexported string", &mf, &mf.fStr},
		{"string pointer", &mf, &mf.FpStr},
		{"string slice", &mf, &mf.FsStr},
		{"string array", &mf, &mf.FaStr},
		{"struct pointer", &mf, &mf.SPtr},
		{"struct", &mf, &mf.SVal},
		{"nil struct pointer", &mf, &mf.SNil},
		{"string field of an embedded struct pointer", &ep, &ep.FStr},
		{"pointer to string of an embedded struct pointer", &ep, &ep.FStrPtr},
		{"string field of an embedded struct", &em, &em.FStrPtr},
		{"pointer to string of an embedded struct", &em, &em.FStr},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			s := reflect.ValueOf(tc.sp).Elem()
			f := reflect.ValueOf(tc.field)

			// --- When ---
			_, ok := findStructField(s, f)

			// --- Then ---
			assert.True(t, ok)
		})
	}
}

func Test_findStructField_not_found_tabular(t *testing.T) {
	var mf TStruct

	tt := []struct {
		testN string

		sp    any // Pointer to struct.
		field any
	}{
		{"1", &mf, mf.FpStr},
		{"2", &mf, mf.FsStr},
		{"3", &mf, mf.SPtr},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			s := reflect.ValueOf(tc.sp).Elem()
			f := reflect.ValueOf(tc.field)

			// --- When ---
			_, ok := findStructField(s, f)

			// --- Then ---
			assert.False(t, ok)
		})
	}
}

func Test_getErrorFieldName_tabular(t *testing.T) {
	var s1 TStruct

	tt := []struct {
		testN string

		sp    any // Pointer to struct.
		field any
		tag   string
		name  string
	}{
		{
			"use the default tag name when not provided",
			&s1,
			&s1.FStr,
			"",
			"f_json",
		},
		{
			"use the field name when no tags are present",
			&s1,
			&s1.SPtr,
			"",
			"SPtr",
		},
		{
			"field name when the default tag name is set to ignore",
			&s1,
			&s1.FpStr,
			"",
			"FpStr",
		},
		{"get tag name", &s1, &s1.FStr, "json", "f_json"},
		{
			"the filed name when provided tag does not exist",
			&s1,
			&s1.SPtr,
			"json",
			"SPtr",
		},
		{
			"field name when provided tag name is set to ignore",
			&s1,
			&s1.FpStr,
			"json",
			"FpStr",
		},
		{
			"use the default tag name when not provided and multiple exist",
			&s1,
			&s1.FsStr,
			"",
			"fs_str",
		},
		{
			"get the correct tag name when multiple exists",
			&s1,
			&s1.FsStr,
			"custom",
			"custom",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			elem := reflect.ValueOf(tc.sp).Elem()
			sf, _ := findStructField(elem, reflect.ValueOf(tc.field))

			// --- When ---
			have := getErrorFieldName(tc.tag, &sf)

			// --- Then ---
			assert.Equal(t, tc.name, have)
		})
	}
}

func Test_getErrorFieldName_json_tabular(t *testing.T) {
	type A struct {
		Field0 string `custom:"custom"`
		Field1 string `json:"f1"`
		Field2 string `json:"f2,omitempty"`
		Field3 string `json:",omitempty"`
		Field4 string `json:"f4,x1,omitempty"` //nolint:staticcheck
	}

	tt := []struct {
		testN string

		field   string
		tagName string
		name    string
	}{
		{"default - field name when not present", "Field0", "", "Field0"},
		{"default - present", "Field1", "", "f1"},
		{"default - has options", "Field2", "", "f2"},
		{"default - without name but with options", "Field3", "", "Field3"},
		{"default - with options", "Field4", "", "f4"},

		{"field name when the tag does not exist", "Field0", "json", "Field0"},
		{"tag name when exists", "Field1", "json", "f1"},
		{"tag name with options", "Field2", "json", "f2"},
		{"field name when the tag has no name", "Field3", "json", "Field3"},
		{"tag name with multiple options", "Field4", "json", "f4"},

		{"custom tag", "Field0", "custom", "custom"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			a := reflect.TypeFor[A]()

			// --- When ---
			have, _ := a.FieldByName(tc.field)

			// --- Then ---
			assert.Equal(t, tc.name, getErrorFieldName(tc.tagName, &have))
		})
	}
}

func BenchmarkValidateStruct(b *testing.B) {
	b.ReportAllocs()
	b.StopTimer()
	mf := NewTStruct()
	b.StartTimer()

	var err error
	for i := 0; i < b.N; i++ {
		err = ValidateStruct(&mf)
	}
	_ = err
}
