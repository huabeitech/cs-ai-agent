package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var KnowledgeChunkRepository = newKnowledgeChunkRepository()

func newKnowledgeChunkRepository() *knowledgeChunkRepository {
	return &knowledgeChunkRepository{}
}

type knowledgeChunkRepository struct {
}

func (r *knowledgeChunkRepository) Get(db *gorm.DB, id int64) *models.KnowledgeChunk {
	ret := &models.KnowledgeChunk{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *knowledgeChunkRepository) Take(db *gorm.DB, where ...interface{}) *models.KnowledgeChunk {
	ret := &models.KnowledgeChunk{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *knowledgeChunkRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.KnowledgeChunk) {
	cnd.Find(db, &list)
	return
}

func (r *knowledgeChunkRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.KnowledgeChunk {
	ret := &models.KnowledgeChunk{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *knowledgeChunkRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.KnowledgeChunk, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *knowledgeChunkRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.KnowledgeChunk, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.KnowledgeChunk{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *knowledgeChunkRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.KnowledgeChunk{})
}

func (r *knowledgeChunkRepository) Create(db *gorm.DB, t *models.KnowledgeChunk) (err error) {
	err = db.Create(t).Error
	return
}

func (r *knowledgeChunkRepository) BatchCreate(db *gorm.DB, list []models.KnowledgeChunk) (err error) {
	err = db.Create(&list).Error
	return
}

func (r *knowledgeChunkRepository) Update(db *gorm.DB, t *models.KnowledgeChunk) (err error) {
	err = db.Save(t).Error
	return
}

func (r *knowledgeChunkRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.KnowledgeChunk{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *knowledgeChunkRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.KnowledgeChunk{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *knowledgeChunkRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.KnowledgeChunk{}, "id = ?", id)
}

func (r *knowledgeChunkRepository) DeleteByDocumentID(db *gorm.DB, documentID int64) {
	db.Delete(&models.KnowledgeChunk{}, "document_id = ?", documentID)
}

func (r *knowledgeChunkRepository) DeleteByFaqID(db *gorm.DB, faqID int64) {
	db.Delete(&models.KnowledgeChunk{}, "faq_id = ?", faqID)
}

func (r *knowledgeChunkRepository) FindByDocumentID(db *gorm.DB, documentID int64) (list []models.KnowledgeChunk) {
	db.Where("document_id = ?", documentID).Order("chunk_no asc").Find(&list)
	return
}

func (r *knowledgeChunkRepository) FindByFaqID(db *gorm.DB, faqID int64) (list []models.KnowledgeChunk) {
	db.Where("faq_id = ?", faqID).Order("chunk_no asc").Find(&list)
	return
}

func (r *knowledgeChunkRepository) CountByDocumentID(db *gorm.DB, documentID int64) int64 {
	var count int64
	db.Model(&models.KnowledgeChunk{}).Where("document_id = ?", documentID).Count(&count)
	return count
}

func (r *knowledgeChunkRepository) CountByFaqID(db *gorm.DB, faqID int64) int64 {
	var count int64
	db.Model(&models.KnowledgeChunk{}).Where("faq_id = ?", faqID).Count(&count)
	return count
}

func (r *knowledgeChunkRepository) CountByKnowledgeBaseID(db *gorm.DB, knowledgeBaseID int64) int64 {
	var count int64
	db.Model(&models.KnowledgeChunk{}).Where("knowledge_base_id = ?", knowledgeBaseID).Count(&count)
	return count
}
