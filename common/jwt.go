package common

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AdminRole 管理员角色：超级管理员 / 财务 / 门店操作员
type AdminRole string

const (
	RoleSuperAdmin AdminRole = "super_admin"
	RoleFinance    AdminRole = "finance"
	RoleOperator   AdminRole = "operator"
)

// AdminClaims JWT 载荷
type AdminClaims struct {
	AdminID  uint64    `json:"admin_id"`
	Username string    `json:"username"`
	Role     AdminRole `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成管理员登录 token
func GenerateToken(adminID uint64, username string, role AdminRole, secret string, expireHours int) (string, error) {
	claims := AdminClaims{
		AdminID:  adminID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    ServiceName,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析并校验 token
func ParseToken(tokenStr, secret string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AdminClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*AdminClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// HasRole 检查是否拥有指定角色之一
func HasRole(role AdminRole, allowed ...AdminRole) bool {
	for _, r := range allowed {
		if role == r {
			return true
		}
	}
	return false
}
