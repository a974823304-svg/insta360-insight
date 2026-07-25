package model

import "time"

// User 平台账号。PasswordHash 禁止出现在任何 JSON 响应中。
type User struct {
	ID           int64     `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Role         string    `db:"role" json:"role"` // "admin" | "viewer"
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	// 个人资料(profile 功能)
	Nickname string `db:"nickname" json:"nickname"`
	Avatar   string `db:"avatar"   json:"avatar"` // URL 或 "preset:xxx"
	Contact  string `db:"contact"  json:"contact"` // 自由文本:电话/微信/邮箱
	Bio      string `db:"bio"      json:"bio"`
}
