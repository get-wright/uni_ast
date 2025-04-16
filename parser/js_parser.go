package parser

import (
	"fmt"

	"universal-parser/ast"
)

// JavaScriptParser implements the Parser interface for JavaScript
type JavaScriptParser struct {
	BaseParser
}

// NewJavaScriptParser creates a new JavaScript parser
func NewJavaScriptParser() Parser {
	return &JavaScriptParser{}
}

// Initialize sets up the JavaScript parser
func (p *JavaScriptParser) Initialize(source string) {
	p.BaseParser.Initialize(source)
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
}

// GetLanguage returns the name of the language
func (p *JavaScriptParser) GetLanguage() string {
	return "JavaScript"
}

// Parse parses the source code and returns an AST
func (p *JavaScriptParser) Parse(source string) (*ast.Node, error) {
	p.Initialize(source)
	
	// Create the program node
	start := p.CurrentPosition()
	programNode := ast.CreateNode(ast.NodeProgram, start, ast.Position{})
	
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
		if p.CheckKeyword("function") {
			functionNode, err := p.parseFunctionDeclaration()
			if err != nil {
				return nil, err
			}
			programNode.Children = append(programNode.Children, functionNode)
		} else if p.CheckKeyword("var") || p.CheckKeyword("let") || p.CheckKeyword("const") {
			varNode, err := p.parseVariableDeclaration()
			if err != nil {
				return nil, err
			}
			programNode.Children = append(programNode.Children, varNode)
		} else if p.CheckKeyword("if") {
			ifNode, err := p.parseIfStatement()
			if err != nil {
				return nil, err
			}
			programNode.Children = append(programNode.Children, ifNode)
		} else if p.CheckKeyword("for") {
			forNode, err := p.parseForStatement()
			if err != nil {
				return nil, err
			}
			programNode.Children = append(programNode.Children, forNode)
		} else if p.CheckKeyword("while") {
			whileNode, err := p.parseWhileStatement()
			if err != nil {
				return nil, err
			}
			programNode.Children = append(programNode.Children, whileNode)
		} else if p.CheckKeyword("class") {
			classNode, err := p.parseClassDeclaration()
			if err != nil {
				return nil, err
			}
			programNode.Children = append(programNode.Children, classNode)
		} else if p.CheckKeyword("import") {
			importNode, err := p.parseImportStatement()
			if err != nil {
				return nil, err
			}
			programNode.Children = append(programNode.Children, importNode)
		} else if p.CheckKeyword("export") {
			exportNode, err := p.parseExportStatement()
			if err != nil {
				return nil, err
			}
			programNode.Children = append(programNode.Children, exportNode)
		} else {
			// Try to parse as expression statement
			exprNode, err := p.parseExpressionStatement()
			if err != nil {
				return nil, err
			}
			programNode.Children = append(programNode.Children, exprNode)
		}
	}
	
	programNode.End = p.CurrentPosition()
	return programNode, nil
}

// Parse function declaration
func (p *JavaScriptParser) parseFunctionDeclaration() (*ast.Node, error) {
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
	
	nameNode := ast.CreateNode(ast.NodeIdentifier, nameStart, nameEnd)
	nameNode.Value = name
	
	functionNode := ast.CreateNode(ast.NodeFunctionDecl, start, ast.Position{})
	functionNode.Children = append(functionNode.Children, nameNode)
	
	// Parse parameters
	p.SkipWhitespace()
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	p.Advance()
	
	// Parse parameters
	paramList := []*ast.Node{}
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
		
		paramNode := ast.CreateNode(ast.NodeParameter, paramStart, paramEnd)
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
	p.Advance()
	
	// Create block statement node
	blockStart := p.CurrentPosition()
	blockNode := ast.CreateNode(ast.NodeBlockStatement, blockStart, ast.Position{})
	
	// Parse function body (simplified for now)
	braceCount := 1
	for p.pos < len(p.source) && braceCount > 0 {
		if p.current == '{' {
			braceCount++
		} else if p.current == '}' {
			braceCount--
		}
		p.Advance()
	}
	
	if braceCount > 0 {
		return nil, fmt.Errorf("unclosed function body at %d:%d", p.line, p.column)
	}
	
	blockNode.End = p.CurrentPosition()
	functionNode.Children = append(functionNode.Children, blockNode)
	functionNode.End = p.CurrentPosition()
	
	return functionNode, nil
}

// Parse variable declaration
func (p *JavaScriptParser) parseVariableDeclaration() (*ast.Node, error) {
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
	
	node := ast.CreateNode(ast.NodeVariableDecl, start, ast.Position{})
	node.Attrs["kind"] = kind
	
	p.SkipWhitespace()
	
	// Parse variable name (identifier)
	nameStart := p.CurrentPosition()
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}
	nameEnd := p.CurrentPosition()
	
	identNode := ast.CreateNode(ast.NodeIdentifier, nameStart, nameEnd)
	identNode.Value = name
	node.Children = append(node.Children, identNode)
	
	// Skip to the end of the declaration
	for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
		p.Advance()
	}
	if p.current == ';' {
		p.Advance()
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse if statement
func (p *JavaScriptParser) parseIfStatement() (*ast.Node, error) {
	start := p.CurrentPosition()
	node := ast.CreateNode(ast.NodeIfStatement, start, ast.Position{})
	
	// Skip 'if' keyword
	for i := 0; i < 2; i++ {
		p.Advance()
	}
	
	p.SkipWhitespace()
	
	// Parse condition (simplified)
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	
	// Skip to the end of the condition
	parenCount := 1
	p.Advance() // Skip the opening paren
	
	for p.pos < len(p.source) && parenCount > 0 {
		if p.current == '(' {
			parenCount++
		} else if p.current == ')' {
			parenCount--
		}
		p.Advance()
	}
	
	if parenCount > 0 {
		return nil, fmt.Errorf("unclosed condition at %d:%d", p.line, p.column)
	}
	
	p.SkipWhitespace()
	
	// Parse the if body (simplified)
	if p.current == '{' {
		braceCount := 1
		p.Advance() // Skip the opening brace
		
		for p.pos < len(p.source) && braceCount > 0 {
			if p.current == '{' {
				braceCount++
			} else if p.current == '}' {
				braceCount--
			}
			p.Advance()
		}
		
		if braceCount > 0 {
			return nil, fmt.Errorf("unclosed if body at %d:%d", p.line, p.column)
		}
	} else {
		// Single statement body
		for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
			p.Advance()
		}
		if p.current == ';' {
			p.Advance()
		}
	}
	
	// Check for else clause (simplified)
	p.SkipWhitespace()
	if p.CheckKeyword("else") {
		// Skip 'else' keyword
		for i := 0; i < 4; i++ {
			p.Advance()
		}
		
		p.SkipWhitespace()
		
		// Parse the else body (simplified)
		if p.current == '{' {
			braceCount := 1
			p.Advance() // Skip the opening brace
			
			for p.pos < len(p.source) && braceCount > 0 {
				if p.current == '{' {
					braceCount++
				} else if p.current == '}' {
					braceCount--
				}
				p.Advance()
			}
			
			if braceCount > 0 {
				return nil, fmt.Errorf("unclosed else body at %d:%d", p.line, p.column)
			}
		} else {
			// Single statement body
			for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
				p.Advance()
			}
			if p.current == ';' {
				p.Advance()
			}
		}
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse for statement
func (p *JavaScriptParser) parseForStatement() (*ast.Node, error) {
	start := p.CurrentPosition()
	node := ast.CreateNode(ast.NodeForStatement, start, ast.Position{})
	
	// Skip 'for' keyword
	for i := 0; i < 3; i++ {
		p.Advance()
	}
	
	p.SkipWhitespace()
	
	// Parse for loop header
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	
	// Skip to the end of the for loop header
	parenCount := 1
	p.Advance() // Skip the opening paren
	
	for p.pos < len(p.source) && parenCount > 0 {
		if p.current == '(' {
			parenCount++
		} else if p.current == ')' {
			parenCount--
		}
		p.Advance()
	}
	
	if parenCount > 0 {
		return nil, fmt.Errorf("unclosed for loop header at %d:%d", p.line, p.column)
	}
	
	p.SkipWhitespace()
	
	// Parse the for loop body
	if p.current == '{' {
		braceCount := 1
		p.Advance() // Skip the opening brace
		
		for p.pos < len(p.source) && braceCount > 0 {
			if p.current == '{' {
				braceCount++
			} else if p.current == '}' {
				braceCount--
			}
			p.Advance()
		}
		
		if braceCount > 0 {
			return nil, fmt.Errorf("unclosed for loop body at %d:%d", p.line, p.column)
		}
	} else {
		// Single statement body
		for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
			p.Advance()
		}
		if p.current == ';' {
			p.Advance()
		}
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse while statement
func (p *JavaScriptParser) parseWhileStatement() (*ast.Node, error) {
	start := p.CurrentPosition()
	node := ast.CreateNode(ast.NodeWhileStatement, start, ast.Position{})
	
	// Skip 'while' keyword
	for i := 0; i < 5; i++ {
		p.Advance()
	}
	
	p.SkipWhitespace()
	
	// Parse condition
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	
	// Skip to the end of the condition
	parenCount := 1
	p.Advance() // Skip the opening paren
	
	for p.pos < len(p.source) && parenCount > 0 {
		if p.current == '(' {
			parenCount++
		} else if p.current == ')' {
			parenCount--
		}
		p.Advance()
	}
	
	if parenCount > 0 {
		return nil, fmt.Errorf("unclosed condition at %d:%d", p.line, p.column)
	}
	
	p.SkipWhitespace()
	
	// Parse the while body
	if p.current == '{' {
		braceCount := 1
		p.Advance() // Skip the opening brace
		
		for p.pos < len(p.source) && braceCount > 0 {
			if p.current == '{' {
				braceCount++
			} else if p.current == '}' {
				braceCount--
			}
			p.Advance()
		}
		
		if braceCount > 0 {
			return nil, fmt.Errorf("unclosed while body at %d:%d", p.line, p.column)
		}
	} else {
		// Single statement body
		for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
			p.Advance()
		}
		if p.current == ';' {
			p.Advance()
		}
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse class declaration
func (p *JavaScriptParser) parseClassDeclaration() (*ast.Node, error) {
	start := p.CurrentPosition()
	node := ast.CreateNode(ast.NodeClassDeclaration, start, ast.Position{})
	
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
	
	nameNode := ast.CreateNode(ast.NodeIdentifier, nameStart, nameEnd)
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
		
		superNode := ast.CreateNode(ast.NodeIdentifier, superStart, superEnd)
		superNode.Value = superName
		superNode.Attrs["role"] = "superclass"
		node.Children = append(node.Children, superNode)
	}
	
	p.SkipWhitespace()
	
	// Parse class body
	if p.current != '{' {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.line, p.column)
	}
	
	braceCount := 1
	p.Advance() // Skip the opening brace
	
	// Simplified parsing of class methods and properties
	for p.pos < len(p.source) && braceCount > 0 {
		if p.current == '{' {
			braceCount++
		} else if p.current == '}' {
			braceCount--
			if braceCount == 0 {
				p.Advance()
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
func (p *JavaScriptParser) parseImportStatement() (*ast.Node, error) {
	start := p.CurrentPosition()
	node := ast.CreateNode(ast.NodeImportStatement, start, ast.Position{})
	
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
func (p *JavaScriptParser) parseExportStatement() (*ast.Node, error) {
	start := p.CurrentPosition()
	node := ast.CreateNode(ast.NodeExportStatement, start, ast.Position{})
	
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

// Parse expression statement
func (p *JavaScriptParser) parseExpressionStatement() (*ast.Node, error) {
	start := p.CurrentPosition()
	node := ast.CreateNode(ast.NodeExpressionStatement, start, ast.Position{})
	
	// Skip to the end of the expression statement
	for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
		p.Advance()
	}
	if p.current == ';' {
		p.Advance()
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse return statement
func (p *JavaScriptParser) parseReturnStatement() (*ast.Node, error) {
	start := p.CurrentPosition()
	node := ast.CreateNode(ast.NodeReturnStatement, start, ast.Position{})
	
	// Skip 'return' keyword
	for i := 0; i < 6; i++ {
		p.Advance()
	}
	
	// Skip to the end of the return statement
	for p.pos < len(p.source) && p.current != ';' && p.current != '\n' {
		p.Advance()
	}
	if p.current == ';' {
		p.Advance()
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse try statement
func (p *JavaScriptParser) parseTryStatement() (*ast.Node, error) {
	start := p.CurrentPosition()
	node := ast.CreateNode(ast.NodeTryStatement, start, ast.Position{})
	
	// Skip 'try' keyword
	for i := 0; i < 3; i++ {
		p.Advance()
	}
	
	p.SkipWhitespace()
	
	// Parse try block
	if p.current != '{' {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.line, p.column)
	}
	
	braceCount := 1
	p.Advance() // Skip the opening brace
	
	for p.pos < len(p.source) && braceCount > 0 {
		if p.current == '{' {
			braceCount++
		} else if p.current == '}' {
			braceCount--
		}
		p.Advance()
	}
	
	if braceCount > 0 {
		return nil, fmt.Errorf("unclosed try block at %d:%d", p.line, p.column)
	}
	
	p.SkipWhitespace()
	
	// Check for catch clause
	if p.CheckKeyword("catch") {
		// Skip 'catch' keyword
		for i := 0; i < 5; i++ {
			p.Advance()
		}
		
		p.SkipWhitespace()
		
		// Skip catch parameter if present
		if p.current == '(' {
			parenCount := 1
			p.Advance() // Skip the opening paren
			
			for p.pos < len(p.source) && parenCount > 0 {
				if p.current == '(' {
					parenCount++
				} else if p.current == ')' {
					parenCount--
				}
				p.Advance()
			}
		}
		
		p.SkipWhitespace()
		
		// Parse catch block
		if p.current != '{' {
			return nil, fmt.Errorf("expected '{' at %d:%d", p.line, p.column)
		}
		
		braceCount = 1
		p.Advance() // Skip the opening brace
		
		for p.pos < len(p.source) && braceCount > 0 {
			if p.current == '{' {
				braceCount++
			} else if p.current == '}' {
				braceCount--
			}
			p.Advance()
		}
		
		if braceCount > 0 {
			return nil, fmt.Errorf("unclosed catch block at %d:%d", p.line, p.column)
		}
	}
	
	p.SkipWhitespace()
	
	// Check for finally clause
	if p.CheckKeyword("finally") {
		// Skip 'finally' keyword
		for i := 0; i < 7; i++ {
			p.Advance()
		}
		
		p.SkipWhitespace()
		
		// Parse finally block
		if p.current != '{' {
			return nil, fmt.Errorf("expected '{' at %d:%d", p.line, p.column)
		}
		
		braceCount = 1
		p.Advance() // Skip the opening brace
		
		for p.pos < len(p.source) && braceCount > 0 {
			if p.current == '{' {
				braceCount++
			} else if p.current == '}' {
				braceCount--
			}
			p.Advance()
		}
		
		if braceCount > 0 {
			return nil, fmt.Errorf("unclosed finally block at %d:%d", p.line, p.column)
		}
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}

// Parse switch statement
func (p *JavaScriptParser) parseSwitchStatement() (*ast.Node, error) {
	start := p.CurrentPosition()
	node := ast.CreateNode(ast.NodeSwitchStatement, start, ast.Position{})
	
	// Skip 'switch' keyword
	for i := 0; i < 6; i++ {
		p.Advance()
	}
	
	p.SkipWhitespace()
	
	// Parse switch discriminant
	if p.current != '(' {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.line, p.column)
	}
	
	parenCount := 1
	p.Advance() // Skip the opening paren
	
	for p.pos < len(p.source) && parenCount > 0 {
		if p.current == '(' {
			parenCount++
		} else if p.current == ')' {
			parenCount--
		}
		p.Advance()
	}
	
	if parenCount > 0 {
		return nil, fmt.Errorf("unclosed switch discriminant at %d:%d", p.line, p.column)
	}
	
	p.SkipWhitespace()
	
	// Parse switch body
	if p.current != '{' {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.line, p.column)
	}
	
	braceCount := 1
	p.Advance() // Skip the opening brace
	
	for p.pos < len(p.source) && braceCount > 0 {
		if p.current == '{' {
			braceCount++
		} else if p.current == '}' {
			braceCount--
		}
		p.Advance()
	}
	
	if braceCount > 0 {
		return nil, fmt.Errorf("unclosed switch body at %d:%d", p.line, p.column)
	}
	
	node.End = p.CurrentPosition()
	return node, nil
}