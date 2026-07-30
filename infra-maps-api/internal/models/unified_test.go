package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateStatus(t *testing.T) {
	tests := []struct {
		name     string
		cpu      float64
		memPct   float64
		restarts int
		want     NodeStatus
	}{
		{"healthy — all ok", 50, 60, 0, StatusHealthy},
		{"warning — cpu at 70", 70, 60, 0, StatusWarning},
		{"warning — cpu 75", 75, 60, 0, StatusWarning},
		{"critical — cpu above 85", 90, 60, 0, StatusCritical},
		{"warning — memory at 80", 40, 80, 0, StatusWarning},
		{"critical — memory above 90", 30, 92, 0, StatusCritical},
		{"warning — restarts", 40, 50, 3, StatusWarning},
		{"critical wins over restarts", 90, 50, 3, StatusCritical},
		{"boundary — cpu 85 is warning", 85, 50, 0, StatusWarning},
		{"boundary — mem 90 is warning", 10, 90, 0, StatusWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, CalculateStatus(tt.cpu, tt.memPct, tt.restarts))
		})
	}
}

func TestPropagateStatus(t *testing.T) {
	tree := &UnifiedNode{
		ID: "root", Status: StatusHealthy,
		Children: []*UnifiedNode{
			{ID: "cluster", Status: StatusHealthy, Children: []*UnifiedNode{
				{ID: "pod-1", Status: StatusHealthy},
				{ID: "pod-2", Status: StatusCritical},
			}},
			{ID: "island", Status: StatusHealthy, Children: []*UnifiedNode{
				{ID: "vm-1", Status: StatusWarning},
			}},
		},
	}

	got := PropagateStatus(tree)

	assert.Equal(t, StatusCritical, got)
	assert.Equal(t, StatusCritical, tree.Children[0].Status)
	assert.Equal(t, StatusWarning, tree.Children[1].Status)
}

func TestPropagateStatus_LeafKeepsOwnStatus(t *testing.T) {
	leaf := &UnifiedNode{ID: "vm", Status: StatusUnknown}
	assert.Equal(t, StatusUnknown, PropagateStatus(leaf))
}
