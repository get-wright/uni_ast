package ast

// Position represents a position in the source code
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Offset int `json:"offset"` // Character offset in the source
	File   string `json:"file,omitempty"` // File name, useful for multi-file parsing
}

// Range represents a range in the source code
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Node represents a node in the AST with enhanced position tracking
type Node struct {
	Type     string            `json:"type"`
	Value    string            `json:"value,omitempty"`
	Children []*Node           `json:"children,omitempty"`
	Attrs    map[string]string `json:"attrs,omitempty"`
	Range    Range             `json:"range"` // Using a Range instead of separate Start/End
	
	// Cached token information for quick access
	FirstToken *Token `json:"-"` // The leftmost token of this node
	LastToken  *Token `json:"-"` // The rightmost token of this node
}

// Token represents a token with origin information
type Token struct {
	Kind     string   `json:"kind"`
	Value    string   `json:"value"`
	Position Position `json:"position"`
	Origin   TokenOrigin `json:"origin"`
}

// TokenOrigin represents the origin of a token
type TokenOrigin int

const (
	OriginalToken TokenOrigin = iota // From the source code
	FakeToken                        // Generated during parsing
	RecoveryToken                    // Generated during error recovery
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

// Helper to create nodes with proper range tracking
func CreateNode(nodeType string, start, end Position) *Node {
	return &Node{
		Type:     nodeType,
		Range:    Range{Start: start, End: end},
		Children: []*Node{},
		Attrs:    make(map[string]string),
	}
}

// SetRange sets the range of a node based on first and last tokens
func (n *Node) SetRange(first, last *Token) {
	n.Range.Start = first.Position
	n.Range.End = last.Position
	n.FirstToken = first
	n.LastToken = last
}

// ComputeRangeFromChildren computes the range based on the node's children
func (n *Node) ComputeRangeFromChildren() {
	if len(n.Children) == 0 {
		return
	}
	
	// Start with the range of the first child
	first := n.Children[0]
	last := n.Children[len(n.Children)-1]
	
	n.Range.Start = first.Range.Start
	n.Range.End = last.Range.End
}

// CombineRanges combines the ranges of multiple nodes
func CombineRanges(nodes []*Node) Range {
	if len(nodes) == 0 {
		return Range{}
	}
	
	result := nodes[0].Range
	
	for _, node := range nodes[1:] {
		// Update end position if this node ends after the current end
		if node.Range.End.Offset > result.End.Offset {
			result.End = node.Range.End
		}
		
		// Update start position if this node starts before the current start
		if node.Range.Start.Offset < result.Start.Offset {
			result.Start = node.Range.Start
		}
	}
	
	return result
}