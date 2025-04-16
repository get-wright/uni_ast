package main

import (
	"fmt"
)

// Node types
const (
	NodeProgram           = "Program"
	NodeVariableDecl      = "VariableDeclaration"
	NodeFunctionDecl      = "FunctionDeclaration"
	NodeParameter         = "Parameter"
	NodeReturnStatement   = "ReturnStatement"
	NodeCallExpression    = "CallExpression"
	NodeIfStatement       = "IfStatement"
	NodeForStatement      = "ForStatement"
	NodeWhileStatement    = "WhileStatement"
	NodeBlockStatement    = "BlockStatement"
	NodeBinaryExpression  = "BinaryExpression"
	NodeIdentifier        = "Identifier"
	NodeLiteral           = "Literal"
	NodeAssignmentExpr    = "AssignmentExpression"
	NodeClassDeclaration  = "ClassDeclaration"
	NodePropertyDefinition = "PropertyDefinition"
	NodeMethodDefinition  = "MethodDefinition"
	NodeImportStatement   = "ImportStatement"
	NodeExportStatement   = "ExportStatement"
	NodeExpressionStatement = "ExpressionStatement"
	NodeArrayLiteral      = "ArrayLiteral"
	NodeObjectLiteral     = "ObjectLiteral"
	NodeProperty          = "Property"
	NodeArrowFunction     = "ArrowFunction"
	NodeTemplateLiteral   = "TemplateLiteral"
	NodeTryStatement      = "TryStatement"
	NodeCatchClause       = "CatchClause"
	NodeSwitchStatement   = "SwitchStatement"
	NodeSwitchCase        = "SwitchCase"
	NodeThisExpression    = "ThisExpression"
	NodeSuperExpression   = "SuperExpression"
	NodeSpreadElement     = "SpreadElement"
	NodeRestElement       = "RestElement"
	NodeDestructuring     = "Destructuring"
)

// Position represents a position in the source code
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Node represents a node in the AST
type Node struct {
	Type     string            `json:"type"`
	Value    string            `json:"value,omitempty"`
	Children []*Node           `json:"children,omitempty"`
	Attrs    map[string]string `json:"attrs,omitempty"`
	Start    Position          `json:"start"`
	End      Position          `json:"end"`
}

// Helper to create nodes
func CreateNode(nodeType string, start, end Position) *Node {
	return &Node{
		Type:     nodeType,
		Start:    start,
		End:      end,
		Children: []*Node{},
		Attrs:    make(map[string]string),
	}
}

// AST Printer for visualization
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