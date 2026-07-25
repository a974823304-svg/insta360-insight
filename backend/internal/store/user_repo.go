// Package store 负责账号等持久化数据访问。当前用 SQLite(纯 Go,无 CGO)。
package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"time"

	"insta360-insight/internal/model"

	_ "modernc.org/sqlite" // 驱动名 "sqlite"
)

// UserRepo 封装 users 表的访问。
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo 打开(或创建)SQLite 文件库,自动建目录与建表。
func NewUserRepo(dbPath string) (*UserRepo, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	r := &UserRepo{db: db}
	if err := r.migrate(); err != nil {
		return nil, err
	}
	return r, nil
}

// userCols 是所有查询统一选取的列(含资料字段),避免 SELECT * 与 Scan 失配。
const userCols = "id, username, password_hash, role, created_at, nickname, avatar, contact, bio"

func (r *UserRepo) migrate() error {
	if _, err := r.db.Exec(`
CREATE TABLE IF NOT EXISTS users (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	username      TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	role          TEXT NOT NULL DEFAULT 'viewer',
	created_at    DATETIME NOT NULL
)`); err != nil {
		return err
	}
	// 幂等加列:先查已有列,只补缺失的,支持旧库升级不丢数据
	have, err := r.columns()
	if err != nil {
		return err
	}
	for _, c := range []string{"nickname TEXT DEFAULT ''", "avatar TEXT DEFAULT ''", "contact TEXT DEFAULT ''", "bio TEXT DEFAULT ''"} {
		name := strings.Fields(c)[0]
		if have[name] {
			continue
		}
		if _, err := r.db.Exec("ALTER TABLE users ADD COLUMN " + c); err != nil {
			return err
		}
	}
	return nil
}

// columns 返回 users 表现有列名集合(SQLite PRAGMA table_info)。
func (r *UserRepo) columns() (map[string]bool, error) {
	rows, err := r.db.Query("PRAGMA table_info(users)")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		m[name] = true
	}
	return m, nil
}

// CreateUser 插入新账号,并回填自增主键到 u.ID。
func (r *UserRepo) CreateUser(u *model.User) error {
	if u.CreatedAt.IsZero() {
		u.CreatedAt = timeNow()
	}
	res, err := r.db.Exec(
		`INSERT INTO users (username, password_hash, role, created_at) VALUES (?, ?, ?, ?)`,
		u.Username, u.PasswordHash, u.Role, u.CreatedAt,
	)
	if err != nil {
		return err
	}
	if id, err := res.LastInsertId(); err == nil {
		u.ID = id
	}
	return nil
}

// GetByUsername 按用户名查账号;不存在返回 sql.ErrNoRows。
func (r *UserRepo) GetByUsername(name string) (*model.User, error) {
	row := r.db.QueryRow(`SELECT `+userCols+` FROM users WHERE username = ?`, name)
	return r.scanUser(row)
}

// GetByID 按主键查账号(资料接口用)。
func (r *UserRepo) GetByID(id int64) (*model.User, error) {
	row := r.db.QueryRow(`SELECT `+userCols+` FROM users WHERE id = ?`, id)
	return r.scanUser(row)
}

// scanUser 统一解析 users 行。profile 列(nickname/avatar/contact/bio)可能为 NULL,
// 用 sql.NullString 容错,避免 "converting NULL to string" 报错, NULL 读作空串。
func (r *UserRepo) scanUser(row *sql.Row) (*model.User, error) {
	u := &model.User{}
	var nick, avatar, contact, bio sql.NullString
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt,
		&nick, &avatar, &contact, &bio); err != nil {
		return nil, err
	}
	u.Nickname = nick.String
	u.Avatar = avatar.String
	u.Contact = contact.String
	u.Bio = bio.String
	return u, nil
}

// UpdateProfile 只更新资料列(昵称/头像/联系方式/简介),不动账号/密码/角色。
func (r *UserRepo) UpdateProfile(id int64, u *model.User) error {
	_, err := r.db.Exec(
		`UPDATE users SET nickname=?, avatar=?, contact=?, bio=? WHERE id=?`,
		u.Nickname, u.Avatar, u.Contact, u.Bio, id)
	return err
}

// Count 返回用户总数(种子判断表空用)。
func (r *UserRepo) Count() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

// Close 释放连接(进程退出时调用)。
func (r *UserRepo) Close() error { return r.db.Close() }

// timeNow 抽出来便于测试;生产用 time.Now。
var timeNow = func() time.Time { return time.Now() }
