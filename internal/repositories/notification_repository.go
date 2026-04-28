package repositories

import (
	"time"

	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var NotificationRepository = newNotificationRepository()

func newNotificationRepository() *notificationRepository {
	return &notificationRepository{}
}

type notificationRepository struct {
}

func (r *notificationRepository) Get(db *gorm.DB, id int64) *models.Notification {
	ret := &models.Notification{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *notificationRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.Notification) {
	cnd.Find(db, &list)
	return
}

func (r *notificationRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.Notification, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *notificationRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.Notification, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.Notification{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *notificationRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.Notification{})
}

func (r *notificationRepository) Create(db *gorm.DB, item *models.Notification) error {
	return db.Create(item).Error
}

func (r *notificationRepository) Updates(db *gorm.DB, id int64, columns map[string]any) error {
	return db.Model(&models.Notification{}).Where("id = ?", id).Updates(columns).Error
}

func (r *notificationRepository) MarkAllRead(db *gorm.DB, userID int64, readAt time.Time) error {
	return db.Model(&models.Notification{}).
		Where("recipient_user_id = ? AND read_at IS NULL", userID).
		Updates(map[string]any{"read_at": readAt}).Error
}
