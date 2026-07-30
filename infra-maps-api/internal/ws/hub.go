package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message est l'enveloppe des messages WS (API_CONTRACT.md).
type Message struct {
	Type    string      `json:"type"` // "alert_fired", "alert_resolved", "ping"
	Payload interface{} `json:"payload,omitempty"`
}

// Hub diffuse les messages à tous les clients WebSocket connectés.
type Hub struct {
	mu       sync.RWMutex
	clients  map[*websocket.Conn]struct{}
	upgrader websocket.Upgrader
	log      *slog.Logger
}

func NewHub(log *slog.Logger) *Hub {
	return &Hub{
		clients: map[*websocket.Conn]struct{}{},
		upgrader: websocket.Upgrader{
			// CORS est géré au niveau HTTP ; le handshake WS suit la même politique.
			CheckOrigin: func(*http.Request) bool { return true },
		},
		log: log,
	}
}

// Handle upgrade la connexion et l'enregistre jusqu'à déconnexion.
func (h *Hub) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Warn("ws upgrade failed", "error", err)
		return
	}
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()

	// Lecture bloquante : détecte la fermeture côté client.
	go func() {
		defer h.remove(conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

func (h *Hub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	_ = conn.Close()
}

// Broadcast envoie le message à tous les clients. Les clients en erreur sont retirés.
func (h *Hub) Broadcast(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		h.log.Error("ws marshal failed", "error", err)
		return
	}

	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.clients))
	for c := range h.clients {
		conns = append(conns, c)
	}
	h.mu.RUnlock()

	for _, c := range conns {
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			h.remove(c)
		}
	}
}

// ClientCount retourne le nombre de clients connectés (tests, métriques).
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// RunPing envoie un keepalive toutes les 30s jusqu'à annulation.
func (h *Hub) RunPing(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			h.Broadcast(Message{Type: "ping"})
		case <-ctx.Done():
			return
		}
	}
}
