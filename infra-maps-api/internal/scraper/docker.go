package scraper

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const composeProjectLabel = "com.docker.compose.project"

// dockerAPI est le sous-ensemble du client Docker utilisé — mockable en tests.
type dockerAPI interface {
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
}

// Docker scrape les containers standalone via le socket. Les containers d'un
// même projet Compose sont groupés sous un compose-group.
type Docker struct {
	api dockerAPI
	now func() time.Time

	mu      sync.RWMutex
	lastErr error
}

func NewDocker(socket string) (*Docker, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHost("unix://"+strings.TrimPrefix(socket, "unix://")),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return &Docker{api: cli, now: time.Now}, nil
}

// NewDockerWithAPI est le constructeur de test.
func NewDockerWithAPI(api dockerAPI) *Docker {
	return &Docker{api: api, now: time.Now}
}

func (d *Docker) Name() string { return "docker" }

func (d *Docker) Health() error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.lastErr
}

func (d *Docker) setErr(err error) {
	d.mu.Lock()
	d.lastErr = err
	d.mu.Unlock()
}

func (d *Docker) Scrape(ctx context.Context) ([]*models.UnifiedNode, error) {
	containers, err := d.api.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		err = fmt.Errorf("docker list containers: %w", err)
		d.setErr(err)
		return nil, err
	}

	seen := d.now()
	groups := map[string]*models.UnifiedNode{}
	var nodes []*models.UnifiedNode

	for _, c := range containers {
		id := c.ID
		if len(id) > 12 {
			id = id[:12]
		}
		name := id
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		parentID := ""
		if project := c.Labels[composeProjectLabel]; project != "" {
			groupID := "compose-" + project
			if _, ok := groups[groupID]; !ok {
				group := &models.UnifiedNode{
					ID: groupID, Name: project, Type: models.NodeTypeDockerComposeGroup,
					Status: models.StatusHealthy, Source: "docker", LastSeen: seen,
				}
				groups[groupID] = group
				nodes = append(nodes, group)
			}
			parentID = groupID
		}

		nodes = append(nodes, &models.UnifiedNode{
			ID: id, Name: name, Type: models.NodeTypeContainer, ParentID: parentID,
			Status: containerStatus(c.State), Source: "docker", LastSeen: seen,
			RawLabels: c.Labels,
		})
	}

	d.setErr(nil)
	return nodes, nil
}

func (d *Docker) Connections(ctx context.Context) ([]*models.Connection, error) {
	return nil, nil
}

func containerStatus(state container.ContainerState) models.NodeStatus {
	switch state {
	case container.StateRunning:
		return models.StatusHealthy
	case container.StateRestarting, container.StatePaused:
		return models.StatusWarning
	case container.StateExited, container.StateDead:
		return models.StatusCritical
	default:
		return models.StatusUnknown
	}
}
