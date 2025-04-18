package parser

import (
	"fmt"
	"universal-parser/ast"
)

// TokenType represents the type of a token
type TokenType int

const (
	TokenIdentifier TokenType = iota
	TokenKeyword
	TokenOperator
	TokenString
	TokenNumber
	TokenComment
	TokenWhitespace
	TokenEOF
	TokenInvalid
	// Add more token types as needed
)

// TokenInfo stores detailed information about a token
type TokenInfo struct {
	Type     TokenType
	Value    string
	Line     int
	Column   int
	Offset   int
	Length   int
	Origin   ast.TokenOrigin
	File     string // Source file name
}

// BaseParser implements common functionality for all parsers with enhanced token tracking
type BaseParser struct {
	source     string
	filename   string
	pos        int
	line       int
	column     int
	current    byte
	keywords   map[string]bool // Language-specific keywords
	
	// Token handling enhancements
	tokens      []TokenInfo    // Pre-tokenized input (if available)
	tokenIndex  int            // Current token index
	tokenCache  map[int]ast.Token // Cache of ast.Token objects
	
	// Error recovery
	recoverMode bool
	errors      []ParserError
	
	// For handling included files
	fileResolver FileResolver
}

// ParserError represents a parsing error with location information
type ParserError struct {
	Message  string
	Position ast.Position
}

// FileResolver is an interface for resolving file inclusions
type FileResolver interface {
	ResolveFile(name string) (string, error)
}

// Initialize sets up the parser with the given source and filename
func (p *BaseParser) Initialize(source, filename string) {
	p.source = source
	p.filename = filename
	p.pos = 0
	p.line = 1
	p.column = 1
	p.keywords = make(map[string]bool)
	p.tokenCache = make(map[int]ast.Token)
	
	if len(source) > 0 {
		p.current = source[0]
	}
}

// CreateToken creates a token with the current position information
func (p *BaseParser) CreateToken(value string, origin ast.TokenOrigin) ast.Token {
	return ast.Token{
		Kind:  "token", // Can be more specific based on context
		Value: value,
		Position: ast.Position{
			Line:   p.line,
			Column: p.column,
			Offset: p.pos,
			File:   p.filename,
		},
		Origin: origin,
	}
}

// CreateFakeToken creates a fake token for synthetic nodes
func (p *BaseParser) CreateFakeToken(value string) ast.Token {
	return p.CreateToken(value, ast.FakeToken)
}

// Advance moves to the next character with enhanced tracking
func (p *BaseParser) Advance() {
	if p.pos < len(p.source) {
		// Check for newline to properly track line and column
		if p.current == '\n' {
			p.line++
			p.column = 1
		} else {
			p.column++
		}
		p.pos++
	}
	
	if p.pos >= len(p.source) {
		p.current = 0
		return
	}
	
	p.current = p.source[p.pos]
}

// Peek returns the next character without advancing
func (p *BaseParser) Peek() byte {
	if p.pos+1 >= len(p.source) {
		return 0
	}
	return p.source[p.pos+1]
}

// PeekAhead returns the character at the specified offset without advancing
func (p *BaseParser) PeekAhead(offset int) byte {
	if p.pos+offset >= len(p.source) {
		return 0
	}
	return p.source[p.pos+offset]
}

// PeekRange returns a range of characters ahead without advancing
func (p *BaseParser) PeekRange(start, length int) string {
	if p.pos+start >= len(p.source) {
		return ""
	}
	
	end := p.pos + start + length
	if end > len(p.source) {
		end = len(p.source)
	}
	
	return p.source[p.pos+start:end]
}

// SkipWhitespace skips spaces, tabs, newlines, and carriage returns
func (p *BaseParser) SkipWhitespace() {
	for p.pos < len(p.source) && (p.current == ' ' || p.current == '\t' || p.current == '\n' || p.current == '\r') {
		p.Advance()
	}
}

// SkipLineComment skips a line comment (// in many languages)
func (p *BaseParser) SkipLineComment() {
	for p.pos < len(p.source) && p.current != '\n' {
		p.Advance()
	}
	if p.pos < len(p.source) {
		p.Advance() // Skip the newline
	}
}

// SkipBlockComment skips a block comment (/* ... */ in many languages)
func (p *BaseParser) SkipBlockComment(startSequence, endSequence string) {
	// Skip the opening sequence
	for i := 0; i < len(startSequence); i++ {
		p.Advance()
	}
	
	// Look for the closing sequence
	for p.pos < len(p.source) {
		if p.current == endSequence[0] {
			matched := true
			for i := 1; i < len(endSequence); i++ {
				if p.PeekAhead(i) != endSequence[i] {
					matched = false
					break
				}
			}
			
			if matched {
				// Skip the closing sequence
				for i := 0; i < len(endSequence); i++ {
					p.Advance()
				}
				return
			}
		}
		p.Advance()
	}
}

// CurrentPosition returns the current position as an ast.Position
func (p *BaseParser) CurrentPosition() ast.Position {
	return ast.Position{
		Line:   p.line,
		Column: p.column,
		Offset: p.pos,
		File:   p.filename,
	}
}

// CheckKeyword checks if the current position contains the given keyword
func (p *BaseParser) CheckKeyword(keyword string) bool {
	if p.pos+len(keyword) > len(p.source) {
		return false
	}
	
	for i := 0; i < len(keyword); i++ {
		if p.source[p.pos+i] != keyword[i] {
			return false
		}
	}
	
	// Make sure it's a complete keyword
	if p.pos+len(keyword) < len(p.source) {
		next := p.source[p.pos+len(keyword)]
		if (next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') || next == '_' || (next >= '0' && next <= '9') {
			return false
		}
	}
	
	return true
}

// IsAlpha checks if the character is a letter
func (p *BaseParser) IsAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// IsDigit checks if the character is a digit
func (p *BaseParser) IsDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// IsAlphaNumeric checks if the character is a letter or digit
func (p *BaseParser) IsAlphaNumeric(ch byte) bool {
	return p.IsAlpha(ch) || p.IsDigit(ch)
}

// IsIdentifierStart checks if the character can start an identifier
func (p *BaseParser) IsIdentifierStart(ch byte) bool {
	return p.IsAlpha(ch)
}

// IsIdentifierPart checks if the character can be part of an identifier
func (p *BaseParser) IsIdentifierPart(ch byte) bool {
	return p.IsAlphaNumeric(ch)
}

// ParseIdentifier parses an identifier
func (p *BaseParser) ParseIdentifier() (string, error) {
	if !p.IsIdentifierStart(p.current) {
		return "", fmt.Errorf("expected identifier at %d:%d, got '%c'", p.line, p.column, p.current)
	}
	
	identifier := string(p.current)
	p.Advance()
	
	for p.pos < len(p.source) && p.IsIdentifierPart(p.current) {
		identifier += string(p.current)
		p.Advance()
	}
	
	return identifier, nil
}

// ConsumeToken consumes a token and returns its value and position
func (p *BaseParser) ConsumeToken(tokenType TokenType) (ast.Token, error) {
	startPos := p.CurrentPosition()
	startIndex := p.pos
	
	switch tokenType {
	case TokenIdentifier:
		if !p.IsIdentifierStart(p.current) {
			return ast.Token{}, fmt.Errorf("expected identifier at %d:%d", p.line, p.column)
		}
		
		identifier := string(p.current)
		p.Advance()
		
		for p.pos < len(p.source) && p.IsIdentifierPart(p.current) {
			identifier += string(p.current)
			p.Advance()
		}
		
		token := ast.Token{
			Kind:     "identifier",
			Value:    identifier,
			Position: startPos,
			Origin:   ast.OriginalToken,
		}
		
		p.tokenCache[startIndex] = token
		return token, nil
		
	case TokenString:
		// Handle string literals
		// ...
		
	case TokenNumber:
		// Handle number literals
		// ...
	}
	
	return ast.Token{}, fmt.Errorf("unsupported token type %v", tokenType)
}

// EnterErrorRecoveryMode enables error recovery mode
func (p *BaseParser) EnterErrorRecoveryMode() {
	p.recoverMode = true
}

// ExitErrorRecoveryMode disables error recovery mode
func (p *BaseParser) ExitErrorRecoveryMode() {
	p.recoverMode = false
}

// RecordError records a parsing error
func (p *BaseParser) RecordError(message string) {
	p.errors = append(p.errors, ParserError{
		Message:  message,
		Position: p.CurrentPosition(),
	})
	
	// If we're in recovery mode, try to recover
	if p.recoverMode {
		p.TryRecover()
	}
}

// TryRecover attempts to recover from a parsing error
func (p *BaseParser) TryRecover() {
	// Skip until we find a synchronization point (e.g., ';', '}', etc.)
	for p.pos < len(p.source) && p.current != ';' && p.current != '}' && p.current != '\n' {
		p.Advance()
	}
	
	// Skip the synchronization token
	if p.pos < len(p.source) {
		p.Advance()
	}
}

// GetErrors returns all parsing errors
func (p *BaseParser) GetErrors() []ParserError {
	return p.errors
}

// SkipWhitespaceAndComments skips spaces, tabs, newlines, and comments
func (p *BaseParser) SkipWhitespaceAndComments() {
	for p.pos < len(p.source) {
		// Skip whitespace
		if p.current == ' ' || p.current == '\t' || p.current == '\n' || p.current == '\r' {
			p.Advance()
			continue
		}
		
		// Check for line comment
		if p.current == '/' && p.PeekAhead(1) == '/' {
			p.SkipLineComment()
			continue
		}
		
		// Check for block comment
		if p.current == '/' && p.PeekAhead(1) == '*' {
			p.SkipBlockComment("/*", "*/")
			continue
		}
		
		// No more whitespace or comments
		break
	}
}

// CreateNodeWithTokens creates a node with tokens for range information
func (p *BaseParser) CreateNodeWithTokens(nodeType string, firstToken, lastToken ast.Token) *ast.Node {
	node := ast.CreateNode(nodeType, firstToken.Position, lastToken.Position)
	node.FirstToken = &firstToken
	node.LastToken = &lastToken
	return node
}

// IncludeFile processes an include directive and updates the parser state
func (p *BaseParser) IncludeFile(filename string) error {
	if p.fileResolver == nil {
		return fmt.Errorf("no file resolver available for including %s", filename)
	}
	
	content, err := p.fileResolver.ResolveFile(filename)
	if err != nil {
		return err
	}
	
	// Save current parser state
	oldSource := p.source
	oldPos := p.pos
	oldLine := p.line
	oldColumn := p.column
	oldCurrent := p.current
	oldFilename := p.filename
	
	// Set new parser state
	p.Initialize(content, filename)
	
	// Parse the included file
	// This would typically call a method to parse the file contents
	
	// Restore original parser state
	p.source = oldSource
	p.pos = oldPos
	p.line = oldLine
	p.column = oldColumn
	p.current = oldCurrent
	p.filename = oldFilename
	
	return nil
}