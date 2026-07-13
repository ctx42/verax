// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package verax

import (
	"regexp"
	"testing"
	"time"
)

// Sink variables prevent the compiler from optimising away benchmark calls.
var (
	sinkErr  error
	sinkBool bool
	sinkAny  any
)

// Benchmark structs used for ValidateStruct benchmarks.
type benchUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

var benchEmailRx = regexp.MustCompile(
	`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
)

// --- IsNil ---

func Benchmark_IsNil_nil(b *testing.B) {
	for b.Loop() {
		sinkBool = IsNil(nil)
	}
}

func Benchmark_IsNil_string(b *testing.B) {
	v := "hello"
	for b.Loop() {
		sinkBool = IsNil(v)
	}
}

func Benchmark_IsNil_nilPointer(b *testing.B) {
	var p *string
	for b.Loop() {
		sinkBool = IsNil(p)
	}
}

// --- IsEmpty ---

func Benchmark_IsEmpty_string_nonempty(b *testing.B) {
	v := "hello"
	for b.Loop() {
		sinkBool = IsEmpty(v)
	}
}

func Benchmark_IsEmpty_string_empty(b *testing.B) {
	v := ""
	for b.Loop() {
		sinkBool = IsEmpty(v)
	}
}

func Benchmark_IsEmpty_int_zero(b *testing.B) {
	var v int
	for b.Loop() {
		sinkBool = IsEmpty(v)
	}
}

func Benchmark_IsEmpty_int_nonzero(b *testing.B) {
	v := 42
	for b.Loop() {
		sinkBool = IsEmpty(v)
	}
}

func Benchmark_IsEmpty_time_zero(b *testing.B) {
	v := time.Time{}
	for b.Loop() {
		sinkBool = IsEmpty(v)
	}
}

func Benchmark_IsEmpty_time_nonzero(b *testing.B) {
	v := time.Now()
	for b.Loop() {
		sinkBool = IsEmpty(v)
	}
}

// --- Indirect ---

func Benchmark_Indirect_string(b *testing.B) {
	v := "hello"
	for b.Loop() {
		sinkAny = Indirect(v)
	}
}

func Benchmark_Indirect_pointer(b *testing.B) {
	s := "hello"
	v := &s
	for b.Loop() {
		sinkAny = Indirect(v)
	}
}

// --- RequiredRule ---

func Benchmark_RequiredRule_valid_string(b *testing.B) {
	r := Required
	v := "hello"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_RequiredRule_invalid_string(b *testing.B) {
	r := Required
	v := ""
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_RequiredRule_valid_int(b *testing.B) {
	r := Required
	v := 42
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_RequiredRule_invalid_int(b *testing.B) {
	r := Required
	var v int
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

// --- RangeRule (Min/Max) ---

func Benchmark_RangeRule_Min_valid_int(b *testing.B) {
	r := Min(18)
	v := 25
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_RangeRule_Min_invalid_int(b *testing.B) {
	r := Min(18)
	v := 15
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_RangeRule_Min_valid_float(b *testing.B) {
	r := Min(1.0)
	v := 3.14
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_RangeRule_Min_valid_time(b *testing.B) {
	cutoff := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	r := Min(cutoff)
	v := time.Now()
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

// --- LengthRule ---

func Benchmark_LengthRule_valid_string(b *testing.B) {
	r := Length(2, 50)
	v := "hello"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_LengthRule_invalid_string(b *testing.B) {
	r := Length(2, 50)
	v := "x"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_LengthRule_valid_slice(b *testing.B) {
	r := Length(1, 10)
	v := []int{1, 2, 3}
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

// --- EqualRule ---

func Benchmark_EqualRule_valid_int(b *testing.B) {
	r := Equal(42)
	v := 42
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_EqualRule_invalid_int(b *testing.B) {
	r := Equal(42)
	v := 99
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_EqualRule_valid_string(b *testing.B) {
	r := Equal("hello")
	v := "hello"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

// --- InRule ---

func Benchmark_InRule_valid_3(b *testing.B) {
	r := In("a", "b", "c")
	v := "b"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_InRule_invalid_3(b *testing.B) {
	r := In("a", "b", "c")
	v := "z"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_InRule_valid_10(b *testing.B) {
	r := In("a", "b", "c", "d", "e", "f", "g", "h", "i", "j")
	v := "j"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_InRule_invalid_10(b *testing.B) {
	r := In("a", "b", "c", "d", "e", "f", "g", "h", "i", "j")
	v := "z"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

// --- MatchRule ---

func Benchmark_MatchRule_valid(b *testing.B) {
	r := Match(benchEmailRx)
	v := "user@example.com"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_MatchRule_invalid(b *testing.B) {
	r := Match(benchEmailRx)
	v := "bad-email"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

// --- Validate (entry point, multiple rules) ---

func Benchmark_Validate_string_1rule(b *testing.B) {
	v := "hello@example.com"
	for b.Loop() {
		sinkErr = Validate(v, Required)
	}
}

func Benchmark_Validate_string_3rules(b *testing.B) {
	r := Match(benchEmailRx)
	v := "hello@example.com"
	for b.Loop() {
		sinkErr = Validate(v, Required, Length(5, 100), r)
	}
}

func Benchmark_Validate_int_3rules(b *testing.B) {
	v := 25
	for b.Loop() {
		sinkErr = Validate(v, Required, Min(18), Max(120))
	}
}

// --- ValidateStruct ---

func Benchmark_ValidateStruct_valid(b *testing.B) {
	r := Match(benchEmailRx)
	u := &benchUser{Name: "Alice", Email: "alice@example.com", Age: 30}
	for b.Loop() {
		sinkErr = ValidateStruct(u,
			Field(&u.Name, Required, Length(2, 50)),
			Field(&u.Email, Required, r),
			Field(&u.Age, Required, Min(18), Max(120)),
		)
	}
}

func Benchmark_ValidateStruct_invalid(b *testing.B) {
	r := Match(benchEmailRx)
	u := &benchUser{Name: "A", Email: "bad", Age: 15}
	for b.Loop() {
		sinkErr = ValidateStruct(u,
			Field(&u.Name, Required, Length(2, 50)),
			Field(&u.Email, Required, r),
			Field(&u.Age, Required, Min(18), Max(120)),
		)
	}
}

// --- StringOrBytes ---

func Benchmark_StringOrBytes_string(b *testing.B) {
	v := "hello"
	for b.Loop() {
		_, _, _, _ = StringOrBytes(v)
	}
}

func Benchmark_StringOrBytes_bytes(b *testing.B) {
	v := []byte("hello")
	for b.Loop() {
		_, _, _, _ = StringOrBytes(v)
	}
}

// --- IsEmpty (additional types) ---

func Benchmark_IsEmpty_bool(b *testing.B) {
	v := true
	for b.Loop() {
		sinkBool = IsEmpty(v)
	}
}

func Benchmark_IsEmpty_slice_nonempty(b *testing.B) {
	v := []int{1, 2, 3}
	for b.Loop() {
		sinkBool = IsEmpty(v)
	}
}

func Benchmark_IsEmpty_nilPointer(b *testing.B) {
	var v *string
	for b.Loop() {
		sinkBool = IsEmpty(v)
	}
}

// --- InRule (single element baseline) ---

func Benchmark_InRule_valid_1(b *testing.B) {
	r := In("a")
	v := "a"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_InRule_invalid_1(b *testing.B) {
	r := In("a")
	v := "z"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

// --- EachRule (slice iteration) ---

func Benchmark_EachRule_valid_3(b *testing.B) {
	r := Each(Required, Length(1, 20))
	v := []string{"foo", "bar", "baz"}
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_EachRule_invalid_3(b *testing.B) {
	r := Each(Required, Length(1, 20))
	v := []string{"foo", "", "baz"}
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_EachRule_valid_10(b *testing.B) {
	r := Each(Required, Length(1, 20))
	v := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

// --- ContainRule ---

func Benchmark_ContainRule_valid_3(b *testing.B) {
	r := Contain(Equal("b"))
	v := []string{"a", "b", "c"}
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_ContainRule_invalid_3(b *testing.B) {
	r := Contain(Equal("z"))
	v := []string{"a", "b", "c"}
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

// --- RangeRule (uint, to compare with int) ---

func Benchmark_RangeRule_Min_valid_uint(b *testing.B) {
	r := Min(uint(18))
	v := uint(25)
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

// --- LengthRule (rune counting) ---

func Benchmark_RuneLengthRule_valid_ascii(b *testing.B) {
	r := RuneLength(2, 50)
	v := "hello"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

func Benchmark_RuneLengthRule_valid_multibyte(b *testing.B) {
	r := RuneLength(2, 50)
	v := "héllo wörld"
	for b.Loop() {
		sinkErr = r.Validate(v)
	}
}

// --- Validate entry-point overhead ---

func Benchmark_Validate_norules(b *testing.B) {
	v := "hello"
	for b.Loop() {
		sinkErr = Validate(v)
	}
}

func Benchmark_Rule_direct_vs_Validate(b *testing.B) {
	r := Required
	v := "hello"
	b.Run("direct", func(b *testing.B) {
		for b.Loop() {
			sinkErr = r.Validate(v)
		}
	})
	b.Run("via_Validate", func(b *testing.B) {
		for b.Loop() {
			sinkErr = Validate(v, r)
		}
	})
}

// --- ValidateStruct scaling ---

type benchUser5 struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Email string `json:"email"`
	Age   int    `json:"age"`
	Zip   string `json:"zip"`
}

type benchUser10 struct {
	First   string `json:"first"`
	Last    string `json:"last"`
	Email   string `json:"email"`
	Age     int    `json:"age"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Country string `json:"country"`
	Phone   string `json:"phone"`
	Role    string `json:"role"`
	Active  string `json:"active"`
}

func Benchmark_ValidateStruct_5fields_valid(b *testing.B) {
	rx := benchEmailRx
	u := &benchUser5{
		First: "Alice",
		Last:  "Smith",
		Email: "alice@example.com",
		Age:   30,
		Zip:   "12345",
	}
	for b.Loop() {
		sinkErr = ValidateStruct(u,
			Field(&u.First, Required, Length(1, 50)),
			Field(&u.Last, Required, Length(1, 50)),
			Field(&u.Email, Required, Match(rx)),
			Field(&u.Age, Required, Min(18), Max(120)),
			Field(&u.Zip, Required, Length(5, 10)),
		)
	}
}

func Benchmark_ValidateStruct_10fields_valid(b *testing.B) {
	rx := benchEmailRx
	u := &benchUser10{
		First:   "Alice",
		Last:    "Smith",
		Email:   "alice@example.com",
		Age:     30,
		Zip:     "12345",
		City:    "Berlin",
		Country: "DE",
		Phone:   "+491234567",
		Role:    "admin",
		Active:  "yes",
	}
	for b.Loop() {
		sinkErr = ValidateStruct(u,
			Field(&u.First, Required, Length(1, 50)),
			Field(&u.Last, Required, Length(1, 50)),
			Field(&u.Email, Required, Match(rx)),
			Field(&u.Age, Required, Min(18), Max(120)),
			Field(&u.Zip, Required, Length(5, 10)),
			Field(&u.City, Required, Length(1, 100)),
			Field(&u.Country, Required, Length(2, 3)),
			Field(&u.Phone, Required, Length(7, 20)),
			Field(&u.Role, Required, Length(1, 50)),
			Field(&u.Active, Required),
		)
	}
}
