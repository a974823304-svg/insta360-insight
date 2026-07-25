package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

// Auth 账号相关 HTTP 处理器(瘦层)。
type Auth struct {
	svc       *service.AuthService
	avatarDir string
}

func NewAuth(svc *service.AuthService, avatarDir string) *Auth {
	return &Auth{svc: svc, avatarDir: avatarDir}
}

type authReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// Register POST /api/auth/register
func (h *Auth) Register(c *gin.Context) {
	var q authReq
	if err := c.ShouldBindJSON(&q); err != nil || q.Username == "" || q.Password == "" {
		c.JSON(http.StatusOK, model.Fail(400, "用户名和密码必填"))
		return
	}
	u, err := h.svc.Register(q.Username, q.Password, q.Role)
	if err != nil {
		c.JSON(http.StatusOK, model.Fail(409, err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.OK(gin.H{"user": u}))
}

// Login POST /api/auth/login
func (h *Auth) Login(c *gin.Context) {
	var q authReq
	if err := c.ShouldBindJSON(&q); err != nil || q.Username == "" || q.Password == "" {
		c.JSON(http.StatusOK, model.Fail(400, "用户名和密码必填"))
		return
	}
	token, u, err := h.svc.Login(q.Username, q.Password)
	if err != nil {
		c.JSON(http.StatusOK, model.Fail(401, "用户名或密码错误"))
		return
	}
	c.JSON(http.StatusOK, model.OK(gin.H{"token": token, "user": u}))
}

// ProfileGet GET /api/user/profile —— 返回当前登录用户资料。
func (h *Auth) ProfileGet(c *gin.Context) {
	claims := c.MustGet("claims").(service.Claims)
	u, err := h.svc.GetProfile(claims.UserID)
	if err != nil {
		c.JSON(http.StatusOK, model.Fail(404, err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.OK(u))
}

// ProfileUpdate PUT /api/user/profile —— 更新昵称/头像/联系方式/简介。
func (h *Auth) ProfileUpdate(c *gin.Context) {
	var q struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
		Contact  string `json:"contact"`
		Bio      string `json:"bio"`
	}
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusOK, model.Fail(400, "请求格式错误"))
		return
	}
	claims := c.MustGet("claims").(service.Claims)
	u, err := h.svc.UpdateProfile(claims.UserID, q.Nickname, q.Avatar, q.Contact, q.Bio)
	if err != nil {
		c.JSON(http.StatusOK, model.Fail(400, err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.OK(u))
}

const avatarMaxBytes = 2 << 20 // 2MB

// allowedAvatarExt 按 Content-Type 返回允许的扩展名,不允许返回 ""。
func allowedAvatarExt(contentType string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

// AvatarUpload POST /api/user/avatar —— 接收图片文件,落盘后返回可访问 URL。
func (h *Auth) AvatarUpload(c *gin.Context) {
	claims := c.MustGet("claims").(service.Claims)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, model.Fail(400, "请选择要上传的图片"))
		return
	}
	if fileHeader.Size > avatarMaxBytes {
		c.JSON(http.StatusOK, model.Fail(400, "图片不能超过 2MB"))
		return
	}
	ext := allowedAvatarExt(fileHeader.Header.Get("Content-Type"))
	if ext == "" {
		c.JSON(http.StatusOK, model.Fail(400, "仅支持 PNG / JPG / WEBP 格式"))
		return
	}
	// 删除旧本地头像(若当前 avatar 为 /avatars/ 开头),避免磁盘膨胀
	if cur, gerr := h.svc.GetProfile(claims.UserID); gerr == nil && strings.HasPrefix(cur.Avatar, "/avatars/") {
		_ = os.Remove(filepath.Join(h.avatarDir, filepath.Base(cur.Avatar)))
	}
	filename := fmt.Sprintf("%d_%d%s", claims.UserID, time.Now().UnixNano(), ext)
	if err := c.SaveUploadedFile(fileHeader, filepath.Join(h.avatarDir, filename)); err != nil {
		c.JSON(http.StatusOK, model.Fail(500, "头像保存失败"))
		return
	}
	c.JSON(http.StatusOK, model.OK(gin.H{"url": "/avatars/" + filename}))
}
