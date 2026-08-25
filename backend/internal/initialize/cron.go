package initialize

import (
	"context"
	"time"

	"go.uber.org/zap"

	"golang-learn/internal/repository"
	"golang-learn/internal/service"
)

// StartCronJobs 启动后台定时任务（非阻塞，内部起 goroutine）：
//   每天凌晨 1 点执行：
//     1. RefreshAllUsersTemperature —— 按最近互动时间刷新全体好友关系温度
//     2. GenerateAllReminders       —— 为全体用户生成生日提醒
//
// 实现说明：教学项目用 time.Ticker 实现（无额外依赖）；
// 生产项目建议使用 robfig/cron 或分布式任务调度。
func StartCronJobs(
	userRepo *repository.UserRepo,
	friendSvc *service.FriendService,
	reminderSvc *service.ReminderService,
	logger *zap.Logger,
) {
	go runCronJobs(userRepo, friendSvc, reminderSvc, logger)
	logger.Info("定时任务已启动：每天凌晨 1 点刷新关系温度 + 生成生日提醒")
}

// runCronJobs 定时任务主循环
func runCronJobs(
	userRepo *repository.UserRepo,
	friendSvc *service.FriendService,
	reminderSvc *service.ReminderService,
	logger *zap.Logger,
) {
	// 首次等待到下一个凌晨 1 点，之后每 24 小时执行一次
	timer := time.NewTimer(nextOneAM())
	defer timer.Stop()

	for range timer.C {
		timer.Reset(24 * time.Hour)

		// panic 兜底：单个任务崩溃不能拖挂整个进程
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("定时任务 panic", zap.Any("error", r))
				}
			}()

			ctx := context.Background()

			// 任务 1：刷新全体用户好友的关系温度
			if err := friendSvc.RefreshAllUsersTemperature(ctx); err != nil {
				logger.Error("定时任务：刷新关系温度失败", zap.Error(err))
			} else {
				logger.Info("定时任务：关系温度刷新完成")
			}

			// 任务 2：为全体用户生成生日提醒
			if err := reminderSvc.GenerateAllReminders(ctx); err != nil {
				logger.Error("定时任务：生成提醒失败", zap.Error(err))
			} else {
				logger.Info("定时任务：提醒生成完成")
			}
		}()
	}
}

// nextOneAM 计算距离下一个凌晨 1 点的时长
func nextOneAM() time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), 1, 0, 0, 0, now.Location())
	// 今天 1 点已过，则定到明天
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next.Sub(now)
}
