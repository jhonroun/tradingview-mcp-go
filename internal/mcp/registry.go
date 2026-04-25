package mcp

import (
	"encoding/json"
	"fmt"
)

// HandlerFunc is the function signature for an MCP tool handler.
type HandlerFunc func(args json.RawMessage) (interface{}, error)

// ToolDef defines an MCP tool registration.
type ToolDef struct {
	Name        string
	Description string
	Schema      InputSchema
	Handler     HandlerFunc
}

// Registry holds all registered MCP tools in insertion order.
type Registry struct {
	tools map[string]*ToolDef
	order []string
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]*ToolDef)}
}

// Register adds a tool definition to the registry.
func (r *Registry) Register(def ToolDef) {
	r.tools[def.Name] = &def
	r.order = append(r.order, def.Name)
}

// Get returns a tool definition by name.
func (r *Registry) Get(name string) (*ToolDef, bool) {
	def, ok := r.tools[name]
	return def, ok
}

// List returns all tools as MCP Tool descriptors in registration order.
func (r *Registry) List() []Tool {
	out := make([]Tool, 0, len(r.order))
	for _, name := range r.order {
		def := r.tools[name]
		out = append(out, Tool{
			Name:        def.Name,
			Description: def.Description,
			InputSchema: def.Schema,
		})
	}
	return out
}

// Call dispatches a tool call by name.
func (r *Registry) Call(name string, args json.RawMessage) (interface{}, error) {
	def, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return def.Handler(args)
}
