package manifest

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func validManifest() Manifest {
	return Manifest{
		ProtocolVersion: "1", RubricVersion: "1", FixtureRevision: "43eeb7f",
		Prompt:              Prompt{ID: "p-1", Fingerprint: "sha256:aa"},
		ActiveStackRevision: "43eeb7f",
		Variant:             "actual",
		Profile:             Profile{Name: "descartable", ConfigFingerprint: "sha256:bb"},
		Seed:                42,
		Cohorts:             []Cohort{cohort("a", nil), cohort("b", nil)},
	}
}

func TestManifestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Manifest)
		wantErr string
	}{
		{"valido", func(m *Manifest) {}, ""},
		{"sin version de protocolo", func(m *Manifest) { m.ProtocolVersion = "" }, "protocol_version"},
		{"sin revision de fixture", func(m *Manifest) { m.FixtureRevision = "" }, "fixture_revision"},
		{"sin cohortes", func(m *Manifest) { m.Cohorts = nil }, "cohortes"},
		{"ids de cohorte duplicados", func(m *Manifest) { m.Cohorts[1].ID = "a" }, "duplicad"},
		{"peso no positivo", func(m *Manifest) { m.Cohorts[0].Tasks[0].Weight = 0 }, "peso"},
		{"cohorte sin tareas", func(m *Manifest) { m.Cohorts[0].Tasks = nil }, "tareas"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validManifest()
			tt.mutate(&m)
			err := m.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestLoad_RoundTripAndUnknownFields(t *testing.T) {
	want := validManifest()
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Load(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load() = %+v, want %+v", got, want)
	}

	if _, err := Load(strings.NewReader(`{"protocol_version":"1","api_key":"secreto"}`)); err == nil {
		t.Error("Load() debería rechazar campos desconocidos")
	}
	if _, err := Load(strings.NewReader(`{`)); err == nil {
		t.Error("Load() debería fallar con JSON inválido")
	}
	bad := validManifest()
	bad.ProtocolVersion = ""
	badData, _ := json.Marshal(bad)
	if _, err := Load(strings.NewReader(string(badData))); err == nil {
		t.Error("Load() debería devolver el error de Validate")
	}
}
