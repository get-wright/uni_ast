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
	
	fmt.Printf(" [%d:%d - %d:%d]\n", node.Start.Line, node.Start.Column, node.End.Line, node.End.Column)
	
	for _, child := range node.Children {
		PrintAST(child, indent+"  ")
	}
}