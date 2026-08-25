package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MyClaims 自定义 JWT 载荷：标准 Claims + 业务字段（user_id / username）
type MyClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT。
// 参数：secret 签名密钥；issuer 签发者；expireHours 有效期（小时）
// 返回：token 字符串、过期时间戳（秒）
func GenerateToken(secret, issuer string, userID uint, username string, expireHours int) (string, int64, error) {
	now := time.Now()
	expireAt := now.Add(time.Duration(expireHours) * time.Hour)

	claims := MyClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	return signed, expireAt.Unix(), nil
}

// ParseToken 解析并校验 JWT（验签 + 过期检查由库完成）。
// 返回错误约定：
//   - 密钥错误/格式错误 → ErrTokenInvalid
//   - 已过期           → ErrTokenExpired
var (
	ErrTokenInvalid = errors.New("token 无效")
	ErrTokenExpired = errors.New("token 已过期")
)

func ParseToken(secret, tokenStr string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {
		// 校验签名算法必须是 HS256，防止算法替换攻击
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		// jwt 库会把过期错误包装在 err 里，用 errors.Is 判断
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*MyClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}
