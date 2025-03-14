package store

import "sync"

type MemoryProxyStore struct {
	mutex sync.RWMutex
	store map[string]string
}

func NewMemoryProxyStore() *MemoryProxyStore {
	return &MemoryProxyStore{
		store: make(map[string]string),
	}
}

// Clear implements ProxyStore.
func (m *MemoryProxyStore) Clear() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.store = make(map[string]string)
	return nil
}

// Delete implements ProxyStore.
func (m *MemoryProxyStore) Delete(iou string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.store, iou)
	return nil
}

// Get implements ProxyStore.
func (m *MemoryProxyStore) Get(iou string) (string, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	pgt, ok := m.store[iou]
	return pgt, ok
}

// Set implements ProxyStore.
func (m *MemoryProxyStore) Set(iou string, pgt string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.store[iou] = pgt
	return nil
}

var _ ProxyStore = &MemoryProxyStore{}
