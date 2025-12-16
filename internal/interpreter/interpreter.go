package interpreter

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xbasic/xbasic/internal/ast"
	"github.com/xbasic/xbasic/internal/builtins"
)

// Interpreter executes BASIC programs
type Interpreter struct {
	program  *ast.Program
	env      *Environment
	state    *ExecutionState
	output   func(string)
	input    func(string) string
	screen   Screen
	builtins *builtins.Registry
	files    map[int]*FileHandle
	graphics *GraphicsBuffer
}

// Screen interface for display operations
type Screen interface {
	Print(s string)
	Println(s string)
	Clear()
	Locate(row, col int)
	SetColor(fg, bg int)
	GetKey() string
	GetSize() (rows, cols int)
	SetCell(x, y int, ch rune)
	Show()
}

// FileHandle represents an open file
type FileHandle struct {
	Name     string
	Mode     string
	File     *os.File
	Reader   *bufio.Reader
	RecLen   int  // Record length for RANDOM mode
	Position int64
}

// GraphicsBuffer represents a text-mode graphics buffer
type GraphicsBuffer struct {
	Width  int
	Height int
	Pixels [][]int // Color values, 0 = off, 1-15 = colors
}

// New creates a new interpreter
func New(program *ast.Program) *Interpreter {
	return &Interpreter{
		program:  program,
		env:      NewEnvironment(),
		state:    NewExecutionState(),
		builtins: builtins.NewRegistry(),
		files:    make(map[int]*FileHandle),
	}
}

// SetOutput sets the output callback
func (i *Interpreter) SetOutput(fn func(string)) {
	i.output = fn
}

// SetInput sets the input callback
func (i *Interpreter) SetInput(fn func(string) string) {
	i.input = fn
}

// SetScreen sets the screen interface
func (i *Interpreter) SetScreen(s Screen) {
	i.screen = s
}

// Run executes the program
func (i *Interpreter) Run() error {
	i.state.Running = true
	i.state.ProgramCounter = 0

	for i.state.Running && i.state.ProgramCounter < len(i.program.Statements) {
		stmt := i.program.Statements[i.state.ProgramCounter]
		err := i.executeStatement(stmt)
		if err != nil {
			return err
		}
		i.state.ProgramCounter++
	}

	return nil
}

// Stop stops the running program
func (i *Interpreter) Stop() {
	i.state.Running = false
}

// Reset resets the interpreter state
func (i *Interpreter) Reset() {
	i.env = NewEnvironment()
	i.state = NewExecutionState()
	i.files = make(map[int]*FileHandle)
}

func (i *Interpreter) executeStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.LineNumberStmt:
		// Line numbers are markers, nothing to execute
		return nil

	case *ast.LabelStmt:
		// Labels are markers, nothing to execute
		return nil

	case *ast.LetStmt:
		return i.executeLetStatement(s)

	case *ast.PrintStmt:
		return i.executePrintStatement(s)

	case *ast.InputStmt:
		return i.executeInputStatement(s)

	case *ast.DimStmt:
		return i.executeDimStatement(s)

	case *ast.IfStmt:
		return i.executeIfStatement(s)

	case *ast.ForStmt:
		return i.executeForStatement(s)

	case *ast.WhileStmt:
		return i.executeWhileStatement(s)

	case *ast.DoLoopStmt:
		return i.executeDoLoopStatement(s)

	case *ast.SelectCaseStmt:
		return i.executeSelectCaseStatement(s)

	case *ast.GotoStmt:
		return i.executeGotoStatement(s)

	case *ast.GosubStmt:
		return i.executeGosubStatement(s)

	case *ast.ReturnStmt:
		return i.executeReturnStatement(s)

	case *ast.ExitStmt:
		return i.executeExitStatement(s)

	case *ast.SubStatement:
		// Skip SUB definitions during normal execution
		return i.skipSubDefinition()

	case *ast.FuncStatement:
		// Skip FUNCTION definitions during normal execution
		return i.skipFunctionDefinition()

	case *ast.CallStmt:
		return i.executeCallStatement(s)

	case *ast.SubCallStmt:
		return i.executeSubCallStatement(s)

	case *ast.DataStmt:
		// DATA statements are collected at parse time
		return nil

	case *ast.ReadStmt:
		return i.executeReadStatement(s)

	case *ast.RestoreStmt:
		return i.executeRestoreStatement(s)

	case *ast.ClsStmt:
		return i.executeClsStatement()

	case *ast.LocateStmt:
		return i.executeLocateStatement(s)

	case *ast.ColorStmt:
		return i.executeColorStatement(s)

	case *ast.ScreenStmt:
		return i.executeScreenStatement(s)

	case *ast.EndStmt:
		i.state.Running = false
		return nil

	case *ast.RemStmt:
		// Comments are ignored
		return nil

	case *ast.SleepStmt:
		return i.executeSleepStatement(s)

	case *ast.BeepStmt:
		return i.executeBeepStatement()

	case *ast.SwapStmt:
		return i.executeSwapStatement(s)

	case *ast.RandomizeStmt:
		return i.executeRandomizeStatement(s)

	case *ast.ConstStmt:
		return i.executeConstStatement(s)

	case *ast.OpenStmt:
		return i.executeOpenStatement(s)

	case *ast.CloseStmt:
		return i.executeCloseStatement(s)

	case *ast.PrintFileStmt:
		return i.executePrintFileStatement(s)

	case *ast.InputFileStmt:
		return i.executeInputFileStatement(s)

	case *ast.LineInputStmt:
		return i.executeLineInputStatement(s)

	case *ast.OnGotoStmt:
		return i.executeOnGotoStatement(s)

	case *ast.OnGosubStmt:
		return i.executeOnGosubStatement(s)

	case *ast.LineInputFileStmt:
		return i.executeLineInputFileStatement(s)

	case *ast.GetStmt:
		return i.executeGetStatement(s)

	case *ast.PutStmt:
		return i.executePutStatement(s)

	case *ast.SeekStmt:
		return i.executeSeekStatement(s)

	case *ast.RedimStmt:
		return i.executeRedimStatement(s)

	case *ast.PrintUsingStmt:
		return i.executePrintUsingStatement(s)

	case *ast.PsetStmt:
		return i.executePsetStatement(s)

	case *ast.LineGraphicsStmt:
		return i.executeLineGraphicsStatement(s)

	case *ast.CircleStmt:
		return i.executeCircleStatement(s)

	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
}

func (i *Interpreter) executeLetStatement(s *ast.LetStmt) error {
	value, err := i.evaluate(s.Value)
	if err != nil {
		return err
	}

	switch target := s.Name.(type) {
	case *ast.Identifier:
		// Check if identifier is actually an array
		if arr, ok := i.env.GetArray(target.Name); ok {
			// Single element array access without subscripts - error
			_ = arr
			return fmt.Errorf("array %s requires subscripts", target.Name)
		}
		i.env.Set(target.Name, value)

	case *ast.ArrayAccess:
		arr, ok := i.env.GetArray(target.Name)
		if !ok {
			return fmt.Errorf("array %s not defined", target.Name)
		}
		subscripts, err := i.evaluateSubscripts(target.Indices)
		if err != nil {
			return err
		}
		return arr.Set(subscripts, value)

	case *ast.CallExpr:
		// Could be array access disguised as function call
		arr, ok := i.env.GetArray(target.Function)
		if ok {
			subscripts, err := i.evaluateSubscripts(target.Arguments)
			if err != nil {
				return err
			}
			return arr.Set(subscripts, value)
		}
		// Otherwise it's an error
		return fmt.Errorf("cannot assign to function call")

	default:
		return fmt.Errorf("invalid assignment target: %T", s.Name)
	}

	return nil
}

func (i *Interpreter) executePrintStatement(s *ast.PrintStmt) error {
	var output strings.Builder
	col := 0

	for _, item := range s.Items {
		if item.Expression != nil {
			val, err := i.evaluate(item.Expression)
			if err != nil {
				return err
			}

			str := i.formatValue(val)
			output.WriteString(str)
			col += len(str)
		}

		switch item.Separator {
		case ";":
			// No space
		case ",":
			// Tab to next 14-character zone
			spaces := 14 - (col % 14)
			output.WriteString(strings.Repeat(" ", spaces))
			col += spaces
		}
	}

	if !s.NoNewline {
		output.WriteString("\n")
	}

	i.print(output.String())
	return nil
}

func (i *Interpreter) executeInputStatement(s *ast.InputStmt) error {
	prompt := "? "
	if s.Prompt != nil {
		prompt = s.Prompt.Value
	}

	// Get input from user
	inputStr := i.getInput(prompt)

	// Parse input values
	values := strings.Split(inputStr, ",")

	for idx, v := range s.Variables {
		if idx >= len(values) {
			break
		}

		inputVal := strings.TrimSpace(values[idx])
		var val Value

		switch target := v.(type) {
		case *ast.Identifier:
			dt := i.env.inferType(target.Name)
			if dt == ast.TypeString {
				val = &StringValue{Val: inputVal}
			} else {
				// Try to parse as number
				var f float64
				fmt.Sscanf(inputVal, "%f", &f)
				val = CoerceValue(&DoubleValue{Val: f}, dt)
			}
			i.env.Set(target.Name, val)

		case *ast.CallExpr:
			// Array element
			arr, ok := i.env.GetArray(target.Function)
			if !ok {
				return fmt.Errorf("array %s not defined", target.Function)
			}
			subscripts, err := i.evaluateSubscripts(target.Arguments)
			if err != nil {
				return err
			}
			if arr.DataType == ast.TypeString {
				val = &StringValue{Val: inputVal}
			} else {
				var f float64
				fmt.Sscanf(inputVal, "%f", &f)
				val = CoerceValue(&DoubleValue{Val: f}, arr.DataType)
			}
			arr.Set(subscripts, val)
		}
	}

	return nil
}

func (i *Interpreter) executeDimStatement(s *ast.DimStmt) error {
	for _, v := range s.Variables {
		if len(v.Dimensions) > 0 {
			// Array declaration
			dims := make([]int, len(v.Dimensions))
			for idx, dimExpr := range v.Dimensions {
				dimVal, err := i.evaluate(dimExpr)
				if err != nil {
					return err
				}
				dims[idx] = int(dimVal.ToInt())
			}
			dt := v.DataType
			if dt == ast.TypeUnknown {
				dt = i.env.inferType(v.Name)
			}
			i.env.DeclareArray(v.Name, dt, dims)
		} else {
			// Scalar variable declaration
			dt := v.DataType
			if dt == ast.TypeUnknown {
				dt = i.env.inferType(v.Name)
			}
			i.env.Set(v.Name, DefaultValue(dt))
		}
	}
	return nil
}

func (i *Interpreter) executeIfStatement(s *ast.IfStmt) error {
	cond, err := i.evaluate(s.Condition)
	if err != nil {
		return err
	}

	if cond.ToBool() {
		for _, stmt := range s.Consequence {
			if err := i.executeStatement(stmt); err != nil {
				return err
			}
		}
	} else if len(s.Alternative) > 0 {
		for _, stmt := range s.Alternative {
			if err := i.executeStatement(stmt); err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *Interpreter) executeForStatement(s *ast.ForStmt) error {
	// Initialize loop variable
	startVal, err := i.evaluate(s.Start)
	if err != nil {
		return err
	}
	i.env.Set(s.Variable.Name, startVal)

	endVal, err := i.evaluate(s.End)
	if err != nil {
		return err
	}

	stepVal := &DoubleValue{Val: 1}
	if s.Step != nil {
		sv, err := i.evaluate(s.Step)
		if err != nil {
			return err
		}
		stepVal = &DoubleValue{Val: sv.ToFloat()}
	}

	// Determine step direction
	stepSign := 1
	if stepVal.ToFloat() < 0 {
		stepSign = -1
	}

	// Execute loop
	for {
		// Check termination condition
		currVal := i.env.GetOrCreate(s.Variable.Name, ast.TypeDouble)
		curr := currVal.ToFloat()
		end := endVal.ToFloat()

		if stepSign > 0 && curr > end {
			break
		}
		if stepSign < 0 && curr < end {
			break
		}

		// Execute body
		for _, stmt := range s.Body {
			if err := i.executeStatement(stmt); err != nil {
				// Check for EXIT FOR
				if exitErr, ok := err.(*ExitError); ok && exitErr.ExitType == "FOR" {
					return nil
				}
				return err
			}
		}

		// Increment loop variable
		newVal := curr + stepVal.ToFloat()
		i.env.Set(s.Variable.Name, &DoubleValue{Val: newVal})
	}

	return nil
}

func (i *Interpreter) executeWhileStatement(s *ast.WhileStmt) error {
	for {
		cond, err := i.evaluate(s.Condition)
		if err != nil {
			return err
		}

		if !cond.ToBool() {
			break
		}

		for _, stmt := range s.Body {
			if err := i.executeStatement(stmt); err != nil {
				if exitErr, ok := err.(*ExitError); ok && exitErr.ExitType == "WHILE" {
					return nil
				}
				return err
			}
		}
	}

	return nil
}

func (i *Interpreter) executeDoLoopStatement(s *ast.DoLoopStmt) error {
	for {
		// Pre-condition
		if s.ConditionPos == "PRE" && s.Condition != nil {
			cond, err := i.evaluate(s.Condition)
			if err != nil {
				return err
			}
			if s.ConditionType == "WHILE" && !cond.ToBool() {
				break
			}
			if s.ConditionType == "UNTIL" && cond.ToBool() {
				break
			}
		}

		// Execute body
		for _, stmt := range s.Body {
			if err := i.executeStatement(stmt); err != nil {
				if exitErr, ok := err.(*ExitError); ok && exitErr.ExitType == "DO" {
					return nil
				}
				return err
			}
		}

		// Post-condition
		if s.ConditionPos == "POST" && s.Condition != nil {
			cond, err := i.evaluate(s.Condition)
			if err != nil {
				return err
			}
			if s.ConditionType == "WHILE" && !cond.ToBool() {
				break
			}
			if s.ConditionType == "UNTIL" && cond.ToBool() {
				break
			}
		}

		// Infinite loop if no condition
		if s.Condition == nil {
			// This would be DO...LOOP without condition
			// Need to break on EXIT DO only
		}
	}

	return nil
}

func (i *Interpreter) executeSelectCaseStatement(s *ast.SelectCaseStmt) error {
	testVal, err := i.evaluate(s.Expression)
	if err != nil {
		return err
	}

	for _, caseClause := range s.Cases {
		matched := false

		for _, cv := range caseClause.Values {
			switch cv.Type {
			case "SINGLE":
				caseVal, err := i.evaluate(cv.Value)
				if err != nil {
					return err
				}
				if Compare(testVal, caseVal) == 0 {
					matched = true
				}

			case "RANGE":
				startVal, err := i.evaluate(cv.Value)
				if err != nil {
					return err
				}
				endVal, err := i.evaluate(cv.EndValue)
				if err != nil {
					return err
				}
				if Compare(testVal, startVal) >= 0 && Compare(testVal, endVal) <= 0 {
					matched = true
				}

			case "IS":
				caseVal, err := i.evaluate(cv.Value)
				if err != nil {
					return err
				}
				cmp := Compare(testVal, caseVal)
				switch cv.Operator {
				case "<":
					matched = cmp < 0
				case ">":
					matched = cmp > 0
				case "<=":
					matched = cmp <= 0
				case ">=":
					matched = cmp >= 0
				case "=":
					matched = cmp == 0
				case "<>":
					matched = cmp != 0
				}
			}

			if matched {
				break
			}
		}

		if matched {
			for _, stmt := range caseClause.Body {
				if err := i.executeStatement(stmt); err != nil {
					return err
				}
			}
			return nil
		}
	}

	// Execute CASE ELSE if no match
	for _, stmt := range s.CaseElse {
		if err := i.executeStatement(stmt); err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) executeGotoStatement(s *ast.GotoStmt) error {
	target := strings.ToUpper(s.Target)

	// Try as line number first
	if lineNum, ok := parseLineNumber(target); ok {
		if idx, ok := i.program.LineNumbers[lineNum]; ok {
			i.state.ProgramCounter = idx - 1 // -1 because it will be incremented
			return nil
		}
	}

	// Try as label
	if idx, ok := i.program.Labels[target]; ok {
		i.state.ProgramCounter = idx - 1
		return nil
	}

	return fmt.Errorf("undefined label or line number: %s", s.Target)
}

func (i *Interpreter) executeGosubStatement(s *ast.GosubStmt) error {
	target := strings.ToUpper(s.Target)

	// Push return address
	frame := CallFrame{
		ReturnIndex: i.state.ProgramCounter,
		Type:        "GOSUB",
	}
	i.state.PushCall(frame)

	// Jump to target
	if lineNum, ok := parseLineNumber(target); ok {
		if idx, ok := i.program.LineNumbers[lineNum]; ok {
			i.state.ProgramCounter = idx - 1
			return nil
		}
	}

	if idx, ok := i.program.Labels[target]; ok {
		i.state.ProgramCounter = idx - 1
		return nil
	}

	return fmt.Errorf("undefined label or line number: %s", s.Target)
}

func (i *Interpreter) executeReturnStatement(s *ast.ReturnStmt) error {
	frame, ok := i.state.PopCall()
	if !ok {
		return fmt.Errorf("RETURN without GOSUB")
	}

	if frame.Type == "GOSUB" {
		i.state.ProgramCounter = frame.ReturnIndex
		return nil
	}

	// RETURN from SUB/FUNCTION
	if s.Value != nil && frame.Type == "FUNCTION" {
		val, err := i.evaluate(s.Value)
		if err != nil {
			return err
		}
		// Restore parent environment and set function result
		if frame.LocalEnv != nil && frame.LocalEnv.parent != nil {
			i.env = frame.LocalEnv.parent
		}
		i.env.Set(frame.FuncName, val)
	} else if frame.LocalEnv != nil && frame.LocalEnv.parent != nil {
		i.env = frame.LocalEnv.parent
	}

	i.state.ProgramCounter = frame.ReturnIndex
	return nil
}

// ExitError signals an EXIT statement
type ExitError struct {
	ExitType string
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("EXIT %s", e.ExitType)
}

func (i *Interpreter) executeExitStatement(s *ast.ExitStmt) error {
	return &ExitError{ExitType: s.ExitType}
}

func (i *Interpreter) skipSubDefinition() error {
	// Skip until END SUB
	for i.state.ProgramCounter < len(i.program.Statements)-1 {
		i.state.ProgramCounter++
		stmt := i.program.Statements[i.state.ProgramCounter]
		if _, ok := stmt.(*ast.EndStmt); ok {
			break
		}
		// Check for END SUB marker
		if end, ok := stmt.(*ast.SubStatement); ok {
			_ = end
			break
		}
	}
	return nil
}

func (i *Interpreter) skipFunctionDefinition() error {
	// Skip until END FUNCTION
	for i.state.ProgramCounter < len(i.program.Statements)-1 {
		i.state.ProgramCounter++
		stmt := i.program.Statements[i.state.ProgramCounter]
		if _, ok := stmt.(*ast.EndStmt); ok {
			break
		}
		if end, ok := stmt.(*ast.FuncStatement); ok {
			_ = end
			break
		}
	}
	return nil
}

func (i *Interpreter) executeCallStatement(s *ast.CallStmt) error {
	return i.callSub(s.Name, s.Arguments)
}

func (i *Interpreter) executeSubCallStatement(s *ast.SubCallStmt) error {
	return i.callSub(s.Name, s.Arguments)
}

func (i *Interpreter) callSub(name string, args []ast.Expression) error {
	name = strings.ToUpper(name)

	sub, ok := i.program.Subs[name]
	if !ok {
		return fmt.Errorf("undefined SUB: %s", name)
	}

	// Evaluate arguments
	argVals := make([]Value, len(args))
	for idx, arg := range args {
		val, err := i.evaluate(arg)
		if err != nil {
			return err
		}
		argVals[idx] = val
	}

	// Create local environment
	localEnv := NewEnclosedEnvironment(i.env)

	// Bind parameters
	for idx, param := range sub.Parameters {
		if idx < len(argVals) {
			localEnv.Set(param.Name, argVals[idx])
		} else {
			localEnv.Set(param.Name, DefaultValue(param.DataType))
		}
	}

	// Save current state
	frame := CallFrame{
		ReturnIndex: i.state.ProgramCounter,
		LocalEnv:    i.env,
		Type:        "SUB",
	}
	i.state.PushCall(frame)

	// Switch to local environment
	i.env = localEnv

	// Execute SUB body
	for _, stmt := range sub.Body {
		if err := i.executeStatement(stmt); err != nil {
			if exitErr, ok := err.(*ExitError); ok && exitErr.ExitType == "SUB" {
				break
			}
			return err
		}
	}

	// Restore environment
	popFrame, _ := i.state.PopCall()
	if popFrame.LocalEnv != nil {
		i.env = popFrame.LocalEnv
	}

	return nil
}

func (i *Interpreter) executeReadStatement(s *ast.ReadStmt) error {
	for _, v := range s.Variables {
		if i.state.DataPointer >= len(i.program.DataItems) {
			return fmt.Errorf("out of DATA")
		}

		dataVal, err := i.evaluate(i.program.DataItems[i.state.DataPointer])
		if err != nil {
			return err
		}
		i.state.DataPointer++

		switch target := v.(type) {
		case *ast.Identifier:
			i.env.Set(target.Name, dataVal)
		case *ast.CallExpr:
			arr, ok := i.env.GetArray(target.Function)
			if !ok {
				return fmt.Errorf("array %s not defined", target.Function)
			}
			subscripts, err := i.evaluateSubscripts(target.Arguments)
			if err != nil {
				return err
			}
			arr.Set(subscripts, dataVal)
		}
	}

	return nil
}

func (i *Interpreter) executeRestoreStatement(s *ast.RestoreStmt) error {
	if s.Target == "" {
		i.state.DataPointer = 0
		return nil
	}

	// Find DATA at specific label/line
	target := strings.ToUpper(s.Target)

	if lineNum, ok := parseLineNumber(target); ok {
		if idx, ok := i.program.LineNumbers[lineNum]; ok {
			// Count DATA items before this line
			count := 0
			for j := 0; j < idx; j++ {
				if data, ok := i.program.Statements[j].(*ast.DataStmt); ok {
					count += len(data.Values)
				}
			}
			i.state.DataPointer = count
			return nil
		}
	}

	return fmt.Errorf("undefined label or line number: %s", s.Target)
}

func (i *Interpreter) executeClsStatement() error {
	if i.screen != nil {
		i.screen.Clear()
	}
	return nil
}

func (i *Interpreter) executeLocateStatement(s *ast.LocateStmt) error {
	row, col := 1, 1

	if s.Row != nil {
		val, err := i.evaluate(s.Row)
		if err != nil {
			return err
		}
		row = int(val.ToInt())
	}

	if s.Column != nil {
		val, err := i.evaluate(s.Column)
		if err != nil {
			return err
		}
		col = int(val.ToInt())
	}

	if i.screen != nil {
		i.screen.Locate(row, col)
	}
	return nil
}

func (i *Interpreter) executeColorStatement(s *ast.ColorStmt) error {
	fg, bg := 7, 0 // default white on black

	if s.Foreground != nil {
		val, err := i.evaluate(s.Foreground)
		if err != nil {
			return err
		}
		fg = int(val.ToInt())
	}

	if s.Background != nil {
		val, err := i.evaluate(s.Background)
		if err != nil {
			return err
		}
		bg = int(val.ToInt())
	}

	if i.screen != nil {
		i.screen.SetColor(fg, bg)
	}
	return nil
}

func (i *Interpreter) executeScreenStatement(s *ast.ScreenStmt) error {
	// SCREEN mode - for terminal, we mainly support text modes
	return nil
}

func (i *Interpreter) executeSleepStatement(s *ast.SleepStmt) error {
	// Sleep is handled by the UI layer
	return nil
}

func (i *Interpreter) executeBeepStatement() error {
	// BEEP - platform-specific
	i.print("\a") // ASCII bell
	return nil
}

func (i *Interpreter) executeSwapStatement(s *ast.SwapStmt) error {
	val1, err := i.evaluate(s.Var1)
	if err != nil {
		return err
	}
	val2, err := i.evaluate(s.Var2)
	if err != nil {
		return err
	}

	// Assign swapped values
	if id1, ok := s.Var1.(*ast.Identifier); ok {
		i.env.Set(id1.Name, val2)
	}
	if id2, ok := s.Var2.(*ast.Identifier); ok {
		i.env.Set(id2.Name, val1)
	}

	return nil
}

func (i *Interpreter) executeRandomizeStatement(s *ast.RandomizeStmt) error {
	if s.Seed != nil {
		val, err := i.evaluate(s.Seed)
		if err != nil {
			return err
		}
		i.builtins.SetRandomSeed(val.ToInt())
	} else {
		i.builtins.RandomizeSeed()
	}
	return nil
}

func (i *Interpreter) executeConstStatement(s *ast.ConstStmt) error {
	val, err := i.evaluate(s.Value)
	if err != nil {
		return err
	}
	return i.env.DefineConst(s.Name, val)
}

func (i *Interpreter) executeOpenStatement(s *ast.OpenStmt) error {
	filename, err := i.evaluate(s.Filename)
	if err != nil {
		return err
	}

	fileNumVal, err := i.evaluate(s.FileNum)
	if err != nil {
		return err
	}
	fileNum := int(fileNumVal.ToInt())

	if _, exists := i.files[fileNum]; exists {
		return fmt.Errorf("file #%d already open", fileNum)
	}

	var file *os.File
	mode := strings.ToUpper(s.Mode)

	switch mode {
	case "INPUT":
		file, err = os.Open(filename.ToString())
	case "OUTPUT":
		file, err = os.Create(filename.ToString())
	case "APPEND":
		file, err = os.OpenFile(filename.ToString(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	case "BINARY":
		file, err = os.OpenFile(filename.ToString(), os.O_RDWR|os.O_CREATE, 0644)
	case "RANDOM":
		file, err = os.OpenFile(filename.ToString(), os.O_RDWR|os.O_CREATE, 0644)
	default:
		return fmt.Errorf("invalid file mode: %s", mode)
	}

	if err != nil {
		return fmt.Errorf("cannot open file: %v", err)
	}

	fh := &FileHandle{
		Name:   filename.ToString(),
		Mode:   mode,
		File:   file,
		RecLen: 128, // default record length
	}

	if mode == "INPUT" {
		fh.Reader = bufio.NewReader(file)
	}

	// Handle record length for RANDOM mode
	if s.RecLen != nil {
		recLen, err := i.evaluate(s.RecLen)
		if err != nil {
			return err
		}
		fh.RecLen = int(recLen.ToInt())
	}

	i.files[fileNum] = fh
	return nil
}

func (i *Interpreter) executeCloseStatement(s *ast.CloseStmt) error {
	if len(s.FileNums) == 0 {
		// Close all files
		for num, fh := range i.files {
			if fh.File != nil {
				fh.File.Close()
			}
			delete(i.files, num)
		}
		return nil
	}

	for _, fileNumExpr := range s.FileNums {
		fileNumVal, err := i.evaluate(fileNumExpr)
		if err != nil {
			return err
		}
		fileNum := int(fileNumVal.ToInt())

		fh, exists := i.files[fileNum]
		if !exists {
			return fmt.Errorf("file #%d not open", fileNum)
		}
		if fh.File != nil {
			fh.File.Close()
		}
		delete(i.files, fileNum)
	}
	return nil
}

func (i *Interpreter) executePrintFileStatement(s *ast.PrintFileStmt) error {
	fileNumVal, err := i.evaluate(s.FileNum)
	if err != nil {
		return err
	}
	fileNum := int(fileNumVal.ToInt())

	fh, exists := i.files[fileNum]
	if !exists {
		return fmt.Errorf("file #%d not open", fileNum)
	}
	if fh.Mode != "OUTPUT" && fh.Mode != "APPEND" && fh.Mode != "BINARY" && fh.Mode != "RANDOM" {
		return fmt.Errorf("file #%d not open for output", fileNum)
	}

	var output strings.Builder
	col := 0

	for _, item := range s.Items {
		if item.Expression != nil {
			val, err := i.evaluate(item.Expression)
			if err != nil {
				return err
			}
			str := i.formatValue(val)
			output.WriteString(str)
			col += len(str)
		}

		switch item.Separator {
		case ";":
			// No space
		case ",":
			// Tab to next 14-character zone
			spaces := 14 - (col % 14)
			output.WriteString(strings.Repeat(" ", spaces))
			col += spaces
		}
	}

	if !s.NoNewline {
		output.WriteString("\n")
	}

	_, err = fh.File.WriteString(output.String())
	return err
}

func (i *Interpreter) executeInputFileStatement(s *ast.InputFileStmt) error {
	fileNumVal, err := i.evaluate(s.FileNum)
	if err != nil {
		return err
	}
	fileNum := int(fileNumVal.ToInt())

	fh, exists := i.files[fileNum]
	if !exists {
		return fmt.Errorf("file #%d not open", fileNum)
	}
	if fh.Mode != "INPUT" && fh.Mode != "BINARY" && fh.Mode != "RANDOM" {
		return fmt.Errorf("file #%d not open for input", fileNum)
	}

	// Create reader if needed
	if fh.Reader == nil {
		fh.Reader = bufio.NewReader(fh.File)
	}

	for _, v := range s.Variables {
		// Read value from file (comma or newline delimited)
		var inputVal string
		var ch byte

		// Skip leading whitespace
		for {
			ch, err = fh.Reader.ReadByte()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			if ch != ' ' && ch != '\t' {
				fh.Reader.UnreadByte()
				break
			}
		}

		// Read until delimiter
		for {
			ch, err = fh.Reader.ReadByte()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			if ch == ',' || ch == '\n' || ch == '\r' {
				if ch == '\r' {
					// Skip LF after CR
					nextCh, err := fh.Reader.ReadByte()
					if err == nil && nextCh != '\n' {
						fh.Reader.UnreadByte()
					}
				}
				break
			}
			inputVal += string(ch)
		}

		inputVal = strings.TrimSpace(inputVal)

		switch target := v.(type) {
		case *ast.Identifier:
			dt := i.env.inferType(target.Name)
			var val Value
			if dt == ast.TypeString {
				val = &StringValue{Val: inputVal}
			} else {
				var f float64
				fmt.Sscanf(inputVal, "%f", &f)
				val = CoerceValue(&DoubleValue{Val: f}, dt)
			}
			i.env.Set(target.Name, val)

		case *ast.CallExpr:
			arr, ok := i.env.GetArray(target.Function)
			if !ok {
				return fmt.Errorf("array %s not defined", target.Function)
			}
			subscripts, err := i.evaluateSubscripts(target.Arguments)
			if err != nil {
				return err
			}
			var val Value
			if arr.DataType == ast.TypeString {
				val = &StringValue{Val: inputVal}
			} else {
				var f float64
				fmt.Sscanf(inputVal, "%f", &f)
				val = CoerceValue(&DoubleValue{Val: f}, arr.DataType)
			}
			arr.Set(subscripts, val)
		}
	}

	return nil
}

func (i *Interpreter) executeLineInputStatement(s *ast.LineInputStmt) error {
	prompt := ""
	if s.Prompt != nil {
		prompt = s.Prompt.Value
	}

	inputStr := i.getInput(prompt)

	if id, ok := s.Variable.(*ast.Identifier); ok {
		i.env.Set(id.Name, &StringValue{Val: inputStr})
	}

	return nil
}

func (i *Interpreter) executeOnGotoStatement(s *ast.OnGotoStmt) error {
	val, err := i.evaluate(s.Expression)
	if err != nil {
		return err
	}

	idx := int(val.ToInt())
	if idx >= 1 && idx <= len(s.Targets) {
		target := s.Targets[idx-1]
		gs := &ast.GotoStmt{Target: target}
		return i.executeGotoStatement(gs)
	}

	return nil
}

func (i *Interpreter) executeOnGosubStatement(s *ast.OnGosubStmt) error {
	val, err := i.evaluate(s.Expression)
	if err != nil {
		return err
	}

	idx := int(val.ToInt())
	if idx >= 1 && idx <= len(s.Targets) {
		target := s.Targets[idx-1]
		gs := &ast.GosubStmt{Target: target}
		return i.executeGosubStatement(gs)
	}

	return nil
}

// Expression evaluation

func (i *Interpreter) evaluate(expr ast.Expression) (Value, error) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return &LongValue{Val: int32(e.Value)}, nil

	case *ast.FloatLiteral:
		return &DoubleValue{Val: e.Value}, nil

	case *ast.StringLiteral:
		return &StringValue{Val: e.Value}, nil

	case *ast.Identifier:
		name := strings.ToUpper(e.Name)
		// Check for built-in functions that can be called without parentheses
		switch name {
		case "RND", "TIMER", "DATE$", "TIME$", "INKEY$", "_PI", "PI":
			args := []builtins.Value{}
			result, err := i.builtins.Call(name, args)
			if err != nil {
				return nil, err
			}
			return builtinToValue(result), nil
		case "FREEFILE":
			return &IntegerValue{Val: int16(i.GetNextFreeFile())}, nil
		}
		val, ok := i.env.Get(e.Name)
		if !ok {
			// Auto-create variable with default value
			val = i.env.GetOrCreate(e.Name, e.TypeHint)
		}
		return val, nil

	case *ast.ArrayAccess:
		arr, ok := i.env.GetArray(e.Name)
		if !ok {
			return nil, fmt.Errorf("array %s not defined", e.Name)
		}
		subscripts, err := i.evaluateSubscripts(e.Indices)
		if err != nil {
			return nil, err
		}
		return arr.Get(subscripts)

	case *ast.BinaryExpr:
		return i.evaluateBinaryExpr(e)

	case *ast.UnaryExpr:
		return i.evaluateUnaryExpr(e)

	case *ast.CallExpr:
		return i.evaluateCallExpr(e)

	case *ast.GroupedExpr:
		return i.evaluate(e.Expression)

	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

func (i *Interpreter) evaluateBinaryExpr(e *ast.BinaryExpr) (Value, error) {
	left, err := i.evaluate(e.Left)
	if err != nil {
		return nil, err
	}

	right, err := i.evaluate(e.Right)
	if err != nil {
		return nil, err
	}

	// String concatenation
	if e.Operator == "+" && (left.Type() == ast.TypeString || right.Type() == ast.TypeString) {
		return &StringValue{Val: left.ToString() + right.ToString()}, nil
	}

	// String comparison
	if left.Type() == ast.TypeString && right.Type() == ast.TypeString {
		ls := left.(*StringValue).Val
		rs := right.(*StringValue).Val

		switch e.Operator {
		case "=":
			return boolToValue(ls == rs), nil
		case "<>":
			return boolToValue(ls != rs), nil
		case "<":
			return boolToValue(ls < rs), nil
		case ">":
			return boolToValue(ls > rs), nil
		case "<=":
			return boolToValue(ls <= rs), nil
		case ">=":
			return boolToValue(ls >= rs), nil
		default:
			return nil, fmt.Errorf("invalid operator %s for strings", e.Operator)
		}
	}

	// Numeric operations
	lf := left.ToFloat()
	rf := right.ToFloat()

	switch e.Operator {
	case "+":
		return &DoubleValue{Val: lf + rf}, nil
	case "-":
		return &DoubleValue{Val: lf - rf}, nil
	case "*":
		return &DoubleValue{Val: lf * rf}, nil
	case "/":
		if rf == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return &DoubleValue{Val: lf / rf}, nil
	case "\\":
		if rf == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return &LongValue{Val: int32(int64(lf) / int64(rf))}, nil
	case "^":
		return &DoubleValue{Val: pow(lf, rf)}, nil
	case "MOD":
		if rf == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return &LongValue{Val: int32(int64(lf) % int64(rf))}, nil
	case "=":
		return boolToValue(lf == rf), nil
	case "<>":
		return boolToValue(lf != rf), nil
	case "<":
		return boolToValue(lf < rf), nil
	case ">":
		return boolToValue(lf > rf), nil
	case "<=":
		return boolToValue(lf <= rf), nil
	case ">=":
		return boolToValue(lf >= rf), nil
	case "AND":
		return &LongValue{Val: int32(int64(lf) & int64(rf))}, nil
	case "OR":
		return &LongValue{Val: int32(int64(lf) | int64(rf))}, nil
	case "XOR":
		return &LongValue{Val: int32(int64(lf) ^ int64(rf))}, nil
	case "EQV":
		return &LongValue{Val: int32(^(int64(lf) ^ int64(rf)))}, nil
	case "IMP":
		return &LongValue{Val: int32(^int64(lf) | int64(rf))}, nil
	default:
		return nil, fmt.Errorf("unknown operator: %s", e.Operator)
	}
}

func (i *Interpreter) evaluateUnaryExpr(e *ast.UnaryExpr) (Value, error) {
	right, err := i.evaluate(e.Right)
	if err != nil {
		return nil, err
	}

	switch e.Operator {
	case "-":
		return &DoubleValue{Val: -right.ToFloat()}, nil
	case "NOT":
		return &LongValue{Val: int32(^int64(right.ToInt()))}, nil
	default:
		return nil, fmt.Errorf("unknown unary operator: %s", e.Operator)
	}
}

func (i *Interpreter) evaluateCallExpr(e *ast.CallExpr) (Value, error) {
	name := strings.ToUpper(e.Function)

	// Check if it's an array access
	if arr, ok := i.env.GetArray(name); ok {
		subscripts, err := i.evaluateSubscripts(e.Arguments)
		if err != nil {
			return nil, err
		}
		return arr.Get(subscripts)
	}

	// Check for user-defined function
	if fn, ok := i.program.Functions[name]; ok {
		return i.callFunction(fn, e.Arguments)
	}

	// Handle file I/O functions that need interpreter access
	switch name {
	case "EOF":
		if len(e.Arguments) < 1 {
			return nil, fmt.Errorf("EOF requires 1 argument")
		}
		fileNumVal, err := i.evaluate(e.Arguments[0])
		if err != nil {
			return nil, err
		}
		fileNum := int(fileNumVal.ToInt())
		fh, exists := i.files[fileNum]
		if !exists {
			return nil, fmt.Errorf("file #%d not open", fileNum)
		}
		// Check if at end of file
		if fh.Reader != nil {
			_, err := fh.Reader.Peek(1)
			if err == io.EOF {
				return &IntegerValue{Val: -1}, nil // TRUE
			}
		}
		return &IntegerValue{Val: 0}, nil // FALSE

	case "LOF":
		if len(e.Arguments) < 1 {
			return nil, fmt.Errorf("LOF requires 1 argument")
		}
		fileNumVal, err := i.evaluate(e.Arguments[0])
		if err != nil {
			return nil, err
		}
		fileNum := int(fileNumVal.ToInt())
		fh, exists := i.files[fileNum]
		if !exists {
			return nil, fmt.Errorf("file #%d not open", fileNum)
		}
		info, err := fh.File.Stat()
		if err != nil {
			return nil, err
		}
		return &LongValue{Val: int32(info.Size())}, nil

	case "LOC":
		if len(e.Arguments) < 1 {
			return nil, fmt.Errorf("LOC requires 1 argument")
		}
		fileNumVal, err := i.evaluate(e.Arguments[0])
		if err != nil {
			return nil, err
		}
		fileNum := int(fileNumVal.ToInt())
		fh, exists := i.files[fileNum]
		if !exists {
			return nil, fmt.Errorf("file #%d not open", fileNum)
		}
		pos, err := fh.File.Seek(0, 1) // Get current position
		if err != nil {
			return nil, err
		}
		if fh.Mode == "RANDOM" {
			return &LongValue{Val: int32(pos / int64(fh.RecLen))}, nil
		}
		return &LongValue{Val: int32(pos)}, nil

	case "FREEFILE":
		return &IntegerValue{Val: int16(i.GetNextFreeFile())}, nil
	}

	// Evaluate arguments
	args := make([]builtins.Value, len(e.Arguments))
	for idx, arg := range e.Arguments {
		val, err := i.evaluate(arg)
		if err != nil {
			return nil, err
		}
		args[idx] = valueToBuiltin(val)
	}

	// Call built-in function
	result, err := i.builtins.Call(name, args)
	if err != nil {
		return nil, err
	}
	return builtinToValue(result), nil
}

func (i *Interpreter) callFunction(fn *ast.FuncStatement, args []ast.Expression) (Value, error) {
	// Evaluate arguments
	argVals := make([]Value, len(args))
	for idx, arg := range args {
		val, err := i.evaluate(arg)
		if err != nil {
			return nil, err
		}
		argVals[idx] = val
	}

	// Create local environment
	localEnv := NewEnclosedEnvironment(i.env)

	// Bind parameters
	for idx, param := range fn.Parameters {
		if idx < len(argVals) {
			localEnv.Set(param.Name, argVals[idx])
		} else {
			localEnv.Set(param.Name, DefaultValue(param.DataType))
		}
	}

	// Initialize return variable (function name)
	localEnv.Set(fn.Name, DefaultValue(fn.ReturnType))

	// Save current environment
	savedEnv := i.env
	i.env = localEnv

	// Execute function body
	for _, stmt := range fn.Body {
		if err := i.executeStatement(stmt); err != nil {
			if exitErr, ok := err.(*ExitError); ok && exitErr.ExitType == "FUNCTION" {
				break
			}
			i.env = savedEnv
			return nil, err
		}
	}

	// Get return value
	retVal, _ := i.env.Get(fn.Name)
	if retVal == nil {
		retVal = DefaultValue(fn.ReturnType)
	}

	// Restore environment
	i.env = savedEnv

	return retVal, nil
}

func (i *Interpreter) evaluateSubscripts(exprs []ast.Expression) ([]int, error) {
	subscripts := make([]int, len(exprs))
	for idx, expr := range exprs {
		val, err := i.evaluate(expr)
		if err != nil {
			return nil, err
		}
		subscripts[idx] = int(val.ToInt())
	}
	return subscripts, nil
}

// Helper methods

func (i *Interpreter) print(s string) {
	if i.output != nil {
		i.output(s)
	} else if i.screen != nil {
		i.screen.Print(s)
	}
}

func (i *Interpreter) getInput(prompt string) string {
	if i.input != nil {
		return i.input(prompt)
	}
	return ""
}

func (i *Interpreter) formatValue(val Value) string {
	switch v := val.(type) {
	case *StringValue:
		return v.Val
	case *IntegerValue:
		if v.Val >= 0 {
			return " " + v.String()
		}
		return v.String()
	case *LongValue:
		if v.Val >= 0 {
			return " " + v.String()
		}
		return v.String()
	case *SingleValue:
		if v.Val >= 0 {
			return " " + v.String()
		}
		return v.String()
	case *DoubleValue:
		if v.Val >= 0 {
			return " " + v.String()
		}
		return v.String()
	}
	return val.String()
}

func boolToValue(b bool) Value {
	if b {
		return &IntegerValue{Val: -1} // TRUE in QBasic
	}
	return &IntegerValue{Val: 0} // FALSE
}

func parseLineNumber(s string) (int, bool) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err == nil
}

func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	if exp == 1 {
		return base
	}
	if exp == 2 {
		return base * base
	}
	// Use math.Pow for general case
	result := 1.0
	for exp > 0 {
		if int(exp)%2 == 1 {
			result *= base
		}
		base *= base
		exp = float64(int(exp) / 2)
	}
	return result
}

// Conversion functions between interpreter and builtins value types

func valueToBuiltin(v Value) builtins.Value {
	switch val := v.(type) {
	case *IntegerValue:
		return &builtins.IntegerValue{Val: val.Val}
	case *LongValue:
		return &builtins.LongValue{Val: val.Val}
	case *SingleValue:
		return &builtins.SingleValue{Val: val.Val}
	case *DoubleValue:
		return &builtins.DoubleValue{Val: val.Val}
	case *StringValue:
		return &builtins.StringValue{Val: val.Val}
	default:
		return &builtins.DoubleValue{Val: v.ToFloat()}
	}
}

func builtinToValue(v builtins.Value) Value {
	switch val := v.(type) {
	case *builtins.IntegerValue:
		return &IntegerValue{Val: val.Val}
	case *builtins.LongValue:
		return &LongValue{Val: val.Val}
	case *builtins.SingleValue:
		return &SingleValue{Val: val.Val}
	case *builtins.DoubleValue:
		return &DoubleValue{Val: val.Val}
	case *builtins.StringValue:
		return &StringValue{Val: val.Val}
	default:
		return &DoubleValue{Val: v.ToFloat()}
	}
}

func (i *Interpreter) executeLineInputFileStatement(s *ast.LineInputFileStmt) error {
	fileNumVal, err := i.evaluate(s.FileNum)
	if err != nil {
		return err
	}
	fileNum := int(fileNumVal.ToInt())

	fh, exists := i.files[fileNum]
	if !exists {
		return fmt.Errorf("file #%d not open", fileNum)
	}
	if fh.Mode != "INPUT" && fh.Mode != "BINARY" {
		return fmt.Errorf("file #%d not open for input", fileNum)
	}

	if fh.Reader == nil {
		fh.Reader = bufio.NewReader(fh.File)
	}

	// Read entire line
	line, err := fh.Reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}
	line = strings.TrimRight(line, "\r\n")

	if id, ok := s.Variable.(*ast.Identifier); ok {
		i.env.Set(id.Name, &StringValue{Val: line})
	}

	return nil
}

func (i *Interpreter) executeGetStatement(s *ast.GetStmt) error {
	fileNumVal, err := i.evaluate(s.FileNum)
	if err != nil {
		return err
	}
	fileNum := int(fileNumVal.ToInt())

	fh, exists := i.files[fileNum]
	if !exists {
		return fmt.Errorf("file #%d not open", fileNum)
	}
	if fh.Mode != "BINARY" && fh.Mode != "RANDOM" {
		return fmt.Errorf("file #%d not open for binary/random access", fileNum)
	}

	// Seek to position if specified
	if s.Position != nil {
		posVal, err := i.evaluate(s.Position)
		if err != nil {
			return err
		}
		pos := posVal.ToInt() - 1 // QBasic positions are 1-based
		if fh.Mode == "RANDOM" {
			pos = pos * int64(fh.RecLen)
		}
		_, err = fh.File.Seek(pos, 0)
		if err != nil {
			return err
		}
	}

	// Read data based on variable type
	if s.Variable != nil {
		switch target := s.Variable.(type) {
		case *ast.Identifier:
			dt := i.env.inferType(target.Name)
			var val Value

			switch dt {
			case ast.TypeInteger:
				var v int16
				binary.Read(fh.File, binary.LittleEndian, &v)
				val = &IntegerValue{Val: v}
			case ast.TypeLong:
				var v int32
				binary.Read(fh.File, binary.LittleEndian, &v)
				val = &LongValue{Val: v}
			case ast.TypeSingle:
				var v float32
				binary.Read(fh.File, binary.LittleEndian, &v)
				val = &SingleValue{Val: v}
			case ast.TypeDouble:
				var v float64
				binary.Read(fh.File, binary.LittleEndian, &v)
				val = &DoubleValue{Val: v}
			case ast.TypeString:
				buf := make([]byte, fh.RecLen)
				n, _ := fh.File.Read(buf)
				val = &StringValue{Val: strings.TrimRight(string(buf[:n]), "\x00")}
			default:
				var v float64
				binary.Read(fh.File, binary.LittleEndian, &v)
				val = &DoubleValue{Val: v}
			}
			i.env.Set(target.Name, val)
		}
	}

	return nil
}

func (i *Interpreter) executePutStatement(s *ast.PutStmt) error {
	fileNumVal, err := i.evaluate(s.FileNum)
	if err != nil {
		return err
	}
	fileNum := int(fileNumVal.ToInt())

	fh, exists := i.files[fileNum]
	if !exists {
		return fmt.Errorf("file #%d not open", fileNum)
	}
	if fh.Mode != "BINARY" && fh.Mode != "RANDOM" {
		return fmt.Errorf("file #%d not open for binary/random access", fileNum)
	}

	// Seek to position if specified
	if s.Position != nil {
		posVal, err := i.evaluate(s.Position)
		if err != nil {
			return err
		}
		pos := posVal.ToInt() - 1 // QBasic positions are 1-based
		if fh.Mode == "RANDOM" {
			pos = pos * int64(fh.RecLen)
		}
		_, err = fh.File.Seek(pos, 0)
		if err != nil {
			return err
		}
	}

	// Write data based on variable type
	if s.Variable != nil {
		val, err := i.evaluate(s.Variable)
		if err != nil {
			return err
		}

		switch v := val.(type) {
		case *IntegerValue:
			binary.Write(fh.File, binary.LittleEndian, v.Val)
		case *LongValue:
			binary.Write(fh.File, binary.LittleEndian, v.Val)
		case *SingleValue:
			binary.Write(fh.File, binary.LittleEndian, v.Val)
		case *DoubleValue:
			binary.Write(fh.File, binary.LittleEndian, v.Val)
		case *StringValue:
			buf := make([]byte, fh.RecLen)
			copy(buf, v.Val)
			fh.File.Write(buf)
		}
	}

	return nil
}

func (i *Interpreter) executeSeekStatement(s *ast.SeekStmt) error {
	fileNumVal, err := i.evaluate(s.FileNum)
	if err != nil {
		return err
	}
	fileNum := int(fileNumVal.ToInt())

	fh, exists := i.files[fileNum]
	if !exists {
		return fmt.Errorf("file #%d not open", fileNum)
	}

	posVal, err := i.evaluate(s.Position)
	if err != nil {
		return err
	}
	pos := posVal.ToInt() - 1 // QBasic positions are 1-based

	if fh.Mode == "RANDOM" {
		pos = pos * int64(fh.RecLen)
	}

	_, err = fh.File.Seek(pos, 0)
	return err
}

func (i *Interpreter) executeRedimStatement(s *ast.RedimStmt) error {
	for _, v := range s.Variables {
		if len(v.Dimensions) > 0 {
			dims := make([]int, len(v.Dimensions))
			for idx, dimExpr := range v.Dimensions {
				dimVal, err := i.evaluate(dimExpr)
				if err != nil {
					return err
				}
				dims[idx] = int(dimVal.ToInt())
			}

			dt := v.DataType
			if dt == ast.TypeUnknown {
				dt = i.env.inferType(v.Name)
			}

			if s.Preserve {
				// Get existing array data if it exists
				existingArr, exists := i.env.GetArray(v.Name)
				if exists {
					// Create new array with new dimensions
					newArr := i.env.DeclareArray(v.Name, dt, dims)
					// Copy existing data (up to the smaller dimension)
					i.copyArrayData(existingArr, newArr)
				} else {
					i.env.DeclareArray(v.Name, dt, dims)
				}
			} else {
				i.env.DeclareArray(v.Name, dt, dims)
			}
		}
	}
	return nil
}

func (i *Interpreter) copyArrayData(src, dst *Array) {
	// Simple copy for 1D arrays
	if len(src.Dimensions) == 1 && len(dst.Dimensions) == 1 {
		maxIdx := src.Dimensions[0].Upper
		if dst.Dimensions[0].Upper < maxIdx {
			maxIdx = dst.Dimensions[0].Upper
		}
		for idx := 0; idx <= maxIdx; idx++ {
			if val, err := src.Get([]int{idx}); err == nil {
				dst.Set([]int{idx}, val)
			}
		}
	}
}

func (i *Interpreter) executePrintUsingStatement(s *ast.PrintUsingStmt) error {
	formatVal, err := i.evaluate(s.Format)
	if err != nil {
		return err
	}
	format := formatVal.ToString()

	var output strings.Builder

	for _, item := range s.Items {
		if item.Expression != nil {
			val, err := i.evaluate(item.Expression)
			if err != nil {
				return err
			}
			formatted := i.formatWithTemplate(format, val)
			output.WriteString(formatted)
		}
	}

	if !s.NoNewline {
		output.WriteString("\n")
	}

	if s.FileNum != nil {
		fileNumVal, err := i.evaluate(s.FileNum)
		if err != nil {
			return err
		}
		fileNum := int(fileNumVal.ToInt())
		fh, exists := i.files[fileNum]
		if !exists {
			return fmt.Errorf("file #%d not open", fileNum)
		}
		_, err = fh.File.WriteString(output.String())
		return err
	}

	i.print(output.String())
	return nil
}

func (i *Interpreter) formatWithTemplate(format string, val Value) string {
	// Process format specifiers
	// ###.## - numeric
	// $$ - dollar sign
	// ** - asterisk fill
	// + - sign
	// ^^^^ - exponential
	// \ \ - fixed width string
	// ! - first char
	// & - whole string

	if strings.Contains(format, "#") {
		// Numeric format
		return i.formatNumeric(format, val.ToFloat())
	} else if strings.HasPrefix(format, "\\") && strings.HasSuffix(format, "\\") {
		// Fixed width string
		width := len(format)
		s := val.ToString()
		if len(s) > width {
			return s[:width]
		}
		return s + strings.Repeat(" ", width-len(s))
	} else if format == "!" {
		s := val.ToString()
		if len(s) > 0 {
			return string(s[0])
		}
		return " "
	} else if format == "&" {
		return val.ToString()
	}

	return val.ToString()
}

func (i *Interpreter) formatNumeric(format string, val float64) string {
	// Count format characters
	dollarPrefix := strings.Count(format, "$")
	asteriskFill := strings.Count(format, "*")
	plusSign := strings.Contains(format, "+")
	minusSign := strings.HasSuffix(format, "-")
	exponential := strings.Contains(format, "^^^^")

	// Remove special chars to count digit positions
	cleaned := format
	cleaned = strings.ReplaceAll(cleaned, "$", "")
	cleaned = strings.ReplaceAll(cleaned, "*", "")
	cleaned = strings.ReplaceAll(cleaned, "+", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "^", "")

	decPos := strings.Index(cleaned, ".")
	intDigits := strings.Count(cleaned, "#")
	decDigits := 0
	if decPos >= 0 {
		decDigits = strings.Count(cleaned[decPos:], "#")
		intDigits = intDigits - decDigits
	}

	var result string
	if exponential {
		result = fmt.Sprintf("%.*E", decDigits, val)
	} else if decPos >= 0 {
		result = fmt.Sprintf("%*.*f", intDigits+decDigits+1, decDigits, val)
	} else {
		result = fmt.Sprintf("%*.0f", intDigits, val)
	}

	// Apply formatting options
	if plusSign && val >= 0 {
		result = "+" + result
	}
	if minusSign && val < 0 {
		result = strings.TrimPrefix(result, "-")
		result = result + "-"
	}
	if dollarPrefix > 0 {
		result = "$" + strings.TrimLeft(result, " ")
	}
	if asteriskFill > 0 {
		result = strings.ReplaceAll(result, " ", "*")
	}

	return result
}

func (i *Interpreter) executePsetStatement(s *ast.PsetStmt) error {
	xVal, err := i.evaluate(s.X)
	if err != nil {
		return err
	}
	yVal, err := i.evaluate(s.Y)
	if err != nil {
		return err
	}

	x := int(xVal.ToInt())
	y := int(yVal.ToInt())

	color := 15 // default white
	if s.Color != nil {
		colorVal, err := i.evaluate(s.Color)
		if err != nil {
			return err
		}
		color = int(colorVal.ToInt())
	}

	i.drawPixel(x, y, color)
	return nil
}

func (i *Interpreter) executeLineGraphicsStatement(s *ast.LineGraphicsStmt) error {
	x1Val, err := i.evaluate(s.X1)
	if err != nil {
		return err
	}
	y1Val, err := i.evaluate(s.Y1)
	if err != nil {
		return err
	}
	x2Val, err := i.evaluate(s.X2)
	if err != nil {
		return err
	}
	y2Val, err := i.evaluate(s.Y2)
	if err != nil {
		return err
	}

	x1 := int(x1Val.ToInt())
	y1 := int(y1Val.ToInt())
	x2 := int(x2Val.ToInt())
	y2 := int(y2Val.ToInt())

	color := 15
	if s.Color != nil {
		colorVal, err := i.evaluate(s.Color)
		if err != nil {
			return err
		}
		color = int(colorVal.ToInt())
	}

	switch s.BoxFill {
	case "B":
		// Draw box outline
		i.drawLine(x1, y1, x2, y1, color)
		i.drawLine(x2, y1, x2, y2, color)
		i.drawLine(x2, y2, x1, y2, color)
		i.drawLine(x1, y2, x1, y1, color)
	case "BF":
		// Draw filled box
		for y := y1; y <= y2; y++ {
			for x := x1; x <= x2; x++ {
				i.drawPixel(x, y, color)
			}
		}
	default:
		// Draw line
		i.drawLine(x1, y1, x2, y2, color)
	}

	return nil
}

func (i *Interpreter) executeCircleStatement(s *ast.CircleStmt) error {
	xVal, err := i.evaluate(s.X)
	if err != nil {
		return err
	}
	yVal, err := i.evaluate(s.Y)
	if err != nil {
		return err
	}
	radiusVal, err := i.evaluate(s.Radius)
	if err != nil {
		return err
	}

	cx := int(xVal.ToInt())
	cy := int(yVal.ToInt())
	radius := int(radiusVal.ToInt())

	color := 15
	if s.Color != nil {
		colorVal, err := i.evaluate(s.Color)
		if err != nil {
			return err
		}
		color = int(colorVal.ToInt())
	}

	i.drawCircle(cx, cy, radius, color)
	return nil
}

// Graphics helper methods using Unicode block characters

func (i *Interpreter) drawPixel(x, y, color int) {
	if i.screen == nil {
		return
	}

	// In terminal mode, each character cell is 2x2 pixels using Unicode block chars
	// We'll use a simpler approach: one character = one pixel
	rows, cols := i.screen.GetSize()

	// Scale coordinates to terminal size
	termX := x * cols / 320
	termY := y * rows / 200

	if termX >= 0 && termX < cols && termY >= 0 && termY < rows {
		// Use full block character
		i.screen.SetCell(termX, termY, '█')
		i.screen.Show()
	}
}

func (i *Interpreter) drawLine(x1, y1, x2, y2, color int) {
	// Bresenham's line algorithm
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	for {
		i.drawPixel(x1, y1, color)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func (i *Interpreter) drawCircle(cx, cy, radius, color int) {
	// Midpoint circle algorithm
	x := radius
	y := 0
	err := 0

	for x >= y {
		i.drawPixel(cx+x, cy+y, color)
		i.drawPixel(cx+y, cy+x, color)
		i.drawPixel(cx-y, cy+x, color)
		i.drawPixel(cx-x, cy+y, color)
		i.drawPixel(cx-x, cy-y, color)
		i.drawPixel(cx-y, cy-x, color)
		i.drawPixel(cx+y, cy-x, color)
		i.drawPixel(cx+x, cy-y, color)

		y++
		if err <= 0 {
			err += 2*y + 1
		}
		if err > 0 {
			x--
			err -= 2*x + 1
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// GetFileHandle returns a file handle for built-in functions
func (i *Interpreter) GetFileHandle(fileNum int) (*FileHandle, bool) {
	fh, ok := i.files[fileNum]
	return fh, ok
}

// GetNextFreeFile returns the next available file number (1-255)
func (i *Interpreter) GetNextFreeFile() int {
	for n := 1; n <= 255; n++ {
		if _, exists := i.files[n]; !exists {
			return n
		}
	}
	return 0 // No free file numbers available
}
