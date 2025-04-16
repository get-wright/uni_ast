package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
)

// Parser interface for language-specific parsers
type Parser interface {
	// Core parsing methods
	Initialize(source string)
	Parse(source string) (*Node, error)
	GetLanguage() string
	
	// Specialized parsing methods
	ParseExpression(source string) (*Node, error)
	ParseStatement(source string) (*Node, error)
	ParseType(source string) (*Node, error)
	
	// Configuration
	SetOptions(options ParserOptions)
}

// ParserOptions contains configuration options for parsers
type ParserOptions struct {
	StrictMode       bool
	TargetECMAScript int    // For JavaScript/TypeScript
	Dialect          string // For language-specific dialects
	IncludeComments  bool
	SourcePath       string // For import resolution
}

// ParseError represents a detailed parsing error
type ParseError struct {
	Line    int
	Column  int
	Message string
	Context string // Surrounding code snippet
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Parse error at line %d, column %d: %s\n%s", 
		e.Line, e.Column, e.Message, e.Context)
}

// Symbol represents a named entity in the code
type Symbol struct {
	Name      string
	Kind      string // "variable", "function", "class", etc.
	Type      string
	IsExported bool
	References []*Node
	Definition *Node
	Scope      *SymbolTable
}

// SymbolTable for tracking symbols and scopes
type SymbolTable struct {
	Symbols map[string]*Symbol
	Parent  *SymbolTable
	Children []*SymbolTable
	Node     *Node // AST node associated with this scope
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable(parent *SymbolTable) *SymbolTable {
	st := &SymbolTable{
		Symbols:  make(map[string]*Symbol),
		Parent:   parent,
		Children: []*SymbolTable{},
	}
	if parent != nil {
		parent.Children = append(parent.Children, st)
	}
	return st
}

// Add adds a symbol to the table
func (st *SymbolTable) Add(name, kind, symbolType string, node *Node) *Symbol {
	sym := &Symbol{
		Name:       name,
		Kind:       kind,
		Type:       symbolType,
		Definition: node,
		Scope:      st,
	}
	st.Symbols[name] = sym
	if node != nil {
		node.Symbol = sym
	}
	return sym
}

// Lookup finds a symbol in this table or a parent table
func (st *SymbolTable) Lookup(name string) *Symbol {
	if sym, ok := st.Symbols[name]; ok {
		return sym
	}
	if st.Parent != nil {
		return st.Parent.Lookup(name)
	}
	return nil
}

// Parser factory
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
	case "php":
		return &PHPParser{}, nil
	case "ruby":
		return &RubyParser{}, nil
	case "rust":
		return &RustParser{}, nil
	case "cs", "csharp":
		return &CSharpParser{}, nil
	case "swift":
		return &SwiftParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

// ParserCache for improving performance
var (
	parserCache = make(map[string]*Node)
	cacheMutex  sync.RWMutex
)

// ParseWithCache parses with caching
func ParseWithCache(parser Parser, source string) (*Node, error) {
	cacheKey := fmt.Sprintf("%s_%x", parser.GetLanguage(), sha256.Sum256([]byte(source)))
	
	// Check cache first
	cacheMutex.RLock()
	if cachedNode, exists := parserCache[cacheKey]; exists {
		cacheMutex.RUnlock()
		return cachedNode, nil
	}
	cacheMutex.RUnlock()
	
	// Parse and cache the result
	node, err := parser.Parse(source)
	if err == nil {
		cacheMutex.Lock()
		parserCache[cacheKey] = node
		cacheMutex.Unlock()
	}
	return node, err
}

// BasicParser implements common functionality for all parsers
type BasicParser struct {
	source   string
	pos      int
	line     int
	column   int
	current  byte
	keywords map[string]bool // Language-specific keywords
	options  ParserOptions
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
	p.keywords = make(map[string]bool)
}

// SetOptions sets parser configuration options
func (p *BasicParser) SetOptions(options ParserOptions) {
	p.options = options
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

// Default implementations for specialized parsing methods
func (p *BasicParser) ParseExpression(source string) (*Node, error) {
	p.Initialize(source)
	// This should be overridden by language-specific parsers
	return nil, fmt.Errorf("ParseExpression not implemented for this language")
}

func (p *BasicParser) ParseStatement(source string) (*Node, error) {
	p.Initialize(source)
	// This should be overridden by language-specific parsers
	return nil, fmt.Errorf("ParseStatement not implemented for this language")
}

func (p *BasicParser) ParseType(source string) (*Node, error) {
	p.Initialize(source)
	// This should be overridden by language-specific parsers
	return nil, fmt.Errorf("ParseType not implemented for this language")
}

// GetContext extracts a context snippet for error messages
func (p *BasicParser) GetContext(lineNum, col int, contextSize int) string {
	lines := strings.Split(p.source, "\n")
	if lineNum <= 0 || lineNum > len(lines) {
		return ""
	}
	
	start := max(1, lineNum-contextSize)
	end := min(len(lines), lineNum+contextSize)
	
	var result strings.Builder
	for i := start; i <= end; i++ {
		lineStr := lines[i-1]
		result.WriteString(fmt.Sprintf("%4d | %s\n", i, lineStr))
		
		if i == lineNum {
			// Add pointer to the error position
			result.WriteString("     | ")
			for j := 1; j < col; j++ {
				result.WriteString(" ")
			}
			result.WriteString("^\n")
		}
	}
	
	return result.String()
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}