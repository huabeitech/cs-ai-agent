package bootstrap

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"cs-agent/internal/ai/mcps"
	_ "cs-agent/internal/ai/runtime"
	"cs-agent/internal/controllers/api"
	"cs-agent/internal/controllers/dashboard"
	"cs-agent/internal/controllers/third"
	"cs-agent/internal/middleware"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/cors"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/mvc"

	_ "cs-agent/internal/services/wx_callback_handlers"
)

func NewServer() (*iris.Application, error) {
	cfg := config.Current()
	app := iris.New()
	corsHandler := cors.New().
		AllowOrigin("*").
		AllowHeaders("Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Guest-Id", "X-Channel-Id", "X-External-Id", "X-External-Name").
		MaxAge(600).
		ExposeHeaders("Content-Length", "Content-Type", "Authorization", "X-Guest-Id", "X-Channel-Id", "X-External-Id", "X-External-Name").
		Handler()
	app.UseRouter(func(ctx iris.Context) {
		// WebSocket upgrade is validated by the upgrader's origin policy.
		if isWebsocketUpgrade(ctx) {
			ctx.Next()
			return
		}
		corsHandler(ctx)
	})
	app.UseRouter(recover.New())
	app.UseRouter(func(ctx iris.Context) {
		start := time.Now()
		path := ctx.Path()
		method := ctx.Method()
		ctx.Next()

		slog.Info("http request",
			"method", method,
			"path", path,
			"status", ctx.GetStatusCode(),
			"elapsed", time.Since(start).Milliseconds(),
			"clientIp", ctx.RemoteAddr(),
		)
	})
	app.UseRouter(func(ctx iris.Context) {
		ctx.SetMaxRequestBodySize(cfg.Storage.MaxRequestBodySizeBytes())
		ctx.Next()
	})

	// 注册路由
	addRouter(app)

	// 注册本地存储静态资源服务
	app.HandleDir(cfg.Storage.Local.BaseURL, iris.Dir(cfg.Storage.Local.Root), iris.DirOptions{
		ShowList: false,
	})

	// 注册dashboard静态资源服务
	app.HandleDir("/", iris.Dir("web/out"), iris.DirOptions{
		IndexName: "index.html",
		Compress:  true,
		ShowList:  false,
	})

	return app, nil
}

func isWebsocketUpgrade(ctx iris.Context) bool {
	if !strings.EqualFold(ctx.GetHeader("Upgrade"), "websocket") {
		return false
	}
	return strings.Contains(strings.ToLower(ctx.GetHeader("Connection")), "upgrade")
}

func addRouter(app *iris.Application) {
	mcpHandler := mcps.NewHTTPHandler()

	app.Any("/api/mcp", iris.FromStd(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mcpHandler.ServeHTTP(w, r)
	})))

	mvc.Configure(app.Party("/api"), func(m *mvc.Application) {
		m.Party("/auth").Handle(new(api.AuthController))
		m.Party("/channel").Handle(new(api.ChannelController))
		m.Party("/conversation", middleware.ExternalUserMiddleware).Handle(new(api.ConversationController))
		m.Party("/message", middleware.ExternalUserMiddleware).Handle(new(api.MessageController))
	})

	mvc.Configure(app.Party("/api/ws"), func(m *mvc.Application) {
		m.Router.Get("/dashboard", middleware.AuthMiddleware, services.WsService.HandleDashboardWS)
		m.Router.Get("/dashboard/notification", middleware.AuthMiddleware, services.WsService.HandleDashboardNotificationWS)
		m.Router.Get("/open", services.WsService.HandleOpenWS)
	})

	mvc.Configure(app.Party("/api/dashboard", middleware.AuthMiddleware), func(m *mvc.Application) {
		m.Party("/dashboard").Handle(new(dashboard.DashboardController))
		m.Party("/user").Handle(new(dashboard.UserController))
		m.Party("/company").Handle(new(dashboard.CompanyController))
		m.Party("/customer").Handle(new(dashboard.CustomerController))
		m.Party("/customer-contact").Handle(new(dashboard.CustomerContactController))
		m.Party("/role").Handle(new(dashboard.RoleController))
		m.Party("/permission").Handle(new(dashboard.PermissionController))
		m.Party("/session").Handle(new(dashboard.SessionController))
		m.Party("/tag").Handle(new(dashboard.TagController))
		m.Party("/conversation").Handle(new(dashboard.ConversationController))
		m.Party("/ticket").Handle(new(dashboard.TicketController))
		m.Party("/notification").Handle(new(dashboard.NotificationController))
		m.Party("/ticket-resolution-code").Handle(new(dashboard.TicketResolutionCodeController))
		m.Party("/ticket-priority-config").Handle(new(dashboard.TicketPriorityConfigController))
		m.Party("/quick-reply").Handle(new(dashboard.QuickReplyController))
		m.Party("/channel").Handle(new(dashboard.ChannelController))
		m.Party("/agent").Handle(new(dashboard.AgentController))
		m.Party("/agent-team").Handle(new(dashboard.AgentTeamController))
		m.Party("/agent-team-schedule").Handle(new(dashboard.AgentTeamScheduleController))
		m.Party("/ai-agent").Handle(new(dashboard.AIAgentController))
		m.Party("/ai-config").Handle(new(dashboard.AIConfigController))
		m.Party("/asset").Handle(new(dashboard.AssetController))
		m.Party("/knowledge-base").Handle(new(dashboard.KnowledgeBaseController))
		m.Party("/knowledge-document").Handle(new(dashboard.KnowledgeDocumentController))
		m.Party("/knowledge-faq").Handle(new(dashboard.KnowledgeFAQController))
		m.Party("/knowledge-retrieve").Handle(new(dashboard.KnowledgeRetrieveController))
		m.Party("/knowledge-retrieve-log").Handle(new(dashboard.KnowledgeRetrieveLogController))
		m.Party("/agent-run-log").Handle(new(dashboard.AgentRunLogController))
		m.Party("/skill-definition").Handle(new(dashboard.SkillDefinitionController))
		m.Party("/mcp").Handle(new(dashboard.MCPController))
	})

	mvc.Configure(app.Party("/api/third"), func(m *mvc.Application) {
		m.Party("/wechat").Handle(new(third.WechatController))
	})
}
