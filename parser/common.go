package parser

import (
	"universal-parser/ast"
)

// Parser interface for language-specific parsers
type Parser interface {
	Parse(source string) (*ast.Node, error)
	GetLanguage() string
}