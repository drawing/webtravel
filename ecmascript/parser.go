package ecmascript

import (
	"bytes"
	"errors"
	// "fmt"
)

type parser struct {
	source     []byte
	index      int
	lineNumber int

	strict            bool
	lookahead         Token
	lookahead_end_pos int

	state_allowIn        bool
	state_inFunctionBody bool
	state_inSwitch       bool
	state_inIteration    bool
}

func (p *parser) skipComment() error {
	blockComment := false
	lineComment := false

	for p.index < len(p.source) {
		ch := p.source[p.index]

		if lineComment {
			p.index++
			if isLineTerminator(ch) {
				lineComment = false
				if ch == 13 && p.source[p.index] == 10 {
					p.index++
				}
				p.lineNumber++
			}
		} else if blockComment {
			if isLineTerminator(ch) {
				if ch == 13 && p.source[p.index+1] == 10 {
					p.index++
				}
				p.index++
				p.lineNumber++
				if p.index >= len(p.source) {
					return errors.New("UnexpectedToken")
				}
			} else {
				ch = p.source[p.index]
				p.index++
				if p.index >= len(p.source) {
					return errors.New("UnexpectedToken")
				}

				if ch == 42 {
					ch = p.source[p.index]
					if ch == 47 {
						p.index++
						blockComment = false
					}
				}
			}
		} else if ch == 47 {
			ch = p.source[p.index+1]
			if ch == 47 {
				p.index += 2
				lineComment = true
			} else if ch == 42 {
				p.index += 2
				blockComment = true
				if p.index >= len(p.source) {
					return errors.New("UnexpectedToken")
				}
			} else {
				break
			}
		} else if isWhiteSpace(ch) {
			p.index++
		} else if isLineTerminator(ch) {
			p.index++
			if ch == 13 && p.source[p.index] == 10 {
				p.index++
			}
			p.lineNumber++
		} else {
			break
		}
	}

	return nil
}

func (p *parser) getIdentifier() ([]byte, error) {
	start := p.index
	p.index++
	for p.index < len(p.source) {
		ch := p.source[p.index]
		if ch == 92 {
			return p.source[start:p.index], errors.New("EscapeIdentifier Unsupported")
		}
		if isIdentifierPart(ch) {
			p.index++
		} else {
			break
		}
	}
	return p.source[start:p.index], nil
}

func (p *parser) scanIdentifier() (Token, error) {
	start := p.index
	id, err := p.getIdentifier()

	if err != nil {
		return NULLToken, err
	}

	sid := string(id)
	if len(id) == 1 {
		return Token{IdentifierToken, p.source[start:p.index]}, nil
	} else if isKeyword(id, p.strict) {
		return Token{KeywordToken, p.source[start:p.index]}, nil
	} else if sid == "null" {
		return Token{NullLiteralToken, p.source[start:p.index]}, nil
	} else if sid == "true" || sid == "false" {
		return Token{BooleanLiteralToken, p.source[start:p.index]}, nil
	} else {
		return Token{IdentifierToken, p.source[start:p.index]}, nil
	}
	return NULLToken, nil
}

func (p *parser) scanPunctuator() (Token, error) {
	start := p.index
	code := p.source[start]
	switch code {
	case 46, 40, 41, 59, 44, 123, 125, 91, 93, 58, 63, 126:
		p.index++
		return Token{PunctuatorToken, p.source[start:p.index]}, nil
	default:
		code2 := p.source[start+1]
		if code2 == 61 {
			switch code {
			case 37, 38, 42, 43, 45, 47, 60, 62, 94, 124:
				p.index += 2
				return Token{PunctuatorToken, p.source[start:p.index]}, nil
			case 33, 61:
				p.index += 2
				if p.source[p.index] == 61 {
					p.index++
					return Token{PunctuatorToken, p.source[start:p.index]}, nil
				}
				return Token{PunctuatorToken, p.source[start:p.index]}, nil
			}
		}
	}

	ch1 := p.source[p.index]
	ch2 := p.source[p.index+1]
	ch3 := p.source[p.index+2]
	ch4 := p.source[p.index+3]

	if ch1 == '>' && ch2 == '>' && ch3 == '>' {
		if ch4 == '=' {
			p.index += 4
		} else {
			p.index += 3
		}
		return Token{PunctuatorToken, p.source[start:p.index]}, nil
	}

	if ch1 == '<' && ch2 == '<' && ch3 == '=' {
		p.index += 3
		return Token{PunctuatorToken, p.source[start:p.index]}, nil
	}
	if ch1 == '>' && ch2 == '>' && ch3 == '=' {
		p.index += 3
		return Token{PunctuatorToken, p.source[start:p.index]}, nil
	}

	if ch1 == ch2 && bytes.Contains([]byte("+-<>&|"), []byte{ch1}) {
		p.index += 2
		return Token{PunctuatorToken, p.source[start:p.index]}, nil
	}
	if bytes.Contains([]byte("<>=!+-*%&|^/"), []byte{ch1}) {
		p.index++
		return Token{PunctuatorToken, p.source[start:p.index]}, nil
	}
	return NULLToken, errors.New("UnexpectedToken")
}

func (p *parser) scanNumericLiteral() (Token, error) {
	ch := p.source[p.index]
	start := p.index
	if ch != '.' {
		p.index++

		if ch == '0' {
			ch = p.source[p.index]
			if ch == 'x' || ch == 'X' {
				p.index++
				for p.index < len(p.source) {
					if !isHexDigit(p.source[p.index]) {
						break
					}
					p.index++
				}
				return Token{NumericLiteralToken, p.source[start:p.index]}, nil
			}
			if isOctalDigit(ch) {
				p.index++
				for p.index < len(p.source) {
					if !isOctalDigit(p.source[p.index]) {
						break
					}
					p.index++
				}
				return Token{NumericLiteralToken, p.source[start:p.index]}, nil
			}
		}

		for isDecimalDigit(p.source[p.index]) {
			p.index++
		}
		ch = p.source[p.index]
	}

	if ch == '.' {
		p.index++
		for isDecimalDigit(p.source[p.index]) {
			p.index++
		}
		ch = p.source[p.index]
	}

	if ch == 'e' || ch == 'E' {
		p.index++
		ch = p.source[p.index]
		if ch == '+' || ch == '-' {
			p.index++
		}
		for isDecimalDigit(p.source[p.index]) {
			p.index++
		}
	}

	return Token{NumericLiteralToken, p.source[start:p.index]}, nil
}

func (p *parser) scanStringLiteral() (Token, error) {
	quote := p.source[p.index]
	if quote != '\'' && quote != '"' {
		return NULLToken, errors.New("StringLiteral error")
	}

	start := p.index
	p.index++
	escape := false

	for p.index < len(p.source) {
		ch := p.source[p.index]
		p.index++

		if escape {
			escape = false
			continue
		}
		if ch == quote {
			break
		}

		if ch == '\\' {
			escape = true
		}
	}

	return Token{StringLiteralToken, p.source[start:p.index]}, nil
}

func (p *parser) scanRegExp() ([]byte, error) {
	p.skipComment()
	start := p.index
	ch := p.source[p.index]
	if ch != '/' {
		return p.source[start:p.index], errors.New("RegExp Start Error")
	}

	p.index++

	classMarker := false
	terminated := false
	for p.index < len(p.source) {
		ch = p.source[p.index]
		p.index++
		if classMarker {
			if ch == ']' {
				classMarker = false
			}
		} else {
			if ch == '\\' {
				ch = p.source[p.index]
				p.index++
				if isLineTerminator(ch) {
					return p.source[start:p.index], errors.New("UnterminatedRegExp")
				}
			} else if ch == '/' {
				terminated = true
				break
			} else if ch == '[' {
				classMarker = true
			} else if isLineTerminator(ch) {
				return p.source[start:p.index], errors.New("UnterminatedRegExp")
			}
		}
	}
	if !terminated {
		return p.source[start:p.index], errors.New("UnterminatedRegExp")
	}

	for p.index < len(p.source) {
		ch = p.source[p.index]
		if !isIdentifierPart(ch) {
			break
		}
		p.index++
	}
	p.peek()
	return p.source[start:p.index], nil
}

func isIdentifierName(token Token) bool {
	switch token.token_type {
	case IdentifierToken, KeywordToken, BooleanLiteralToken, NullLiteralToken:
		return true
	}
	return false
}

func (p *parser) advance() (Token, error) {
	p.skipComment()

	if p.index >= len(p.source) {
		return NULLToken, nil
	}

	ch := p.source[p.index]
	// fmt.Println("aab:", ch)

	if ch == 40 || ch == 41 || ch == 58 {
		return p.scanPunctuator()
	}

	if ch == 39 || ch == 34 {
		return p.scanStringLiteral()
	}

	if isIdentifierStart(ch) {
		return p.scanIdentifier()
	}

	if ch == 46 {
		if isDecimalDigit(p.source[p.index+1]) {
			return p.scanNumericLiteral()
		}
		return p.scanPunctuator()
	}

	if isDecimalDigit(ch) {
		return p.scanNumericLiteral()
	}
	return p.scanPunctuator()
}

func (p *parser) lex() (Token, error) {
	var err error
	token := p.lookahead
	token_end_pos := p.lookahead_end_pos

	p.index = p.lookahead_end_pos
	p.lookahead, err = p.advance()
	if err != nil {
		return NULLToken, err
	}

	p.lookahead_end_pos = p.index
	p.index = token_end_pos

	return token, nil
}

func (p *parser) peek() error {
	var err error
	index := p.index
	lineNumber := p.lineNumber
	p.lookahead, err = p.advance()
	if err != nil {
		return err
	}

	p.lookahead_end_pos = p.index
	p.index = index
	p.lineNumber = lineNumber
	return nil
}

func (p *parser) peekLineTerminator() bool {
	found := false
	index := p.index
	lineNumber := p.lineNumber
	p.skipComment()
	if lineNumber != p.lineNumber {
		found = true
	}
	p.index = index
	p.lineNumber = lineNumber

	return found
}

func (p *parser) expect(value string) error {
	token, err := p.lex()

	if err != nil {
		return err
	}
	if token.token_type != PunctuatorToken ||
		value != string(token.token_value) {
		return errors.New("expect error")
	}
	return nil
}

func (p *parser) expectKeyword(value string) error {
	token, err := p.lex()
	if err != nil {
		return err
	}
	if token.token_type != KeywordToken || value != string(token.token_value) {
		return errors.New("expectKeyword error")
	}
	return nil
}

func (p *parser) match(value string) bool {
	if p.lookahead.token_type == PunctuatorToken &&
		value == string(p.lookahead.token_value) {
		return true
	}
	return false
}

func (p *parser) matchKeyword(value string) bool {
	if p.lookahead.token_type == KeywordToken &&
		value == string(p.lookahead.token_value) {
		return true
	}
	return false
}

func (p *parser) matchAssign() bool {
	if p.lookahead.token_type != PunctuatorToken {
		return false
	}
	switch string(p.lookahead.token_value) {
	case "=", "*=", "/=", "%=", "+=", "-=", "<<=", ">>=", ">>>=", "&=", "^=", "|=":
		return true
	}
	return false
}

func (p *parser) consumeSemicolon() error {
	if p.source[p.index] == 59 {
		p.lex()
		return nil
	}

	lineNumber := p.lineNumber
	err := p.skipComment()
	if err != nil {
		return err
	}
	if lineNumber != p.lineNumber {
		return nil
	}

	if p.match(";") {
		p.lex()
		return nil
	}

	if p.lookahead.token_type != EOFToken && !p.match("}") {
		return errors.New("consumeSemicolon failed")
	}

	return nil
}

func isLeftHandSide(stmt Stmt) bool {
	switch stmt.(type) {
	case *Identifier, *MemberExpression:
		return true
	}
	return false
}

func (p *parser) parseArrayInitialiser() (Stmt, error) {
	var elements []Stmt

	err := p.expect("[")
	if err != nil {
		return nil, err
	}

	for !p.match("]") {
		if p.match(",") {
			p.lex()
			elements = append(elements, nil)
		} else {
			elem, err := p.parseAssignmentExpression()
			if err != nil {
				return nil, err
			}
			elements = append(elements, elem)
			if !p.match("]") {
				err = p.expect(",")
				if err != nil {
					return nil, err
				}
			}
		}
	}

	err = p.expect("]")
	if err != nil {
		return nil, err
	}

	return &ArrayExpression{elements: elements}, nil
}

func (p *parser) parsePropertyFunction(param []Stmt) (Stmt, error) {
	previousStrict := p.strict
	body, err := p.parseFunctionSourceElements()
	if err != nil {
		return nil, err
	}
	p.strict = previousStrict
	return &FunctionExpression{header: false, params: param, body: body}, nil
}

func (p *parser) parseObjectPropertyKey() (Stmt, error) {
	token, err := p.lex()
	if err != nil {
		return nil, err
	}
	if token.token_type == StringLiteralToken ||
		token.token_type == NumericLiteralToken {
		return &Literal{value: token.token_value}, nil
	}

	return &Identifier{value: token.token_value}, nil
}

func (p *parser) parseObjectProperty() (Stmt, error) {
	token := p.lookahead

	if token.token_type == IdentifierToken {
		id, err := p.parseObjectPropertyKey()
		if err != nil {
			return nil, err
		}

		if string(token.token_value) == "get" && !p.match(":") {
			key, err := p.parseObjectPropertyKey()
			if err != nil {
				return nil, err
			}
			err = p.expect("(")
			if err != nil {
				return nil, err
			}
			err = p.expect(")")
			if err != nil {
				return nil, err
			}
			value, err := p.parsePropertyFunction(nil)
			if err != nil {
				return nil, err
			}
			return &Property{"get", key, value}, nil
		}
		if string(token.token_value) == "set" && !p.match(":") {
			key, err := p.parseObjectPropertyKey()
			if err != nil {
				return nil, err
			}
			err = p.expect("(")
			if err != nil {
				return nil, err
			}

			token := p.lookahead
			if token.token_type != IdentifierToken {
				return nil, errors.New("Unexpected")
			}
			param, err := p.parseVariableIdentifier()
			if err != nil {
				return nil, err
			}
			var params []Stmt
			params = append(params, param)

			err = p.expect(")")
			if err != nil {
				return nil, err
			}

			value, err := p.parsePropertyFunction(params)
			if err != nil {
				return nil, err
			}
			return &Property{"set", key, value}, nil
		}
		err = p.expect(":")
		if err != nil {
			return nil, err
		}
		value, err := p.parseAssignmentExpression()
		if err != nil {
			return nil, err
		}
		return &Property{"init", id, value}, nil
	}

	if token.token_type == EOFToken || token.token_type == PunctuatorToken {
		return nil, errors.New("Unexpected")
	} else {
		key, err := p.parseObjectPropertyKey()
		if err != nil {
			return nil, err
		}
		err = p.expect(":")
		if err != nil {
			return nil, err
		}
		value, err := p.parseAssignmentExpression()
		if err != nil {
			return nil, err
		}
		return &Property{"init", key, value}, nil
	}

	return nil, nil
}

func (p *parser) parseObjectInitialiser() (Stmt, error) {
	err := p.expect("{")
	if err != nil {
		return nil, err
	}

	var properties []Stmt

	for !p.match("}") {
		property, err := p.parseObjectProperty()
		if err != nil {
			return nil, err
		}

		properties = append(properties, property)

		if !p.match("}") {
			err = p.expect(",")
			if err != nil {
				return nil, err
			}
		}
	}

	err = p.expect("}")
	if err != nil {
		return nil, err
	}
	return &ObjectExpression{properties}, nil
}

func (p *parser) parseGroupExpression() (Stmt, error) {
	err := p.expect("(")
	if err != nil {
		return nil, err
	}

	expr, err := p.parseExpression()

	err = p.expect(")")
	if err != nil {
		return nil, err
	}

	return &GroupStatement{expr}, nil
}

func (p *parser) parsePrimaryExpression() (Stmt, error) {
	token_type := p.lookahead.token_type

	if token_type == IdentifierToken {
		token, err := p.lex()
		if err != nil {
			return nil, err
		}
		return &Identifier{value: token.token_value}, nil
	}

	if token_type == StringLiteralToken || token_type == NumericLiteralToken {
		token, err := p.lex()
		if err != nil {
			return nil, err
		}
		return &Literal{value: token.token_value}, nil
	}

	if token_type == KeywordToken {
		if p.matchKeyword("this") {
			p.lex()
			return &ThisExpression{}, nil
		}

		if p.matchKeyword("function") {
			return p.parseFunctionExpression()
		}
	}

	if token_type == BooleanLiteralToken || token_type == NullLiteralToken {
		token, err := p.lex()
		if err != nil {
			return nil, err
		}
		return &Literal{value: token.token_value}, nil
	}

	if p.match("[") {
		return p.parseArrayInitialiser()
	}

	if p.match("{") {
		return p.parseObjectInitialiser()
	}

	if p.match("(") {
		return p.parseGroupExpression()
	}

	if p.match("/") || p.match("/=") {
		value, err := p.scanRegExp()
		if err != nil {
			return nil, err
		}
		return &Literal{value: value}, nil
	}

	return nil, errors.New("Unexpected")
}

func (p *parser) parseArguments() ([]Stmt, error) {
	var args []Stmt
	err := p.expect("(")
	if err != nil {
		return nil, err
	}

	if !p.match(")") {
		for p.index < len(p.source) {
			element, err := p.parseAssignmentExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, element)

			if p.match(")") {
				break
			}
			err = p.expect(",")
			if err != nil {
				return nil, err
			}
		}
	}

	err = p.expect(")")
	if err != nil {
		return nil, err
	}

	return args, nil
}

func (p *parser) parseNonComputedProperty() (Stmt, error) {
	token, err := p.lex()
	if err != nil {
		return nil, err
	}
	if !isIdentifierName(token) {
		return nil, errors.New("NonComputedProperty failed")
	}
	return &Identifier{value: token.token_value}, nil
}

func (p *parser) parseNonComputedMember() (Stmt, error) {
	err := p.expect(".")
	if err != nil {
		return nil, err
	}

	return p.parseNonComputedProperty()
}

func (p *parser) parseComputedMember() (Stmt, error) {
	err := p.expect("[")
	if err != nil {
		return nil, err
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	err = p.expect("]")
	if err != nil {
		return nil, err
	}

	return expr, nil
}

func (p *parser) parseNewExpression() (Stmt, error) {
	err := p.expectKeyword("new")
	if err != nil {
		return nil, err
	}

	callee, err := p.parseLeftHandSideExpression()
	if err != nil {
		return nil, err
	}
	var args []Stmt = nil
	if p.match("(") {
		args, err = p.parseArguments()
		if err != nil {
			return nil, err
		}
	}
	return &NewExpression{callee, args}, nil
}

func (p *parser) parseLeftHandSideExpressionAllowCall() (Stmt, error) {
	var expr Stmt
	var property Stmt
	var err error

	if p.matchKeyword("new") {
		expr, err = p.parseNewExpression()
	} else {
		expr, err = p.parsePrimaryExpression()
	}
	if err != nil {
		return nil, err
	}

	for p.match(".") || p.match("[") || p.match("(") {
		if p.match("(") {
			args, err := p.parseArguments()
			if err != nil {
				return nil, err
			}
			expr = &CallExpression{expr, args}
		} else if p.match("[") {
			property, err = p.parseComputedMember()
			if err != nil {
				return nil, err
			}
			expr = &MemberExpression{true, expr, property}
		} else {
			property, err = p.parseNonComputedMember()
			if err != nil {
				return nil, err
			}
			expr = &MemberExpression{false, expr, property}
		}
	}

	return expr, nil
}

func (p *parser) parseLeftHandSideExpression() (Stmt, error) {
	var expr Stmt
	var property Stmt
	var err error

	if p.matchKeyword("new") {
		expr, err = p.parseNewExpression()
	} else {
		expr, err = p.parsePrimaryExpression()
	}

	if err != nil {
		return nil, err
	}

	for p.match(".") || p.match("[") {
		var computed bool = false
		if p.match("[") {
			computed = true
			property, err = p.parseComputedMember()
		} else {
			property, err = p.parseNonComputedMember()
		}

		if err != nil {
			return nil, err
		}

		expr = &MemberExpression{computed, expr, property}
	}

	return expr, nil
}

func (p *parser) parsePostfixExpression() (Stmt, error) {
	expr, err := p.parseLeftHandSideExpressionAllowCall()
	if err != nil {
		return nil, err
	}

	if p.lookahead.token_type != PunctuatorToken {
		return expr, nil
	}

	if (p.match("++") || p.match("--")) && !p.peekLineTerminator() {
		if !isLeftHandSide(expr) {
			return nil, errors.New("1-InvalidLHSInAssignment")
		}

		token, err := p.lex()
		if err != nil {
			return nil, err
		}

		return &PostfixExpression{token.token_value, expr}, nil
	}

	return expr, nil
}

func (p *parser) parseUnaryExpression() (Stmt, error) {
	if p.lookahead.token_type != PunctuatorToken &&
		p.lookahead.token_type != KeywordToken {
		return p.parsePostfixExpression()
	}

	if p.match("++") || p.match("--") {
		token, err := p.lex()
		if err != nil {
			return nil, err
		}

		expr, err := p.parseUnaryExpression()
		if err != nil {
			return nil, err
		}

		if !isLeftHandSide(expr) {
			return nil, errors.New("2-InvalidLHSInAssignment")
		}

		return &UnaryExpression{token.token_value, expr, true}, nil
	}

	if p.match("+") || p.match("-") || p.match("~") || p.match("!") {
		token, err := p.lex()
		if err != nil {
			return nil, err
		}

		expr, err := p.parseUnaryExpression()
		if err != nil {
			return nil, err
		}

		return &UnaryExpression{token.token_value, expr, false}, nil
	}

	if p.matchKeyword("delete") || p.matchKeyword("void") || p.matchKeyword("typeof") {
		token, err := p.lex()
		if err != nil {
			return nil, err
		}

		expr, err := p.parseUnaryExpression()
		if err != nil {
			return nil, err
		}

		return &UnaryExpression{token.token_value, expr, false}, nil
	}

	return p.parsePostfixExpression()
}

func (p *parser) binaryPrecedence(token Token, allowIn bool) int {
	var prec int = 0

	if token.token_type != PunctuatorToken &&
		token.token_type != KeywordToken {
		return 0
	}

	switch string(token.token_value) {
	case "||":
		prec = 1
	case "&&":
		prec = 2
	case "|":
		prec = 3
	case "^":
		prec = 4
	case "&":
		prec = 5
	case "==", "!=", "===", "!==":
		prec = 6
	case "<", ">", "<=", ">=", "instanceof":
		prec = 7
	case "in":
		prec = 0
		if allowIn {
			prec = 7
		}
	case "<<", ">>", ">>>":
		prec = 8
	case "+", "-":
		prec = 9
	case "*", "/", "%":
		prec = 11
	}
	return prec
}

func (p *parser) parseBinaryExpression() (Stmt, error) {
	previousAllowIn := p.state_allowIn
	p.state_allowIn = true

	expr, err := p.parseUnaryExpression()
	if err != nil {
		return nil, err
	}

	token := p.lookahead
	// fmt.Println("test:", token.token_type, string(token.token_value))
	prec := p.binaryPrecedence(token, previousAllowIn)
	// fmt.Println("test2:", prec)
	if prec == 0 {
		return expr, nil
	}

	for prec > 0 {
		token, err = p.lex()
		if err != nil {
			return nil, err
		}

		right, err := p.parseUnaryExpression()
		if err != nil {
			return nil, err
		}
		expr = &BinaryExpression{token.token_value, expr, right}

		prec = p.binaryPrecedence(p.lookahead, previousAllowIn)
	}

	p.state_allowIn = previousAllowIn

	return expr, nil
}

// 11.12 Conditional Operator
func (p *parser) parseConditionalExpression() (Stmt, error) {
	expr, err := p.parseBinaryExpression()
	if err != nil {
		return nil, err
	}

	var retExpr Stmt
	// var alternate Stmt
	var consequent Stmt

	retExpr = expr

	if p.match("?") {
		p.lex()
		previosAllowIn := p.state_allowIn
		p.state_allowIn = true
		consequent, err = p.parseAssignmentExpression()
		if err != nil {
			return nil, err
		}
		p.state_allowIn = previosAllowIn
		err = p.expect(":")
		if err != nil {
			return nil, err
		}
		alternate, err := p.parseAssignmentExpression()
		if err != nil {
			return nil, err
		}

		retExpr = &ConditionalExpression{expr: expr, consequent: consequent, alternate: alternate}
	}

	return retExpr, nil
}

func (p *parser) parseAssignmentExpression() (Stmt, error) {
	token := p.lookahead

	left, err := p.parseConditionalExpression()
	if err != nil {
		return nil, err
	}

	if p.matchAssign() {
		if !isLeftHandSide(left) {
			return nil, errors.New("LeftHandSide error")
		}

		token, err = p.lex()
		if err != nil {
			return nil, err
		}

		right, err := p.parseAssignmentExpression()
		if err != nil {
			return nil, err
		}

		return &AssignmentExpression{token.token_value, left, right}, nil
	}

	return left, nil
}

func (p *parser) parseExpression() (Stmt, error) {
	expr, err := p.parseAssignmentExpression()
	if err != nil {
		return nil, err
	}

	if !p.match(",") {
		return expr, nil
	}

	var exprlist []Stmt
	exprlist = append(exprlist, expr)

	for p.index < len(p.source) {
		if !p.match(",") {
			break
		}
		p.lex()

		expr, err = p.parseAssignmentExpression()
		if err != nil {
			return nil, err
		}
		exprlist = append(exprlist, expr)
	}

	return &SequenceExpression{expr: exprlist}, nil
}

func (p *parser) parseStatementList() ([]Stmt, error) {
	var list []Stmt

	for p.index < len(p.source) {
		if p.match("}") {
			break
		}
		statement, err := p.parseSourceElement()
		if err != nil {
			return nil, err
		}
		if statement == nil {
			break
		}
		list = append(list, statement)
	}
	return list, nil
}

func (p *parser) parseBlock() (Stmt, error) {
	err := p.expect("{")
	if err != nil {
		return nil, err
	}

	block, err := p.parseStatementList()
	if err != nil {
		return nil, err
	}

	err = p.expect("}")
	if err != nil {
		return nil, err
	}

	return &BlockStatement{elements: block}, nil
}

func (p *parser) parseVariableIdentifier() (Stmt, error) {
	token, err := p.lex()
	if err != nil {
		return nil, err
	}

	if token.token_type != IdentifierToken {
		return nil, errors.New("Unexpected")
	}

	return &Identifier{value: token.token_value}, nil
}

func (p *parser) parseVariableDeclaration(kind []byte) (Stmt, error) {
	id, err := p.parseVariableIdentifier()
	if err != nil {
		return nil, err
	}

	var init Stmt

	/*if p.strict && isRestrictedWord(id.value) {
		return nil, errors.New("StrictVarName")
	}*/

	if string(kind) == "const" {
		err := p.expect("=")
		if err != nil {
			return nil, err
		}
		init, err = p.parseAssignmentExpression()
		if err != nil {
			return nil, err
		}
	} else if p.match("=") {
		p.lex()
		init, err = p.parseAssignmentExpression()
		if err != nil {
			return nil, err
		}
	}

	return &VariableDeclarator{id: id, init: init}, nil
}

func (p *parser) parseVariableDeclarationList(kind []byte) ([]Stmt, error) {
	var list []Stmt
	for p.index < len(p.source) {
		stmt, err := p.parseVariableDeclaration(kind)
		if err != nil {
			return nil, err
		}
		list = append(list, stmt)

		if p.lookahead.token_type != PunctuatorToken || p.lookahead.token_value[0] != ',' {
			break
		}
		p.lex()
	}
	return list, nil
}

func (p *parser) parseVariableStatement() (Stmt, error) {
	err := p.expectKeyword("var")
	if err != nil {
		return nil, err
	}

	declarations, err := p.parseVariableDeclarationList([]byte("var"))
	if err != nil {
		return nil, err
	}

	err = p.consumeSemicolon()
	if err != nil {
		return nil, err
	}

	return &VariableDeclaration{[]byte("var"), declarations, true}, nil
}

func (p *parser) parseConstLetDeclaration(kind []byte) (Stmt, error) {
	err := p.expectKeyword(string(kind))
	if err != nil {
		return nil, err
	}

	declarations, err := p.parseVariableDeclarationList(kind)
	if err != nil {
		return nil, err
	}

	err = p.consumeSemicolon()
	if err != nil {
		return nil, err
	}

	return &VariableDeclaration{kind, declarations, true}, nil
}

func (p *parser) parseEmptyStatement() (Stmt, error) {
	err := p.expect(";")
	if err != nil {
		return nil, err
	}
	return &EmptyStatement{}, nil
}

func (p *parser) parseExpressionStatement() (Stmt, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	p.consumeSemicolon()
	return &ExpressionStatement{expr, true}, nil
}

func (p *parser) parseIfStatement() (Stmt, error) {
	err := p.expectKeyword("if")
	if err != nil {
		return nil, err
	}

	err = p.expect("(")
	if err != nil {
		return nil, err
	}

	test, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	err = p.expect(")")
	if err != nil {
		return nil, err
	}

	consequent, err := p.parseStatement()
	if err != nil {
		return nil, err
	}

	var alternate Stmt
	if p.matchKeyword("else") {
		p.lex()
		alternate, err = p.parseStatement()
		if err != nil {
			return nil, err
		}
	}
	return &IfStatement{test, consequent, alternate}, nil
}

func (p *parser) parseDoWhileStatement() (Stmt, error) {
	err := p.expectKeyword("do")
	if err != nil {
		return nil, err
	}

	oldInIteration := p.state_inIteration
	p.state_inIteration = true

	body, err := p.parseStatement()
	if err != nil {
		return nil, err
	}

	p.state_inIteration = oldInIteration

	err = p.expectKeyword("while")
	if err != nil {
		return nil, err
	}

	err = p.expect("(")
	if err != nil {
		return nil, err
	}

	test, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	err = p.expect(")")
	if err != nil {
		return nil, err
	}

	if p.match(";") {
		p.lex()
	}
	return &DoWhileStatement{body, test}, nil
}

func (p *parser) parseWhileStatement() (Stmt, error) {
	err := p.expectKeyword("while")
	if err != nil {
		return nil, err
	}

	err = p.expect("(")
	if err != nil {
		return nil, err
	}

	test, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	err = p.expect(")")
	if err != nil {
		return nil, err
	}

	oldInInteration := p.state_inIteration
	p.state_inIteration = true

	body, err := p.parseStatement()
	if err != nil {
		return nil, err
	}

	p.state_inIteration = oldInInteration
	return &WhileStatement{test, body}, nil
}

func (p *parser) parseForVariableDeclaration() (Stmt, error) {
	token, err := p.lex()
	if err != nil {
		return nil, err
	}

	declarations, err := p.parseVariableDeclarationList([]byte(""))
	if err != nil {
		return nil, err
	}

	return &VariableDeclaration{token.token_value, declarations, false}, nil
}

func (p *parser) parseForStatement() (Stmt, error) {
	err := p.expectKeyword("for")
	if err != nil {
		return nil, err
	}

	err = p.expect("(")
	if err != nil {
		return nil, err
	}

	var init Stmt
	var left Stmt
	var right Stmt
	if p.match(";") {
		p.lex()
	} else {
		if p.matchKeyword("var") || p.matchKeyword("let") {
			p.state_allowIn = false
			init, err = p.parseForVariableDeclaration()
			p.state_allowIn = true

			if /*len(init.declarations) == 1 && */ p.matchKeyword("in") {
				p.lex()
				left = init
				right, err = p.parseExpression()
				if err != nil {
					return nil, err
				}
			}
		} else {
			p.state_allowIn = false
			init, err = p.parseExpression()
			p.state_allowIn = true

			if p.matchKeyword("in") {
				if !isLeftHandSide(init) {
					return nil, errors.New("InvalidLHSInForIn")
				}

				p.lex()
				left = init
				right, err = p.parseExpression()
				if err != nil {
					return nil, err
				}
				init = nil
			}
		}

		if left == nil {
			err = p.expect(";")
			if err != nil {
				return nil, err
			}
		}
	}

	var test Stmt
	var update Stmt
	if left == nil {
		if !p.match(";") {
			test, err = p.parseExpression()
		}

		err = p.expect(";")
		if err != nil {
			return nil, err
		}

		if !p.match(")") {
			update, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
		}
	}

	err = p.expect(")")
	if err != nil {
		return nil, err
	}

	oldInIteration := p.state_inIteration
	p.state_inIteration = true

	body, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	p.state_inIteration = oldInIteration

	if left == nil {
		return &ForStatement{init, test, update, body}, nil
	}
	return &ForInStatement{left, right, body}, nil
}

func (p *parser) parseContinueStatement() (Stmt, error) {
	err := p.expectKeyword("continue")
	if err != nil {
		return nil, err
	}

	if p.source[p.index] == 59 {
		p.lex()
		if !p.state_inIteration {
			return nil, errors.New("IllegalContinue")
		}

		return &ContinueStatement{}, nil
	}

	if p.peekLineTerminator() {
		if !p.state_inIteration {
			return nil, errors.New("IllegalContinue")
		}
		return &ContinueStatement{}, nil
	}

	var label Stmt = nil
	if p.lookahead.token_type == IdentifierToken {
		label, err = p.parseVariableIdentifier()
		if err != nil {
			return nil, err
		}
	}

	p.consumeSemicolon()

	if label == nil && !p.state_inIteration {
		return nil, errors.New("IllegalContinue")
	}
	return &ContinueStatement{label}, nil
}

func (p *parser) parseBreakStatement() (Stmt, error) {
	err := p.expectKeyword("break")
	if err != nil {
		return nil, err
	}

	if p.source[p.index] == 59 {
		p.lex()
		if !p.state_inIteration && !p.state_inSwitch {
			return nil, errors.New("IllegalBreak")
		}

		return &BreakStatement{}, nil
	}

	if p.peekLineTerminator() {
		if !p.state_inIteration && !p.state_inSwitch {
			return nil, errors.New("IllegalBreak")
		}
		return &BreakStatement{}, nil
	}

	var label Stmt = nil
	if p.lookahead.token_type == IdentifierToken {
		label, err = p.parseVariableIdentifier()
		if err != nil {
			return nil, err
		}
	}

	p.consumeSemicolon()

	if label == nil && !(p.state_inIteration || p.state_inSwitch) {
		return nil, errors.New("IllegalBreak")
	}
	return &BreakStatement{label}, nil
}

func (p *parser) parseReturnStatement() (Stmt, error) {
	var argument Stmt

	err := p.expectKeyword("return")
	if err != nil {
		return nil, err
	}

	if !p.state_inFunctionBody {
		return nil, errors.New("IllegalReturn")
	}

	if p.source[p.index] == 32 {
		if isIdentifierStart(p.source[p.index+1]) {
			argument, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
			err = p.consumeSemicolon()
			if err != nil {
				return nil, err
			}
			return &ReturnStatement{argument: argument}, nil
		}
	}

	if p.peekLineTerminator() {
		return &ReturnStatement{}, nil
	}

	if !p.match(";") {
		if !p.match("}") && p.lookahead.token_type != EOFToken {
			argument, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
		}
	}

	err = p.consumeSemicolon()
	if err != nil {
		return nil, err
	}

	return &ReturnStatement{argument: argument}, nil
}

func (p *parser) parseWithStatement() (Stmt, error) {
	if p.strict {
		return nil, errors.New("StrictModeWith")
	}

	err := p.expectKeyword("with")
	if err != nil {
		return nil, err
	}

	err = p.expect("(")
	if err != nil {
		return nil, err
	}

	object, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	err = p.expect(")")
	if err != nil {
		return nil, err
	}

	body, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	return &WithStatement{object: object, body: body}, nil
}

func (p *parser) parseSwitchCase() (Stmt, error) {
	var test Stmt
	var consequent []Stmt
	if p.matchKeyword("default") {
		p.lex()
		test = nil
	} else {
		err := p.expectKeyword("case")
		if err != nil {
			return nil, err
		}
		test, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	err := p.expect(":")
	if err != nil {
		return nil, err
	}

	for p.index < len(p.source) {
		if p.match("}") || p.matchKeyword("default") || p.matchKeyword("case") {
			break
		}

		statement, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		consequent = append(consequent, statement)
	}

	return &SwitchCase{test: test, consequent: consequent}, nil
}

func (p *parser) parseSwitchStatement() (Stmt, error) {
	err := p.expectKeyword("switch")
	if err != nil {
		return nil, err
	}

	err = p.expect("(")
	if err != nil {
		return nil, err
	}

	discriminant, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	err = p.expect(")")
	if err != nil {
		return nil, err
	}

	err = p.expect("{")
	if err != nil {
		return nil, err
	}

	if p.match("}") {
		p.lex()
		return &SwitchStatement{discriminant: discriminant}, nil
	}

	var cases []Stmt

	oldInSwitch := p.state_inSwitch
	p.state_inSwitch = true
	// defaultFound := false

	for p.index < len(p.source) {
		if p.match("}") {
			break
		}
		clause, err := p.parseSwitchCase()
		if err != nil {
			return nil, err
		}
		/*
			if clause.test == nil {
				if defaultFound {
					return nil, errors.New("MultipleDefaultsInSwitch")
				}
				defaultFound = true
			}
		*/
		cases = append(cases, clause)
	}

	p.state_inSwitch = oldInSwitch

	err = p.expect("}")
	if err != nil {
		return nil, err
	}

	return &SwitchStatement{discriminant: discriminant, cases: cases}, nil
}

func (p *parser) parseThrowStatement() (Stmt, error) {
	err := p.expectKeyword("throw")
	if err != nil {
		return nil, err
	}

	if p.peekLineTerminator() {
		return nil, errors.New("NewlineAfterThrow")
	}

	argument, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	err = p.consumeSemicolon()
	if err != nil {
		return nil, err
	}

	return &ThrowStatement{argument}, nil
}

func (p *parser) parseCatchClause() (Stmt, error) {
	err := p.expectKeyword("catch")
	if err != nil {
		return nil, err
	}

	err = p.expect("(")
	if err != nil {
		return nil, err
	}

	if p.match(")") {
		return nil, errors.New("CatchClause failed")
	}

	param, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	err = p.expect(")")
	if err != nil {
		return nil, err
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return &CatchClause{param: param, body: body}, nil
}

func (p *parser) parseTryStatement() (Stmt, error) {
	var handlers []Stmt
	var finalizer Stmt
	err := p.expectKeyword("try")
	if err != nil {
		return nil, err
	}

	block, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	if p.matchKeyword("catch") {
		stmt, err := p.parseCatchClause()
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, stmt)
	}

	if p.matchKeyword("finally") {
		p.lex()
		finalizer, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}

	return &TryStatement{block: block, handers: handlers, finalizer: finalizer}, nil
}

func (p *parser) parseDebuggerStatement() (Stmt, error) {
	err := p.expectKeyword("debugger")
	if err != nil {
		return nil, err
	}

	err = p.consumeSemicolon()
	if err != nil {
		return nil, err
	}

	return &DebuggerStatement{}, nil
}

func (p *parser) parseStatement() (Stmt, error) {
	token_type := p.lookahead.token_type

	if token_type == EOFToken {
		return nil, errors.New("Unexpected")
	}

	if token_type == PunctuatorToken {
		switch string(p.lookahead.token_value) {
		case ";":
			return p.parseEmptyStatement()
		case "{":
			return p.parseBlock()
		case "(":
			return p.parseExpressionStatement()
		}
	}

	if token_type == KeywordToken {
		switch string(p.lookahead.token_value) {
		case "break":
			return p.parseBreakStatement()
		case "continue":
			return p.parseContinueStatement()
		case "debugger":
			return p.parseDebuggerStatement()
		case "do":
			return p.parseDoWhileStatement()
		case "for":
			return p.parseForStatement()
		case "function":
			return p.parseFunctionDeclaration()
		case "if":
			return p.parseIfStatement()
		case "return":
			return p.parseReturnStatement()
		case "switch":
			return p.parseSwitchStatement()
		case "throw":
			return p.parseThrowStatement()
		case "try":
			return p.parseTryStatement()
		case "var":
			return p.parseVariableStatement()
		case "while":
			return p.parseWhileStatement()
		case "with":
			return p.parseWithStatement()
		}
	}

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	switch expr.(type) {
	case *Identifier:
		if p.match(":") {
			p.lex()

			labeledBody, err := p.parseStatement()
			if err != nil {
				return nil, err
			}

			return &LabeledStatement{expr: expr, body: labeledBody}, nil
		}
	}

	err = p.consumeSemicolon()
	if err != nil {
		return nil, err
	}

	return &ExpressionStatement{expr, false}, nil
}

func (p *parser) parseFunctionSourceElements() (Stmt, error) {
	var sourceElements []Stmt
	err := p.expect("{")
	if err != nil {
		return nil, err
	}

	for p.index < len(p.source) {
		if p.lookahead.token_type != StringLiteralToken {
			break
		}
		// token := p.lookahead

		sourceElement, err := p.parseSourceElement()
		if err != nil {
			return nil, err
		}
		sourceElements = append(sourceElements, sourceElement)

		should_break := false
		switch sourceElement.(type) {
		case *Literal:
			should_break = true
		}
		if should_break {
			break
		}

		/*directive := sourceElement.value[1:len(sourceElement.value)-1]
		if types.Compare(directive, []byte("use strict")) == 0 {
			p.strict = true
		}*/
	}

	oldInIteration := p.state_inIteration
	oldInSwitch := p.state_inSwitch
	oldInFunctionBody := p.state_inFunctionBody

	p.state_inIteration = false
	p.state_inSwitch = false
	p.state_inFunctionBody = true

	for p.index < len(p.source) {
		if p.match("}") {
			break
		}
		sourceElement, err := p.parseSourceElement()
		if err != nil {
			return nil, err
		}
		if sourceElement == nil {
			break
		}
		sourceElements = append(sourceElements, sourceElement)
	}

	err = p.expect("}")
	if err != nil {
		return nil, err
	}

	p.state_inIteration = oldInIteration
	p.state_inSwitch = oldInSwitch
	p.state_inFunctionBody = oldInFunctionBody

	return &BlockStatement{elements: sourceElements}, nil
}

func (p *parser) parseParams() ([]Stmt, error) {
	var params []Stmt

	err := p.expect("(")
	if err != nil {
		return nil, err
	}

	if !p.match(")") {
		for p.index < len(p.source) {
			// token := p.lookahead
			param, err := p.parseVariableIdentifier()
			if err != nil {
				return nil, err
			}

			params = append(params, param)

			if p.match(")") {
				break
			}
			err = p.expect(",")
			if err != nil {
				return nil, err
			}
		}
	}

	err = p.expect(")")
	if err != nil {
		return nil, err
	}

	return params, nil
}

func (p *parser) parseFunctionDeclaration() (Stmt, error) {
	err := p.expectKeyword("function")
	if err != nil {
		return nil, err
	}

	token := p.lookahead

	id, err := p.parseVariableIdentifier()
	if err != nil {
		return nil, err
	}
	if p.strict {
		if isRestrictedWord(token.token_value) {
			return nil, errors.New("strict mode error")
		}
	}

	params, err := p.parseParams()
	if err != nil {
		return nil, err
	}

	previousStrict := p.strict
	body, err := p.parseFunctionSourceElements()
	if err != nil {
		return nil, err
	}
	p.strict = previousStrict

	return &FunctionDeclaration{id: id, params: params, body: body}, nil
}

func (p *parser) parseFunctionExpression() (Stmt, error) {
	var id Stmt
	err := p.expectKeyword("function")
	if err != nil {
		return nil, err
	}

	if !p.match("(") {
		token := p.lookahead
		id, err = p.parseVariableIdentifier()
		if err != nil {
			return nil, err
		}
		if p.strict {
			if isRestrictedWord(token.token_value) {
				return nil, errors.New("strict mode error")
			}
		}
	}

	params, err := p.parseParams()
	if err != nil {
		return nil, err
	}

	previousStrict := p.strict
	body, err := p.parseFunctionSourceElements()
	if err != nil {
		return nil, err
	}
	p.strict = previousStrict

	return &FunctionExpression{header: true, id: id, params: params, body: body}, nil
}

func (p *parser) parseSourceElement() (Stmt, error) {
	if p.lookahead.token_type == KeywordToken {
		switch string(p.lookahead.token_value) {
		case "const", "let":
			return p.parseConstLetDeclaration(p.lookahead.token_value)
		case "function":
			return p.parseFunctionDeclaration()
		default:
			return p.parseStatement()
		}
	}
	if p.lookahead.token_type != EOFToken {
		return p.parseStatement()
	}
	return nil, nil
}

func (p *parser) parseSourceElements() ([]Stmt, error) {
	var sourceElements []Stmt

	for p.index < len(p.source) {
		token := p.lookahead

		if token.token_type != StringLiteralToken {
			break
		}

		sourceElement, err := p.parseSourceElement()
		if err != nil {
			return nil, err
		}
		sourceElements = append(sourceElements, sourceElement)

		should_break := false
		switch sourceElement.(type) {
		case *Literal:
			should_break = true
		}
		if should_break {
			break
		}
		/*
			directive := sourceElement.value[1:len(sourceElement.value)-1]
			if types.Compare(directive, []byte("use strict")) == 0 {
				p.strict = true
			}
		*/
	}

	for p.index < len(p.source) {
		sourceElement, err := p.parseSourceElement()
		if err != nil {
			return nil, err
		}
		if sourceElement == nil {
			break
		}
		sourceElements = append(sourceElements, sourceElement)
	}
	return sourceElements, nil
}

func (p *parser) parseProgram(source []byte) (Stmt, error) {
	p.index = 0
	p.source = source
	p.strict = false
	p.lineNumber = 0

	p.peek()

	body, err := p.parseSourceElements()
	if err != nil {
		return nil, err
	}

	return &Program{body: body}, nil
}

type JSTransform struct {
}

func (t *JSTransform) Process(source []byte) ([]byte, error) {
	var p parser

	for i := 0; i < 7; i += 1 {
		source = append(source, '\n')
	}

	s, err := p.parseProgram(source)
	if err != nil {
		return source, err
	}

	b := new(bytes.Buffer)
	s.generate(b)

	return b.Bytes(), nil
}
