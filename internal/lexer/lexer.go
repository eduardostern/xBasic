package lexer

import (
	"strings"
	"unicode"
)

// Lexer tokenizes QBasic source code
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line number
	column       int  // current column number
	lineStart    bool // true if at start of line (for line numbers)
}

// New creates a new Lexer
func New(input string) *Lexer {
	l := &Lexer{
		input:     input,
		line:      1,
		column:    0,
		lineStart: true,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) peekCharN(n int) byte {
	pos := l.readPosition + n - 1
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '=':
		tok = l.newToken(TOKEN_EQ, string(l.ch))
	case '+':
		tok = l.newToken(TOKEN_PLUS, string(l.ch))
	case '-':
		tok = l.newToken(TOKEN_MINUS, string(l.ch))
	case '*':
		tok = l.newToken(TOKEN_ASTERISK, string(l.ch))
	case '/':
		tok = l.newToken(TOKEN_SLASH, string(l.ch))
	case '\\':
		tok = l.newToken(TOKEN_BACKSLASH, string(l.ch))
	case '^':
		tok = l.newToken(TOKEN_CARET, string(l.ch))
	case '(':
		tok = l.newToken(TOKEN_LPAREN, string(l.ch))
	case ')':
		tok = l.newToken(TOKEN_RPAREN, string(l.ch))
	case ',':
		tok = l.newToken(TOKEN_COMMA, string(l.ch))
	case ';':
		tok = l.newToken(TOKEN_SEMICOLON, string(l.ch))
	case ':':
		tok = l.newToken(TOKEN_COLON, string(l.ch))
	case '#':
		tok = l.newToken(TOKEN_HASH, string(l.ch))
	case '$':
		tok = l.newToken(TOKEN_DOLLAR, string(l.ch))
	case '%':
		tok = l.newToken(TOKEN_PERCENT, string(l.ch))
	case '&':
		tok = l.newToken(TOKEN_AMPERSAND, string(l.ch))
	case '!':
		tok = l.newToken(TOKEN_BANG, string(l.ch))
	case '?':
		// ? is shorthand for PRINT
		tok = l.newToken(TOKEN_PRINT, "?")
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_LE, string(ch)+string(l.ch))
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_NE, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(TOKEN_LT, string(l.ch))
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_GE, string(ch)+string(l.ch))
		} else if l.peekChar() == '<' {
			ch := l.ch
			l.readChar()
			tok = l.newToken(TOKEN_NE, string(ch)+string(l.ch))
		} else {
			tok = l.newToken(TOKEN_GT, string(l.ch))
		}
	case '"':
		tok.Type = TOKEN_STRING
		tok.Literal = l.readString()
		tok.Line = l.line
		return tok
	case '\'':
		// Comment - skip to end of line
		tok.Type = TOKEN_REM
		tok.Literal = l.readComment()
		tok.Line = l.line
		return tok
	case '\n':
		tok = l.newToken(TOKEN_NEWLINE, "\\n")
		l.line++
		l.column = 0
		l.lineStart = true
		l.readChar()
		return tok
	case '\r':
		// Handle \r\n (Windows) or just \r (old Mac)
		l.readChar()
		if l.ch == '\n' {
			l.readChar()
		}
		tok = Token{Type: TOKEN_NEWLINE, Literal: "\\n", Line: l.line, Column: l.column}
		l.line++
		l.column = 0
		l.lineStart = true
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = TOKEN_EOF
		return tok
	default:
		if isDigit(l.ch) {
			return l.readNumber()
		} else if isLetter(l.ch) {
			return l.readIdentifier()
		} else {
			tok = l.newToken(TOKEN_ILLEGAL, string(l.ch))
		}
	}

	l.lineStart = false
	l.readChar()
	return tok
}

func (l *Lexer) newToken(tokenType TokenType, literal string) Token {
	return Token{Type: tokenType, Literal: literal, Line: l.line, Column: l.column}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' {
		l.readChar()
	}
}

func (l *Lexer) readString() string {
	var result strings.Builder
	l.readChar() // skip opening quote

	for l.ch != '"' && l.ch != 0 && l.ch != '\n' && l.ch != '\r' {
		result.WriteByte(l.ch)
		l.readChar()
	}

	if l.ch == '"' {
		l.readChar() // skip closing quote
	}

	return result.String()
}

func (l *Lexer) readComment() string {
	var result strings.Builder
	l.readChar() // skip the '

	for l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
		result.WriteByte(l.ch)
		l.readChar()
	}

	return result.String()
}

func (l *Lexer) readNumber() Token {
	startCol := l.column
	var result strings.Builder
	isFloat := false
	hasExponent := false

	// Read integer part
	for isDigit(l.ch) {
		result.WriteByte(l.ch)
		l.readChar()
	}

	// Check for decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		result.WriteByte(l.ch)
		l.readChar()

		// Read decimal part
		for isDigit(l.ch) {
			result.WriteByte(l.ch)
			l.readChar()
		}
	}

	// Check for exponent (E or D notation)
	if l.ch == 'E' || l.ch == 'e' || l.ch == 'D' || l.ch == 'd' {
		isFloat = true
		hasExponent = true
		result.WriteByte(l.ch)
		l.readChar()

		// Optional sign
		if l.ch == '+' || l.ch == '-' {
			result.WriteByte(l.ch)
			l.readChar()
		}

		// Exponent digits
		for isDigit(l.ch) {
			result.WriteByte(l.ch)
			l.readChar()
		}
	}

	// Check for type suffix on number
	if l.ch == '#' || l.ch == '!' || l.ch == '%' || l.ch == '&' {
		if l.ch == '#' || l.ch == '!' {
			isFloat = true
		}
		result.WriteByte(l.ch)
		l.readChar()
	}

	tok := Token{
		Literal: result.String(),
		Line:    l.line,
		Column:  startCol,
	}

	// Determine if this is a line number (integer at start of line)
	if l.lineStart && !isFloat && !hasExponent {
		tok.Type = TOKEN_LINE_NUMBER
	} else if isFloat {
		tok.Type = TOKEN_FLOAT
	} else {
		tok.Type = TOKEN_INTEGER
	}

	l.lineStart = false
	return tok
}

func (l *Lexer) readIdentifier() Token {
	startCol := l.column
	var result strings.Builder

	// Read identifier characters
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		result.WriteByte(l.ch)
		l.readChar()
	}

	// Check for type suffix
	typeSuffix := ""
	if l.ch == '$' || l.ch == '%' || l.ch == '&' || l.ch == '!' || l.ch == '#' {
		typeSuffix = string(l.ch)
		result.WriteByte(l.ch)
		l.readChar()
	}

	literal := result.String()
	upperLiteral := strings.ToUpper(literal)

	// Remove suffix for keyword lookup
	lookupName := upperLiteral
	if typeSuffix != "" {
		lookupName = strings.ToUpper(literal[:len(literal)-1])
	}

	tok := Token{
		Literal: literal,
		Line:    l.line,
		Column:  startCol,
	}

	// Check if it's a keyword (without type suffix)
	if typeSuffix == "" {
		if keywordType := LookupIdent(lookupName); keywordType != TOKEN_IDENT {
			tok.Type = keywordType

			// Handle REM specially - read rest of line as comment
			if keywordType == TOKEN_REM {
				l.skipWhitespace()
				comment := l.readComment()
				tok.Literal = literal + " " + comment
			}
		} else {
			tok.Type = TOKEN_IDENT
		}
	} else {
		tok.Type = TOKEN_IDENT
	}

	l.lineStart = false
	return tok
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// Tokenize returns all tokens from the input
func Tokenize(input string) []Token {
	l := New(input)
	var tokens []Token

	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}

	return tokens
}
