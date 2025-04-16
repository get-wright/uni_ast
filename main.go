package main

import (
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
	outputFormat := flag.String("format", "pretty", "Output format (pretty, json)")
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
	
	// Parse the source
	ast, err := parser.Parse(string(source))
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
		// TODO: Implement JSON output
		fmt.Println("JSON output not yet implemented")
	default:
		fmt.Printf("Unknown output format: %s\n", *outputFormat)
		os.Exit(1)
	}
}