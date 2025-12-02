package repository

import (
	"sync"
	
	"Permia/bot-service/internal/domain"
)

// SessionRepository defines the interface for a user session store.
type SessionRepository interface {
	SetState(userID int64, state domain.UserState)
	GetState(userID int64) domain.UserState
}

// InMemorySessionRepository is an in-memory implementation of SessionRepository.
type InMemorySessionRepository struct {
	states sync.Map
}

func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{}
}

func (r *InMemorySessionRepository) SetState(userID int64, state domain.UserState) {
	r.states.Store(userID, state)
}

func (r *InMemorySessionRepository) GetState(userID int64) domain.UserState {
	state, ok := r.states.Load(userID)
	if !ok {
		return domain.StateNone
	}
	return state.(domain.UserState)
}