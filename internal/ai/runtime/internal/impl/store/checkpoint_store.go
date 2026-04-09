package store

import (
	"context"
	"encoding/base64"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/cloudwego/eino/adk"
	"github.com/mlogclub/simple/sqls"
)

var DefaultCheckPointStore adk.CheckPointStore = NewDBCheckPointStore()

type DBCheckPointStore struct{}

func NewDBCheckPointStore() *DBCheckPointStore {
	return &DBCheckPointStore{}
}

func (s *DBCheckPointStore) Get(_ context.Context, checkPointID string) ([]byte, bool, error) {
	item := repositories.ConversationInterruptRepository.GetByCheckPointID(sqls.DB(), checkPointID)
	if item == nil || item.CheckPointData == "" {
		return nil, false, nil
	}
	return decodeCheckPointData(item.CheckPointData)
}

func (s *DBCheckPointStore) Set(_ context.Context, checkPointID string, checkPoint []byte) error {
	item := repositories.ConversationInterruptRepository.GetByCheckPointID(sqls.DB(), checkPointID)
	if item == nil {
		item = buildEmptyInterrupt(checkPointID)
	}
	item.CheckPointData = encodeCheckPointData(checkPoint)
	if item.ConversationID == 0 && item.AIAgentID == 0 && item.SourceMessageID == 0 && item.Status == "" {
		return repositories.ConversationInterruptRepository.Create(sqls.DB(), item)
	}
	return repositories.ConversationInterruptRepository.UpsertByCheckPointID(sqls.DB(), item)
}

func encodeCheckPointData(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

func decodeCheckPointData(value string) ([]byte, bool, error) {
	if value == "" {
		return nil, false, nil
	}
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

func buildEmptyInterrupt(checkPointID string) *models.ConversationInterrupt {
	now := time.Now()
	return &models.ConversationInterrupt{
		CheckPointID: checkPointID,
		Status:       "checkpointed",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
