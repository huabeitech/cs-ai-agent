package repositories

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/enums"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var AgentTeamScheduleRepository = newAgentTeamScheduleRepository()

func newAgentTeamScheduleRepository() *agentTeamScheduleRepository {
	return &agentTeamScheduleRepository{}
}

type agentTeamScheduleRepository struct {
}

func (r *agentTeamScheduleRepository) Get(db *gorm.DB, id int64) *models.AgentTeamSchedule {
	ret := &models.AgentTeamSchedule{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *agentTeamScheduleRepository) Take(db *gorm.DB, where ...interface{}) *models.AgentTeamSchedule {
	ret := &models.AgentTeamSchedule{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *agentTeamScheduleRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.AgentTeamSchedule) {
	cnd.Find(db, &list)
	return
}

func (r *agentTeamScheduleRepository) FindByTimeRange(db *gorm.DB, startAt, endAt time.Time, teamID int64) (list []models.AgentTeamSchedule) {
	query := db.Model(&models.AgentTeamSchedule{}).
		Where("start_at < ? AND end_at > ?", endAt, startAt)
	if teamID > 0 {
		query = query.Where("team_id = ?", teamID)
	}
	query.Order("team_id ASC").Order("start_at ASC").Order("id ASC").Find(&list)
	return
}

func (r *agentTeamScheduleRepository) FindOverlappingByTeamIDsAndTimeRange(db *gorm.DB, teamIDs []int64, startAt, endAt time.Time) (list []models.AgentTeamSchedule) {
	if len(teamIDs) == 0 {
		return
	}
	db.Model(&models.AgentTeamSchedule{}).
		Where("team_id IN ? AND status = ? AND start_at < ? AND end_at > ?", teamIDs, enums.StatusOk, endAt, startAt).
		Order("team_id ASC").
		Order("start_at ASC").
		Order("id ASC").
		Find(&list)
	return
}

func (r *agentTeamScheduleRepository) CreateBatch(db *gorm.DB, list []models.AgentTeamSchedule) error {
	if len(list) == 0 {
		return nil
	}
	return db.Create(&list).Error
}

func (r *agentTeamScheduleRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.AgentTeamSchedule {
	ret := &models.AgentTeamSchedule{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *agentTeamScheduleRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.AgentTeamSchedule, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *agentTeamScheduleRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.AgentTeamSchedule, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.AgentTeamSchedule{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *agentTeamScheduleRepository) FindBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (list []models.AgentTeamSchedule) {
	db.Raw(sqlStr, paramArr...).Scan(&list)
	return
}

func (r *agentTeamScheduleRepository) CountBySql(db *gorm.DB, sqlStr string, paramArr ...interface{}) (count int64) {
	db.Raw(sqlStr, paramArr...).Count(&count)
	return
}

func (r *agentTeamScheduleRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.AgentTeamSchedule{})
}

func (r *agentTeamScheduleRepository) Create(db *gorm.DB, t *models.AgentTeamSchedule) (err error) {
	err = db.Create(t).Error
	return
}

func (r *agentTeamScheduleRepository) Update(db *gorm.DB, t *models.AgentTeamSchedule) (err error) {
	err = db.Save(t).Error
	return
}

func (r *agentTeamScheduleRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.AgentTeamSchedule{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *agentTeamScheduleRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.AgentTeamSchedule{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *agentTeamScheduleRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.AgentTeamSchedule{}, "id = ?", id)
}
