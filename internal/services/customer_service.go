package services

import (
	"crypto/md5"
	"encoding/hex"
	"log/slog"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/openidentity"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"strings"
	"time"

	"github.com/mlogclub/simple/common/strs"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var CustomerService = newCustomerService()

func newCustomerService() *customerService {
	return &customerService{}
}

type customerService struct {
}

func (s *customerService) Get(id int64) *models.Customer {
	return repositories.CustomerRepository.Get(sqls.DB(), id)
}

func (s *customerService) Take(where ...interface{}) *models.Customer {
	return repositories.CustomerRepository.Take(sqls.DB(), where...)
}

func (s *customerService) Find(cnd *sqls.Cnd) []models.Customer {
	return repositories.CustomerRepository.Find(sqls.DB(), cnd)
}

func (s *customerService) FindOne(cnd *sqls.Cnd) *models.Customer {
	return repositories.CustomerRepository.FindOne(sqls.DB(), cnd)
}

func (s *customerService) FindPageByParams(params *params.QueryParams) (list []models.Customer, paging *sqls.Paging) {
	return repositories.CustomerRepository.FindPageByParams(sqls.DB(), params)
}

func (s *customerService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Customer, paging *sqls.Paging) {
	return repositories.CustomerRepository.FindPageByCnd(sqls.DB(), cnd)
}

// ListCustomers 客户分页列表（连联系方式表，支持按非主联系方式检索）。
func (s *customerService) ListCustomers(req request.CustomerListRequest) (list []models.Customer, paging *sqls.Paging) {
	if err := s.newCustomerListQuery(req).Distinct("c.*").Offset(req.Offset()).Order("c.id DESC").Limit(req.GetLimit()).Scan(&list).Error; err != nil {
		slog.Error("customer list scan failed", slog.Any("error", err))
	}

	var total int64
	if err := s.newCustomerListQuery(req).Distinct("c.id").Count(&total).Error; err != nil {
		slog.Error("customer list count failed", slog.Any("error", err))
	}

	paging = &sqls.Paging{
		Page:  req.GetPage(),
		Limit: req.GetLimit(),
		Total: total,
	}
	return
}

func (s *customerService) newCustomerListQuery(req request.CustomerListRequest) *gorm.DB {
	deleted := int(enums.StatusDeleted)
	tx := sqls.DB().
		Table("t_customer AS c").
		Joins("LEFT JOIN t_customer_contact AS cc ON cc.customer_id = c.id AND cc.status <> ?", deleted).
		Joins("LEFT JOIN t_company AS co ON co.id = c.company_id")

	tx.Where("c.status <> ?", enums.StatusDeleted)

	if req.Status != nil {
		tx.Where("c.status = ?", *req.Status)
	}
	if req.Gender != nil {
		tx.Where("c.gender = ?", *req.Gender)
	}
	if req.CompanyID != nil && *req.CompanyID > 0 {
		tx.Where("c.company_id = ?", *req.CompanyID)
	}
	if kw := strings.TrimSpace(req.Keyword); strs.IsNotBlank(kw) {
		pat := "%" + kw + "%"
		tx.Where(`(
c.name LIKE ? OR
c.primary_mobile LIKE ? OR
c.primary_email LIKE ? OR
cc.contact_value LIKE ? OR
co.name LIKE ?
)`, pat, pat, pat, pat, pat)
	}
	return tx
}

func (s *customerService) Count(cnd *sqls.Cnd) int64 {
	return repositories.CustomerRepository.Count(sqls.DB(), cnd)
}

func (s *customerService) CountByCompanyIDs(companyIDs []int64) map[int64]int64 {
	return repositories.CustomerRepository.CountByCompanyIDs(sqls.DB(), companyIDs, int(enums.StatusDeleted))
}

func (s *customerService) EnsureExternalCustomer(db *gorm.DB, externalInfo openidentity.ExternalInfo) (int64, error) {
	externalSource := externalInfo.ExternalSource
	externalID := strings.TrimSpace(externalInfo.ExternalID)
	if strings.TrimSpace(string(externalSource)) == "" || externalID == "" {
		return 0, errorsx.Unauthorized("外部用户标识不能为空")
	}
	now := time.Now()
	if identity := repositories.CustomerIdentityRepository.GetBy(db, externalSource, externalID); identity != nil {
		_ = repositories.CustomerRepository.Updates(db, identity.CustomerID, map[string]any{
			"last_active_at": now,
			"updated_at":     now,
		})
		return identity.CustomerID, nil
	}

	customer := &models.Customer{
		Name:         buildExternalCustomerName(externalInfo),
		LastActiveAt: &now,
		Status:       enums.StatusOk,
		AuditFields:  utils.BuildAuditFields(nil),
	}
	if err := repositories.CustomerRepository.Create(db, customer); err != nil {
		return 0, err
	}
	if err := repositories.CustomerIdentityRepository.Create(db, &models.CustomerIdentity{
		CustomerID:     customer.ID,
		ExternalSource: externalSource,
		ExternalID:     externalID,
		Status:         enums.StatusOk,
		AuditFields:    utils.BuildAuditFields(nil),
	}); err != nil {
		return 0, err
	}
	return customer.ID, nil
}

func buildExternalCustomerName(externalInfo openidentity.ExternalInfo) string {
	if strs.IsNotBlank(externalInfo.ExternalName) {
		return externalInfo.ExternalName
	}
	return "访客" + hashUUID(externalInfo.ExternalID)
}

func hashUUID(uuid string) string {
	if uuid == "" {
		return "unknown"
	}

	h := md5.Sum([]byte(uuid))
	return hex.EncodeToString(h[:])[:8]
}

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

	if err := repositories.CustomerRepository.Create(sqls.DB(), item); err != nil {
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

	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		now := time.Now()
		if err := repositories.CustomerRepository.Updates(ctx.Tx, req.ID, map[string]any{
			"name":             name,
			"gender":           req.Gender,
			"company_id":       req.CompanyID,
			"primary_mobile":   strings.TrimSpace(req.PrimaryMobile),
			"primary_email":    strings.TrimSpace(req.PrimaryEmail),
			"remark":           strings.TrimSpace(req.Remark),
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       now,
		}); err != nil {
			return err
		}
		return s.syncConversationCustomerName(ctx.Tx, req.ID, name, operator, now)
	})
}

func (s *customerService) DeleteCustomer(id int64, operator dto.AuthPrincipal) error {
	item := s.Get(id)
	if item == nil {
		return errorsx.InvalidParam("客户不存在")
	}
	return repositories.CustomerRepository.Updates(sqls.DB(), id, map[string]any{
		"status":           enums.StatusDeleted,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *customerService) syncConversationCustomerName(db *gorm.DB, customerID int64, name string, operator *dto.AuthPrincipal, now time.Time) error {
	if customerID <= 0 {
		return nil
	}
	updates := map[string]any{
		"customer_name": strings.TrimSpace(name),
		"updated_at":    now,
	}
	if operator != nil {
		updates["update_user_id"] = operator.UserID
		updates["update_user_name"] = operator.Username
	}
	return repositories.ConversationRepository.UpdatesByCustomerID(db, customerID, updates)
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
	return repositories.CustomerRepository.Updates(sqls.DB(), id, map[string]any{
		"status":           status,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

// SaveCustomerProfile 单事务保存客户主信息与联系方式全量（新建或更新）。
func (s *customerService) SaveCustomerProfile(req request.SaveCustomerProfileRequest, operator *dto.AuthPrincipal) (*models.Customer, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errorsx.InvalidParam("客户名称不能为空")
	}
	if req.CompanyID > 0 {
		if CompanyService.Get(req.CompanyID) == nil {
			return nil, errorsx.InvalidParam("所属公司不存在")
		}
	}
	createMode := req.ID == nil || *req.ID <= 0

	var out *models.Customer
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		var customerID int64
		if createMode {
			c := &models.Customer{
				Name:          name,
				Gender:        enums.Gender(req.Gender),
				CompanyID:     req.CompanyID,
				PrimaryMobile: "",
				PrimaryEmail:  "",
				Status:        enums.StatusOk,
				Remark:        strings.TrimSpace(req.Remark),
				AuditFields:   utils.BuildAuditFields(operator),
			}
			if err := repositories.CustomerRepository.Create(ctx.Tx, c); err != nil {
				return err
			}
			customerID = c.ID
			out = c
		} else {
			customerID = *req.ID
			cur := repositories.CustomerRepository.Get(ctx.Tx, customerID)
			if cur == nil {
				return errorsx.InvalidParam("客户不存在")
			}
			now := time.Now()
			if err := repositories.CustomerRepository.Updates(ctx.Tx, customerID, map[string]any{
				"name":             name,
				"gender":           req.Gender,
				"company_id":       req.CompanyID,
				"remark":           strings.TrimSpace(req.Remark),
				"update_user_id":   operator.UserID,
				"update_user_name": operator.Username,
				"updated_at":       now,
			}); err != nil {
				return err
			}
			if err := s.syncConversationCustomerName(ctx.Tx, customerID, name, operator, now); err != nil {
				return err
			}
			out = repositories.CustomerRepository.Get(ctx.Tx, customerID)
		}
		return CustomerContactService.ReplaceAllForCustomerInTx(ctx, customerID, req.Contacts, operator)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
