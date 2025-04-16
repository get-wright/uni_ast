package main

import (
	"fmt"
	"strings"
)

// Parser interface for language-specific parsers
type Parser interface {
	Parse(source string) (*Node, error)
	GetLanguage() string
}

// ParserFactory creates a parser for a specific language
func ParserFactory(language string) (Parser, error) {
	switch strings.ToLower(language) {
	case "js", "javascript":
		return &JavaScriptParser{}, nil
	case "py", "python":
		return &PythonParser{}, nil
	case "go", "golang":
		return &GoParser{}, nil
	case "java":
		return &JavaParser{}, nil
	case "c":
		return &CParser{}, nil
	case "cpp", "c++":
		return &CPPParser{}, nil
	case "ts", "typescript":
		return &TypeScriptParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

// BasicParser implements common functionality for all parsers
type BasicParser struct {
	source   string
	pos      int
	line     int
	column   int
	current  byte
	keywords map[string]bool // Language-specific keywords
}

// Initialize sets up the parser with the given source
func (p *BasicParser) Initialize(source string) {
	p.source = source
	p.pos = 0
	p.line = 1
	p.column = 1
	if len(source) > 0 {
		p.current = source[0]
	}
}

// Advance moves to the next character
func (p *BasicParser) Advance() {
	p.pos++
	p.column++
	
	if p.pos >= len(p.source) {
		p.current = 0
		return
	}
	
	p.current = p.source[p.pos]
	
	if p.current == '\n' {
		p.line++
		p.column = 1
	}
}

// Peek returns the next character without advancing
func (p *BasicParser) Peek() byte {
	if p.pos+1 >= len(p.source) {
		return 0
	}
	return p.source[p.pos+1]
}

// PeekAhead returns the character at the specified offset without advancing
func (p *BasicParser) PeekAhead(offset int) byte {
	if p.pos+offset >= len(p.source) {
		return 0
	}
	return p.source[p.pos+offset]
}

// Skip whitespace
func (p *BasicParser) SkipWhitespace() {
	for p.pos < len(p.source) && (p.current == ' ' || p.current == '\t' || p.current == '\n' || p.current == '\r') {
		p.Advance()
	}
}

// SkipLineComment skips a line comment (// in many languages)
func (p *BasicParser) SkipLineComment() {
	for p.pos < len(p.source) && p.current != '\n' {
		p.Advance()
	}
	if p.pos < len(p.source) {
		p.Advance() // Skip the newline
	}
}

// SkipBlockComment skips a block comment (/* ... */ in many languages)
func (p *BasicParser) SkipBlockComment(startSequence, endSequence string) {
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

// Current position
func (p *BasicParser) CurrentPosition() Position {
	return Position{Line: p.line, Column: p.column}
}

// CheckKeyword checks if the current position contains the given keyword
func (p *BasicParser) CheckKeyword(keyword string) bool {
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
func (p *BasicParser) IsAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// IsDigit checks if the character is a digit
func (p *BasicParser) IsDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// IsAlphaNumeric checks if the character is a letter or digit
func (p *BasicParser) IsAlphaNumeric(ch byte) bool {
	return p.IsAlpha(ch) || p.IsDigit(ch)
}

// IsIdentifierStart checks if the character can start an identifier
func (p *BasicParser) IsIdentifierStart(ch byte) bool {
	return p.IsAlpha(ch)
}

// IsIdentifierPart checks if the character can be part of an identifier
func (p *BasicParser) IsIdentifierPart(ch byte) bool {
	return p.IsAlphaNumeric(ch)
}

// ParseIdentifier parses an identifier
func (p *BasicParser) ParseIdentifier() (string, error) {
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