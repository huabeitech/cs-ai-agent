package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var KnowledgeFAQRepository = newKnowledgeFAQRepository()

func newKnowledgeFAQRepository() *knowledgeFAQRepository {
	return &knowledgeFAQRepository{}
}

type knowledgeFAQRepository struct{}

func (r *knowledgeFAQRepository) Get(db *gorm.DB, id int64) *models.KnowledgeFAQ {
	ret := &models.KnowledgeFAQ{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *knowledgeFAQRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.KnowledgeFAQ) {
	cnd.Find(db, &list)
	return
}

func (r *knowledgeFAQRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.KnowledgeFAQ, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.KnowledgeFAQ{})
	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *knowledgeFAQRepository) FindPageByParams(db *gorm.DB, queryParams *params.QueryParams) (list []models.KnowledgeFAQ, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &queryParams.Cnd)
}

func (r *knowledgeFAQRepository) Create(db *gorm.DB, t *models.KnowledgeFAQ) error {
	return db.Create(t).Error
}

func (r *knowledgeFAQRepository) Updates(db *gorm.DB, id int64, columns map[string]any) error {
	return db.Model(&models.KnowledgeFAQ{}).Where("id = ?", id).Updates(columns).Error
}

func (r *knowledgeFAQRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.KnowledgeFAQ{}, "id = ?", id)
}

func (r *knowledgeFAQRepository) FindByIDs(db *gorm.DB, ids []int64) (list []models.KnowledgeFAQ) {
	if len(ids) == 0 {
		return nil
	}
	db.Where("id IN ?", ids).Find(&list)
	return
}

func (r *knowledgeFAQRepository) CountByKnowledgeBaseID(db *gorm.DB, knowledgeBaseID int64) int64 {
	var count int64
	db.Model(&models.KnowledgeFAQ{}).Where("knowledge_base_id = ?", knowledgeBaseID).Count(&count)
	return count
}
