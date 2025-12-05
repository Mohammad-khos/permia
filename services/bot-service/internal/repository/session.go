package repository

import (
	"sync"
	
	"Permia/bot-service/internal/domain"
)

// SessionRepository defines the interface for a user session store.
type SessionRepository interface {
	SetState(userID int64, state domain.UserState)
	GetState(userID int64) domain.UserState
	SetDraft(userID int64, key string, value string)
	GetDraft(userID int64, key string) string
	ClearDraft(userID int64)
}

// InMemorySessionRepository is an in-memory implementation of SessionRepository.
type InMemorySessionRepository struct {
	states map[int64]domain.UserState
	drafts map[int64]map[string]string
	mu     sync.RWMutex
}

func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		states: make(map[int64]domain.UserState),
		drafts: make(map[int64]map[string]string),
	}
}

func (r *InMemorySessionRepository) SetState(userID int64, state domain.UserState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states[userID] = state
}

func (r *InMemorySessionRepository) GetState(userID int64) domain.UserState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.states[userID]
}
// پیاده‌سازی متدهای Draft
func (r *InMemorySessionRepository) SetDraft(userID int64, key string, value string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.drafts[userID] == nil {
		r.drafts[userID] = make(map[string]string)
	}
	r.drafts[userID][key] = value
}

func (r *InMemorySessionRepository) GetDraft(userID int64, key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if data, ok := r.drafts[userID]; ok {
		return data[key]
	}
	return ""
}

func (r *InMemorySessionRepository) ClearDraft(userID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.drafts, userID)
}