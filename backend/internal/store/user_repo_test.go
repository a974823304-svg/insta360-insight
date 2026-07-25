package store

import (
	"path/filepath"
	"testing"

	"insta360-insight/internal/model"
)

func newTestRepo(t *testing.T) *UserRepo {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	repo, err := NewUserRepo(dbPath)
	if err != nil {
		t.Fatalf("NewUserRepo: %v", err)
	}
	t.Cleanup(func() { repo.Close() })
	return repo
}

func TestMigrateIdempotent(t *testing.T) {
	repo := newTestRepo(t)
	if err := repo.migrate(); err != nil { // 第二次迁移应安全(列已存在)
		t.Fatalf("migrate 2nd: %v", err)
	}
}

func TestProfilePersistRoundTrip(t *testing.T) {
	repo := newTestRepo(t)
	if err := repo.CreateUser(&model.User{Username: "alice", PasswordHash: "x", Role: "viewer"}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	u, err := repo.GetByUsername("alice")
	if err != nil {
		t.Fatalf("GetByUsername: %v", err)
	}
	u.Nickname, u.Avatar, u.Contact, u.Bio = "Alice", "preset:blue", "wx:alice", "hi"
	if err := repo.UpdateProfile(u.ID, u); err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	again, err := repo.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if again.Nickname != "Alice" || again.Avatar != "preset:blue" || again.Contact != "wx:alice" || again.Bio != "hi" {
		t.Fatalf("profile not persisted: %+v", again)
	}
}

func TestCreateUserSetsID(t *testing.T) {
	repo := newTestRepo(t)
	u := &model.User{Username: "idtest", PasswordHash: "x", Role: "viewer"}
	if err := repo.CreateUser(u); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID <= 0 {
		t.Fatalf("expected positive ID after CreateUser, got %d", u.ID)
	}
	got, err := repo.GetByID(u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Username != "idtest" {
		t.Fatalf("unexpected user %+v", got)
	}
}
