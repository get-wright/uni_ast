// token_utils.go

package parser

import (
	"sort"
	"universal-parser/ast"
)

// TokenRange represents a range of tokens in the source
type TokenRange struct {
	Tokens   []ast.Token
	StartPos ast.Position
	EndPos   ast.Position
}

// CollectTokens collects all tokens from an AST node
func CollectTokens(node *ast.Node) []ast.Token {
	var tokens []ast.Token
	
	// Add first and last tokens if available
	if node.FirstToken != nil {
		tokens = append(tokens, *node.FirstToken)
	}
	if node.LastToken != nil && node.LastToken != node.FirstToken {
		tokens = append(tokens, *node.LastToken)
	}
	
	// Recursively collect tokens from children
	for _, child := range node.Children {
		tokens = append(tokens, CollectTokens(child)...)
	}
	
	return tokens
}

// GetNodeRange gets the range of a node, calculating it from children if necessary
func GetNodeRange(node *ast.Node) ast.Range {
	// If range is already set, return it
	if node.Range.Start.Line > 0 || node.Range.End.Line > 0 {
		return node.Range
	}
	
	// If we have first and last tokens, compute from those
	if node.FirstToken != nil && node.LastToken != nil {
		return ast.Range{
			Start: node.FirstToken.Position,
			End:   node.LastToken.Position,
		}
	}
	
	// Compute from children
	if len(node.Children) > 0 {
		firstChild := node.Children[0]
		lastChild := node.Children[len(node.Children)-1]
		
		firstRange := GetNodeRange(firstChild)
		lastRange := GetNodeRange(lastChild)
		
		return ast.Range{
			Start: firstRange.Start,
			End:   lastRange.End,
		}
	}
	
	// Default empty range
	return ast.Range{}
}

// SortTokensByPosition sorts tokens by their position in the source
func SortTokensByPosition(tokens []ast.Token) {
	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i].Position.Line != tokens[j].Position.Line {
			return tokens[i].Position.Line < tokens[j].Position.Line
		}
		return tokens[i].Position.Column < tokens[j].Position.Column
	})
}

// TokenRangeFromNode creates a TokenRange from an AST node
func TokenRangeFromNode(node *ast.Node) TokenRange {
	tokens := CollectTokens(node)
	SortTokensByPosition(tokens)
	
	if len(tokens) == 0 {
		range_ := GetNodeRange(node)
		return TokenRange{
			Tokens:   []ast.Token{},
			StartPos: range_.Start,
			EndPos:   range_.End,
		}
	}
	
	return TokenRange{
		Tokens:   tokens,
		StartPos: tokens[0].Position,
		EndPos:   tokens[len(tokens)-1].Position,
	}
}

// ExtractTextFromRange extracts the text from a range in the source
func ExtractTextFromRange(source string, range_ ast.Range) string {
	// Calculate start and end positions
	startLine := range_.Start.Line
	startCol := range_.Start.Column
	endLine := range_.End.Line
	endCol := range_.End.Column
	
	// Handle simple case: single line
	if startLine == endLine {
		// Convert to byte offsets (this is simplified and may need adjustment for Unicode)
		startOffset := range_.Start.Offset
		endOffset := range_.End.Offset
		
		if startOffset >= 0 && endOffset > startOffset && endOffset <= len(source) {
			return source[startOffset:endOffset]
		}
		return ""
	}
	
	// Multi-line case would require more sophisticated handling
	// Calculate offsets by counting lines and columns
	// ...
	
	return ""
}

// FindNearestNodeToPosition finds the nearest AST node to a given position
func FindNearestNodeToPosition(root *ast.Node, pos ast.Position) *ast.Node {
	// Start with the root node
	closest := root
	minDistance := positionDistance(GetNodeRange(root).Start, pos)
	
	// Helper function to recursively search for the closest node
	var findClosest func(node *ast.Node)
	findClosest = func(node *ast.Node) {
		// Check if this node is closer
		nodeRange := GetNodeRange(node)
		if pointInRange(pos, nodeRange) {
			// The position is inside this node
			
			// Check if any of the children contain the position
			for _, child := range node.Children {
				childRange := GetNodeRange(child)
				if pointInRange(pos, childRange) {
					findClosest(child)
					return
				}
			}
			
			// No child contains the position, but this node does
			closest = node
			return
		}
		
		// Calculate distance to this node
		distance := positionDistance(nodeRange.Start, pos)
		if distance < minDistance {
			minDistance = distance
			closest = node
		}
		
		// Check children
		for _, child := range node.Children {
			findClosest(child)
		}
	}
	
	findClosest(root)
	return closest
}

// Helper function to check if a point is inside a range
func pointInRange(pos ast.Position, range_ ast.Range) bool {
	// Check if the position is after the start
	if pos.Line < range_.Start.Line {
		return false
	}
	if pos.Line == range_.Start.Line && pos.Column < range_.Start.Column {
		return false
	}
	
	// Check if the position is before the end
	if pos.Line > range_.End.Line {
		return false
	}
	if pos.Line == range_.End.Line && pos.Column > range_.End.Column {
		return false
	}
	
	return true
}

// Helper function to calculate distance between two positions
func positionDistance(pos1, pos2 ast.Position) int {
	// Simple line-based distance, could be refined with column information
	lineDiff := abs(pos1.Line - pos2.Line)
	if lineDiff > 0 {
		return lineDiff * 1000 // Weight lines more heavily
	}
	return abs(pos1.Column - pos2.Column)
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}