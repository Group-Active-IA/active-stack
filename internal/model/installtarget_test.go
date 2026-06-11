package model_test

import (
	"testing"

	"github.com/Group-Active-IA/active-stack/internal/model"
)

func TestInstallTarget_ZeroValueIsMachine(t *testing.T) {
	var target model.InstallTarget
	if target != model.Machine {
		t.Errorf("zero-value of InstallTarget = %v, want Machine", target)
	}
}

func TestInstallTarget_Values(t *testing.T) {
	tests := []struct {
		name   string
		target model.InstallTarget
		want   string
	}{
		{
			name:   "Machine",
			target: model.Machine,
			want:   "machine",
		},
		{
			name:   "Project",
			target: model.Project,
			want:   "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.target.String(); got != tt.want {
				t.Errorf("InstallTarget(%v).String() = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}

func TestInstallTarget_ZeroValueString(t *testing.T) {
	var target model.InstallTarget
	if got := target.String(); got != "machine" {
		t.Errorf("zero-value InstallTarget.String() = %q, want %q", got, "machine")
	}
}
