package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config regroupe toutes les variables d'environnement (API_CONTRACT.md).
type Config struct {
	Port           string
	PrometheusURL  string
	K8sEnabled     bool
	K8sInCluster   bool
	K8sKubeconfig  string
	DockerEnabled  bool
	DockerSocket   string
	LokiURL        string
	ScrapeInterval time.Duration
	CacheTTL       time.Duration
	CORSOrigin     string
	// MockEnabled force le scraper mocké (dev sans infra). Activé
	// automatiquement si PROMETHEUS_URL est vide.
	MockEnabled bool
}

func Load(log *slog.Logger) Config {
	cfg := Config{
		Port:           str("PORT", "8080"),
		PrometheusURL:  str("PROMETHEUS_URL", ""),
		K8sEnabled:     boolean("K8S_ENABLED", true, log),
		K8sInCluster:   boolean("K8S_IN_CLUSTER", true, log),
		K8sKubeconfig:  str("K8S_KUBECONFIG", os.Getenv("HOME")+"/.kube/config"),
		DockerEnabled:  boolean("DOCKER_ENABLED", true, log),
		DockerSocket:   str("DOCKER_SOCKET", "/var/run/docker.sock"),
		LokiURL:        str("LOKI_URL", ""),
		ScrapeInterval: duration("SCRAPE_INTERVAL", 30*time.Second, log),
		CacheTTL:       duration("CACHE_TTL", 30*time.Second, log),
		CORSOrigin:     str("CORS_ORIGIN", "*"),
		MockEnabled:    boolean("MOCK_ENABLED", false, log),
	}
	if cfg.PrometheusURL == "" && !cfg.MockEnabled {
		log.Warn("PROMETHEUS_URL empty — falling back to mock scraper")
		cfg.MockEnabled = true
	}
	return cfg
}

func str(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func boolean(key string, def bool, log *slog.Logger) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		log.Warn("invalid bool env, using default", "key", key, "value", raw, "default", def)
		return def
	}
	return v
}

func duration(key string, def time.Duration, log *slog.Logger) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		log.Warn("invalid duration env, using default", "key", key, "value", raw, "default", def.String())
		return def
	}
	return d
}
