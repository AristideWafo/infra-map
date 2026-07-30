package scraper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoki_Logs(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("query")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"result": []map[string]interface{}{{
					"stream": map[string]string{"pod": "api-1", "container": "api"},
					"values": [][2]string{
						{"1700000000000000000", "ERROR connection refused to postgres:5432"},
						{"1700000001000000000", "request handled in 12ms"},
					},
				}},
			},
		})
	}))
	defer srv.Close()

	l := NewLoki(srv.URL, 100)
	entries, err := l.Logs(context.Background(), "api-1", 50, time.Unix(0, 0), "")
	require.NoError(t, err)
	require.Len(t, entries, 2)

	assert.Equal(t, `{pod="api-1"}`, gotQuery)
	assert.Equal(t, "error", entries[0].Level)
	assert.Equal(t, "api-1", entries[0].Pod)
	assert.Equal(t, "api", entries[0].Container)
	assert.Equal(t, "info", entries[1].Level)
}

func TestLoki_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	l := NewLoki(srv.URL, 100)
	_, err := l.Logs(context.Background(), "api-1", 50, time.Unix(0, 0), "")
	assert.Error(t, err)
}

func TestLoki_Enabled(t *testing.T) {
	assert.False(t, NewLoki("", 100).Enabled())
	assert.True(t, NewLoki("http://loki:3100", 100).Enabled())
	var nilLoki *Loki
	assert.False(t, nilLoki.Enabled())
}

func TestDetectLevel(t *testing.T) {
	assert.Equal(t, "error", detectLevel("ERROR boom"))
	assert.Equal(t, "warn", detectLevel("WARN disk"))
	assert.Equal(t, "debug", detectLevel("debug trace"))
	assert.Equal(t, "info", detectLevel("hello"))
}
