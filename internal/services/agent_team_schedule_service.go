package services

import (
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/pkg/errorsx"
	"cs-agent/internal/pkg/utils"
	"cs-agent/internal/repositories"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var AgentTeamScheduleService = newAgentTeamScheduleService()

func newAgentTeamScheduleService() *agentTeamScheduleService {
	return &agentTeamScheduleService{}
}

type agentTeamScheduleService struct {
	writeMu sync.Mutex
}

const maxAgentTeamScheduleBatchItems = 500

type AgentTeamScheduleBatchPreviewResult struct {
	Total    int
	Conflict bool
	Items    []AgentTeamScheduleBatchPreviewItem
}

type AgentTeamScheduleBatchPreviewItem struct {
	TeamID         int64
	TeamName       string
	Date           time.Time
	Weekday        int
	StartAt        time.Time
	EndAt          time.Time
	Remark         string
	Conflict       bool
	ConflictReason string
}

type AgentTeamScheduleBatchGenerateResult struct {
	Created int
}

type batchScheduleCandidate struct {
	TeamID   int64
	TeamName string
	Date     time.Time
	StartAt  time.Time
	EndAt    time.Time
	Remark   string
}

func (s *agentTeamScheduleService) Get(id int64) *models.AgentTeamSchedule {
	return repositories.AgentTeamScheduleRepository.Get(sqls.DB(), id)
}

func (s *agentTeamScheduleService) Take(where ...interface{}) *models.AgentTeamSchedule {
	return repositories.AgentTeamScheduleRepository.Take(sqls.DB(), where...)
}

func (s *agentTeamScheduleService) Find(cnd *sqls.Cnd) []models.AgentTeamSchedule {
	return repositories.AgentTeamScheduleRepository.Find(sqls.DB(), cnd)
}

func (s *agentTeamScheduleService) FindOne(cnd *sqls.Cnd) *models.AgentTeamSchedule {
	return repositories.AgentTeamScheduleRepository.FindOne(sqls.DB(), cnd)
}

func (s *agentTeamScheduleService) FindPageByParams(params *params.QueryParams) (list []models.AgentTeamSchedule, paging *sqls.Paging) {
	return repositories.AgentTeamScheduleRepository.FindPageByParams(sqls.DB(), params)
}

func (s *agentTeamScheduleService) FindPageByCnd(cnd *sqls.Cnd) (list []models.AgentTeamSchedule, paging *sqls.Paging) {
	return repositories.AgentTeamScheduleRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *agentTeamScheduleService) Count(cnd *sqls.Cnd) int64 {
	return repositories.AgentTeamScheduleRepository.Count(sqls.DB(), cnd)
}

func (s *agentTeamScheduleService) FindCalendarSchedules(req request.AgentTeamScheduleCalendarRequest) ([]models.AgentTeamSchedule, error) {
	startAtValue, err := parseRequiredDateTime(req.StartAt, "开始时间格式错误")
	if err != nil {
		return nil, err
	}
	endAtValue, err := parseRequiredDateTime(req.EndAt, "结束时间格式错误")
	if err != nil {
		return nil, err
	}
	if !endAtValue.After(startAtValue) {
		return nil, errorsx.InvalidParam("结束时间必须晚于开始时间")
	}
	return repositories.AgentTeamScheduleRepository.FindByTimeRange(sqls.DB(), startAtValue, endAtValue, req.TeamID), nil
}

func (s *agentTeamScheduleService) Create(t *models.AgentTeamSchedule) error {
	return repositories.AgentTeamScheduleRepository.Create(sqls.DB(), t)
}

func (s *agentTeamScheduleService) Update(t *models.AgentTeamSchedule) error {
	return repositories.AgentTeamScheduleRepository.Update(sqls.DB(), t)
}

func (s *agentTeamScheduleService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.AgentTeamScheduleRepository.Updates(sqls.DB(), id, columns)
}

func (s *agentTeamScheduleService) UpdateColumn(id int64, name string, value interface{}) error {
	return repositories.AgentTeamScheduleRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *agentTeamScheduleService) Delete(id int64) {
	repositories.AgentTeamScheduleRepository.Delete(sqls.DB(), id)
}

func (s *agentTeamScheduleService) CreateAgentTeamSchedule(req request.CreateAgentTeamScheduleRequest, operator *dto.AuthPrincipal) (*models.AgentTeamSchedule, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	s.writeMu.Lock()
	item, err := s.buildScheduleModel(0, req.TeamID, req.StartAt, req.EndAt, req.Remark)
	if err != nil {
		s.writeMu.Unlock()
		return nil, err
	}
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := repositories.AgentTeamScheduleRepository.Create(sqls.DB(), item); err != nil {
		s.writeMu.Unlock()
		return nil, err
	}
	s.writeMu.Unlock()
	s.dispatchPendingConversationsIfActive(item)
	return item, nil
}

func (s *agentTeamScheduleService) UpdateAgentTeamSchedule(req request.UpdateAgentTeamScheduleRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	s.writeMu.Lock()
	if s.Get(req.ID) == nil {
		s.writeMu.Unlock()
		return errorsx.InvalidParam("客服组排班不存在")
	}
	item, err := s.buildScheduleModel(req.ID, req.TeamID, req.StartAt, req.EndAt, req.Remark)
	if err != nil {
		s.writeMu.Unlock()
		return err
	}
	if err := repositories.AgentTeamScheduleRepository.Updates(sqls.DB(), req.ID, map[string]any{
		"team_id":          item.TeamID,
		"start_at":         item.StartAt,
		"end_at":           item.EndAt,
		"remark":           item.Remark,
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	}); err != nil {
		s.writeMu.Unlock()
		return err
	}
	s.writeMu.Unlock()
	s.dispatchPendingConversationsIfActive(item)
	return nil
}

func (s *agentTeamScheduleService) DeleteAgentTeamSchedule(id int64) error {
	if s.Get(id) == nil {
		return errorsx.InvalidParam("客服组排班不存在")
	}
	repositories.AgentTeamScheduleRepository.Delete(sqls.DB(), id)
	return nil
}

func (s *agentTeamScheduleService) BatchPreview(req request.AgentTeamScheduleBatchRequest, operator *dto.AuthPrincipal) (*AgentTeamScheduleBatchPreviewResult, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	candidates, err := s.buildBatchScheduleCandidates(req)
	if err != nil {
		return nil, err
	}
	conflicts := s.findBatchConflict(candidates)
	return buildBatchPreviewResult(candidates, conflicts), nil
}

func (s *agentTeamScheduleService) BatchGenerate(req request.AgentTeamScheduleBatchRequest, operator *dto.AuthPrincipal) (*AgentTeamScheduleBatchGenerateResult, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	s.writeMu.Lock()
	candidates, err := s.buildBatchScheduleCandidates(req)
	if err != nil {
		s.writeMu.Unlock()
		return nil, err
	}
	conflicts := s.findBatchConflict(candidates)
	for _, conflict := range conflicts {
		if conflict != "" {
			s.writeMu.Unlock()
			return nil, errorsx.InvalidParam("存在冲突排班，请先处理冲突")
		}
	}

	schedules := make([]models.AgentTeamSchedule, 0, len(candidates))
	for _, candidate := range candidates {
		schedules = append(schedules, models.AgentTeamSchedule{
			TeamID:      candidate.TeamID,
			StartAt:     candidate.StartAt,
			EndAt:       candidate.EndAt,
			Remark:      candidate.Remark,
			Status:      enums.StatusOk,
			AuditFields: utils.BuildAuditFields(operator),
		})
	}
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		conflicts := s.findBatchConflictByDB(ctx.Tx, candidates)
		for _, conflict := range conflicts {
			if conflict != "" {
				return errorsx.InvalidParam("存在冲突排班，请先处理冲突")
			}
		}
		return repositories.AgentTeamScheduleRepository.CreateBatch(ctx.Tx, schedules)
	}); err != nil {
		s.writeMu.Unlock()
		return nil, err
	}
	s.writeMu.Unlock()
	for i := range schedules {
		s.dispatchPendingConversationsIfActive(&schedules[i])
	}
	return &AgentTeamScheduleBatchGenerateResult{Created: len(schedules)}, nil
}

func (s *agentTeamScheduleService) buildScheduleModel(id, teamID int64, startAt, endAt, remark string) (*models.AgentTeamSchedule, error) {
	if teamID <= 0 {
		return nil, errorsx.InvalidParam("请选择客服组")
	}
	team := AgentTeamService.Get(teamID)
	if team == nil {
		return nil, errorsx.InvalidParam("客服组不存在")
	}
	if !slices.Contains(enums.StatusValues, team.Status) {
		return nil, errorsx.InvalidParam("客服组状态不合法")
	}
	startAtValue, err := parseRequiredDateTime(startAt, "开始时间格式错误")
	if err != nil {
		return nil, err
	}
	endAtValue, err := parseRequiredDateTime(endAt, "结束时间格式错误")
	if err != nil {
		return nil, err
	}
	if !endAtValue.After(startAtValue) {
		return nil, errorsx.InvalidParam("结束时间必须晚于开始时间")
	}
	if !sameLocalDay(startAtValue, endAtValue) {
		return nil, errorsx.InvalidParam("单条排班记录不能跨天")
	}
	if startAtValue.Before(startOfLocalDay(time.Now())) {
		return nil, errorsx.InvalidParam("不能添加或修改历史日期的排班")
	}
	overlapping := repositories.AgentTeamScheduleRepository.FindOverlappingByTeamIDsAndTimeRange(sqls.DB(), []int64{teamID}, startAtValue, endAtValue)
	for _, item := range overlapping {
		if item.ID != id {
			return nil, errorsx.InvalidParam("该客服组在所选时间段已存在排班")
		}
	}
	return &models.AgentTeamSchedule{
		TeamID:  teamID,
		StartAt: startAtValue,
		EndAt:   endAtValue,
		Remark:  strings.TrimSpace(remark),
	}, nil
}

func (s *agentTeamScheduleService) buildBatchScheduleCandidates(req request.AgentTeamScheduleBatchRequest) ([]batchScheduleCandidate, error) {
	teamIDs := uniquePositiveInt64s(req.TeamIDs)
	if len(teamIDs) == 0 {
		return nil, errorsx.InvalidParam("请选择客服组")
	}
	weekdays, err := normalizeBatchWeekdays(req.Weekdays)
	if err != nil {
		return nil, err
	}
	startDate, err := parseRequiredDate(req.StartDate, "开始日期格式错误")
	if err != nil {
		return nil, err
	}
	endDate, err := parseRequiredDate(req.EndDate, "结束日期格式错误")
	if err != nil {
		return nil, err
	}
	if endDate.Before(startDate) {
		return nil, errorsx.InvalidParam("结束日期必须晚于或等于开始日期")
	}
	if startDate.Before(startOfLocalDay(time.Now())) {
		return nil, errorsx.InvalidParam("不能添加或修改历史日期的排班")
	}
	startClock, err := parseRequiredClock(req.StartTime, "开始时间格式错误")
	if err != nil {
		return nil, err
	}
	endClock, err := parseRequiredClock(req.EndTime, "结束时间格式错误")
	if err != nil {
		return nil, err
	}
	firstStartAt := combineDateAndClock(startDate, startClock)
	firstEndAt := combineDateAndClock(startDate, endClock)
	if !firstEndAt.After(firstStartAt) {
		return nil, errorsx.InvalidParam("结束时间必须晚于开始时间")
	}

	teams := AgentTeamService.FindByIds(teamIDs)
	teamsByID := make(map[int64]models.AgentTeam, len(teams))
	for _, team := range teams {
		teamsByID[team.ID] = team
	}
	for _, teamID := range teamIDs {
		team, ok := teamsByID[teamID]
		if !ok || team.Status == enums.StatusDeleted {
			return nil, errorsx.InvalidParam("客服组不存在")
		}
		if !slices.Contains(enums.StatusValues, team.Status) {
			return nil, errorsx.InvalidParam("客服组状态不合法")
		}
	}

	weekdaySet := make(map[int]struct{}, len(weekdays))
	for _, weekday := range weekdays {
		weekdaySet[weekday] = struct{}{}
	}
	candidates := make([]batchScheduleCandidate, 0)
	remark := strings.TrimSpace(req.Remark)
	for _, teamID := range teamIDs {
		team := teamsByID[teamID]
		for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
			if _, ok := weekdaySet[weekdayForBatchRequest(date)]; !ok {
				continue
			}
			if len(candidates) >= maxAgentTeamScheduleBatchItems {
				return nil, errorsx.InvalidParam(fmt.Sprintf("单次最多生成 %d 条排班", maxAgentTeamScheduleBatchItems))
			}
			candidates = append(candidates, batchScheduleCandidate{
				TeamID:   teamID,
				TeamName: team.Name,
				Date:     date,
				StartAt:  combineDateAndClock(date, startClock),
				EndAt:    combineDateAndClock(date, endClock),
				Remark:   remark,
			})
		}
	}
	if len(candidates) == 0 {
		return nil, errorsx.InvalidParam("未生成任何排班")
	}
	return candidates, nil
}

func parseRequiredDate(value, message string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errorsx.InvalidParam(message)
	}
	ret, err := time.ParseInLocation(time.DateOnly, value, time.Local)
	if err != nil {
		return time.Time{}, errorsx.InvalidParam(message + "，请使用 yyyy-MM-dd")
	}
	return startOfLocalDay(ret), nil
}

func parseRequiredClock(value, message string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errorsx.InvalidParam(message)
	}
	layouts := []string{"15:04", "15:04:05"}
	for _, layout := range layouts {
		if ret, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return ret, nil
		}
	}
	return time.Time{}, errorsx.InvalidParam(message + "，请使用 HH:mm 或 HH:mm:ss")
}

func combineDateAndClock(date, clock time.Time) time.Time {
	year, month, day := date.In(time.Local).Date()
	hour, minute, second := clock.In(time.Local).Clock()
	return time.Date(year, month, day, hour, minute, second, 0, time.Local)
}

func buildBatchPreviewResult(candidates []batchScheduleCandidate, conflicts map[int]string) *AgentTeamScheduleBatchPreviewResult {
	items := make([]AgentTeamScheduleBatchPreviewItem, 0, len(candidates))
	hasConflict := false
	for i, candidate := range candidates {
		conflictReason := conflicts[i]
		conflict := conflictReason != ""
		if conflict {
			hasConflict = true
		}
		items = append(items, AgentTeamScheduleBatchPreviewItem{
			TeamID:         candidate.TeamID,
			TeamName:       candidate.TeamName,
			Date:           candidate.Date,
			Weekday:        weekdayForBatchRequest(candidate.Date),
			StartAt:        candidate.StartAt,
			EndAt:          candidate.EndAt,
			Remark:         candidate.Remark,
			Conflict:       conflict,
			ConflictReason: conflictReason,
		})
	}
	return &AgentTeamScheduleBatchPreviewResult{
		Total:    len(items),
		Conflict: hasConflict,
		Items:    items,
	}
}

func (s *agentTeamScheduleService) findBatchConflict(candidates []batchScheduleCandidate) map[int]string {
	return s.findBatchConflictByDB(sqls.DB(), candidates)
}

func (s *agentTeamScheduleService) findBatchConflictByDB(db *gorm.DB, candidates []batchScheduleCandidate) map[int]string {
	conflicts := make(map[int]string)
	if len(candidates) == 0 {
		return conflicts
	}
	teamIDs := make([]int64, 0, len(candidates))
	startAt := candidates[0].StartAt
	endAt := candidates[0].EndAt
	for _, candidate := range candidates {
		teamIDs = append(teamIDs, candidate.TeamID)
		if candidate.StartAt.Before(startAt) {
			startAt = candidate.StartAt
		}
		if candidate.EndAt.After(endAt) {
			endAt = candidate.EndAt
		}
	}
	existing := repositories.AgentTeamScheduleRepository.FindOverlappingByTeamIDsAndTimeRange(db, uniquePositiveInt64s(teamIDs), startAt, endAt)
	for i, candidate := range candidates {
		for _, item := range existing {
			if item.TeamID != candidate.TeamID {
				continue
			}
			if item.StartAt.Before(candidate.EndAt) && item.EndAt.After(candidate.StartAt) {
				conflicts[i] = fmt.Sprintf("该客服组在 %s 至 %s 已存在排班", item.StartAt.Format(time.DateTime), item.EndAt.Format(time.DateTime))
				break
			}
		}
	}
	return conflicts
}

func normalizeBatchWeekdays(values []int) ([]int, error) {
	seen := make(map[int]struct{}, len(values))
	ret := make([]int, 0, len(values))
	for _, value := range values {
		if value < 1 || value > 7 {
			return nil, errorsx.InvalidParam("星期必须在 1 到 7 之间")
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		ret = append(ret, value)
	}
	if len(ret) == 0 {
		return nil, errorsx.InvalidParam("请选择星期")
	}
	return ret, nil
}

func weekdayForBatchRequest(value time.Time) int {
	if value.Weekday() == time.Sunday {
		return 7
	}
	return int(value.Weekday())
}

func uniquePositiveInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	ret := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		ret = append(ret, value)
	}
	return ret
}

func parseRequiredDateTime(value, message string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errorsx.InvalidParam(message)
	}
	ret, err := parseDateTimeValue(value)
	if err != nil {
		return time.Time{}, errorsx.InvalidParam(message + "，请使用 yyyy-MM-dd HH:mm:ss 或 RFC3339")
	}
	return ret, nil
}

func parseDateTimeValue(value string) (time.Time, error) {
	layouts := []string{
		time.DateTime,
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if ret, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return ret, nil
		}
	}
	return time.Time{}, errorsx.InvalidParam("时间格式错误")
}

func startOfLocalDay(value time.Time) time.Time {
	year, month, day := value.In(time.Local).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

func sameLocalDay(a, b time.Time) bool {
	aYear, aMonth, aDay := a.In(time.Local).Date()
	bYear, bMonth, bDay := b.In(time.Local).Date()
	return aYear == bYear && aMonth == bMonth && aDay == bDay
}

func (s *agentTeamScheduleService) dispatchPendingConversationsIfActive(item *models.AgentTeamSchedule) {
	if item == nil {
		return
	}
	if item.Status != enums.StatusOk {
		return
	}
	now := time.Now()
	if item.StartAt.After(now) || !item.EndAt.After(now) {
		return
	}
	_, _ = ConversationDispatchService.DispatchPendingConversations(0)
}
