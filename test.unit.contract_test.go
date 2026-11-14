package contract

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestHelloResponseValidate(t *testing.T) {
	tests := []struct {
		name    string
		resp    HelloResponse
		wantErr bool
	}{
		{
			name: "valid service",
			resp: HelloResponse{
				ID:              "messenger",
				Kind:            KindService,
				ContractVersion: ContractVersion,
			},
		},
		{
			name: "valid plugin with deps",
			resp: HelloResponse{
				ID:              "esphome-plugin",
				Kind:            KindPlugin,
				ContractVersion: ContractVersion,
				DependsOn:       []string{"messenger", "registry"},
				Definitions:     DescribePlugin("ESPHome", "Discovers ESPHome devices."),
			},
		},
		{
			name:    "missing id",
			resp:    HelloResponse{Kind: KindService, ContractVersion: ContractVersion},
			wantErr: true,
		},
		{
			name:    "bad kind",
			resp:    HelloResponse{ID: "x", Kind: "worker", ContractVersion: ContractVersion},
			wantErr: true,
		},
		{
			name:    "wrong version",
			resp:    HelloResponse{ID: "x", Kind: KindService, ContractVersion: 99},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resp.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWriteAndReadJSON(t *testing.T) {
	var buf bytes.Buffer

	msg := RuntimeMessage{Type: RuntimeReady}
	if err := WriteJSON(&buf, msg); err != nil {
		t.Fatal(err)
	}

	dec := json.NewDecoder(&buf)
	var got RuntimeMessage
	if err := ReadJSON(dec, &got); err != nil {
		t.Fatal(err)
	}

	if got.Type != RuntimeReady {
		t.Errorf("got type %q, want %q", got.Type, RuntimeReady)
	}
}

func TestSetupDefinitionValidate(t *testing.T) {
	setup := SelectableSetup("Media",
		SetupField{Key: "url", Environment: "SERVER_URL", Name: "Server URL", Control: SetupText, Format: "url", Required: true},
		SetupField{Key: "password", Environment: "SERVER_PASSWORD", Name: "Password", Control: SetupSecret, Sensitive: true},
		SetupField{Key: "profile", Environment: "PROFILE", Name: "Profile", Control: SetupSelect, Default: json.RawMessage(`"portable"`), Options: []SetupOption{{Value: "portable", Label: "Portable"}}},
	)
	setup.Constraints = []SetupConstraint{{Kind: "all_or_none", Fields: []string{"url", "password"}}}
	setup.Tools = []ToolRequirement{{ID: "ffmpeg", Name: "FFmpeg", Source: ToolSourceSystem, Kind: ToolKindExecutable, Environment: "FFMPEG_PATH", Required: true, Checks: []ToolCheck{{Arguments: []string{"-version"}, Contains: []string{"ffmpeg version"}}}}}
	setup.Fields = append(setup.Fields, SetupField{Key: "port", Environment: "PORT", Name: "Port", Control: SetupText, Generate: "free_port", Advanced: true})
	if err := setup.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	bad := *setup
	bad.Fields = append(append([]SetupField(nil), setup.Fields...), SetupField{Key: "token", Name: "Token", Control: SetupSecret})
	if err := bad.Validate(); err == nil {
		t.Fatal("expected an unmarked secret to fail validation")
	}
}

func TestSetupDefinitionRejectsInvalidTool(t *testing.T) {
	setup := RequiredSetup("Core")
	setup.Tools = []ToolRequirement{{ID: "ffmpeg", Name: "FFmpeg", Source: ToolSourceSystem, Kind: ToolKindExecutable, Required: true}}
	if err := setup.Validate(); err == nil {
		t.Fatal("expected required tool without checks to fail validation")
	}
	setup.Tools[0].Checks = []ToolCheck{{Arguments: []string{"-version"}}}
	setup.Tools[0].Source = "download"
	if err := setup.Validate(); err == nil {
		t.Fatal("expected unsupported tool source to fail validation")
	}
}

func TestHelloResponseRejectsInvalidSetup(t *testing.T) {
	resp := HelloResponse{ID: "plugin", Kind: KindPlugin, ContractVersion: ContractVersion, Definitions: DescribePlugin("Plugin", "A plugin."), Setup: &SetupDefinition{Mode: SetupSelectable}}
	if err := resp.Validate(); err == nil {
		t.Fatal("expected invalid setup to fail hello validation")
	}
}
