package contract

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type SetupMode string

const (
	SetupRequired    SetupMode = "required"
	SetupSelectable  SetupMode = "selectable"
	SetupDependency  SetupMode = "dependency"
	SetupDevelopment SetupMode = "development"
)

type SetupControl string

const (
	SetupText        SetupControl = "text"
	SetupSecret      SetupControl = "secret"
	SetupBoolean     SetupControl = "boolean"
	SetupSelect      SetupControl = "select"
	SetupMultiSelect SetupControl = "multiselect"
	SetupFile        SetupControl = "file"
	SetupDirectory   SetupControl = "directory"
)

type SetupOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type SetupField struct {
	Key         string          `json:"key"`
	Environment string          `json:"environment,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Control     SetupControl    `json:"control"`
	Format      string          `json:"format,omitempty"`
	Required    bool            `json:"required,omitempty"`
	Sensitive   bool            `json:"sensitive,omitempty"`
	Advanced    bool            `json:"advanced,omitempty"`
	Default     json.RawMessage `json:"default,omitempty"`
	Options     []SetupOption   `json:"options,omitempty"`
	Generate    string          `json:"generate,omitempty"`
	RelativeTo  string          `json:"relativeTo,omitempty"`
	Export      string          `json:"export,omitempty"`
}

type SetupConstraint struct {
	Kind   string   `json:"kind"`
	Fields []string `json:"fields"`
}

type ToolSource string

const ToolSourceSystem ToolSource = "system"

type ToolKind string

const (
	ToolKindExecutable ToolKind = "executable"
	ToolKindFile       ToolKind = "file"
)

type ToolCheck struct {
	Arguments []string `json:"arguments,omitempty"`
	Contains  []string `json:"contains,omitempty"`
}

type ToolRequirement struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Source      ToolSource  `json:"source"`
	Kind        ToolKind    `json:"kind"`
	Environment string      `json:"environment,omitempty"`
	Default     string      `json:"default,omitempty"`
	Required    bool        `json:"required,omitempty"`
	InstallHint string      `json:"installHint,omitempty"`
	Checks      []ToolCheck `json:"checks,omitempty"`
}

type ExternalResourceKind string

const (
	ExternalResourceService ExternalResourceKind = "service"
	ExternalResourceDevice  ExternalResourceKind = "device"
)

type ResourceActivationMode string

const (
	ResourceActivationAlways     ResourceActivationMode = "always"
	ResourceActivationConfigured ResourceActivationMode = "configured"
)

type ResourceActivation struct {
	Mode   ResourceActivationMode `json:"mode"`
	Fields []string               `json:"fields,omitempty"`
}

// ExternalResourceDefinition declares a non-SlideBolt service or device whose
// protocol and health checks are owned by the declaring component.
type ExternalResourceDefinition struct {
	ID          string               `json:"id"`
	Kind        ExternalResourceKind `json:"kind"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Fields      []string             `json:"fields,omitempty"`
	Activation  ResourceActivation   `json:"activation"`
}

type SetupDefinition struct {
	Mode        SetupMode                    `json:"mode"`
	Category    string                       `json:"category"`
	Fields      []SetupField                 `json:"fields,omitempty"`
	Constraints []SetupConstraint            `json:"constraints,omitempty"`
	Tools       []ToolRequirement            `json:"tools,omitempty"`
	Resources   []ExternalResourceDefinition `json:"resources,omitempty"`
}

func RequiredSetup(category string, fields ...SetupField) *SetupDefinition {
	return &SetupDefinition{Mode: SetupRequired, Category: category, Fields: fields}
}

func SelectableSetup(category string, fields ...SetupField) *SetupDefinition {
	return &SetupDefinition{Mode: SetupSelectable, Category: category, Fields: fields}
}

func DevelopmentSetup(category string) *SetupDefinition {
	return &SetupDefinition{Mode: SetupDevelopment, Category: category}
}

// DependencySetup describes a hidden component that is installed only when a
// selected or required component depends on it.
func DependencySetup(category string, fields ...SetupField) *SetupDefinition {
	return &SetupDefinition{Mode: SetupDependency, Category: category, Fields: fields}
}

var environmentName = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
var externalResourceID = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func (s *SetupDefinition) Validate() error {
	if s == nil {
		return nil
	}
	if s.Mode != SetupRequired && s.Mode != SetupSelectable && s.Mode != SetupDependency && s.Mode != SetupDevelopment {
		return fmt.Errorf("invalid mode %q", s.Mode)
	}
	if s.Category == "" {
		return fmt.Errorf("category is required")
	}
	fields := make(map[string]SetupField, len(s.Fields))
	environments := make(map[string]string, len(s.Fields))
	for _, field := range s.Fields {
		if field.Key == "" || field.Name == "" {
			return fmt.Errorf("field key and name are required")
		}
		if _, exists := fields[field.Key]; exists {
			return fmt.Errorf("duplicate field %q", field.Key)
		}
		fields[field.Key] = field
		if field.Environment != "" {
			if !environmentName.MatchString(field.Environment) {
				return fmt.Errorf("field %q has invalid environment %q", field.Key, field.Environment)
			}
			if owner, exists := environments[field.Environment]; exists {
				return fmt.Errorf("fields %q and %q share environment %q", owner, field.Key, field.Environment)
			}
			environments[field.Environment] = field.Key
		}
		switch field.Control {
		case SetupText, SetupSecret, SetupBoolean, SetupSelect, SetupMultiSelect, SetupFile, SetupDirectory:
		default:
			return fmt.Errorf("field %q has invalid control %q", field.Key, field.Control)
		}
		if field.Control == SetupSecret && !field.Sensitive {
			return fmt.Errorf("secret field %q must be sensitive", field.Key)
		}
		if (field.Control == SetupSelect || field.Control == SetupMultiSelect) && len(field.Options) == 0 {
			return fmt.Errorf("field %q requires options", field.Key)
		}
		seenOptions := map[string]struct{}{}
		for _, option := range field.Options {
			if option.Value == "" || option.Label == "" {
				return fmt.Errorf("field %q has incomplete option", field.Key)
			}
			if _, exists := seenOptions[option.Value]; exists {
				return fmt.Errorf("field %q has duplicate option %q", field.Key, option.Value)
			}
			seenOptions[option.Value] = struct{}{}
		}
		if len(field.Default) > 0 && !json.Valid(field.Default) {
			return fmt.Errorf("field %q has invalid default JSON", field.Key)
		}
		if field.RelativeTo != "" && field.RelativeTo != "sandbox" {
			return fmt.Errorf("field %q has invalid relativeTo %q", field.Key, field.RelativeTo)
		}
		switch field.Export {
		case "", "value", "file_contents", "csv":
		default:
			return fmt.Errorf("field %q has invalid export %q", field.Key, field.Export)
		}
		switch field.Generate {
		case "", "uuid", "secret", "local_mac", "free_port":
		default:
			return fmt.Errorf("field %q has invalid generator %q", field.Key, field.Generate)
		}
	}
	for _, constraint := range s.Constraints {
		if constraint.Kind != "all_or_none" || len(constraint.Fields) < 2 {
			return fmt.Errorf("invalid constraint %q", constraint.Kind)
		}
		for _, key := range constraint.Fields {
			if _, exists := fields[key]; !exists {
				return fmt.Errorf("constraint references unknown field %q", key)
			}
		}
	}
	tools := map[string]struct{}{}
	for _, tool := range s.Tools {
		if tool.ID == "" || tool.Name == "" {
			return fmt.Errorf("tool id and name are required")
		}
		if _, exists := tools[tool.ID]; exists {
			return fmt.Errorf("duplicate tool %q", tool.ID)
		}
		tools[tool.ID] = struct{}{}
		if tool.Environment != "" && !environmentName.MatchString(tool.Environment) {
			return fmt.Errorf("tool %q has invalid environment %q", tool.ID, tool.Environment)
		}
		if tool.Source != ToolSourceSystem {
			return fmt.Errorf("tool %q has invalid source %q", tool.ID, tool.Source)
		}
		if tool.Kind != ToolKindExecutable && tool.Kind != ToolKindFile {
			return fmt.Errorf("tool %q has invalid kind %q", tool.ID, tool.Kind)
		}
		if tool.Kind == ToolKindExecutable && tool.Required && len(tool.Checks) == 0 {
			return fmt.Errorf("required tool %q must define checks", tool.ID)
		}
		if tool.Kind == ToolKindFile && len(tool.Checks) != 0 {
			return fmt.Errorf("file tool %q cannot define executable checks", tool.ID)
		}
		for i, check := range tool.Checks {
			if len(check.Arguments) == 0 {
				return fmt.Errorf("tool %q check %d requires arguments", tool.ID, i)
			}
			for _, argument := range check.Arguments {
				if argument == "" {
					return fmt.Errorf("tool %q check %d has an empty argument", tool.ID, i)
				}
			}
			for _, expected := range check.Contains {
				if expected == "" {
					return fmt.Errorf("tool %q check %d has an empty output match", tool.ID, i)
				}
			}
		}
	}
	resources := map[string]struct{}{}
	for _, resource := range s.Resources {
		if !externalResourceID.MatchString(resource.ID) || resource.Name == "" || resource.Description == "" {
			return fmt.Errorf("invalid external resource %q", resource.ID)
		}
		if _, exists := resources[resource.ID]; exists {
			return fmt.Errorf("duplicate external resource %q", resource.ID)
		}
		resources[resource.ID] = struct{}{}
		if resource.Kind != ExternalResourceService && resource.Kind != ExternalResourceDevice {
			return fmt.Errorf("external resource %q has invalid kind %q", resource.ID, resource.Kind)
		}
		resourceFields := map[string]struct{}{}
		for _, key := range resource.Fields {
			if _, exists := fields[key]; !exists {
				return fmt.Errorf("external resource %q references unknown field %q", resource.ID, key)
			}
			if _, exists := resourceFields[key]; exists {
				return fmt.Errorf("external resource %q repeats field %q", resource.ID, key)
			}
			resourceFields[key] = struct{}{}
		}
		switch resource.Activation.Mode {
		case ResourceActivationAlways:
			if len(resource.Activation.Fields) != 0 {
				return fmt.Errorf("always-active external resource %q cannot have activation fields", resource.ID)
			}
		case ResourceActivationConfigured:
			if len(resource.Activation.Fields) == 0 {
				return fmt.Errorf("configured external resource %q requires activation fields", resource.ID)
			}
			for _, key := range resource.Activation.Fields {
				if _, exists := resourceFields[key]; !exists {
					return fmt.Errorf("external resource %q activation references undeclared field %q", resource.ID, key)
				}
			}
		default:
			return fmt.Errorf("external resource %q has invalid activation mode %q", resource.ID, resource.Activation.Mode)
		}
	}
	return nil
}
