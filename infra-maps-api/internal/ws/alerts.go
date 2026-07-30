package ws

import (
	"context"
	"log/slog"
	"sync"
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

	mu     sync.RWMutex
	active map[string]models.Alert
}

func NewAlertWatcher(p AlertsProvider, hub *Hub, interval time.Duration, log *slog.Logger) *AlertWatcher {
	w := &AlertWatcher{provider: p, hub: hub, interval: interval, log: log, active: map[string]models.Alert{}}
	hub.SetWelcome(w.snapshot)
	return w
}

// snapshot rejoue les alertes actives aux nouveaux clients.
func (w *AlertWatcher) snapshot() []Message {
	w.mu.RLock()
	defer w.mu.RUnlock()
	msgs := make([]Message, 0, len(w.active))
	for _, a := range w.active {
		msgs = append(msgs, Message{Type: "alert_fired", Payload: a})
	}
	return msgs
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

	w.mu.RLock()
	previous := w.active
	w.mu.RUnlock()

	current := make(map[string]models.Alert, len(alerts))
	for _, a := range alerts {
		current[a.ID] = a
		if _, known := previous[a.ID]; !known {
			w.hub.Broadcast(Message{Type: "alert_fired", Payload: a})
		}
	}
	for id, a := range previous {
		if _, still := current[id]; !still {
			w.hub.Broadcast(Message{Type: "alert_resolved", Payload: map[string]interface{}{
				"id": a.ID, "nodeId": a.NodeID, "resolvedAt": time.Now().UTC(),
			}})
		}
	}
	w.mu.Lock()
	w.active = current
	w.mu.Unlock()
}
