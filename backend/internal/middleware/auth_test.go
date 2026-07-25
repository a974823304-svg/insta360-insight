// backend/internal/middleware/auth_test.go
package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/model"
	"insta360-insight/internal/service"
)

type fakeValidator struct {
	token string
	err   error
}

func (f fakeValidator) ParseToken(t string) (*service.Claims, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &service.Claims{UserID: 1, Username: "u", Role: "admin"}, nil
}

func okHandler() gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(200, model.OK("ok")) }
}

func TestJWTAuth_ValidTokenPasses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(fakeValidator{token: "good"}, false, service.Claims{}))
	r.GET("/x", okHandler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer good")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestJWTAuth_MissingTokenRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(fakeValidator{}, false, service.Claims{}))
	r.GET("/x", okHandler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_InvalidTokenRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(fakeValidator{err: ErrFake}, false, service.Claims{}))
	r.GET("/x", okHandler())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer bad")
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_DisableInjectsDevAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(fakeValidator{err: ErrFake}, true, service.Claims{}))
	r.GET("/x", func(c *gin.Context) {
		cl := c.MustGet("claims").(service.Claims)
		c.JSON(200, model.OK(cl.Username))
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

var ErrFake = errors.New("fake")
