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
	"github.com/mlogclub/simple/web/params"
)

var CompanyService = newCompanyService()

func newCompanyService() *companyService {
	return &companyService{}
}

type companyService struct {
}

func (s *companyService) Get(id int64) *models.Company {
	return repositories.CompanyRepository.Get(sqls.DB(), id)
}

func (s *companyService) Take(where ...interface{}) *models.Company {
	return repositories.CompanyRepository.Take(sqls.DB(), where...)
}

func (s *companyService) Find(cnd *sqls.Cnd) []models.Company {
	return repositories.CompanyRepository.Find(sqls.DB(), cnd)
}

func (s *companyService) FindOne(cnd *sqls.Cnd) *models.Company {
	return repositories.CompanyRepository.FindOne(sqls.DB(), cnd)
}

func (s *companyService) FindPageByParams(params *params.QueryParams) (list []models.Company, paging *sqls.Paging) {
	return repositories.CompanyRepository.FindPageByParams(sqls.DB(), params)
}

func (s *companyService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Company, paging *sqls.Paging) {
	return repositories.CompanyRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *companyService) Count(cnd *sqls.Cnd) int64 {
	return repositories.CompanyRepository.Count(sqls.DB(), cnd)
}

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
	if err := repositories.CompanyRepository.Create(sqls.DB(), item); err != nil {
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
	if err := repositories.CompanyRepository.Updates(sqls.DB(), req.ID, map[string]any{
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

	repositories.CompanyRepository.Delete(sqls.DB(), id)
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
	return repositories.CompanyRepository.Updates(sqls.DB(), id, map[string]any{
		"status":           status,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       now,
	})
}
