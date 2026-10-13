// Package user is the reference business module: it shows how a feature
// module lives in the kernel (lifecycle), consumes services (database, jwt),
// publishes services (auth) and exposes routes (route.Routable). Copy this
// pattern for your own modules.
package user

import (
	"context"

	v1 "github.com/Wi1low/go-web-template/internal/api/v1"
	"github.com/Wi1low/go-web-template/internal/auth"
	"github.com/Wi1low/go-web-template/internal/config"
	"github.com/Wi1low/go-web-template/internal/kernel"
	"github.com/Wi1low/go-web-template/internal/middleware"
	"github.com/Wi1low/go-web-template/internal/model"
	"github.com/Wi1low/go-web-template/internal/repository"
	"github.com/Wi1low/go-web-template/internal/route"
	"github.com/Wi1low/go-web-template/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type module struct {
	handler *v1.UserHandler
	parser  auth.TokenParser // nil when jwt.signing-key is empty
}

// New returns the demo user module.
func New() kernel.Module { return &module{} }

func (*module) Name() string { return "user" }

// The demo needs the database; disable it with database.enabled: false.
func (*module) Enabled(c *config.Config) bool { return c.Database.Enabled }

func (m *module) Init(ctx context.Context, a *kernel.App) error {
	db := kernel.MustRequire[*gorm.DB](a)

	repo := repository.NewUserRepository(db)
	svc := service.NewUserService(repo)

	if a.Cfg.Server.Mode == config.ModeDev {
		// Auto-migration is a dev convenience; production schemas belong in
		// migrations, not in the boot path.
		if err := db.WithContext(ctx).AutoMigrate(&model.User{}); err != nil {
			return err
		}
		seedDemoUser(ctx, db)
	}

	if a.Cfg.Jwt.SigningKey != "" {
		parser, err := auth.NewTokenParser(&a.Cfg.Jwt)
		if err != nil {
			return err
		}
		m.parser = parser
		// Publish under the interface type so other modules can depend on it.
		kernel.Provide[auth.TokenParser](a, parser)
	} else {
		zap.L().Warn("jwt.signing-key is empty; auth demo routes disabled")
	}

	m.handler = v1.NewUserHandler(svc, m.parser)
	return nil
}

func (*module) Stop(context.Context) error { return nil }

// RegisterRoutes implements route.Routable: mounted under /api/v1 by the
// httpserver module.
func (m *module) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/users", m.handler.Create)
	rg.GET("/users", m.handler.List)
	rg.GET("/users/:id", m.handler.Get)
	rg.PUT("/users/:id", m.handler.Update)
	rg.DELETE("/users/:id", m.handler.Delete)

	if m.parser != nil {
		rg.POST("/login", m.handler.Login)
		authed := rg.Group("", middleware.Auth(m.parser))
		authed.GET("/me", m.handler.Me)
	}
}

// seedDemoUser creates demo@example.com when the table is empty so the login
// demo works out of the box. Remove with the demo itself.
func seedDemoUser(ctx context.Context, db *gorm.DB) {
	var count int64
	if err := db.WithContext(ctx).Model(&model.User{}).Count(&count).Error; err != nil || count > 0 {
		return
	}
	demo := &model.User{Name: "demo", Email: "demo@example.com"}
	if err := db.WithContext(ctx).Create(demo).Error; err != nil {
		zap.L().Warn("seed demo user failed", zap.Error(err))
		return
	}
	zap.L().Info("seeded demo user demo@example.com (demo login accepts password 123456)")
}

// Compile-time assertion that the module exposes routes to the httpserver.
var _ route.Routable = (*module)(nil)
