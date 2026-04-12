package services

import (
	"encoding/json"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/toolx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var SkillDefinitionService = newSkillDefinitionService()

func newSkillDefinitionService() *skillDefinitionService {
	return &skillDefinitionService{}
}

type skillDefinitionService struct {
}

func (s *skillDefinitionService) Get(id int64) *models.SkillDefinition {
	return repositories.SkillDefinitionRepository.Get(sqls.DB(), id)
}

func (s *skillDefinitionService) Take(where ...interface{}) *models.SkillDefinition {
	return repositories.SkillDefinitionRepository.Take(sqls.DB(), where...)
}

func (s *skillDefinitionService) Find(cnd *sqls.Cnd) []models.SkillDefinition {
	return repositories.SkillDefinitionRepository.Find(sqls.DB(), cnd)
}

func (s *skillDefinitionService) FindOne(cnd *sqls.Cnd) *models.SkillDefinition {
	return repositories.SkillDefinitionRepository.FindOne(sqls.DB(), cnd)
}

func (s *skillDefinitionService) FindPageByParams(params *params.QueryParams) (list []models.SkillDefinition, paging *sqls.Paging) {
	return repositories.SkillDefinitionRepository.FindPageByParams(sqls.DB(), params)
}

func (s *skillDefinitionService) FindPageByCnd(cnd *sqls.Cnd) (list []models.SkillDefinition, paging *sqls.Paging) {
	return repositories.SkillDefinitionRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *skillDefinitionService) Count(cnd *sqls.Cnd) int64 {
	return repositories.SkillDefinitionRepository.Count(sqls.DB(), cnd)
}

func (s *skillDefinitionService) Create(t *models.SkillDefinition) error {
	return repositories.SkillDefinitionRepository.Create(sqls.DB(), t)
}

func (s *skillDefinitionService) Update(t *models.SkillDefinition) error {
	return repositories.SkillDefinitionRepository.Update(sqls.DB(), t)
}

func (s *skillDefinitionService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.SkillDefinitionRepository.Updates(sqls.DB(), id, columns)
}

func (s *skillDefinitionService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.SkillDefinitionRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *skillDefinitionService) Delete(id int64) {
	repositories.SkillDefinitionRepository.Delete(sqls.DB(), id)
}

func (s *skillDefinitionService) NextPriority() int {
	if max := repositories.SkillDefinitionRepository.FindOne(sqls.DB(), sqls.NewCnd().Desc("priority").Desc("id")); max != nil {
		return max.Priority + 1
	}
	return 1
}

func (s *skillDefinitionService) UpdatePriority(ids []int64) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for i, id := range ids {
			if err := repositories.SkillDefinitionRepository.UpdateColumn(ctx.Tx, id, "priority", i+1); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *skillDefinitionService) GetByCode(code string) *models.SkillDefinition {
	return repositories.SkillDefinitionRepository.GetByCode(sqls.DB(), code)
}

func (s *skillDefinitionService) CreateSkillDefinition(req request.CreateSkillDefinitionRequest, operator *dto.AuthPrincipal) (*models.SkillDefinition, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	normalized, err := s.normalizeSkillDefinitionRequest(req)
	if err != nil {
		return nil, err
	}
	if s.Take("code = ?", normalized.Code) != nil {
		return nil, errorsx.InvalidParam("Skill 编码已存在")
	}
	item := &models.SkillDefinition{
		Code:             normalized.Code,
		Name:             normalized.Name,
		Description:      normalized.Description,
		Instruction:      normalized.Instruction,
		Examples:         mustMarshalSkillStringArray(normalized.Examples),
		AllowedToolCodes: mustMarshalSkillStringArray(normalized.AllowedToolCodes),
		Priority:         normalized.Priority,
		Status:           enums.StatusOk,
		Remark:           normalized.Remark,
		AuditFields:      utils.BuildAuditFields(operator),
	}
	if item.Priority <= 0 {
		item.Priority = s.NextPriority()
	}
	if err := repositories.SkillDefinitionRepository.Create(sqls.DB(), item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *skillDefinitionService) UpdateSkillDefinition(req request.UpdateSkillDefinitionRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	if req.ID <= 0 {
		return errorsx.InvalidParam("Skill ID 不合法")
	}
	current := s.Get(req.ID)
	if current == nil {
		return errorsx.InvalidParam("Skill 不存在")
	}
	normalized, err := s.normalizeSkillDefinitionRequest(req.CreateSkillDefinitionRequest)
	if err != nil {
		return err
	}
	if exists := s.Take("code = ? AND id <> ?", normalized.Code, req.ID); exists != nil {
		return errorsx.InvalidParam("Skill 编码已存在")
	}
	return repositories.SkillDefinitionRepository.Updates(sqls.DB(), req.ID, map[string]any{
		"code":               normalized.Code,
		"name":               normalized.Name,
		"description":        normalized.Description,
		"content":            normalized.Instruction,
		"examples":           mustMarshalSkillStringArray(normalized.Examples),
		"allowed_tool_codes": mustMarshalSkillStringArray(normalized.AllowedToolCodes),
		"priority":           resolveSkillPriorityForService(normalized.Priority, current.Priority),
		"remark":             normalized.Remark,
		"update_user_id":     operator.UserID,
		"update_user_name":   operator.Username,
		"updated_at":         time.Now(),
	})
}

func (s *skillDefinitionService) normalizeSkillDefinitionRequest(req request.CreateSkillDefinitionRequest) (*request.CreateSkillDefinitionRequest, error) {
	normalized := &request.CreateSkillDefinitionRequest{
		Code:        strings.TrimSpace(req.Code),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Instruction: strings.TrimSpace(req.Instruction),
		Priority:    normalizeSkillPriorityForService(req.Priority),
		Remark:      strings.TrimSpace(req.Remark),
	}
	if normalized.Code == "" {
		return nil, errorsx.InvalidParam("Skill 编码不能为空")
	}
	if normalized.Name == "" {
		return nil, errorsx.InvalidParam("Skill 名称不能为空")
	}
	if normalized.Instruction == "" {
		return nil, errorsx.InvalidParam("技能说明不能为空")
	}
	examples, err := normalizeSkillStringArray(req.Examples)
	if err != nil {
		return nil, err
	}
	allowedToolCodes, err := normalizeSkillStringArray(req.AllowedToolCodes)
	if err != nil {
		return nil, err
	}
	for _, toolCode := range allowedToolCodes {
		if err := ToolCatalogService.ValidateMCPToolCode(toolCode); err != nil {
			return nil, err
		}
	}
	normalized.Examples = examples
	normalized.AllowedToolCodes = allowedToolCodes
	return normalized, nil
}

func normalizeSkillStringArray(input []string) ([]string, error) {
	buf, err := json.Marshal(input)
	if err != nil {
		return nil, errorsx.InvalidParam("JSON 数组格式不合法")
	}
	var ret []string
	if err := json.Unmarshal(buf, &ret); err != nil {
		return nil, errorsx.InvalidParam("JSON 数组格式不合法")
	}
	normalized := make([]string, 0, len(ret))
	seen := make(map[string]struct{}, len(ret))
	for _, item := range ret {
		item = strings.TrimSpace(item)
		item = toolx.NormalizeToolCodeAlias(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		normalized = append(normalized, item)
	}
	return normalized, nil
}

func mustMarshalSkillStringArray(input []string) string {
	items, err := normalizeSkillStringArray(input)
	if err != nil || len(items) == 0 {
		return "[]"
	}
	buf, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(buf)
}

func normalizeSkillPriorityForService(priority int) int {
	if priority < 0 {
		return 0
	}
	return priority
}

func resolveSkillPriorityForService(input, current int) int {
	input = normalizeSkillPriorityForService(input)
	if input <= 0 {
		return current
	}
	return input
}
