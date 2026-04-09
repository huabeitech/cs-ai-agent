package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var WxWorkKFConversationRepository = newWxWorkKFConversationRepository()

func newWxWorkKFConversationRepository() *wxWorkKFConversationRepository {
	return &wxWorkKFConversationRepository{}
}

type wxWorkKFConversationRepository struct {
}

func (r *wxWorkKFConversationRepository) Get(db *gorm.DB, id int64) *models.WxWorkKFConversation {
	ret := &models.WxWorkKFConversation{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *wxWorkKFConversationRepository) Take(db *gorm.DB, where ...interface{}) *models.WxWorkKFConversation {
	ret := &models.WxWorkKFConversation{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *wxWorkKFConversationRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.WxWorkKFConversation) {
	cnd.Find(db, &list)
	return
}

func (r *wxWorkKFConversationRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.WxWorkKFConversation {
	ret := &models.WxWorkKFConversation{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *wxWorkKFConversationRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.WxWorkKFConversation, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *wxWorkKFConversationRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.WxWorkKFConversation, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.WxWorkKFConversation{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *wxWorkKFConversationRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.WxWorkKFConversation) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *wxWorkKFConversationRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *wxWorkKFConversationRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.WxWorkKFConversation{})
}

func (r *wxWorkKFConversationRepository) Create(db *gorm.DB, t *models.WxWorkKFConversation) (err error) {
	err = db.Create(t).Error
	return
}

func (r *wxWorkKFConversationRepository) Update(db *gorm.DB, t *models.WxWorkKFConversation) (err error) {
	err = db.Save(t).Error
	return
}

func (r *wxWorkKFConversationRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.WxWorkKFConversation{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *wxWorkKFConversationRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.WxWorkKFConversation{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *wxWorkKFConversationRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.WxWorkKFConversation{}, "id = ?", id)
}

