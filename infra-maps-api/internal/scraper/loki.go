package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
)

// Loki interroge l'API query_range de Loki à la demande (pas de polling).
// Pas un Scraper : les logs ne rentrent pas dans l'arbre, ils sont proxifiés.
type Loki struct {
	baseURL  string
	client   *http.Client
	maxLines int
}

func NewLoki(baseURL string, maxLines int) *Loki {
	if maxLines <= 0 {
		maxLines = 100
	}
	return &Loki{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		client:   &http.Client{Timeout: 10 * time.Second},
		maxLines: maxLines,
	}
}

// Enabled indique si Loki est configuré.
func (l *Loki) Enabled() bool { return l != nil && l.baseURL != "" }

type lokiResponse struct {
	Data struct {
		Result []struct {
			Stream map[string]string `json:"stream"`
			Values [][2]string       `json:"values"` // [ts_ns, line]
		} `json:"result"`
	} `json:"data"`
}

// Logs retourne les entrées récentes pour un pod/container (label pod=nodeName).
func (l *Loki) Logs(ctx context.Context, nodeName string, limit int, from time.Time, level string) ([]models.LogEntry, error) {
	if limit <= 0 || limit > l.maxLines {
		limit = l.maxLines
	}
	selector := fmt.Sprintf(`{pod=%q}`, nodeName)
	if level != "" {
		selector += fmt.Sprintf(` |= %q`, level)
	}

	q := url.Values{}
	q.Set("query", selector)
	q.Set("limit", strconv.Itoa(limit))
	q.Set("start", strconv.FormatInt(from.UnixNano(), 10))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.baseURL+"/loki/api/v1/query_range?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("loki request: %w", err)
	}
	res, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("loki query: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("loki status %d", res.StatusCode)
	}

	var body lokiResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("loki decode: %w", err)
	}

	var entries []models.LogEntry
	for _, r := range body.Data.Result {
		for _, v := range r.Values {
			ns, err := strconv.ParseInt(v[0], 10, 64)
			if err != nil {
				continue
			}
			entries = append(entries, models.LogEntry{
				Timestamp: time.Unix(0, ns),
				Level:     detectLevel(v[1]),
				Message:   v[1],
				Pod:       r.Stream["pod"],
				Container: r.Stream["container"],
			})
		}
	}
	return entries, nil
}

func detectLevel(line string) string {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "error"):
		return "error"
	case strings.Contains(lower, "warn"):
		return "warn"
	case strings.Contains(lower, "debug"):
		return "debug"
	default:
		return "info"
	}
}
