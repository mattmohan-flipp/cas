package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryProxyStore(t *testing.T) {
	store := NewMemoryProxyStore()
	assert.NotNil(t, store)

	// Try before anything is set
	pgt, ok := store.Get("iou")
	assert.False(t, ok)
	assert.Equal(t, "", pgt)

	// Try after setting
	store.Set("iou", "pgt")
	pgt, ok = store.Get("iou")
	assert.True(t, ok)
	assert.Equal(t, "pgt", pgt)

	// Try after deleting
	store.Delete("iou")
	pgt, ok = store.Get("iou")
	assert.False(t, ok)
	assert.Equal(t, "", pgt)

	// Set again and then clear
	store.Set("iou", "pgt")
	store.Clear()
	pgt, ok = store.Get("iou")
	assert.False(t, ok)
	assert.Equal(t, "", pgt)
}
