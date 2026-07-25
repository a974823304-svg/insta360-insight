package service

import (
	"testing"

	"insta360-insight/internal/store"
)

func newTestAuth(t *testing.T) *AuthService {
	t.Helper()
	repo, err := store.NewUserRepo(t.TempDir() + "/a.db")
	if err != nil {
		t.Fatalf("repo: %v", err)
	}
	t.Cleanup(func() { repo.Close() })
	return NewAuthService(repo, "test-secret")
}

func TestRegisterAndLogin(t *testing.T) {
	s := newTestAuth(t)
	u, err := s.Register("alice", "secret123", "viewer")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if u.PasswordHash != "" && u.Role != "viewer" {
		t.Fatalf("unexpected user %+v", u)
	}
	token, got, err := s.Login("alice", "secret123")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if token == "" || got.Username != "alice" {
		t.Fatalf("bad login result token=%q user=%+v", token, got)
	}
}

func TestLoginWrongPasswordFails(t *testing.T) {
	s := newTestAuth(t)
	_, _ = s.Register("bob", "secret123", "viewer")
	if _, _, err := s.Login("bob", "wrong"); err == nil {
		t.Fatal("expected login with wrong password to fail")
	}
}

func TestRegisterWeakPasswordFails(t *testing.T) {
	s := newTestAuth(t)
	if _, err := s.Register("c", "123", "viewer"); err == nil {
		t.Fatal("expected weak password to fail")
	}
}

func TestRegisterDuplicateFails(t *testing.T) {
	s := newTestAuth(t)
	_, _ = s.Register("d", "secret123", "viewer")
	if _, err := s.Register("d", "secret123", "viewer"); err == nil {
		t.Fatal("expected duplicate username to fail")
	}
}

func TestParseTokenRoundTrip(t *testing.T) {
	s := newTestAuth(t)
	_, _ = s.Register("e", "secret123", "admin")
	token, _, _ := s.Login("e", "secret123")
	claims, err := s.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if claims.Username != "e" || claims.Role != "admin" {
		t.Fatalf("bad claims %+v", claims)
	}
}

func TestParseTokenRejectsGarbage(t *testing.T) {
	s := newTestAuth(t)
	if _, err := s.ParseToken("not-a-token"); err == nil {
		t.Fatal("expected garbage token to fail")
	}
}

func TestSeedAdminIdempotent(t *testing.T) {
	s := newTestAuth(t)
	if _, err := s.SeedAdmin("admin", "insta360"); err != nil {
		t.Fatalf("seed1: %v", err)
	}
	if _, err := s.SeedAdmin("admin", "insta360"); err != nil {
		t.Fatalf("seed2(idempotent) should not error: %v", err)
	}
}

func TestUpdateProfileFlow(t *testing.T) {
	s := newTestAuth(t)
	if _, err := s.Register("u1", "abc12345", "viewer"); err != nil {
		t.Fatalf("register: %v", err)
	}
	_, u, err := s.Login("u1", "abc12345")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	// 缺昵称
	if _, err := s.UpdateProfile(u.ID, "", "preset:x", "wx:x", "bio"); err == nil {
		t.Fatal("expected error when nickname empty")
	}
	// 缺联系方式
	if _, err := s.UpdateProfile(u.ID, "Name", "preset:x", "", "bio"); err == nil {
		t.Fatal("expected error when contact empty")
	}
	// 正常
	upd, err := s.UpdateProfile(u.ID, "Name", "preset:blue", "wx:x", "hello")
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if upd.Nickname != "Name" || upd.Contact != "wx:x" || upd.Bio != "hello" {
		t.Fatalf("mismatch: %+v", upd)
	}
	// 取回确认持久化
	got, err := s.GetProfile(u.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Nickname != "Name" || got.Contact != "wx:x" {
		t.Fatalf("not persisted: %+v", got)
	}
}
