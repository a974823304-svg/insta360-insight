package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
	"insta360-insight/internal/store"
)

func setupAuthEngine(t *testing.T) (*gin.Engine, *service.AuthService) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	repo, err := store.NewUserRepo(t.TempDir() + "/h.db")
	if err != nil {
		t.Fatalf("repo: %v", err)
	}
	t.Cleanup(func() { repo.Close() })
	svc := service.NewAuthService(repo, "test-secret")
	r := gin.New()
	a := NewAuth(svc, t.TempDir())
	r.POST("/api/auth/register", a.Register)
	r.POST("/api/auth/login", a.Login)
	return r, svc
}

func TestLoginSuccess(t *testing.T) {
	r, svc := setupAuthEngine(t)
	_, _ = svc.Register("u1", "secret123", "admin")

	body, _ := json.Marshal(map[string]string{"username": "u1", "password": "secret123"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp model.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d msg=%s", resp.Code, resp.Message)
	}
	data := resp.Data.(map[string]interface{})
	if data["token"] == nil || data["token"] == "" {
		t.Fatal("expected token in data")
	}
	ub, _ := json.Marshal(data["user"])
	if bytes.Contains(ub, []byte("password_hash")) {
		t.Fatal("user JSON must not contain password_hash")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	r, svc := setupAuthEngine(t)
	_, _ = svc.Register("u2", "secret123", "viewer")
	body, _ := json.Marshal(map[string]string{"username": "u2", "password": "bad"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var resp model.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code == 0 {
		t.Fatal("expected non-zero code for wrong password")
	}
}

func TestRegisterEndpoint(t *testing.T) {
	r, _ := setupAuthEngine(t)
	body, _ := json.Marshal(map[string]string{"username": "new", "password": "secret123"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var resp model.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("register failed: %s", resp.Message)
	}
}

// 最小 1x1 PNG,用于构造合法上传体
var minPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89,
	0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00, 0x05, 0x00,
	0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

func setupAvatarEngine(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	repo, err := store.NewUserRepo(t.TempDir() + "/av.db")
	if err != nil {
		t.Fatalf("repo: %v", err)
	}
	t.Cleanup(func() { repo.Close() })
	svc := service.NewAuthService(repo, "test-secret")
	u, err := svc.Register("tester", "secret123", "admin")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	avatarDir := t.TempDir()
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	r := gin.New()
	a := NewAuth(svc, avatarDir)
	r.Use(func(c *gin.Context) {
		c.Set("claims", service.Claims{UserID: u.ID, Username: u.Username, Role: u.Role})
		c.Next()
	})
	r.POST("/api/user/avatar", a.AvatarUpload)
	return r, avatarDir
}

func TestAvatarUploadSuccess(t *testing.T) {
	r, avatarDir := setupAvatarEngine(t)
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	// 真实浏览器上传时,part 的 Content-Type 会带上 image/png;这里显式设置以贴近真实请求。
	ph := make(textproto.MIMEHeader)
	ph.Set("Content-Disposition", `form-data; name="file"; filename="a.png"`)
	ph.Set("Content-Type", "image/png")
	fw, _ := w.CreatePart(ph)
	fw.Write(minPNG)
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/user/avatar", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var resp model.APIResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d msg=%s", resp.Code, resp.Message)
	}
	data := resp.Data.(map[string]interface{})
	url, _ := data["url"].(string)
	if !strings.HasPrefix(url, "/avatars/") {
		t.Fatalf("expected /avatars/ url, got %q", url)
	}
	if _, err := os.Stat(filepath.Join(avatarDir, filepath.Base(url))); err != nil {
		t.Fatalf("avatar file not saved: %v", err)
	}
}

func TestAvatarUploadRejectsNonImage(t *testing.T) {
	r, _ := setupAvatarEngine(t)
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, _ := w.CreateFormFile("file", "a.txt")
	fw.Write([]byte("not an image"))
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/user/avatar", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var resp model.APIResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Code == 0 {
		t.Fatal("expected non-zero code for non-image upload")
	}
}

func TestAvatarUploadMissingFile(t *testing.T) {
	r, _ := setupAvatarEngine(t)
	req := httptest.NewRequest(http.MethodPost, "/api/user/avatar", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var resp model.APIResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Code == 0 {
		t.Fatal("expected non-zero code when file missing")
	}
}
