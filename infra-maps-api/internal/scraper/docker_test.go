package scraper

import (
	"context"
	"errors"
	"testing"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDockerAPI struct {
	containers []container.Summary
	err        error
}

func (f *fakeDockerAPI) ContainerList(_ context.Context, _ container.ListOptions) ([]container.Summary, error) {
	return f.containers, f.err
}

func TestDocker_ScrapeStandaloneAndComposeGroups(t *testing.T) {
	fake := &fakeDockerAPI{containers: []container.Summary{
		{
			ID: "abc123def456789", Names: []string{"/web"}, State: container.StateRunning,
			Labels: map[string]string{composeProjectLabel: "myapp"},
		},
		{
			ID: "fed654cba321000", Names: []string{"/db"}, State: container.StateExited,
			Labels: map[string]string{composeProjectLabel: "myapp"},
		},
		{ID: "aaa111bbb222ccc", Names: []string{"/standalone"}, State: container.StateRunning},
	}}
	d := NewDockerWithAPI(fake)

	nodes, err := d.Scrape(context.Background())
	require.NoError(t, err)
	require.Len(t, nodes, 4) // 1 group + 3 containers

	group := nodes[0]
	assert.Equal(t, "compose-myapp", group.ID)
	assert.Equal(t, models.NodeTypeDockerComposeGroup, group.Type)

	web := nodes[1]
	assert.Equal(t, "abc123def456", web.ID) // tronqué à 12
	assert.Equal(t, "web", web.Name)        // slash retiré
	assert.Equal(t, "compose-myapp", web.ParentID)
	assert.Equal(t, models.StatusHealthy, web.Status)

	db := nodes[2]
	assert.Equal(t, models.StatusCritical, db.Status) // exited

	standalone := nodes[3]
	assert.Empty(t, standalone.ParentID)
}

func TestDocker_ErrorPropagatesToHealth(t *testing.T) {
	fake := &fakeDockerAPI{err: errors.New("socket not found")}
	d := NewDockerWithAPI(fake)

	_, err := d.Scrape(context.Background())
	require.Error(t, err)
	assert.Error(t, d.Health())
}

func TestContainerStatus(t *testing.T) {
	tests := []struct {
		state container.ContainerState
		want  models.NodeStatus
	}{
		{container.StateRunning, models.StatusHealthy},
		{container.StateRestarting, models.StatusWarning},
		{container.StatePaused, models.StatusWarning},
		{container.StateExited, models.StatusCritical},
		{container.StateDead, models.StatusCritical},
		{container.StateCreated, models.StatusUnknown},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, containerStatus(tt.state), string(tt.state))
	}
}
