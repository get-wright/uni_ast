// js_parser.go

package parser

import (
	"fmt"
	"universal-parser/ast"
)

// JavaScriptParser implements the Parser interface for JavaScript with enhanced token tracking
type JavaScriptParser struct {
	BaseParser
	config ParserConfig
}

// NewJavaScriptParser creates a new JavaScript parser with the given configuration
func NewJavaScriptParser(config ParserConfig) Parser {
	return &JavaScriptParser{config: config}
}

// GetLanguage returns the name of the language
func (p *JavaScriptParser) GetLanguage() string {
	return "JavaScript"
}

// Initialize sets up the JavaScript parser
func (p *JavaScriptParser) Initialize(source, filename string) {
	p.BaseParser.Initialize(source, filename)
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
	
	// Add file resolver if provided
	if p.config.FileResolver != nil {
		p.fileResolver = p.config.FileResolver
	}
}

// Parse parses the source code and returns an AST with enhanced token tracking
func (p *JavaScriptParser) Parse(source string) (*ast.Node, error) {
	p.Initialize(source, p.filename)
	
	// Enable error recovery if configured
	if p.config.RecoverFromErrors {
		p.EnterErrorRecoveryMode()
	}
	
	// Create the program node
	startToken := p.CreateToken("program", ast.OriginalToken)
	programNode := p.CreateNodeWithTokens(ast.NodeProgram, startToken, startToken)
	
	// Parse statements
	for p.pos < len(p.source) {
		p.SkipWhitespaceAndComments()
		if p.pos >= len(p.source) {
			break
		}
		
		// Skip comments if not configured to include them
		if p.current == '/' && p.PeekAhead(1) == '/' {
			if p.config.IncludeComments {
				comment, err := p.parseLineComment()
				if err == nil {
					programNode.Children = append(programNode.Children, comment)
				}
			} else {
				p.SkipLineComment()
			}
			continue
		}
		
		if p.current == '/' && p.PeekAhead(1) == '*' {
			if p.config.IncludeComments {
				comment, err := p.parseBlockComment()
				if err == nil {
					programNode.Children = append(programNode.Children, comment)
				}
			} else {
				p.SkipBlockComment("/*", "*/")
			}
			continue
		}
		
		// Parse statements with enhanced error handling
		statement, err := p.parseStatement()
		if err != nil {
			// Record the error
			p.RecordError(err.Error())
			
			// If in recovery mode, we can continue parsing
			if !p.recoverMode {
				return nil, err
			}
			
			// Skip to the next statement
			continue
		}
		
		if statement != nil {
			programNode.Children = append(programNode.Children, statement)
		}
	}
	
	// Set the end position of the program node
	if len(programNode.Children) > 0 {
		lastChild := programNode.Children[len(programNode.Children)-1]
		programNode.SetRange(
			&startToken, 
			lastChild.LastToken,
		)
	} else {
		endToken := p.CreateToken("end", ast.OriginalToken)
		programNode.SetRange(&startToken, &endToken)
	}
	
	// Check for errors
	if len(p.errors) > 0 && !p.recoverMode {
		return programNode, fmt.Errorf("parsing completed with %d errors", len(p.errors))
	}
	
	return programNode, nil
}

// Parse a statement with enhanced token tracking
func (p *JavaScriptParser) parseStatement() (*ast.Node, error) {
	// Skip whitespace and comments
	p.SkipWhitespaceAndComments()
	
	startPos := p.CurrentPosition()
	startToken := p.CreateToken("", ast.OriginalToken)
	
	// Check for keywords and patterns
	if p.CheckKeyword("function") {
		// Parse function declaration
		functionNode, err := p.parseFunctionDeclaration()
		if err != nil {
			return nil, err
		}
		return functionNode, nil
	} else if p.CheckKeyword("var") || p.CheckKeyword("let") || p.CheckKeyword("const") {
		// Parse variable declaration
		varNode, err := p.parseVariableDeclaration()
		if err != nil {
			return nil, err
		}
		return varNode, nil
	} else if p.CheckKeyword("if") {
		// Parse if statement
		ifNode, err := p.parseIfStatement()
		if err != nil {
			return nil, err
		}
		return ifNode, nil
	} else if p.CheckKeyword("for") {
		// Parse for statement
		forNode, err := p.parseForStatement()
		if err != nil {
			return nil, err
		}
		return forNode, nil
	} else if p.CheckKeyword("while") {
		// Parse while statement
		whileNode, err := p.parseWhileStatement()
		if err != nil {
			return nil, err
		}
		return whileNode, nil
	} else if p.CheckKeyword("do") {
		// Parse do-while statement
		doWhileNode, err := p.parseDoWhileStatement()
		if err != nil {
			return nil, err
		}
		return doWhileNode, nil
	} else if p.CheckKeyword("class") {
		// Parse class declaration
		classNode, err := p.parseClassDeclaration()
		if err != nil {
			return nil, err
		}
		return classNode, nil
	} else if p.CheckKeyword("import") {
		// Parse import statement
		importNode, err := p.parseImportStatement()
		if err != nil {
			return nil, err
		}
		return importNode, nil
	} else if p.CheckKeyword("export") {
		// Parse export statement
		exportNode, err := p.parseExportStatement()
		if err != nil {
			return nil, err
		}
		return exportNode, nil
	} else if p.CheckKeyword("return") {
		// Parse return statement
		returnNode, err := p.parseReturnStatement()
		if err != nil {
			return nil, err
		}
		return returnNode, nil
	} else if p.CheckKeyword("switch") {
		// Parse switch statement
		switchNode, err := p.parseSwitchStatement()
		if err != nil {
			return nil, err
		}
		return switchNode, nil
	} else if p.CheckKeyword("try") {
		// Parse try statement
		tryNode, err := p.parseTryStatement()
		if err != nil {
			return nil, err
		}
		return tryNode, nil
	} else if p.CheckKeyword("break") {
		// Parse break statement
		breakNode, err := p.parseBreakStatement()
		if err != nil {
			return nil, err
		}
		return breakNode, nil
	} else if p.CheckKeyword("continue") {
		// Parse continue statement
		continueNode, err := p.parseContinueStatement()
		if err != nil {
			return nil, err
		}
		return continueNode, nil
	} else if p.CheckKeyword("throw") {
		// Parse throw statement
		throwNode, err := p.parseThrowStatement()
		if err != nil {
			return nil, err
		}
		return throwNode, nil
	} else if p.current == '{' {
		// Parse block statement
		blockNode, err := p.parseBlockStatement()
		if err != nil {
			return nil, err
		}
		return blockNode, nil
	} else {
		// Try to parse as expression statement
		exprNode, err := p.parseExpressionStatement()
		if err != nil {
			return nil, err
		}
		return exprNode, nil
	}
}

// Parse a function declaration with enhanced token tracking
func (p *JavaScriptParser) parseFunctionDeclaration() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	
	// Create token for 'function' keyword
	functionToken := p.CreateToken("function", ast.OriginalToken)
	
	// Advance past 'function' keyword
	for i := 0; i < 8; i++ {
		p.Advance()
	}
	
	p.SkipWhitespaceAndComments()
	
	// Parse function name
	nameStartPos := p.CurrentPosition()
	nameToken, err := p.ConsumeToken(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	
	// Create the function node with tokens
	functionNode := p.CreateNodeWithTokens(ast.NodeFunctionDecl, functionToken, nameToken)
	
	// Create the name node with tokens
	nameNode := p.CreateNodeWithTokens(ast.NodeIdentifier, nameToken, nameToken)
	nameNode.Value = nameToken.Value
	
	functionNode.Children = append(functionNode.Children, nameNode)
	
	p.SkipWhitespaceAndComments()
	
	// Parse parameters
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	
	// Token for opening parenthesis
	openParenToken := p.CreateToken("(", ast.OriginalToken)
	p.Advance()
	
	// Parse parameters
	paramList := []*ast.Node{}
	
	for p.pos < len(p.source) && p.current != ')' {
		p.SkipWhitespaceAndComments()
		
		if p.current == ',' {
			p.Advance()
			continue
		}
		
		// Parse parameter name
		paramStartPos := p.CurrentPosition()
		paramToken, err := p.ConsumeToken(TokenIdentifier)
		if err != nil {
			return nil, err
		}
		
		// Create parameter node with tokens
		paramNode := p.CreateNodeWithTokens(ast.NodeParameter, paramToken, paramToken)
		paramNode.Value = paramToken.Value
		
		paramList = append(paramList, paramNode)
		
		p.SkipWhitespaceAndComments()
	}
	
	if p.current != ')' {
		return nil, fmt.Errorf("expected ')' at %d:%d", p.line, p.column)
	}
	
	// Token for closing parenthesis
	closeParenToken := p.CreateToken(")", ast.OriginalToken)
	p.Advance()
	
	// Add parameters to function node
	for _, param := range paramList {
		functionNode.Children = append(functionNode.Children, param)
	}
	
	p.SkipWhitespaceAndComments()
	
	// Parse function body
	if p.current != '{' {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.line, p.column)
	}
	
	// Token for opening brace
	openBraceToken := p.CreateToken("{", ast.OriginalToken)
	p.Advance()
	
	// Create block statement node
	blockNode := p.CreateNodeWithTokens(ast.NodeBlockStatement, openBraceToken, openBraceToken)
	
	// Parse function body (collect statements)
	braceCount := 1
	
	// Parse body statements
	for p.pos < len(p.source) && braceCount > 0 {
		if p.current == '{' {
			p.Advance()
			braceCount++
			continue
		}
		
		if p.current == '}' {
			if braceCount == 1 {
				// Last closing brace
				closeBraceToken := p.CreateToken("}", ast.OriginalToken)
				p.Advance()
				
				// Update block node end token
				blockNode.SetRange(&openBraceToken, &closeBraceToken)
				break
			}
			
			p.Advance()
			braceCount--
			continue
		}
		
		// Try to parse a statement
		stmt, err := p.parseStatement()
		if err != nil {
			// In recovery mode, we continue
			if p.recoverMode {
				p.RecordError(err.Error())
				p.TryRecover()
				continue
			}
			return nil, err
		}
		
		if stmt != nil {
			blockNode.Children = append(blockNode.Children, stmt)
		}
		
		p.SkipWhitespaceAndComments()
	}
	
	// Add block node to function node
	functionNode.Children = append(functionNode.Children, blockNode)
	
	// Update function node end token
	if blockNode.LastToken != nil {
		functionNode.SetRange(&functionToken, blockNode.LastToken)
	}
	
	return functionNode, nil
}

// Parse variable declaration
func (p *JavaScriptParser) parseVariableDeclaration() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	
	// Determine variable kind (var, let, const)
	var kind string
	var kindToken ast.Token
	
	if p.CheckKeyword("var") {
		kind = "var"
		kindToken = p.CreateToken(kind, ast.OriginalToken)
		for i := 0; i < 3; i++ {
			p.Advance()
		}
	} else if p.CheckKeyword("let") {
		kind = "let"
		kindToken = p.CreateToken(kind, ast.OriginalToken)
		for i := 0; i < 3; i++ {
			p.Advance()
		}
	} else if p.CheckKeyword("const") {
		kind = "const"
		kindToken = p.CreateToken(kind, ast.OriginalToken)
		for i := 0; i < 5; i++ {
			p.Advance()
		}
	}
	
	// Create var declaration node
	node := p.CreateNodeWithTokens(ast.NodeVariableDecl, kindToken, kindToken)
	node.Attrs["kind"] = kind
	
	p.SkipWhitespaceAndComments()
	
	// Parse variable name (identifier)
	nameToken, err := p.ConsumeToken(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	
	// Create identifier node
	identNode := p.CreateNodeWithTokens(ast.NodeIdentifier, nameToken, nameToken)
	identNode.Value = nameToken.Value
	
	node.Children = append(node.Children, identNode)
	
	// Parse initializer if present
	p.SkipWhitespaceAndComments()
	if p.current == '=' {
		// Token for equals
		equalsToken := p.CreateToken("=", ast.OriginalToken)
		p.Advance()
		
		p.SkipWhitespaceAndComments()
		
		// Parse expression (simplified)
		// In a real implementation, call a proper expression parser here
		exprStart := p.CurrentPosition()
		exprToken := p.CreateToken("expr", ast.OriginalToken)
		
		// Skip to the end of the expression
		for p.pos < len(p.source) && p.current != ';' && p.current != '\n' && p.current != ',' {
			p.Advance()
		}
		
		exprEnd := p.CurrentPosition()
		exprNode := p.CreateNodeWithTokens(ast.NodeLiteral, exprToken, exprToken)
		exprNode.Value = "expression" // Simplified
		
		node.Children = append(node.Children, exprNode)
	}
	
	// Skip to the semicolon or EOL
	p.SkipWhitespaceAndComments()
	var endToken ast.Token
	
	if p.current == ';' {
		endToken = p.CreateToken(";", ast.OriginalToken)
		p.Advance()
	} else if p.current == '\n' {
		endToken = p.CreateToken("\n", ast.OriginalToken)
		p.Advance()
	} else {
		// If at EOF or no proper terminator, use the last position
		endToken = p.CreateToken("", ast.OriginalToken)
	}
	
	// Update node's end position
	node.SetRange(&kindToken, &endToken)
	
	return node, nil
}

// Parse if statement
func (p *JavaScriptParser) parseIfStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	ifToken := p.CreateToken("if", ast.OriginalToken)
	
	// Skip 'if' keyword
	for i := 0; i < 2; i++ {
		p.Advance()
	}
	
	// Create if statement node
	node := p.CreateNodeWithTokens(ast.NodeIfStatement, ifToken, ifToken)
	
	p.SkipWhitespaceAndComments()
	
	// Parse condition
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	
	// Token for opening parenthesis
	openParenToken := p.CreateToken("(", ast.OriginalToken)
	p.Advance()
	
	// Parse condition (simplified)
	// In a real implementation, call a proper expression parser here
	condStart := p.CurrentPosition()
	condToken := p.CreateToken("condition", ast.OriginalToken)
	
	// Skip to the closing parenthesis
	parenCount := 1
	for p.pos < len(p.source) && parenCount > 0 {
		if p.current == '(' {
			parenCount++
		} else if p.current == ')' {
			parenCount--
		}
		
		if parenCount > 0 {
			p.Advance()
		}
	}
	
	if p.current != ')' {
		return nil, fmt.Errorf("expected ')' at %d:%d", p.line, p.column)
	}
	
	// Token for closing parenthesis
	closeParenToken := p.CreateToken(")", ast.OriginalToken)
	p.Advance()
	
	// Create condition node
	condNode := p.CreateNodeWithTokens("Condition", condToken, closeParenToken)
	node.Children = append(node.Children, condNode)
	
	p.SkipWhitespaceAndComments()
	
	// Parse if body
	thenStmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	
	node.Children = append(node.Children, thenStmt)
	
	// Check for else
	p.SkipWhitespaceAndComments()
	if p.CheckKeyword("else") {
		elseToken := p.CreateToken("else", ast.OriginalToken)
		
		// Skip 'else' keyword
		for i := 0; i < 4; i++ {
			p.Advance()
		}
		
		p.SkipWhitespaceAndComments()
		
		// Parse else body
		elseStmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		
		node.Children = append(node.Children, elseStmt)
		
		// Update node end position
		node.SetRange(&ifToken, elseStmt.LastToken)
	} else {
		// No else clause, update end position
		node.SetRange(&ifToken, thenStmt.LastToken)
	}
	
	return node, nil
}

// parseForStatement parses a for statement
func (p *JavaScriptParser) parseForStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	forToken := p.CreateToken("for", ast.OriginalToken)
	
	// Skip 'for' keyword
	for i := 0; i < 3; i++ {
		p.Advance()
	}
	
	// Create for statement node
	node := p.CreateNodeWithTokens(ast.NodeForStatement, forToken, forToken)
	
	p.SkipWhitespaceAndComments()
	
	// Parse for loop header
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	
	// Token for opening parenthesis
	openParenToken := p.CreateToken("(", ast.OriginalToken)
	p.Advance()
	
	// Parse initialization (simplified)
	initToken := p.CreateToken("init", ast.OriginalToken)
	initNode := p.CreateNodeWithTokens("Initialization", initToken, initToken)
	
	// Skip to the first semicolon
	for p.pos < len(p.source) && p.current != ';' {
		p.Advance()
	}
	
	if p.current != ';' {
		return nil, fmt.Errorf("expected ';' at %d:%d", p.line, p.column)
	}
	
	firstSemiToken := p.CreateToken(";", ast.OriginalToken)
	p.Advance()
	
	// Parse condition (simplified)
	condToken := p.CreateToken("condition", ast.OriginalToken)
	condNode := p.CreateNodeWithTokens("Condition", condToken, condToken)
	
	// Skip to the second semicolon
	for p.pos < len(p.source) && p.current != ';' {
		p.Advance()
	}
	
	if p.current != ';' {
		return nil, fmt.Errorf("expected ';' at %d:%d", p.line, p.column)
	}
	
	secondSemiToken := p.CreateToken(";", ast.OriginalToken)
	p.Advance()
	
	// Parse update (simplified)
	updateToken := p.CreateToken("update", ast.OriginalToken)
	updateNode := p.CreateNodeWithTokens("Update", updateToken, updateToken)
	
	// Skip to the closing parenthesis
	for p.pos < len(p.source) && p.current != ')' {
		p.Advance()
	}
	
	if p.current != ')' {
		return nil, fmt.Errorf("expected ')' at %d:%d", p.line, p.column)
	}
	
	closeParenToken := p.CreateToken(")", ast.OriginalToken)
	p.Advance()
	
	// Add initialization, condition, and update to the for node
	node.Children = append(node.Children, initNode)
	node.Children = append(node.Children, condNode)
	node.Children = append(node.Children, updateNode)
	
	// Parse for body
	p.SkipWhitespaceAndComments()
	bodyStmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	
	node.Children = append(node.Children, bodyStmt)
	
	// Update node end position
	node.SetRange(&forToken, bodyStmt.LastToken)
	
	return node, nil
}

// parseWhileStatement parses a while statement
func (p *JavaScriptParser) parseWhileStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	whileToken := p.CreateToken("while", ast.OriginalToken)
	
	// Skip 'while' keyword
	for i := 0; i < 5; i++ {
		p.Advance()
	}
	
	// Create while statement node
	node := p.CreateNodeWithTokens(ast.NodeWhileStatement, whileToken, whileToken)
	
	p.SkipWhitespaceAndComments()
	
	// Parse condition
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	
	// Token for opening parenthesis
	openParenToken := p.CreateToken("(", ast.OriginalToken)
	p.Advance()
	
	// Parse condition (simplified)
	condToken := p.CreateToken("condition", ast.OriginalToken)
	
	// Skip to the closing parenthesis
	parenCount := 1
	for p.pos < len(p.source) && parenCount > 0 {
		if p.current == '(' {
			parenCount++
		} else if p.current == ')' {
			parenCount--
		}
		
		if parenCount > 0 {
			p.Advance()
		}
	}
	
	if p.current != ')' {
		return nil, fmt.Errorf("expected ')' at %d:%d", p.line, p.column)
	}
	
	closeParenToken := p.CreateToken(")", ast.OriginalToken)
	p.Advance()
	
	// Create condition node
	condNode := p.CreateNodeWithTokens("Condition", condToken, closeParenToken)
	node.Children = append(node.Children, condNode)
	
	// Parse while body
	p.SkipWhitespaceAndComments()
	bodyStmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	
	node.Children = append(node.Children, bodyStmt)
	
	// Update node end position
	node.SetRange(&whileToken, bodyStmt.LastToken)
	
	return node, nil
}

// parseDoWhileStatement parses a do-while statement
func (p *JavaScriptParser) parseDoWhileStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	doToken := p.CreateToken("do", ast.OriginalToken)
	
	// Skip 'do' keyword
	for i := 0; i < 2; i++ {
		p.Advance()
	}
	
	// Create do-while statement node
	node := p.CreateNodeWithTokens("DoWhileStatement", doToken, doToken)
	
	// Parse do body
	p.SkipWhitespaceAndComments()
	bodyStmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	
	node.Children = append(node.Children, bodyStmt)
	
	// Parse while part
	p.SkipWhitespaceAndComments()
	if !p.CheckKeyword("while") {
		return nil, fmt.Errorf("expected 'while' at %d:%d", p.line, p.column)
	}
	
	whileToken := p.CreateToken("while", ast.OriginalToken)
	
	// Skip 'while' keyword
	for i := 0; i < 5; i++ {
		p.Advance()
	}
	
	p.SkipWhitespaceAndComments()
	
	// Parse condition
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	
	openParenToken := p.CreateToken("(", ast.OriginalToken)
	p.Advance()
	
	// Parse condition (simplified)
	condToken := p.CreateToken("condition", ast.OriginalToken)
	
	// Skip to the closing parenthesis
	parenCount := 1
	for p.pos < len(p.source) && parenCount > 0 {
		if p.current == '(' {
			parenCount++
		} else if p.current == ')' {
			parenCount--
		}
		
		if parenCount > 0 {
			p.Advance()
		}
	}
	
	if p.current != ')' {
		return nil, fmt.Errorf("expected ')' at %d:%d", p.line, p.column)
	}
	
	closeParenToken := p.CreateToken(")", ast.OriginalToken)
	p.Advance()
	
	// Create condition node
	condNode := p.CreateNodeWithTokens("Condition", condToken, closeParenToken)
	node.Children = append(node.Children, condNode)
	
	// Check for semicolon
	p.SkipWhitespaceAndComments()
	var endToken ast.Token
	
	if p.current == ';' {
		endToken = p.CreateToken(";", ast.OriginalToken)
		p.Advance()
	} else {
		// If no semicolon, use the closing parenthesis as the end token
		endToken = closeParenToken
	}
	
	// Update node end position
	node.SetRange(&doToken, &endToken)
	
	return node, nil
}

// parseBlockStatement parses a block statement
func (p *JavaScriptParser) parseBlockStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	openBraceToken := p.CreateToken("{", ast.OriginalToken)
	
	// Skip '{' character
	p.Advance()
	
	// Create block statement node
	node := p.CreateNodeWithTokens(ast.NodeBlockStatement, openBraceToken, openBraceToken)
	
	// Parse statements in the block
	for p.pos < len(p.source) && p.current != '}' {
		p.SkipWhitespaceAndComments()
		
		if p.current == '}' {
			break
		}
		
		// Parse a statement
		stmt, err := p.parseStatement()
		if err != nil {
			// In recovery mode, we continue
			if p.recoverMode {
				p.RecordError(err.Error())
				p.TryRecover()
				continue
			}
			return nil, err
		}
		
		if stmt != nil {
			node.Children = append(node.Children, stmt)
		}
	}
	
	// Check for closing brace
	if p.current != '}' {
		return nil, fmt.Errorf("expected '}' at %d:%d", p.line, p.column)
	}
	
	closeBraceToken := p.CreateToken("}", ast.OriginalToken)
	p.Advance()
	
	// Update node end position
	node.SetRange(&openBraceToken, &closeBraceToken)
	
	return node, nil
}

// parseClassDeclaration parses a class declaration
func (p *JavaScriptParser) parseClassDeclaration() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	classToken := p.CreateToken("class", ast.OriginalToken)
	
	// Skip 'class' keyword
	for i := 0; i < 5; i++ {
		p.Advance()
	}
	
	// Create class declaration node
	node := p.CreateNodeWithTokens(ast.NodeClassDeclaration, classToken, classToken)
	
	p.SkipWhitespaceAndComments()
	
	// Parse class name
	nameToken, err := p.ConsumeToken(TokenIdentifier)
	if err != nil {
		return nil, err
	}
	
	// Create class name node
	nameNode := p.CreateNodeWithTokens(ast.NodeIdentifier, nameToken, nameToken)
	nameNode.Value = nameToken.Value
	
	node.Children = append(node.Children, nameNode)
	
	// Check for extends clause
	p.SkipWhitespaceAndComments()
	if p.CheckKeyword("extends") {
		extendsToken := p.CreateToken("extends", ast.OriginalToken)
		
		// Skip 'extends' keyword
		for i := 0; i < 7; i++ {
			p.Advance()
		}
		
		p.SkipWhitespaceAndComments()
		
		// Parse parent class name
		parentToken, err := p.ConsumeToken(TokenIdentifier)
		if err != nil {
			return nil, err
		}
		
		// Create parent class node
		parentNode := p.CreateNodeWithTokens("ExtendsClause", parentToken, parentToken)
		parentNode.Value = parentToken.Value
		
		node.Children = append(node.Children, parentNode)
	}
	
	// Parse class body
	p.SkipWhitespaceAndComments()
	if p.current != '{' {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.line, p.column)
	}
	
	openBraceToken := p.CreateToken("{", ast.OriginalToken)
	p.Advance()
	
	// Create class body node
	bodyNode := p.CreateNodeWithTokens("ClassBody", openBraceToken, openBraceToken)
	
	// Parse class methods and properties (simplified)
	for p.pos < len(p.source) && p.current != '}' {
		p.SkipWhitespaceAndComments()
		
		if p.current == '}' {
			break
		}
		
		// Check for method or property definition
		// For simplicity, we'll just skip to the next method/property
		
		// Skip to the next method or property
		methodStart := p.CurrentPosition()
		methodToken := p.CreateToken("method", ast.OriginalToken)
		
		// Skip until we find a semicolon, opening brace, or closing brace
		for p.pos < len(p.source) && p.current != ';' && p.current != '{' && p.current != '}' {
			p.Advance()
		}
		
		if p.current == '{' {
			// Method with body
			braceCount := 1
			p.Advance()
			
			for p.pos < len(p.source) && braceCount > 0 {
				if p.current == '{' {
					braceCount++
				} else if p.current == '}' {
					braceCount--
				}
				
				if braceCount > 0 {
					p.Advance()
				}
			}
			
			if p.current == '}' {
				p.Advance()
			}
		} else if p.current == ';' {
			// Property
			p.Advance()
		}
		
		// Create method/property node
		methodEnd := p.CurrentPosition()
		methodNode := p.CreateNodeWithTokens("ClassMember", methodToken, methodToken)
		
		bodyNode.Children = append(bodyNode.Children, methodNode)
	}
	
	// Check for closing brace
	if p.current != '}' {
		return nil, fmt.Errorf("expected '}' at %d:%d", p.line, p.column)
	}
	
	closeBraceToken := p.CreateToken("}", ast.OriginalToken)
	p.Advance()
	
	// Update body node end position
	bodyNode.SetRange(&openBraceToken, &closeBraceToken)
	
	// Add body node to class node
	node.Children = append(node.Children, bodyNode)
	
	// Update class node end position
	node.SetRange(&classToken, &closeBraceToken)
	
	return node, nil
}

// parseImportStatement parses an import statement
func (p *JavaScriptParser) parseImportStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	importToken := p.CreateToken("import", ast.OriginalToken)
	
	// Skip 'import' keyword
	for i := 0; i < 6; i++ {
		p.Advance()
	}
	
	// Create import statement node
	node := p.CreateNodeWithTokens(ast.NodeImportStatement, importToken, importToken)
	
	// Parse import statement (simplified)
	// For simplicity, we'll just skip to the semicolon
	
	// Skip to the semicolon
	for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
		p.Advance()
	}
	
	// Create end token
	var endToken ast.Token
	if p.current == ';' {
		endToken = p.CreateToken(";", ast.OriginalToken)
		p.Advance()
	} else if p.current == '\n' {
		endToken = p.CreateToken("\n", ast.OriginalToken)
		p.Advance()
	} else {
		endToken = p.CreateToken("", ast.OriginalToken)
	}
	
	// Update node end position
	node.SetRange(&importToken, &endToken)
	
	return node, nil
}

// parseExportStatement parses an export statement
func (p *JavaScriptParser) parseExportStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	exportToken := p.CreateToken("export", ast.OriginalToken)
	
	// Skip 'export' keyword
	for i := 0; i < 6; i++ {
		p.Advance()
	}
	
	// Create export statement node
	node := p.CreateNodeWithTokens(ast.NodeExportStatement, exportToken, exportToken)
	
	// Parse export statement (simplified)
	// For simplicity, we'll just skip to the semicolon
	
	// Skip to the semicolon
	for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
		p.Advance()
	}
	
	// Create end token
	var endToken ast.Token
	if p.current == ';' {
		endToken = p.CreateToken(";", ast.OriginalToken)
		p.Advance()
	} else if p.current == '\n' {
		endToken = p.CreateToken("\n", ast.OriginalToken)
		p.Advance()
	} else {
		endToken = p.CreateToken("", ast.OriginalToken)
	}
	
	// Update node end position
	node.SetRange(&exportToken, &endToken)
	
	return node, nil
}

// parseReturnStatement parses a return statement
func (p *JavaScriptParser) parseReturnStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	returnToken := p.CreateToken("return", ast.OriginalToken)
	
	// Skip 'return' keyword
	for i := 0; i < 6; i++ {
		p.Advance()
	}
	
	// Create return statement node
	node := p.CreateNodeWithTokens(ast.NodeReturnStatement, returnToken, returnToken)
	
	// Check if there's an expression
	p.SkipWhitespaceAndComments()
	if p.current != ';' && p.current != '\n' && p.current != '}' {
		// Parse return value expression (simplified)
		exprToken := p.CreateToken("expr", ast.OriginalToken)
		
		// Skip to the semicolon or end of line
		for p.pos < len(p.source) && p.current != ';' && p.current != '\n' && p.current != '}' {
			p.Advance()
		}
		
		// Create expression node
		exprNode := p.CreateNodeWithTokens("Expression", exprToken, exprToken)
		node.Children = append(node.Children, exprNode)
	}
	
	// Check for semicolon
	var endToken ast.Token
	if p.current == ';' {
		endToken = p.CreateToken(";", ast.OriginalToken)
		p.Advance()
	} else if p.current == '\n' {
		endToken = p.CreateToken("\n", ast.OriginalToken)
		p.Advance()
	} else {
		endToken = p.CreateToken("", ast.OriginalToken)
	}
	
	// Update node end position
	node.SetRange(&returnToken, &endToken)
	
	return node, nil
}

// parseSwitchStatement parses a switch statement
func (p *JavaScriptParser) parseSwitchStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	switchToken := p.CreateToken("switch", ast.OriginalToken)
	
	// Skip 'switch' keyword
	for i := 0; i < 6; i++ {
		p.Advance()
	}
	
	// Create switch statement node
	node := p.CreateNodeWithTokens(ast.NodeSwitchStatement, switchToken, switchToken)
	
	p.SkipWhitespaceAndComments()
	
	// Parse switch discriminant
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	
	openParenToken := p.CreateToken("(", ast.OriginalToken)
	p.Advance()
	
	// Parse discriminant expression (simplified)
	discrToken := p.CreateToken("discriminant", ast.OriginalToken)
	
	// Skip to the closing parenthesis
	parenCount := 1
	for p.pos < len(p.source) && parenCount > 0 {
		if p.current == '(' {
			parenCount++
		} else if p.current == ')' {
			parenCount--
		}
		
		if parenCount > 0 {
			p.Advance()
		}
	}
	
	if p.current != ')' {
		return nil, fmt.Errorf("expected ')' at %d:%d", p.line, p.column)
	}
	
	closeParenToken := p.CreateToken(")", ast.OriginalToken)
	p.Advance()
	
	// Create discriminant node
	discrNode := p.CreateNodeWithTokens("SwitchDiscriminant", discrToken, closeParenToken)
	node.Children = append(node.Children, discrNode)
	
	// Parse switch body
	p.SkipWhitespaceAndComments()
	if p.current != '{' {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.line, p.column)
	}
	
	openBraceToken := p.CreateToken("{", ast.OriginalToken)
	p.Advance()
	
	// Parse case clauses
	for p.pos < len(p.source) && p.current != '}' {
		p.SkipWhitespaceAndComments()
		
		if p.current == '}' {
			break
		}
		
		// Check for case or default
		if p.CheckKeyword("case") {
			caseToken := p.CreateToken("case", ast.OriginalToken)
			
			// Skip 'case' keyword
			for i := 0; i < 4; i++ {
				p.Advance()
			}
			
			// Parse case expression (simplified)
			p.SkipWhitespaceAndComments()
			caseExprToken := p.CreateToken("caseExpr", ast.OriginalToken)
			
			// Skip to the colon
			for p.pos < len(p.source) && p.current != ':' {
				p.Advance()
			}
			
			if p.current != ':' {
				return nil, fmt.Errorf("expected ':' at %d:%d", p.line, p.column)
			}
			
			colonToken := p.CreateToken(":", ast.OriginalToken)
			p.Advance()
			
			// Create case clause node
			caseNode := p.CreateNodeWithTokens(ast.NodeSwitchCase, caseToken, colonToken)
			node.Children = append(node.Children, caseNode)
		} else if p.CheckKeyword("default") {
			defaultToken := p.CreateToken("default", ast.OriginalToken)
			
			// Skip 'default' keyword
			for i := 0; i < 7; i++ {
				p.Advance()
			}
			
			// Skip to the colon
			p.SkipWhitespaceAndComments()
			if p.current != ':' {
				return nil, fmt.Errorf("expected ':' at %d:%d", p.line, p.column)
			}
			
			colonToken := p.CreateToken(":", ast.OriginalToken)
			p.Advance()
			
			// Create default clause node
			defaultNode := p.CreateNodeWithTokens("DefaultCase", defaultToken, colonToken)
			node.Children = append(node.Children, defaultNode)
		} else {
			// Parse case body statements
			stmt, err := p.parseStatement()
			if err != nil {
				// In recovery mode, we continue
				if p.recoverMode {
					p.RecordError(err.Error())
					p.TryRecover()
					continue
				}
				return nil, err
			}
			
			if stmt != nil {
				node.Children = append(node.Children, stmt)
			}
		}
	}
	
	// Check for closing brace
	if p.current != '}' {
		return nil, fmt.Errorf("expected '}' at %d:%d", p.line, p.column)
	}
	
	closeBraceToken := p.CreateToken("}", ast.OriginalToken)
	p.Advance()
	
	// Update node end position
	node.SetRange(&switchToken, &closeBraceToken)
	
	return node, nil
}

// parseTryStatement parses a try statement
func (p *JavaScriptParser) parseTryStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	tryToken := p.CreateToken("try", ast.OriginalToken)
	
	// Skip 'try' keyword
	for i := 0; i < 3; i++ {
		p.Advance()
	}
	
	// Create try statement node
	node := p.CreateNodeWithTokens(ast.NodeTryStatement, tryToken, tryToken)
	
	// Parse try block
	p.SkipWhitespaceAndComments()
	tryBlock, err := p.parseBlockStatement()
	if err != nil {
		return nil, err
	}
	
	node.Children = append(node.Children, tryBlock)
	
	// Parse catch clause if present
	p.SkipWhitespaceAndComments()
	var finalToken ast.Token = *tryBlock.LastToken
	
	if p.CheckKeyword("catch") {
		catchToken := p.CreateToken("catch", ast.OriginalToken)
		
		// Skip 'catch' keyword
		for i := 0; i < 5; i++ {
			p.Advance()
		}
		
		// Create catch clause node
		catchNode := p.CreateNodeWithTokens(ast.NodeCatchClause, catchToken, catchToken)
		
		// Parse catch parameter if present
		p.SkipWhitespaceAndComments()
		if p.current == '(' {
			openParenToken := p.CreateToken("(", ast.OriginalToken)
			p.Advance()
			
			// Parse parameter (simplified)
			p.SkipWhitespaceAndComments()
			paramToken, err := p.ConsumeToken(TokenIdentifier)
			if err != nil {
				return nil, err
			}
			
			// Skip to closing parenthesis
			for p.pos < len(p.source) && p.current != ')' {
				p.Advance()
			}
			
			if p.current != ')' {
				return nil, fmt.Errorf("expected ')' at %d:%d", p.line, p.column)
			}
			
			closeParenToken := p.CreateToken(")", ast.OriginalToken)
			p.Advance()
			
			// Create parameter node
			paramNode := p.CreateNodeWithTokens("CatchParameter", paramToken, closeParenToken)
			catchNode.Children = append(catchNode.Children, paramNode)
		}
		
		// Parse catch block
		p.SkipWhitespaceAndComments()
		catchBlock, err := p.parseBlockStatement()
		if err != nil {
			return nil, err
		}
		
		catchNode.Children = append(catchNode.Children, catchBlock)
		
		// Update catch node end position
		catchNode.SetRange(&catchToken, catchBlock.LastToken)
		
		// Add catch node to try node
		node.Children = append(node.Children, catchNode)
		
		finalToken = *catchBlock.LastToken
	}
	
	// Parse finally clause if present
	p.SkipWhitespaceAndComments()
	if p.CheckKeyword("finally") {
		finallyToken := p.CreateToken("finally", ast.OriginalToken)
		
		// Skip 'finally' keyword
		for i := 0; i < 7; i++ {
			p.Advance()
		}
		
		// Create finally node
		finallyNode := p.CreateNodeWithTokens("Finally", finallyToken, finallyToken)
		
		// Parse finally block
		p.SkipWhitespaceAndComments()
		finallyBlock, err := p.parseBlockStatement()
		if err != nil {
			return nil, err
		}
		
		finallyNode.Children = append(finallyNode.Children, finallyBlock)
		
		// Update finally node end position
		finallyNode.SetRange(&finallyToken, finallyBlock.LastToken)
		
		// Add finally node to try node
		node.Children = append(node.Children, finallyNode)
		
		finalToken = *finallyBlock.LastToken
	}
	
	// Update try node end position
	node.SetRange(&tryToken, &finalToken)
	
	return node, nil
}

// parseBreakStatement parses a break statement
func (p *JavaScriptParser) parseBreakStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	breakToken := p.CreateToken("break", ast.OriginalToken)
	
	// Skip 'break' keyword
	for i := 0; i < 5; i++ {
		p.Advance()
	}
	
	// Create break statement node
	node := p.CreateNodeWithTokens("BreakStatement", breakToken, breakToken)
	
	// Parse label if present
	p.SkipWhitespaceAndComments()
	if p.IsIdentifierStart(p.current) {
		labelToken, err := p.ConsumeToken(TokenIdentifier)
		if err != nil {
			return nil, err
		}
		
		// Create label node
		labelNode := p.CreateNodeWithTokens("Label", labelToken, labelToken)
		labelNode.Value = labelToken.Value
		
		node.Children = append(node.Children, labelNode)
	}
	
	// Check for semicolon
	var endToken ast.Token
	p.SkipWhitespaceAndComments()
	if p.current == ';' {
		endToken = p.CreateToken(";", ast.OriginalToken)
		p.Advance()
	} else if p.current == '\n' {
		endToken = p.CreateToken("\n", ast.OriginalToken)
		p.Advance()
	} else {
		endToken = p.CreateToken("", ast.OriginalToken)
	}
	
	// Update node end position
	node.SetRange(&breakToken, &endToken)
	
	return node, nil
}

// parseContinueStatement parses a continue statement
func (p *JavaScriptParser) parseContinueStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	continueToken := p.CreateToken("continue", ast.OriginalToken)
	
	// Skip 'continue' keyword
	for i := 0; i < 8; i++ {
		p.Advance()
	}
	
	// Create continue statement node
	node := p.CreateNodeWithTokens("ContinueStatement", continueToken, continueToken)
	
	// Parse label if present
	p.SkipWhitespaceAndComments()
	if p.IsIdentifierStart(p.current) {
		labelToken, err := p.ConsumeToken(TokenIdentifier)
		if err != nil {
			return nil, err
		}
		
		// Create label node
		labelNode := p.CreateNodeWithTokens("Label", labelToken, labelToken)
		labelNode.Value = labelToken.Value
		
		node.Children = append(node.Children, labelNode)
	}
	
	// Check for semicolon
	var endToken ast.Token
	p.SkipWhitespaceAndComments()
	if p.current == ';' {
		endToken = p.CreateToken(";", ast.OriginalToken)
		p.Advance()
	} else if p.current == '\n' {
		endToken = p.CreateToken("\n", ast.OriginalToken)
		p.Advance()
	} else {
		endToken = p.CreateToken("", ast.OriginalToken)
	}
	
	// Update node end position
	node.SetRange(&continueToken, &endToken)
	
	return node, nil
}

// parseThrowStatement parses a throw statement
func (p *JavaScriptParser) parseThrowStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	throwToken := p.CreateToken("throw", ast.OriginalToken)
	
	// Skip 'throw' keyword
	for i := 0; i < 5; i++ {
		p.Advance()
	}
	
	// Create throw statement node
	node := p.CreateNodeWithTokens("ThrowStatement", throwToken, throwToken)
	
	// Parse throw expression
	p.SkipWhitespaceAndComments()
	
	// Parse expression (simplified)
	exprToken := p.CreateToken("expr", ast.OriginalToken)
	
	// Skip to the semicolon or end of line
	for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
		p.Advance()
	}
	
	// Create expression node
	exprNode := p.CreateNodeWithTokens("Expression", exprToken, exprToken)
	node.Children = append(node.Children, exprNode)
	
	// Check for semicolon
	var endToken ast.Token
	if p.current == ';' {
		endToken = p.CreateToken(";", ast.OriginalToken)
		p.Advance()
	} else if p.current == '\n' {
		endToken = p.CreateToken("\n", ast.OriginalToken)
		p.Advance()
	} else {
		endToken = p.CreateToken("", ast.OriginalToken)
	}
	
	// Update node end position
	node.SetRange(&throwToken, &endToken)
	
	return node, nil
}

// parseExpressionStatement parses an expression statement
func (p *JavaScriptParser) parseExpressionStatement() (*ast.Node, error) {
	startPos := p.CurrentPosition()
	startToken := p.CreateToken("expr", ast.OriginalToken)
	
	// Create expression statement node
	node := p.CreateNodeWithTokens(ast.NodeExpressionStatement, startToken, startToken)
	
	// Parse expression (simplified)
	// In a real implementation, call a proper expression parser here
	exprStart := p.CurrentPosition()
	exprToken := p.CreateToken("expr", ast.OriginalToken)
	
	// Skip to the end of the expression
	for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
		p.Advance()
	}
	
	// Create expression node
	exprNode := p.CreateNodeWithTokens("Expression", exprToken, exprToken)
	node.Children = append(node.Children, exprNode)
	
	// Check for semicolon
	var endToken ast.Token
	if p.current == ';' {
		endToken = p.CreateToken(";", ast.OriginalToken)
		p.Advance()
	} else if p.current == '\n' {
		endToken = p.CreateToken("\n", ast.OriginalToken)
		p.Advance()
	} else {
		// If at EOF or no proper terminator, use the last position
		endToken = p.CreateToken("", ast.OriginalToken)
	}
	
	// Update node end position
	node.SetRange(&startToken, &endToken)
	
	return node, nil
}

// Parse a line comment with enhanced token tracking
func (p *JavaScriptParser) parseLineComment() (*ast.Node, error) {
	if p.current != '/' || p.PeekAhead(1) != '/' {
		return nil, fmt.Errorf("expected line comment at %d:%d", p.line, p.column)
	}
	
	startPos := p.CurrentPosition()
	startToken := p.CreateToken("//", ast.OriginalToken)
	
	// Skip the //
	p.Advance()
	p.Advance()
	
	commentText := ""
	
	// Collect comment text
	for p.pos < len(p.source) && p.current != '\n' {
		commentText += string(p.current)
		p.Advance()
	}
	
	// Include the newline in the comment
	if p.pos < len(p.source) && p.current == '\n' {
		endToken := p.CreateToken("\n", ast.OriginalToken)
		p.Advance()
		
		// Create comment node
		commentNode := p.CreateNodeWithTokens("Comment", startToken, endToken)
		commentNode.Value = commentText
		commentNode.Attrs["type"] = "line"
		
		return commentNode, nil
	}
	
	// End of file without newline
	endToken := p.CreateToken("", ast.OriginalToken)
	commentNode := p.CreateNodeWithTokens("Comment", startToken, endToken)
	commentNode.Value = commentText
	commentNode.Attrs["type"] = "line"
	
	return commentNode, nil
}

// Parse a block comment with enhanced token tracking
func (p *JavaScriptParser) parseBlockComment() (*ast.Node, error) {
	if p.current != '/' || p.PeekAhead(1) != '*' {
		return nil, fmt.Errorf("expected block comment at %d:%d", p.line, p.column)
	}
	
	startPos := p.CurrentPosition()
	startToken := p.CreateToken("/*", ast.OriginalToken)
	
	// Skip the /*
	p.Advance()
	p.Advance()
	
	commentText := ""
	
	// Collect comment text
	for p.pos < len(p.source) {
		if p.current == '*' && p.PeekAhead(1) == '/' {
			// End of comment
			p.Advance() // Skip *
			p.Advance() // Skip /
			
			endToken := p.CreateToken("*/", ast.OriginalToken)
			
			// Create comment node
			commentNode := p.CreateNodeWithTokens("Comment", startToken, endToken)
			commentNode.Value = commentText
			commentNode.Attrs["type"] = "block"
			
			return commentNode, nil
		}
		
		commentText += string(p.current)
		p.Advance()
	}
	
	// Unterminated comment
	return nil, fmt.Errorf("unterminated block comment at %d:%d", startPos.Line, startPos.Column)
}