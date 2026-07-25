package service

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"insta360-insight/internal/model"
	"insta360-insight/internal/store"
)

// Claims 是 JWT 载荷,同时被 handler/middleware 复用。
type Claims struct {
	UserID   int64  `json:"sub"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService 账号与令牌业务逻辑。
type AuthService struct {
	repo      *store.UserRepo
	jwtSecret string
}

// NewAuthService 构造。
func NewAuthService(repo *store.UserRepo, jwtSecret string) *AuthService {
	return &AuthService{repo: repo, jwtSecret: jwtSecret}
}

// Register 注册新账号(密码 bcrypt 哈希后落库)。
func (s *AuthService) Register(username, password, role string) (*model.User, error) {
	if len(password) < 6 {
		return nil, errors.New("密码至少 6 位")
	}
	if role == "" {
		role = "viewer"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.CreateUser(u); err != nil {
		// 唯一约束冲突(sqlite error 2067)统一为"用户名已存在"
		return nil, errors.New("用户名已存在")
	}
	return u, nil
}

// Login 校验凭证并签发 JWT。
func (s *AuthService) Login(username, password string) (string, *model.User, error) {
	u, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}
	token, err := s.sign(u)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

// GetProfile 取当前用户资料(含昵称/头像/联系方式/简介)。
func (s *AuthService) GetProfile(userID int64) (*model.User, error) {
	return s.repo.GetByID(userID)
}

// UpdateProfile 更新资料;昵称与联系方式必填,bio 限长 500。
func (s *AuthService) UpdateProfile(userID int64, nickname, avatar, contact, bio string) (*model.User, error) {
	if strings.TrimSpace(nickname) == "" {
		return nil, errors.New("昵称为必填项")
	}
	if strings.TrimSpace(contact) == "" {
		return nil, errors.New("联系方式为必填项")
	}
	if len([]rune(bio)) > 500 {
		return nil, errors.New("个人简介不能超过 500 字")
	}
	u, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	u.Nickname = nickname
	u.Avatar = avatar
	u.Contact = contact
	u.Bio = bio
	if err := s.repo.UpdateProfile(u.ID, u); err != nil {
		return nil, err
	}
	return u, nil
}

// ParseToken 校验并解析 JWT。
func (s *AuthService) ParseToken(token string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("非预期签名算法")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// SeedAdmin 表为空时插入 admin 账号,已存在则跳过(幂等)。返回该账号。
func (s *AuthService) SeedAdmin(user, pass string) (*model.User, error) {
	if existing, err := s.repo.GetByUsername(user); err == nil {
		return existing, nil
	}
	u, err := s.Register(user, pass, "admin")
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *AuthService) sign(u *model.User) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   u.ID,
		Username: u.Username,
		Role:     u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   itoa(u.ID),
			Issuer:    "insta360-insight",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(s.jwtSecret))
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
