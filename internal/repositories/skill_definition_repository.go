package repositories

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var SkillDefinitionRepository = newSkillDefinitionRepository()

func newSkillDefinitionRepository() *skillDefinitionRepository {
	return &skillDefinitionRepository{}
}

type skillDefinitionRepository struct {
}

func (r *skillDefinitionRepository) Get(db *gorm.DB, id int64) *models.SkillDefinition {
	ret := &models.SkillDefinition{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *skillDefinitionRepository) Take(db *gorm.DB, where ...interface{}) *models.SkillDefinition {
	ret := &models.SkillDefinition{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *skillDefinitionRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.SkillDefinition) {
	cnd.Find(db, &list)
	return
}

func (r *skillDefinitionRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.SkillDefinition {
	ret := &models.SkillDefinition{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *skillDefinitionRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.SkillDefinition, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *skillDefinitionRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.SkillDefinition, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.SkillDefinition{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *skillDefinitionRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.SkillDefinition) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *skillDefinitionRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *skillDefinitionRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.SkillDefinition{})
}

func (r *skillDefinitionRepository) Create(db *gorm.DB, t *models.SkillDefinition) (err error) {
	err = db.Create(t).Error
	return
}

func (r *skillDefinitionRepository) Update(db *gorm.DB, t *models.SkillDefinition) (err error) {
	err = db.Save(t).Error
	return
}

func (r *skillDefinitionRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.SkillDefinition{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *skillDefinitionRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.SkillDefinition{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *skillDefinitionRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.SkillDefinition{}, "id = ?", id)
}

func (r *skillDefinitionRepository) GetByCode(db *gorm.DB, code string) *models.SkillDefinition {
	return r.FindOne(db, sqls.NewCnd().Where("code = ?", code))
}

func (r *skillDefinitionRepository) GetByIDs(db *gorm.DB, ids []int64) map[int64]models.SkillDefinition {
	if len(ids) == 0 {
		return nil
	}
	list := r.Find(db, sqls.NewCnd().Where("id IN (?)", ids))
	if len(list) == 0 {
		return nil
	}
	result := make(map[int64]models.SkillDefinition, len(list))
	for _, item := range list {
		result[item.ID] = item
	}
	return result
}
