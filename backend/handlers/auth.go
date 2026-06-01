package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"clubbix/backend/config"
	"clubbix/backend/models"
	"clubbix/backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

func memberCode(usuarioID uint) string {
	return "CLB-2026-" + strconv.FormatUint(uint64(usuarioID), 10)
}

func (h *AuthHandler) frontendSession(usuario *models.Usuario, token string) *models.FrontendSession {
	session := &models.FrontendSession{
		Nome:     usuario.NomeCompleto,
		Email:    usuario.Email,
		MemberID: memberCode(usuario.IDUsuario),
		Token:    token,
	}

	var assinatura models.Assinatura
	err := h.db.Preload("Plano").
		Where("id_usuario = ?", usuario.IDUsuario).
		Order("created_at DESC").
		First(&assinatura).Error
	if err == nil && assinatura.Plano.NomePlano != "" {
		session.Plano = "Plano " + assinatura.Plano.NomePlano
	}

	return session
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email e senha sao obrigatorios"})
		return
	}

	email := utils.NormalizeEmail(req.Email)
	if !utils.ValidateEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email invalido"})
		return
	}

	var usuario models.Usuario
	if err := h.db.Where("email = ?", email).First(&usuario).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Email ou senha incorretos"})
		return
	}

	if usuario.StatusConta != models.StatusContaAtiva {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Conta inativa ou bloqueada"})
		return
	}

	if !utils.VerifyPassword(usuario.SenhaHash, req.Senha) {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Email ou senha incorretos"})
		return
	}

	now := time.Now()
	h.db.Model(&usuario).Update("data_ultimo_acesso", now)

	token, err := utils.GenerateJWT(usuario.IDUsuario, usuario.Email, h.cfg.JWTSecret, h.cfg.JWTExpiration)
	if err != nil {
		log.Printf("Erro ao gerar JWT: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	usuario.SenhaHash = ""
	c.JSON(http.StatusOK, models.AuthResponse{
		Token:          token,
		Usuario:        &usuario,
		ExpirationTime: time.Now().Add(h.cfg.JWTExpiration),
		Session:        h.frontendSession(&usuario, token),
	})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados invalidos"})
		return
	}

	email := utils.NormalizeEmail(req.Email)
	if !utils.ValidateEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email invalido"})
		return
	}
	if !utils.ValidateCPF(req.CPF) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "CPF invalido"})
		return
	}
	if err := utils.ValidatePassword(req.Senha); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	hash, err := utils.HashPassword(req.Senha)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao proteger senha"})
		return
	}

	usuario := models.Usuario{
		NomeCompleto: utils.SanitizeText(req.NomeCompleto, 100),
		CPFHash:      utils.HashCPF(req.CPF),
		Email:        email,
		Telefone:     utils.SanitizePhone(req.Telefone),
		SenhaHash:    hash,
		TipoUsuario:  models.TipoUsuarioAssociado,
		StatusConta:  models.StatusContaAtiva,
	}

	if err := h.db.Create(&usuario).Error; err != nil {
		log.Printf("Erro ao cadastrar usuario: %v", err)
		c.JSON(http.StatusConflict, gin.H{"erro": "Nao foi possivel cadastrar usuario"})
		return
	}

	token, err := utils.GenerateJWT(usuario.IDUsuario, usuario.Email, h.cfg.JWTSecret, h.cfg.JWTExpiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	usuario.SenhaHash = ""
	c.JSON(http.StatusCreated, models.AuthResponse{
		Token:          token,
		Usuario:        &usuario,
		ExpirationTime: time.Now().Add(h.cfg.JWTExpiration),
		Session:        h.frontendSession(&usuario, token),
	})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	usuarioIDInterface, exists := c.Get("usuarioID")
	if !exists {
		usuarioIDInterface, exists = c.Get("membroID")
	}
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Nao autenticado"})
		return
	}

	usuarioID, err := strconv.ParseUint(usuarioIDInterface.(string), 10, 32)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "ID de usuario invalido"})
		return
	}

	var usuario models.Usuario
	if err := h.db.Preload("Assinaturas.Plano").First(&usuario, uint(usuarioID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Usuario nao encontrado"})
		return
	}

	usuario.SenhaHash = ""
	c.JSON(http.StatusOK, usuario)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email e obrigatorio"})
		return
	}
	if !utils.ValidateEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email invalido"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mensagem": "Se o email estiver registrado, voce recebera um link de recuperacao.",
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"mensagem": "Desconectado com sucesso"})
}

func (h *AuthHandler) PublicSubscribe(c *gin.Context) {
	var req models.PublicAssinaturaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados invalidos"})
		return
	}

	email := utils.NormalizeEmail(req.Email)
	if !utils.ValidateEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email invalido"})
		return
	}
	if !utils.ValidateCPF(req.CPF) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "CPF invalido"})
		return
	}

	var plano models.Plano
	planoKey := utils.SanitizeSlug(req.Plano)
	planoNome := map[string]string{
		"basico": "Basico",
		"plus":   "Plus",
		"max":    "Max",
	}[planoKey]
	if planoNome == "" {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Plano invalido"})
		return
	}
	if err := h.db.Where("ativo = ? AND nome_plano = ?", true, planoNome).First(&plano).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Plano nao encontrado"})
		return
	}

	var usuario models.Usuario
	var assinatura models.Assinatura
	var pagamento models.Pagamento
	err := h.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("email = ?", email).First(&usuario).Error
		if err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			password, passwordErr := utils.GenerateRandomPassword(20)
			if passwordErr != nil {
				return passwordErr
			}
			hash, hashErr := utils.HashPassword(password)
			if hashErr != nil {
				return hashErr
			}
			usuario = models.Usuario{
				NomeCompleto: utils.SanitizeText(req.NomeCompleto, 100),
				CPFHash:      utils.HashCPF(req.CPF),
				Email:        email,
				Telefone:     utils.SanitizePhone(req.Telefone),
				SenhaHash:    hash,
				TipoUsuario:  models.TipoUsuarioAssociado,
				StatusConta:  models.StatusContaAtiva,
			}
			if err := tx.Create(&usuario).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&usuario).Updates(map[string]interface{}{
				"nome_completo": utils.SanitizeText(req.NomeCompleto, 100),
				"telefone":      utils.SanitizePhone(req.Telefone),
				"status_conta":  models.StatusContaAtiva,
			}).Error; err != nil {
				return err
			}
		}

		now := time.Now()
		dataFim := now.AddDate(0, plano.DuracaoMeses, 0)
		if plano.DuracaoTipo == models.DuracaoTipoDias {
			dataFim = now.AddDate(0, 0, plano.DuracaoMeses)
		}
		assinatura = models.Assinatura{
			IDUsuario:           usuario.IDUsuario,
			IDPlano:             plano.IDPlano,
			DataInicio:          now,
			DataFim:             dataFim,
			Status:              models.StatusAssinaturaPendente,
			ValorPago:           plano.Valor,
			RenovacaoAutomatica: req.RenovacaoAutomatica,
			CodigoTransacao:     utils.GenerateCodigoDesconto("TRX"),
		}
		if err := tx.Create(&assinatura).Error; err != nil {
			return err
		}

		pagamento = models.Pagamento{
			IDAssinatura:    assinatura.IDAssinatura,
			MetodoPagamento: utils.SanitizeSlug(req.MetodoPagamento),
			Valor:           plano.Valor,
			StatusPagamento: models.StatusPagamentoPendente,
			IDGateway:       utils.GenerateCodigoDesconto("LOCAL"),
		}
		return tx.Create(&pagamento).Error
	})
	if err != nil {
		log.Printf("Erro ao registrar assinatura publica: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Nao foi possivel registrar assinatura"})
		return
	}

	token, err := utils.GenerateJWT(usuario.IDUsuario, usuario.Email, h.cfg.JWTSecret, h.cfg.JWTExpiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao gerar token"})
		return
	}

	assinatura.Plano = plano
	usuario.SenhaHash = ""
	session := h.frontendSession(&usuario, token)
	session.Plano = "Plano " + plano.NomePlano
	c.JSON(http.StatusCreated, models.PublicAssinaturaResponse{
		Mensagem:    "Assinatura recebida com sucesso",
		Token:       token,
		Session:     session,
		Usuario:     &usuario,
		Assinatura:  &assinatura,
		Pagamento:   &pagamento,
		PaymentNote: "Pagamento registrado como pendente ate integracao com gateway.",
	})
}
