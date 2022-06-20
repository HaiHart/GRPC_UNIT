package utils

import (
	"fmt"
	"strings"
)

// NodeType represents flag indicating node type (Gateway, Relay, etc.)
type NodeType int

const (
	// Gateway collects all the various gateway types
	Gateway NodeType = 1 << iota
)

var nodeTypeNames = map[NodeType]string{
	Gateway: "GATEWAY",
}

var nodeNameTypes = map[string]NodeType{
	"GATEWAY": Gateway,
}

// String returns the string representation of a node type for use (e.g. in JSON dumps)
func (n NodeType) String() string {
	s, ok := nodeTypeNames[n]
	if ok {
		return s
	}
	return "UNKNOWN"
}

// FromStringToNodeType return nodeType of string name
func FromStringToNodeType(s string) (NodeType, error) {
	cs := strings.Replace(s, "-", "", -1)
	cs = strings.ToUpper(cs)
	nt, ok := nodeNameTypes[cs]
	if ok {
		return nt, nil
	}
	return 0, fmt.Errorf("could not deserialize unknown node value %v", cs)
}
