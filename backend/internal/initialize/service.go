package initialize

import (
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"golang-learn/internal/config"
	"golang-learn/internal/repository"
	"golang-learn/internal/service"
)

// Container 依赖注入容器：
// 统一持有 Repository 与 Service 实例，Router / AI Tools / MCP / 定时任务
// 都从这里取依赖，避免到处手写构造链条。
type Container struct {
	// Repository 层
	UserRepo        *repository.UserRepo
	FriendRepo      *repository.FriendRepo
	GroupRepo       *repository.GroupRepo
	TagRepo         *repository.TagRepo
	InteractionRepo *repository.InteractionRepo
	ReminderRepo    *repository.ReminderRepo
	ChatRepo        *repository.ChatRepo

	// Service 层
	UserService        *service.UserService
	FriendService      *service.FriendService
	GroupService       *service.GroupService
	TagService         *service.TagService
	InteractionService *service.InteractionService
	StatsService       *service.StatsService
	ReminderService    *service.ReminderService
}

// InitContainer 组装依赖：Repository（依赖 DB）→ Service（依赖 Repository）。
// 构造顺序即依赖顺序，一目了然。
func InitContainer(
	db *gorm.DB,
	rdb *redis.Client,
	cfg *config.Config,
	logger *zap.Logger,
) *Container {
	// ===== Repository =====
	userRepo := repository.NewUserRepo(db)
	friendRepo := repository.NewFriendRepo(db)
	groupRepo := repository.NewGroupRepo(db)
	tagRepo := repository.NewTagRepo(db)
	interactionRepo := repository.NewInteractionRepo(db)
	reminderRepo := repository.NewReminderRepo(db)
	chatRepo := repository.NewChatRepo(db)

	// ===== Service =====
	userService := service.NewUserService(userRepo, &cfg.JWT)
	friendService := service.NewFriendService(friendRepo, logger)
	groupService := service.NewGroupService(groupRepo)
	tagService := service.NewTagService(tagRepo)
	interactionService := service.NewInteractionService(interactionRepo, friendRepo, logger)
	statsService := service.NewStatsService(friendRepo, interactionRepo, rdb)
	reminderService := service.NewReminderService(reminderRepo, friendRepo, rdb, logger)

	return &Container{
		UserRepo:        userRepo,
		FriendRepo:      friendRepo,
		GroupRepo:       groupRepo,
		TagRepo:         tagRepo,
		InteractionRepo: interactionRepo,
		ReminderRepo:    reminderRepo,
		ChatRepo:        chatRepo,

		UserService:        userService,
		FriendService:      friendService,
		GroupService:       groupService,
		TagService:         tagService,
		InteractionService: interactionService,
		StatsService:       statsService,
		ReminderService:    reminderService,
	}
}
