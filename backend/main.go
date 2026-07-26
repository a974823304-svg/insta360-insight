// Package main 是 Insta360 达人营销数据洞察平台后端服务的入口。
package main

import (
	"log"
	"os"

	"insta360-insight/internal/api/router"
	"insta360-insight/internal/middleware"
	"insta360-insight/internal/service"
	"insta360-insight/internal/service/source"
	"insta360-insight/internal/store"
)

func main() {
	// 1. 选择数据源(阶段二:可插拔接入层 adapter)
	srcKind := os.Getenv("SOURCE")
	adapter, err := source.NewDataSource(srcKind)
	if err != nil {
		log.Fatalf("初始化数据源失败: %v", err)
	}
	log.Printf("[insta360-insight] data source = %s", adapter.Name())

	// 2. 装载业务服务(依赖注入,后续方便做单测)
	insightSvc := service.NewInsightService(adapter)
	aiSvc := service.NewAIService(adapter)
	creatorSvc := service.NewCreatorService(adapter)
	contentSvc := service.NewContentService(adapter)
	marketSvc := service.NewMarketService(adapter)
	brandSvc := service.NewBrandService(adapter)

	// 3. 账号与鉴权(阶段一)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-only-insecure-secret-change-me"
		log.Println("[WARN] JWT_SECRET 未设置,使用开发默认值;生产环境必须设置强随机值")
	}
	disableAuth := os.Getenv("AUTH_DISABLE") == "1" || os.Getenv("ENV") == "dev"

	avatarDir := envOrDefault("AVATAR_DIR", "data/avatars")
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		log.Fatalf("创建头像目录失败: %v", err)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/app.db"
	}
	userRepo, err := store.NewUserRepo(dbPath)
	if err != nil {
		log.Fatalf("初始化用户库失败: %v", err)
	}
	defer userRepo.Close()

	authSvc := service.NewAuthService(userRepo, jwtSecret)
	seedUser := envOrDefault("SEED_ADMIN_USER", "admin")
	seedPass := envOrDefault("SEED_ADMIN_PASS", "insta360")
	admin, err := authSvc.SeedAdmin(seedUser, seedPass)
	if err != nil {
		log.Fatalf("种子管理员失败: %v", err)
	}
	// dev 模式(AUTH_DISABLE=1/ENV=dev)下,JWTAuth 会注入 devUser 作为当前身份,
	// 必须是真实账号(否则资料接口会落到不存在的 user 0)。
	var devUser service.Claims
	if admin != nil {
		devUser = service.Claims{UserID: admin.ID, Username: admin.Username, Role: admin.Role}
	}

	// 4. 注册路由 & 全局中间件(鉴权在 router 内部按组挂载)
	engine := router.New(insightSvc, aiSvc, authSvc, creatorSvc, contentSvc, marketSvc, brandSvc, disableAuth, devUser, avatarDir)
	engine.Use(middleware.CORS(), middleware.AccessLog())

	// 5. 启动 HTTP 服务,监听 8080
	addr := ":8080"
	log.Printf("[insta360-insight] backend listening on %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
