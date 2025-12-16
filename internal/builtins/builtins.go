package builtins

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/xbasic/xbasic/internal/ast"
)

// Value interface for runtime values
type Value interface {
	Type() ast.DataType
	String() string
	ToFloat() float64
	ToInt() int64
	ToBool() bool
	ToString() string
}

// IntegerValue represents an integer
type IntegerValue struct{ Val int16 }

func (v *IntegerValue) Type() ast.DataType { return ast.TypeInteger }
func (v *IntegerValue) String() string     { return fmt.Sprintf("%d", v.Val) }
func (v *IntegerValue) ToFloat() float64   { return float64(v.Val) }
func (v *IntegerValue) ToInt() int64       { return int64(v.Val) }
func (v *IntegerValue) ToBool() bool       { return v.Val != 0 }
func (v *IntegerValue) ToString() string   { return fmt.Sprintf("%d", v.Val) }

// LongValue represents a long integer
type LongValue struct{ Val int32 }

func (v *LongValue) Type() ast.DataType { return ast.TypeLong }
func (v *LongValue) String() string     { return fmt.Sprintf("%d", v.Val) }
func (v *LongValue) ToFloat() float64   { return float64(v.Val) }
func (v *LongValue) ToInt() int64       { return int64(v.Val) }
func (v *LongValue) ToBool() bool       { return v.Val != 0 }
func (v *LongValue) ToString() string   { return fmt.Sprintf("%d", v.Val) }

// SingleValue represents a single-precision float
type SingleValue struct{ Val float32 }

func (v *SingleValue) Type() ast.DataType { return ast.TypeSingle }
func (v *SingleValue) String() string     { return fmt.Sprintf("%g", v.Val) }
func (v *SingleValue) ToFloat() float64   { return float64(v.Val) }
func (v *SingleValue) ToInt() int64       { return int64(v.Val) }
func (v *SingleValue) ToBool() bool       { return v.Val != 0 }
func (v *SingleValue) ToString() string   { return fmt.Sprintf("%g", v.Val) }

// DoubleValue represents a double-precision float
type DoubleValue struct{ Val float64 }

func (v *DoubleValue) Type() ast.DataType { return ast.TypeDouble }
func (v *DoubleValue) String() string     { return fmt.Sprintf("%g", v.Val) }
func (v *DoubleValue) ToFloat() float64   { return v.Val }
func (v *DoubleValue) ToInt() int64       { return int64(v.Val) }
func (v *DoubleValue) ToBool() bool       { return v.Val != 0 }
func (v *DoubleValue) ToString() string   { return fmt.Sprintf("%g", v.Val) }

// StringValue represents a string
type StringValue struct{ Val string }

func (v *StringValue) Type() ast.DataType { return ast.TypeString }
func (v *StringValue) String() string     { return v.Val }
func (v *StringValue) ToFloat() float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(v.Val), 64)
	return f
}
func (v *StringValue) ToInt() int64 {
	i, _ := strconv.ParseInt(strings.TrimSpace(v.Val), 10, 64)
	return i
}
func (v *StringValue) ToBool() bool    { return v.Val != "" }
func (v *StringValue) ToString() string { return v.Val }

// BuiltinFunc is the signature for built-in functions
type BuiltinFunc func(args []Value) (Value, error)

// Registry holds all built-in functions
type Registry struct {
	functions map[string]BuiltinFunc
	rng       *rand.Rand
}

// NewRegistry creates a new function registry with all built-ins
func NewRegistry() *Registry {
	r := &Registry{
		functions: make(map[string]BuiltinFunc),
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	r.registerAll()
	return r
}

// Call invokes a built-in function
func (r *Registry) Call(name string, args []Value) (Value, error) {
	fn, ok := r.functions[name]
	if !ok {
		return nil, fmt.Errorf("unknown function: %s", name)
	}
	return fn(args)
}

// SetRandomSeed sets the random number generator seed
func (r *Registry) SetRandomSeed(seed int64) {
	r.rng = rand.New(rand.NewSource(seed))
}

// RandomizeSeed randomizes the seed based on current time
func (r *Registry) RandomizeSeed() {
	r.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (r *Registry) registerAll() {
	// String functions
	r.functions["LEN"] = r.fnLen
	r.functions["LEFT$"] = r.fnLeft
	r.functions["RIGHT$"] = r.fnRight
	r.functions["MID$"] = r.fnMid
	r.functions["INSTR"] = r.fnInstr
	r.functions["UCASE$"] = r.fnUCase
	r.functions["LCASE$"] = r.fnLCase
	r.functions["STR$"] = r.fnStr
	r.functions["VAL"] = r.fnVal
	r.functions["CHR$"] = r.fnChr
	r.functions["ASC"] = r.fnAsc
	r.functions["STRING$"] = r.fnString
	r.functions["SPACE$"] = r.fnSpace
	r.functions["LTRIM$"] = r.fnLTrim
	r.functions["RTRIM$"] = r.fnRTrim
	r.functions["TRIM$"] = r.fnTrim

	// Math functions
	r.functions["ABS"] = r.fnAbs
	r.functions["SGN"] = r.fnSgn
	r.functions["INT"] = r.fnInt
	r.functions["FIX"] = r.fnFix
	r.functions["SQR"] = r.fnSqr
	r.functions["SIN"] = r.fnSin
	r.functions["COS"] = r.fnCos
	r.functions["TAN"] = r.fnTan
	r.functions["ATN"] = r.fnAtn
	r.functions["LOG"] = r.fnLog
	r.functions["EXP"] = r.fnExp
	r.functions["RND"] = r.fnRnd

	// Date/Time functions
	r.functions["TIMER"] = r.fnTimer
	r.functions["DATE$"] = r.fnDate
	r.functions["TIME$"] = r.fnTime

	// Conversion functions
	r.functions["CINT"] = r.fnCInt
	r.functions["CLNG"] = r.fnCLng
	r.functions["CSNG"] = r.fnCSng
	r.functions["CDBL"] = r.fnCDbl

	// Other functions
	r.functions["HEX$"] = r.fnHex
	r.functions["OCT$"] = r.fnOct
	r.functions["TAB"] = r.fnTab
	r.functions["SPC"] = r.fnSpc

	// Additional math functions
	r.functions["ATAN2"] = r.fnAtan2
	r.functions["ATN2"] = r.fnAtan2 // alias
	r.functions["ROUND"] = r.fnRound
	r.functions["_PI"] = r.fnPi
	r.functions["PI"] = r.fnPi // alias without underscore
}

// String functions

func (r *Registry) fnLen(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("LEN requires 1 argument")
	}
	return &LongValue{Val: int32(len(args[0].ToString()))}, nil
}

func (r *Registry) fnLeft(args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("LEFT$ requires 2 arguments")
	}
	s := args[0].ToString()
	n := int(args[1].ToInt())
	if n < 0 {
		return nil, fmt.Errorf("illegal function call")
	}
	if n > len(s) {
		n = len(s)
	}
	return &StringValue{Val: s[:n]}, nil
}

func (r *Registry) fnRight(args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("RIGHT$ requires 2 arguments")
	}
	s := args[0].ToString()
	n := int(args[1].ToInt())
	if n < 0 {
		return nil, fmt.Errorf("illegal function call")
	}
	if n > len(s) {
		n = len(s)
	}
	return &StringValue{Val: s[len(s)-n:]}, nil
}

func (r *Registry) fnMid(args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("MID$ requires at least 2 arguments")
	}
	s := args[0].ToString()
	start := int(args[1].ToInt()) - 1 // QBasic is 1-indexed

	if start < 0 {
		return nil, fmt.Errorf("illegal function call")
	}
	if start >= len(s) {
		return &StringValue{Val: ""}, nil
	}

	length := len(s) - start
	if len(args) >= 3 {
		length = int(args[2].ToInt())
		if length < 0 {
			return nil, fmt.Errorf("illegal function call")
		}
	}

	end := start + length
	if end > len(s) {
		end = len(s)
	}

	return &StringValue{Val: s[start:end]}, nil
}

func (r *Registry) fnInstr(args []Value) (Value, error) {
	var start int = 0
	var s1, s2 string

	if len(args) == 2 {
		s1 = args[0].ToString()
		s2 = args[1].ToString()
	} else if len(args) >= 3 {
		start = int(args[0].ToInt()) - 1
		s1 = args[1].ToString()
		s2 = args[2].ToString()
	} else {
		return nil, fmt.Errorf("INSTR requires 2 or 3 arguments")
	}

	if start < 0 {
		return nil, fmt.Errorf("illegal function call")
	}

	if start >= len(s1) {
		return &LongValue{Val: 0}, nil
	}

	idx := strings.Index(s1[start:], s2)
	if idx == -1 {
		return &LongValue{Val: 0}, nil
	}
	return &LongValue{Val: int32(idx + start + 1)}, nil // 1-indexed
}

func (r *Registry) fnUCase(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("UCASE$ requires 1 argument")
	}
	return &StringValue{Val: strings.ToUpper(args[0].ToString())}, nil
}

func (r *Registry) fnLCase(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("LCASE$ requires 1 argument")
	}
	return &StringValue{Val: strings.ToLower(args[0].ToString())}, nil
}

func (r *Registry) fnStr(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("STR$ requires 1 argument")
	}
	v := args[0]
	s := v.ToString()
	// QBasic adds leading space for positive numbers
	if v.ToFloat() >= 0 {
		s = " " + s
	}
	return &StringValue{Val: s}, nil
}

func (r *Registry) fnVal(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("VAL requires 1 argument")
	}
	s := strings.TrimSpace(args[0].ToString())
	f, _ := strconv.ParseFloat(s, 64)
	return &DoubleValue{Val: f}, nil
}

func (r *Registry) fnChr(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("CHR$ requires 1 argument")
	}
	n := args[0].ToInt()
	if n < 0 || n > 255 {
		return nil, fmt.Errorf("illegal function call")
	}
	return &StringValue{Val: string(rune(n))}, nil
}

func (r *Registry) fnAsc(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ASC requires 1 argument")
	}
	s := args[0].ToString()
	if len(s) == 0 {
		return nil, fmt.Errorf("illegal function call")
	}
	return &LongValue{Val: int32(s[0])}, nil
}

func (r *Registry) fnString(args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("STRING$ requires 2 arguments")
	}
	n := int(args[0].ToInt())
	if n < 0 {
		return nil, fmt.Errorf("illegal function call")
	}

	var ch string
	if args[1].Type() == ast.TypeString {
		s := args[1].ToString()
		if len(s) > 0 {
			ch = string(s[0])
		} else {
			ch = " "
		}
	} else {
		code := args[1].ToInt()
		if code < 0 || code > 255 {
			return nil, fmt.Errorf("illegal function call")
		}
		ch = string(rune(code))
	}

	return &StringValue{Val: strings.Repeat(ch, n)}, nil
}

func (r *Registry) fnSpace(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("SPACE$ requires 1 argument")
	}
	n := int(args[0].ToInt())
	if n < 0 {
		return nil, fmt.Errorf("illegal function call")
	}
	return &StringValue{Val: strings.Repeat(" ", n)}, nil
}

func (r *Registry) fnLTrim(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("LTRIM$ requires 1 argument")
	}
	return &StringValue{Val: strings.TrimLeft(args[0].ToString(), " ")}, nil
}

func (r *Registry) fnRTrim(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("RTRIM$ requires 1 argument")
	}
	return &StringValue{Val: strings.TrimRight(args[0].ToString(), " ")}, nil
}

func (r *Registry) fnTrim(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("TRIM$ requires 1 argument")
	}
	return &StringValue{Val: strings.TrimSpace(args[0].ToString())}, nil
}

// Math functions

func (r *Registry) fnAbs(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ABS requires 1 argument")
	}
	return &DoubleValue{Val: math.Abs(args[0].ToFloat())}, nil
}

func (r *Registry) fnSgn(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("SGN requires 1 argument")
	}
	v := args[0].ToFloat()
	if v > 0 {
		return &IntegerValue{Val: 1}, nil
	} else if v < 0 {
		return &IntegerValue{Val: -1}, nil
	}
	return &IntegerValue{Val: 0}, nil
}

func (r *Registry) fnInt(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("INT requires 1 argument")
	}
	return &DoubleValue{Val: math.Floor(args[0].ToFloat())}, nil
}

func (r *Registry) fnFix(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("FIX requires 1 argument")
	}
	return &DoubleValue{Val: math.Trunc(args[0].ToFloat())}, nil
}

func (r *Registry) fnSqr(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("SQR requires 1 argument")
	}
	v := args[0].ToFloat()
	if v < 0 {
		return nil, fmt.Errorf("illegal function call")
	}
	return &DoubleValue{Val: math.Sqrt(v)}, nil
}

func (r *Registry) fnSin(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("SIN requires 1 argument")
	}
	return &DoubleValue{Val: math.Sin(args[0].ToFloat())}, nil
}

func (r *Registry) fnCos(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("COS requires 1 argument")
	}
	return &DoubleValue{Val: math.Cos(args[0].ToFloat())}, nil
}

func (r *Registry) fnTan(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("TAN requires 1 argument")
	}
	return &DoubleValue{Val: math.Tan(args[0].ToFloat())}, nil
}

func (r *Registry) fnAtn(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ATN requires 1 argument")
	}
	return &DoubleValue{Val: math.Atan(args[0].ToFloat())}, nil
}

func (r *Registry) fnLog(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("LOG requires 1 argument")
	}
	v := args[0].ToFloat()
	if v <= 0 {
		return nil, fmt.Errorf("illegal function call")
	}
	return &DoubleValue{Val: math.Log(v)}, nil
}

func (r *Registry) fnExp(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("EXP requires 1 argument")
	}
	return &DoubleValue{Val: math.Exp(args[0].ToFloat())}, nil
}

func (r *Registry) fnRnd(args []Value) (Value, error) {
	// RND with no args or positive arg returns next random number
	// RND(0) returns previous random number (not implemented)
	// RND(negative) seeds and returns
	if len(args) > 0 {
		n := args[0].ToFloat()
		if n < 0 {
			r.rng = rand.New(rand.NewSource(int64(n)))
		}
	}
	return &SingleValue{Val: float32(r.rng.Float64())}, nil
}

// Date/Time functions

func (r *Registry) fnTimer(args []Value) (Value, error) {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	seconds := now.Sub(midnight).Seconds()
	return &SingleValue{Val: float32(seconds)}, nil
}

func (r *Registry) fnDate(args []Value) (Value, error) {
	now := time.Now()
	return &StringValue{Val: now.Format("01-02-2006")}, nil
}

func (r *Registry) fnTime(args []Value) (Value, error) {
	now := time.Now()
	return &StringValue{Val: now.Format("15:04:05")}, nil
}

// Conversion functions

func (r *Registry) fnCInt(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("CINT requires 1 argument")
	}
	v := args[0].ToFloat()
	// Round to nearest, ties to even (banker's rounding)
	return &IntegerValue{Val: int16(math.RoundToEven(v))}, nil
}

func (r *Registry) fnCLng(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("CLNG requires 1 argument")
	}
	v := args[0].ToFloat()
	return &LongValue{Val: int32(math.RoundToEven(v))}, nil
}

func (r *Registry) fnCSng(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("CSNG requires 1 argument")
	}
	return &SingleValue{Val: float32(args[0].ToFloat())}, nil
}

func (r *Registry) fnCDbl(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("CDBL requires 1 argument")
	}
	return &DoubleValue{Val: args[0].ToFloat()}, nil
}

// Other functions

func (r *Registry) fnHex(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("HEX$ requires 1 argument")
	}
	n := args[0].ToInt()
	return &StringValue{Val: fmt.Sprintf("%X", n)}, nil
}

func (r *Registry) fnOct(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("OCT$ requires 1 argument")
	}
	n := args[0].ToInt()
	return &StringValue{Val: fmt.Sprintf("%o", n)}, nil
}

func (r *Registry) fnTab(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("TAB requires 1 argument")
	}
	n := int(args[0].ToInt())
	if n < 1 {
		n = 1
	}
	// TAB returns spaces to reach column n
	// This is a simplification - actual TAB depends on current cursor position
	return &StringValue{Val: strings.Repeat(" ", n-1)}, nil
}

func (r *Registry) fnSpc(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("SPC requires 1 argument")
	}
	n := int(args[0].ToInt())
	if n < 0 {
		n = 0
	}
	return &StringValue{Val: strings.Repeat(" ", n)}, nil
}

// Additional math functions

func (r *Registry) fnAtan2(args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("ATAN2 requires 2 arguments")
	}
	y := args[0].ToFloat()
	x := args[1].ToFloat()
	return &DoubleValue{Val: math.Atan2(y, x)}, nil
}

func (r *Registry) fnRound(args []Value) (Value, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ROUND requires at least 1 argument")
	}
	v := args[0].ToFloat()
	decimals := 0
	if len(args) >= 2 {
		decimals = int(args[1].ToInt())
	}
	multiplier := math.Pow(10, float64(decimals))
	return &DoubleValue{Val: math.Round(v*multiplier) / multiplier}, nil
}

func (r *Registry) fnPi(args []Value) (Value, error) {
	return &DoubleValue{Val: math.Pi}, nil
}
