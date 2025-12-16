package interpreter

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/xbasic/xbasic/internal/ast"
)

// Value represents a runtime value
type Value interface {
	Type() ast.DataType
	String() string
	Clone() Value
	ToFloat() float64
	ToInt() int64
	ToBool() bool
	ToString() string
}

// IntegerValue represents a 16-bit signed integer (INTEGER / %)
type IntegerValue struct {
	Val int16
}

func (iv *IntegerValue) Type() ast.DataType { return ast.TypeInteger }
func (iv *IntegerValue) String() string     { return fmt.Sprintf("%d", iv.Val) }
func (iv *IntegerValue) Clone() Value       { return &IntegerValue{Val: iv.Val} }
func (iv *IntegerValue) ToFloat() float64   { return float64(iv.Val) }
func (iv *IntegerValue) ToInt() int64       { return int64(iv.Val) }
func (iv *IntegerValue) ToBool() bool       { return iv.Val != 0 }
func (iv *IntegerValue) ToString() string   { return fmt.Sprintf("%d", iv.Val) }

// LongValue represents a 32-bit signed integer (LONG / &)
type LongValue struct {
	Val int32
}

func (lv *LongValue) Type() ast.DataType { return ast.TypeLong }
func (lv *LongValue) String() string     { return fmt.Sprintf("%d", lv.Val) }
func (lv *LongValue) Clone() Value       { return &LongValue{Val: lv.Val} }
func (lv *LongValue) ToFloat() float64   { return float64(lv.Val) }
func (lv *LongValue) ToInt() int64       { return int64(lv.Val) }
func (lv *LongValue) ToBool() bool       { return lv.Val != 0 }
func (lv *LongValue) ToString() string   { return fmt.Sprintf("%d", lv.Val) }

// SingleValue represents a 32-bit float (SINGLE / !)
type SingleValue struct {
	Val float32
}

func (sv *SingleValue) Type() ast.DataType { return ast.TypeSingle }
func (sv *SingleValue) String() string     { return formatFloat(float64(sv.Val)) }
func (sv *SingleValue) Clone() Value       { return &SingleValue{Val: sv.Val} }
func (sv *SingleValue) ToFloat() float64   { return float64(sv.Val) }
func (sv *SingleValue) ToInt() int64       { return int64(sv.Val) }
func (sv *SingleValue) ToBool() bool       { return sv.Val != 0 }
func (sv *SingleValue) ToString() string   { return formatFloat(float64(sv.Val)) }

// DoubleValue represents a 64-bit float (DOUBLE / #)
type DoubleValue struct {
	Val float64
}

func (dv *DoubleValue) Type() ast.DataType { return ast.TypeDouble }
func (dv *DoubleValue) String() string     { return formatFloat(dv.Val) }
func (dv *DoubleValue) Clone() Value       { return &DoubleValue{Val: dv.Val} }
func (dv *DoubleValue) ToFloat() float64   { return dv.Val }
func (dv *DoubleValue) ToInt() int64       { return int64(dv.Val) }
func (dv *DoubleValue) ToBool() bool       { return dv.Val != 0 }
func (dv *DoubleValue) ToString() string   { return formatFloat(dv.Val) }

// StringValue represents a string (STRING / $)
type StringValue struct {
	Val string
}

func (sv *StringValue) Type() ast.DataType { return ast.TypeString }
func (sv *StringValue) String() string     { return sv.Val }
func (sv *StringValue) Clone() Value       { return &StringValue{Val: sv.Val} }
func (sv *StringValue) ToFloat() float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(sv.Val), 64)
	return v
}
func (sv *StringValue) ToInt() int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(sv.Val), 10, 64)
	return v
}
func (sv *StringValue) ToBool() bool     { return sv.Val != "" }
func (sv *StringValue) ToString() string { return sv.Val }

// Array represents a BASIC array
type Array struct {
	DataType   ast.DataType
	Dimensions []ArrayDimension
	Data       []Value
}

// ArrayDimension represents array bounds
type ArrayDimension struct {
	Lower int // typically 0 (OPTION BASE 0) or 1 (OPTION BASE 1)
	Upper int
}

// NewArray creates a new array with given dimensions
func NewArray(dt ast.DataType, dims []ArrayDimension) *Array {
	size := 1
	for _, d := range dims {
		size *= (d.Upper - d.Lower + 1)
	}

	arr := &Array{
		DataType:   dt,
		Dimensions: dims,
		Data:       make([]Value, size),
	}

	// Initialize with default values
	defaultVal := DefaultValue(dt)
	for i := range arr.Data {
		arr.Data[i] = defaultVal.Clone()
	}

	return arr
}

// GetIndex calculates the linear index from subscripts
func (a *Array) GetIndex(subscripts []int) (int, error) {
	if len(subscripts) != len(a.Dimensions) {
		return 0, fmt.Errorf("wrong number of dimensions: expected %d, got %d",
			len(a.Dimensions), len(subscripts))
	}

	index := 0
	multiplier := 1

	for i := len(a.Dimensions) - 1; i >= 0; i-- {
		sub := subscripts[i]
		dim := a.Dimensions[i]

		if sub < dim.Lower || sub > dim.Upper {
			return 0, fmt.Errorf("subscript out of range: %d not in [%d, %d]",
				sub, dim.Lower, dim.Upper)
		}

		index += (sub - dim.Lower) * multiplier
		multiplier *= (dim.Upper - dim.Lower + 1)
	}

	return index, nil
}

// Get retrieves value at given subscripts
func (a *Array) Get(subscripts []int) (Value, error) {
	index, err := a.GetIndex(subscripts)
	if err != nil {
		return nil, err
	}
	return a.Data[index], nil
}

// Set stores value at given subscripts
func (a *Array) Set(subscripts []int, value Value) error {
	index, err := a.GetIndex(subscripts)
	if err != nil {
		return err
	}
	a.Data[index] = value
	return nil
}

// Helper functions

// DefaultValue returns the default value for a data type
func DefaultValue(dt ast.DataType) Value {
	switch dt {
	case ast.TypeInteger:
		return &IntegerValue{Val: 0}
	case ast.TypeLong:
		return &LongValue{Val: 0}
	case ast.TypeSingle:
		return &SingleValue{Val: 0}
	case ast.TypeDouble:
		return &DoubleValue{Val: 0}
	case ast.TypeString:
		return &StringValue{Val: ""}
	default:
		return &SingleValue{Val: 0} // default to SINGLE
	}
}

// NewValue creates a new value of the given type from a Go value
func NewValue(dt ast.DataType, val interface{}) Value {
	switch dt {
	case ast.TypeInteger:
		switch v := val.(type) {
		case int:
			return &IntegerValue{Val: int16(v)}
		case int64:
			return &IntegerValue{Val: int16(v)}
		case float64:
			return &IntegerValue{Val: int16(v)}
		}
	case ast.TypeLong:
		switch v := val.(type) {
		case int:
			return &LongValue{Val: int32(v)}
		case int64:
			return &LongValue{Val: int32(v)}
		case float64:
			return &LongValue{Val: int32(v)}
		}
	case ast.TypeSingle:
		switch v := val.(type) {
		case float32:
			return &SingleValue{Val: v}
		case float64:
			return &SingleValue{Val: float32(v)}
		case int64:
			return &SingleValue{Val: float32(v)}
		}
	case ast.TypeDouble:
		switch v := val.(type) {
		case float64:
			return &DoubleValue{Val: v}
		case float32:
			return &DoubleValue{Val: float64(v)}
		case int64:
			return &DoubleValue{Val: float64(v)}
		}
	case ast.TypeString:
		if v, ok := val.(string); ok {
			return &StringValue{Val: v}
		}
	}
	return DefaultValue(dt)
}

// CoerceValue converts a value to the target type
func CoerceValue(val Value, targetType ast.DataType) Value {
	if val.Type() == targetType {
		return val
	}

	switch targetType {
	case ast.TypeInteger:
		return &IntegerValue{Val: int16(val.ToInt())}
	case ast.TypeLong:
		return &LongValue{Val: int32(val.ToInt())}
	case ast.TypeSingle:
		return &SingleValue{Val: float32(val.ToFloat())}
	case ast.TypeDouble:
		return &DoubleValue{Val: val.ToFloat()}
	case ast.TypeString:
		return &StringValue{Val: val.ToString()}
	}

	return val
}

// PromoteType returns the wider of two numeric types
func PromoteType(t1, t2 ast.DataType) ast.DataType {
	// STRING cannot be promoted with numbers
	if t1 == ast.TypeString || t2 == ast.TypeString {
		return ast.TypeString
	}

	// Promotion order: INTEGER < LONG < SINGLE < DOUBLE
	order := map[ast.DataType]int{
		ast.TypeInteger: 1,
		ast.TypeLong:    2,
		ast.TypeSingle:  3,
		ast.TypeDouble:  4,
	}

	if order[t1] > order[t2] {
		return t1
	}
	return t2
}

// formatFloat formats a float for QBasic-style output
func formatFloat(f float64) string {
	if f == float64(int64(f)) && math.Abs(f) < 1e15 {
		return fmt.Sprintf("%d", int64(f))
	}
	s := strconv.FormatFloat(f, 'G', -1, 64)
	return s
}

// IsNumeric returns true if the value is a numeric type
func IsNumeric(v Value) bool {
	switch v.Type() {
	case ast.TypeInteger, ast.TypeLong, ast.TypeSingle, ast.TypeDouble:
		return true
	}
	return false
}

// Compare compares two values, returns -1, 0, or 1
func Compare(a, b Value) int {
	// String comparison
	if a.Type() == ast.TypeString && b.Type() == ast.TypeString {
		as := a.(*StringValue).Val
		bs := b.(*StringValue).Val
		if as < bs {
			return -1
		} else if as > bs {
			return 1
		}
		return 0
	}

	// Numeric comparison
	af := a.ToFloat()
	bf := b.ToFloat()
	if af < bf {
		return -1
	} else if af > bf {
		return 1
	}
	return 0
}
