// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package verax

import (
	"reflect"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/xrr/pkg/xrr/xrrtest"
)

func Test_ErrFieldNotFound(t *testing.T) {
	// --- Given ---
	err := ErrFieldNotFound(123)

	// --- Then ---
	assert.ErrorEqual(t, "the field #123 cannot be found in the struct", err)
	xrrtest.AssertCode(t, ECInternal, err)
}

func Test_ErrFieldPointer(t *testing.T) {
	// --- Given ---
	err := ErrFieldPointer(123)

	// --- Then ---
	assert.ErrorEqual(t, "field #123 must be specified as a pointer", err)
	xrrtest.AssertCode(t, ECInternal, err)
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
			have := findStructField(s, f)

			// --- Then ---
			assert.NotNil(t, have)
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
			have := findStructField(s, f)

			// --- Then ---
			assert.Nil(t, have)
		})
	}
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

	t.Run("valid no rules", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()

		// --- When ---
		err := ValidateStruct(&mf)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid no field rules", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		fr := []*FieldRules{
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
		fr := []*FieldRules{
			Field(&mf.FStr, StrRule("FStr")),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid not exported field", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		fr := []*FieldRules{
			Field(&mf.FStr, StrRule("FStr")),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid field pointer", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		fr := []*FieldRules{
			Field(&mf.FpStr, StrRule("TStruct.FpStr")),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		assert.NoError(t, err)
	})

	t.Run("valid slice field each rule", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		fr := []*FieldRules{
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
		rs := []*FieldRules{
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
		rs := []*FieldRules{
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

		rs := []*FieldRules{
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

		fr := []*FieldRules{
			Field(&mf.FStr, StrRule("other")),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		xrrtest.AssertEqual(t, "f_json: must be 'other' (ECMustOther)", err)
	})

	t.Run("invalid field without JSON tag", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()

		fr := []*FieldRules{
			Field(&mf.fStr, StrRule("other")),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		xrrtest.AssertEqual(t, "fStr: must be 'other' (ECMustOther)", err)
	})

	t.Run("invalid field with ignored JSON tag", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()

		fr := []*FieldRules{
			Field(&mf.FpStr, StrRule("other")),
		}

		// --- When ---
		err := ValidateStruct(&mf, fr...)

		// --- Then ---
		xrrtest.AssertEqual(t, "FpStr: must be 'other' (ECMustOther)", err)
	})

	t.Run("invalid field from embedded", func(t *testing.T) {
		// --- Given ---
		s := Model{
			ModelVal: ModelVal{"abc"},
			SvSM1:    ModelVal{"abc"},
			SpSM1:    &ModelVal{"abc"},
			SpSM2:    &ModelPtr{"abc"},
		}

		fr := []*FieldRules{
			Field(&s.FStr, StrRule("other")),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		xrrtest.AssertEqual(t, "FStr: must be 'other' (ECMustOther)", err)
	})

	t.Run("invalid field with value struct", func(t *testing.T) {
		// --- Given ---
		s := Model{
			SvSM1: ModelVal{"abc"},
			SpSM1: &ModelVal{"abc"},
			SpSM2: &ModelPtr{"abc"},
		}

		fr := []*FieldRules{
			Field(&s.ModelVal),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
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

		fr := []*FieldRules{
			Field(&s.FStr, InternalErrRule),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		xrrtest.AssertEqual(t, "FStr: internal error (ECInternal)", err)
	})

	t.Run("internal error uses json field name", func(t *testing.T) {
		// --- Given ---
		s := TStruct{
			FStr: "",
		}

		fr := []*FieldRules{
			Field(&s.FStr, InternalErrRule),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		xrrtest.AssertEqual(t, "f_json: internal error (ECInternal)", err)
	})

	t.Run("invalid field from value struct", func(t *testing.T) {
		// --- Given ---
		s := Model{
			ModelVal: ModelVal{"abc"},
			SvSM1:    ModelVal{"invalid"},
			SpSM1:    &ModelVal{"abc"},
			SpSM2:    &ModelPtr{"abc"},
		}

		fr := []*FieldRules{
			Field(&s.SvSM1),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		xrrtest.AssertEqual(t, "SvSM1.FStr: must be 'abc' (ECMustAbc)", err)
	})

	t.Run("invalid field from pointer struct value receiver", func(t *testing.T) {
		// --- Given ---
		s := Model{
			ModelVal: ModelVal{"abc"},
			SvSM1:    ModelVal{"abc"},
			SpSM1:    &ModelVal{"invalid"},
			SpSM2:    &ModelPtr{"abc"},
		}

		fr := []*FieldRules{
			Field(&s.SpSM1),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		xrrtest.AssertEqual(t, "SpSM1.FStr: must be 'abc' (ECMustAbc)", err)
	})

	t.Run("invalid field from pointer struct pointer receiver", func(t *testing.T) {
		// --- Given ---
		s := Model{
			ModelVal: ModelVal{"abc"},
			SvSM1:    ModelVal{"abc"},
			SpSM1:    &ModelVal{"abc"},
			SpSM2:    &ModelPtr{"invalid"},
		}

		fr := []*FieldRules{
			Field(&s.SpSM2),
		}

		// --- When ---
		err := ValidateStruct(&s, fr...)

		// --- Then ---
		xrrtest.AssertEqual(t, "SpSM2.FStr: must be 'abc' (ECMustAbc)", err)
	})

	t.Run("invalid multiple errors in slice", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		rs := []*FieldRules{
			Field(&mf.FaStr, Each(Length(2, 2))),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		exp := "" +
			"FaStr.0: the length must be exactly 2 (ECInvLength); " +
			"FaStr.1: the length must be exactly 2 (ECInvLength); " +
			"FaStr.2: the length must be exactly 2 (ECInvLength); " +
			"FaStr.3: the length must be exactly 2 (ECInvLength)"
		xrrtest.AssertFieldsEqual(t, exp, err)
	})

	t.Run("invalid multiple field errors", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		rs := []*FieldRules{
			Field(&mf.FpStr, Length(2, 2)),
			Field(&mf.FaStr, Length(2, 2)),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		exp := "" +
			"FaStr: the length must be exactly 2 (ECInvLength); " +
			"FpStr: the length must be exactly 2 (ECInvLength)"
		xrrtest.AssertFieldsEqual(t, exp, err)
	})

	t.Run("non-struct pointer", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()

		// --- When ---
		err := ValidateStruct(mf, nil)

		// --- Then ---
		assert.ErrorIs(t, ErrNotStructPtr, err)
	})

	t.Run("field not found", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		rs := []*FieldRules{
			Field(&mf),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		var e ErrFieldNotFound
		assert.ErrorAs(t, &e, err)
		xrrtest.AssertCode(t, ECInternal, err)
	})

	t.Run("field not pointer", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		rs := []*FieldRules{
			Field(mf.FStr),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		var e ErrFieldPointer
		assert.ErrorAs(t, &e, err)
		xrrtest.AssertCode(t, ECInternal, err)
	})

	t.Run("field must not be nil", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		mf.FpStr = nil

		rs := []*FieldRules{
			Field(&mf.FpStr, Required),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
		xrrtest.AssertEqual(t, "FpStr: cannot be blank (ECRequired)", err)
	})

	t.Run("field must not be empty", func(t *testing.T) {
		// --- Given ---
		mf := NewTStruct()
		mf.FStr = ""

		rs := []*FieldRules{
			Field(&mf.FStr, Required),
		}

		// --- When ---
		err := ValidateStruct(&mf, rs...)

		// --- Then ---
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
		wMsg := "Value: the length must be between 5 and 10 (ECInvLength)"
		xrrtest.AssertEqual(t, wMsg, err)
	})
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
			sf := findStructField(elem, reflect.ValueOf(tc.field))

			// --- When ---
			name := getErrorFieldName(tc.tag, sf)

			// --- Then ---
			assert.Equal(t, tc.name, name)
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
			a := reflect.TypeOf(A{})

			// --- When ---
			field, _ := a.FieldByName(tc.field)

			// --- Then ---
			assert.Equal(t, tc.name, getErrorFieldName(tc.tagName, &field))
		})
	}
}

func Test_Field_Tag(t *testing.T) {
	t.Run("tag not set", func(t *testing.T) {
		// --- Given ---
		var s1 TStruct

		// --- When ---
		fr := Field(s1.FStr)

		// --- Then ---
		assert.Equal(t, "", fr.tag)
	})

	t.Run("tag set", func(t *testing.T) {
		// --- Given ---
		var s1 TStruct

		// --- When ---
		fr := Field(s1.FStr).Tag("custom")

		// --- Then ---
		assert.Equal(t, "custom", fr.tag)
	})
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
