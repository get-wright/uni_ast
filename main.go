package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Parse command line arguments
	filePath := flag.String("file", "", "Path to the source file to parse")
	language := flag.String("lang", "", "Language of the source file (js, py, go, java, c, cpp, ts)")
	outputFormat := flag.String("format", "pretty", "Output format (pretty, json, symbols)")
	useCache := flag.Bool("cache", true, "Use parsing cache for better performance")
	strictMode := flag.Bool("strict", false, "Enable strict mode parsing")
	ecmaVersion := flag.Int("ecma", 2020, "ECMAScript version for JavaScript parsing")
	
	flag.Parse()
	
	if *filePath == "" {
		fmt.Println("Usage: parser -file <file_path> [-lang <language>] [-format <format>]")
		flag.PrintDefaults()
		os.Exit(1)
	}
	
	// Read the source file
	source, err := os.ReadFile(*filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}
	
	// Infer language from file extension if not provided
	if *language == "" {
		ext := strings.ToLower(filepath.Ext(*filePath))
		switch ext {
		case ".js":
			*language = "js"
		case ".py":
			*language = "py"
		case ".go":
			*language = "go"
		case ".java":
			*language = "java"
		case ".c":
			*language = "c"
		case ".cpp", ".cc", ".cxx":
			*language = "cpp"
		case ".ts":
			*language = "ts"
		default:
			fmt.Println("Could not infer language from file extension. Please specify using -lang flag.")
			os.Exit(1)
		}
	}
	
	// Create the appropriate parser
	parser, err := ParserFactory(*language)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	// Configure parser
	parser.SetOptions(ParserOptions{
		StrictMode:       *strictMode,
		TargetECMAScript: *ecmaVersion,
		SourcePath:       *filePath,
		IncludeComments:  true,
	})
	
	// Parse the source
	var ast *Node
	if *useCache {
		ast, err = ParseWithCache(parser, string(source))
	} else {
		ast, err = parser.Parse(string(source))
	}
	
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		os.Exit(1)
	}
	
	// Output the AST based on format
	switch strings.ToLower(*outputFormat) {
	case "pretty":
		fmt.Printf("AST for %s file:\n", parser.GetLanguage())
		PrintAST(ast, "")
	case "json":
		jsonData, err := ast.ToJSON()
		if err != nil {
			fmt.Printf("Error converting to JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	case "symbols":
		// Output symbol table
		printSymbolTable(ast)
	default:
		fmt.Printf("Unknown output format: %s\n", *outputFormat)
		os.Exit(1)
	}
}

// Print symbol table
func printSymbolTable(node *Node) {
	fmt.Println("Symbol Table:")
	
	// Collect symbols from nodes
	var symbols []*Symbol
	visitor := &symbolCollector{symbols: &symbols}
	node.Accept(visitor)
	
	// Print collected symbols
	for _, sym := range symbols {
		if sym == nil || sym.Definition == nil {
			continue
		}
		
		fmt.Printf("- %s (%s): defined at line %d, column %d\n",
			sym.Name, sym.Kind, sym.Definition.Start.Line, sym.Definition.Start.Column)
		
		if len(sym.References) > 0 {
			fmt.Printf("  References: ")
			for i, ref := range sym.References {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("line %d, column %d", ref.Start.Line, ref.Start.Column)
			}
			fmt.Println()
		}
	}
}

// Symbol collector visitor
type symbolCollector struct {
	symbols *[]*Symbol
}

func (s *symbolCollector) Visit(node *Node) Visitor {
	if node.Symbol != nil && node.Symbol.Definition == node {
		*s.symbols = append(*s.symbols, node.Symbol)
	}
	return s
}