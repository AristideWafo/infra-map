package ws

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func dial(t *testing.T, hub *Hub) *websocket.Conn {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(hub.Handle))
	t.Cleanup(s.Close)
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http"), nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func TestHub_BroadcastReachesClient(t *testing.T) {
	hub := NewHub(testLog())
	conn := dial(t, hub)

	// Attendre l'enregistrement
	require.Eventually(t, func() bool { return hub.ClientCount() == 1 }, time.Second, 10*time.Millisecond)

	hub.Broadcast(Message{Type: "alert_fired", Payload: models.Alert{ID: "a1", NodeID: "pod-1"}})

	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	_, data, err := conn.ReadMessage()
	require.NoError(t, err)

	var msg Message
	require.NoError(t, json.Unmarshal(data, &msg))
	assert.Equal(t, "alert_fired", msg.Type)
}

type fakeAlerts struct {
	alerts []models.Alert
	err    error
}

func (f *fakeAlerts) Alerts(_ context.Context) ([]models.Alert, error) { return f.alerts, f.err }

func TestAlertWatcher_FiredThenResolved(t *testing.T) {
	hub := NewHub(testLog())
	conn := dial(t, hub)
	require.Eventually(t, func() bool { return hub.ClientCount() == 1 }, time.Second, 10*time.Millisecond)

	provider := &fakeAlerts{alerts: []models.Alert{{ID: "a1", NodeID: "pod-1", Name: "HighCPU"}}}
	w := NewAlertWatcher(provider, hub, time.Hour, testLog())

	read := func() Message {
		_ = conn.SetReadDeadline(time.Now().Add(time.Second))
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)
		var m Message
		require.NoError(t, json.Unmarshal(data, &m))
		return m
	}

	w.Poll(context.Background())
	assert.Equal(t, "alert_fired", read().Type)

	// Même alerte au poll suivant : pas de re-broadcast
	w.Poll(context.Background())

	// Alerte disparue : resolved
	provider.alerts = nil
	w.Poll(context.Background())
	assert.Equal(t, "alert_resolved", read().Type)
}

func TestAlertWatcher_ProviderErrorKeepsState(t *testing.T) {
	hub := NewHub(testLog())
	provider := &fakeAlerts{alerts: []models.Alert{{ID: "a1"}}}
	w := NewAlertWatcher(provider, hub, time.Hour, testLog())

	w.Poll(context.Background())
	provider.err = errors.New("prometheus down")
	w.Poll(context.Background()) // ne panique pas, ne "résout" pas l'alerte
	assert.Len(t, w.active, 1)
}

func TestHub_WelcomeReplaysActiveAlerts(t *testing.T) {
	hub := NewHub(testLog())
	provider := &fakeAlerts{alerts: []models.Alert{{ID: "a1", NodeID: "pod-1"}}}
	w := NewAlertWatcher(provider, hub, time.Hour, testLog())

	// L'alerte se déclenche AVANT toute connexion client
	w.Poll(context.Background())

	conn := dial(t, hub)
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	_, data, err := conn.ReadMessage()
	require.NoError(t, err)

	var msg Message
	require.NoError(t, json.Unmarshal(data, &msg))
	assert.Equal(t, "alert_fired", msg.Type)
}
