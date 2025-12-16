package interpreter

import (
	"fmt"
	"strings"

	"github.com/xbasic/xbasic/internal/ast"
)

// Environment manages variable scopes
type Environment struct {
	variables map[string]Value
	arrays    map[string]*Array
	constants map[string]Value
	parent    *Environment // for SUB/FUNCTION scope
	shared    *Environment // module-level shared variables
}

// NewEnvironment creates a new global environment
func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		arrays:    make(map[string]*Array),
		constants: make(map[string]Value),
	}
}

// NewEnclosedEnvironment creates a new local scope
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.parent = outer
	env.shared = outer
	return env
}

// Get retrieves a variable value
func (e *Environment) Get(name string) (Value, bool) {
	name = strings.ToUpper(name)

	// Check constants first
	if val, ok := e.constants[name]; ok {
		return val, true
	}

	// Check local variables
	if val, ok := e.variables[name]; ok {
		return val, true
	}

	// Check parent scope
	if e.parent != nil {
		return e.parent.Get(name)
	}

	return nil, false
}

// Set stores a variable value
func (e *Environment) Set(name string, val Value) {
	name = strings.ToUpper(name)

	// Don't allow modifying constants
	if _, ok := e.constants[name]; ok {
		return
	}

	// If variable exists in parent and we're in local scope, create local copy
	e.variables[name] = val
}

// SetShared stores a variable in the shared (module-level) scope
func (e *Environment) SetShared(name string, val Value) {
	name = strings.ToUpper(name)
	if e.shared != nil {
		e.shared.variables[name] = val
	} else {
		e.variables[name] = val
	}
}

// GetOrCreate gets an existing variable or creates a new one with default value
func (e *Environment) GetOrCreate(name string, dt ast.DataType) Value {
	name = strings.ToUpper(name)

	if val, ok := e.Get(name); ok {
		return val
	}

	// Create with default value based on type suffix in name
	if dt == ast.TypeUnknown {
		dt = e.inferType(name)
	}

	val := DefaultValue(dt)
	e.Set(name, val)
	return val
}

// inferType infers the data type from a variable name's suffix
func (e *Environment) inferType(name string) ast.DataType {
	if len(name) == 0 {
		return ast.TypeSingle // default
	}

	lastChar := name[len(name)-1:]
	switch lastChar {
	case "%":
		return ast.TypeInteger
	case "&":
		return ast.TypeLong
	case "!":
		return ast.TypeSingle
	case "#":
		return ast.TypeDouble
	case "$":
		return ast.TypeString
	default:
		return ast.TypeSingle // default is SINGLE in QBasic
	}
}

// DefineConst defines a constant
func (e *Environment) DefineConst(name string, val Value) error {
	name = strings.ToUpper(name)

	if _, ok := e.constants[name]; ok {
		return fmt.Errorf("constant %s already defined", name)
	}

	e.constants[name] = val
	return nil
}

// GetArray retrieves an array
func (e *Environment) GetArray(name string) (*Array, bool) {
	name = strings.ToUpper(name)

	if arr, ok := e.arrays[name]; ok {
		return arr, true
	}

	if e.parent != nil {
		return e.parent.GetArray(name)
	}

	return nil, false
}

// SetArray stores an array
func (e *Environment) SetArray(name string, arr *Array) {
	name = strings.ToUpper(name)
	e.arrays[name] = arr
}

// DeclareArray declares a new array with given dimensions
func (e *Environment) DeclareArray(name string, dt ast.DataType, dims []int) *Array {
	name = strings.ToUpper(name)

	// Convert dims to ArrayDimension (0 to dims[i])
	adims := make([]ArrayDimension, len(dims))
	for i, d := range dims {
		adims[i] = ArrayDimension{Lower: 0, Upper: d}
	}

	arr := NewArray(dt, adims)
	e.arrays[name] = arr
	return arr
}

// ExecutionState tracks program execution
type ExecutionState struct {
	ProgramCounter int         // current statement index
	CallStack      []CallFrame // call stack for GOSUB/SUB/FUNCTION
	DataPointer    int         // current position in DATA items
	ForStack       []ForFrame  // stack for FOR loops
	DoStack        []DoFrame   // stack for DO loops
	WhileStack     []WhileFrame // stack for WHILE loops
	Running        bool        // is program running?
	StepMode       bool        // single-step mode
}

// CallFrame represents a stack frame for subroutine calls
type CallFrame struct {
	ReturnIndex int          // statement index to return to
	LocalEnv    *Environment // local variables
	Type        string       // "GOSUB", "SUB", or "FUNCTION"
	FuncName    string       // function name (for assigning return value)
}

// ForFrame tracks FOR loop state
type ForFrame struct {
	Variable   string
	EndValue   Value
	StepValue  Value
	BodyStart  int // statement index of first body statement
	BodyEnd    int // statement index after NEXT
	StepSign   int // 1 or -1
}

// DoFrame tracks DO loop state
type DoFrame struct {
	LoopStart int // statement index of DO
	LoopEnd   int // statement index after LOOP
}

// WhileFrame tracks WHILE loop state
type WhileFrame struct {
	LoopStart int // statement index of WHILE
	LoopEnd   int // statement index after WEND
}

// NewExecutionState creates a new execution state
func NewExecutionState() *ExecutionState {
	return &ExecutionState{
		ProgramCounter: 0,
		CallStack:      []CallFrame{},
		DataPointer:    0,
		ForStack:       []ForFrame{},
		DoStack:        []DoFrame{},
		WhileStack:     []WhileFrame{},
		Running:        false,
		StepMode:       false,
	}
}

// PushCall pushes a call frame onto the stack
func (es *ExecutionState) PushCall(frame CallFrame) {
	es.CallStack = append(es.CallStack, frame)
}

// PopCall pops a call frame from the stack
func (es *ExecutionState) PopCall() (CallFrame, bool) {
	if len(es.CallStack) == 0 {
		return CallFrame{}, false
	}
	frame := es.CallStack[len(es.CallStack)-1]
	es.CallStack = es.CallStack[:len(es.CallStack)-1]
	return frame, true
}

// PushFor pushes a FOR frame onto the stack
func (es *ExecutionState) PushFor(frame ForFrame) {
	es.ForStack = append(es.ForStack, frame)
}

// PopFor pops a FOR frame from the stack
func (es *ExecutionState) PopFor() (ForFrame, bool) {
	if len(es.ForStack) == 0 {
		return ForFrame{}, false
	}
	frame := es.ForStack[len(es.ForStack)-1]
	es.ForStack = es.ForStack[:len(es.ForStack)-1]
	return frame, true
}

// PeekFor returns the top FOR frame without popping
func (es *ExecutionState) PeekFor() (ForFrame, bool) {
	if len(es.ForStack) == 0 {
		return ForFrame{}, false
	}
	return es.ForStack[len(es.ForStack)-1], true
}

// UpdateFor updates the top FOR frame
func (es *ExecutionState) UpdateFor(frame ForFrame) {
	if len(es.ForStack) > 0 {
		es.ForStack[len(es.ForStack)-1] = frame
	}
}

// PushDo pushes a DO frame onto the stack
func (es *ExecutionState) PushDo(frame DoFrame) {
	es.DoStack = append(es.DoStack, frame)
}

// PopDo pops a DO frame from the stack
func (es *ExecutionState) PopDo() (DoFrame, bool) {
	if len(es.DoStack) == 0 {
		return DoFrame{}, false
	}
	frame := es.DoStack[len(es.DoStack)-1]
	es.DoStack = es.DoStack[:len(es.DoStack)-1]
	return frame, true
}

// PushWhile pushes a WHILE frame onto the stack
func (es *ExecutionState) PushWhile(frame WhileFrame) {
	es.WhileStack = append(es.WhileStack, frame)
}

// PopWhile pops a WHILE frame from the stack
func (es *ExecutionState) PopWhile() (WhileFrame, bool) {
	if len(es.WhileStack) == 0 {
		return WhileFrame{}, false
	}
	frame := es.WhileStack[len(es.WhileStack)-1]
	es.WhileStack = es.WhileStack[:len(es.WhileStack)-1]
	return frame, true
}
