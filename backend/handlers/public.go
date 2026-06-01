package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"clubbix/backend/models"
	"clubbix/backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PublicHandler struct {
	db *gorm.DB
}

type publicParceiroResponse struct {
	IDParceiro         uint    `json:"id_parceiro"`
	NomeEmpresa        string  `json:"nome_empresa"`
	Categoria          string  `json:"categoria"`
	CategoriaNome      string  `json:"categoria_nome"`
	CategoriaIcone     string  `json:"categoria_icone"`
	Descricao          string  `json:"descricao"`
	PercentualDesconto float64 `json:"percentual_desconto"`
	Endereco           string  `json:"endereco"`
	Telefone           string  `json:"telefone"`
	Email              string  `json:"email"`
	Site               string  `json:"site"`
	LogoURL            string  `json:"logo_url"`
	Codigo             string  `json:"codigo"`
}

func NewPublicHandler(db *gorm.DB) *PublicHandler {
	return &PublicHandler{db: db}
}

func (h *PublicHandler) ListPlanos(c *gin.Context) {
	var planos []models.Plano
	if err := h.db.Where("ativo = ?", true).Find(&planos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar planos"})
		return
	}
	c.JSON(http.StatusOK, planos)
}

func (h *PublicHandler) GetPlano(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "ID de plano invalido"})
		return
	}

	var plano models.Plano
	if err := h.db.Where("ativo = ?", true).First(&plano, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Plano nao encontrado"})
		return
	}
	c.JSON(http.StatusOK, plano)
}

func (h *PublicHandler) ListParceiros(c *gin.Context) {
	var parceiros []models.Parceiro
	query := h.db.Where("status = ?", models.StatusParceiroAtivo)

	if categoria := c.Query("categoria"); categoria != "" {
		query = query.Where("categoria = ?", utils.SanitizeSlug(categoria))
	}
	if busca := c.Query("busca"); busca != "" {
		busca = "%" + utils.SanitizeText(busca, 80) + "%"
		query = query.Where("nome_empresa LIKE ? OR descricao LIKE ? OR categoria LIKE ?", busca, busca, busca)
	}

	if err := query.Limit(100).Find(&parceiros).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar parceiros"})
		return
	}

	var categorias []models.Categoria
	if err := h.db.Where("ativo = ?", true).Find(&categorias).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar categorias"})
		return
	}
	bySlug := map[string]models.Categoria{}
	for _, categoria := range categorias {
		bySlug[categoria.Slug] = categoria
	}

	response := make([]publicParceiroResponse, 0, len(parceiros))
	for _, parceiro := range parceiros {
		categoria := bySlug[parceiro.Categoria]
		response = append(response, publicParceiroResponse{
			IDParceiro:         parceiro.IDParceiro,
			NomeEmpresa:        parceiro.NomeEmpresa,
			Categoria:          parceiro.Categoria,
			CategoriaNome:      categoria.Nome,
			CategoriaIcone:     categoria.Icone,
			Descricao:          parceiro.Descricao,
			PercentualDesconto: parceiro.PercentualDesconto,
			Endereco:           parceiro.Endereco,
			Telefone:           parceiro.Telefone,
			Email:              parceiro.Email,
			Site:               parceiro.Site,
			LogoURL:            parceiro.LogoURL,
			Codigo:             "CLB-" + strconv.FormatUint(uint64(100+parceiro.IDParceiro), 10),
		})
	}

	c.JSON(http.StatusOK, gin.H{"parceiros": response, "total": len(response)})
}

func (h *PublicHandler) ListCategorias(c *gin.Context) {
	type categoriaResponse struct {
		Slug  string `json:"slug"`
		Nome  string `json:"nome"`
		Icone string `json:"icone"`
		Total int64  `json:"total"`
	}

	var categorias []models.Categoria
	if err := h.db.Where("ativo = ?", true).Order("ordem ASC, nome ASC").Find(&categorias).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar categorias"})
		return
	}

	var counts []struct {
		Categoria string
		Total     int64
	}
	if err := h.db.Model(&models.Parceiro{}).
		Select("categoria, count(*) as total").
		Where("status = ?", models.StatusParceiroAtivo).
		Group("categoria").
		Scan(&counts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao contar categorias"})
		return
	}
	countBySlug := map[string]int64{}
	for _, count := range counts {
		countBySlug[count.Categoria] = count.Total
	}

	response := make([]categoriaResponse, 0, len(categorias))
	for _, categoria := range categorias {
		response = append(response, categoriaResponse{
			Slug:  categoria.Slug,
			Nome:  categoria.Nome,
			Icone: categoria.Icone,
			Total: countBySlug[categoria.Slug],
		})
	}

	c.JSON(http.StatusOK, gin.H{"categorias": response, "total": len(response)})
}

func (h *PublicHandler) GetParceiro(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "ID de parceiro invalido"})
		return
	}

	var parceiro models.Parceiro
	if err := h.db.Preload("Beneficios", "ativo = ?", true).Preload("Cursos", "ativo = ?", true).
		Where("status = ?", models.StatusParceiroAtivo).
		First(&parceiro, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Parceiro nao encontrado"})
		return
	}
	c.JSON(http.StatusOK, parceiro)
}

func (h *PublicHandler) ListBeneficios(c *gin.Context) {
	var beneficios []models.Beneficio
	if err := h.db.Preload("Parceiro").Where("ativo = ?", true).Limit(100).Find(&beneficios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar beneficios"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"beneficios": beneficios, "total": len(beneficios)})
}

func (h *PublicHandler) ListCursos(c *gin.Context) {
	var cursos []models.Curso
	if err := h.db.Preload("Parceiro").Where("ativo = ?", true).Limit(100).Find(&cursos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar cursos"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cursos": cursos, "total": len(cursos)})
}

func (h *PublicHandler) Contact(c *gin.Context) {
	var req models.ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados invalidos"})
		return
	}
	email := utils.NormalizeEmail(req.Email)
	if email != "" && !utils.ValidateEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email invalido"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mensagem": "Mensagem recebida com sucesso. Nossa equipe retornara em breve.",
		"contato": gin.H{
			"nome":     utils.SanitizeText(req.Nome, 100),
			"assunto":  utils.SanitizeText(req.Assunto, 120),
			"email":    email,
			"telefone": utils.SanitizePhone(req.Telefone),
		},
	})
}

func (h *PublicHandler) FAQ(c *gin.Context) {
	type faqItem struct {
		ID       string   `json:"id"`
		Pergunta string   `json:"pergunta"`
		Resposta string   `json:"resposta"`
		Tags     []string `json:"tags"`
	}

	items := []faqItem{
		{ID: "oque", Pergunta: "O que e o Clubbix?", Resposta: "Clube de descontos com acesso a parceiros selecionados e beneficios exclusivos para membros.", Tags: []string{"clube", "beneficios", "descontos"}},
		{ID: "planos", Pergunta: "Quais sao os planos?", Resposta: "Basico, Plus e Max, com valores e duracoes exibidos na pagina de assinatura.", Tags: []string{"planos", "preco", "assinar"}},
		{ID: "acesso", Pergunta: "Como uso os beneficios?", Resposta: "Apos a assinatura, o membro acessa sua area do clube e apresenta o codigo ou cartao digital ao parceiro.", Tags: []string{"cartao", "parceiro", "codigo"}},
		{ID: "univag", Pergunta: "Existe beneficio para graduacao?", Resposta: "Cursos participantes podem ter beneficios adicionais conforme as regras comerciais vigentes.", Tags: []string{"univag", "curso", "graduacao"}},
		{ID: "indicacao", Pergunta: "Como funciona a indicacao?", Resposta: "Membros podem gerar um link ou codigo de indicacao para compartilhar com amigos.", Tags: []string{"indicacao", "convite", "link"}},
	}

	q := strings.ToLower(utils.SanitizeText(c.Query("q"), 80))
	if q == "" {
		c.JSON(http.StatusOK, gin.H{"faq": items, "total": len(items)})
		return
	}

	filtered := make([]faqItem, 0)
	for _, item := range items {
		text := strings.ToLower(item.Pergunta + " " + item.Resposta + " " + strings.Join(item.Tags, " "))
		if strings.Contains(text, q) {
			filtered = append(filtered, item)
		}
	}
	c.JSON(http.StatusOK, gin.H{"faq": filtered, "total": len(filtered)})
}

type AssinaturaHandler struct {
	db *gorm.DB
}

func NewAssinaturaHandler(db *gorm.DB) *AssinaturaHandler {
	return &AssinaturaHandler{db: db}
}

func (h *AssinaturaHandler) CreateAssinatura(c *gin.Context) {
	rawID, exists := c.Get("usuarioID")
	if !exists {
		rawID, exists = c.Get("membroID")
	}
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Nao autenticado"})
		return
	}

	usuarioID, err := strconv.ParseUint(rawID.(string), 10, 32)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "ID de usuario invalido"})
		return
	}

	var req models.AssinaturaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados invalidos"})
		return
	}

	var plano models.Plano
	if err := h.db.Where("ativo = ?", true).First(&plano, req.IDPlano).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Plano nao encontrado"})
		return
	}

	now := time.Now()
	dataFim := now.AddDate(0, plano.DuracaoMeses, 0)
	if plano.DuracaoTipo == models.DuracaoTipoDias {
		dataFim = now.AddDate(0, 0, plano.DuracaoMeses)
	}

	assinatura := models.Assinatura{
		IDUsuario:           uint(usuarioID),
		IDPlano:             plano.IDPlano,
		DataInicio:          now,
		DataFim:             dataFim,
		Status:              models.StatusAssinaturaPendente,
		ValorPago:           plano.Valor,
		RenovacaoAutomatica: req.RenovacaoAutomatica,
		CodigoTransacao:     utils.GenerateCodigoDesconto("TRX"),
	}

	if err := h.db.Create(&assinatura).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao criar assinatura"})
		return
	}
	c.JSON(http.StatusCreated, assinatura)
}

func (h *AssinaturaHandler) GetAssinaturaMe(c *gin.Context) {
	rawID, exists := c.Get("usuarioID")
	if !exists {
		rawID, exists = c.Get("membroID")
	}
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Nao autenticado"})
		return
	}

	usuarioID, err := strconv.ParseUint(rawID.(string), 10, 32)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "ID de usuario invalido"})
		return
	}

	var assinatura models.Assinatura
	if err := h.db.Where("id_usuario = ?", uint(usuarioID)).Preload("Plano").Order("created_at DESC").First(&assinatura).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Assinatura nao encontrada"})
		return
	}
	c.JSON(http.StatusOK, assinatura)
}

func (h *AssinaturaHandler) WebhookPagamento(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"mensagem": "Webhook recebido"})
}
