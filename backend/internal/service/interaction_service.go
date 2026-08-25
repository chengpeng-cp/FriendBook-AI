package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"golang-learn/internal/model"
	"golang-learn/internal/repository"
)

// InteractionParams 创建/更新互动记录入参
type InteractionParams struct {
	FriendID    int       // 好友 ID（必填）
	Type        int       // 互动类型 1-6（0 默认为聊天）
	Content     string    // 内容摘要
	Location    string    // 地点
	Mood        string    // 心情
	PhotoURLs   []string  // 照片 URL 列表
	IsMemory    bool      // 是否标记为记忆卡片
	MemoryTitle string    // 记忆卡片标题
	InteractAt  time.Time // 互动时间（零值取当前时间）
}

// InteractionService 互动记录业务逻辑
type InteractionService struct {
	interactionRepo *repository.InteractionRepo
	friendRepo      *repository.FriendRepo
	logger          *zap.Logger
}

// NewInteractionService 创建互动记录 Service
func NewInteractionService(
	interactionRepo *repository.InteractionRepo,
	friendRepo *repository.FriendRepo,
	logger *zap.Logger,
) *InteractionService {
	return &InteractionService{
		interactionRepo: interactionRepo,
		friendRepo:      friendRepo,
		logger:          logger,
	}
}

// Create 记录一次互动，并联动更新好友的最近互动时间与关系温度。
// 联动规则（业务核心）：
//  1. 好友 last_interact 更新为本次互动时间
//  2. 按互动间隔重算关系温度（刚互动完温度会回升）
func (s *InteractionService) Create(ctx context.Context, userID uint, p *InteractionParams) (*model.Interaction, error) {
	if p.FriendID <= 0 {
		return nil, errors.New("好友 ID 不能为空")
	}

	// 校验好友存在且属于当前用户
	friend, err := s.friendRepo.GetByID(ctx, userID, uint(p.FriendID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("好友不存在")
		}
		return nil, err
	}

	// 互动时间零值兜底为当前时间
	if p.InteractAt.IsZero() {
		p.InteractAt = time.Now()
	}
	// 互动类型兜底为"聊天"
	if p.Type <= 0 {
		p.Type = int(model.InteractChat)
	}

	interaction := &model.Interaction{
		FriendID:    friend.ID,
		UserID:      userID,
		Type:        model.InteractionType(p.Type),
		Content:     p.Content,
		Location:    p.Location,
		Mood:        p.Mood,
		IsMemory:    p.IsMemory,
		MemoryTitle: p.MemoryTitle,
		InteractAt:  p.InteractAt,
	}
	// 照片列表序列化为 JSON 字符串存储。
	// 注意：MySQL JSON 列不接受空字符串（Error 3140: The document is empty），
	// 空列表必须显式写 "[]"，不能留 Go 零值 ""。
	if len(p.PhotoURLs) > 0 {
		interaction.PhotoURLs = mustJSON(p.PhotoURLs)
	} else {
		interaction.PhotoURLs = "[]"
	}

	if err := s.interactionRepo.Create(ctx, interaction); err != nil {
		return nil, err
	}

	// ======== 联动更新好友 ========
	// 只有更新的互动时间比原 last_interact 更新，才触发联动（补录历史互动不应刷新温度）
	shouldRefresh := friend.LastInteract == nil || p.InteractAt.After(*friend.LastInteract)
	if shouldRefresh {
		updates := map[string]interface{}{
			"last_interact": p.InteractAt,
			"temperature":   InteractDaysToTemperature(p.InteractAt),
		}
		if err := s.friendRepo.Update(ctx, userID, friend.ID, updates); err != nil {
			// 联动失败不影响互动记录本身，只记日志
			s.logger.Warn("联动更新好友互动信息失败",
				zap.Uint("friend_id", friend.ID), zap.Error(err))
		}
	}
	return interaction, nil
}

// InteractDaysToTemperature 按互动时间距今的天数推断温度（刚互动=夯）
func InteractDaysToTemperature(interactAt time.Time) model.Temperature {
	days := int(time.Since(interactAt).Hours() / 24)
	switch {
	case days <= 7:
		return model.TempHot
	case days <= 30:
		return model.TempWarm
	case days <= 90:
		return model.TempMild
	case days <= 180:
		return model.TempCool
	default:
		return model.TempCold
	}
}

// Update 更新互动记录
func (s *InteractionService) Update(ctx context.Context, userID, interactionID uint, p *InteractionParams) error {
	updates := map[string]interface{}{}
	if p.Type > 0 {
		updates["type"] = p.Type
	}
	if p.Content != "" {
		updates["content"] = p.Content
	}
	if p.Location != "" {
		updates["location"] = p.Location
	}
	if p.Mood != "" {
		updates["mood"] = p.Mood
	}
	if !p.InteractAt.IsZero() {
		updates["interact_at"] = p.InteractAt
	}
	if len(updates) == 0 {
		return nil
	}
	err := s.interactionRepo.Update(ctx, userID, interactionID, updates)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("互动记录不存在")
	}
	return err
}

// Delete 删除互动记录
func (s *InteractionService) Delete(ctx context.Context, userID, interactionID uint) error {
	err := s.interactionRepo.Delete(ctx, userID, interactionID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("互动记录不存在")
	}
	return err
}

// List 分页查询互动记录（friendID=0 查全部）
func (s *InteractionService) List(ctx context.Context, userID, friendID uint, page, pageSize int) ([]model.Interaction, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return s.interactionRepo.List(ctx, userID, friendID, page, pageSize)
}

// ListByFriendID 查询好友全部互动（AI 生成破冰话术引用共同记忆用）
func (s *InteractionService) ListByFriendID(ctx context.Context, userID, friendID uint) ([]model.Interaction, error) {
	return s.interactionRepo.ListByFriendID(ctx, userID, friendID)
}

// mustJSON 序列化失败返回空数组字符串（照片列表为辅助信息，不因序列化失败中断主流程）
func mustJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(data)
}
