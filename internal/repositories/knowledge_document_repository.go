package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var KnowledgeDocumentRepository = newKnowledgeDocumentRepository()

func newKnowledgeDocumentRepository() *knowledgeDocumentRepository {
	return &knowledgeDocumentRepository{}
}

type knowledgeDocumentRepository struct {
}

func (r *knowledgeDocumentRepository) Get(db *gorm.DB, id int64) *models.KnowledgeDocument {
	ret := &models.KnowledgeDocument{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *knowledgeDocumentRepository) Take(db *gorm.DB, where ...interface{}) *models.KnowledgeDocument {
	ret := &models.KnowledgeDocument{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *knowledgeDocumentRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.KnowledgeDocument) {
	cnd.Find(db, &list)
	return
}

func (r *knowledgeDocumentRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.KnowledgeDocument {
	ret := &models.KnowledgeDocument{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *knowledgeDocumentRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.KnowledgeDocument, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *knowledgeDocumentRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.KnowledgeDocument, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.KnowledgeDocument{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *knowledgeDocumentRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.KnowledgeDocument{})
}

func (r *knowledgeDocumentRepository) Create(db *gorm.DB, t *models.KnowledgeDocument) (err error) {
	err = db.Create(t).Error
	return
}

func (r *knowledgeDocumentRepository) Update(db *gorm.DB, t *models.KnowledgeDocument) (err error) {
	err = db.Save(t).Error
	return
}

func (r *knowledgeDocumentRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.KnowledgeDocument{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *knowledgeDocumentRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.KnowledgeDocument{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *knowledgeDocumentRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.KnowledgeDocument{}, "id = ?", id)
}

func (r *knowledgeDocumentRepository) CountByKnowledgeBaseID(db *gorm.DB, knowledgeBaseID int64) int64 {
	var count int64
	db.Model(&models.KnowledgeDocument{}).Where("knowledge_base_id = ?", knowledgeBaseID).Count(&count)
	return count
}
