package ast

import (
	"fmt"
)

// PrintAST prints the AST in a hierarchical format for visualization
func PrintAST(node *Node, indent string) {
	fmt.Printf("%s- %s", indent, node.Type)

	if node.Value != "" {
		fmt.Printf(" (%s)", node.Value)
	}

	// Corrected line to access Start and End via Range
	fmt.Printf(" [%d:%d - %d:%d]\n", node.Range.Start.Line, node.Range.Start.Column, node.Range.End.Line, node.Range.End.Column)

	for _, child := range node.Children {
		PrintAST(child, indent+"  ")
	}
}