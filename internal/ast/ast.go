package ast

import (
	"bytes"
	"strings"
)

// DataType represents BASIC data types
type DataType int

const (
	TypeUnknown DataType = iota
	TypeInteger          // %
	TypeLong             // &
	TypeSingle           // !
	TypeDouble           // #
	TypeString           // $
)

func (dt DataType) String() string {
	switch dt {
	case TypeInteger:
		return "INTEGER"
	case TypeLong:
		return "LONG"
	case TypeSingle:
		return "SINGLE"
	case TypeDouble:
		return "DOUBLE"
	case TypeString:
		return "STRING"
	default:
		return "UNKNOWN"
	}
}

// TypeSuffix returns the type suffix character
func (dt DataType) Suffix() string {
	switch dt {
	case TypeInteger:
		return "%"
	case TypeLong:
		return "&"
	case TypeSingle:
		return "!"
	case TypeDouble:
		return "#"
	case TypeString:
		return "$"
	default:
		return ""
	}
}

// DataTypeFromSuffix returns the DataType from a suffix character
func DataTypeFromSuffix(suffix string) DataType {
	switch suffix {
	case "%":
		return TypeInteger
	case "&":
		return TypeLong
	case "!":
		return TypeSingle
	case "#":
		return TypeDouble
	case "$":
		return TypeString
	default:
		return TypeUnknown
	}
}

// Node is the base interface for all AST nodes
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a statement node
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST
type Program struct {
	Statements  []Statement
	Labels      map[string]int            // label name -> statement index
	LineNumbers map[int]int               // line number -> statement index
	DataItems   []Expression              // collected DATA items
	Subs        map[string]*SubStatement  // SUB definitions
	Functions   map[string]*FuncStatement // FUNCTION definitions
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	return out.String()
}

// NewProgram creates a new Program
func NewProgram() *Program {
	return &Program{
		Statements:  []Statement{},
		Labels:      make(map[string]int),
		LineNumbers: make(map[int]int),
		DataItems:   []Expression{},
		Subs:        make(map[string]*SubStatement),
		Functions:   make(map[string]*FuncStatement),
	}
}

// Parameter represents a SUB/FUNCTION parameter
type Parameter struct {
	Name     string
	DataType DataType
	ByVal    bool // if false, ByRef (default in QBasic)
}

func (p *Parameter) String() string {
	var out bytes.Buffer
	if p.ByVal {
		out.WriteString("BYVAL ")
	}
	out.WriteString(p.Name)
	if p.DataType != TypeUnknown {
		out.WriteString(" AS ")
		out.WriteString(p.DataType.String())
	}
	return out.String()
}

// DimVariable represents a variable in a DIM statement
type DimVariable struct {
	Name       string
	Dimensions []Expression // nil for scalar, expressions for array bounds
	DataType   DataType
}

func (dv *DimVariable) String() string {
	var out bytes.Buffer
	out.WriteString(dv.Name)
	if len(dv.Dimensions) > 0 {
		out.WriteString("(")
		dims := make([]string, len(dv.Dimensions))
		for i, d := range dv.Dimensions {
			dims[i] = d.String()
		}
		out.WriteString(strings.Join(dims, ", "))
		out.WriteString(")")
	}
	if dv.DataType != TypeUnknown {
		out.WriteString(" AS ")
		out.WriteString(dv.DataType.String())
	}
	return out.String()
}

// PrintItem represents an item in a PRINT statement
type PrintItem struct {
	Expression Expression
	Separator  string // "", ";", or ","
}

// CaseClause represents a CASE clause in SELECT CASE
type CaseClause struct {
	Values []CaseValue
	Body   []Statement
}

// CaseValue represents a value in a CASE clause
type CaseValue struct {
	Type     string     // "SINGLE", "RANGE", "IS"
	Value    Expression // for SINGLE and RANGE (start)
	EndValue Expression // for RANGE (end)
	Operator string     // for IS (<, >, =, etc.)
}

func (cv *CaseValue) String() string {
	switch cv.Type {
	case "RANGE":
		return cv.Value.String() + " TO " + cv.EndValue.String()
	case "IS":
		return "IS " + cv.Operator + " " + cv.Value.String()
	default:
		return cv.Value.String()
	}
}
