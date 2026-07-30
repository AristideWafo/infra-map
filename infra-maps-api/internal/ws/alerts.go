package ws

import (
	"context"
	"log/slog"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
)

// AlertsProvider fournit l'état courant des alertes (Prometheus ou mock).
type AlertsProvider interface {
	Alerts(ctx context.Context) ([]models.Alert, error)
}

// AlertWatcher poll le provider et diffuse les transitions fired/resolved.
type AlertWatcher struct {
	provider AlertsProvider
	hub      *Hub
	interval time.Duration
	log      *slog.Logger

	active map[string]models.Alert
}

func NewAlertWatcher(p AlertsProvider, hub *Hub, interval time.Duration, log *slog.Logger) *AlertWatcher {
	return &AlertWatcher{provider: p, hub: hub, interval: interval, log: log, active: map[string]models.Alert{}}
}

func (w *AlertWatcher) Run(ctx context.Context) {
	w.Poll(ctx)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.Poll(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// Poll compare l'état courant à l'état précédent et broadcast les diffs.
func (w *AlertWatcher) Poll(ctx context.Context) {
	alerts, err := w.provider.Alerts(ctx)
	if err != nil {
		w.log.Warn("alerts poll failed", "error", err)
		return
	}

	current := make(map[string]models.Alert, len(alerts))
	for _, a := range alerts {
		current[a.ID] = a
		if _, known := w.active[a.ID]; !known {
			w.hub.Broadcast(Message{Type: "alert_fired", Payload: a})
		}
	}
	for id, a := range w.active {
		if _, still := current[id]; !still {
			w.hub.Broadcast(Message{Type: "alert_resolved", Payload: map[string]interface{}{
				"id": a.ID, "nodeId": a.NodeID, "resolvedAt": time.Now().UTC(),
			}})
		}
	}
	w.active = current
}
