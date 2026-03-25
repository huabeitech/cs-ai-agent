package services

import (
	"testing"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"

	"github.com/mlogclub/simple/sqls"
)

func TestDeleteAgentTeamSoftDeletesAndAllowsReuseName(t *testing.T) {
	setupIMServiceTestDB(t)

	operator := &dto.AuthPrincipal{
		UserID:   1001,
		Username: "admin",
	}
	if err := sqls.DB().Create(&models.AgentTeam{
		ID:     1,
		Name:   "售前组",
		Status: enums.StatusOk,
		AuditFields: models.AuditFields{
			CreateUserID:   operator.UserID,
			CreateUserName: operator.Username,
			UpdateUserID:   operator.UserID,
			UpdateUserName: operator.Username,
		},
	}).Error; err != nil {
		t.Fatalf("seed agent team failed: %v", err)
	}

	if err := AgentTeamService.DeleteAgentTeam(1, operator); err != nil {
		t.Fatalf("delete agent team failed: %v", err)
	}

	saved := AgentTeamService.Get(1)
	if saved == nil {
		t.Fatal("expected agent team still exists after soft delete")
	}
	if saved.Status != enums.StatusDeleted {
		t.Fatalf("expected status deleted, got %d", saved.Status)
	}

	created, err := AgentTeamService.CreateAgentTeam(request.CreateAgentTeamRequest{
		Name:         "售前组",
		LeaderUserID: 0,
		Status:       int(enums.StatusOk),
		Description:  "new team",
		Remark:       "",
	}, operator)
	if err != nil {
		t.Fatalf("expected create same name after soft delete, got error: %v", err)
	}
	if created.ID == 1 {
		t.Fatalf("expected new team id, got reused id %d", created.ID)
	}
}

func TestDeleteAgentTeamRejectsBoundAIAgent(t *testing.T) {
	setupIMServiceTestDB(t)

	operator := &dto.AuthPrincipal{
		UserID:   1001,
		Username: "admin",
	}
	if err := sqls.DB().Create(&models.AgentTeam{
		ID:     1,
		Name:   "售后组",
		Status: enums.StatusOk,
		AuditFields: models.AuditFields{
			CreateUserID:   operator.UserID,
			CreateUserName: operator.Username,
			UpdateUserID:   operator.UserID,
			UpdateUserName: operator.Username,
		},
	}).Error; err != nil {
		t.Fatalf("seed agent team failed: %v", err)
	}
	if err := sqls.DB().Create(&models.AIAgent{
		ID:      2,
		Name:    "智能客服",
		Status:  enums.StatusOk,
		TeamIDs: "1",
		AuditFields: models.AuditFields{
			CreateUserID:   operator.UserID,
			CreateUserName: operator.Username,
			UpdateUserID:   operator.UserID,
			UpdateUserName: operator.Username,
		},
	}).Error; err != nil {
		t.Fatalf("seed ai agent failed: %v", err)
	}

	if err := AgentTeamService.DeleteAgentTeam(1, operator); err == nil {
		t.Fatal("expected delete agent team rejected when bound by ai agent")
	}
}
