package ast

import (
	"bytes"
	"fmt"
	"strings"
)

// LineNumberStmt represents a line number marker
type LineNumberStmt struct {
	Line   int
	Number int
}

func (ls *LineNumberStmt) statementNode()       {}
func (ls *LineNumberStmt) TokenLiteral() string { return fmt.Sprintf("%d", ls.Number) }
func (ls *LineNumberStmt) String() string       { return fmt.Sprintf("%d", ls.Number) }

// LabelStmt represents a label
type LabelStmt struct {
	Line int
	Name string
}

func (ls *LabelStmt) statementNode()       {}
func (ls *LabelStmt) TokenLiteral() string { return ls.Name }
func (ls *LabelStmt) String() string       { return ls.Name + ":" }

// LetStmt represents variable assignment (LET is optional)
type LetStmt struct {
	Line  int
	Name  Expression // Identifier or ArrayAccess
	Value Expression
}

func (ls *LetStmt) statementNode()       {}
func (ls *LetStmt) TokenLiteral() string { return "LET" }
func (ls *LetStmt) String() string {
	return ls.Name.String() + " = " + ls.Value.String()
}

// PrintStmt represents PRINT statement
type PrintStmt struct {
	Line      int
	Items     []PrintItem
	NoNewline bool // ends with ; or ,
}

func (ps *PrintStmt) statementNode()       {}
func (ps *PrintStmt) TokenLiteral() string { return "PRINT" }
func (ps *PrintStmt) String() string {
	var out bytes.Buffer
	out.WriteString("PRINT ")
	for i, item := range ps.Items {
		if i > 0 {
			if ps.Items[i-1].Separator != "" {
				out.WriteString(ps.Items[i-1].Separator)
				out.WriteString(" ")
			}
		}
		if item.Expression != nil {
			out.WriteString(item.Expression.String())
		}
	}
	return out.String()
}

// InputStmt represents INPUT statement
type InputStmt struct {
	Line      int
	Prompt    *StringLiteral
	Variables []Expression
}

func (is *InputStmt) statementNode()       {}
func (is *InputStmt) TokenLiteral() string { return "INPUT" }
func (is *InputStmt) String() string {
	var out bytes.Buffer
	out.WriteString("INPUT ")
	if is.Prompt != nil {
		out.WriteString(is.Prompt.String())
		out.WriteString("; ")
	}
	vars := make([]string, len(is.Variables))
	for i, v := range is.Variables {
		vars[i] = v.String()
	}
	out.WriteString(strings.Join(vars, ", "))
	return out.String()
}

// DimStmt represents array/variable declaration
type DimStmt struct {
	Line      int
	Variables []DimVariable
	Shared    bool
	Static    bool
}

func (ds *DimStmt) statementNode()       {}
func (ds *DimStmt) TokenLiteral() string { return "DIM" }
func (ds *DimStmt) String() string {
	var out bytes.Buffer
	out.WriteString("DIM ")
	if ds.Shared {
		out.WriteString("SHARED ")
	}
	if ds.Static {
		out.WriteString("STATIC ")
	}
	vars := make([]string, len(ds.Variables))
	for i, v := range ds.Variables {
		vars[i] = v.String()
	}
	out.WriteString(strings.Join(vars, ", "))
	return out.String()
}

// IfStmt represents IF/THEN/ELSE/END IF
type IfStmt struct {
	Line        int
	Condition   Expression
	Consequence []Statement
	Alternative []Statement // ELSE block
	SingleLine  bool        // single-line IF...THEN...ELSE
}

func (is *IfStmt) statementNode()       {}
func (is *IfStmt) TokenLiteral() string { return "IF" }
func (is *IfStmt) String() string {
	var out bytes.Buffer
	out.WriteString("IF ")
	out.WriteString(is.Condition.String())
	out.WriteString(" THEN")
	if is.SingleLine {
		out.WriteString(" ")
		for _, s := range is.Consequence {
			out.WriteString(s.String())
		}
		if len(is.Alternative) > 0 {
			out.WriteString(" ELSE ")
			for _, s := range is.Alternative {
				out.WriteString(s.String())
			}
		}
	} else {
		out.WriteString("\n")
		for _, s := range is.Consequence {
			out.WriteString("  ")
			out.WriteString(s.String())
			out.WriteString("\n")
		}
		if len(is.Alternative) > 0 {
			out.WriteString("ELSE\n")
			for _, s := range is.Alternative {
				out.WriteString("  ")
				out.WriteString(s.String())
				out.WriteString("\n")
			}
		}
		out.WriteString("END IF")
	}
	return out.String()
}

// ForStmt represents FOR/NEXT loop
type ForStmt struct {
	Line     int
	Variable *Identifier
	Start    Expression
	End      Expression
	Step     Expression // nil means step 1
	Body     []Statement
}

func (fs *ForStmt) statementNode()       {}
func (fs *ForStmt) TokenLiteral() string { return "FOR" }
func (fs *ForStmt) String() string {
	var out bytes.Buffer
	out.WriteString("FOR ")
	out.WriteString(fs.Variable.String())
	out.WriteString(" = ")
	out.WriteString(fs.Start.String())
	out.WriteString(" TO ")
	out.WriteString(fs.End.String())
	if fs.Step != nil {
		out.WriteString(" STEP ")
		out.WriteString(fs.Step.String())
	}
	out.WriteString("\n")
	for _, s := range fs.Body {
		out.WriteString("  ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	out.WriteString("NEXT")
	return out.String()
}

// WhileStmt represents WHILE/WEND loop
type WhileStmt struct {
	Line      int
	Condition Expression
	Body      []Statement
}

func (ws *WhileStmt) statementNode()       {}
func (ws *WhileStmt) TokenLiteral() string { return "WHILE" }
func (ws *WhileStmt) String() string {
	var out bytes.Buffer
	out.WriteString("WHILE ")
	out.WriteString(ws.Condition.String())
	out.WriteString("\n")
	for _, s := range ws.Body {
		out.WriteString("  ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	out.WriteString("WEND")
	return out.String()
}

// DoLoopStmt represents DO/LOOP
type DoLoopStmt struct {
	Line          int
	Condition     Expression
	ConditionPos  string // "PRE" or "POST"
	ConditionType string // "WHILE" or "UNTIL"
	Body          []Statement
}

func (dl *DoLoopStmt) statementNode()       {}
func (dl *DoLoopStmt) TokenLiteral() string { return "DO" }
func (dl *DoLoopStmt) String() string {
	var out bytes.Buffer
	out.WriteString("DO")
	if dl.ConditionPos == "PRE" && dl.Condition != nil {
		out.WriteString(" ")
		out.WriteString(dl.ConditionType)
		out.WriteString(" ")
		out.WriteString(dl.Condition.String())
	}
	out.WriteString("\n")
	for _, s := range dl.Body {
		out.WriteString("  ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	out.WriteString("LOOP")
	if dl.ConditionPos == "POST" && dl.Condition != nil {
		out.WriteString(" ")
		out.WriteString(dl.ConditionType)
		out.WriteString(" ")
		out.WriteString(dl.Condition.String())
	}
	return out.String()
}

// SelectCaseStmt represents SELECT CASE
type SelectCaseStmt struct {
	Line       int
	Expression Expression
	Cases      []CaseClause
	CaseElse   []Statement
}

func (sc *SelectCaseStmt) statementNode()       {}
func (sc *SelectCaseStmt) TokenLiteral() string { return "SELECT" }
func (sc *SelectCaseStmt) String() string {
	var out bytes.Buffer
	out.WriteString("SELECT CASE ")
	out.WriteString(sc.Expression.String())
	out.WriteString("\n")
	for _, c := range sc.Cases {
		out.WriteString("CASE ")
		vals := make([]string, len(c.Values))
		for i, v := range c.Values {
			vals[i] = v.String()
		}
		out.WriteString(strings.Join(vals, ", "))
		out.WriteString("\n")
		for _, s := range c.Body {
			out.WriteString("  ")
			out.WriteString(s.String())
			out.WriteString("\n")
		}
	}
	if len(sc.CaseElse) > 0 {
		out.WriteString("CASE ELSE\n")
		for _, s := range sc.CaseElse {
			out.WriteString("  ")
			out.WriteString(s.String())
			out.WriteString("\n")
		}
	}
	out.WriteString("END SELECT")
	return out.String()
}

// GotoStmt represents GOTO
type GotoStmt struct {
	Line   int
	Target string // label name or line number as string
}

func (gs *GotoStmt) statementNode()       {}
func (gs *GotoStmt) TokenLiteral() string { return "GOTO" }
func (gs *GotoStmt) String() string       { return "GOTO " + gs.Target }

// GosubStmt represents GOSUB
type GosubStmt struct {
	Line   int
	Target string
}

func (gs *GosubStmt) statementNode()       {}
func (gs *GosubStmt) TokenLiteral() string { return "GOSUB" }
func (gs *GosubStmt) String() string       { return "GOSUB " + gs.Target }

// ReturnStmt represents RETURN
type ReturnStmt struct {
	Line  int
	Value Expression // for FUNCTION return, nil for GOSUB return
}

func (rs *ReturnStmt) statementNode()       {}
func (rs *ReturnStmt) TokenLiteral() string { return "RETURN" }
func (rs *ReturnStmt) String() string {
	if rs.Value != nil {
		return "RETURN " + rs.Value.String()
	}
	return "RETURN"
}

// ExitStmt represents EXIT FOR/DO/SUB/FUNCTION
type ExitStmt struct {
	Line     int
	ExitType string // "FOR", "DO", "SUB", "FUNCTION", "WHILE"
}

func (es *ExitStmt) statementNode()       {}
func (es *ExitStmt) TokenLiteral() string { return "EXIT" }
func (es *ExitStmt) String() string       { return "EXIT " + es.ExitType }

// SubStatement represents SUB definition
type SubStatement struct {
	Line       int
	Name       string
	Parameters []Parameter
	Body       []Statement
	Static     bool
}

func (ss *SubStatement) statementNode()       {}
func (ss *SubStatement) TokenLiteral() string { return "SUB" }
func (ss *SubStatement) String() string {
	var out bytes.Buffer
	out.WriteString("SUB ")
	out.WriteString(ss.Name)
	out.WriteString("(")
	params := make([]string, len(ss.Parameters))
	for i, p := range ss.Parameters {
		params[i] = p.String()
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if ss.Static {
		out.WriteString(" STATIC")
	}
	out.WriteString("\n")
	for _, s := range ss.Body {
		out.WriteString("  ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	out.WriteString("END SUB")
	return out.String()
}

// FuncStatement represents FUNCTION definition
type FuncStatement struct {
	Line       int
	Name       string
	Parameters []Parameter
	ReturnType DataType
	Body       []Statement
	Static     bool
}

func (fs *FuncStatement) statementNode()       {}
func (fs *FuncStatement) TokenLiteral() string { return "FUNCTION" }
func (fs *FuncStatement) String() string {
	var out bytes.Buffer
	out.WriteString("FUNCTION ")
	out.WriteString(fs.Name)
	out.WriteString("(")
	params := make([]string, len(fs.Parameters))
	for i, p := range fs.Parameters {
		params[i] = p.String()
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if fs.Static {
		out.WriteString(" STATIC")
	}
	out.WriteString("\n")
	for _, s := range fs.Body {
		out.WriteString("  ")
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	out.WriteString("END FUNCTION")
	return out.String()
}

// DataStmt represents DATA statement
type DataStmt struct {
	Line   int
	Values []Expression
}

func (ds *DataStmt) statementNode()       {}
func (ds *DataStmt) TokenLiteral() string { return "DATA" }
func (ds *DataStmt) String() string {
	var out bytes.Buffer
	out.WriteString("DATA ")
	vals := make([]string, len(ds.Values))
	for i, v := range ds.Values {
		vals[i] = v.String()
	}
	out.WriteString(strings.Join(vals, ", "))
	return out.String()
}

// ReadStmt represents READ statement
type ReadStmt struct {
	Line      int
	Variables []Expression
}

func (rs *ReadStmt) statementNode()       {}
func (rs *ReadStmt) TokenLiteral() string { return "READ" }
func (rs *ReadStmt) String() string {
	var out bytes.Buffer
	out.WriteString("READ ")
	vars := make([]string, len(rs.Variables))
	for i, v := range rs.Variables {
		vars[i] = v.String()
	}
	out.WriteString(strings.Join(vars, ", "))
	return out.String()
}

// RestoreStmt represents RESTORE statement
type RestoreStmt struct {
	Line   int
	Target string // optional label/line number
}

func (rs *RestoreStmt) statementNode()       {}
func (rs *RestoreStmt) TokenLiteral() string { return "RESTORE" }
func (rs *RestoreStmt) String() string {
	if rs.Target != "" {
		return "RESTORE " + rs.Target
	}
	return "RESTORE"
}

// ClsStmt represents CLS statement
type ClsStmt struct {
	Line int
}

func (cs *ClsStmt) statementNode()       {}
func (cs *ClsStmt) TokenLiteral() string { return "CLS" }
func (cs *ClsStmt) String() string       { return "CLS" }

// LocateStmt represents LOCATE statement
type LocateStmt struct {
	Line   int
	Row    Expression
	Column Expression
}

func (ls *LocateStmt) statementNode()       {}
func (ls *LocateStmt) TokenLiteral() string { return "LOCATE" }
func (ls *LocateStmt) String() string {
	var out bytes.Buffer
	out.WriteString("LOCATE ")
	if ls.Row != nil {
		out.WriteString(ls.Row.String())
	}
	out.WriteString(", ")
	if ls.Column != nil {
		out.WriteString(ls.Column.String())
	}
	return out.String()
}

// ColorStmt represents COLOR statement
type ColorStmt struct {
	Line       int
	Foreground Expression
	Background Expression
}

func (cs *ColorStmt) statementNode()       {}
func (cs *ColorStmt) TokenLiteral() string { return "COLOR" }
func (cs *ColorStmt) String() string {
	var out bytes.Buffer
	out.WriteString("COLOR ")
	if cs.Foreground != nil {
		out.WriteString(cs.Foreground.String())
	}
	if cs.Background != nil {
		out.WriteString(", ")
		out.WriteString(cs.Background.String())
	}
	return out.String()
}

// ScreenStmt represents SCREEN statement
type ScreenStmt struct {
	Line int
	Mode Expression
}

func (ss *ScreenStmt) statementNode()       {}
func (ss *ScreenStmt) TokenLiteral() string { return "SCREEN" }
func (ss *ScreenStmt) String() string {
	return "SCREEN " + ss.Mode.String()
}

// EndStmt represents END statement
type EndStmt struct {
	Line int
}

func (es *EndStmt) statementNode()       {}
func (es *EndStmt) TokenLiteral() string { return "END" }
func (es *EndStmt) String() string       { return "END" }

// RemStmt represents REM (comment) statement
type RemStmt struct {
	Line    int
	Comment string
}

func (rs *RemStmt) statementNode()       {}
func (rs *RemStmt) TokenLiteral() string { return "REM" }
func (rs *RemStmt) String() string       { return "REM " + rs.Comment }

// CallStmt represents CALL statement (calling a SUB)
type CallStmt struct {
	Line      int
	Name      string
	Arguments []Expression
}

func (cs *CallStmt) statementNode()       {}
func (cs *CallStmt) TokenLiteral() string { return "CALL" }
func (cs *CallStmt) String() string {
	var out bytes.Buffer
	out.WriteString("CALL ")
	out.WriteString(cs.Name)
	if len(cs.Arguments) > 0 {
		out.WriteString("(")
		args := make([]string, len(cs.Arguments))
		for i, arg := range cs.Arguments {
			args[i] = arg.String()
		}
		out.WriteString(strings.Join(args, ", "))
		out.WriteString(")")
	}
	return out.String()
}

// SubCallStmt represents an implicit sub call (without CALL keyword)
type SubCallStmt struct {
	Line      int
	Name      string
	Arguments []Expression
}

func (sc *SubCallStmt) statementNode()       {}
func (sc *SubCallStmt) TokenLiteral() string { return sc.Name }
func (sc *SubCallStmt) String() string {
	var out bytes.Buffer
	out.WriteString(sc.Name)
	if len(sc.Arguments) > 0 {
		out.WriteString(" ")
		args := make([]string, len(sc.Arguments))
		for i, arg := range sc.Arguments {
			args[i] = arg.String()
		}
		out.WriteString(strings.Join(args, ", "))
	}
	return out.String()
}

// SleepStmt represents SLEEP statement
type SleepStmt struct {
	Line    int
	Seconds Expression // nil for indefinite
}

func (ss *SleepStmt) statementNode()       {}
func (ss *SleepStmt) TokenLiteral() string { return "SLEEP" }
func (ss *SleepStmt) String() string {
	if ss.Seconds != nil {
		return "SLEEP " + ss.Seconds.String()
	}
	return "SLEEP"
}

// BeepStmt represents BEEP statement
type BeepStmt struct {
	Line int
}

func (bs *BeepStmt) statementNode()       {}
func (bs *BeepStmt) TokenLiteral() string { return "BEEP" }
func (bs *BeepStmt) String() string       { return "BEEP" }

// SwapStmt represents SWAP statement
type SwapStmt struct {
	Line int
	Var1 Expression
	Var2 Expression
}

func (ss *SwapStmt) statementNode()       {}
func (ss *SwapStmt) TokenLiteral() string { return "SWAP" }
func (ss *SwapStmt) String() string {
	return "SWAP " + ss.Var1.String() + ", " + ss.Var2.String()
}

// RandomizeStmt represents RANDOMIZE statement
type RandomizeStmt struct {
	Line int
	Seed Expression // nil for RANDOMIZE TIMER
}

func (rs *RandomizeStmt) statementNode()       {}
func (rs *RandomizeStmt) TokenLiteral() string { return "RANDOMIZE" }
func (rs *RandomizeStmt) String() string {
	if rs.Seed != nil {
		return "RANDOMIZE " + rs.Seed.String()
	}
	return "RANDOMIZE"
}

// ConstStmt represents CONST statement
type ConstStmt struct {
	Line  int
	Name  string
	Value Expression
}

func (cs *ConstStmt) statementNode()       {}
func (cs *ConstStmt) TokenLiteral() string { return "CONST" }
func (cs *ConstStmt) String() string {
	return "CONST " + cs.Name + " = " + cs.Value.String()
}

// OpenStmt represents OPEN statement for file I/O
type OpenStmt struct {
	Line     int
	Filename Expression
	Mode     string     // "INPUT", "OUTPUT", "APPEND", "BINARY", "RANDOM"
	FileNum  Expression // #n
	RecLen   Expression // for RANDOM access
}

func (os *OpenStmt) statementNode()       {}
func (os *OpenStmt) TokenLiteral() string { return "OPEN" }
func (os *OpenStmt) String() string {
	var out bytes.Buffer
	out.WriteString("OPEN ")
	out.WriteString(os.Filename.String())
	out.WriteString(" FOR ")
	out.WriteString(os.Mode)
	out.WriteString(" AS #")
	out.WriteString(os.FileNum.String())
	return out.String()
}

// CloseStmt represents CLOSE statement
type CloseStmt struct {
	Line     int
	FileNums []Expression // empty means close all
}

func (cs *CloseStmt) statementNode()       {}
func (cs *CloseStmt) TokenLiteral() string { return "CLOSE" }
func (cs *CloseStmt) String() string {
	var out bytes.Buffer
	out.WriteString("CLOSE")
	if len(cs.FileNums) > 0 {
		out.WriteString(" #")
		nums := make([]string, len(cs.FileNums))
		for i, n := range cs.FileNums {
			nums[i] = n.String()
		}
		out.WriteString(strings.Join(nums, ", #"))
	}
	return out.String()
}

// PrintFileStmt represents PRINT #n statement
type PrintFileStmt struct {
	Line      int
	FileNum   Expression
	Items     []PrintItem
	NoNewline bool
}

func (pf *PrintFileStmt) statementNode()       {}
func (pf *PrintFileStmt) TokenLiteral() string { return "PRINT" }
func (pf *PrintFileStmt) String() string {
	var out bytes.Buffer
	out.WriteString("PRINT #")
	out.WriteString(pf.FileNum.String())
	out.WriteString(", ")
	for i, item := range pf.Items {
		if i > 0 && pf.Items[i-1].Separator != "" {
			out.WriteString(pf.Items[i-1].Separator)
			out.WriteString(" ")
		}
		if item.Expression != nil {
			out.WriteString(item.Expression.String())
		}
	}
	return out.String()
}

// InputFileStmt represents INPUT #n statement
type InputFileStmt struct {
	Line      int
	FileNum   Expression
	Variables []Expression
}

func (if_ *InputFileStmt) statementNode()       {}
func (if_ *InputFileStmt) TokenLiteral() string { return "INPUT" }
func (if_ *InputFileStmt) String() string {
	var out bytes.Buffer
	out.WriteString("INPUT #")
	out.WriteString(if_.FileNum.String())
	out.WriteString(", ")
	vars := make([]string, len(if_.Variables))
	for i, v := range if_.Variables {
		vars[i] = v.String()
	}
	out.WriteString(strings.Join(vars, ", "))
	return out.String()
}

// LineInputStmt represents LINE INPUT statement
type LineInputStmt struct {
	Line     int
	Prompt   *StringLiteral
	Variable Expression
}

func (li *LineInputStmt) statementNode()       {}
func (li *LineInputStmt) TokenLiteral() string { return "LINE INPUT" }
func (li *LineInputStmt) String() string {
	var out bytes.Buffer
	out.WriteString("LINE INPUT ")
	if li.Prompt != nil {
		out.WriteString(li.Prompt.String())
		out.WriteString("; ")
	}
	out.WriteString(li.Variable.String())
	return out.String()
}

// LineInputFileStmt represents LINE INPUT #n statement
type LineInputFileStmt struct {
	Line     int
	FileNum  Expression
	Variable Expression
}

func (lif *LineInputFileStmt) statementNode()       {}
func (lif *LineInputFileStmt) TokenLiteral() string { return "LINE INPUT" }
func (lif *LineInputFileStmt) String() string {
	var out bytes.Buffer
	out.WriteString("LINE INPUT #")
	out.WriteString(lif.FileNum.String())
	out.WriteString(", ")
	out.WriteString(lif.Variable.String())
	return out.String()
}

// OnGotoStmt represents ON...GOTO statement
type OnGotoStmt struct {
	Line       int
	Expression Expression
	Targets    []string
}

func (og *OnGotoStmt) statementNode()       {}
func (og *OnGotoStmt) TokenLiteral() string { return "ON" }
func (og *OnGotoStmt) String() string {
	var out bytes.Buffer
	out.WriteString("ON ")
	out.WriteString(og.Expression.String())
	out.WriteString(" GOTO ")
	out.WriteString(strings.Join(og.Targets, ", "))
	return out.String()
}

// OnGosubStmt represents ON...GOSUB statement
type OnGosubStmt struct {
	Line       int
	Expression Expression
	Targets    []string
}

func (og *OnGosubStmt) statementNode()       {}
func (og *OnGosubStmt) TokenLiteral() string { return "ON" }
func (og *OnGosubStmt) String() string {
	var out bytes.Buffer
	out.WriteString("ON ")
	out.WriteString(og.Expression.String())
	out.WriteString(" GOSUB ")
	out.WriteString(strings.Join(og.Targets, ", "))
	return out.String()
}

// GetStmt represents GET #n, position, variable (binary file I/O)
type GetStmt struct {
	Line     int
	FileNum  Expression
	Position Expression // optional position
	Variable Expression
}

func (gs *GetStmt) statementNode()       {}
func (gs *GetStmt) TokenLiteral() string { return "GET" }
func (gs *GetStmt) String() string {
	var out bytes.Buffer
	out.WriteString("GET #")
	out.WriteString(gs.FileNum.String())
	if gs.Position != nil {
		out.WriteString(", ")
		out.WriteString(gs.Position.String())
	}
	out.WriteString(", ")
	out.WriteString(gs.Variable.String())
	return out.String()
}

// PutStmt represents PUT #n, position, variable (binary file I/O)
type PutStmt struct {
	Line     int
	FileNum  Expression
	Position Expression // optional position
	Variable Expression
}

func (ps *PutStmt) statementNode()       {}
func (ps *PutStmt) TokenLiteral() string { return "PUT" }
func (ps *PutStmt) String() string {
	var out bytes.Buffer
	out.WriteString("PUT #")
	out.WriteString(ps.FileNum.String())
	if ps.Position != nil {
		out.WriteString(", ")
		out.WriteString(ps.Position.String())
	}
	out.WriteString(", ")
	out.WriteString(ps.Variable.String())
	return out.String()
}

// SeekStmt represents SEEK #n, position
type SeekStmt struct {
	Line     int
	FileNum  Expression
	Position Expression
}

func (ss *SeekStmt) statementNode()       {}
func (ss *SeekStmt) TokenLiteral() string { return "SEEK" }
func (ss *SeekStmt) String() string {
	return fmt.Sprintf("SEEK #%s, %s", ss.FileNum.String(), ss.Position.String())
}

// RedimStmt represents REDIM [PRESERVE] array(newsize)
type RedimStmt struct {
	Line      int
	Preserve  bool
	Variables []DimVariable
}

func (rs *RedimStmt) statementNode()       {}
func (rs *RedimStmt) TokenLiteral() string { return "REDIM" }
func (rs *RedimStmt) String() string {
	var out bytes.Buffer
	out.WriteString("REDIM ")
	if rs.Preserve {
		out.WriteString("PRESERVE ")
	}
	vars := make([]string, len(rs.Variables))
	for i, v := range rs.Variables {
		vars[i] = v.String()
	}
	out.WriteString(strings.Join(vars, ", "))
	return out.String()
}

// PsetStmt represents PSET (x, y), color
type PsetStmt struct {
	Line  int
	X     Expression
	Y     Expression
	Color Expression // optional
}

func (ps *PsetStmt) statementNode()       {}
func (ps *PsetStmt) TokenLiteral() string { return "PSET" }
func (ps *PsetStmt) String() string {
	var out bytes.Buffer
	out.WriteString("PSET (")
	out.WriteString(ps.X.String())
	out.WriteString(", ")
	out.WriteString(ps.Y.String())
	out.WriteString(")")
	if ps.Color != nil {
		out.WriteString(", ")
		out.WriteString(ps.Color.String())
	}
	return out.String()
}

// LineGraphicsStmt represents LINE (x1, y1)-(x2, y2), color, BF
type LineGraphicsStmt struct {
	Line   int
	X1     Expression
	Y1     Expression
	X2     Expression
	Y2     Expression
	Color  Expression // optional
	BoxFill string    // "B", "BF", or empty
}

func (ls *LineGraphicsStmt) statementNode()       {}
func (ls *LineGraphicsStmt) TokenLiteral() string { return "LINE" }
func (ls *LineGraphicsStmt) String() string {
	var out bytes.Buffer
	out.WriteString("LINE (")
	out.WriteString(ls.X1.String())
	out.WriteString(", ")
	out.WriteString(ls.Y1.String())
	out.WriteString(")-(")
	out.WriteString(ls.X2.String())
	out.WriteString(", ")
	out.WriteString(ls.Y2.String())
	out.WriteString(")")
	if ls.Color != nil {
		out.WriteString(", ")
		out.WriteString(ls.Color.String())
	}
	if ls.BoxFill != "" {
		out.WriteString(", ")
		out.WriteString(ls.BoxFill)
	}
	return out.String()
}

// CircleStmt represents CIRCLE (x, y), radius, color
type CircleStmt struct {
	Line   int
	X      Expression
	Y      Expression
	Radius Expression
	Color  Expression // optional
}

func (cs *CircleStmt) statementNode()       {}
func (cs *CircleStmt) TokenLiteral() string { return "CIRCLE" }
func (cs *CircleStmt) String() string {
	var out bytes.Buffer
	out.WriteString("CIRCLE (")
	out.WriteString(cs.X.String())
	out.WriteString(", ")
	out.WriteString(cs.Y.String())
	out.WriteString("), ")
	out.WriteString(cs.Radius.String())
	if cs.Color != nil {
		out.WriteString(", ")
		out.WriteString(cs.Color.String())
	}
	return out.String()
}

// PrintUsingStmt represents PRINT USING format$; expression
type PrintUsingStmt struct {
	Line      int
	FileNum   Expression // nil for screen output, #n for file output
	Format    Expression
	Items     []PrintItem
	NoNewline bool
}

func (pu *PrintUsingStmt) statementNode()       {}
func (pu *PrintUsingStmt) TokenLiteral() string { return "PRINT USING" }
func (pu *PrintUsingStmt) String() string {
	var out bytes.Buffer
	out.WriteString("PRINT ")
	if pu.FileNum != nil {
		out.WriteString("#")
		out.WriteString(pu.FileNum.String())
		out.WriteString(", ")
	}
	out.WriteString("USING ")
	out.WriteString(pu.Format.String())
	out.WriteString("; ")
	for i, item := range pu.Items {
		if i > 0 {
			out.WriteString("; ")
		}
		if item.Expression != nil {
			out.WriteString(item.Expression.String())
		}
	}
	return out.String()
}

// LineInputFileStmt already exists, but adding for completeness
