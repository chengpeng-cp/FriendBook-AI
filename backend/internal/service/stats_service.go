package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"golang-learn/internal/model"
	"golang-learn/internal/repository"
)

// ==================== 统计数据结构 ====================

// StatsOverview 统计概览（数据看板 + AI 社交分析共用）
type StatsOverview struct {
	Total          int64                       `json:"total"`           // 好友总数
	NewThisMonth   int64                       `json:"new_this_month"`  // 本月新增
	TodayReminders int64                       `json:"today_reminders"` // 今日提醒数（本月过生日好友数）
	ColdCount      int64                       `json:"cold_count"`      // 关系变凉数（凉+拉）
	AvgFavor       float64                     `json:"avg_favor"`       // 平均好感度
	TempDist       map[model.Temperature]int64 `json:"temp_dist"`       // 温度分布
	PlatformDist   map[string]int64            `json:"platform_dist"`   // 平台分布
	InteractTrend  []DailyCount                `json:"interact_trend"`  // 近 30 天互动趋势
}

// DailyCount 单日互动量
type DailyCount struct {
	Date  string `json:"date"` // yyyy-mm-dd
	Count int64  `json:"count"`
}

// PlatformDistItem 平台分布项（HTTP 接口返回用）
type PlatformDistItem struct {
	Platform string `json:"platform"`
	Count    int64  `json:"count"`
}

// TemperatureDist 温度分布（HTTP 接口返回用，字段展开便于前端画图）
type TemperatureDist struct {
	Hot  int64 `json:"hot"`  // 夯
	Warm int64 `json:"warm"` // 热
	Mild int64 `json:"mild"` // 温
	Cool int64 `json:"cool"` // 凉
	Cold int64 `json:"cold"` // 拉
}

// TrendPoint 互动趋势点（HTTP 接口返回用）
type TrendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// ==================== Service ====================

// StatsService 统计业务逻辑。
// 缓存策略：Cache-Aside（读时回源），所有统计结果缓存 5 分钟，
// 统计数据允许短暂不一致，可显著降低数据库聚合查询压力。
type StatsService struct {
	friendRepo      *repository.FriendRepo
	interactionRepo *repository.InteractionRepo
	rdb             *redis.Client
}

// NewStatsService 创建统计 Service
func NewStatsService(
	friendRepo *repository.FriendRepo,
	interactionRepo *repository.InteractionRepo,
	rdb *redis.Client,
) *StatsService {
	return &StatsService{
		friendRepo:      friendRepo,
		interactionRepo: interactionRepo,
		rdb:             rdb,
	}
}

// statsCacheTTL 统计缓存有效期
const statsCacheTTL = 5 * time.Minute

// GetOverview 统计概览（含 AI 分析需要的全量维度），带 Redis 缓存
func (s *StatsService) GetOverview(ctx context.Context, userID uint) (*StatsOverview, error) {
	cacheKey := fmt.Sprintf("stats:overview:%d", userID)

	// 1. 先查缓存
	if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var data StatsOverview
		if json.Unmarshal([]byte(cached), &data) == nil {
			return &data, nil
		}
	}

	// 2. 回源查询
	db := s.friendRepo.DB().WithContext(ctx)
	data := &StatsOverview{
		TempDist:      map[model.Temperature]int64{},
		PlatformDist:  map[string]int64{},
		InteractTrend: []DailyCount{},
	}

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 好友总数
	if err := db.Model(&model.Friend{}).
		Where("user_id = ?", userID).Count(&data.Total).Error; err != nil {
		return nil, err
	}
	// 本月新增
	if err := db.Model(&model.Friend{}).
		Where("user_id = ? AND created_at >= ?", userID, monthStart).
		Count(&data.NewThisMonth).Error; err != nil {
		return nil, err
	}
	// 今日提醒（简化口径：本月过生日的好友数）
	if err := db.Model(&model.Friend{}).
		Where("user_id = ? AND MONTH(birthday) = ?", userID, int(now.Month())).
		Count(&data.TodayReminders).Error; err != nil {
		return nil, err
	}
	// 关系变凉（凉+拉）
	if err := db.Model(&model.Friend{}).
		Where("user_id = ? AND temperature IN ?", userID,
			[]model.Temperature{model.TempCool, model.TempCold}).
		Count(&data.ColdCount).Error; err != nil {
		return nil, err
	}
	// 平均好感度（COALESCE 保证空表返回 0）
	if err := db.Model(&model.Friend{}).
		Where("user_id = ?", userID).
		Select("COALESCE(AVG(favorability), 0)").
		Scan(&data.AvgFavor).Error; err != nil {
		return nil, err
	}
	// 温度分布
	var tempRows []struct {
		Temperature model.Temperature
		Count       int64
	}
	if err := db.Model(&model.Friend{}).
		Select("temperature, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("temperature").
		Scan(&tempRows).Error; err != nil {
		return nil, err
	}
	for _, row := range tempRows {
		data.TempDist[row.Temperature] = row.Count
	}
	// 平台分布
	var platformRows []PlatformDistItem
	if err := s.friendRepo.DB().WithContext(ctx).
		Table("accounts").
		Select("accounts.platform as platform, COUNT(*) as count").
		Joins("JOIN friends ON friends.id = accounts.friend_id").
		Where("friends.user_id = ?", userID).
		Group("accounts.platform").
		Order("count DESC").
		Scan(&platformRows).Error; err != nil {
		return nil, err
	}
	for _, row := range platformRows {
		data.PlatformDist[row.Platform] = row.Count
	}
	// 近 30 天互动趋势
	since := now.AddDate(0, 0, -30)
	var trendRows []TrendPoint
	if err := s.interactionRepo.DB().WithContext(ctx).
		Model(&model.Interaction{}).
		Select("DATE_FORMAT(interact_at, '%Y-%m-%d') as date, COUNT(*) as count").
		Where("user_id = ? AND interact_at >= ?", userID, since).
		Group("date").
		Order("date ASC").
		Scan(&trendRows).Error; err != nil {
		return nil, err
	}
	for _, row := range trendRows {
		data.InteractTrend = append(data.InteractTrend, DailyCount{Date: row.Date, Count: row.Count})
	}

	// 3. 写缓存（失败不影响主流程）
	if encoded, err := json.Marshal(data); err == nil {
		_ = s.rdb.Set(ctx, cacheKey, encoded, statsCacheTTL).Err()
	}
	return data, nil
}

// PlatformDistribution 平台分布（HTTP 专用，返回数组便于前端渲染）
func (s *StatsService) PlatformDistribution(ctx context.Context, userID uint) ([]PlatformDistItem, error) {
	cacheKey := fmt.Sprintf("stats:platform:%d", userID)
	if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var result []PlatformDistItem
		if json.Unmarshal([]byte(cached), &result) == nil {
			return result, nil
		}
	}

	var result []PlatformDistItem
	err := s.friendRepo.DB().WithContext(ctx).
		Table("accounts").
		Select("accounts.platform as platform, COUNT(*) as count").
		Joins("JOIN friends ON friends.id = accounts.friend_id").
		Where("friends.user_id = ?", userID).
		Group("accounts.platform").
		Order("count DESC").
		Scan(&result).Error
	if err != nil {
		return nil, err
	}

	if encoded, err := json.Marshal(result); err == nil {
		_ = s.rdb.Set(ctx, cacheKey, encoded, statsCacheTTL).Err()
	}
	return result, nil
}

// TemperatureDistribution 温度分布（HTTP 专用）
func (s *StatsService) TemperatureDistribution(ctx context.Context, userID uint) (*TemperatureDist, error) {
	cacheKey := fmt.Sprintf("stats:temp:%d", userID)
	if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var result TemperatureDist
		if json.Unmarshal([]byte(cached), &result) == nil {
			return &result, nil
		}
	}

	var rows []struct {
		Temperature model.Temperature
		Count       int64
	}
	err := s.friendRepo.DB().WithContext(ctx).
		Model(&model.Friend{}).
		Select("temperature, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("temperature").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	// 按温度档位映射到具名字段
	result := &TemperatureDist{}
	for _, row := range rows {
		switch row.Temperature {
		case model.TempHot:
			result.Hot = row.Count
		case model.TempWarm:
			result.Warm = row.Count
		case model.TempMild:
			result.Mild = row.Count
		case model.TempCool:
			result.Cool = row.Count
		case model.TempCold:
			result.Cold = row.Count
		}
	}

	if encoded, err := json.Marshal(result); err == nil {
		_ = s.rdb.Set(ctx, cacheKey, encoded, statsCacheTTL).Err()
	}
	return result, nil
}

// InteractionTrend 近 30 天互动趋势（HTTP 专用）
func (s *StatsService) InteractionTrend(ctx context.Context, userID uint) ([]TrendPoint, error) {
	cacheKey := fmt.Sprintf("stats:trend:%d", userID)
	if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var result []TrendPoint
		if json.Unmarshal([]byte(cached), &result) == nil {
			return result, nil
		}
	}

	since := time.Now().AddDate(0, 0, -30)
	var result []TrendPoint
	err := s.interactionRepo.DB().WithContext(ctx).
		Model(&model.Interaction{}).
		Select("DATE_FORMAT(interact_at, '%Y-%m-%d') as date, COUNT(*) as count").
		Where("user_id = ? AND interact_at >= ?", userID, since).
		Group("date").
		Order("date ASC").
		Scan(&result).Error
	if err != nil {
		return nil, err
	}

	if encoded, err := json.Marshal(result); err == nil {
		_ = s.rdb.Set(ctx, cacheKey, encoded, statsCacheTTL).Err()
	}
	return result, nil
}

// ToSummary 统计概览转自然语言摘要（AI 工具 get_friend_stats 返回用）
func (o *StatsOverview) ToSummary() string {
	return fmt.Sprintf(
		"社交圈统计快照：\n- 好友总数：%d\n- 本月新增：%d\n- 关系变凉（凉+拉档位）：%d\n- 平均好感度：%.1f\n- 温度分布：夯%d 热%d 温%d 凉%d 拉%d",
		o.Total, o.NewThisMonth, o.ColdCount, o.AvgFavor,
		o.TempDist[model.TempHot], o.TempDist[model.TempWarm],
		o.TempDist[model.TempMild], o.TempDist[model.TempCool],
		o.TempDist[model.TempCold],
	)
}
