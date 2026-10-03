package telegram

import "sync"

type ConversationStore struct {
	mu sync.RWMutex
	states map[int64]*ConversationState
}

func NewConversationStore() *ConversationStore {
	return &ConversationStore{
		states: make(map[int64]*ConversationState),
	}
}

func (s *ConversationStore) Get(chatID int64) (*ConversationState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.states[chatID]
	if !ok {
		return nil, false
	}

	return state, true
}

func (s *ConversationStore) Set(chatID int64, state *ConversationState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.states[chatID] = state
}

func (s *ConversationStore) Delete(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.states, chatID)
}