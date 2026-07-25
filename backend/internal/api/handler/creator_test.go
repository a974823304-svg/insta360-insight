package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"insta360-insight/internal/service"
	"insta360-insight/internal/service/source"
)

func TestCreatorHandlerKpi(t *testing.T) {
	h := NewCreator(service.NewCreatorService(source.NewMockAdapter()))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/creator/kpi", nil)
	h.Kpi(c)
	if w.Code != http.StatusOK {
		t.Fatalf("期望 200, 实际 %d", w.Code)
	}
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			Key string `json:"key"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json 解析失败: %v", err)
	}
	if resp.Code != 0 || len(resp.Data) != 5 {
		t.Fatalf("期望 code=0 且 5 条, 实际 code=%d len=%d", resp.Code, len(resp.Data))
	}
}
