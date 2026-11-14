// Package contract defines the shared wire types and constants for the
// SlideBolt manager ↔ binary protocol.
package contract

import (
	"encoding/json"
	"fmt"
	"io"
)

// ContractVersion is the current protocol version.
const ContractVersion = 1

// Binary kinds.
const (
	KindService = "service"
	KindPlugin  = "plugin"
)

// Control message types (manager → binary, over stdin).
const (
	ControlShutdown   = "shutdown"
	ControlDependency = "dependency"
)

// Runtime message types (binary → manager, over stdout).
const (
	RuntimeReady = "ready"
	RuntimeError = "error"
	RuntimeLog   = "log"
)

// HelloResponse is the JSON a binary prints to stdout on `<bin> hello`.
type HelloResponse struct {
	ID              string             `json:"id"`
	Kind            string             `json:"kind"`
	ContractVersion int                `json:"contractVersion"`
	DependsOn       []string           `json:"dependsOn,omitempty"`
	Definitions     *PluginDefinitions `json:"definitions,omitempty"`
}

// PluginDefinitions is executable, serialized documentation supplied by a
// plugin. Consumers must render this data generically; plugin IDs are never a
// dispatch mechanism.
type PluginDefinitions struct {
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	DeviceClasses []DeviceClassDefinition `json:"deviceClasses,omitempty"`
	EntityTypes   []string                `json:"entityTypes,omitempty"`
	Commands      []CommandDefinition     `json:"commands,omitempty"`
	Extensions    []ExtensionDefinition   `json:"extensions,omitempty"`
}

type DeviceClassDefinition struct {
	Class       string   `json:"class"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	EntityTypes []string `json:"entityTypes,omitempty"`
}

type ExtensionDefinition struct {
	Type        string              `json:"type"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	StateSchema json.RawMessage     `json:"stateSchema"`
	Commands    []CommandDefinition `json:"commands,omitempty"`
}

type CommandDefinition struct {
	Action        string          `json:"action"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	PayloadSchema json.RawMessage `json:"payloadSchema"`
}

// DescribePlugin builds the common manifest form. It is deliberately free of
// plugin knowledge: each plugin owns the data passed to it.
func DescribePlugin(name, description string, entityTypes ...string) *PluginDefinitions {
	return &PluginDefinitions{Name: name, Description: description, EntityTypes: entityTypes}
}

func DescribeCommand(action, name, description, schema string) CommandDefinition {
	return CommandDefinition{Action: action, Name: name, Description: description, PayloadSchema: json.RawMessage(schema)}
}

// Validate checks that the hello response satisfies the minimum contract.
func (h *HelloResponse) Validate() error {
	if h.ID == "" {
		return fmt.Errorf("hello: id is required")
	}
	if h.Kind != KindService && h.Kind != KindPlugin {
		return fmt.Errorf("hello: kind must be %q or %q, got %q", KindService, KindPlugin, h.Kind)
	}
	if h.ContractVersion != ContractVersion {
		return fmt.Errorf("hello: unsupported contractVersion %d (want %d)", h.ContractVersion, ContractVersion)
	}
	if h.Kind == KindPlugin {
		if h.Definitions == nil {
			return fmt.Errorf("hello: plugin definitions are required")
		}
		if h.Definitions.Name == "" || h.Definitions.Description == "" {
			return fmt.Errorf("hello: plugin definition name and description are required")
		}
		seen := map[string]struct{}{}
		for _, entityType := range h.Definitions.EntityTypes {
			if entityType == "" {
				return fmt.Errorf("hello: empty entity type definition")
			}
			if _, exists := seen[entityType]; exists {
				return fmt.Errorf("hello: duplicate entity type %q", entityType)
			}
			seen[entityType] = struct{}{}
		}
		for _, command := range h.Definitions.Commands {
			if command.Action == "" || command.Name == "" || command.Description == "" || len(command.PayloadSchema) == 0 {
				return fmt.Errorf("hello: incomplete command definition %q", command.Action)
			}
			var schema any
			if err := json.Unmarshal(command.PayloadSchema, &schema); err != nil {
				return fmt.Errorf("hello: command %q schema: %w", command.Action, err)
			}
		}
	}
	return nil
}

// ControlMessage is a JSON line sent from manager to binary over stdin.
type ControlMessage struct {
	Type    string          `json:"type"`
	ID      string          `json:"id,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// RuntimeMessage is a JSON line sent from binary to manager over stdout.
type RuntimeMessage struct {
	Type    string          `json:"type"`
	Level   string          `json:"level,omitempty"`
	Message string          `json:"message,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// WriteJSON encodes v as a single JSON line to w.
func WriteJSON(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

// ReadJSON decodes one JSON line from a decoder into v.
func ReadJSON(dec *json.Decoder, v any) error {
	return dec.Decode(v)
}
