package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"clubbix/backend/config"
	"clubbix/backend/models"
	"clubbix/backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RateLimiter implementa rate limiting por IP (proteção contra A04 - DoS)
type RateLimiter struct {
	maxRequests int
	window      time.Duration
	visitors    map[string]*Visitor
	mu          sync.RWMutex
}

type Visitor struct {
	lastSeen time.Time
	requests int
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
		visitors:    make(map[string]*Visitor),
	}

	// Limpar visitors antigos a cada 10 minutos
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.mu.Lock()
			now := time.Now()
			for ip, visitor := range rl.visitors {
				if now.Sub(visitor.lastSeen) > window*2 {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()

	return rl
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		now := time.Now()

		visitor, exists := rl.visitors[ip]
		if !exists {
			rl.visitors[ip] = &Visitor{
				lastSeen: now,
				requests: 1,
			}
		} else {
			if now.Sub(visitor.lastSeen) > rl.window {
				visitor.requests = 1
				visitor.lastSeen = now
			} else if visitor.requests >= rl.maxRequests {
				rl.mu.Unlock()
				c.AbortWithStatusJSON(429, gin.H{
					"erro": "Muitas requisições. Tente novamente mais tarde.",
				})
				return
			} else {
				visitor.requests++
			}
		}
		rl.mu.Unlock()

		c.Next()
	}
}

// SecurityHeaders adiciona headers de segurança (proteção contra A05 - Security Misconfiguration)
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevenir clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevenir MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Ativar XSS Protection (para navegadores antigos)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' https://wa.me https://api.whatsapp.com;")

		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Feature Policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")
		if len(c.Request.URL.Path) >= 5 && c.Request.URL.Path[:5] == "/api/" {
			c.Header("Cache-Control", "no-store")
		}

		// HSTS (apenas em produção)
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

func RecoveryMiddleware(environment string) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if environment == "development" {
					log.Printf("panic: %v\n%s", recovered, debug.Stack())
				} else {
					log.Printf("panic: %v", recovered)
				}
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"erro": "Erro interno do servidor",
				})
			}
		}()
		c.Next()
	}
}

// CORS configura CORS de forma segura (proteção contra A05 - Security Misconfiguration)
func CORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Verificar se a origin é permitida
		isAllowed := false
		for _, allowed := range cfg.AllowedOrigins {
			if origin == allowed {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
			c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Requested-With")
			c.Header("Access-Control-Max-Age", "3600")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// AuthMiddleware valida o token JWT (proteção contra A07 - Authentication Failures)
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"erro": "Token não fornecido",
			})
			return
		}

		// Extrair token do header "Bearer TOKEN"
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.AbortWithStatusJSON(401, gin.H{
				"erro": "Formato de token inválido",
			})
			return
		}

		tokenString := authHeader[7:]

		claims, err := utils.ParseJWT(tokenString, jwtSecret)
		if err != nil {
			log.Printf("Erro ao fazer parse do JWT: %v", err)
			c.AbortWithStatusJSON(401, gin.H{
				"erro": "Token inválido ou expirado",
			})
			return
		}

		// Armazenar claims no contexto para usar nos handlers
		c.Set("usuarioID", claims["sub"])
		c.Set("membroID", claims["sub"])
		c.Set("email", claims["email"])

		c.Next()
	}
}

func AdminOnly(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawID, exists := c.Get("usuarioID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "Nao autenticado"})
			return
		}

		usuarioID, err := strconv.ParseUint(fmt.Sprintf("%v", rawID), 10, 32)
		if err != nil || usuarioID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "Usuario invalido"})
			return
		}

		var usuario models.Usuario
		if err := db.Select("id_usuario", "tipo_usuario", "status_conta").First(&usuario, uint(usuarioID)).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "Usuario invalido"})
			return
		}
		if usuario.StatusConta != models.StatusContaAtiva || usuario.TipoUsuario != models.TipoUsuarioAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"erro": "Acesso restrito a administradores"})
			return
		}

		c.Next()
	}
}

// Logging middleware com proteção de dados sensíveis (proteção contra A09 - Logging & Monitoring)
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Capturar detalhes da requisição
		method := c.Request.Method
		path := c.Request.RequestURI
		statusCode := 0

		// Usar um custom writer para capturar o status code
		c.Next()

		statusCode = c.Writer.Status()
		duration := time.Since(startTime).Milliseconds()

		// Log seguro (não logar dados sensíveis como senhas)
		logMessage := fmt.Sprintf(
			"[%s] %s %s - Status: %d - Duração: %dms - IP: %s",
			time.Now().Format("2006-01-02 15:04:05"),
			method,
			path,
			statusCode,
			duration,
			c.ClientIP(),
		)

		log.Println(logMessage)
	}
}

// InputValidationMiddleware valida o tamanho das requisições (proteção contra A04 - Insecure Design)
func InputValidationMiddleware(maxBodySize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Limitar tamanho do corpo da requisição
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodySize)

		c.Next()
	}
}

// GetClientIP retorna o IP real do cliente (considerando proxies)
func GetClientIP(c *gin.Context) string {
	ip := c.GetHeader("X-Forwarded-For")
	if ip == "" {
		ip = c.GetHeader("X-Real-IP")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(c.Request.RemoteAddr)
	}
	return ip
}

// TrustProxyMiddleware configura quais IPs de proxy devem ser confiados
func TrustProxyMiddleware(trustedProxies []string) gin.HandlerFunc {
	if len(trustedProxies) > 0 {
		// Gin pode usar SetTrustedProxies, mas vamos usar isso no main.go
	}
	return func(c *gin.Context) {
		c.Next()
	}
}
