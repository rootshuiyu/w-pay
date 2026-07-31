package service

import (
	"time"

	"wpay/common"
	"wpay/config"
	"wpay/dao"
	"wpay/model"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	adminDAO *dao.AdminDAO
}

func NewAuthService() *AuthService {
	return &AuthService{adminDAO: dao.NewAdminDAO()}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	ExpireAt int64  `json:"expire_at"`
	Role     string `json:"role"`
	Username string `json:"username"`
}

func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	if err := common.SanitizeString(req.Username); err != nil {
		return nil, err
	}

	admin, err := s.adminDAO.FindByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, common.ErrInvalidInput("用户名或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		return nil, common.ErrInvalidInput("用户名或密码错误")
	}

	expireHours := config.Global.JWT.ExpireHours
	token, err := common.GenerateToken(admin.ID, admin.Username, common.AdminRole(admin.Role), config.Global.JWT.Secret, expireHours)
	if err != nil {
		return nil, err
	}

	expire := time.Duration(expireHours) * time.Hour
	if err := dao.SetAdminToken(admin.ID, token, expire); err != nil {
		return nil, err
	}

	expireAt := time.Now().Add(expire).Unix()
	return &LoginResponse{
		Token:    token,
		ExpireAt: expireAt,
		Role:     string(admin.Role),
		Username: admin.Username,
	}, nil
}

func (s *AuthService) Logout(adminID uint64) error {
	return dao.DeleteAdminToken(adminID)
}

// HashPassword 密码哈希
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// InitDefaultAdmin 初始化默认超级管理员（仅开发环境或首次部署）
func InitDefaultAdmin() error {
	admin, err := dao.NewAdminDAO().FindByUsername("admin")
	if err != nil {
		return err
	}
	if admin != nil {
		return nil
	}
	hash, err := HashPassword("admin123")
	if err != nil {
		return err
	}
	return dao.NewAdminDAO().Create(&model.Admin{
		Username:     "admin",
		PasswordHash: hash,
		Role:         model.AdminRoleSuperAdmin,
		Status:       1,
		RealName:     "超级管理员",
	})
}
