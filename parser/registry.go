package parser

import (
	"fmt"
	"strings"
)

// ParserRegistry handles parser registration and lookup
type ParserRegistry struct {
	parsers map[string]func() Parser
}

// NewParserRegistry creates a new parser registry
func NewParserRegistry() *ParserRegistry {
	registry := &ParserRegistry{
		parsers: make(map[string]func() Parser),
	}
	
	// Register default parsers - for now only JavaScript is implemented
	registry.Register("js", NewJavaScriptParser)
	registry.Register("javascript", NewJavaScriptParser)
	
	// Other parsers will be registered when implemented
	// For now, we'll use placeholder functions that return appropriate errors
	registry.Register("py", createPlaceholderParser("Python"))
	registry.Register("python", createPlaceholderParser("Python"))
	registry.Register("go", createPlaceholderParser("Go"))
	registry.Register("golang", createPlaceholderParser("Go"))
	registry.Register("java", createPlaceholderParser("Java"))
	registry.Register("c", createPlaceholderParser("C"))
	registry.Register("cpp", createPlaceholderParser("C++"))
	registry.Register("c++", createPlaceholderParser("C++"))
	registry.Register("ts", createPlaceholderParser("TypeScript"))
	registry.Register("typescript", createPlaceholderParser("TypeScript"))
	
	return registry
}

// createPlaceholderParser returns a factory function for a not-yet-implemented parser
func createPlaceholderParser(language string) func() Parser {
	return func() Parser {
		return &placeholderParser{language: language}
	}
}

// placeholderParser is a stub implementation that returns an error when trying to parse
type placeholderParser struct {
	language string
}

func (p *placeholderParser) Parse(source string) (*ast.Node, error) {
	return nil, fmt.Errorf("%s parser not yet implemented", p.language)
}

func (p *placeholderParser) GetLanguage() string {
	return p.language
}

// Register adds a parser factory function to the registry
func (r *ParserRegistry) Register(name string, factory func() Parser) {
	r.parsers[strings.ToLower(name)] = factory
}

// Get returns a parser for the given language
func (r *ParserRegistry) Get(language string) (Parser, error) {
	factory, ok := r.parsers[strings.ToLower(language)]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
	
	return factory(), nil
}

// DefaultRegistry is the global parser registry
var DefaultRegistry = NewParserRegistry()

// GetParser returns a parser for the given language using the default registry
func GetParser(language string) (Parser, error) {
	return DefaultRegistry.Get(language)
}