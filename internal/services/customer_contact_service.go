package services

import (
	"sort"
	"strings"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var CustomerContactService = newCustomerContactService()

func newCustomerContactService() *customerContactService {
	return &customerContactService{}
}

type customerContactService struct {
}

func (s *customerContactService) Get(id int64) *models.CustomerContact {
	return repositories.CustomerContactRepository.Get(sqls.DB(), id)
}

func (s *customerContactService) Take(where ...interface{}) *models.CustomerContact {
	return repositories.CustomerContactRepository.Take(sqls.DB(), where...)
}

func (s *customerContactService) Find(cnd *sqls.Cnd) []models.CustomerContact {
	return repositories.CustomerContactRepository.Find(sqls.DB(), cnd)
}

func (s *customerContactService) FindOne(cnd *sqls.Cnd) *models.CustomerContact {
	return repositories.CustomerContactRepository.FindOne(sqls.DB(), cnd)
}

func (s *customerContactService) FindPageByParams(params *params.QueryParams) (list []models.CustomerContact, paging *sqls.Paging) {
	return repositories.CustomerContactRepository.FindPageByParams(sqls.DB(), params)
}

func (s *customerContactService) FindPageByCnd(cnd *sqls.Cnd) (list []models.CustomerContact, paging *sqls.Paging) {
	return repositories.CustomerContactRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *customerContactService) Count(cnd *sqls.Cnd) int64 {
	return repositories.CustomerContactRepository.Count(sqls.DB(), cnd)
}

func (s *customerContactService) Create(t *models.CustomerContact) error {
	return repositories.CustomerContactRepository.Create(sqls.DB(), t)
}

func (s *customerContactService) Update(t *models.CustomerContact) error {
	return repositories.CustomerContactRepository.Update(sqls.DB(), t)
}

func (s *customerContactService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.CustomerContactRepository.Updates(sqls.DB(), id, columns)
}

func (s *customerContactService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.CustomerContactRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *customerContactService) Delete(id int64) {
	repositories.CustomerContactRepository.Delete(sqls.DB(), id)
}

// FindActiveByCustomerID 返回某客户下未删除的联系方式列表。
func (s *customerContactService) FindActiveByCustomerID(customerID int64) []models.CustomerContact {
	if customerID <= 0 {
		return nil
	}
	cnd := sqls.NewCnd().
		Where("customer_id = ?", customerID).
		Where("status <> ?", enums.StatusDeleted).
		Asc("id")
	return repositories.CustomerContactRepository.Find(sqls.DB(), cnd)
}

func normalizeContactSource(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "manual"
	}
	return v
}

func (s *customerContactService) hasDuplicateContact(
	db *gorm.DB,
	customerID int64,
	contactType enums.ContactType,
	contactValue string,
	excludeID int64,
) bool {
	cnd := sqls.NewCnd().
		Where("customer_id = ?", customerID).
		Where("contact_type = ?", contactType).
		Where("contact_value = ?", contactValue).
		Where("status <> ?", enums.StatusDeleted)
	if excludeID > 0 {
		cnd = cnd.Where("id <> ?", excludeID)
	}
	return repositories.CustomerContactRepository.FindOne(db, cnd) != nil
}

// syncCustomerPrimaryFromContacts 根据当前主联系方式更新客户表冗余字段（列表检索用）。
func (s *customerContactService) syncCustomerPrimaryFromContacts(db *gorm.DB, customerID int64) error {
	if customerID <= 0 {
		return nil
	}
	if repositories.CustomerRepository.Get(db, customerID) == nil {
		return nil
	}
	cnd := sqls.NewCnd().
		Where("customer_id = ?", customerID).
		Where("is_primary = ?", true).
		Where("status <> ?", enums.StatusDeleted)
	primary := repositories.CustomerContactRepository.FindOne(db, cnd)
	pm, pe := "", ""
	if primary != nil {
		val := strings.TrimSpace(primary.ContactValue)
		switch primary.ContactType {
		case enums.ContactTypeEmail:
			pe = val
		default:
			pm = val
		}
	}
	return repositories.CustomerRepository.Updates(db, customerID, map[string]any{
		"primary_mobile": pm,
		"primary_email":  pe,
		"updated_at":     time.Now(),
	})
}

// ReplaceAllForCustomerInTx 在事务内全量替换客户联系方式（软删未出现在 payload 中的记录），并同步客户主联系方式冗余字段。
func (s *customerContactService) ReplaceAllForCustomerInTx(
	ctx *sqls.TxContext,
	customerID int64,
	raw []request.CustomerProfileContactItem,
	operator *dto.AuthPrincipal,
) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	type line struct {
		id      *int64
		ct      enums.ContactType
		val     string
		remark  string
		primary bool
	}
	var items []line
	for _, r := range raw {
		ct := strings.TrimSpace(r.ContactType)
		val := strings.TrimSpace(r.ContactValue)
		if val == "" {
			continue
		}
		if !enums.IsValidContactType(ct) {
			return errorsx.InvalidParam("联系方式类型不合法")
		}
		items = append(items, line{
			id:      r.ID,
			ct:      enums.ContactType(ct),
			val:     val,
			remark:  strings.TrimSpace(r.Remark),
			primary: r.IsPrimary,
		})
	}
	if len(items) > 0 {
		primaryCount := 0
		for i := range items {
			if items[i].primary {
				primaryCount++
			}
		}
		if primaryCount == 0 {
			items[0].primary = true
		} else if primaryCount > 1 {
			return errorsx.InvalidParam("仅能指定一条主联系方式")
		}
	}

	existing := repositories.CustomerContactRepository.Find(ctx.Tx, sqls.NewCnd().
		Where("customer_id = ?", customerID).
		Where("status <> ?", enums.StatusDeleted).
		Asc("id"))

	wantIDs := map[int64]struct{}{}
	for i := range items {
		if items[i].id != nil && *items[i].id > 0 {
			wantIDs[*items[i].id] = struct{}{}
		}
	}
	now := time.Now()
	for _, ex := range existing {
		if _, ok := wantIDs[ex.ID]; !ok {
			if err := repositories.CustomerContactRepository.Updates(ctx.Tx, ex.ID, map[string]any{
				"status":           enums.StatusDeleted,
				"update_user_id":   operator.UserID,
				"update_user_name": operator.Username,
				"updated_at":       now,
			}); err != nil {
				return err
			}
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return !items[i].primary && items[j].primary
	})

	for _, it := range items {
		if it.id != nil && *it.id > 0 {
			row := repositories.CustomerContactRepository.Get(ctx.Tx, *it.id)
			if row == nil || row.CustomerID != customerID || row.Status == enums.StatusDeleted {
				return errorsx.InvalidParam("联系方式不存在")
			}
			if s.hasDuplicateContact(ctx.Tx, customerID, it.ct, it.val, *it.id) {
				return errorsx.InvalidParam("该联系方式已存在")
			}
			if it.primary {
				if err := s.clearPrimaryExcept(ctx.Tx, customerID, *it.id); err != nil {
					return err
				}
			}
			if err := repositories.CustomerContactRepository.Updates(ctx.Tx, *it.id, map[string]any{
				"contact_type":     it.ct,
				"contact_value":    it.val,
				"is_primary":       it.primary,
				"remark":           it.remark,
				"update_user_id":   operator.UserID,
				"update_user_name": operator.Username,
				"updated_at":       now,
			}); err != nil {
				return err
			}
			continue
		}
		if s.hasDuplicateContact(ctx.Tx, customerID, it.ct, it.val, 0) {
			return errorsx.InvalidParam("该联系方式已存在")
		}
		if it.primary {
			if err := s.clearPrimaryExcept(ctx.Tx, customerID, 0); err != nil {
				return err
			}
		}
		item := &models.CustomerContact{
			CustomerID:   customerID,
			ContactType:  it.ct,
			ContactValue: it.val,
			IsPrimary:    it.primary,
			IsVerified:   false,
			Source:       normalizeContactSource("manual"),
			Status:       enums.StatusOk,
			Remark:       it.remark,
			AuditFields:  utils.BuildAuditFields(operator),
		}
		if err := repositories.CustomerContactRepository.Create(ctx.Tx, item); err != nil {
			return err
		}
	}
	return s.syncCustomerPrimaryFromContacts(ctx.Tx, customerID)
}

func (s *customerContactService) clearPrimaryExcept(db *gorm.DB, customerID int64, exceptID int64) error {
	cnd := sqls.NewCnd().
		Where("customer_id = ?", customerID).
		Where("is_primary = ?", true)
	if exceptID > 0 {
		cnd = cnd.Where("id <> ?", exceptID)
	}
	list := repositories.CustomerContactRepository.Find(db, cnd)
	for i := range list {
		if err := repositories.CustomerContactRepository.UpdateColumn(db, list[i].ID, "is_primary", false); err != nil {
			return err
		}
	}
	return nil
}

func (s *customerContactService) validateContactStatus(status int) error {
	if !enums.IsValidStatus(status) {
		return errorsx.InvalidParam("状态值不合法")
	}
	if status == int(enums.StatusDeleted) {
		return errorsx.InvalidParam("状态值不合法")
	}
	return nil
}

// CreateCustomerContact 创建联系方式；主联系方式在同一客户下唯一。
func (s *customerContactService) CreateCustomerContact(req request.CreateCustomerContactRequest, operator *dto.AuthPrincipal) (*models.CustomerContact, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	if req.CustomerID <= 0 {
		return nil, errorsx.InvalidParam("客户不存在")
	}
	if CustomerService.Get(req.CustomerID) == nil {
		return nil, errorsx.InvalidParam("客户不存在")
	}
	ct := strings.TrimSpace(req.ContactType)
	if !enums.IsValidContactType(ct) {
		return nil, errorsx.InvalidParam("联系方式类型不合法")
	}
	val := strings.TrimSpace(req.ContactValue)
	if val == "" {
		return nil, errorsx.InvalidParam("联系方式不能为空")
	}
	if err := s.validateContactStatus(req.Status); err != nil {
		return nil, err
	}
	status := enums.Status(req.Status)
	if status == 0 {
		status = enums.StatusOk
	}

	var created *models.CustomerContact
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if s.hasDuplicateContact(ctx.Tx, req.CustomerID, enums.ContactType(ct), val, 0) {
			return errorsx.InvalidParam("该联系方式已存在")
		}
		if req.IsPrimary {
			if err := s.clearPrimaryExcept(ctx.Tx, req.CustomerID, 0); err != nil {
				return err
			}
		}
		now := time.Now()
		var verifiedAt *time.Time
		if req.IsVerified {
			verifiedAt = &now
		}
		item := &models.CustomerContact{
			CustomerID:   req.CustomerID,
			ContactType:  enums.ContactType(ct),
			ContactValue: val,
			IsPrimary:    req.IsPrimary,
			IsVerified:   req.IsVerified,
			VerifiedAt:   verifiedAt,
			Source:       normalizeContactSource(req.Source),
			Status:       status,
			Remark:       strings.TrimSpace(req.Remark),
			AuditFields:  utils.BuildAuditFields(operator),
		}
		if err := repositories.CustomerContactRepository.Create(ctx.Tx, item); err != nil {
			return err
		}
		created = item
		if err := s.syncCustomerPrimaryFromContacts(ctx.Tx, req.CustomerID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// UpdateCustomerContact 更新联系方式。
func (s *customerContactService) UpdateCustomerContact(req request.UpdateCustomerContactRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	if req.ID <= 0 {
		return errorsx.InvalidParam("联系方式不存在")
	}
	current := s.Get(req.ID)
	if current == nil {
		return errorsx.InvalidParam("联系方式不存在")
	}
	ct := strings.TrimSpace(req.ContactType)
	if !enums.IsValidContactType(ct) {
		return errorsx.InvalidParam("联系方式类型不合法")
	}
	val := strings.TrimSpace(req.ContactValue)
	if val == "" {
		return errorsx.InvalidParam("联系方式不能为空")
	}
	if err := s.validateContactStatus(req.Status); err != nil {
		return err
	}

	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		if s.hasDuplicateContact(ctx.Tx, current.CustomerID, enums.ContactType(ct), val, req.ID) {
			return errorsx.InvalidParam("该联系方式已存在")
		}
		if req.IsPrimary {
			if err := s.clearPrimaryExcept(ctx.Tx, current.CustomerID, req.ID); err != nil {
				return err
			}
		}
		now := time.Now()
		verifiedAt := current.VerifiedAt
		if req.IsVerified {
			if verifiedAt == nil {
				verifiedAt = &now
			}
		} else {
			verifiedAt = nil
		}
		if err := repositories.CustomerContactRepository.Updates(ctx.Tx, req.ID, map[string]any{
			"contact_type":     enums.ContactType(ct),
			"contact_value":    val,
			"is_primary":       req.IsPrimary,
			"is_verified":      req.IsVerified,
			"verified_at":      verifiedAt,
			"source":           normalizeContactSource(req.Source),
			"status":           req.Status,
			"remark":           strings.TrimSpace(req.Remark),
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       now,
		}); err != nil {
			return err
		}
		return s.syncCustomerPrimaryFromContacts(ctx.Tx, current.CustomerID)
	})
}

// DeleteCustomerContact 软删除联系方式并同步客户主联系方式冗余字段。
func (s *customerContactService) DeleteCustomerContact(id int64, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	if id <= 0 {
		return errorsx.InvalidParam("联系方式不存在")
	}
	current := s.Get(id)
	if current == nil {
		return errorsx.InvalidParam("联系方式不存在")
	}
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		now := time.Now()
		if err := repositories.CustomerContactRepository.Updates(ctx.Tx, id, map[string]any{
			"status":           enums.StatusDeleted,
			"update_user_id":   operator.UserID,
			"update_user_name": operator.Username,
			"updated_at":       now,
		}); err != nil {
			return err
		}
		return s.syncCustomerPrimaryFromContacts(ctx.Tx, current.CustomerID)
	})
}
