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
