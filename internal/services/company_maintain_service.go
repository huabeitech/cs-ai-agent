package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"strings"
	"time"

	"github.com/mlogclub/simple/sqls"
)

func (s *companyService) CreateCompany(req request.CreateCompanyRequest, operator *dto.AuthPrincipal) (*models.Company, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errorsx.InvalidParam("公司名称不能为空")
	}

	existing := repositories.CompanyRepository.GetByName(sqls.DB(), name)
	if existing != nil && existing.Status != enums.StatusDeleted {
		return nil, errorsx.InvalidParam("公司名称已存在")
	}

	item := &models.Company{
		Name:        name,
		Code:        strings.TrimSpace(req.Code),
		Status:      enums.StatusOk,
		Remark:      strings.TrimSpace(req.Remark),
		AuditFields: utils.BuildAuditFields(operator),
	}
	if err := s.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *companyService) UpdateCompany(req request.UpdateCompanyRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(req.ID)
	if item == nil {
		return errorsx.InvalidParam("公司不存在")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errorsx.InvalidParam("公司名称不能为空")
	}

	existing := repositories.CompanyRepository.GetByName(sqls.DB(), name)
	if existing != nil && existing.ID != req.ID {
		return errorsx.InvalidParam("公司名称已存在")
	}

	now := time.Now()
	if err := s.Updates(req.ID, map[string]any{
		"name":             name,
		"code":             strings.TrimSpace(req.Code),
		"remark":           strings.TrimSpace(req.Remark),
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       now,
	}); err != nil {
		return err
	}
	return nil
}

func (s *companyService) DeleteCompany(id int64) error {
	item := s.Get(id)
	if item == nil {
		return errorsx.InvalidParam("公司不存在")
	}

	// 存在归属客户则不允许删除（避免客户“失联”）
	if CustomerService.Count(sqls.NewCnd().Eq("company_id", id)) > 0 {
		return errorsx.InvalidParam("该公司下存在客户，无法删除")
	}

	s.Delete(id)
	return nil
}

func (s *companyService) UpdateStatus(id int64, status int, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(id)
	if item == nil {
		return errorsx.InvalidParam("公司不存在")
	}
	if status != int(enums.StatusOk) && status != int(enums.StatusDisabled) {
		return errorsx.InvalidParam("状态值不合法")
	}
	now := time.Now()
	return s.Updates(id, map[string]any{
		"status":           status,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       now,
	})
}
