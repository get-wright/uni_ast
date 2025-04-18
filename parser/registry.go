package parser

import (
	"fmt"
	"strings"
	"sync"
	"universal-parser/ast"
)

// Parser interface for language-specific parsers
type Parser interface {
	Parse(source string) (*ast.Node, error)
	GetLanguage() string
}

// ParserConfig contains configuration options for a parser
type ParserConfig struct {
	IncludeComments     bool
	RecoverFromErrors   bool
	TrackWhitespace     bool
	PreservePositions   bool
	MaxErrors           int
	FileResolver        FileResolver
	// Add more configuration options as needed
}

// ParserRegistry handles parser registration and lookup with enhanced features
type ParserRegistry struct {
	parsers     map[string]func(ParserConfig) Parser
	aliases     map[string]string
	extensions  map[string]string
	mutex       sync.RWMutex // For thread safety
	defaultConf ParserConfig
}

// NewParserRegistry creates a new parser registry with default configuration
func NewParserRegistry() *ParserRegistry {
	registry := &ParserRegistry{
		parsers:    make(map[string]func(ParserConfig) Parser),
		aliases:    make(map[string]string),
		extensions: make(map[string]string),
		defaultConf: ParserConfig{
			IncludeComments:   true,
			RecoverFromErrors: true,
			PreservePositions: true,
			MaxErrors:         100,
		},
	}
	
	// Register default parsers - for now only JavaScript is implemented
	registry.Register("js", NewJavaScriptParser)
	registry.RegisterAlias("javascript", "js")
	registry.RegisterFileExtension(".js", "js")
	registry.RegisterFileExtension(".jsx", "js")
	
	// Other parsers will be registered when implemented
	registry.RegisterPlaceholder("py", "Python")
	registry.RegisterPlaceholder("go", "Go")
	registry.RegisterPlaceholder("java", "Java")
	registry.RegisterPlaceholder("c", "C")
	registry.RegisterPlaceholder("cpp", "C++")
	registry.RegisterPlaceholder("ts", "TypeScript")
	
	return registry
}

// Register adds a parser factory function to the registry
func (r *ParserRegistry) Register(name string, factory func(ParserConfig) Parser) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.parsers[strings.ToLower(name)] = factory
}

// RegisterAlias registers an alternative name for a language
func (r *ParserRegistry) RegisterAlias(alias, target string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.aliases[strings.ToLower(alias)] = strings.ToLower(target)
}

// RegisterFileExtension registers a file extension to a language
func (r *ParserRegistry) RegisterFileExtension(ext, language string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.extensions[strings.ToLower(ext)] = strings.ToLower(language)
}

// RegisterPlaceholder registers a placeholder for a not-yet-implemented parser
func (r *ParserRegistry) RegisterPlaceholder(name, language string) {
	r.Register(name, func(conf ParserConfig) Parser {
		return &placeholderParser{language: language}
	})
	
	// Also register common file extensions
	ext := "." + name
	r.RegisterFileExtension(ext, name)
}

// SetDefaultConfig sets the default configuration for all parsers
func (r *ParserRegistry) SetDefaultConfig(config ParserConfig) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.defaultConf = config
}

// GetByLanguage returns a parser for the given language name with custom configuration
func (r *ParserRegistry) GetByLanguage(language string, config *ParserConfig) (Parser, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	language = strings.ToLower(language)
	
	// Check for alias
	if alias, ok := r.aliases[language]; ok {
		language = alias
	}
	
	factory, ok := r.parsers[language]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
	
	// Use default config if none provided
	conf := r.defaultConf
	if config != nil {
		conf = *config
	}
	
	return factory(conf), nil
}

// GetByFilename returns a parser based on the file extension
func (r *ParserRegistry) GetByFilename(filename string, config *ParserConfig) (Parser, error) {
	r.mutex.RLock()
	
	// Extract extension
	ext := strings.ToLower(filename)
	lastDot := strings.LastIndex(ext, ".")
	if lastDot >= 0 {
		ext = ext[lastDot:]
	} else {
		ext = ""
	}
	
	language, ok := r.extensions[ext]
	r.mutex.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
	
	return r.GetByLanguage(language, config)
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

// DefaultRegistry is the global parser registry
var DefaultRegistry = NewParserRegistry()

// GetParser returns a parser for the given language using the default registry
func GetParser(language string) (Parser, error) {
	return DefaultRegistry.GetByLanguage(language, nil)
}

// GetParserForFile returns a parser for the given filename using the default registry
func GetParserForFile(filename string) (Parser, error) {
	return DefaultRegistry.GetByFilename(filename, nil)
}

// ParseFile parses a file with the appropriate parser
func ParseFile(filename, content string, config *ParserConfig) (*ast.Node, error) {
	parser, err := DefaultRegistry.GetByFilename(filename, config)
	if err != nil {
		return nil, err
	}
	
	return parser.Parse(content)
}