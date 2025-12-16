package lexer

import "fmt"

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF
	TOKEN_NEWLINE

	// Literals
	TOKEN_INTEGER  // 123
	TOKEN_FLOAT    // 123.45 or 1.23E+10
	TOKEN_STRING   // "hello"
	TOKEN_IDENT    // variable name

	// Line number / label
	TOKEN_LINE_NUMBER // 10, 20, 30...

	// Keywords - Control Flow
	TOKEN_IF
	TOKEN_THEN
	TOKEN_ELSE
	TOKEN_ELSEIF
	TOKEN_END
	TOKEN_FOR
	TOKEN_TO
	TOKEN_STEP
	TOKEN_NEXT
	TOKEN_WHILE
	TOKEN_WEND
	TOKEN_DO
	TOKEN_LOOP
	TOKEN_UNTIL
	TOKEN_SELECT
	TOKEN_CASE
	TOKEN_GOTO
	TOKEN_GOSUB
	TOKEN_RETURN
	TOKEN_EXIT
	TOKEN_ON

	// Keywords - Declarations
	TOKEN_DIM
	TOKEN_AS
	TOKEN_SUB
	TOKEN_FUNCTION
	TOKEN_STATIC
	TOKEN_SHARED
	TOKEN_CONST
	TOKEN_TYPE
	TOKEN_LET
	TOKEN_DECLARE
	TOKEN_BYVAL
	TOKEN_BYREF

	// Keywords - Data Types
	TOKEN_INTEGER_TYPE  // INTEGER
	TOKEN_LONG_TYPE     // LONG
	TOKEN_SINGLE_TYPE   // SINGLE
	TOKEN_DOUBLE_TYPE   // DOUBLE
	TOKEN_STRING_TYPE   // STRING

	// Keywords - Data
	TOKEN_DATA
	TOKEN_READ
	TOKEN_RESTORE

	// Keywords - I/O
	TOKEN_PRINT
	TOKEN_INPUT
	TOKEN_OPEN
	TOKEN_CLOSE
	TOKEN_OUTPUT
	TOKEN_APPEND
	TOKEN_BINARY
	TOKEN_RANDOM
	TOKEN_ACCESS
	TOKEN_WRITE
	TOKEN_LINE
	TOKEN_USING
	TOKEN_TAB
	TOKEN_SPC

	// Keywords - Screen/Graphics
	TOKEN_SCREEN
	TOKEN_CLS
	TOKEN_LOCATE
	TOKEN_COLOR
	TOKEN_BEEP

	// Keywords - Misc
	TOKEN_REM
	TOKEN_OPTION
	TOKEN_BASE
	TOKEN_DEF
	TOKEN_SEG
	TOKEN_CALL
	TOKEN_SLEEP
	TOKEN_SYSTEM
	TOKEN_SHELL
	TOKEN_SWAP
	TOKEN_RANDOMIZE
	TOKEN_REDIM
	TOKEN_PRESERVE
	TOKEN_GET
	TOKEN_PUT
	TOKEN_SEEK
	TOKEN_LEN_KW // LEN keyword for OPEN ... LEN = n

	// Keywords - Graphics
	TOKEN_PSET
	TOKEN_CIRCLE

	// Operators - Arithmetic
	TOKEN_PLUS      // +
	TOKEN_MINUS     // -
	TOKEN_ASTERISK  // *
	TOKEN_SLASH     // /
	TOKEN_BACKSLASH // \ (integer division)
	TOKEN_CARET     // ^ (exponentiation)
	TOKEN_MOD       // MOD

	// Operators - Comparison
	TOKEN_EQ // =
	TOKEN_NE // <> or ><
	TOKEN_LT // <
	TOKEN_GT // >
	TOKEN_LE // <=
	TOKEN_GE // >=

	// Operators - Logical
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	TOKEN_XOR
	TOKEN_EQV
	TOKEN_IMP

	// Delimiters
	TOKEN_LPAREN    // (
	TOKEN_RPAREN    // )
	TOKEN_COMMA     // ,
	TOKEN_SEMICOLON // ;
	TOKEN_COLON     // :
	TOKEN_HASH      // # (file number prefix)
	TOKEN_DOLLAR    // $ (string type suffix)
	TOKEN_PERCENT   // % (integer type suffix)
	TOKEN_AMPERSAND // & (long type suffix)
	TOKEN_BANG      // ! (single type suffix)
	TOKEN_QUESTION  // ?
)

var tokenNames = map[TokenType]string{
	TOKEN_ILLEGAL:      "ILLEGAL",
	TOKEN_EOF:          "EOF",
	TOKEN_NEWLINE:      "NEWLINE",
	TOKEN_INTEGER:      "INTEGER",
	TOKEN_FLOAT:        "FLOAT",
	TOKEN_STRING:       "STRING",
	TOKEN_IDENT:        "IDENT",
	TOKEN_LINE_NUMBER:  "LINE_NUMBER",
	TOKEN_IF:           "IF",
	TOKEN_THEN:         "THEN",
	TOKEN_ELSE:         "ELSE",
	TOKEN_ELSEIF:       "ELSEIF",
	TOKEN_END:          "END",
	TOKEN_FOR:          "FOR",
	TOKEN_TO:           "TO",
	TOKEN_STEP:         "STEP",
	TOKEN_NEXT:         "NEXT",
	TOKEN_WHILE:        "WHILE",
	TOKEN_WEND:         "WEND",
	TOKEN_DO:           "DO",
	TOKEN_LOOP:         "LOOP",
	TOKEN_UNTIL:        "UNTIL",
	TOKEN_SELECT:       "SELECT",
	TOKEN_CASE:         "CASE",
	TOKEN_GOTO:         "GOTO",
	TOKEN_GOSUB:        "GOSUB",
	TOKEN_RETURN:       "RETURN",
	TOKEN_EXIT:         "EXIT",
	TOKEN_ON:           "ON",
	TOKEN_DIM:          "DIM",
	TOKEN_AS:           "AS",
	TOKEN_SUB:          "SUB",
	TOKEN_FUNCTION:     "FUNCTION",
	TOKEN_STATIC:       "STATIC",
	TOKEN_SHARED:       "SHARED",
	TOKEN_CONST:        "CONST",
	TOKEN_TYPE:         "TYPE",
	TOKEN_LET:          "LET",
	TOKEN_DECLARE:      "DECLARE",
	TOKEN_BYVAL:        "BYVAL",
	TOKEN_BYREF:        "BYREF",
	TOKEN_INTEGER_TYPE: "INTEGER_TYPE",
	TOKEN_LONG_TYPE:    "LONG_TYPE",
	TOKEN_SINGLE_TYPE:  "SINGLE_TYPE",
	TOKEN_DOUBLE_TYPE:  "DOUBLE_TYPE",
	TOKEN_STRING_TYPE:  "STRING_TYPE",
	TOKEN_DATA:         "DATA",
	TOKEN_READ:         "READ",
	TOKEN_RESTORE:      "RESTORE",
	TOKEN_PRINT:        "PRINT",
	TOKEN_INPUT:        "INPUT",
	TOKEN_OPEN:         "OPEN",
	TOKEN_CLOSE:        "CLOSE",
	TOKEN_OUTPUT:       "OUTPUT",
	TOKEN_APPEND:       "APPEND",
	TOKEN_BINARY:       "BINARY",
	TOKEN_RANDOM:       "RANDOM",
	TOKEN_ACCESS:       "ACCESS",
	TOKEN_WRITE:        "WRITE",
	TOKEN_LINE:         "LINE",
	TOKEN_USING:        "USING",
	TOKEN_TAB:          "TAB",
	TOKEN_SPC:          "SPC",
	TOKEN_SCREEN:       "SCREEN",
	TOKEN_CLS:          "CLS",
	TOKEN_LOCATE:       "LOCATE",
	TOKEN_COLOR:        "COLOR",
	TOKEN_BEEP:         "BEEP",
	TOKEN_REM:          "REM",
	TOKEN_OPTION:       "OPTION",
	TOKEN_BASE:         "BASE",
	TOKEN_DEF:          "DEF",
	TOKEN_SEG:          "SEG",
	TOKEN_CALL:         "CALL",
	TOKEN_SLEEP:        "SLEEP",
	TOKEN_SYSTEM:       "SYSTEM",
	TOKEN_SHELL:        "SHELL",
	TOKEN_SWAP:         "SWAP",
	TOKEN_RANDOMIZE:    "RANDOMIZE",
	TOKEN_REDIM:        "REDIM",
	TOKEN_PRESERVE:     "PRESERVE",
	TOKEN_GET:          "GET",
	TOKEN_PUT:          "PUT",
	TOKEN_SEEK:         "SEEK",
	TOKEN_LEN_KW:       "LEN_KW",
	TOKEN_PSET:         "PSET",
	TOKEN_CIRCLE:       "CIRCLE",
	TOKEN_PLUS:         "PLUS",
	TOKEN_MINUS:        "MINUS",
	TOKEN_ASTERISK:     "ASTERISK",
	TOKEN_SLASH:        "SLASH",
	TOKEN_BACKSLASH:    "BACKSLASH",
	TOKEN_CARET:        "CARET",
	TOKEN_MOD:          "MOD",
	TOKEN_EQ:           "EQ",
	TOKEN_NE:           "NE",
	TOKEN_LT:           "LT",
	TOKEN_GT:           "GT",
	TOKEN_LE:           "LE",
	TOKEN_GE:           "GE",
	TOKEN_AND:          "AND",
	TOKEN_OR:           "OR",
	TOKEN_NOT:          "NOT",
	TOKEN_XOR:          "XOR",
	TOKEN_EQV:          "EQV",
	TOKEN_IMP:          "IMP",
	TOKEN_LPAREN:       "LPAREN",
	TOKEN_RPAREN:       "RPAREN",
	TOKEN_COMMA:        "COMMA",
	TOKEN_SEMICOLON:    "SEMICOLON",
	TOKEN_COLON:        "COLON",
	TOKEN_HASH:         "HASH",
	TOKEN_DOLLAR:       "DOLLAR",
	TOKEN_PERCENT:      "PERCENT",
	TOKEN_AMPERSAND:    "AMPERSAND",
	TOKEN_BANG:         "BANG",
	TOKEN_QUESTION:     "QUESTION",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", t)
}

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

func (t Token) String() string {
	return fmt.Sprintf("Token{%s, %q, %d:%d}", t.Type, t.Literal, t.Line, t.Column)
}

// Keywords maps keyword strings to their token types (case-insensitive)
var Keywords = map[string]TokenType{
	"IF":        TOKEN_IF,
	"THEN":      TOKEN_THEN,
	"ELSE":      TOKEN_ELSE,
	"ELSEIF":    TOKEN_ELSEIF,
	"END":       TOKEN_END,
	"FOR":       TOKEN_FOR,
	"TO":        TOKEN_TO,
	"STEP":      TOKEN_STEP,
	"NEXT":      TOKEN_NEXT,
	"WHILE":     TOKEN_WHILE,
	"WEND":      TOKEN_WEND,
	"DO":        TOKEN_DO,
	"LOOP":      TOKEN_LOOP,
	"UNTIL":     TOKEN_UNTIL,
	"SELECT":    TOKEN_SELECT,
	"CASE":      TOKEN_CASE,
	"GOTO":      TOKEN_GOTO,
	"GOSUB":     TOKEN_GOSUB,
	"RETURN":    TOKEN_RETURN,
	"EXIT":      TOKEN_EXIT,
	"ON":        TOKEN_ON,
	"DIM":       TOKEN_DIM,
	"AS":        TOKEN_AS,
	"SUB":       TOKEN_SUB,
	"FUNCTION":  TOKEN_FUNCTION,
	"STATIC":    TOKEN_STATIC,
	"SHARED":    TOKEN_SHARED,
	"CONST":     TOKEN_CONST,
	"TYPE":      TOKEN_TYPE,
	"LET":       TOKEN_LET,
	"DECLARE":   TOKEN_DECLARE,
	"BYVAL":     TOKEN_BYVAL,
	"BYREF":     TOKEN_BYREF,
	"INTEGER":   TOKEN_INTEGER_TYPE,
	"LONG":      TOKEN_LONG_TYPE,
	"SINGLE":    TOKEN_SINGLE_TYPE,
	"DOUBLE":    TOKEN_DOUBLE_TYPE,
	"STRING":    TOKEN_STRING_TYPE,
	"DATA":      TOKEN_DATA,
	"READ":      TOKEN_READ,
	"RESTORE":   TOKEN_RESTORE,
	"PRINT":     TOKEN_PRINT,
	"INPUT":     TOKEN_INPUT,
	"OPEN":      TOKEN_OPEN,
	"CLOSE":     TOKEN_CLOSE,
	"OUTPUT":    TOKEN_OUTPUT,
	"APPEND":    TOKEN_APPEND,
	"BINARY":    TOKEN_BINARY,
	"RANDOM":    TOKEN_RANDOM,
	"ACCESS":    TOKEN_ACCESS,
	"WRITE":     TOKEN_WRITE,
	"LINE":      TOKEN_LINE,
	"USING":     TOKEN_USING,
	"TAB":       TOKEN_TAB,
	"SPC":       TOKEN_SPC,
	"SCREEN":    TOKEN_SCREEN,
	"CLS":       TOKEN_CLS,
	"LOCATE":    TOKEN_LOCATE,
	"COLOR":     TOKEN_COLOR,
	"BEEP":      TOKEN_BEEP,
	"REM":       TOKEN_REM,
	"OPTION":    TOKEN_OPTION,
	"BASE":      TOKEN_BASE,
	"DEF":       TOKEN_DEF,
	"SEG":       TOKEN_SEG,
	"CALL":      TOKEN_CALL,
	"SLEEP":     TOKEN_SLEEP,
	"SYSTEM":    TOKEN_SYSTEM,
	"SHELL":     TOKEN_SHELL,
	"SWAP":      TOKEN_SWAP,
	"RANDOMIZE": TOKEN_RANDOMIZE,
	"REDIM":     TOKEN_REDIM,
	"PRESERVE":  TOKEN_PRESERVE,
	"GET":       TOKEN_GET,
	"PUT":       TOKEN_PUT,
	"SEEK":      TOKEN_SEEK,
	"PSET":      TOKEN_PSET,
	"CIRCLE":    TOKEN_CIRCLE,
	"MOD":       TOKEN_MOD,
	"AND":       TOKEN_AND,
	"OR":        TOKEN_OR,
	"NOT":       TOKEN_NOT,
	"XOR":       TOKEN_XOR,
	"EQV":       TOKEN_EQV,
	"IMP":       TOKEN_IMP,
}

// LookupIdent checks if an identifier is a keyword
func LookupIdent(ident string) TokenType {
	if tok, ok := Keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENT
}
