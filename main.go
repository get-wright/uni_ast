// main.go

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"universal-parser/ast"
	"universal-parser/parser"
)

// FileSystem implements the parser.FileResolver interface for local files
type FileSystem struct {
	RootDir string
}

func (fs *FileSystem) ResolveFile(name string) (string, error) {
	path := filepath.Join(fs.RootDir, name)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func main() {
	// Parse command line arguments
	var (
		filename     = flag.String("file", "", "Source file to parse")
		language     = flag.String("lang", "", "Language to use (autodetected from extension if not provided)")
		outputFormat = flag.String("format", "json", "Output format (json, text)")
		includeComments = flag.Bool("comments", true, "Include comments in AST")
		includePositions = flag.Bool("positions", true, "Include position information in AST")
		recover     = flag.Bool("recover", true, "Recover from parse errors")
	)
	
	flag.Parse()
	
	if *filename == "" {
		fmt.Println("Error: No input file specified")
		flag.Usage()
		os.Exit(1)
	}
	
	// Read the file
	content, err := ioutil.ReadFile(*filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}
	
	// Set up parser configuration
	config := parser.ParserConfig{
		IncludeComments:   *includeComments,
		RecoverFromErrors: *recover,
		PreservePositions: *includePositions,
		MaxErrors:         100,
		FileResolver:      &FileSystem{RootDir: filepath.Dir(*filename)},
	}
	
	// Parse the file
	var root *ast.Node
	var parseErr error
	
	if *language != "" {
		// Use specified language
		p, err := parser.DefaultRegistry.GetByLanguage(*language, &config)
		if err != nil {
			fmt.Printf("Error getting parser: %v\n", err)
			os.Exit(1)
		}
		
		root, parseErr = p.Parse(string(content))
	} else {
		// Auto-detect language from extension
		root, parseErr = parser.ParseFile(*filename, string(content), &config)
	}
	
	// Check for parse errors
	if parseErr != nil {
		fmt.Printf("Warning: Parsing completed with errors: %v\n", parseErr)
	}
	
	if root == nil {
		fmt.Println("Error: Failed to parse file")
		os.Exit(1)
	}
	
	// Output the AST
	switch *outputFormat {
	case "json":
		outputJSON(root, *includePositions)
	case "text":
		outputText(root, 0)
	default:
		fmt.Printf("Unknown output format: %s\n", *outputFormat)
		os.Exit(1)
	}
	
	// If there were parse errors, exit with non-zero status
	if parseErr != nil {
		os.Exit(1)
	}
}

// outputJSON outputs the AST as JSON
func outputJSON(node *ast.Node, includePositions bool) {
	// Create a simplified view for JSON output
	type SimpleNode struct {
		Type     string                 `json:"type"`
		Value    string                 `json:"value,omitempty"`
		Children []SimpleNode           `json:"children,omitempty"`
		Attrs    map[string]string      `json:"attrs,omitempty"`
		Range    *ast.Range             `json:"range,omitempty"`
	}
	
	var simplify func(*ast.Node) SimpleNode
	simplify = func(node *ast.Node) SimpleNode {
		result := SimpleNode{
			Type:  node.Type,
			Value: node.Value,
			Attrs: node.Attrs,
		}
		
		if includePositions {
			result.Range = &node.Range
		}
		
		// Recursively simplify children
		for _, child := range node.Children {
			result.Children = append(result.Children, simplify(child))
		}
		
		return result
	}
	
	// Convert the AST to a simplified form
	simple := simplify(node)
	
	// Output as JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(simple)
}

// outputText outputs the AST as a hierarchical text format
func outputText(node *ast.Node, indent int) {
	// Print node type and value
	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}
	
	fmt.Printf("%s- %s", indentStr, node.Type)
	
	if node.Value != "" {
		fmt.Printf(" (%s)", node.Value)
	}
	
	// Print position information
	fmt.Printf(" [%d:%d - %d:%d]\n", 
		node.Range.Start.Line, node.Range.Start.Column,
		node.Range.End.Line, node.Range.End.Column)
	
	// Print attributes if any
	for key, value := range node.Attrs {
		fmt.Printf("%s  %s: %s\n", indentStr, key, value)
	}
	
	// Recursively print children
	for _, child := range node.Children {
		outputText(child, indent+1)
	}
}