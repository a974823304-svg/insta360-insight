package model

import "testing"

func TestFail_WrapsCodeAndMessage(t *testing.T) {
	resp := Fail(401, "未授权")
	if resp.Code != 401 {
		t.Fatalf("expected code 401, got %d", resp.Code)
	}
	if resp.Message != "未授权" {
		t.Fatalf("expected message 未授权, got %q", resp.Message)
	}
	if resp.Data != nil {
		t.Fatalf("expected nil data on error, got %v", resp.Data)
	}
}
