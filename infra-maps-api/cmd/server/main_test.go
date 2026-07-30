package main

import (
	"io"
	"log/slog"
	"testing"

	"github.com/aristidewafo/infra-maps-api/internal/config"
	"github.com/stretchr/testify/assert"
)

func discardLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func scraperNames(t *testing.T, cfg config.Config) []string {
	t.Helper()
	scrapers, _ := buildScrapers(cfg, discardLog())
	names := make([]string, len(scrapers))
	for i, s := range scrapers {
		names[i] = s.Name()
	}
	return names
}

// Phase 3 — DEPLOYMENT.md exige un mode fonctionnel par combinaison de
// sources désactivées. Prometheus reste la seule source obligatoire
// (CLAUDE.md) : elle est toujours présente, réelle ou en Noop de secours.

func TestBuildScrapers_VMOnlyMode(t *testing.T) {
	cfg := config.Config{PrometheusURL: "http://prom:9090", K8sEnabled: false, DockerEnabled: false}
	assert.Equal(t, []string{"prometheus"}, scraperNames(t, cfg))
}

func TestBuildScrapers_KubernetesOnlyMode(t *testing.T) {
	cfg := config.Config{
		PrometheusURL: "http://prom:9090", K8sEnabled: true, K8sInCluster: false,
		K8sKubeconfig: "/nonexistent/kubeconfig", DockerEnabled: false,
	}
	names := scraperNames(t, cfg)
	assert.Contains(t, names, "prometheus")
	assert.Contains(t, names, "kubernetes")
	assert.Len(t, names, 2)
}

func TestBuildScrapers_DockerOnlyMode(t *testing.T) {
	cfg := config.Config{
		PrometheusURL: "http://prom:9090", K8sEnabled: false,
		DockerEnabled: true, DockerSocket: "/var/run/docker.sock",
	}
	names := scraperNames(t, cfg)
	assert.Contains(t, names, "prometheus")
	assert.Contains(t, names, "docker")
	assert.Len(t, names, 2)
}

func TestBuildScrapers_AllSourcesEnabled(t *testing.T) {
	cfg := config.Config{
		PrometheusURL: "http://prom:9090", K8sEnabled: true, K8sInCluster: false,
		K8sKubeconfig: "/nonexistent/kubeconfig", DockerEnabled: true, DockerSocket: "/var/run/docker.sock",
	}
	assert.Len(t, scraperNames(t, cfg), 3)
}

func TestBuildScrapers_MockModeIgnoresSourceToggles(t *testing.T) {
	// Le mock tient lieu de source obligatoire en dev — K8s/Docker n'entrent
	// pas en jeu tant que MockEnabled force la simulation complète.
	cfg := config.Config{MockEnabled: true, K8sEnabled: true, DockerEnabled: true}
	assert.Equal(t, []string{"prometheus"}, scraperNames(t, cfg))
}

func TestBuildScrapers_KubernetesInitFailureFallsBackToNoop(t *testing.T) {
	// K8sInCluster=true hors cluster : rest.InClusterConfig() échoue toujours.
	cfg := config.Config{
		PrometheusURL: "http://prom:9090", K8sEnabled: true, K8sInCluster: true, DockerEnabled: false,
	}
	names := scraperNames(t, cfg)
	assert.Contains(t, names, "kubernetes") // Noop porte le même nom
	assert.Len(t, names, 2)
}
