package parser

import "github.com/xbasic/xbasic/internal/lexer"

// Operator precedence levels (lowest to highest)
const (
	_ int = iota
	LOWEST
	IMP        // IMP
	EQV        // EQV
	XOR        // XOR
	OR         // OR
	AND        // AND
	NOT        // NOT (unary)
	COMPARISON // =, <>, <, >, <=, >=
	SUM        // +, -
	MOD_OP     // MOD
	INTDIV     // \ (integer division)
	PRODUCT    // *, /
	NEGATE     // - (unary negation)
	POWER      // ^
	CALL       // function calls
)

// precedences maps token types to their precedence levels
var precedences = map[lexer.TokenType]int{
	lexer.TOKEN_IMP:       IMP,
	lexer.TOKEN_EQV:       EQV,
	lexer.TOKEN_XOR:       XOR,
	lexer.TOKEN_OR:        OR,
	lexer.TOKEN_AND:       AND,
	lexer.TOKEN_EQ:        COMPARISON,
	lexer.TOKEN_NE:        COMPARISON,
	lexer.TOKEN_LT:        COMPARISON,
	lexer.TOKEN_GT:        COMPARISON,
	lexer.TOKEN_LE:        COMPARISON,
	lexer.TOKEN_GE:        COMPARISON,
	lexer.TOKEN_PLUS:      SUM,
	lexer.TOKEN_MINUS:     SUM,
	lexer.TOKEN_MOD:       MOD_OP,
	lexer.TOKEN_BACKSLASH: INTDIV,
	lexer.TOKEN_ASTERISK:  PRODUCT,
	lexer.TOKEN_SLASH:     PRODUCT,
	lexer.TOKEN_CARET:     POWER,
	lexer.TOKEN_LPAREN:    CALL,
}

// tokenToOperator maps token types to operator strings
var tokenToOperator = map[lexer.TokenType]string{
	lexer.TOKEN_PLUS:      "+",
	lexer.TOKEN_MINUS:     "-",
	lexer.TOKEN_ASTERISK:  "*",
	lexer.TOKEN_SLASH:     "/",
	lexer.TOKEN_BACKSLASH: "\\",
	lexer.TOKEN_CARET:     "^",
	lexer.TOKEN_MOD:       "MOD",
	lexer.TOKEN_EQ:        "=",
	lexer.TOKEN_NE:        "<>",
	lexer.TOKEN_LT:        "<",
	lexer.TOKEN_GT:        ">",
	lexer.TOKEN_LE:        "<=",
	lexer.TOKEN_GE:        ">=",
	lexer.TOKEN_AND:       "AND",
	lexer.TOKEN_OR:        "OR",
	lexer.TOKEN_XOR:       "XOR",
	lexer.TOKEN_EQV:       "EQV",
	lexer.TOKEN_IMP:       "IMP",
	lexer.TOKEN_NOT:       "NOT",
}
