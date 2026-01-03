package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/buildwithgo/amaro/addons/cache"
)

// Session holds the session data.
// T is the type of data stored in the session.
type Session[T any] struct {
	ID    string
	Data  T
	store Store[T]
	isNew bool
}

// Get retrieves a value if T is a map[string]interface{}.
// This provides backward compatibility and dynamic access.
func (s *Session[T]) Get(key string) interface{} {
	if any(s.Data) == nil {
		return nil
	}
	if m, ok := any(s.Data).(map[string]interface{}); ok {
		return m[key]
	}
	return nil
}

// Set stores a value if T is a map[string]interface{}.
// If s.Data is nil or a nil map, it initializes it.
func (s *Session[T]) Set(key string, value interface{}) {
	var m map[string]interface{}

	// Check if T is map[string]interface{}
	if existing, ok := any(s.Data).(map[string]interface{}); ok {
		if existing == nil {
			m = make(map[string]interface{})
		} else {
			m = existing
		}
	} else if any(s.Data) == nil {
		// T is likely interface{} or map (but nil value).
		// Try to create map and see if it fits T.
		m = make(map[string]interface{})
		if _, ok := any(m).(T); !ok {
			return // T doesn't support map
		}
	} else {
		return // T is likely a struct (not a map)
	}

	m[key] = value
	s.Data = any(m).(T)
}

// Save persists the session to the store.
func (s *Session[T]) Save() error {
	return s.store.Save(s)
}

// Store (Provider) interface.
type Store[T any] interface {
	Get(id string) (*Session[T], error)
	Save(session *Session[T]) error
	Delete(id string) error
	NewSession() *Session[T]
	CookieConfig() (name string, ttl time.Duration)
}

// Provider represents the same interface as Store.
type Provider[T any] interface {
	Store[T]
}

// Manager manages sessions.
type Manager[T any] struct {
	cookieName string
	ttl        time.Duration
	// Backend cache is now ANY type
	cache cache.Cache
}

// New creates a new session manager with map[string]interface{} as the data type.
// This is a helper for the most common use case.
func New(cache cache.Cache, cookieName string, ttl time.Duration) *Manager[map[string]interface{}] {
	return NewManager[map[string]interface{}](cache, cookieName, ttl)
}

// NewManager creates a new session manager using Any Cache backend.
func NewManager[T any](cache cache.Cache, cookieName string, ttl time.Duration) *Manager[T] {
	return &Manager[T]{
		cache:      cache,
		cookieName: cookieName,
		ttl:        ttl,
	}
}

// CookieConfig returns the cookie configuration.
func (m *Manager[T]) CookieConfig() (string, time.Duration) {
	return m.cookieName, m.ttl
}

// Get returns the session stored in the cache.
func (m *Manager[T]) Get(id string) (*Session[T], error) {
	val, ok := m.cache.Get(id)
	if !ok {
		return m.NewSession(), nil
	}

	// Assert that retrieved value is T
	data, ok := val.(T)
	if !ok {
		// If type mismatch in cache, treat as new session (safe fallback)
		return m.NewSession(), nil
	}

	return &Session[T]{
		ID:    id,
		Data:  data,
		store: m,
		isNew: false,
	}, nil
}

// Save persists the session.
func (m *Manager[T]) Save(s *Session[T]) error {
	// Cache accepts interface{}, so we pass T directly
	m.cache.Set(s.ID, s.Data, m.ttl)
	return nil
}

// Delete deletes the session.
func (m *Manager[T]) Delete(id string) error {
	m.cache.Delete(id)
	return nil
}

func (m *Manager[T]) NewSession() *Session[T] {
	var data T
	return &Session[T]{
		ID:    base64.URLEncoding.EncodeToString(generateRandomBytes(32)),
		Data:  data,
		store: m,
		isNew: true,
	}
}

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	rand.Read(b)
	return b
}
