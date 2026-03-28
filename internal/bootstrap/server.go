package bootstrap

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"cs-agent/internal/ai/mcps"
	"cs-agent/internal/ai/rag/vectordb"
	"cs-agent/internal/controllers/api"
	"cs-agent/internal/controllers/console"
	"cs-agent/internal/controllers/open"
	"cs-agent/internal/pkg/config"
	"cs-agent/internal/pkg/logx"
	"cs-agent/internal/services/cronx"
	"cs-agent/internal/wxwork"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/cors"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/mvc"
)

func NewServer(configPath string) (*iris.Application, *config.Config, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, nil, err
	}

	logx.Init(logx.Config{
		Level:     cfg.Logger.Level,
		Format:    cfg.Logger.Format,
		AddSource: cfg.Logger.AddSource,
	})

	if _, err = InitDB(cfg.DB); err != nil {
		return nil, nil, err
	}
	if err = InitMigrations(); err != nil {
		return nil, nil, err
	}
	if err = vectordb.Init(&cfg.VectorDB); err != nil {
		return nil, nil, err
	}
	wxwork.Init(cfg)

	// 启动任务调度器
	cronx.Init()

	app := iris.New()
	corsHandler := cors.New().
		AllowOrigin("*").
		AllowHeaders("Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Visitor-Id", "X-Widget-App-Id", "X-External-Source", "X-External-Id", "X-External-Name").
		MaxAge(600).
		ExposeHeaders("Content-Length", "Content-Type", "Authorization", "X-Visitor-Id", "X-Widget-App-Id", "X-External-Source", "X-External-Id", "X-External-Name").
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
	addRouter(app, cfg)

	// 注册本地存储静态资源服务
	app.HandleDir(cfg.Storage.Local.BaseURL, iris.Dir(cfg.Storage.Local.Root), iris.DirOptions{
		ShowList: false,
	})

	// 注册web静态资源服务
	app.HandleDir("/", iris.Dir("web/out"), iris.DirOptions{
		IndexName: "index.html",
		Compress:  true,
		ShowList:  false,
	})

	// 注册widget静态资源服务
	app.HandleDir("/widget", iris.Dir("widget/out"), iris.DirOptions{
		IndexName: "index.html",
		Compress:  true,
		ShowList:  false,
	})

	return app, cfg, nil
}

func isWebsocketUpgrade(ctx iris.Context) bool {
	if !strings.EqualFold(ctx.GetHeader("Upgrade"), "websocket") {
		return false
	}
	return strings.Contains(strings.ToLower(ctx.GetHeader("Connection")), "upgrade")
}

func addRouter(app *iris.Application, cfg *config.Config) {
	mcpHandler := mcps.NewHTTPHandler(cfg)

	app.Any("/api/mcp", iris.FromStd(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mcpHandler.ServeHTTP(w, r)
	})))

	app.Get("/api/admin/ws", AuthMiddleware, api.HandleAdminWebsocket)
	app.Get("/api/open/im/ws", open.HandleImWebsocket)

	mvc.Configure(app.Party("/api/auth"), func(m *mvc.Application) {
		m.Register(cfg)
		m.Handle(new(api.AuthController))
	})

	mvc.Configure(app.Party("/api/console", AuthMiddleware), func(m *mvc.Application) {
		m.Register(cfg)
		m.Party("/dashboard").Handle(new(console.DashboardController))
		m.Party("/user").Handle(new(console.UserController))
		m.Party("/company").Handle(new(console.CompanyController))
		m.Party("/customer").Handle(new(console.CustomerController))
		m.Party("/role").Handle(new(console.RoleController))
		m.Party("/permission").Handle(new(console.PermissionController))
		m.Party("/session").Handle(new(console.SessionController))
		m.Party("/tag").Handle(new(console.TagController))
		m.Party("/conversation").Handle(new(console.ConversationController))
		m.Party("/quick-reply").Handle(new(console.QuickReplyController))
		m.Party("/widget-site").Handle(new(console.WidgetSiteController))
		m.Party("/agent").Handle(new(console.AgentController))
		m.Party("/agent-team").Handle(new(console.AgentTeamController))
		m.Party("/agent-team-schedule").Handle(new(console.AgentTeamScheduleController))
		m.Party("/ai-agent").Handle(new(console.AIAgentController))
		m.Party("/ai-config").Handle(new(console.AIConfigController))
		m.Party("/asset").Handle(new(console.AssetController))
		m.Party("/knowledge-base").Handle(new(console.KnowledgeBaseController))
		m.Party("/knowledge-document").Handle(new(console.KnowledgeDocumentController))
		m.Party("/knowledge-retrieve").Handle(new(console.KnowledgeRetrieveController))
		m.Party("/knowledge-retrieve-log").Handle(new(console.KnowledgeRetrieveLogController))
		m.Party("/skill-definition").Handle(new(console.SkillDefinitionController))
		m.Party("/mcp").Handle(new(console.MCPController))
	})

	mvc.Configure(app.Party("/api/open/im"), func(m *mvc.Application) {
		m.Register(cfg)
		m.Party("/widget").Handle(new(open.ImWidgetController))
		m.Party("/conversation").Handle(new(open.ImConversationController))
		m.Party("/message").Handle(new(open.ImMessageController))
	})
}
