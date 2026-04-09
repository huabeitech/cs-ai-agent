package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var SkillRunLogRepository = newSkillRunLogRepository()

func newSkillRunLogRepository() *skillRunLogRepository {
	return &skillRunLogRepository{}
}

type skillRunLogRepository struct {
}

func (r *skillRunLogRepository) Get(db *gorm.DB, id int64) *models.SkillRunLog {
	ret := &models.SkillRunLog{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *skillRunLogRepository) Take(db *gorm.DB, where ...interface{}) *models.SkillRunLog {
	ret := &models.SkillRunLog{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *skillRunLogRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.SkillRunLog) {
	cnd.Find(db, &list)
	return
}

func (r *skillRunLogRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.SkillRunLog {
	ret := &models.SkillRunLog{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *skillRunLogRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.SkillRunLog, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *skillRunLogRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.SkillRunLog, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.SkillRunLog{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *skillRunLogRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (list []models.SkillRunLog) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *skillRunLogRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr... interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *skillRunLogRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.SkillRunLog{})
}

func (r *skillRunLogRepository) Create(db *gorm.DB, t *models.SkillRunLog) (err error) {
	err = db.Create(t).Error
	return
}

func (r *skillRunLogRepository) Update(db *gorm.DB, t *models.SkillRunLog) (err error) {
	err = db.Save(t).Error
	return
}

func (r *skillRunLogRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.SkillRunLog{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *skillRunLogRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.SkillRunLog{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *skillRunLogRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.SkillRunLog{}, "id = ?", id)
}

