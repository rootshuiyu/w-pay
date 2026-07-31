package router

import (
	"os"

	"wpay/common"
	"wpay/common/middleware"
	"wpay/controller"

	"github.com/gin-gonic/gin"
)

// Setup 路由分组（交付文档 §5.2）
func Setup(mode string) *gin.Engine {
	gin.SetMode(mode)
	r := gin.New()

	// 可信反向代理由白名单 registry 热更新（见 middleware.BindEngine）
	middleware.BindEngine(r)
	middleware.ReapplyTrustedProxies()

	r.Use(middleware.Recovery())
	r.Use(middleware.RateLimit())

	// 管理后台前端（web/dist 存在时托管 SPA，hash 路由无需 fallback）
	if _, err := os.Stat("web/dist/index.html"); err == nil {
		r.GET("/", func(c *gin.Context) {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
			c.File("./web/dist/index.html")
		})
		r.Static("/assets", "./web/dist/assets")
	} else {
		r.GET("/", controller.Index)
	}
	r.GET("/health", controller.Health)

	authCtrl := controller.NewAuthController()
	storeCtrl := controller.NewStoreController()
	channelCtrl := controller.NewPayChannelController()
	platformCtrl := controller.NewPayPlatformController()
	orderCtrl := controller.NewOrderController()
	callbackCtrl := controller.NewCallbackController()
	reconcileCtrl := controller.NewReconcileController()
	whitelistCtrl := controller.NewIPWhitelistController()

	// ① 收银端开放接口（内网收银机；可配置来源 IP 白名单，默认不限制）
	payGroup := r.Group("/api/pay")
	payGroup.Use(middleware.DynamicIPWhitelist("pay"))
	{
		payGroup.POST("/create", orderCtrl.PayCreate)
		payGroup.GET("/query", orderCtrl.PayQuery)
	}

	// ② 支付回调公开接口（验签 + 幂等，无登录；生产限支付平台官方 IP）
	notifyGroup := r.Group("/api/notify")
	notifyGroup.Use(middleware.DynamicIPWhitelist("callback"))
	{
		notifyGroup.POST("/wx", callbackCtrl.Wechat)
		notifyGroup.POST("/alipay", callbackCtrl.Alipay)
	}

	// ③ 后台鉴权接口（必须 Token，分级权限；IP 白名单可后台热更新）
	admin := r.Group("/api/admin")
	{
		admin.POST("/login", authCtrl.Login)

		// 白名单管理：仅 super_admin，且不受 IP 白名单拦截（避免误配后无法自救）
		wlGroup := admin.Group("/whitelist")
		wlGroup.Use(middleware.AuthToken())
		wlGroup.Use(middleware.RequireRole(common.RoleSuperAdmin))
		{
			wlGroup.GET("/overview", whitelistCtrl.Overview)
			wlGroup.PUT("/policy", whitelistCtrl.UpdatePolicy)
			wlGroup.POST("/add", whitelistCtrl.Add)
			wlGroup.PUT("/edit", whitelistCtrl.Edit)
			wlGroup.PUT("/status", whitelistCtrl.UpdateStatus)
			wlGroup.DELETE("/del", whitelistCtrl.Del)
		}

		protected := admin.Group("")
		protected.Use(middleware.DynamicIPWhitelist("admin"))
		protected.Use(middleware.AuthToken())
		{
			protected.POST("/logout", authCtrl.Logout)
			protected.GET("/profile", authCtrl.Profile)

			// 门店管理（无限增删改查）
			storeGroup := protected.Group("/store")
			storeGroup.Use(middleware.RequireRole(common.RoleSuperAdmin, common.RoleOperator))
			{
				storeGroup.POST("/add", storeCtrl.Add)
				storeGroup.PUT("/edit", storeCtrl.Edit)
				storeGroup.GET("/list", storeCtrl.List)
				storeGroup.PUT("/status", storeCtrl.UpdateStatus)
				storeGroup.DELETE("/del", storeCtrl.Del)
			}

			// 门店支付渠道（可随时更换商户配置）
			channelGroup := protected.Group("/channel")
			channelGroup.Use(middleware.RequireRole(common.RoleSuperAdmin))
			{
				channelGroup.POST("/add", channelCtrl.Add)
				channelGroup.POST("/pool-add", channelCtrl.PoolAdd)
				channelGroup.PUT("/edit", channelCtrl.Edit)
				channelGroup.GET("/list", channelCtrl.List)
				channelGroup.GET("/pool", channelCtrl.Pool)
				channelGroup.PUT("/status", channelCtrl.UpdateStatus)
			}

			platformGroup := protected.Group("/platform")
			platformGroup.Use(middleware.RequireRole(common.RoleSuperAdmin))
			{
				platformGroup.POST("/add", platformCtrl.Add)
				platformGroup.POST("/quick-setup", platformCtrl.QuickSetup)
				platformGroup.POST("/add-channel", platformCtrl.AddChannel)
				platformGroup.GET("/detail", platformCtrl.Detail)
				platformGroup.PUT("/edit", platformCtrl.Edit)
				platformGroup.GET("/list", platformCtrl.List)
				platformGroup.PUT("/status", platformCtrl.UpdateStatus)
				platformGroup.DELETE("/del", platformCtrl.Del)
				platformGroup.GET("/channels", platformCtrl.Channels)
				platformGroup.GET("/available-channels", platformCtrl.AvailableChannels)
				platformGroup.PUT("/set-channels", platformCtrl.SetChannels)
				platformGroup.PUT("/bind-channels", platformCtrl.BindChannels)
				platformGroup.PUT("/unbind-channel", platformCtrl.UnbindChannel)
				platformGroup.PUT("/regenerate-key", platformCtrl.RegenerateKey)
				platformGroup.GET("/pool", platformCtrl.Pool)
			}

			// 订单查询
			orderGroup := protected.Group("/order")
			{
				orderGroup.GET("/list", orderCtrl.List)
				orderGroup.GET("/detail", orderCtrl.QueryByParam)
			}

			// 对账统计与导出
			statGroup := protected.Group("/stat")
			statGroup.Use(middleware.RequireRole(common.RoleSuperAdmin, common.RoleFinance))
			{
				statGroup.GET("/summary", reconcileCtrl.Stat)
			}

			exportGroup := protected.Group("/export")
			exportGroup.Use(middleware.RequireRole(common.RoleSuperAdmin, common.RoleFinance))
			{
				exportGroup.GET("/orders", reconcileCtrl.ExportOrders)
				exportGroup.GET("/stat", reconcileCtrl.ExportStat)
			}
		}
	}

	return r
}
