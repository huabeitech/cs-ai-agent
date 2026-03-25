package agentteam

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/constants"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/repositories"
	"fmt"
	"time"

	"github.com/mlogclub/simple/sqls"
	"golang.org/x/crypto/bcrypt"
)

type InitResult struct {
	TeamCreated     bool
	UsersCreated    int
	ProfilesCreated int
	UpdatesApplied  int
}

// Init 初始化客服组和客服用户
// 创建：
// 1. 客服组，组长为管理员用户
// 2. 客服A 用户
// 3. 客服B 用户
// 4. 为客服A和客服B创建客服档案，关联到该客服组
func Init() (*InitResult, error) {
	result := &InitResult{}

	// 获取管理员用户
	adminUser := repositories.UserRepository.Take(
		sqls.DB(),
		"username = ?",
		constants.BootstrapAdminUsername,
	)
	if adminUser == nil {
		return result, fmt.Errorf("bootstrap admin user not found")
	}

	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		return initTeamAndUsers(ctx, adminUser, result)
	})
	if err != nil {
		return result, fmt.Errorf("init team and users failed: %w", err)
	}

	return result, nil
}

func initTeamAndUsers(ctx *sqls.TxContext, leaderUser *models.User, result *InitResult) error {
	teamName := "默认客服组"
	now := time.Now()

	team := repositories.AgentTeamRepository.Take(ctx.Tx, "name = ?", teamName)
	if team != nil {
		if err := ctx.Tx.Model(team).Updates(map[string]any{
			"leader_user_id":   leaderUser.ID,
			"update_user_id":   constants.SystemAuditUserID,
			"update_user_name": constants.SystemAuditUserName,
			"updated_at":       now,
		}).Error; err != nil {
			return err
		}
	} else {
		team = &models.AgentTeam{
			Name:         teamName,
			LeaderUserID: leaderUser.ID,
			Status:       enums.StatusOk,
			Description:  "Local testdata seed - default service team",
			AuditFields: models.AuditFields{
				CreatedAt:      now,
				CreateUserID:   constants.SystemAuditUserID,
				CreateUserName: constants.SystemAuditUserName,
				UpdatedAt:      now,
				UpdateUserID:   constants.SystemAuditUserID,
				UpdateUserName: constants.SystemAuditUserName,
			},
		}
		if err := ctx.Tx.Create(team).Error; err != nil {
			return err
		}
		result.TeamCreated = true
	}

	agentUsers := []struct {
		username string
		nickname string
		code     string
	}{
		{
			username: leaderUser.Username,
			nickname: "客服组长",
			code:     "AGENT_LEADER_A",
		},
		{
			username: "agent_a",
			nickname: "客服A",
			code:     "AGENT_A",
		},
		{
			username: "agent_b",
			nickname: "客服B",
			code:     "AGENT_B",
		},
	}

	for _, agentUser := range agentUsers {
		userID, userCreated, err := createOrGetUser(ctx, agentUser.username, agentUser.nickname)
		if err != nil {
			return err
		}
		if userCreated {
			result.UsersCreated++
		}

		// 创建或更新客服档案
		profileCreated, err := createOrUpdateProfile(ctx, userID, team.ID, agentUser.code, agentUser.nickname)
		if err != nil {
			return err
		}
		if profileCreated {
			result.ProfilesCreated++
		} else {
			result.UpdatesApplied++
		}
	}

	return nil
}

func createOrGetUser(ctx *sqls.TxContext, username, nickname string) (int64, bool, error) {
	user := repositories.UserRepository.Take(
		ctx.Tx,
		"username = ?",
		username,
	)

	if user != nil {
		return user.ID, false, nil
	}

	// 创建新用户
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(constants.BootstrapAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, false, err
	}

	now := time.Now()
	newUser := &models.User{
		Username: username,
		Nickname: nickname,
		Password: string(hashedPassword),
		Status:   enums.StatusOk,
		Remark:   "Local testdata seed",
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   constants.SystemAuditUserID,
			CreateUserName: constants.SystemAuditUserName,
			UpdatedAt:      now,
			UpdateUserID:   constants.SystemAuditUserID,
			UpdateUserName: constants.SystemAuditUserName,
		},
	}

	if err := ctx.Tx.Create(newUser).Error; err != nil {
		return 0, false, err
	}

	return newUser.ID, true, nil
}

func createOrUpdateProfile(ctx *sqls.TxContext, userID, teamID int64, agentCode, displayName string) (bool, error) {
	profile := repositories.AgentProfileRepository.Take(
		ctx.Tx,
		"user_id = ?",
		userID,
	)

	now := time.Now()

	if profile != nil {
		// 档案已存在，更新关键信息
		return false, ctx.Tx.Model(profile).Updates(map[string]any{
			"team_id":          teamID,
			"agent_code":       agentCode,
			"display_name":     displayName,
			"status":           enums.StatusOk,
			"update_user_id":   constants.SystemAuditUserID,
			"update_user_name": constants.SystemAuditUserName,
			"updated_at":       now,
		}).Error
	}

	// 创建新档案
	newProfile := &models.AgentProfile{
		UserID:             userID,
		TeamID:             teamID,
		AgentCode:          agentCode,
		DisplayName:        displayName,
		Avatar:             "",
		ServiceStatus:      enums.ServiceStatusIdle,
		MaxConcurrentCount: 5,
		PriorityLevel:      10,
		AutoAssignEnabled:  true,
		Status:             enums.StatusOk,
		Remark:             "Local testdata seed",
		AuditFields: models.AuditFields{
			CreatedAt:      now,
			CreateUserID:   constants.SystemAuditUserID,
			CreateUserName: constants.SystemAuditUserName,
			UpdatedAt:      now,
			UpdateUserID:   constants.SystemAuditUserID,
			UpdateUserName: constants.SystemAuditUserName,
		},
	}

	return true, ctx.Tx.Create(newProfile).Error
}
