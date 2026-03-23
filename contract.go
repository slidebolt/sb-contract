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
	ID              string   `json:"id"`
	Kind            string   `json:"kind"`
	ContractVersion int      `json:"contractVersion"`
	DependsOn       []string `json:"dependsOn,omitempty"`
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
