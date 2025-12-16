package ast

import (
	"bytes"
	"fmt"
	"strings"
)

// IntegerLiteral represents an integer constant
type IntegerLiteral struct {
	Line  int
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return fmt.Sprintf("%d", il.Value) }
func (il *IntegerLiteral) String() string       { return fmt.Sprintf("%d", il.Value) }

// FloatLiteral represents a floating-point constant
type FloatLiteral struct {
	Line  int
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fmt.Sprintf("%g", fl.Value) }
func (fl *FloatLiteral) String() string       { return fmt.Sprintf("%g", fl.Value) }

// StringLiteral represents a string constant
type StringLiteral struct {
	Line  int
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Value }
func (sl *StringLiteral) String() string       { return fmt.Sprintf("%q", sl.Value) }

// Identifier represents a variable reference
type Identifier struct {
	Line     int
	Name     string
	TypeHint DataType // derived from suffix (%, &, !, #, $)
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Name }
func (i *Identifier) String() string       { return i.Name }

// ArrayAccess represents array subscript access
type ArrayAccess struct {
	Line    int
	Name    string
	Indices []Expression
}

func (aa *ArrayAccess) expressionNode()      {}
func (aa *ArrayAccess) TokenLiteral() string { return aa.Name }
func (aa *ArrayAccess) String() string {
	var out bytes.Buffer
	out.WriteString(aa.Name)
	out.WriteString("(")
	indices := make([]string, len(aa.Indices))
	for i, idx := range aa.Indices {
		indices[i] = idx.String()
	}
	out.WriteString(strings.Join(indices, ", "))
	out.WriteString(")")
	return out.String()
}

// BinaryExpr represents a binary operation
type BinaryExpr struct {
	Line     int
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpr) expressionNode()      {}
func (be *BinaryExpr) TokenLiteral() string { return be.Operator }
func (be *BinaryExpr) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(be.Left.String())
	out.WriteString(" ")
	out.WriteString(be.Operator)
	out.WriteString(" ")
	out.WriteString(be.Right.String())
	out.WriteString(")")
	return out.String()
}

// UnaryExpr represents a unary operation
type UnaryExpr struct {
	Line     int
	Operator string
	Right    Expression
}

func (ue *UnaryExpr) expressionNode()      {}
func (ue *UnaryExpr) TokenLiteral() string { return ue.Operator }
func (ue *UnaryExpr) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ue.Operator)
	out.WriteString(ue.Right.String())
	out.WriteString(")")
	return out.String()
}

// CallExpr represents a function/sub call
type CallExpr struct {
	Line      int
	Function  string
	Arguments []Expression
}

func (ce *CallExpr) expressionNode()      {}
func (ce *CallExpr) TokenLiteral() string { return ce.Function }
func (ce *CallExpr) String() string {
	var out bytes.Buffer
	out.WriteString(ce.Function)
	out.WriteString("(")
	args := make([]string, len(ce.Arguments))
	for i, arg := range ce.Arguments {
		args[i] = arg.String()
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// GroupedExpr represents a parenthesized expression
type GroupedExpr struct {
	Line       int
	Expression Expression
}

func (ge *GroupedExpr) expressionNode()      {}
func (ge *GroupedExpr) TokenLiteral() string { return "(" }
func (ge *GroupedExpr) String() string {
	return "(" + ge.Expression.String() + ")"
}
