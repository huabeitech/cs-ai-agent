package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/utils"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kataras/iris/v12"
)

var WsService = newWsService()

type wsService struct {
	upgrader websocket.Upgrader
	seq      atomic.Uint64
	manager  *WsConnectionManager
}

func newWsService() *wsService {
	return &wsService{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		manager: newWsConnectionManager(),
	}
}

func (s *wsService) UpgradeUserConnection(ctx iris.Context, principal *dto.AuthPrincipal) error {
	return s.upgradeConnection(ctx, principal, realtimeRoleUser)
}

func (s *wsService) UpgradeAdminConnection(ctx iris.Context, principal *dto.AuthPrincipal) error {
	return s.upgradeConnection(ctx, principal, realtimeRoleAdmin)
}

func (s *wsService) upgradeConnection(ctx iris.Context, principal *dto.AuthPrincipal, role string) error {
	conn, err := s.upgrader.Upgrade(ctx.ResponseWriter().Naive(), ctx.Request(), nil)
	if err != nil {
		return err
	}

	session := &ClientSession{
		ID:           s.nextID("conn"),
		Conn:         conn,
		Principal:    principal,
		Role:         role,
		TerminalType: s.resolveTerminalType(ctx, role),
		Topics:       make(map[string]struct{}),
		Send:         make(chan []byte, realtimeSendBufferSize),
	}
	session.touch()

	conn.SetReadLimit(realtimeMaxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(realtimePongWait))
	conn.SetPongHandler(func(string) error {
		session.touch()
		return conn.SetReadDeadline(time.Now().Add(realtimePongWait))
	})

	sessionCount := s.manager.Register(session, s.defaultTopics(session))
	slog.Info("realtime client connected",
		"connId", session.ID,
		"role", session.Role,
		"userId", session.Principal.UserID,
		"visitorId", session.Principal.VisitorID,
		"terminalType", session.TerminalType,
		"topicCount", len(session.Topics),
		"sessionCount", sessionCount,
	)

	go s.writePump(session)
	go s.readPump(session)

	session.enqueueEvent(s.newEvent("", RealtimeConnectedEvent{
		Payload: RealtimeConnectedPayload{
			ConnID:       session.ID,
			UserID:       principal.UserID,
			VisitorID:    principal.VisitorID,
			Role:         role,
			TerminalType: session.TerminalType,
			Topics:       session.topicList(),
		},
	}))
	return nil
}

func (s *wsService) readPump(session *ClientSession) {
	defer s.closeSession(session)

	for {
		_, body, err := session.Conn.ReadMessage()
		if err != nil {
			return
		}
		session.touch()

		input := realtimeClientMessage{}
		if err := json.Unmarshal(body, &input); err != nil {
			session.enqueueEvent(s.newEvent("", RealtimeResyncRequiredEvent{
				Payload: RealtimeResyncRequiredPayload{
					Reason: enums.IMRealtimeResyncReasonInvalidPayload,
				},
			}))
			continue
		}

		switch strings.TrimSpace(input.Type) {
		case enums.IMRealtimeClientTypePing:
			session.enqueueEvent(s.newEvent("", RealtimePongEvent{}))
		case enums.IMRealtimeClientTypeSubscribe:
			topics := s.subscribeTopics(session, input.Topics)
			if len(topics) > 0 {
				session.enqueueEvent(s.newEvent("", RealtimeSubscribedEvent{
					Payload: RealtimeTopicsPayload{Topics: topics},
				}))
			}
		case enums.IMRealtimeClientTypeUnsubscribe:
			topics := s.unsubscribeTopics(session, input.Topics)
			if len(topics) > 0 {
				session.enqueueEvent(s.newEvent("", RealtimeUnsubscribedEvent{
					Payload: RealtimeTopicsPayload{Topics: topics},
				}))
			}
		case enums.IMRealtimeClientTypeAck:
			slog.Debug("realtime event ack",
				"connId", session.ID,
				"eventId", strings.TrimSpace(input.EventID),
			)
		default:
			session.enqueueEvent(s.newEvent("", RealtimeResyncRequiredEvent{
				Payload: RealtimeResyncRequiredPayload{
					Reason: enums.IMRealtimeResyncReasonUnsupportedMessageType,
				},
			}))
		}
	}
}

func (s *wsService) writePump(session *ClientSession) {
	ticker := time.NewTicker(realtimePingPeriod)
	defer func() {
		ticker.Stop()
		s.closeSession(session)
	}()

	for {
		select {
		case payload, ok := <-session.Send:
			_ = session.Conn.SetWriteDeadline(time.Now().Add(realtimeWriteWait))
			if !ok {
				_ = session.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := session.Conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			_ = session.Conn.SetWriteDeadline(time.Now().Add(realtimeWriteWait))
			if err := session.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *wsService) closeSession(session *ClientSession) {
	if session == nil {
		return
	}
	session.closeOnce.Do(func() {
		session.Closed.Store(true)
		remaining := s.manager.Unregister(session)

		close(session.Send)
		_ = session.Conn.Close()

		slog.Info("realtime client disconnected",
			"connId", session.ID,
			"role", session.Role,
			"userId", session.Principal.UserID,
			"visitorId", session.Principal.VisitorID,
			"terminalType", session.TerminalType,
			"sessionCount", remaining,
		)
	})
}

func (s *wsService) subscribeTopics(session *ClientSession, topics []string) []string {
	allowed := s.filterAllowedTopics(session, topics)
	if len(allowed) == 0 {
		return nil
	}
	return s.manager.Subscribe(session, allowed)
}

func (s *wsService) unsubscribeTopics(session *ClientSession, topics []string) []string {
	allowed := s.filterAllowedTopics(session, topics)
	if len(allowed) == 0 {
		return nil
	}
	return s.manager.Unsubscribe(session, allowed, sliceToSet(s.defaultTopics(session)))
}

func (s *wsService) PublishMessageCreated(conversation *models.Conversation, message *models.Message) {
	if conversation == nil || message == nil {
		return
	}

	event := s.newEvent(s.conversationTopic(conversation.ID), RealtimeMessageCreatedEvent{
		Payload: RealtimeMessageCreatedPayload{
			ConversationID:    conversation.ID,
			MessageID:         message.ID,
			Status:            conversation.Status,
			CurrentAssigneeID: conversation.CurrentAssigneeID,
			SenderType:        message.SenderType,
			SenderID:          message.SenderID,
			MessageType:       message.MessageType,
			Content:           message.Content,
			Payload:           message.Payload,
			SeqNo:             message.SeqNo,
			SendStatus:        message.SendStatus,
			SentAt:            formatWsTime(message.SentAt),
		},
	})
	s.PublishToTopics(s.routeConversationTopics(conversation), event)
}

func (s *wsService) PublishConversationChanged(conversation *models.Conversation, eventType string) {
	if conversation == nil {
		return
	}
	agentReadState, customerReadState := ConversationReadStateService.GetConversationReadStates(conversation.ID)

	event := s.newEvent(s.conversationTopic(conversation.ID), RealtimeConversationChangedEvent{
		Type: eventType,
		Payload: RealtimeConversationChangedPayload{
			ConversationID:            conversation.ID,
			Status:                    conversation.Status,
			ServiceMode:               conversation.ServiceMode,
			CurrentAssigneeID:         conversation.CurrentAssigneeID,
			LastMessageID:             conversation.LastMessageID,
			LastMessageAt:             formatWsTime(conversation.LastMessageAt),
			LastMessageSummary:        conversation.LastMessageSummary,
			CustomerUnreadCount:       conversation.CustomerUnreadCount,
			AgentUnreadCount:          conversation.AgentUnreadCount,
			CustomerLastReadMessageID: readStateMessageID(customerReadState),
			CustomerLastReadSeqNo:     readStateSeqNo(customerReadState),
			CustomerLastReadAt:        readStateAt(customerReadState),
			AgentLastReadMessageID:    readStateMessageID(agentReadState),
			AgentLastReadSeqNo:        readStateSeqNo(agentReadState),
			AgentLastReadAt:           readStateAt(agentReadState),
		},
	})
	s.PublishToTopics(s.routeConversationTopics(conversation), event)
}

func (s *wsService) PublishResyncRequired(topics []string, reason string) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = enums.IMRealtimeResyncReasonManual
	}
	s.PublishToTopics(topics, s.newEvent("", RealtimeResyncRequiredEvent{
		Payload: RealtimeResyncRequiredPayload{Reason: reason},
	}))
}

func readStateMessageID(state *models.ConversationReadState) int64 {
	if state == nil {
		return 0
	}
	return state.LastReadMessageID
}

func readStateSeqNo(state *models.ConversationReadState) int64 {
	if state == nil {
		return 0
	}
	return state.LastReadSeqNo
}

func readStateAt(state *models.ConversationReadState) string {
	if state == nil {
		return ""
	}
	return utils.FormatTimePtr(state.LastReadAt)
}

func (s *wsService) Publish(event RealtimeEvent) {
	if strings.TrimSpace(event.Topic) == "" {
		return
	}
	s.PublishToTopic(event.Topic, event)
}

func (s *wsService) PublishToTopic(topic string, event RealtimeEvent) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return
	}
	if event.Topic == "" {
		event.Topic = topic
	}
	s.PublishToTopics([]string{topic}, event)
}

func (s *wsService) PublishToTopics(topics []string, event RealtimeEvent) {
	normalized := normalizeRealtimeTopics(topics)
	if len(normalized) == 0 {
		return
	}

	targets := s.manager.FindByTopics(normalized)
	if len(targets) == 0 {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		slog.Error("marshal realtime event failed", "error", err, "type", event.Type)
		return
	}

	for _, session := range targets {
		if session.enqueue(payload) {
			continue
		}
		slog.Warn("drop slow realtime client",
			"connId", session.ID,
			"role", session.Role,
			"type", event.Type,
			"topic", event.Topic,
		)
		go s.closeSession(session)
	}
}

func (s *wsService) routeConversationTopics(conversation *models.Conversation) []string {
	if conversation == nil {
		return nil
	}

	topics := []string{s.conversationTopic(conversation.ID)}
	if conversation.SourceUserID > 0 {
		topics = append(topics, s.userTopic(conversation.SourceUserID))
	}
	if strings.TrimSpace(conversation.ExternalUserID) != "" {
		topics = append(topics, s.visitorTopic(conversation.ExternalUserID))
	}
	if conversation.CurrentAssigneeID > 0 {
		topics = append(topics, s.adminTopic(conversation.CurrentAssigneeID))
	} else {
		topics = append(topics, realtimeTopicAdminAll)
	}
	return normalizeRealtimeTopics(topics)
}

func (s *wsService) defaultTopics(session *ClientSession) []string {
	if session == nil || session.Principal == nil {
		return nil
	}

	switch session.Role {
	case realtimeRoleAdmin:
		if session.Principal.UserID <= 0 {
			return []string{realtimeTopicAdminAll}
		}
		return []string{s.adminTopic(session.Principal.UserID), realtimeTopicAdminAll}
	default:
		if strings.TrimSpace(session.Principal.VisitorID) != "" {
			return []string{s.visitorTopic(session.Principal.VisitorID)}
		}
		if session.Principal.UserID > 0 {
			return []string{s.userTopic(session.Principal.UserID)}
		}
		return nil
	}
}

func (s *wsService) filterAllowedTopics(session *ClientSession, topics []string) []string {
	normalized := normalizeRealtimeTopics(topics)
	if len(normalized) == 0 || session == nil || session.Principal == nil {
		return nil
	}

	defaultTopics := sliceToSet(s.defaultTopics(session))
	ret := make([]string, 0, len(normalized))
	for _, topic := range normalized {
		if _, ok := defaultTopics[topic]; ok {
			ret = append(ret, topic)
			continue
		}
		if conversationID, ok := parseConversationTopic(topic); ok && s.canSubscribeConversation(session, conversationID) {
			ret = append(ret, topic)
		}
	}
	return ret
}

func (s *wsService) canSubscribeConversation(session *ClientSession, conversationID int64) bool {
	if session == nil || session.Principal == nil || conversationID <= 0 {
		return false
	}

	conversation := ConversationService.Get(conversationID)
	if conversation == nil {
		return false
	}
	if session.Role == realtimeRoleAdmin {
		return true
	}
	return ConversationService.IsCustomerConversationOwner(conversation, session.Principal)
}

func (s *wsService) resolveTerminalType(ctx iris.Context, role string) string {
	if ctx == nil {
		return "web"
	}
	terminalType := strings.TrimSpace(ctx.URLParam("terminalType"))
	if terminalType != "" {
		return terminalType
	}
	if role == realtimeRoleAdmin {
		return "console_web"
	}
	return "web"
}

func (s *wsService) newEvent(topic string, event RealtimeDomainEvent) RealtimeEvent {
	if event == nil {
		return RealtimeEvent{
			EventID: s.nextID("evt"),
			Topic:   topic,
			At:      time.Now().Format(time.DateTime),
		}
	}
	return RealtimeEvent{
		EventID: s.nextID("evt"),
		Type:    event.EventType(),
		Topic:   topic,
		Data:    event.EventPayload(),
		At:      time.Now().Format(time.DateTime),
	}
}

func (s *wsService) nextID(prefix string) string {
	seq := s.seq.Add(1)
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), seq)
}

func (s *wsService) userTopic(userID int64) string {
	return realtimeTopicUserPrefix + strconv.FormatInt(userID, 10)
}

func (s *wsService) visitorTopic(visitorID string) string {
	return realtimeTopicVisitorPrefix + strings.TrimSpace(visitorID)
}

func (s *wsService) adminTopic(userID int64) string {
	return realtimeTopicAdminPrefix + strconv.FormatInt(userID, 10)
}

func (s *wsService) conversationTopic(conversationID int64) string {
	return realtimeTopicConversationPrefix + strconv.FormatInt(conversationID, 10)
}

func normalizeRealtimeTopics(topics []string) []string {
	if len(topics) == 0 {
		return nil
	}
	ret := make([]string, 0, len(topics))
	seen := make(map[string]struct{}, len(topics))
	for _, topic := range topics {
		item := strings.TrimSpace(topic)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		ret = append(ret, item)
	}
	return ret
}

func parseConversationTopic(topic string) (int64, bool) {
	if !strings.HasPrefix(topic, realtimeTopicConversationPrefix) {
		return 0, false
	}
	value := strings.TrimPrefix(topic, realtimeTopicConversationPrefix)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func sliceToSet(items []string) map[string]struct{} {
	ret := make(map[string]struct{}, len(items))
	for _, item := range items {
		ret[item] = struct{}{}
	}
	return ret
}

func formatWsTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(time.DateTime)
}
