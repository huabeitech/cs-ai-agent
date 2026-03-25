package services

import "sync"

type WsConnectionManager struct {
	mu       sync.RWMutex
	sessions map[string]*ClientSession
	topics   map[string]map[string]*ClientSession
}

func newWsConnectionManager() *WsConnectionManager {
	return &WsConnectionManager{
		sessions: make(map[string]*ClientSession),
		topics:   make(map[string]map[string]*ClientSession),
	}
}

func (m *WsConnectionManager) Register(session *ClientSession, defaultTopics []string) int {
	if session == nil {
		return 0
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[session.ID] = session
	for _, topic := range defaultTopics {
		m.subscribeLocked(session, topic)
	}
	return len(m.sessions)
}

func (m *WsConnectionManager) Unregister(session *ClientSession) int {
	if session == nil {
		return 0
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, session.ID)
	for topic := range session.Topics {
		m.unsubscribeLocked(session, topic)
	}
	return len(m.sessions)
}

func (m *WsConnectionManager) Subscribe(session *ClientSession, topics []string) []string {
	if session == nil || len(topics) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	ret := make([]string, 0, len(topics))
	for _, topic := range topics {
		if _, exists := session.Topics[topic]; exists {
			continue
		}
		m.subscribeLocked(session, topic)
		ret = append(ret, topic)
	}
	return ret
}

func (m *WsConnectionManager) Unsubscribe(session *ClientSession, topics []string, keep map[string]struct{}) []string {
	if session == nil || len(topics) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	ret := make([]string, 0, len(topics))
	for _, topic := range topics {
		if _, isDefault := keep[topic]; isDefault {
			continue
		}
		if _, exists := session.Topics[topic]; !exists {
			continue
		}
		m.unsubscribeLocked(session, topic)
		ret = append(ret, topic)
	}
	return ret
}

func (m *WsConnectionManager) FindByTopics(topics []string) []*ClientSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uniq := make(map[string]*ClientSession)
	for _, topic := range topics {
		for connID, session := range m.topics[topic] {
			uniq[connID] = session
		}
	}

	ret := make([]*ClientSession, 0, len(uniq))
	for _, session := range uniq {
		ret = append(ret, session)
	}
	return ret
}

func (m *WsConnectionManager) subscribeLocked(session *ClientSession, topic string) {
	if _, exists := session.Topics[topic]; exists {
		return
	}
	if m.topics[topic] == nil {
		m.topics[topic] = make(map[string]*ClientSession)
	}
	m.topics[topic][session.ID] = session
	session.Topics[topic] = struct{}{}
}

func (m *WsConnectionManager) unsubscribeLocked(session *ClientSession, topic string) {
	if sessions, ok := m.topics[topic]; ok {
		delete(sessions, session.ID)
		if len(sessions) == 0 {
			delete(m.topics, topic)
		}
	}
	delete(session.Topics, topic)
}
