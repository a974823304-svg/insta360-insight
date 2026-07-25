package router

import (
	"github.com/gin-gonic/gin"

	"insta360-insight/internal/api/handler"
	"insta360-insight/internal/middleware"
	"insta360-insight/internal/service"
)

// New 构造完整 *gin.Engine。
// authSvc 用于挂载登录/注册端点与 JWT 中间件;disableAuth 控制是否跳过鉴权。
func New(insightSvc *service.InsightService, aiSvc *service.AIService, authSvc *service.AuthService, creatorSvc *service.CreatorService, disableAuth bool, devUser service.Claims, avatarDir string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	health := handler.NewHealth()
	insight := handler.NewInsight(insightSvc, aiSvc)
	auth := handler.NewAuth(authSvc, avatarDir)
	creator := handler.NewCreator(creatorSvc)

	// 健康检查 + 登录/注册:公开
	r.GET("/api/health", health.Handle)
	r.POST("/api/auth/register", auth.Register)
	r.POST("/api/auth/login", auth.Login)

	// 数据洞察组:受 JWTAuth 保护(health 已在外面,不受影响)
	g := r.Group("/api")
	g.Use(middleware.JWTAuth(authSvc, disableAuth, devUser))
	{
		g.GET("/kpi", insight.Kpi)
		g.GET("/views-trend", insight.ViewsTrend)
		g.GET("/platform-distribution", insight.PlatformShare)
		g.GET("/track-performance", insight.TrackPerformance)
		g.GET("/explosive-radar", insight.Radar)
		g.GET("/audience-age", insight.AudienceAge)
		g.GET("/top-creators", insight.TopCreators)
		g.GET("/insights", insight.Insights)
		g.GET("/filters/options", insight.FilterOptions)
		g.GET("/user/profile", auth.ProfileGet)
		g.PUT("/user/profile", auth.ProfileUpdate)
		g.POST("/user/avatar", auth.AvatarUpload)
		g.GET("/creator/kpi", creator.Kpi)
		g.GET("/creator/trend", creator.Trend)
		g.GET("/creator/platforms", creator.Platforms)
		g.GET("/creator/tracks", creator.Tracks)
		g.GET("/creator/audience", creator.Audience)
		g.GET("/creator/list", creator.List)
	}

	r.Static("/avatars", avatarDir)

	return r
}
