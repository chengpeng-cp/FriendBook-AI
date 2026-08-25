// Package utils 通用工具函数：密码哈希、JWT 等。
package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword 使用 bcrypt 对明文密码加密（cost 使用库默认值 10）。
// bcrypt 是单向哈希，同一明文每次加密结果不同，校验时用 CheckPassword。
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 校验明文密码与 bcrypt 哈希是否匹配。
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
