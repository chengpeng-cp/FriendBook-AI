package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"golang-learn/internal/model"
)

// FriendQuery 好友列表查询条件（分页 + 多条件筛选 + 排序）
type FriendQuery struct {
	Keyword     string // 关键词（姓名/备注/手机号模糊匹配）
	GroupID     uint   // 分组筛选
	TagID       uint   // 标签筛选
	Temperature int    // 关系温度筛选（1-5，0 表示不筛选）
	MinFavor    int    // 最低好感度（0 表示不筛选）
	MaxFavor    int    // 最高好感度（0 表示不筛选）
	Sort        string // 排序字段：favorability_desc / favorability_asc / last_interact_asc ...
	Page        int    // 页码（从 1 开始）
	PageSize    int    // 每页条数
}

// FriendRepo 好友数据访问对象
type FriendRepo struct {
	db *gorm.DB
}

// NewFriendRepo 创建好友 Repository
func NewFriendRepo(db *gorm.DB) *FriendRepo {
	return &FriendRepo{db: db}
}

// DB 暴露底层 *gorm.DB（统计服务需要做聚合查询，不复用 CRUD 方法）
func (r *FriendRepo) DB() *gorm.DB {
	return r.db
}

// Create 创建好友（含关联的平台账号，使用事务保证一致性）
func (r *FriendRepo) Create(ctx context.Context, friend *model.Friend, accounts []model.Account) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(friend).Error; err != nil {
			return err
		}
		// 回填账号的好友 ID 再批量创建
		for i := range accounts {
			accounts[i].FriendID = friend.ID
		}
		if len(accounts) > 0 {
			if err := tx.Create(&accounts).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetByID 查询好友详情（必须带 user_id 条件，保证数据隔离）
// Preload 预加载关联的账号/分组/标签，避免 N+1 查询
func (r *FriendRepo) GetByID(ctx context.Context, userID, friendID uint) (*model.Friend, error) {
	var friend model.Friend
	err := r.db.WithContext(ctx).
		Preload("Accounts").
		Preload("Groups").
		Preload("Tags").
		Where("id = ? AND user_id = ?", friendID, userID).
		First(&friend).Error
	if err != nil {
		return nil, err
	}
	return &friend, nil
}

// Update 按字段更新好友（必须带 user_id 条件）
func (r *FriendRepo) Update(ctx context.Context, userID, friendID uint, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.Friend{}).
		Where("id = ? AND user_id = ?", friendID, userID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除好友（级联删除其账号/分组/标签关联/互动记录/提醒，使用事务）
func (r *FriendRepo) Delete(ctx context.Context, userID, friendID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先校验好友存在且属于当前用户
		var count int64
		if err := tx.Model(&model.Friend{}).
			Where("id = ? AND user_id = ?", friendID, userID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		// 删除好友本体
		if err := tx.Where("id = ? AND user_id = ?", friendID, userID).
			Delete(&model.Friend{}).Error; err != nil {
			return err
		}
		// 级联清理关联数据
		if err := tx.Where("friend_id = ?", friendID).Delete(&model.Account{}).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM friend_groups WHERE friend_id = ?", friendID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM friend_tags WHERE friend_id = ?", friendID).Error; err != nil {
			return err
		}
		if err := tx.Where("friend_id = ?", friendID).Delete(&model.Interaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("friend_id = ?", friendID).Delete(&model.Reminder{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// List 分页查询好友列表（支持关键词/分组/标签/温度/好感度筛选）
func (r *FriendRepo) List(ctx context.Context, userID uint, q *FriendQuery) ([]model.Friend, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Friend{}).Where("user_id = ?", userID)

	// 关键词模糊搜索：姓名 / 备注 / 手机号
	if q.Keyword != "" {
		kw := "%" + q.Keyword + "%"
		query = query.Where("name LIKE ? OR remark LIKE ? OR phone LIKE ?", kw, kw, kw)
	}
	// 好感度区间
	if q.MinFavor > 0 {
		query = query.Where("favorability >= ?", q.MinFavor)
	}
	if q.MaxFavor > 0 {
		query = query.Where("favorability <= ?", q.MaxFavor)
	}
	// 温度筛选
	if q.Temperature > 0 {
		query = query.Where("temperature = ?", q.Temperature)
	}
	// 分组筛选：通过中间表反查
	if q.GroupID > 0 {
		query = query.Where(
			"id IN (SELECT friend_id FROM friend_groups WHERE group_id = ?)", q.GroupID)
	}
	// 标签筛选：通过中间表反查
	if q.TagID > 0 {
		query = query.Where(
			"id IN (SELECT friend_id FROM friend_tags WHERE tag_id = ?)", q.TagID)
	}

	// 先查总数（在排序/分页前）
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序（白名单方式，防止 SQL 注入）
	switch q.Sort {
	case "favorability_desc":
		query = query.Order("favorability DESC")
	case "favorability_asc":
		query = query.Order("favorability ASC")
	case "last_interact_asc":
		query = query.Order("last_interact ASC")
	case "last_interact_desc":
		query = query.Order("last_interact DESC")
	default:
		query = query.Order("id DESC")
	}

	// 分页
	var friends []model.Friend
	err := query.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Find(&friends).Error
	if err != nil {
		return nil, 0, err
	}
	return friends, total, nil
}

// ListByUserID 查询用户所有好友（不分页，导出 CSV / 刷新温度 / 生成提醒用）
func (r *FriendRepo) ListByUserID(ctx context.Context, userID uint) ([]model.Friend, error) {
	var friends []model.Friend
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&friends).Error
	return friends, err
}

// ListAllUserIDs 返回有好友数据的所有用户 ID（定时任务用）
func (r *FriendRepo) ListAllUserIDs(ctx context.Context) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&model.Friend{}).Distinct().Pluck("user_id", &ids).Error
	return ids, err
}

// ListColdFriends 分页查询"关系变凉"（凉/拉档位）的好友，按最久未互动排序
func (r *FriendRepo) ListColdFriends(ctx context.Context, userID uint, page, pageSize int) ([]model.Friend, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Friend{}).
		Where("user_id = ? AND temperature IN ?", userID,
			[]model.Temperature{model.TempCool, model.TempCold})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var friends []model.Friend
	err := query.Order("last_interact ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&friends).Error
	if err != nil {
		return nil, 0, err
	}
	return friends, total, nil
}

// ErrFriendNotFound 好友不存在（含不属于当前用户两种情况）
var ErrFriendNotFound = errors.New("好友不存在")

// ListAccounts 查询好友的平台账号列表（必须校验好友属于当前用户）
func (r *FriendRepo) ListAccounts(ctx context.Context, userID, friendID uint) ([]model.Account, error) {
	// 先校验归属，防止越权查询他人好友的账号
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Friend{}).
		Where("id = ? AND user_id = ?", friendID, userID).
		Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var accounts []model.Account
	err := r.db.WithContext(ctx).
		Where("friend_id = ?", friendID).
		Order("id ASC").
		Find(&accounts).Error
	return accounts, err
}

// AddAccount 给好友添加一个平台账号
func (r *FriendRepo) AddAccount(ctx context.Context, userID uint, account *model.Account) error {
	// 校验好友归属
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Friend{}).
		Where("id = ? AND user_id = ?", account.FriendID, userID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.db.WithContext(ctx).Create(account).Error
}

// DeleteAccount 删除好友的一个平台账号（通过子查询校验归属）
func (r *FriendRepo) DeleteAccount(ctx context.Context, userID, accountID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND friend_id IN (SELECT id FROM friends WHERE user_id = ?)",
			accountID, userID).
		Delete(&model.Account{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateGroups 更新好友的分组关联（先删后建，简单可靠）
func (r *FriendRepo) UpdateGroups(ctx context.Context, userID, friendID uint, groupIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var friend model.Friend
		if err := tx.Where("id = ? AND user_id = ?", friendID, userID).First(&friend).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM friend_groups WHERE friend_id = ?", friendID).Error; err != nil {
			return err
		}
		if len(groupIDs) > 0 {
			var groups []model.Group
			if err := tx.Where("user_id = ? AND id IN ?", userID, groupIDs).Find(&groups).Error; err != nil {
				return err
			}
			if err := tx.Model(&friend).Association("Groups").Replace(groups); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateTags 更新好友的标签关联（先删后建）
func (r *FriendRepo) UpdateTags(ctx context.Context, userID, friendID uint, tagIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var friend model.Friend
		if err := tx.Where("id = ? AND user_id = ?", friendID, userID).First(&friend).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM friend_tags WHERE friend_id = ?", friendID).Error; err != nil {
			return err
		}
		if len(tagIDs) > 0 {
			var tags []model.Tag
			if err := tx.Where("user_id = ? AND id IN ?", userID, tagIDs).Find(&tags).Error; err != nil {
				return err
			}
			if err := tx.Model(&friend).Association("Tags").Replace(tags); err != nil {
				return err
			}
		}
		return nil
	})
}
