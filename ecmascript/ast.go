package ecmascript

import (
	// "log"
	"bytes"
	"io"
	// "fmt"
)

type Stmt interface {
	stmtNode()
	generate(w io.Writer)
}

type (
	ArrayExpression struct {
		elements []Stmt
	}
	AssignmentExpression struct {
		operator []byte
		left     Stmt
		right    Stmt
	}
	BinaryExpression struct {
		operator []byte
		left     Stmt
		right    Stmt
	}
	BlockStatement struct {
		elements []Stmt
	}
	BreakStatement struct {
		label Stmt
	}
	CallExpression struct {
		callee    Stmt
		arguments []Stmt
	}
	CatchClause struct {
		param Stmt
		body  Stmt
	}

	ConditionalExpression struct {
		expr       Stmt
		consequent Stmt
		alternate  Stmt
	}

	ContinueStatement struct {
		label Stmt
	}
	DebuggerStatement struct {
	}
	DoWhileStatement struct {
		body Stmt
		test Stmt
	}
	EmptyStatement struct {
	}
	ExpressionStatement struct {
		expr     Stmt
		brackets bool
	}
	ForStatement struct {
		init   Stmt
		test   Stmt
		update Stmt
		body   Stmt
	}
	ForInStatement struct {
		left  Stmt
		right Stmt
		body  Stmt
	}
	FunctionDeclaration struct {
		id     Stmt
		params []Stmt
		body   Stmt
	}
	FunctionExpression struct {
		header   bool
		id       Stmt
		params   []Stmt
		defaults string
		body     Stmt
	}
	Identifier struct {
		value []byte
	}
	IfStatement struct {
		test       Stmt
		consequent Stmt
		alternate  Stmt
	}
	LabeledStatement struct {
		expr Stmt
		body Stmt
	}
	Literal struct {
		value []byte
	}

	MemberExpression struct {
		computed bool
		object   Stmt
		property Stmt
	}
	NewExpression struct {
		callee Stmt
		args   []Stmt
	}
	ObjectExpression struct {
		properties []Stmt
	}
	PostfixExpression struct {
		operator []byte
		argument Stmt
	}
	Program struct {
		body []Stmt
	}
	Property struct {
		kind  string
		key   Stmt
		value Stmt
	}

	ReturnStatement struct {
		argument Stmt
	}
	SequenceExpression struct {
		expr []Stmt
	}
	SwitchCase struct {
		test       Stmt
		consequent []Stmt
	}
	SwitchStatement struct {
		discriminant Stmt
		cases        []Stmt
	}

	ThisExpression struct {
	}
	ThrowStatement struct {
		argument Stmt
	}
	TryStatement struct {
		block     Stmt
		handers   []Stmt
		finalizer Stmt
	}
	UnaryExpression struct {
		operator []byte
		argument Stmt
		prefix   bool
	}

	VariableDeclaration struct {
		kind         []byte
		declarations []Stmt
		semicolon    bool
	}
	VariableDeclarator struct {
		id   Stmt
		init Stmt
	}

	WhileStatement struct {
		test Stmt
		body Stmt
	}
	WithStatement struct {
		object Stmt
		body   Stmt
	}
)

type GroupStatement struct {
	expr Stmt
}

func (s *ArrayExpression) stmtNode()       {}
func (s *AssignmentExpression) stmtNode()  {}
func (s *BinaryExpression) stmtNode()      {}
func (s *BlockStatement) stmtNode()        {}
func (s *BreakStatement) stmtNode()        {}
func (s *CallExpression) stmtNode()        {}
func (s *CatchClause) stmtNode()           {}
func (s *ConditionalExpression) stmtNode() {}
func (s *ContinueStatement) stmtNode()     {}
func (s *DebuggerStatement) stmtNode()     {}
func (s *DoWhileStatement) stmtNode()      {}
func (s *EmptyStatement) stmtNode()        {}
func (s *ExpressionStatement) stmtNode()   {}
func (s *ForStatement) stmtNode()          {}
func (s *ForInStatement) stmtNode()        {}
func (s *FunctionDeclaration) stmtNode()   {}
func (s *FunctionExpression) stmtNode()    {}
func (s *Identifier) stmtNode()            {}
func (s *IfStatement) stmtNode()           {}
func (s *LabeledStatement) stmtNode()      {}
func (s *Literal) stmtNode()               {}
func (s *MemberExpression) stmtNode()      {}
func (s *NewExpression) stmtNode()         {}
func (s *ObjectExpression) stmtNode()      {}
func (s *PostfixExpression) stmtNode()     {}
func (s *Program) stmtNode()               {}
func (s *Property) stmtNode()              {}
func (s *ReturnStatement) stmtNode()       {}
func (s *SequenceExpression) stmtNode()    {}
func (s *SwitchCase) stmtNode()            {}
func (s *SwitchStatement) stmtNode()       {}
func (s *ThisExpression) stmtNode()        {}
func (s *ThrowStatement) stmtNode()        {}
func (s *TryStatement) stmtNode()          {}
func (s *UnaryExpression) stmtNode()       {}
func (s *VariableDeclaration) stmtNode()   {}
func (s *VariableDeclarator) stmtNode()    {}
func (s *WhileStatement) stmtNode()        {}
func (s *WithStatement) stmtNode()         {}

func (s *GroupStatement) stmtNode() {}

func (s *ArrayExpression) generate(w io.Writer) {
	w.Write([]byte("["))
	for k, v := range s.elements {
		if k != 0 {
			w.Write([]byte(","))
		}
		if v != nil {
			v.generate(w)
		}
	}
	w.Write([]byte("]"))
}

func (s *AssignmentExpression) generate(w io.Writer) {
	// fmt.Println("AssignmentExpression", string(s.operator))
	if bytes.Compare(s.operator, []byte("=")) != 0 {
		s.left.generate(w)
		w.Write(s.operator)
		s.right.generate(w)
	} else {
		b := new(bytes.Buffer)
		s.left.generate(b)

		hook_func := assignmentHook(b.Bytes())

		w.Write(b.Bytes())
		w.Write(s.operator)

		if hook_func != nil {
			w.Write(hook_func)
			w.Write([]byte("("))
		}
		s.right.generate(w)
		if hook_func != nil {
			w.Write([]byte(")"))
		}
	}
}
func (s *BinaryExpression) generate(w io.Writer) {
	s.left.generate(w)
	w.Write([]byte(" "))
	w.Write(s.operator)
	w.Write([]byte(" "))
	s.right.generate(w)
}
func (s *BlockStatement) generate(w io.Writer) {
	w.Write([]byte("{"))
	for _, v := range s.elements {
		v.generate(w)
	}
	w.Write([]byte("}"))
}
func (s *BreakStatement) generate(w io.Writer) {
	w.Write([]byte("break "))
	if s.label != nil {
		s.label.generate(w)
	}
	w.Write([]byte(";"))
}
func (s *CallExpression) generate(w io.Writer) {
	b := new(bytes.Buffer)
	s.callee.generate(b)
	bs := b.Bytes()

	replaceFunc, replaceNum, argumentType := calleeHook(bs, len(s.arguments))
	w.Write(bs)
	w.Write([]byte("("))
	for k, v := range s.arguments {
		if k != 0 {
			w.Write([]byte(","))
		}
		if replaceFunc != nil && replaceNum == k {
			w.Write([]byte(replaceFunc))
			w.Write([]byte("("))
			if argumentType == 0 {
				v.generate(w)
			} else if argumentType == 1 {
				for k1, v1 := range s.arguments {
					if k1 != 0 {
						w.Write([]byte(","))
					}
					v1.generate(w)
				}
			}
			w.Write([]byte(")"))
		} else {
			v.generate(w)
		}
	}
	w.Write([]byte(")"))
}
func (s *CatchClause) generate(w io.Writer) {
	w.Write([]byte("catch("))
	s.param.generate(w)
	w.Write([]byte(")"))
	s.body.generate(w)
}
func (s *ConditionalExpression) generate(w io.Writer) {
	// w.Write([]byte("("))
	s.expr.generate(w)
	w.Write([]byte("?"))
	s.consequent.generate(w)
	w.Write([]byte(":"))
	s.alternate.generate(w)
	// w.Write([]byte(")"))
}
func (s *ContinueStatement) generate(w io.Writer) {
	w.Write([]byte("continue "))
	if s.label != nil {
		s.label.generate(w)
	}
	w.Write([]byte(";"))
}
func (s *DebuggerStatement) generate(w io.Writer) {
	w.Write([]byte("debugger;"))
}
func (s *DoWhileStatement) generate(w io.Writer) {
	w.Write([]byte("do "))
	s.body.generate(w)
	w.Write([]byte(" while("))
	s.test.generate(w)
	w.Write([]byte(")"))
}
func (s *EmptyStatement) generate(w io.Writer) {
	w.Write([]byte(";"))
}
func (s *ExpressionStatement) generate(w io.Writer) {
	if s.brackets {
		w.Write([]byte("("))
	}
	s.expr.generate(w)
	if s.brackets {
		w.Write([]byte(")"))
	}
	w.Write([]byte(";"))
}
func (s *ForStatement) generate(w io.Writer) {
	w.Write([]byte("for("))
	if s.init != nil {
		s.init.generate(w)
	}
	w.Write([]byte(";"))
	if s.test != nil {
		s.test.generate(w)
	}
	w.Write([]byte(";"))
	if s.update != nil {
		s.update.generate(w)
	}
	w.Write([]byte(")"))
	s.body.generate(w)
}
func (s *ForInStatement) generate(w io.Writer) {
	w.Write([]byte("for("))
	s.left.generate(w)
	w.Write([]byte(" in "))
	s.right.generate(w)
	w.Write([]byte(")"))
	s.body.generate(w)
}
func (s *FunctionDeclaration) generate(w io.Writer) {
	w.Write([]byte("function "))
	if s.id != nil {
		s.id.generate(w)
	}

	w.Write([]byte("("))
	for k, v := range s.params {
		if k != 0 {
			w.Write([]byte(","))
		}
		v.generate(w)
	}
	w.Write([]byte(")"))

	s.body.generate(w)
}
func (s *FunctionExpression) generate(w io.Writer) {
	if s.header {
		w.Write([]byte("function "))
	}

	if s.id != nil {
		s.id.generate(w)
	}
	w.Write([]byte("("))
	for k, v := range s.params {
		if k != 0 {
			w.Write([]byte(","))
		}
		v.generate(w)
	}
	w.Write([]byte(")"))
	s.body.generate(w)
}
func (s *Identifier) generate(w io.Writer) {
	w.Write(s.value)
}
func (s *IfStatement) generate(w io.Writer) {
	w.Write([]byte("if ("))
	s.test.generate(w)
	w.Write([]byte(")"))
	s.consequent.generate(w)
	if s.alternate != nil {
		w.Write([]byte("else "))
		s.alternate.generate(w)
	}
}
func (s *LabeledStatement) generate(w io.Writer) {
	s.expr.generate(w)
	w.Write([]byte(":"))
	if s.body != nil {
		s.body.generate(w)
	}
}
func (s *Literal) generate(w io.Writer) {
	w.Write(s.value)
}
func (s *MemberExpression) generate(w io.Writer) {
	b := new(bytes.Buffer)
	s.object.generate(b)
	if s.computed {
		b.Write([]byte("["))
		s.property.generate(b)
		b.Write([]byte("]"))
	} else {
		b.Write([]byte("."))
		s.property.generate(b)
	}

	convert_func, present := memberConfigList[b.String()]
	if present {
		w.Write(convert_func)
		w.Write([]byte("("))
		w.Write(b.Bytes())
		w.Write([]byte(")"))
	} else {
		w.Write(b.Bytes())
	}
}
func (s *NewExpression) generate(w io.Writer) {
	w.Write([]byte("new "))
	s.callee.generate(w)
	w.Write([]byte("("))
	if s.args != nil {
		for k, v := range s.args {
			if k != 0 {
				w.Write([]byte(","))
			}
			v.generate(w)
		}
	}
	w.Write([]byte(")"))
}
func (s *ObjectExpression) generate(w io.Writer) {
	w.Write([]byte("{"))
	for k, v := range s.properties {
		if k != 0 {
			w.Write([]byte(","))
		}
		v.generate(w)
	}
	w.Write([]byte("}"))
}
func (s *PostfixExpression) generate(w io.Writer) {
	s.argument.generate(w)
	w.Write(s.operator)
}
func (s *Program) generate(w io.Writer) {
	for _, v := range s.body {
		v.generate(w)
	}
}
func (s *Property) generate(w io.Writer) {
	if s.kind != "init" {
		w.Write([]byte(s.kind))
		w.Write([]byte(" "))
		s.key.generate(w)
		s.value.generate(w)
	} else {
		s.key.generate(w)
		w.Write([]byte(":"))
		s.value.generate(w)
	}
}
func (s *ReturnStatement) generate(w io.Writer) {
	w.Write([]byte("return "))
	if s.argument != nil {
		s.argument.generate(w)
	}
	w.Write([]byte(";"))
}
func (s *SequenceExpression) generate(w io.Writer) {
	for k, v := range s.expr {
		if k != 0 {
			w.Write([]byte(","))
		}
		v.generate(w)
	}
}
func (s *SwitchCase) generate(w io.Writer) {
	if s.test == nil {
		w.Write([]byte("default:"))
	} else {
		w.Write([]byte("case "))
		s.test.generate(w)
		w.Write([]byte(":"))
	}
	for _, v := range s.consequent {
		v.generate(w)
	}

}
func (s *SwitchStatement) generate(w io.Writer) {
	w.Write([]byte("switch("))
	s.discriminant.generate(w)
	w.Write([]byte("){"))
	if s.cases != nil {
		for _, v := range s.cases {
			v.generate(w)
		}
	}
	w.Write([]byte("}"))
}
func (s *ThisExpression) generate(w io.Writer) {
	w.Write([]byte("this"))
}
func (s *ThrowStatement) generate(w io.Writer) {
	w.Write([]byte("throw "))
	s.argument.generate(w)
	w.Write([]byte(";"))
}
func (s *TryStatement) generate(w io.Writer) {
	w.Write([]byte("try"))
	s.block.generate(w)
	if s.handers != nil {
		for _, v := range s.handers {
			v.generate(w)
		}
	}
	if s.finalizer != nil {
		w.Write([]byte("finally"))
		s.finalizer.generate(w)
	}
}
func (s *UnaryExpression) generate(w io.Writer) {
	w.Write(s.operator)
	w.Write([]byte(" "))
	s.argument.generate(w)
}
func (s *VariableDeclaration) generate(w io.Writer) {
	w.Write(s.kind)
	w.Write([]byte(" "))

	for k, v := range s.declarations {
		if k != 0 {
			w.Write([]byte(","))
		}
		v.generate(w)
	}
	if s.semicolon {
		w.Write([]byte(";"))
	}
}
func (s *VariableDeclarator) generate(w io.Writer) {
	s.id.generate(w)
	if s.init != nil {
		w.Write([]byte("="))
		s.init.generate(w)
	}
}
func (s *WhileStatement) generate(w io.Writer) {
	w.Write([]byte("while("))
	s.test.generate(w)
	w.Write([]byte(")"))
	s.body.generate(w)
}
func (s *WithStatement) generate(w io.Writer) {
	w.Write([]byte("with("))
	s.object.generate(w)
	w.Write([]byte(")"))
	s.body.generate(w)
}

func (s *GroupStatement) generate(w io.Writer) {
	w.Write([]byte("("))
	s.expr.generate(w)
	w.Write([]byte(")"))
}
