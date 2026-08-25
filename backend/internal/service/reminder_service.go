package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"golang-learn/internal/model"
	"golang-learn/internal/repository"
)

// ReminderService 提醒业务逻辑：
// - 定时任务为有过生日的好友自动生成提醒（Upsert 去重）
// - 今日/近期提醒查询带 Redis 缓存（当天 23:59 失效）
type ReminderService struct {
	reminderRepo *repository.ReminderRepo
	friendRepo   *repository.FriendRepo
	rdb          *redis.Client
	logger       *zap.Logger
}

// NewReminderService 创建提醒 Service
func NewReminderService(
	reminderRepo *repository.ReminderRepo,
	friendRepo *repository.FriendRepo,
	rdb *redis.Client,
	logger *zap.Logger,
) *ReminderService {
	return &ReminderService{
		reminderRepo: reminderRepo,
		friendRepo:   friendRepo,
		rdb:          rdb,
		logger:       logger,
	}
}

// GenerateBirthdayReminders 为某用户的所有有生日的好友生成生日提醒。
// 使用 Upsert：好友生日被修改后，重跑任务会刷新提醒日期，不会重复插入。
func (s *ReminderService) GenerateBirthdayReminders(ctx context.Context, userID uint) error {
	friends, err := s.friendRepo.ListByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for i := range friends {
		f := &friends[i]
		if f.Birthday == nil {
			continue // 没填生日的好友跳过
		}
		reminder := &model.Reminder{
			FriendID:      f.ID,
			UserID:        userID,
			Type:          model.ReminderBirthday,
			Title:         fmt.Sprintf("%s 的生日", f.Name),
			Date:          *f.Birthday,
			DaysInAdvance: 3, // 默认提前 3 天开始提醒
		}
		if err := s.reminderRepo.Upsert(ctx, reminder); err != nil {
			// 单个失败只记日志，继续处理其他好友
			s.logger.Error("生成生日提醒失败",
				zap.Uint("friend_id", f.ID), zap.Error(err))
		}
	}
	return nil
}

// GenerateAllReminders 为全部用户生成提醒（定时任务调用）
func (s *ReminderService) GenerateAllReminders(ctx context.Context) error {
	userIDs, err := s.friendRepo.ListAllUserIDs(ctx)
	if err != nil {
		return err
	}

	successCount := 0
	for _, uid := range userIDs {
		if err := s.GenerateBirthdayReminders(ctx, uid); err != nil {
			s.logger.Error("为用户生成提醒失败", zap.Uint("user_id", uid), zap.Error(err))
			continue
		}
		successCount++
	}
	s.logger.Info("批量生成提醒完成",
		zap.Int("total", len(userIDs)), zap.Int("success", successCount))
	return nil
}

// GetTodayReminders 获取今日提醒（含未来 3 天内的提醒，提前提醒语义）。
// 带 Redis 缓存，缓存到当天 23:59:59 失效。
func (s *ReminderService) GetTodayReminders(ctx context.Context, userID uint) ([]model.Reminder, error) {
	cacheKey := fmt.Sprintf("reminder:today:%d", userID)

	// 1. 查缓存
	if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var reminders []model.Reminder
		if json.Unmarshal([]byte(cached), &reminders) == nil {
			return reminders, nil
		}
	}

	// 2. 回源：今天 + 未来 3 天（days_in_advance 默认 3）
	now := time.Now()
	reminders, err := s.reminderRepo.GetByDateRange(ctx, userID, now, now.AddDate(0, 0, 3))
	if err != nil {
		return nil, err
	}

	// 3. 写缓存，TTL 到当天结束
	if data, err := json.Marshal(reminders); err == nil {
		_ = s.rdb.Set(ctx, cacheKey, data, untilEndOfDay()).Err()
	}
	return reminders, nil
}

// GetUpcomingReminders 获取近期提醒（未来 7 天）
func (s *ReminderService) GetUpcomingReminders(ctx context.Context, userID uint) ([]model.Reminder, error) {
	now := time.Now()
	return s.reminderRepo.GetByDateRange(ctx, userID, now, now.AddDate(0, 0, 7))
}

// GetReminders AI 工具统一入口：type = "today" / "upcoming"（默认 today）
func (s *ReminderService) GetReminders(ctx context.Context, userID uint, typ string) ([]model.Reminder, error) {
	if typ == "upcoming" {
		return s.GetUpcomingReminders(ctx, userID)
	}
	return s.GetTodayReminders(ctx, userID)
}

// untilEndOfDay 计算距离当天 23:59:59 的时长（提醒缓存 TTL）
func untilEndOfDay() time.Duration {
	now := time.Now()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	return endOfDay.Sub(now)
}
