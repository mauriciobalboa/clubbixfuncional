package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"clubbix/backend/config"
	"clubbix/backend/database"
	"clubbix/backend/handlers"
	"clubbix/backend/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	db := database.InitDB(cfg.DBPath, cfg.Environment)

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.RecoveryMiddleware(cfg.Environment))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.CORS(cfg))
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.NewRateLimiter(cfg.RateLimitPerMin, time.Minute).Limit())
	router.Use(middleware.InputValidationMiddleware(cfg.MaxRequestSize))
	router.Use(redirectHTMLFileRoutes())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "OK",
			"time":    time.Now(),
			"version": "1.0.0",
		})
	})

	publicGroup := router.Group("/api")
	{
		authHandler := handlers.NewAuthHandler(db, cfg)
		publicGroup.POST("/auth/login", authHandler.Login)
		publicGroup.POST("/auth/register", authHandler.Register)
		publicGroup.POST("/auth/esqueci-senha", authHandler.ForgotPassword)
		publicGroup.POST("/assinaturas/publica", authHandler.PublicSubscribe)
		publicGroup.POST("/checkout/assinatura", authHandler.PublicSubscribe)

		publicHandler := handlers.NewPublicHandler(db)
		publicGroup.GET("/planos", publicHandler.ListPlanos)
		publicGroup.GET("/planos/:id", publicHandler.GetPlano)
		publicGroup.GET("/parceiros", publicHandler.ListParceiros)
		publicGroup.GET("/parceiros/:id", publicHandler.GetParceiro)
		publicGroup.GET("/categorias", publicHandler.ListCategorias)
		publicGroup.GET("/beneficios", publicHandler.ListBeneficios)
		publicGroup.GET("/cursos", publicHandler.ListCursos)
		publicGroup.POST("/contato", publicHandler.Contact)
		publicGroup.GET("/faq", publicHandler.FAQ)
	}

	protectedGroup := router.Group("/api")
	protectedGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		authHandler := handlers.NewAuthHandler(db, cfg)
		protectedGroup.GET("/auth/me", authHandler.GetMe)
		protectedGroup.POST("/auth/logout", authHandler.Logout)

		assinaturaHandler := handlers.NewAssinaturaHandler(db)
		paymentHandler := handlers.NewPaymentHandler(db, cfg)
		protectedGroup.POST("/assinatura", assinaturaHandler.CreateAssinatura)
		protectedGroup.GET("/assinatura/me", assinaturaHandler.GetAssinaturaMe)
		protectedGroup.POST("/pagamento", paymentHandler.CreatePagamento)
	}

	adminGroup := router.Group("/api/admin")
	adminGroup.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	adminGroup.Use(middleware.AdminOnly(db))
	{
		adminHandler := handlers.NewAdminHandler(db)

		adminGroup.GET("/usuarios", adminHandler.ListUsuarios)
		adminGroup.GET("/usuarios/:id", adminHandler.GetUsuario)
		adminGroup.POST("/usuarios", adminHandler.CreateUsuario)
		adminGroup.PUT("/usuarios/:id", adminHandler.UpdateUsuario)
		adminGroup.DELETE("/usuarios/:id", adminHandler.DeleteUsuario)

		adminGroup.GET("/planos", adminHandler.ListPlanos)
		adminGroup.GET("/planos/:id", adminHandler.GetPlano)
		adminGroup.POST("/planos", adminHandler.CreatePlano)
		adminGroup.PUT("/planos/:id", adminHandler.UpdatePlano)
		adminGroup.DELETE("/planos/:id", adminHandler.DeletePlano)

		adminGroup.GET("/parceiros", adminHandler.ListParceiros)
		adminGroup.GET("/parceiros/:id", adminHandler.GetParceiro)
		adminGroup.POST("/parceiros", adminHandler.CreateParceiro)
		adminGroup.PUT("/parceiros/:id", adminHandler.UpdateParceiro)
		adminGroup.DELETE("/parceiros/:id", adminHandler.DeleteParceiro)

		adminGroup.GET("/assinaturas", adminHandler.ListAssinaturas)
		adminGroup.GET("/assinaturas/:id", adminHandler.GetAssinatura)
		adminGroup.POST("/assinaturas", adminHandler.CreateAssinatura)
		adminGroup.PUT("/assinaturas/:id", adminHandler.UpdateAssinatura)
		adminGroup.DELETE("/assinaturas/:id", adminHandler.DeleteAssinatura)

		adminGroup.GET("/pagamentos", adminHandler.ListPagamentos)
		adminGroup.GET("/pagamentos/:id", adminHandler.GetPagamento)
		adminGroup.POST("/pagamentos", adminHandler.CreatePagamento)
		adminGroup.PUT("/pagamentos/:id", adminHandler.UpdatePagamento)
		adminGroup.DELETE("/pagamentos/:id", adminHandler.DeletePagamento)

		adminGroup.GET("/beneficios", adminHandler.ListBeneficios)
		adminGroup.GET("/beneficios/:id", adminHandler.GetBeneficio)
		adminGroup.POST("/beneficios", adminHandler.CreateBeneficio)
		adminGroup.PUT("/beneficios/:id", adminHandler.UpdateBeneficio)
		adminGroup.DELETE("/beneficios/:id", adminHandler.DeleteBeneficio)

		adminGroup.GET("/cursos", adminHandler.ListCursos)
		adminGroup.GET("/cursos/:id", adminHandler.GetCurso)
		adminGroup.POST("/cursos", adminHandler.CreateCurso)
		adminGroup.PUT("/cursos/:id", adminHandler.UpdateCurso)
		adminGroup.DELETE("/cursos/:id", adminHandler.DeleteCurso)

		adminGroup.GET("/responsaveis", adminHandler.ListResponsaveis)
		adminGroup.GET("/responsaveis/:id", adminHandler.GetResponsavel)
		adminGroup.POST("/responsaveis", adminHandler.CreateResponsavel)
		adminGroup.PUT("/responsaveis/:id", adminHandler.UpdateResponsavel)
		adminGroup.DELETE("/responsaveis/:id", adminHandler.DeleteResponsavel)

		adminGroup.GET("/usuario-cursos", adminHandler.ListUsuarioCursos)
		adminGroup.GET("/usuario-cursos/:id", adminHandler.GetUsuarioCurso)
		adminGroup.POST("/usuario-cursos", adminHandler.CreateUsuarioCurso)
		adminGroup.PUT("/usuario-cursos/:id", adminHandler.UpdateUsuarioCurso)
		adminGroup.DELETE("/usuario-cursos/:id", adminHandler.DeleteUsuarioCurso)

		adminGroup.GET("/logs-auditoria", adminHandler.ListLogsAuditoria)
		adminGroup.GET("/logs-auditoria/:id", adminHandler.GetLogAuditoria)
		adminGroup.POST("/logs-auditoria", adminHandler.CreateLogAuditoria)
		adminGroup.PUT("/logs-auditoria/:id", adminHandler.UpdateLogAuditoria)
		adminGroup.DELETE("/logs-auditoria/:id", adminHandler.DeleteLogAuditoria)
	}

	webhookGroup := router.Group("/webhook")
	{
		assinaturaHandler := handlers.NewAssinaturaHandler(db)
		webhookGroup.POST("/pagamento", assinaturaHandler.WebhookPagamento)
	}

	if frontendPath := resolveFrontendPath(); frontendPath != "" {
		router.Static("/assets", filepath.Join(frontendPath, "assets"))
		router.GET("/", serveHTML(frontendPath, "index.html"))
		router.GET("/login", serveHTML(frontendPath, "login.html"))
		router.GET("/clube", serveHTML(frontendPath, "clube.html"))
		router.GET("/membro", serveHTML(frontendPath, "clube.html"))
		router.GET("/assinar", serveHTML(frontendPath, "assinar.html"))
		router.GET("/index.html", redirectRoute("/"))
		router.GET("/login.html", redirectRoute("/login"))
		router.GET("/clube.html", redirectRoute("/clube"))
		router.GET("/assinar.html", redirectRoute("/assinar"))
		log.Printf("Frontend HTML servido de: %s", frontendPath)
	} else {
		log.Printf("Frontend nao encontrado; servidor expondo apenas API")
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"erro":   "Endpoint nao encontrado",
			"path":   c.Request.RequestURI,
			"metodo": c.Request.Method,
		})
	})

	serverAddress := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Servidor Clubbix iniciando em http://localhost:%s", cfg.Port)
	log.Printf("Banco de dados: %s", cfg.DBPath)
	log.Printf("Ambiente: %s", cfg.Environment)

	server := &http.Server{
		Addr:         serverAddress,
		Handler:      router,
		ReadTimeout:  cfg.RequestTimeout,
		WriteTimeout: cfg.RequestTimeout,
		IdleTimeout:  15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func resolveFrontendPath() string {
	candidates := []string{
		"../frontend",
		"./frontend",
	}
	for _, candidate := range candidates {
		abs, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if info, err := os.Stat(filepath.Join(abs, "index.html")); err == nil && !info.IsDir() {
			return abs
		}
	}
	return ""
}

func serveHTML(frontendPath, filename string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.File(filepath.Join(frontendPath, filename))
	}
}

func redirectRoute(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		target := path
		if c.Request.URL.RawQuery != "" {
			target += "?" + c.Request.URL.RawQuery
		}
		c.Redirect(http.StatusPermanentRedirect, target)
	}
}

func redirectHTMLFileRoutes() gin.HandlerFunc {
	return func(c *gin.Context) {
		cleanRoutes := map[string]string{
			"/index.html":   "/",
			"/login.html":   "/login",
			"/clube.html":   "/clube",
			"/assinar.html": "/assinar",
		}

		path := strings.ToLower(c.Request.URL.Path)
		if cleanPath, ok := cleanRoutes[path]; ok {
			if c.Request.URL.RawQuery != "" {
				cleanPath += "?" + c.Request.URL.RawQuery
			}
			c.Redirect(http.StatusPermanentRedirect, cleanPath)
			c.Abort()
			return
		}

		if strings.HasSuffix(path, ".html") {
			c.JSON(http.StatusNotFound, gin.H{"erro": "Endpoint nao encontrado"})
			c.Abort()
			return
		}
		c.Next()
	}
}
