package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var KnowledgeBaseRepository = newKnowledgeBaseRepository()

func newKnowledgeBaseRepository() *knowledgeBaseRepository {
	return &knowledgeBaseRepository{}
}

type knowledgeBaseRepository struct {
}

func (r *knowledgeBaseRepository) Get(db *gorm.DB, id int64) *models.KnowledgeBase {
	ret := &models.KnowledgeBase{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *knowledgeBaseRepository) Take(db *gorm.DB, where ...interface{}) *models.KnowledgeBase {
	ret := &models.KnowledgeBase{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *knowledgeBaseRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.KnowledgeBase) {
	cnd.Find(db, &list)
	return
}

func (r *knowledgeBaseRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.KnowledgeBase {
	ret := &models.KnowledgeBase{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *knowledgeBaseRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.KnowledgeBase, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *knowledgeBaseRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.KnowledgeBase, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.KnowledgeBase{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *knowledgeBaseRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.KnowledgeBase{})
}

func (r *knowledgeBaseRepository) Create(db *gorm.DB, t *models.KnowledgeBase) (err error) {
	err = db.Create(t).Error
	return
}

func (r *knowledgeBaseRepository) Update(db *gorm.DB, t *models.KnowledgeBase) (err error) {
	err = db.Save(t).Error
	return
}

func (r *knowledgeBaseRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.KnowledgeBase{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *knowledgeBaseRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.KnowledgeBase{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *knowledgeBaseRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.KnowledgeBase{}, "id = ?", id)
}
