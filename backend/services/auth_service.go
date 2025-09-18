package services

import (
	"errors"
	"os"
	"time"

	"iptables-management-backend/config"
	"iptables-management-backend/models"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	Username    string `json:"username"`
	Role        string `json:"role"`
}

// Claims JWT声明结构
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	return &AuthService{}
}

// Login 用户登录
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	var user models.User
	
	// 查找用户
	if err := config.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 生成JWT token
	token, err := s.generateToken(user.Username, user.Role)
	if err != nil {
		return nil, errors.New("生成token失败")
	}

	return &LoginResponse{
		AccessToken: token,
		Username:    user.Username,
		Role:        user.Role,
	}, nil
}

// generateToken 生成JWT token
func (s *AuthService) generateToken(username, role string) (string, error) {
	claims := Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// HashPassword 加密密码
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CreateDefaultUsers 创建默认用户
func (s *AuthService) CreateDefaultUsers() error {
	// 检查是否已存在管理员用户
	var count int64
	config.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	
	if count > 0 {
		return nil // 已存在管理员用户
	}

	// 创建默认管理员
	adminPassword, _ := s.HashPassword("admin123")
	admin := models.User{
		Username: "admin",
		Password: adminPassword,
		Role:     "admin",
	}

	// 创建默认普通用户
	userPassword, _ := s.HashPassword("user123")
	user := models.User{
		Username: "user1",
		Password: userPassword,
		Role:     "user",
	}

	// 批量创建用户
	return config.DB.Create([]models.User{admin, user}).Error
}