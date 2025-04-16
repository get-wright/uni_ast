package main

import (
	"encoding/json"
	"fmt"
)

// NodeType enum for better type safety
type NodeType int

const (
	NodeProgram NodeType = iota
	NodeVariableDecl
	NodeFunctionDecl
	NodeParameter
	NodeReturnStatement
	NodeCallExpression
	NodeIfStatement
	NodeForStatement
	NodeWhileStatement
	NodeBlockStatement
	NodeBinaryExpression
	NodeIdentifier
	NodeLiteral
	NodeAssignmentExpr
	NodeClassDeclaration
	NodePropertyDefinition
	NodeMethodDefinition
	NodeImportStatement
	NodeExportStatement
	NodeExpressionStatement
	NodeArrayLiteral
	NodeObjectLiteral
	NodeProperty
	NodeArrowFunction
	NodeTemplateLiteral
	NodeTryStatement
	NodeCatchClause
	NodeSwitchStatement
	NodeSwitchCase
	NodeThisExpression
	NodeSuperExpression
	NodeSpreadElement
	NodeRestElement
	NodeDestructuring
)

// String conversion for NodeType
func (nt NodeType) String() string {
	nodeTypeStrings := [...]string{
		"Program",
		"VariableDeclaration",
		"FunctionDeclaration",
		"Parameter",
		"ReturnStatement",
		"CallExpression",
		"IfStatement",
		"ForStatement",
		"WhileStatement",
		"BlockStatement",
		"BinaryExpression",
		"Identifier",
		"Literal",
		"AssignmentExpression",
		"ClassDeclaration",
		"PropertyDefinition",
		"MethodDefinition",
		"ImportStatement",
		"ExportStatement",
		"ExpressionStatement",
		"ArrayLiteral",
		"ObjectLiteral",
		"Property",
		"ArrowFunction",
		"TemplateLiteral",
		"TryStatement",
		"CatchClause",
		"SwitchStatement",
		"SwitchCase",
		"ThisExpression",
		"SuperExpression",
		"SpreadElement",
		"RestElement",
		"Destructuring",
	}
	
	if int(nt) < len(nodeTypeStrings) {
		return nodeTypeStrings[nt]
	}
	return fmt.Sprintf("Unknown(%d)", nt)
}

// Position represents a position in the source code
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Node represents a node in the AST
type Node struct {
	Type     NodeType          `json:"type"`
	Value    string            `json:"value,omitempty"`
	Children []*Node           `json:"children,omitempty"`
	Attrs    map[string]string `json:"attrs,omitempty"`
	Start    Position          `json:"start"`
	End      Position          `json:"end"`
	
	// Symbol information for semantic analysis
	Symbol *Symbol `json:"-"`
}

// Visitor interface for traversing the AST
type Visitor interface {
	// Visit is called for each node during traversal
	Visit(node *Node) Visitor
}

// Accept implements the visitor pattern
func (n *Node) Accept(v Visitor) {
	if visitor := v.Visit(n); visitor != nil {
		for _, child := range n.Children {
			child.Accept(visitor)
		}
	}
}

// ToJSON serializes the node to JSON
func (n *Node) ToJSON() ([]byte, error) {
	return json.MarshalIndent(n, "", "  ")
}

// FromJSON deserializes a node from JSON
func NodeFromJSON(data []byte) (*Node, error) {
	var node Node
	err := json.Unmarshal(data, &node)
	return &node, err
}

// Helper to create nodes with proper type
func CreateNode(nodeType NodeType, start, end Position) *Node {
	return &Node{
		Type:     nodeType,
		Start:    start,
		End:      end,
		Children: []*Node{},
		Attrs:    make(map[string]string),
	}
}

// Clone creates a deep copy of the node
func (n *Node) Clone() *Node {
	clone := &Node{
		Type:     n.Type,
		Value:    n.Value,
		Start:    n.Start,
		End:      n.End,
		Children: make([]*Node, len(n.Children)),
		Attrs:    make(map[string]string),
	}
	
	for k, v := range n.Attrs {
		clone.Attrs[k] = v
	}
	
	for i, child := range n.Children {
		clone.Children[i] = child.Clone()
	}
	
	return clone
}

// AddChild appends a child node
func (n *Node) AddChild(child *Node) {
	n.Children = append(n.Children, child)
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