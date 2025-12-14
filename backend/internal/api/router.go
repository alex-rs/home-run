package api

import (
	"home-run-backend/internal/api/handlers"
	"home-run-backend/internal/auth"
	"home-run-backend/internal/config"
	"home-run-backend/internal/services"
	"home-run-backend/internal/services/federation"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(cfg *config.Config, manager *services.Manager, aggregator *federation.Aggregator) *gin.Engine {
	r := gin.Default()

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	if cfg.Server.CORSAllowOrigin == "*" {
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = []string{cfg.Server.CORSAllowOrigin}
	}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept"}
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))

	// Session middleware
	store := cookie.NewStore([]byte(cfg.Server.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: 2,     // SameSiteLaxMode
	})
	r.Use(sessions.Sessions("homerun_session", store))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg)
	servicesHandler := handlers.NewServicesHandler(manager, aggregator)
	hostHandler := handlers.NewHostHandler()
	federationHandler := handlers.NewFederationHandler(aggregator)

	// Health check (public)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Public routes
		api.POST("/auth/login", authHandler.Login)

		// Protected routes (session-based)
		protected := api.Group("")
		protected.Use(auth.SessionRequired())
		{
			// Auth
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/auth/check", authHandler.Check)

			// Services
			protected.GET("/services", servicesHandler.List)
			protected.GET("/services/:id", servicesHandler.Get)
			protected.GET("/services/:id/configs/:index", servicesHandler.GetConfig)

			// Host stats
			protected.GET("/host/stats", hostHandler.Stats)
		}

		// Federation endpoint (token-based)
		federationGroup := api.Group("/federation")
		federationGroup.Use(auth.TokenRequired(cfg.Auth.APIToken))
		{
			federationGroup.GET("/services", federationHandler.Services)
		}
	}

	// Serve static assets
	r.Static("/assets", "./dist/assets")

	// SPA fallback - serve index.html for any non-API route
	r.NoRoute(func(c *gin.Context) {
		c.File("./dist/index.html")
	})

	return r
}
