package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"golang-learn/internal/model"
	"golang-learn/internal/repository"
)

// ==================== 参数结构定义 ====================
// Handler 层与 AI Tools 层共用这套参数结构。

// AccountParam 平台账号入参
type AccountParam struct {
	Platform string `json:"platform"` // wechat/qq/douyin/xhs/bilibili/phone/email/custom
	Account  string `json:"account"`  // 账号
	Remark   string `json:"remark"`   // 备注
}

// CreateFriendParams 创建好友入参（HTTP 与 AI 工具共用）
type CreateFriendParams struct {
	Name         string
	Remark       string
	Gender       int8
	Birthday     *time.Time
	Phone        string
	Email        string
	Location     string
	Occupation   string
	Bio          string
	MeetScene    string
	Favorability int // 0-100，默认 50
	MBTI         string
	Accounts     []AccountParam // 多平台账号
	GroupIDs     []uint         // 分组 ID 列表
	TagIDs       []uint         // 标签 ID 列表
	UserID       uint           // 归属用户（数据隔离）
}

// UpdateFriendParams 更新好友入参（零值表示不更新该字段）
type UpdateFriendParams struct {
	Name       *string
	Remark     *string
	Phone      *string
	Email      *string
	Location   *string
	Occupation *string
	Bio        *string
	MBTI       *string
	GroupIDs   []uint
	TagIDs     []uint
}

// FriendListResult 好友列表结果（含分页信息，供 Handler / AI 工具复用）
type FriendListResult struct {
	List     []model.Friend `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// ToSummary 生成列表摘要（供 AI 工具返回，控制 token 量）
func (r *FriendListResult) ToSummary() string {
	s := fmt.Sprintf("共 %d 位好友（第 %d/%d 页）：\n", r.Total, r.Page, (r.Total+int64(r.PageSize)-1)/int64(r.PageSize))
	for i := range r.List {
		s += strconv.Itoa(i+1) + ". " + r.List[i].ToSummary() + "\n"
	}
	return s
}

// ==================== Service 定义 ====================

// FriendService 好友业务逻辑（含好感度/温度联动、CSV 导入导出）
type FriendService struct {
	friendRepo *repository.FriendRepo
	logger     *zap.Logger
}

// NewFriendService 创建好友 Service
func NewFriendService(friendRepo *repository.FriendRepo, logger *zap.Logger) *FriendService {
	return &FriendService{friendRepo: friendRepo, logger: logger}
}

// Create 创建好友：
// 1) 好感度归一化（默认 50）；2) 好感度联动计算温度；3) 事务写入好友 + 账号
func (s *FriendService) Create(ctx context.Context, p *CreateFriendParams) (*model.Friend, error) {
	if p.Name == "" {
		return nil, errors.New("好友姓名不能为空")
	}
	// 好感度默认值与越界处理
	if p.Favorability == 0 {
		p.Favorability = 50
	}
	if p.Favorability < 0 || p.Favorability > 100 {
		return nil, errors.New("好感度必须在 0-100 之间")
	}

	friend := &model.Friend{
		UserID:       p.UserID,
		Name:         p.Name,
		Remark:       p.Remark,
		Gender:       model.Gender(p.Gender),
		Birthday:     p.Birthday,
		Phone:        p.Phone,
		Email:        p.Email,
		Location:     p.Location,
		Occupation:   p.Occupation,
		Bio:          p.Bio,
		MeetScene:    p.MeetScene,
		Favorability: p.Favorability,
		Temperature:  FavorToTemperature(p.Favorability), // 联动计算温度
		MBTI:         p.MBTI,
	}

	// 组装平台账号
	accounts := make([]model.Account, 0, len(p.Accounts))
	for _, a := range p.Accounts {
		if a.Platform == "" || a.Account == "" {
			continue
		}
		accounts = append(accounts, model.Account{
			Platform: a.Platform,
			Account:  a.Account,
			Remark:   a.Remark,
		})
	}

	if err := s.friendRepo.Create(ctx, friend, accounts); err != nil {
		return nil, err
	}

	// 写入分组/标签关联
	if len(p.GroupIDs) > 0 {
		if err := s.friendRepo.UpdateGroups(ctx, p.UserID, friend.ID, p.GroupIDs); err != nil {
			s.logger.Warn("写入分组关联失败", zap.Uint("friend_id", friend.ID), zap.Error(err))
		}
	}
	if len(p.TagIDs) > 0 {
		if err := s.friendRepo.UpdateTags(ctx, p.UserID, friend.ID, p.TagIDs); err != nil {
			s.logger.Warn("写入标签关联失败", zap.Uint("friend_id", friend.ID), zap.Error(err))
		}
	}
	return friend, nil
}

// GetByID 查询好友详情（含账号/分组/标签）
func (s *FriendService) GetByID(ctx context.Context, userID, friendID uint) (*model.Friend, error) {
	return s.friendRepo.GetByID(ctx, userID, friendID)
}

// List 分页查询好友列表
func (s *FriendService) List(ctx context.Context, userID uint, q *repository.FriendQuery) (*FriendListResult, error) {
	// 分页参数兜底
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	if q.PageSize > 100 {
		q.PageSize = 100 // 防止一次拉全表
	}

	list, total, err := s.friendRepo.List(ctx, userID, q)
	if err != nil {
		return nil, err
	}
	return &FriendListResult{List: list, Total: total, Page: q.Page, PageSize: q.PageSize}, nil
}

// Search 按关键词搜索好友（关键词必填）
func (s *FriendService) Search(ctx context.Context, userID uint, keyword string, page, pageSize int) (*FriendListResult, error) {
	if keyword == "" {
		return nil, errors.New("搜索关键词不能为空")
	}
	return s.List(ctx, userID, &repository.FriendQuery{Keyword: keyword, Page: page, PageSize: pageSize})
}

// Update 更新好友基础信息（零值不更新）
func (s *FriendService) Update(ctx context.Context, userID, friendID uint, p *UpdateFriendParams) error {
	updates := map[string]interface{}{}
	if p.Name != nil && *p.Name != "" {
		updates["name"] = *p.Name
	}
	if p.Remark != nil {
		updates["remark"] = *p.Remark
	}
	if p.Phone != nil {
		updates["phone"] = *p.Phone
	}
	if p.Email != nil {
		updates["email"] = *p.Email
	}
	if p.Location != nil {
		updates["location"] = *p.Location
	}
	if p.Occupation != nil {
		updates["occupation"] = *p.Occupation
	}
	if p.Bio != nil {
		updates["bio"] = *p.Bio
	}
	if p.MBTI != nil {
		updates["mbti"] = *p.MBTI
	}

	if len(updates) > 0 {
		if err := s.friendRepo.Update(ctx, userID, friendID, updates); err != nil {
			return err
		}
	}
	if p.GroupIDs != nil {
		if err := s.friendRepo.UpdateGroups(ctx, userID, friendID, p.GroupIDs); err != nil {
			return err
		}
	}
	if p.TagIDs != nil {
		if err := s.friendRepo.UpdateTags(ctx, userID, friendID, p.TagIDs); err != nil {
			return err
		}
	}
	return nil
}

// Delete 删除好友（级联清理关联数据）
func (s *FriendService) Delete(ctx context.Context, userID, friendID uint) error {
	err := s.friendRepo.Delete(ctx, userID, friendID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("好友不存在")
	}
	return err
}

// AdjustFavorability 调整好感度并联动更新关系温度。
// 入参为绝对值（0-100），温度按 favorToTemperature 映射规则联动。
func (s *FriendService) AdjustFavorability(ctx context.Context, userID, friendID uint, favorability int) error {
	if favorability < 0 || favorability > 100 {
		return errors.New("好感度必须在 0-100 之间")
	}
	temp := FavorToTemperature(favorability)
	err := s.friendRepo.Update(ctx, userID, friendID, map[string]interface{}{
		"favorability": favorability,
		"temperature":  temp,
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("好友不存在")
	}
	return err
}

// FavorToTemperature 好感度 → 关系温度映射（导出供测试与 AI 模块复用）
//   >=90 夯 | 75-89 热 | 60-74 温 | 40-59 凉 | <40 拉
func FavorToTemperature(favor int) model.Temperature {
	switch {
	case favor >= 90:
		return model.TempHot
	case favor >= 75:
		return model.TempWarm
	case favor >= 60:
		return model.TempMild
	case favor >= 40:
		return model.TempCool
	default:
		return model.TempCold
	}
}

// RefreshTemperature 按最近互动时间刷新某用户全部好友的关系温度。
// 规则（与好感度映射相互独立，纯按互动间隔）：
//   <=7天 夯 | <=30天 热 | <=90天 温 | <=180天 凉 | >180天 拉
// 从未互动的好友跳过不处理。
func (s *FriendService) RefreshTemperature(ctx context.Context, userID uint) error {
	friends, err := s.friendRepo.ListByUserID(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	for i := range friends {
		f := &friends[i]
		if f.LastInteract == nil {
			continue // 从未互动，跳过
		}
		days := int(now.Sub(*f.LastInteract).Hours() / 24)

		var temp model.Temperature
		switch {
		case days <= 7:
			temp = model.TempHot
		case days <= 30:
			temp = model.TempWarm
		case days <= 90:
			temp = model.TempMild
		case days <= 180:
			temp = model.TempCool
		default:
			temp = model.TempCold
		}

		// 温度没变化就不写库，减少无效 UPDATE
		if temp == f.Temperature {
			continue
		}
		if err := s.friendRepo.Update(ctx, userID, f.ID, map[string]interface{}{
			"temperature": temp,
		}); err != nil {
			// 单个失败只记日志，不中断整批
			s.logger.Error("刷新温度失败", zap.Uint("friend_id", f.ID), zap.Error(err))
		}
	}
	return nil
}

// RefreshAllUsersTemperature 刷新所有用户的关系温度（定时任务调用）
func (s *FriendService) RefreshAllUsersTemperature(ctx context.Context) error {
	userIDs, err := s.friendRepo.ListAllUserIDs(ctx)
	if err != nil {
		return err
	}
	for _, uid := range userIDs {
		if err := s.RefreshTemperature(ctx, uid); err != nil {
			s.logger.Error("刷新用户温度失败", zap.Uint("user_id", uid), zap.Error(err))
			continue
		}
	}
	s.logger.Info("全量温度刷新完成", zap.Int("users", len(userIDs)))
	return nil
}

// ListColdFriends 查询"关系变凉"好友（凉/拉档位），按最久未互动排序
func (s *FriendService) ListColdFriends(ctx context.Context, userID uint, page, pageSize int) ([]model.Friend, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return s.friendRepo.ListColdFriends(ctx, userID, page, pageSize)
}

// ListAccounts 查询好友的平台账号列表
func (s *FriendService) ListAccounts(ctx context.Context, userID, friendID uint) ([]model.Account, error) {
	accounts, err := s.friendRepo.ListAccounts(ctx, userID, friendID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("好友不存在")
	}
	return accounts, err
}

// AddAccount 给好友添加平台账号
func (s *FriendService) AddAccount(ctx context.Context, userID, friendID uint, platform, account, remark string) error {
	if platform == "" || account == "" {
		return errors.New("平台和账号不能为空")
	}
	err := s.friendRepo.AddAccount(ctx, userID, &model.Account{
		FriendID: friendID,
		Platform: platform,
		Account:  account,
		Remark:   remark,
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("好友不存在")
	}
	return err
}

// ==================== CSV 导入导出 ====================

// ExportFriendsCSV 导出当前用户所有好友为 CSV。
// 细节：文件头写 UTF-8 BOM（0xEF 0xBB 0xBF），否则 Excel 打开中文乱码。
func (s *FriendService) ExportFriendsCSV(ctx context.Context, userID uint) ([]byte, error) {
	friends, err := s.friendRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	buf.Write([]byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM
	w := csv.NewWriter(buf)

	// 表头
	if err := w.Write([]string{
		"ID", "姓名", "备注", "性别", "生日", "手机号",
		"所在城市", "职业", "好感度", "关系温度", "MBTI",
	}); err != nil {
		return nil, err
	}

	// 数据行
	for i := range friends {
		f := &friends[i]
		row := []string{
			strconv.FormatUint(uint64(f.ID), 10),
			f.Name,
			f.Remark,
			f.Gender.String(),
			"",
			f.Phone,
			f.Location,
			f.Occupation,
			strconv.Itoa(f.Favorability),
			f.Temperature.String(),
			f.MBTI,
		}
		if f.Birthday != nil {
			row[4] = f.Birthday.Format("2006-01-02")
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	// 必须 Flush，否则缓冲区最后几行会丢
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ImportResult CSV 导入结果
type ImportResult struct {
	SuccessCount int      `json:"success_count"` // 成功导入条数
	FailedRows   []string `json:"failed_rows"`   // 失败行描述（行号 + 原因）
}

// ImportFriendsCSV 从 CSV 字节流批量导入好友。
// CSV 格式与 ExportFriendsCSV 导出格式一致（第一行为表头，自动跳过）。
func (s *FriendService) ImportFriendsCSV(ctx context.Context, userID uint, data []byte) (*ImportResult, error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1 // 允许字段数不一致（Excel 编辑后常多列）

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("解析 CSV 失败: %w", err)
	}
	if len(records) < 2 {
		return nil, errors.New("CSV 文件无数据行")
	}

	result := &ImportResult{FailedRows: []string{}}
	// 遍历数据行（跳过表头）
	for i, record := range records[1:] {
		rowNo := i + 2 // 行号从 2 开始（1 是表头）
		if len(record) < 2 || record[1] == "" {
			result.FailedRows = append(result.FailedRows, fmt.Sprintf("第 %d 行：姓名为空", rowNo))
			continue
		}

		p := &CreateFriendParams{
			Name:   record[1],
			Remark: cell(record, 2),
			UserID: userID,
		}
		// 性别：中文文案转枚举
		switch cell(record, 3) {
		case "男":
			p.Gender = 1
		case "女":
			p.Gender = 2
		default:
			p.Gender = 0
		}
		// 生日
		if b := cell(record, 4); b != "" {
			if t, err := time.Parse("2006-01-02", b); err == nil {
				p.Birthday = &t
			}
		}
		p.Phone = cell(record, 5)
		p.Location = cell(record, 6)
		p.Occupation = cell(record, 7)
		// 好感度
		if v, err := strconv.Atoi(cell(record, 8)); err == nil && v >= 0 && v <= 100 {
			p.Favorability = v
		}
		p.MBTI = cell(record, 10)

		if _, err := s.Create(ctx, p); err != nil {
			result.FailedRows = append(result.FailedRows,
				fmt.Sprintf("第 %d 行（%s）：%v", rowNo, p.Name, err))
			continue
		}
		result.SuccessCount++
	}
	return result, nil
}

// cell 安全取 CSV 某列（越界返回空串）
func cell(record []string, idx int) string {
	if idx < len(record) {
		return record[idx]
	}
	return ""
}

// FriendJSON 把好友序列化为 JSON 字符串（MCP/AI 工具返回用）
func FriendJSON(f *model.Friend) string {
	data, _ := json.Marshal(f)
	return string(data)
}
