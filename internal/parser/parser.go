package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xbasic/xbasic/internal/ast"
	"github.com/xbasic/xbasic/internal/lexer"
)

// Parser parses QBasic source code into an AST
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// New creates a new Parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Register prefix parse functions
	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.TOKEN_IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.TOKEN_INTEGER, p.parseIntegerLiteral)
	p.registerPrefix(lexer.TOKEN_FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.TOKEN_STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TOKEN_MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.TOKEN_NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.TOKEN_LPAREN, p.parseGroupedExpression)

	// Register infix parse functions
	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.TOKEN_PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_ASTERISK, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_BACKSLASH, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_CARET, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_MOD, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_NE, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LT, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_GT, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LE, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_GE, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_AND, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_OR, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_XOR, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_EQV, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_IMP, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LPAREN, p.parseCallExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// Errors returns the parser errors
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("line %d: expected next token to be %s, got %s instead",
		p.peekToken.Line, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("line %d: no prefix parse function for %s found",
		p.curToken.Line, t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// ParseProgram parses the entire program
func (p *Parser) ParseProgram() *ast.Program {
	program := ast.NewProgram()

	for !p.curTokenIs(lexer.TOKEN_EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			// Track line numbers and labels
			switch s := stmt.(type) {
			case *ast.LineNumberStmt:
				program.LineNumbers[s.Number] = len(program.Statements)
			case *ast.LabelStmt:
				program.Labels[strings.ToUpper(s.Name)] = len(program.Statements)
			case *ast.SubStatement:
				program.Subs[strings.ToUpper(s.Name)] = s
			case *ast.FuncStatement:
				program.Functions[strings.ToUpper(s.Name)] = s
			case *ast.DataStmt:
				program.DataItems = append(program.DataItems, s.Values...)
			}
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	// Skip empty lines
	for p.curTokenIs(lexer.TOKEN_NEWLINE) {
		p.nextToken()
	}

	if p.curTokenIs(lexer.TOKEN_EOF) {
		return nil
	}

	// Handle line numbers
	if p.curTokenIs(lexer.TOKEN_LINE_NUMBER) {
		lineNum, _ := strconv.Atoi(p.curToken.Literal)
		stmt := &ast.LineNumberStmt{Line: p.curToken.Line, Number: lineNum}
		if p.peekTokenIs(lexer.TOKEN_NEWLINE) || p.peekTokenIs(lexer.TOKEN_EOF) {
			return stmt
		}
		p.nextToken()
		// Continue parsing the statement after the line number
		// For now, return just the line number marker
		nextStmt := p.parseStatement()
		if nextStmt != nil {
			// We'll handle this in the interpreter
			return stmt
		}
		return stmt
	}

	switch p.curToken.Type {
	case lexer.TOKEN_LET:
		return p.parseLetStatement()
	case lexer.TOKEN_PRINT:
		return p.parsePrintStatement()
	case lexer.TOKEN_INPUT:
		return p.parseInputStatement()
	case lexer.TOKEN_DIM:
		return p.parseDimStatement()
	case lexer.TOKEN_IF:
		return p.parseIfStatement()
	case lexer.TOKEN_FOR:
		return p.parseForStatement()
	case lexer.TOKEN_WHILE:
		return p.parseWhileStatement()
	case lexer.TOKEN_DO:
		return p.parseDoLoopStatement()
	case lexer.TOKEN_SELECT:
		return p.parseSelectCaseStatement()
	case lexer.TOKEN_GOTO:
		return p.parseGotoStatement()
	case lexer.TOKEN_GOSUB:
		return p.parseGosubStatement()
	case lexer.TOKEN_RETURN:
		return p.parseReturnStatement()
	case lexer.TOKEN_EXIT:
		return p.parseExitStatement()
	case lexer.TOKEN_SUB:
		return p.parseSubStatement()
	case lexer.TOKEN_FUNCTION:
		return p.parseFunctionStatement()
	case lexer.TOKEN_DATA:
		return p.parseDataStatement()
	case lexer.TOKEN_READ:
		return p.parseReadStatement()
	case lexer.TOKEN_RESTORE:
		return p.parseRestoreStatement()
	case lexer.TOKEN_CLS:
		return p.parseClsStatement()
	case lexer.TOKEN_LOCATE:
		return p.parseLocateStatement()
	case lexer.TOKEN_COLOR:
		return p.parseColorStatement()
	case lexer.TOKEN_SCREEN:
		return p.parseScreenStatement()
	case lexer.TOKEN_END:
		return p.parseEndStatement()
	case lexer.TOKEN_REM:
		return p.parseRemStatement()
	case lexer.TOKEN_CALL:
		return p.parseCallStatement()
	case lexer.TOKEN_SLEEP:
		return p.parseSleepStatement()
	case lexer.TOKEN_BEEP:
		return p.parseBeepStatement()
	case lexer.TOKEN_SWAP:
		return p.parseSwapStatement()
	case lexer.TOKEN_RANDOMIZE:
		return p.parseRandomizeStatement()
	case lexer.TOKEN_CONST:
		return p.parseConstStatement()
	case lexer.TOKEN_OPEN:
		return p.parseOpenStatement()
	case lexer.TOKEN_CLOSE:
		return p.parseCloseStatement()
	case lexer.TOKEN_LINE:
		return p.parseLineStatement()
	case lexer.TOKEN_ON:
		return p.parseOnStatement()
	case lexer.TOKEN_REDIM:
		return p.parseRedimStatement()
	case lexer.TOKEN_GET:
		return p.parseGetStatement()
	case lexer.TOKEN_PUT:
		return p.parsePutStatement()
	case lexer.TOKEN_SEEK:
		return p.parseSeekStatement()
	case lexer.TOKEN_PSET:
		return p.parsePsetStatement()
	case lexer.TOKEN_CIRCLE:
		return p.parseCircleStatement()
	case lexer.TOKEN_IDENT:
		return p.parseIdentifierStatement()
	default:
		return nil
	}
}

// parseIdentifierStatement handles assignment or sub calls starting with identifier
func (p *Parser) parseIdentifierStatement() ast.Statement {
	line := p.curToken.Line
	name := p.curToken.Literal

	// Check if this is an array access or assignment
	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		// Could be array assignment or function call as statement
		ident := p.parseIdentifier()

		if p.peekTokenIs(lexer.TOKEN_EQ) {
			// Array assignment
			p.nextToken() // move to =
			p.nextToken() // move past =
			value := p.parseExpression(LOWEST)
			return &ast.LetStmt{Line: line, Name: ident, Value: value}
		}
		// It's a sub call with parentheses
		if call, ok := ident.(*ast.CallExpr); ok {
			return &ast.SubCallStmt{Line: line, Name: call.Function, Arguments: call.Arguments}
		}
	}

	if p.peekTokenIs(lexer.TOKEN_EQ) {
		// Simple assignment
		p.nextToken() // move to =
		p.nextToken() // move past =
		value := p.parseExpression(LOWEST)
		return &ast.LetStmt{
			Line:  line,
			Name:  &ast.Identifier{Line: line, Name: name},
			Value: value,
		}
	}

	// Check if it's a label (identifier followed by colon)
	if p.peekTokenIs(lexer.TOKEN_COLON) {
		p.nextToken() // consume colon
		return &ast.LabelStmt{Line: line, Name: name}
	}

	// It's a sub call without parentheses
	p.nextToken()
	var args []ast.Expression
	for !p.curTokenIs(lexer.TOKEN_NEWLINE) && !p.curTokenIs(lexer.TOKEN_EOF) && !p.curTokenIs(lexer.TOKEN_COLON) {
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			args = append(args, expr)
		}
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			p.nextToken()
		} else {
			break
		}
	}
	return &ast.SubCallStmt{Line: line, Name: name, Arguments: args}
}

func (p *Parser) parseLetStatement() ast.Statement {
	line := p.curToken.Line
	p.nextToken() // skip LET

	name := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_EQ) {
		return nil
	}

	p.nextToken()
	value := p.parseExpression(LOWEST)

	return &ast.LetStmt{Line: line, Name: name, Value: value}
}

func (p *Parser) parsePrintStatement() ast.Statement {
	line := p.curToken.Line
	var fileNum ast.Expression

	// Check for file output: PRINT #n, ...
	if p.peekTokenIs(lexer.TOKEN_HASH) {
		p.nextToken() // move to #
		p.nextToken() // move past #
		fileNum = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TOKEN_COMMA) {
			return nil
		}
		p.nextToken()

		// Check for PRINT #n, USING
		if p.curTokenIs(lexer.TOKEN_USING) {
			return p.parsePrintUsingRest(line, fileNum)
		}

		// Parse as PRINT #n statement
		fileStmt := &ast.PrintFileStmt{Line: line, FileNum: fileNum}
		fileStmt.Items, fileStmt.NoNewline = p.parsePrintItems()
		return fileStmt
	}

	p.nextToken()

	// Check for PRINT USING
	if p.curTokenIs(lexer.TOKEN_USING) {
		return p.parsePrintUsingRest(line, nil)
	}

	stmt := &ast.PrintStmt{Line: line}
	stmt.Items, stmt.NoNewline = p.parsePrintItems()
	return stmt
}

func (p *Parser) parsePrintUsingRest(line int, fileNum ast.Expression) ast.Statement {
	stmt := &ast.PrintUsingStmt{Line: line, FileNum: fileNum}

	p.nextToken() // skip USING
	stmt.Format = p.parseExpression(LOWEST)

	// Expect semicolon after format string
	if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
	}
	p.nextToken()

	// Parse items
	for !p.curTokenIs(lexer.TOKEN_NEWLINE) && !p.curTokenIs(lexer.TOKEN_EOF) && !p.curTokenIs(lexer.TOKEN_COLON) {
		if p.curTokenIs(lexer.TOKEN_SEMICOLON) {
			if len(stmt.Items) > 0 {
				stmt.Items[len(stmt.Items)-1].Separator = ";"
			}
			stmt.NoNewline = true
			p.nextToken()
			continue
		}
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			if len(stmt.Items) > 0 {
				stmt.Items[len(stmt.Items)-1].Separator = ","
			}
			stmt.NoNewline = true
			p.nextToken()
			continue
		}

		expr := p.parseExpression(LOWEST)
		if expr != nil {
			stmt.Items = append(stmt.Items, ast.PrintItem{Expression: expr})
			stmt.NoNewline = false
		}
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parsePrintItems() ([]ast.PrintItem, bool) {
	var items []ast.PrintItem
	noNewline := false

	for !p.curTokenIs(lexer.TOKEN_NEWLINE) && !p.curTokenIs(lexer.TOKEN_EOF) && !p.curTokenIs(lexer.TOKEN_COLON) {
		if p.curTokenIs(lexer.TOKEN_SEMICOLON) {
			if len(items) > 0 {
				items[len(items)-1].Separator = ";"
			}
			noNewline = true
			p.nextToken()
			continue
		}
		if p.curTokenIs(lexer.TOKEN_COMMA) {
			if len(items) > 0 {
				items[len(items)-1].Separator = ","
			}
			noNewline = true
			p.nextToken()
			continue
		}

		expr := p.parseExpression(LOWEST)
		if expr != nil {
			items = append(items, ast.PrintItem{Expression: expr})
			noNewline = false
		}
		p.nextToken()
	}

	return items, noNewline
}

func (p *Parser) parseInputStatement() ast.Statement {
	stmt := &ast.InputStmt{Line: p.curToken.Line}

	// Check for file input: INPUT #n, ...
	if p.peekTokenIs(lexer.TOKEN_HASH) {
		p.nextToken() // move to #
		p.nextToken() // move past #
		fileNum := p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TOKEN_COMMA) {
			return nil
		}
		p.nextToken()

		fileStmt := &ast.InputFileStmt{Line: stmt.Line, FileNum: fileNum}
		for !p.curTokenIs(lexer.TOKEN_NEWLINE) && !p.curTokenIs(lexer.TOKEN_EOF) {
			expr := p.parseExpression(LOWEST)
			if expr != nil {
				fileStmt.Variables = append(fileStmt.Variables, expr)
			}
			if p.peekTokenIs(lexer.TOKEN_COMMA) {
				p.nextToken()
				p.nextToken()
			} else {
				break
			}
		}
		return fileStmt
	}

	p.nextToken()

	// Check for prompt string
	if p.curTokenIs(lexer.TOKEN_STRING) {
		stmt.Prompt = &ast.StringLiteral{Line: p.curToken.Line, Value: p.curToken.Literal}
		p.nextToken()
		if p.curTokenIs(lexer.TOKEN_SEMICOLON) || p.curTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
	}

	// Parse variable list
	for !p.curTokenIs(lexer.TOKEN_NEWLINE) && !p.curTokenIs(lexer.TOKEN_EOF) {
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			stmt.Variables = append(stmt.Variables, expr)
		}
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			p.nextToken()
		} else {
			break
		}
	}

	return stmt
}

func (p *Parser) parseDimStatement() ast.Statement {
	stmt := &ast.DimStmt{Line: p.curToken.Line}

	p.nextToken()

	// Check for SHARED or STATIC
	if p.curTokenIs(lexer.TOKEN_SHARED) {
		stmt.Shared = true
		p.nextToken()
	}
	if p.curTokenIs(lexer.TOKEN_STATIC) {
		stmt.Static = true
		p.nextToken()
	}

	// Parse variable declarations
	for {
		dimVar := ast.DimVariable{Name: p.curToken.Literal}

		// Check for array dimensions
		if p.peekTokenIs(lexer.TOKEN_LPAREN) {
			p.nextToken() // move to (
			p.nextToken() // move past (
			for !p.curTokenIs(lexer.TOKEN_RPAREN) {
				dim := p.parseExpression(LOWEST)
				dimVar.Dimensions = append(dimVar.Dimensions, dim)
				if p.peekTokenIs(lexer.TOKEN_COMMA) {
					p.nextToken()
				}
				p.nextToken()
			}
		}

		// Check for AS type
		if p.peekTokenIs(lexer.TOKEN_AS) {
			p.nextToken() // move to AS
			p.nextToken() // move to type
			dimVar.DataType = p.parseDataType()
		}

		stmt.Variables = append(stmt.Variables, dimVar)

		if !p.peekTokenIs(lexer.TOKEN_COMMA) {
			break
		}
		p.nextToken() // move to comma
		p.nextToken() // move past comma
	}

	return stmt
}

func (p *Parser) parseDataType() ast.DataType {
	switch p.curToken.Type {
	case lexer.TOKEN_INTEGER_TYPE:
		return ast.TypeInteger
	case lexer.TOKEN_LONG_TYPE:
		return ast.TypeLong
	case lexer.TOKEN_SINGLE_TYPE:
		return ast.TypeSingle
	case lexer.TOKEN_DOUBLE_TYPE:
		return ast.TypeDouble
	case lexer.TOKEN_STRING_TYPE:
		return ast.TypeString
	default:
		return ast.TypeUnknown
	}
}

func (p *Parser) parseIfStatement() ast.Statement {
	stmt := &ast.IfStmt{Line: p.curToken.Line}

	p.nextToken() // skip IF
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_THEN) {
		return nil
	}

	// Check for single-line IF
	if !p.peekTokenIs(lexer.TOKEN_NEWLINE) && !p.peekTokenIs(lexer.TOKEN_EOF) {
		stmt.SingleLine = true
		p.nextToken()

		// Parse consequence
		conseq := p.parseStatement()
		if conseq != nil {
			stmt.Consequence = append(stmt.Consequence, conseq)
		}

		// Check for ELSE
		if p.curTokenIs(lexer.TOKEN_ELSE) || p.peekTokenIs(lexer.TOKEN_ELSE) {
			if p.peekTokenIs(lexer.TOKEN_ELSE) {
				p.nextToken()
			}
			p.nextToken()
			alt := p.parseStatement()
			if alt != nil {
				stmt.Alternative = append(stmt.Alternative, alt)
			}
		}

		return stmt
	}

	// Multi-line IF
	p.nextToken() // move past THEN
	p.nextToken() // move past newline

	// Parse consequence block
	for !p.curTokenIs(lexer.TOKEN_ELSE) && !p.curTokenIs(lexer.TOKEN_ELSEIF) &&
		!p.curTokenIs(lexer.TOKEN_END) && !p.curTokenIs(lexer.TOKEN_EOF) {
		s := p.parseStatement()
		if s != nil {
			stmt.Consequence = append(stmt.Consequence, s)
		}
		p.nextToken()
	}

	// Handle ELSEIF as nested IF in ELSE
	if p.curTokenIs(lexer.TOKEN_ELSEIF) {
		elseifStmt := p.parseIfStatement()
		if elseifStmt != nil {
			stmt.Alternative = append(stmt.Alternative, elseifStmt)
		}
		return stmt
	}

	// Parse ELSE block
	if p.curTokenIs(lexer.TOKEN_ELSE) {
		p.nextToken()
		if p.curTokenIs(lexer.TOKEN_NEWLINE) {
			p.nextToken()
		}

		for !p.curTokenIs(lexer.TOKEN_END) && !p.curTokenIs(lexer.TOKEN_EOF) {
			s := p.parseStatement()
			if s != nil {
				stmt.Alternative = append(stmt.Alternative, s)
			}
			p.nextToken()
		}
	}

	// Expect END IF
	if p.curTokenIs(lexer.TOKEN_END) {
		p.nextToken() // move past END
		// IF is optional after END
	}

	return stmt
}

func (p *Parser) parseForStatement() ast.Statement {
	stmt := &ast.ForStmt{Line: p.curToken.Line}

	p.nextToken() // skip FOR
	stmt.Variable = &ast.Identifier{Line: p.curToken.Line, Name: p.curToken.Literal}

	if !p.expectPeek(lexer.TOKEN_EQ) {
		return nil
	}

	p.nextToken()
	stmt.Start = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_TO) {
		return nil
	}

	p.nextToken()
	stmt.End = p.parseExpression(LOWEST)

	// Optional STEP
	if p.peekTokenIs(lexer.TOKEN_STEP) {
		p.nextToken()
		p.nextToken()
		stmt.Step = p.parseExpression(LOWEST)
	}

	// Skip to body
	for p.curTokenIs(lexer.TOKEN_NEWLINE) || p.peekTokenIs(lexer.TOKEN_NEWLINE) {
		p.nextToken()
	}

	// Parse body until NEXT
	for !p.curTokenIs(lexer.TOKEN_NEXT) && !p.curTokenIs(lexer.TOKEN_EOF) {
		// Skip empty lines
		if p.curTokenIs(lexer.TOKEN_NEWLINE) {
			p.nextToken()
			continue
		}
		s := p.parseStatement()
		if s != nil {
			stmt.Body = append(stmt.Body, s)
		}
		// Move to next token if we're not already at NEXT
		if !p.peekTokenIs(lexer.TOKEN_NEXT) {
			p.nextToken()
		} else {
			p.nextToken()
			break
		}
	}

	// Skip optional variable name after NEXT (e.g., NEXT i)
	if p.peekTokenIs(lexer.TOKEN_IDENT) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseWhileStatement() ast.Statement {
	stmt := &ast.WhileStmt{Line: p.curToken.Line}

	p.nextToken() // skip WHILE
	stmt.Condition = p.parseExpression(LOWEST)

	// Skip to body
	for p.curTokenIs(lexer.TOKEN_NEWLINE) || p.peekTokenIs(lexer.TOKEN_NEWLINE) {
		p.nextToken()
	}

	// Parse body until WEND
	for !p.curTokenIs(lexer.TOKEN_WEND) && !p.curTokenIs(lexer.TOKEN_EOF) {
		// Skip empty lines
		if p.curTokenIs(lexer.TOKEN_NEWLINE) {
			p.nextToken()
			continue
		}
		s := p.parseStatement()
		if s != nil {
			stmt.Body = append(stmt.Body, s)
		}
		// Move to next token if we're not already at WEND
		if !p.peekTokenIs(lexer.TOKEN_WEND) {
			p.nextToken()
		} else {
			p.nextToken()
			break
		}
	}

	return stmt
}

func (p *Parser) parseDoLoopStatement() ast.Statement {
	stmt := &ast.DoLoopStmt{Line: p.curToken.Line}

	p.nextToken() // skip DO

	// Check for pre-condition: DO WHILE/UNTIL condition
	if p.curTokenIs(lexer.TOKEN_WHILE) {
		stmt.ConditionType = "WHILE"
		stmt.ConditionPos = "PRE"
		p.nextToken()
		stmt.Condition = p.parseExpression(LOWEST)
	} else if p.curTokenIs(lexer.TOKEN_UNTIL) {
		stmt.ConditionType = "UNTIL"
		stmt.ConditionPos = "PRE"
		p.nextToken()
		stmt.Condition = p.parseExpression(LOWEST)
	}

	// Skip to body
	for p.curTokenIs(lexer.TOKEN_NEWLINE) {
		p.nextToken()
	}

	// Parse body until LOOP
	for !p.curTokenIs(lexer.TOKEN_LOOP) && !p.curTokenIs(lexer.TOKEN_EOF) {
		// Skip empty lines
		if p.curTokenIs(lexer.TOKEN_NEWLINE) {
			p.nextToken()
			continue
		}
		s := p.parseStatement()
		if s != nil {
			stmt.Body = append(stmt.Body, s)
		}
		// Move to next token if we're not already at LOOP
		if !p.peekTokenIs(lexer.TOKEN_LOOP) {
			p.nextToken()
		} else {
			p.nextToken()
			break
		}
	}

	// Check for post-condition: LOOP WHILE/UNTIL condition
	if p.peekTokenIs(lexer.TOKEN_WHILE) {
		p.nextToken()
		stmt.ConditionType = "WHILE"
		stmt.ConditionPos = "POST"
		p.nextToken()
		stmt.Condition = p.parseExpression(LOWEST)
	} else if p.peekTokenIs(lexer.TOKEN_UNTIL) {
		p.nextToken()
		stmt.ConditionType = "UNTIL"
		stmt.ConditionPos = "POST"
		p.nextToken()
		stmt.Condition = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseSelectCaseStatement() ast.Statement {
	stmt := &ast.SelectCaseStmt{Line: p.curToken.Line}

	if !p.expectPeek(lexer.TOKEN_CASE) {
		return nil
	}

	p.nextToken()
	stmt.Expression = p.parseExpression(LOWEST)

	// Move past newline
	for p.peekTokenIs(lexer.TOKEN_NEWLINE) {
		p.nextToken()
	}
	p.nextToken()

	// Parse CASE clauses
	for p.curTokenIs(lexer.TOKEN_CASE) {
		p.nextToken()

		// Check for CASE ELSE
		if p.curTokenIs(lexer.TOKEN_ELSE) {
			p.nextToken()
			if p.curTokenIs(lexer.TOKEN_NEWLINE) {
				p.nextToken()
			}
			// Parse CASE ELSE body
			for !p.curTokenIs(lexer.TOKEN_END) && !p.curTokenIs(lexer.TOKEN_CASE) && !p.curTokenIs(lexer.TOKEN_EOF) {
				s := p.parseStatement()
				if s != nil {
					stmt.CaseElse = append(stmt.CaseElse, s)
				}
				p.nextToken()
			}
			continue
		}

		// Parse CASE values
		clause := ast.CaseClause{}
		for {
			caseVal := p.parseCaseValue()
			clause.Values = append(clause.Values, caseVal)

			if !p.peekTokenIs(lexer.TOKEN_COMMA) {
				break
			}
			p.nextToken()
			p.nextToken()
		}

		// Move past newline
		for p.peekTokenIs(lexer.TOKEN_NEWLINE) {
			p.nextToken()
		}
		p.nextToken()

		// Parse CASE body
		for !p.curTokenIs(lexer.TOKEN_CASE) && !p.curTokenIs(lexer.TOKEN_END) && !p.curTokenIs(lexer.TOKEN_EOF) {
			s := p.parseStatement()
			if s != nil {
				clause.Body = append(clause.Body, s)
			}
			p.nextToken()
		}

		stmt.Cases = append(stmt.Cases, clause)
	}

	// Expect END SELECT
	if p.curTokenIs(lexer.TOKEN_END) {
		p.nextToken() // skip END
		// SELECT is optional
	}

	return stmt
}

func (p *Parser) parseCaseValue() ast.CaseValue {
	// Check for IS operator
	if p.curToken.Literal == "IS" || p.curTokenIs(lexer.TOKEN_LT) || p.curTokenIs(lexer.TOKEN_GT) ||
		p.curTokenIs(lexer.TOKEN_LE) || p.curTokenIs(lexer.TOKEN_GE) || p.curTokenIs(lexer.TOKEN_EQ) ||
		p.curTokenIs(lexer.TOKEN_NE) {

		var op string
		if p.curToken.Literal == "IS" {
			p.nextToken()
			op = tokenToOperator[p.curToken.Type]
			p.nextToken()
		} else {
			op = tokenToOperator[p.curToken.Type]
			p.nextToken()
		}

		value := p.parseExpression(LOWEST)
		return ast.CaseValue{Type: "IS", Operator: op, Value: value}
	}

	// Parse first value
	value := p.parseExpression(LOWEST)

	// Check for TO (range)
	if p.peekTokenIs(lexer.TOKEN_TO) {
		p.nextToken()
		p.nextToken()
		endValue := p.parseExpression(LOWEST)
		return ast.CaseValue{Type: "RANGE", Value: value, EndValue: endValue}
	}

	return ast.CaseValue{Type: "SINGLE", Value: value}
}

func (p *Parser) parseGotoStatement() ast.Statement {
	stmt := &ast.GotoStmt{Line: p.curToken.Line}
	p.nextToken()
	stmt.Target = p.curToken.Literal
	return stmt
}

func (p *Parser) parseGosubStatement() ast.Statement {
	stmt := &ast.GosubStmt{Line: p.curToken.Line}
	p.nextToken()
	stmt.Target = p.curToken.Literal
	return stmt
}

func (p *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStmt{Line: p.curToken.Line}

	if !p.peekTokenIs(lexer.TOKEN_NEWLINE) && !p.peekTokenIs(lexer.TOKEN_EOF) && !p.peekTokenIs(lexer.TOKEN_COLON) {
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseExitStatement() ast.Statement {
	stmt := &ast.ExitStmt{Line: p.curToken.Line}
	p.nextToken()
	stmt.ExitType = strings.ToUpper(p.curToken.Literal)
	return stmt
}

func (p *Parser) parseSubStatement() ast.Statement {
	stmt := &ast.SubStatement{Line: p.curToken.Line}

	p.nextToken() // skip SUB
	stmt.Name = p.curToken.Literal

	// Parse parameters
	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		stmt.Parameters = p.parseParameters()
	}

	// Check for STATIC
	if p.peekTokenIs(lexer.TOKEN_STATIC) {
		p.nextToken()
		stmt.Static = true
	}

	// Move past newline
	for p.peekTokenIs(lexer.TOKEN_NEWLINE) {
		p.nextToken()
	}
	p.nextToken()

	// Parse body until END SUB
	for !p.isEndSub() && !p.curTokenIs(lexer.TOKEN_EOF) {
		s := p.parseStatement()
		if s != nil {
			stmt.Body = append(stmt.Body, s)
		}
		p.nextToken()
	}

	// Skip END SUB
	if p.curTokenIs(lexer.TOKEN_END) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) isEndSub() bool {
	return p.curTokenIs(lexer.TOKEN_END) && p.peekTokenIs(lexer.TOKEN_SUB)
}

func (p *Parser) isEndFunction() bool {
	return p.curTokenIs(lexer.TOKEN_END) && p.peekTokenIs(lexer.TOKEN_FUNCTION)
}

func (p *Parser) parseFunctionStatement() ast.Statement {
	stmt := &ast.FuncStatement{Line: p.curToken.Line}

	p.nextToken() // skip FUNCTION
	stmt.Name = p.curToken.Literal

	// Check for type suffix in name
	if len(stmt.Name) > 0 {
		lastChar := stmt.Name[len(stmt.Name)-1:]
		if dt := ast.DataTypeFromSuffix(lastChar); dt != ast.TypeUnknown {
			stmt.ReturnType = dt
		}
	}

	// Parse parameters
	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		stmt.Parameters = p.parseParameters()
	}

	// Check for AS type
	if p.peekTokenIs(lexer.TOKEN_AS) {
		p.nextToken()
		p.nextToken()
		stmt.ReturnType = p.parseDataType()
	}

	// Check for STATIC
	if p.peekTokenIs(lexer.TOKEN_STATIC) {
		p.nextToken()
		stmt.Static = true
	}

	// Move past newline
	for p.peekTokenIs(lexer.TOKEN_NEWLINE) {
		p.nextToken()
	}
	p.nextToken()

	// Parse body until END FUNCTION
	for !p.isEndFunction() && !p.curTokenIs(lexer.TOKEN_EOF) {
		s := p.parseStatement()
		if s != nil {
			stmt.Body = append(stmt.Body, s)
		}
		p.nextToken()
	}

	// Skip END FUNCTION
	if p.curTokenIs(lexer.TOKEN_END) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseParameters() []ast.Parameter {
	var params []ast.Parameter

	p.nextToken() // skip (

	for !p.curTokenIs(lexer.TOKEN_RPAREN) && !p.curTokenIs(lexer.TOKEN_EOF) {
		param := ast.Parameter{}

		// Check for BYVAL/BYREF
		if p.curTokenIs(lexer.TOKEN_BYVAL) {
			param.ByVal = true
			p.nextToken()
		} else if p.curTokenIs(lexer.TOKEN_BYREF) {
			p.nextToken()
		}

		param.Name = p.curToken.Literal

		// Check for type suffix or AS type
		if len(param.Name) > 0 {
			lastChar := param.Name[len(param.Name)-1:]
			if dt := ast.DataTypeFromSuffix(lastChar); dt != ast.TypeUnknown {
				param.DataType = dt
			}
		}

		if p.peekTokenIs(lexer.TOKEN_AS) {
			p.nextToken()
			p.nextToken()
			param.DataType = p.parseDataType()
		}

		params = append(params, param)

		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}

	return params
}

func (p *Parser) parseDataStatement() ast.Statement {
	stmt := &ast.DataStmt{Line: p.curToken.Line}

	p.nextToken() // skip DATA

	for !p.curTokenIs(lexer.TOKEN_NEWLINE) && !p.curTokenIs(lexer.TOKEN_EOF) {
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			stmt.Values = append(stmt.Values, expr)
		}
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReadStatement() ast.Statement {
	stmt := &ast.ReadStmt{Line: p.curToken.Line}

	p.nextToken() // skip READ

	for !p.curTokenIs(lexer.TOKEN_NEWLINE) && !p.curTokenIs(lexer.TOKEN_EOF) {
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			stmt.Variables = append(stmt.Variables, expr)
		}
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			p.nextToken()
		} else {
			break
		}
	}

	return stmt
}

func (p *Parser) parseRestoreStatement() ast.Statement {
	stmt := &ast.RestoreStmt{Line: p.curToken.Line}

	if !p.peekTokenIs(lexer.TOKEN_NEWLINE) && !p.peekTokenIs(lexer.TOKEN_EOF) {
		p.nextToken()
		stmt.Target = p.curToken.Literal
	}

	return stmt
}

func (p *Parser) parseClsStatement() ast.Statement {
	return &ast.ClsStmt{Line: p.curToken.Line}
}

func (p *Parser) parseLocateStatement() ast.Statement {
	stmt := &ast.LocateStmt{Line: p.curToken.Line}

	p.nextToken() // skip LOCATE

	if !p.curTokenIs(lexer.TOKEN_COMMA) {
		stmt.Row = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		stmt.Column = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseColorStatement() ast.Statement {
	stmt := &ast.ColorStmt{Line: p.curToken.Line}

	p.nextToken() // skip COLOR
	stmt.Foreground = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		stmt.Background = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseScreenStatement() ast.Statement {
	stmt := &ast.ScreenStmt{Line: p.curToken.Line}
	p.nextToken()
	stmt.Mode = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseEndStatement() ast.Statement {
	// Check what kind of END this is
	if p.peekTokenIs(lexer.TOKEN_IF) || p.peekTokenIs(lexer.TOKEN_SUB) ||
		p.peekTokenIs(lexer.TOKEN_FUNCTION) || p.peekTokenIs(lexer.TOKEN_SELECT) {
		return nil // handled by other parsers
	}
	return &ast.EndStmt{Line: p.curToken.Line}
}

func (p *Parser) parseRemStatement() ast.Statement {
	return &ast.RemStmt{Line: p.curToken.Line, Comment: p.curToken.Literal}
}

func (p *Parser) parseCallStatement() ast.Statement {
	stmt := &ast.CallStmt{Line: p.curToken.Line}

	p.nextToken() // skip CALL
	stmt.Name = p.curToken.Literal

	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken()
		p.nextToken()
		for !p.curTokenIs(lexer.TOKEN_RPAREN) && !p.curTokenIs(lexer.TOKEN_EOF) {
			expr := p.parseExpression(LOWEST)
			if expr != nil {
				stmt.Arguments = append(stmt.Arguments, expr)
			}
			if p.peekTokenIs(lexer.TOKEN_COMMA) {
				p.nextToken()
			}
			p.nextToken()
		}
	}

	return stmt
}

func (p *Parser) parseSleepStatement() ast.Statement {
	stmt := &ast.SleepStmt{Line: p.curToken.Line}

	if !p.peekTokenIs(lexer.TOKEN_NEWLINE) && !p.peekTokenIs(lexer.TOKEN_EOF) {
		p.nextToken()
		stmt.Seconds = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseBeepStatement() ast.Statement {
	return &ast.BeepStmt{Line: p.curToken.Line}
}

func (p *Parser) parseSwapStatement() ast.Statement {
	stmt := &ast.SwapStmt{Line: p.curToken.Line}

	p.nextToken()
	stmt.Var1 = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_COMMA) {
		return nil
	}

	p.nextToken()
	stmt.Var2 = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseRandomizeStatement() ast.Statement {
	stmt := &ast.RandomizeStmt{Line: p.curToken.Line}

	if !p.peekTokenIs(lexer.TOKEN_NEWLINE) && !p.peekTokenIs(lexer.TOKEN_EOF) {
		p.nextToken()
		// RANDOMIZE TIMER is common - check for it
		if p.curToken.Literal != "TIMER" {
			stmt.Seed = p.parseExpression(LOWEST)
		}
	}

	return stmt
}

func (p *Parser) parseConstStatement() ast.Statement {
	stmt := &ast.ConstStmt{Line: p.curToken.Line}

	p.nextToken()
	stmt.Name = p.curToken.Literal

	if !p.expectPeek(lexer.TOKEN_EQ) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseOpenStatement() ast.Statement {
	stmt := &ast.OpenStmt{Line: p.curToken.Line}

	p.nextToken()
	stmt.Filename = p.parseExpression(LOWEST)

	// FOR mode
	if p.peekTokenIs(lexer.TOKEN_FOR) {
		p.nextToken()
		p.nextToken()
		stmt.Mode = strings.ToUpper(p.curToken.Literal)
	}

	// AS #n
	for !p.curTokenIs(lexer.TOKEN_AS) && !p.curTokenIs(lexer.TOKEN_EOF) && !p.curTokenIs(lexer.TOKEN_NEWLINE) {
		p.nextToken()
	}
	if p.curTokenIs(lexer.TOKEN_AS) {
		p.nextToken()
		if p.curTokenIs(lexer.TOKEN_HASH) {
			p.nextToken()
		}
		stmt.FileNum = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseCloseStatement() ast.Statement {
	stmt := &ast.CloseStmt{Line: p.curToken.Line}

	for p.peekTokenIs(lexer.TOKEN_HASH) {
		p.nextToken()
		p.nextToken()
		stmt.FileNums = append(stmt.FileNums, p.parseExpression(LOWEST))
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
	}

	return stmt
}

func (p *Parser) parseLineStatement() ast.Statement {
	line := p.curToken.Line

	p.nextToken() // skip LINE

	// LINE INPUT
	if p.curTokenIs(lexer.TOKEN_INPUT) {
		// Check for file: LINE INPUT #n
		if p.peekTokenIs(lexer.TOKEN_HASH) {
			p.nextToken()
			p.nextToken()
			fileNum := p.parseExpression(LOWEST)
			if !p.expectPeek(lexer.TOKEN_COMMA) {
				return nil
			}
			p.nextToken()
			variable := p.parseExpression(LOWEST)
			return &ast.LineInputFileStmt{Line: line, FileNum: fileNum, Variable: variable}
		}

		stmt := &ast.LineInputStmt{Line: line}
		p.nextToken()

		// Check for prompt
		if p.curTokenIs(lexer.TOKEN_STRING) {
			stmt.Prompt = &ast.StringLiteral{Line: p.curToken.Line, Value: p.curToken.Literal}
			p.nextToken()
			if p.curTokenIs(lexer.TOKEN_SEMICOLON) || p.curTokenIs(lexer.TOKEN_COMMA) {
				p.nextToken()
			}
		}

		stmt.Variable = p.parseExpression(LOWEST)
		return stmt
	}

	// LINE (x1, y1)-(x2, y2), color, BF - graphics
	if p.curTokenIs(lexer.TOKEN_LPAREN) {
		stmt := &ast.LineGraphicsStmt{Line: line}

		p.nextToken() // skip (
		stmt.X1 = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TOKEN_COMMA) {
			return nil
		}
		p.nextToken()
		stmt.Y1 = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}

		// Expect -
		if !p.expectPeek(lexer.TOKEN_MINUS) {
			return nil
		}

		// Expect (
		if !p.expectPeek(lexer.TOKEN_LPAREN) {
			return nil
		}
		p.nextToken()
		stmt.X2 = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TOKEN_COMMA) {
			return nil
		}
		p.nextToken()
		stmt.Y2 = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}

		// Optional color
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			p.nextToken()
			if !p.curTokenIs(lexer.TOKEN_COMMA) && !p.curTokenIs(lexer.TOKEN_NEWLINE) && !p.curTokenIs(lexer.TOKEN_EOF) {
				stmt.Color = p.parseExpression(LOWEST)
			}
		}

		// Optional B or BF
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
			p.nextToken()
			if p.curTokenIs(lexer.TOKEN_IDENT) {
				stmt.BoxFill = strings.ToUpper(p.curToken.Literal)
			}
		}

		return stmt
	}

	return nil
}

func (p *Parser) parseOnStatement() ast.Statement {
	line := p.curToken.Line

	p.nextToken() // skip ON
	expr := p.parseExpression(LOWEST)

	p.nextToken()

	if p.curTokenIs(lexer.TOKEN_GOTO) {
		stmt := &ast.OnGotoStmt{Line: line, Expression: expr}
		p.nextToken()
		for !p.curTokenIs(lexer.TOKEN_NEWLINE) && !p.curTokenIs(lexer.TOKEN_EOF) {
			stmt.Targets = append(stmt.Targets, p.curToken.Literal)
			if p.peekTokenIs(lexer.TOKEN_COMMA) {
				p.nextToken()
			}
			p.nextToken()
		}
		return stmt
	}

	if p.curTokenIs(lexer.TOKEN_GOSUB) {
		stmt := &ast.OnGosubStmt{Line: line, Expression: expr}
		p.nextToken()
		for !p.curTokenIs(lexer.TOKEN_NEWLINE) && !p.curTokenIs(lexer.TOKEN_EOF) {
			stmt.Targets = append(stmt.Targets, p.curToken.Literal)
			if p.peekTokenIs(lexer.TOKEN_COMMA) {
				p.nextToken()
			}
			p.nextToken()
		}
		return stmt
	}

	return nil
}

// Expression parsing

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.TOKEN_NEWLINE) && !p.peekTokenIs(lexer.TOKEN_EOF) &&
		precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	line := p.curToken.Line
	name := p.curToken.Literal

	// Determine type hint from suffix
	var typeHint ast.DataType
	if len(name) > 0 {
		lastChar := name[len(name)-1:]
		typeHint = ast.DataTypeFromSuffix(lastChar)
	}

	// Check for array access or function call
	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken() // move to (
		p.nextToken() // move past (

		var args []ast.Expression
		for !p.curTokenIs(lexer.TOKEN_RPAREN) && !p.curTokenIs(lexer.TOKEN_EOF) {
			arg := p.parseExpression(LOWEST)
			if arg != nil {
				args = append(args, arg)
			}
			if p.peekTokenIs(lexer.TOKEN_COMMA) {
				p.nextToken()
			}
			p.nextToken()
		}

		// Could be array access or function call - we'll determine at runtime
		return &ast.CallExpr{Line: line, Function: name, Arguments: args}
	}

	return &ast.Identifier{Line: line, Name: name, TypeHint: typeHint}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Line: p.curToken.Line}

	// Remove any type suffix for parsing
	numStr := p.curToken.Literal
	numStr = strings.TrimRight(numStr, "%&!#")

	value, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("line %d: could not parse %q as integer", p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Line: p.curToken.Line}

	// Remove any type suffix and normalize D notation
	numStr := p.curToken.Literal
	numStr = strings.TrimRight(numStr, "!#")
	numStr = strings.Replace(numStr, "D", "E", 1)
	numStr = strings.Replace(numStr, "d", "e", 1)

	value, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		msg := fmt.Sprintf("line %d: could not parse %q as float", p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Line: p.curToken.Line, Value: p.curToken.Literal}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.UnaryExpr{
		Line:     p.curToken.Line,
		Operator: tokenToOperator[p.curToken.Type],
	}

	p.nextToken()
	expression.Right = p.parseExpression(NEGATE)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.BinaryExpr{
		Line:     p.curToken.Line,
		Operator: tokenToOperator[p.curToken.Type],
		Left:     left,
	}

	precedence := p.curPrecedence()

	// Right-associative for exponentiation
	if p.curTokenIs(lexer.TOKEN_CARET) {
		precedence--
	}

	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	return &ast.GroupedExpr{Line: p.curToken.Line, Expression: exp}
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	ident, ok := function.(*ast.Identifier)
	if !ok {
		return function
	}

	exp := &ast.CallExpr{Line: p.curToken.Line, Function: ident.Name}
	exp.Arguments = p.parseExpressionList(lexer.TOKEN_RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	var list []ast.Expression

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// parseRedimStatement parses REDIM [PRESERVE] array(newsize)
func (p *Parser) parseRedimStatement() ast.Statement {
	stmt := &ast.RedimStmt{Line: p.curToken.Line}

	p.nextToken() // skip REDIM

	// Check for PRESERVE
	if p.curTokenIs(lexer.TOKEN_PRESERVE) {
		stmt.Preserve = true
		p.nextToken()
	}

	// Parse variable declarations (similar to DIM)
	for {
		dimVar := ast.DimVariable{Name: p.curToken.Literal}

		// Check for array dimensions
		if p.peekTokenIs(lexer.TOKEN_LPAREN) {
			p.nextToken() // move to (
			p.nextToken() // move past (
			for !p.curTokenIs(lexer.TOKEN_RPAREN) {
				dim := p.parseExpression(LOWEST)
				dimVar.Dimensions = append(dimVar.Dimensions, dim)
				if p.peekTokenIs(lexer.TOKEN_COMMA) {
					p.nextToken()
				}
				p.nextToken()
			}
		}

		// Check for AS type
		if p.peekTokenIs(lexer.TOKEN_AS) {
			p.nextToken() // move to AS
			p.nextToken() // move to type
			dimVar.DataType = p.parseDataType()
		}

		stmt.Variables = append(stmt.Variables, dimVar)

		if !p.peekTokenIs(lexer.TOKEN_COMMA) {
			break
		}
		p.nextToken() // move to comma
		p.nextToken() // move past comma
	}

	return stmt
}

// parseGetStatement parses GET #n, position, variable
func (p *Parser) parseGetStatement() ast.Statement {
	stmt := &ast.GetStmt{Line: p.curToken.Line}

	// Expect #
	if !p.expectPeek(lexer.TOKEN_HASH) {
		return nil
	}
	p.nextToken()
	stmt.FileNum = p.parseExpression(LOWEST)

	// Expect comma
	if !p.expectPeek(lexer.TOKEN_COMMA) {
		return nil
	}
	p.nextToken()

	// Position is optional (can be empty between commas)
	if !p.curTokenIs(lexer.TOKEN_COMMA) {
		stmt.Position = p.parseExpression(LOWEST)
	}

	// Expect comma before variable
	if p.curTokenIs(lexer.TOKEN_COMMA) || p.peekTokenIs(lexer.TOKEN_COMMA) {
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		p.nextToken()
		stmt.Variable = p.parseExpression(LOWEST)
	}

	return stmt
}

// parsePutStatement parses PUT #n, position, variable
func (p *Parser) parsePutStatement() ast.Statement {
	stmt := &ast.PutStmt{Line: p.curToken.Line}

	// Expect #
	if !p.expectPeek(lexer.TOKEN_HASH) {
		return nil
	}
	p.nextToken()
	stmt.FileNum = p.parseExpression(LOWEST)

	// Expect comma
	if !p.expectPeek(lexer.TOKEN_COMMA) {
		return nil
	}
	p.nextToken()

	// Position is optional (can be empty between commas)
	if !p.curTokenIs(lexer.TOKEN_COMMA) {
		stmt.Position = p.parseExpression(LOWEST)
	}

	// Expect comma before variable
	if p.curTokenIs(lexer.TOKEN_COMMA) || p.peekTokenIs(lexer.TOKEN_COMMA) {
		if p.peekTokenIs(lexer.TOKEN_COMMA) {
			p.nextToken()
		}
		p.nextToken()
		stmt.Variable = p.parseExpression(LOWEST)
	}

	return stmt
}

// parseSeekStatement parses SEEK #n, position
func (p *Parser) parseSeekStatement() ast.Statement {
	stmt := &ast.SeekStmt{Line: p.curToken.Line}

	// Expect #
	if !p.expectPeek(lexer.TOKEN_HASH) {
		return nil
	}
	p.nextToken()
	stmt.FileNum = p.parseExpression(LOWEST)

	// Expect comma
	if !p.expectPeek(lexer.TOKEN_COMMA) {
		return nil
	}
	p.nextToken()
	stmt.Position = p.parseExpression(LOWEST)

	return stmt
}

// parsePsetStatement parses PSET (x, y), color
func (p *Parser) parsePsetStatement() ast.Statement {
	stmt := &ast.PsetStmt{Line: p.curToken.Line}

	// Expect (
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.X = p.parseExpression(LOWEST)

	// Expect comma
	if !p.expectPeek(lexer.TOKEN_COMMA) {
		return nil
	}
	p.nextToken()
	stmt.Y = p.parseExpression(LOWEST)

	// Expect )
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// Optional color
	if p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		stmt.Color = p.parseExpression(LOWEST)
	}

	return stmt
}

// parseCircleStatement parses CIRCLE (x, y), radius, color
func (p *Parser) parseCircleStatement() ast.Statement {
	stmt := &ast.CircleStmt{Line: p.curToken.Line}

	// Expect (
	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.X = p.parseExpression(LOWEST)

	// Expect comma
	if !p.expectPeek(lexer.TOKEN_COMMA) {
		return nil
	}
	p.nextToken()
	stmt.Y = p.parseExpression(LOWEST)

	// Expect )
	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// Expect comma and radius
	if !p.expectPeek(lexer.TOKEN_COMMA) {
		return nil
	}
	p.nextToken()
	stmt.Radius = p.parseExpression(LOWEST)

	// Optional color
	if p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken()
		p.nextToken()
		stmt.Color = p.parseExpression(LOWEST)
	}

	return stmt
}
