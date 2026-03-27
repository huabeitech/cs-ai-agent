package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"strings"
	"time"
)

func (s *customerService) CreateCustomer(req request.CreateCustomerRequest, operator *dto.AuthPrincipal) (*models.Customer, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errorsx.InvalidParam("客户名称不能为空")
	}

	if req.CompanyID > 0 {
		company := CompanyService.Get(req.CompanyID)
		if company == nil {
			return nil, errorsx.InvalidParam("所属公司不存在")
		}
	}

	item := &models.Customer{
		Name:          name,
		Gender:        enums.Gender(req.Gender),
		CompanyID:     req.CompanyID,
		PrimaryMobile: strings.TrimSpace(req.PrimaryMobile),
		PrimaryEmail:  strings.TrimSpace(req.PrimaryEmail),
		Status:        enums.StatusOk,
		Remark:        strings.TrimSpace(req.Remark),
		AuditFields:   utils.BuildAuditFields(operator),
	}

	if err := s.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *customerService) UpdateCustomer(req request.UpdateCustomerRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(req.ID)
	if item == nil {
		return errorsx.InvalidParam("客户不存在")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errorsx.InvalidParam("客户名称不能为空")
	}

	if req.CompanyID > 0 {
		company := CompanyService.Get(req.CompanyID)
		if company == nil {
			return errorsx.InvalidParam("所属公司不存在")
		}
	}

	return s.Updates(req.ID, map[string]any{
		"name":             name,
		"gender":           req.Gender,
		"company_id":       req.CompanyID,
		"primary_mobile":   strings.TrimSpace(req.PrimaryMobile),
		"primary_email":    strings.TrimSpace(req.PrimaryEmail),
		"remark":           strings.TrimSpace(req.Remark),
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *customerService) DeleteCustomer(id int64) error {
	item := s.Get(id)
	if item == nil {
		return errorsx.InvalidParam("客户不存在")
	}
	s.Delete(id)
	return nil
}

func (s *customerService) UpdateStatus(id int64, status int, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(id)
	if item == nil {
		return errorsx.InvalidParam("客户不存在")
	}
	if status != int(enums.StatusOk) && status != int(enums.StatusDisabled) {
		return errorsx.InvalidParam("状态值不合法")
	}
	return s.Updates(id, map[string]any{
		"status":           status,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}
