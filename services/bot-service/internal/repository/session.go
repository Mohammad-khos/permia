package repository

import "sync"

// UserState represents the state of a user in a conversation.
type UserState int

const (
	StateNone UserState = iota
	StateWaitingForAmount
)

// SessionRepository defines the interface for a user session store.
type SessionRepository interface {
	SetState(userID int64, state UserState)
	GetState(userID int64) UserState
}

// InMemorySessionRepository is an in-memory implementation of SessionRepository.
type InMemorySessionRepository struct {
	states sync.Map
}

func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{}
}

func (r *InMemorySessionRepository) SetState(userID int64, state UserState) {
	r.states.Store(userID, state)
}

func (r *InMemorySessionRepository) GetState(userID int64) UserState {
	state, ok := r.states.Load(userID)
	if !ok {
		return StateNone
	}
	return state.(UserState)
}