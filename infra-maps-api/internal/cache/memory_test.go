package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemory_SetGet(t *testing.T) {
	c := New(30 * time.Second)
	c.Set("tree", "value")

	got, ok := c.Get("tree")
	assert.True(t, ok)
	assert.Equal(t, "value", got)
}

func TestMemory_MissingKey(t *testing.T) {
	c := New(30 * time.Second)
	_, ok := c.Get("absent")
	assert.False(t, ok)
	assert.Equal(t, -1, c.Age("absent"))
}

func TestMemory_Expiration(t *testing.T) {
	c := New(30 * time.Second)
	current := time.Now()
	c.now = func() time.Time { return current }

	c.Set("tree", "value")
	current = current.Add(31 * time.Second)

	_, ok := c.Get("tree")
	assert.False(t, ok)
	// L'âge reste consultable après expiration (données périmées signalées, pas perdues)
	assert.Equal(t, 31, c.Age("tree"))
}

func TestMemory_Age(t *testing.T) {
	c := New(30 * time.Second)
	current := time.Now()
	c.now = func() time.Time { return current }

	c.Set("tree", "value")
	current = current.Add(12 * time.Second)

	assert.Equal(t, 12, c.Age("tree"))
}

func TestMemory_ConcurrentAccess(t *testing.T) {
	c := New(30 * time.Second)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			c.Set("k", i)
		}
		close(done)
	}()
	for i := 0; i < 1000; i++ {
		c.Get("k")
		c.Age("k")
	}
	<-done
}
