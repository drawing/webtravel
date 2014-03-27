package ecmascript

func isDecimalDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') ||
		(ch >= 'a' && ch <= 'f') ||
		(ch >= 'A' && ch <= 'F')
}

func isOctalDigit(ch byte) bool {
	return ch >= '0' && ch <= '7'
}

func isWhiteSpace(ch byte) bool {
	return ch == 32 ||
		ch == 9 ||
		ch == 0xB ||
		ch == 0xC ||
		ch == 0xA0
}

func isLineTerminator(ch byte) bool {
	return ch == 10 || ch == 13
}

func isIdentifierStart(ch byte) bool {
	return ch == 36 || ch == 95 ||
		(ch >= 65 && ch <= 90) ||
		(ch >= 97 && ch <= 122)
}
func isIdentifierPart(ch byte) bool {
	return ch == 36 || ch == 95 ||
		(ch >= 65 && ch <= 90) ||
		(ch >= 97 && ch <= 122) ||
		(ch >= 48 && ch <= 57)
}

func isFutureReservedWord(id []byte) bool {
	sid := string(id)
	if sid == "class" || sid == "enum" || sid == "export" ||
		sid == "extends" || sid == "import" || sid == "super" {
		return true
	}
	return false
}

func isStrictModeReservedWord(id []byte) bool {
	sid := string(id)
	if sid == "implements" || sid == "interface" || sid == "package" ||
		sid == "private" || sid == "protected" || sid == "public" ||
		sid == "static" || sid == "yield" || sid == "let" {
		return true
	}
	return false
}

func isRestrictedWord(id []byte) bool {
	sid := string(id)
	return sid == "eval" || sid == "arguments"
}

func isKeyword(id []byte, strict bool) bool {
	if strict && isStrictModeReservedWord(id) {
		return true
	}
	sid := string(id)

	switch len(sid) {
	case 2:
		return sid == "if" || sid == "in" || sid == "do"
	case 3:
		return sid == "var" || sid == "for" || sid == "new" ||
			sid == "try" || sid == "let"
	case 4:
		return sid == "this" || sid == "else" || sid == "case" ||
			sid == "void" || sid == "with" || sid == "enum"
	case 5:
		return sid == "while" || sid == "break" || sid == "catch" ||
			sid == "throw" || sid == "const" || sid == "yield" ||
			sid == "class" || sid == "super"
	case 6:
		return sid == "return" || sid == "typeof" || sid == "delete" ||
			sid == "switch" || sid == "export" || sid == "import"
	case 7:
		return sid == "default" || sid == "finally" || sid == "extends"
	case 8:
		return sid == "function" || sid == "continue" || sid == "debugger"
	case 10:
		return sid == "instanceof"
	}
	return false
}
