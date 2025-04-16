package main

import (
	"fmt"
)

// JavaScript Parser implementation
type JavaScriptParser struct {
	BasicParser
	strictMode bool
	esVersion  int
}

// Initialize sets up the JavaScript parser
func (p *JavaScriptParser) Initialize(source string) {
	p.BasicParser.Initialize(source)
	p.keywords = map[string]bool{
		"var":       true,
		"let":       true,
		"const":     true,
		"function":  true,
		"return":    true,
		"if":        true,
		"else":      true,
		"for":       true,
		"while":     true,
		"do":        true,
		"break":     true,
		"continue":  true,
		"switch":    true,
		"case":      true,
		"default":   true,
		"try":       true,
		"catch":     true,
		"finally":   true,
		"throw":     true,
		"class":     true,
		"extends":   true,
		"super":     true,
		"this":      true,
		"new":       true,
		"import":    true,
		"export":    true,
		"from":      true,
		"as":        true,
		"async":     true,
		"await":     true,
		"of":        true,
		"in":        true,
		"instanceof": true,
		"typeof":    true,
		"void":      true,
		"delete":    true,
		"null":      true,
		"undefined": true,
		"true":      true,
		"false":     true,
	}
	p.strictMode = false
	p.esVersion = 2020 // Default to ES2020
}

func (p *JavaScriptParser) GetLanguage() string {
	return "JavaScript"
}

func (p *JavaScriptParser) SetOptions(options ParserOptions) {
	p.BasicParser.SetOptions(options)
	p.strictMode = options.StrictMode
	if options.TargetECMAScript > 0 {
		p.esVersion = options.TargetECMAScript
	}
}

// Parse entry point for JavaScript
func (p *JavaScriptParser) Parse(source string) (*Node, error) {
	p.Initialize(source)
	
	// Create the program node
	start := p.CurrentPosition()
	programNode := CreateNode(NodeProgram, start, Position{})
	
	// Parse statements
	for p.pos < len(p.source) {
		p.SkipWhitespace()
		if p.pos >= len(p.source) {
			break
		}
		
		// Skip comments
		if p.current == '/' && p.Peek() == '/' {
			p.SkipLineComment()
			continue
		}
		if p.current == '/' && p.Peek() == '*' {
			p.SkipBlockComment("/*", "*/")
			continue
		}
		
		// Check for keywords and patterns
		var stmtNode *Node
		var err error
		
		if p.CheckKeyword("function") {
			stmtNode, err = p.parseFunctionDeclaration()
		} else if p.CheckKeyword("var") || p.CheckKeyword("let") || p.CheckKeyword("const") {
			stmtNode, err = p.parseVariableDeclaration()
		} else if p.CheckKeyword("if") {
			stmtNode, err = p.parseIfStatement()
		} else if p.CheckKeyword("for") {
			stmtNode, err = p.parseForStatement()
		} else if p.CheckKeyword("while") {
			stmtNode, err = p.parseWhileStatement()
		} else if p.CheckKeyword("class") {
			stmtNode, err = p.parseClassDeclaration()
		} else if p.CheckKeyword("import") {
			stmtNode, err = p.parseImportStatement()
		} else if p.CheckKeyword("export") {
			stmtNode, err = p.parseExportStatement()
		} else {
			// Try to parse as expression statement
			stmtNode, err = p.parseExpressionStatement()
		}
		
		if err != nil {
			// Create detailed error context
			contextStr := p.GetContext(p.line, p.column, 2)
			return nil, &ParseError{
				Line:    p.line,
				Column:  p.column,
				Message: err.Error(),
				Context: contextStr,
			}
		}
		
		programNode.Children = append(programNode.Children, stmtNode)
	}
	
	programNode.End = p.CurrentPosition()
	
	// Build symbol table for the program
	p.buildSymbolTable(programNode)
	
	return programNode, nil
}

// Build symbol table for semantic analysis
func (p *JavaScriptParser) buildSymbolTable(root *Node) *SymbolTable {
	globalScope := NewSymbolTable(nil)
	globalScope.Node = root
	
	// First pass: collect declarations
	p.collectDeclarations(root, globalScope)
	
	// Second pass: resolve references
	p.resolveReferences(root, globalScope)
	
	return globalScope
}

// Collect declarations for symbol table
func (p *JavaScriptParser) collectDeclarations(node *Node, currentScope *SymbolTable) {
	switch node.Type {
	case NodeVariableDecl:
		// Find the identifier child
		for _, child := range node.Children {
			if child.Type == NodeIdentifier {
				kind := node.Attrs["kind"] // var, let, or const
				currentScope.Add(child.Value, "variable", "", child)
				break
			}
		}
	case NodeFunctionDecl:
		if len(node.Children) > 0 && node.Children[0].Type == NodeIdentifier {
			// Function declaration - add to current scope
			fnName := node.Children[0].Value
			fnSymbol := currentScope.Add(fnName, "function", "", node.Children[0])
			
			// Create new scope for function body
			functionScope := NewSymbolTable(currentScope)
			functionScope.Node = node
			
			// Add parameters to function scope
			for i, child := range node.Children {
				if i > 0 && child.Type == NodeParameter {
					functionScope.Add(child.Value, "parameter", "", child)
				}
			}
			
			// Continue collecting declarations in function body
			for _, child := range node.Children {
				if child.Type == NodeBlockStatement {
					p.collectDeclarations(child, functionScope)
					break
				}
			}
		}
	case NodeClassDeclaration:
		if len(node.Children) > 0 && node.Children[0].Type == NodeIdentifier {
			// Class declaration - add to current scope
			className := node.Children[0].Value
			classSymbol := currentScope.Add(className, "class", "", node.Children[0])
			
			// Create new scope for class body
			classScope := NewSymbolTable(currentScope)
			classScope.Node = node
			
			// Process class body
			for _, child := range node.Children {
				if child.Type == NodeMethodDefinition || child.Type == NodePropertyDefinition {
					p.collectDeclarations(child, classScope)
				}
			}
		}
	case NodeBlockStatement:
		// Create new scope for blocks (if, for, while, etc.)
		blockScope := NewSymbolTable(currentScope)
		blockScope.Node = node
		
		// Process block contents
		for _, child := range node.Children {
			p.collectDeclarations(child, blockScope)
		}
	default:
		// Recursively process children
		for _, child := range node.Children {
			p.collectDeclarations(child, currentScope)
		}
	}
}

// Resolve references to symbols
func (p *JavaScriptParser) resolveReferences(node *Node, currentScope *SymbolTable) {
	switch node.Type {
	case NodeIdentifier:
		// Skip identifiers that are part of declarations (they're already handled)
		if node.Symbol == nil {
			// This is a reference, try to resolve it
			if sym := currentScope.Lookup(node.Value); sym != nil {
				node.Symbol = sym
				sym.References = append(sym.References, node)
			}
		}
	case NodeFunctionDecl, NodeClassDeclaration, NodeBlockStatement:
		// Find the appropriate scope for this node
		var newScope *SymbolTable = nil
		for _, scope := range currentScope.Children {
			if scope.Node == node {
				newScope = scope
				break
			}
		}
		
		if newScope == nil {
			// If we don't find a matching scope, continue with current scope
			newScope = currentScope
		}
		
		// Resolve references in children using the new scope
		for _, child := range node.Children {
			p.resolveReferences(child, newScope)
		}
		return
	}
	
	// Recursively process children
	for _, child := range node.Children {
		p.resolveReferences(child, currentScope)
	}
}

// Specialized parsing methods
func (p *JavaScriptParser) ParseExpression(source string) (*Node, error) {
	p.Initialize(source)
	p.SkipWhitespace()
	
	// Parse expression and ensure we consume the entire input
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	
	p.SkipWhitespace()
	if p.pos < len(p.source) {
		return nil, fmt.Errorf("unexpected token at position %d", p.pos)
	}
	
	return expr, nil
}

func (p *JavaScriptParser) ParseStatement(source string) (*Node, error) {
	p.Initialize(source)
	p.SkipWhitespace()
	
	var stmt *Node
	var err error
	
	// Determine statement type
	if p.CheckKeyword("function") {
		stmt, err = p.parseFunctionDeclaration()
	} else if p.CheckKeyword("var") || p.CheckKeyword("let") || p.CheckKeyword("const") {
		stmt, err = p.parseVariableDeclaration()
	} else if p.CheckKeyword("if") {
		stmt, err = p.parseIfStatement()
	} else if p.CheckKeyword("for") {
		stmt, err = p.parseForStatement()
	} else if p.CheckKeyword("while") {
		stmt, err = p.parseWhileStatement()
	} else if p.CheckKeyword("class") {
		stmt, err = p.parseClassDeclaration()
	} else {
		stmt, err = p.parseExpressionStatement()
	}
	
	if err != nil {
		return nil, err
	}
	
	p.SkipWhitespace()
	if p.pos < len(p.source) {
		return nil, fmt.Errorf("unexpected token at position %d", p.pos)
	}
	
	return stmt, nil
}

// Helper to parse expression
func (p *JavaScriptParser) parseExpression() (*Node, error) {
	// This is a simplified implementation
	// In a real parser, this would handle precedence, etc.
	
	// Check for literals, identifiers, and parenthesized expressions
	if p.IsDigit(p.current) {
		return p.parseNumberLiteral()
	} else if p.current == '"' || p.current == '\'' || p.current == '`' {
		return p.parseStringLiteral()
	} else if p.IsIdentifierStart(p.current) {
		id, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		
		// Check for keywords that are values
		if id == "true" || id == "false" {
			start := p.CurrentPosition()
			start.Column -= len(id)
			node := CreateNode(NodeLiteral, start, p.CurrentPosition())
			node.Value = id
			return node, nil
		} else if id == "null" || id == "undefined" {
			start := p.CurrentPosition()
			start.Column -= len(id)
			node := CreateNode(NodeLiteral, start, p.CurrentPosition())
			node.Value = id
			return node, nil
		}
		
		// Regular identifier
		start := p.CurrentPosition()
		start.Column -= len(id)
		node := CreateNode(NodeIdentifier, start, p.CurrentPosition())
		node.Value = id
		
		// Check for function calls: identifier()
		p.SkipWhitespace()
		if p.current == '(' {
			return p.parseCallExpression(node)
		}
		
		return node, nil
	} else if p.current == '(' {
		// Parenthesized expression
		p.Advance() // Skip '('
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.SkipWhitespace()
		if p.current != ')' {
			return nil, fmt.Errorf("expected ')' at %d:%d", p.line, p.column)
		}
		p.Advance() // Skip ')'
		return expr, nil
	}
	
	return nil, fmt.Errorf("unexpected token at %d:%d", p.line, p.column)
}

// Parse number literal
func (p *JavaScriptParser) parseNumberLiteral() (*Node, error) {
	start := p.CurrentPosition()
	
	// Parse integer part
	value := ""
	for p.pos < len(p.source) && p.IsDigit(p.current) {
		value += string(p.current)
		p.Advance()
	}
	
	// Check for decimal point
	if p.current == '.' {
		value += "."
		p.Advance()
		
		// Parse fractional part
		if !p.IsDigit(p.current) {
			return nil, fmt.Errorf("expected digit after decimal point at %d:%d", p.line, p.column)
		}
		
		for p.pos < len(p.source) && p.IsDigit(p.current) {
			value += string(p.current)
			p.Advance()
		}
	}
	
	// Check for exponent
	if p.current == 'e' || p.current == 'E' {
		value += string(p.current)
		p.Advance()
		
		// Check for sign
		if p.current == '+' || p.current == '-' {
			value += string(p.current)
			p.Advance()
		}
		
		// Parse exponent
		if !p.IsDigit(p.current) {
			return nil, fmt.Errorf("expected digit in exponent at %d:%d", p.line, p.column)
		}
		
		for p.pos < len(p.source) && p.IsDigit(p.current) {
			value += string(p.current)
			p.Advance()
		}
	}
	
	node := CreateNode(NodeLiteral, start, p.CurrentPosition())
	node.Value = value
	node.Attrs["type"] = "number"
	return node, nil
}

// Parse string literal
func (p *JavaScriptParser) parseStringLiteral() (*Node, error) {
	start := p.CurrentPosition()
	quote := p.current
	p.Advance() // Skip opening quote
	
	value := ""
	for p.pos < len(p.source) && p.current != quote {
		// Handle escape sequences
		if p.current == '\\' {
			p.Advance() // Skip backslash
			switch p.current {
			case 'n':
				value += "\n"
			case 'r':
				value += "\r"
			case 't':
				value += "\t"
			case '\\':
				value += "\\"
			case '"':
				value += "\""
			case '\'':
				value += "'"
			case '`':
				value += "`"
			default:
				value += string(p.current)
			}
		} else {
			value += string(p.current)
		}
		p.Advance()
	}
	
	if p.current != quote {
		return nil, fmt.Errorf("unterminated string at %d:%d", start.Line, start.Column)
	}
	p.Advance() // Skip closing quote
	
	node := CreateNode(NodeLiteral, start, p.CurrentPosition())
	node.Value = value
	node.Attrs["type"] = "string"
	return node, nil
}

// Parse call expression: identifier(args)
func (p *JavaScriptParser) parseCallExpression(callee *Node) (*Node, error) {
	start := callee.Start
	
	// Parse arguments
	args, err := p.parseArguments()
	if err != nil {
		return nil, err
	}
	
	// Create call node
	callNode := CreateNode(NodeCallExpression, start, p.CurrentPosition())
	callNode.Children = append(callNode.Children, callee) // First child is callee
	callNode.Children = append(callNode.Children, args...) // Rest are arguments
	
	return callNode, nil
}

// Parse arguments in a function call
func (p *JavaScriptParser) parseArguments() ([]*Node, error) {
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	p.Advance() // Skip '('
	
	args := []*Node{}
	
	// Handle empty argument list
	p.SkipWhitespace()
	if p.current == ')' {
		p.Advance() // Skip ')'
		return args, nil
	}
	
	// Parse comma-separated arguments
	for {
		p.SkipWhitespace()
		
		// Parse argument expression
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		
		p.SkipWhitespace()
		if p.current == ')' {
			p.Advance() // Skip ')'
			break
		}
		
		if p.current != ',' {
			return nil, fmt.Errorf("expected ',' or ')' at %d:%d", p.line, p.column)
		}
		p.Advance() // Skip ','
	}
	
	return args, nil
}

// Parse function declaration
func (p *JavaScriptParser) parseFunctionDeclaration() (*Node, error) {
	start := p.CurrentPosition()
	
	// Advance past 'function' keyword
	for i := 0; i < 8; i++ {
		p.Advance()
	}
	
	p.SkipWhitespace()
	
	// Parse function name
	nameStart := p.CurrentPosition()
	name := ""
	for p.pos < len(p.source) && ((p.current >= 'a' && p.current <= 'z') || 
								  (p.current >= 'A' && p.current <= 'Z') || 
								  p.current == '_' || 
								  (p.current >= '0' && p.current <= '9' && len(name) > 0)) {
		name += string(p.current)
		p.Advance()
	}
	nameEnd := p.CurrentPosition()
	
	nameNode := CreateNode(NodeIdentifier, nameStart, nameEnd)
	nameNode.Value = name
	
	functionNode := CreateNode(NodeFunctionDecl, start, Position{})
	functionNode.Children = append(functionNode.Children, nameNode)
	
	// Parse parameters
	p.SkipWhitespace()
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	p.Advance()
	
	// Parse parameters
	paramList := []*Node{}
	for p.pos < len(p.source) && p.current != ')' {
		p.SkipWhitespace()
		
		if p.current == ',' {
			p.Advance()
			continue
		}
		
		// Parse parameter name
		paramStart := p.CurrentPosition()
		paramName := ""
		for p.pos < len(p.source) && ((p.current >= 'a' && p.current <= 'z') || 
									  (p.current >= 'A' && p.current <= 'Z') || 
									  p.current == '_' || 
									  (p.current >= '0' && p.current <= '9' && len(paramName) > 0)) {
			paramName += string(p.current)
			p.Advance()
		}
		paramEnd := p.CurrentPosition()
		
		paramNode := CreateNode(NodeParameter, paramStart, paramEnd)
		paramNode.Value = paramName
		paramList = append(paramList, paramNode)
		
		p.SkipWhitespace()
	}
	
	if p.current != ')' {
		return nil, fmt.Errorf("expected ')' at %d:%d", p.line, p.column)
	}
	p.Advance()
	
	// Add parameters to function node
	for _, param := range paramList {
		functionNode.Children = append(functionNode.Children, param)
	}
	
	// Parse function body
	p.SkipWhitespace()
	if p.current != '{' {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.line, p.column)
	}
	
	blockNode, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}
	
	functionNode.Children = append(functionNode.Children, blockNode)
	functionNode.End = p.CurrentPosition()
	
	return functionNode, nil
}

// Parse block statement
func (p *JavaScriptParser) parseBlockStatement() (*Node, error) {
	if p.current != '{' {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.line, p.column)
	}
	
	start := p.CurrentPosition()
	p.Advance() // Skip '{'
	
	node := CreateNode(NodeBlockStatement, start, Position{})
	
	// Parse statements until closing brace
	for p.pos < len(p.source) && p.current != '}' {
		p.SkipWhitespace()
		
		if p.current == '}' {
			break
		}
		
		// Parse statement
		var stmt *Node
		var err error
		
		if p.CheckKeyword("function") {
			stmt, err = p.parseFunctionDeclaration()
		} else if p.CheckKeyword("var") || p.CheckKeyword("let") || p.CheckKeyword("const") {
			stmt, err = p.parseVariableDeclaration()
		} else if p.CheckKeyword("if") {
			stmt, err = p.parseIfStatement()
		} else if p.CheckKeyword("for") {
			stmt, err = p.parseForStatement()
		} else if p.CheckKeyword("while") {
			stmt, err = p.parseWhileStatement()
		} else if p.CheckKeyword("return") {
			stmt, err = p.parseReturnStatement()
		} else {
			stmt, err = p.parseExpressionStatement()
		}
		
		if err != nil {
			return nil, err
		}
		
		node.Children = append(node.Children, stmt)
	}
	
	if p.current != '}' {
		return nil, fmt.Errorf("expected '}' at %d:%d", p.line, p.column)
	}
	p.Advance() // Skip '}'
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse variable declaration
func (p *JavaScriptParser) parseVariableDeclaration() (*Node, error) {
	start := p.CurrentPosition()
	
	// Determine variable kind (var, let, const)
	var kind string
	if p.CheckKeyword("var") {
		kind = "var"
		for i := 0; i < 3; i++ {
			p.Advance()
		}
	} else if p.CheckKeyword("let") {
		kind = "let"
		for i := 0; i < 3; i++ {
			p.Advance()
		}
	} else if p.CheckKeyword("const") {
		kind = "const"
		for i := 0; i < 5; i++ {
			p.Advance()
		}
	}
	
	node := CreateNode(NodeVariableDecl, start, Position{})
	node.Attrs["kind"] = kind
	
	p.SkipWhitespace()
	
	// Parse variable name (identifier)
	nameStart := p.CurrentPosition()
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	nameEnd := p.CurrentPosition()
	
	identNode := CreateNode(NodeIdentifier, nameStart, nameEnd)
	identNode.Value = name
	node.Children = append(node.Children, identNode)
	
	// Parse initializer if present
	p.SkipWhitespace()
	if p.current == '=' {
		p.Advance() // Skip '='
		p.SkipWhitespace()
		
		initExpr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, initExpr)
	}
	
	// Skip to end of declaration (;)
	p.SkipWhitespace()
	if p.current == ';' {
		p.Advance()
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse if statement
func (p *JavaScriptParser) parseIfStatement() (*Node, error) {
	start := p.CurrentPosition()
	node := CreateNode(NodeIfStatement, start, Position{})
	
	// Skip 'if' keyword
	for i := 0; i < 2; i++ {
		p.Advance()
	}
	
	p.SkipWhitespace()
	
	// Parse condition
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	p.Advance() // Skip '('
	
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.Children = append(node.Children, condition)
	
	p.SkipWhitespace()
	if p.current != ')' {
		return nil, fmt.Errorf("expected ')' at %d:%d", p.line, p.column)
	}
	p.Advance() // Skip ')'
	
	p.SkipWhitespace()
	
	// Parse the if body
	var thenBody *Node
	if p.current == '{' {
		thenBody, err = p.parseBlockStatement()
	} else {
		// Single statement body
		thenBody, err = p.parseStatement()
	}
	
	if err != nil {
		return nil, err
	}
	node.Children = append(node.Children, thenBody)
	
	// Check for else clause
	p.SkipWhitespace()
	if p.CheckKeyword("else") {
		// Skip 'else' keyword
		for i := 0; i < 4; i++ {
			p.Advance()
		}
		
		p.SkipWhitespace()
		
		// Parse the else body
		var elseBody *Node
		if p.current == '{' {
			elseBody, err = p.parseBlockStatement()
		} else if p.CheckKeyword("if") {
			// Handle else if
			elseBody, err = p.parseIfStatement()
		} else {
			// Single statement body
			elseBody, err = p.parseStatement()
		}
		
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, elseBody)
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse for statement
func (p *JavaScriptParser) parseForStatement() (*Node, error) {
	start := p.CurrentPosition()
	node := CreateNode(NodeForStatement, start, Position{})
	
	// Skip 'for' keyword
	for i := 0; i < 3; i++ {
		p.Advance()
	}
	
	p.SkipWhitespace()
	
	// Parse for loop header: for (init; cond; update)
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	p.Advance() // Skip '('
	
	// Parse initialization
	p.SkipWhitespace()
	if p.current != ';' {
		// Has initialization
		var init *Node
		var err error
		
		if p.CheckKeyword("var") || p.CheckKeyword("let") || p.CheckKeyword("const") {
			init, err = p.parseVariableDeclaration()
		} else {
			init, err = p.parseExpression()
			
			// Skip to semicolon
			p.SkipWhitespace()
			if p.current == ';' {
				p.Advance()
			} else {
				return nil, fmt.Errorf("expected ';' at %d:%d", p.line, p.column)
			}
		}
		
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, init)
	} else {
		// Empty initialization
		p.Advance() // Skip ';'
		node.Children = append(node.Children, nil)
	}
	
	// Parse condition
	p.SkipWhitespace()
	if p.current != ';' {
		// Has condition
		condition, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, condition)
	} else {
		// Empty condition
		node.Children = append(node.Children, nil)
	}
	
	// Skip semicolon
	p.SkipWhitespace()
	if p.current != ';' {
		return nil, fmt.Errorf("expected ';' at %d:%d", p.line, p.column)
	}
	p.Advance()
	
	// Parse update
	p.SkipWhitespace()
	if p.current != ')' {
		// Has update
		update, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, update)
	} else {
		// Empty update
		node.Children = append(node.Children, nil)
	}
	
	// Skip closing parenthesis
	p.SkipWhitespace()
	if p.current != ')' {
		return nil, fmt.Errorf("expected ')' at %d:%d", p.line, p.column)
	}
	p.Advance()
	
	// Parse loop body
	p.SkipWhitespace()
	var body *Node
	var err error
	
	if p.current == '{' {
		body, err = p.parseBlockStatement()
	} else {
		// Single statement body
		body, err = p.parseStatement()
	}
	
	if err != nil {
		return nil, err
	}
	node.Children = append(node.Children, body)
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse while statement
func (p *JavaScriptParser) parseWhileStatement() (*Node, error) {
	start := p.CurrentPosition()
	node := CreateNode(NodeWhileStatement, start, Position{})
	
	// Skip 'while' keyword
	for i := 0; i < 5; i++ {
		p.Advance()
	}
	
	p.SkipWhitespace()
	
	// Parse condition
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	p.Advance() // Skip '('
	
	condition, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.Children = append(node.Children, condition)
	
	p.SkipWhitespace()
	if p.current != ')' {
		return nil, fmt.Errorf("expected ')' at %d:%d", p.line, p.column)
	}
	p.Advance() // Skip ')'
	
	// Parse loop body
	p.SkipWhitespace()
	var body *Node
	
	if p.current == '{' {
		body, err = p.parseBlockStatement()
	} else {
		// Single statement body
		body, err = p.parseStatement()
	}
	
	if err != nil {
		return nil, err
	}
	node.Children = append(node.Children, body)
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse return statement
func (p *JavaScriptParser) parseReturnStatement() (*Node, error) {
	start := p.CurrentPosition()
	node := CreateNode(NodeReturnStatement, start, Position{})
	
	// Skip 'return' keyword
	for i := 0; i < 6; i++ {
		p.Advance()
	}
	
	p.SkipWhitespace()
	
	// Parse return value (if any)
	if p.current != ';' && p.current != '}' && p.current != '\n' {
		returnValue, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, returnValue)
	}
	
	// Skip semicolon if present
	p.SkipWhitespace()
	if p.current == ';' {
		p.Advance()
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse expression statement
func (p *JavaScriptParser) parseExpressionStatement() (*Node, error) {
	start := p.CurrentPosition()
	node := CreateNode(NodeExpressionStatement, start, Position{})
	
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	node.Children = append(node.Children, expr)
	
	// Skip semicolon if present
	p.SkipWhitespace()
	if p.current == ';' {
		p.Advance()
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse class declaration
func (p *JavaScriptParser) parseClassDeclaration() (*Node, error) {
	start := p.CurrentPosition()
	node := CreateNode(NodeClassDeclaration, start, Position{})
	
	// Skip 'class' keyword
	for i := 0; i < 5; i++ {
		p.Advance()
	}
	
	p.SkipWhitespace()
	
	// Parse class name
	nameStart := p.CurrentPosition()
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	nameEnd := p.CurrentPosition()
	
	nameNode := CreateNode(NodeIdentifier, nameStart, nameEnd)
	nameNode.Value = name
	node.Children = append(node.Children, nameNode)
	
	p.SkipWhitespace()
	
	// Check for extends clause
	if p.CheckKeyword("extends") {
		// Skip 'extends' keyword
		for i := 0; i < 7; i++ {
			p.Advance()
		}
		
		p.SkipWhitespace()
		
		// Parse superclass name
		superStart := p.CurrentPosition()
		superName, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		superEnd := p.CurrentPosition()
		
		superNode := CreateNode(NodeIdentifier, superStart, superEnd)
		superNode.Value = superName
		superNode.Attrs["role"] = "superclass"
		node.Children = append(node.Children, superNode)
	}
	
	p.SkipWhitespace()
	
	// Parse class body
	if p.current != '{' {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.line, p.column)
	}
	
	// Simple class body parsing - skip for now
	p.Advance() // Skip '{'
	
	braceCount := 1
	for p.pos < len(p.source) && braceCount > 0 {
		if p.current == '{' {
			braceCount++
		} else if p.current == '}' {
			braceCount--
			if braceCount == 0 {
				p.Advance() // Skip closing brace
				break
			}
		}
		p.Advance()
	}
	
	if braceCount > 0 {
		return nil, fmt.Errorf("unclosed class body at %d:%d", p.line, p.column)
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse import statement
func (p *JavaScriptParser) parseImportStatement() (*Node, error) {
	start := p.CurrentPosition()
	node := CreateNode(NodeImportStatement, start, Position{})
	
	// Skip 'import' keyword
	for i := 0; i < 6; i++ {
		p.Advance()
	}
	
	// Skip to the end of the import statement
	for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
		p.Advance()
	}
	if p.current == ';' {
		p.Advance()
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse export statement
func (p *JavaScriptParser) parseExportStatement() (*Node, error) {
	start := p.CurrentPosition()
	node := CreateNode(NodeExportStatement, start, Position{})
	
	// Skip 'export' keyword
	for i := 0; i < 6; i++ {
		p.Advance()
	}
	
	// Skip to the end of the export statement
	for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
		p.Advance()
	}
	if p.current == ';' {
		p.Advance()
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}