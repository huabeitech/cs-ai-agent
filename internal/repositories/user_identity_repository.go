package repositories

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"strings"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var UserIdentityRepository = newUserIdentityRepository()

func newUserIdentityRepository() *userIdentityRepository {
	return &userIdentityRepository{}
}

type userIdentityRepository struct {
}

func (r *userIdentityRepository) Get(db *gorm.DB, id int64) *models.UserIdentity {
	ret := &models.UserIdentity{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *userIdentityRepository) Take(db *gorm.DB, where ...interface{}) *models.UserIdentity {
	ret := &models.UserIdentity{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *userIdentityRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.UserIdentity) {
	cnd.Find(db, &list)
	return
}

func (r *userIdentityRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.UserIdentity {
	ret := &models.UserIdentity{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *userIdentityRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.UserIdentity, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *userIdentityRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.UserIdentity, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.UserIdentity{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *userIdentityRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.UserIdentity) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *userIdentityRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *userIdentityRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.UserIdentity{})
}

func (r *userIdentityRepository) Create(db *gorm.DB, t *models.UserIdentity) (err error) {
	err = db.Create(t).Error
	return
}

func (r *userIdentityRepository) Update(db *gorm.DB, t *models.UserIdentity) (err error) {
	err = db.Save(t).Error
	return
}

func (r *userIdentityRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.UserIdentity{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *userIdentityRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.UserIdentity{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *userIdentityRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.UserIdentity{}, "id = ?", id)
}

func (r *userIdentityRepository) GetBy(db *gorm.DB, provider enums.ThirdProvider, corpId, userId string) *models.UserIdentity {
	return r.FindOne(db, sqls.NewCnd().
		Eq("provider", provider).
		Eq("provider_corp_id", strings.TrimSpace(corpId)).
		Eq("provider_user_id", strings.TrimSpace(userId)))
}
