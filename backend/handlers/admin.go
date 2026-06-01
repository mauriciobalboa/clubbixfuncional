package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"clubbix/backend/models"
	"clubbix/backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "ID invalido"})
		return 0, false
	}
	return uint(id), true
}

func bindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados invalidos"})
		return false
	}
	return true
}

func clean(value string) string {
	return utils.SanitizeString(strings.TrimSpace(value))
}

func cleanText(value string, maxLength int) string {
	return utils.SanitizeText(value, maxLength)
}

func cleanEmail(value string) string {
	return utils.NormalizeEmail(value)
}

func cleanPhone(value string) string {
	return utils.SanitizePhone(value)
}

func cleanSlug(value string) string {
	return utils.SanitizeSlug(value)
}

func cleanURL(value string) string {
	return utils.SanitizeURL(value)
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func (h *AdminHandler) exists(model interface{}, id uint) bool {
	return h.db.First(model, id).Error == nil
}

func (h *AdminHandler) categoryExists(slug string) bool {
	var count int64
	h.db.Model(&models.Categoria{}).Where("slug = ? AND ativo = ?", slug, true).Count(&count)
	return count > 0
}

func (h *AdminHandler) list(c *gin.Context, out interface{}, preload ...string) bool {
	limit := 100
	offset := 0
	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "limit deve estar entre 1 e 200"})
			return false
		}
		limit = parsed
	}
	if raw := c.Query("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "offset deve ser maior ou igual a 0"})
			return false
		}
		offset = parsed
	}
	query := h.db.Limit(limit).Offset(offset)
	for _, item := range preload {
		query = query.Preload(item)
	}
	if err := query.Find(out).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar registros"})
		return false
	}
	return true
}

func notFound(c *gin.Context, name string, err error) bool {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"erro": fmt.Sprintf("%s nao encontrado", name)})
		return true
	}
	return false
}

func (h *AdminHandler) saveAudit(c *gin.Context, usuarioID uint, acao, descricao string) {
	if usuarioID == 0 {
		return
	}
	h.db.Create(&models.LogAuditoria{
		IDUsuario: usuarioID,
		Acao:      acao,
		Descricao: descricao,
		IPUsuario: c.ClientIP(),
	})
}

// Usuarios
func (h *AdminHandler) ListUsuarios(c *gin.Context) {
	var usuarios []models.Usuario
	if !h.list(c, &usuarios) {
		return
	}
	for i := range usuarios {
		usuarios[i].SenhaHash = ""
	}
	c.JSON(http.StatusOK, gin.H{"usuarios": usuarios, "total": len(usuarios)})
}

func (h *AdminHandler) GetUsuario(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var usuario models.Usuario
	if err := h.db.Preload("Assinaturas.Plano").Preload("Responsaveis").Preload("UsuarioCursos.Curso").First(&usuario, id).Error; err != nil {
		if !notFound(c, "Usuario", err) {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar usuario"})
		}
		return
	}
	usuario.SenhaHash = ""
	c.JSON(http.StatusOK, usuario)
}

func (h *AdminHandler) CreateUsuario(c *gin.Context) {
	var req models.UsuarioCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	if !utils.ValidateCPF(req.CPF) || !utils.ValidateEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "CPF ou email invalido"})
		return
	}
	email := cleanEmail(req.Email)
	if err := utils.ValidatePassword(req.Senha); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}
	hash, err := utils.HashPassword(req.Senha)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao proteger senha"})
		return
	}
	tipo := req.TipoUsuario
	if tipo == "" {
		tipo = models.TipoUsuarioAssociado
	}
	status := req.StatusConta
	if status == "" {
		status = models.StatusContaAtiva
	}
	usuario := models.Usuario{
		NomeCompleto: cleanText(req.NomeCompleto, 100),
		CPFHash:      utils.HashCPF(req.CPF),
		Email:        email,
		Telefone:     cleanPhone(req.Telefone),
		SenhaHash:    hash,
		TipoUsuario:  tipo,
		StatusConta:  status,
	}
	if err := h.db.Create(&usuario).Error; err != nil {
		log.Printf("Erro ao criar usuario: %v", err)
		c.JSON(http.StatusConflict, gin.H{"erro": "Nao foi possivel criar usuario"})
		return
	}
	usuario.SenhaHash = ""
	c.JSON(http.StatusCreated, usuario)
}

func (h *AdminHandler) UpdateUsuario(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req models.UsuarioUpdateRequest
	if !bindJSON(c, &req) {
		return
	}
	var usuario models.Usuario
	if err := h.db.First(&usuario, id).Error; err != nil {
		if !notFound(c, "Usuario", err) {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar usuario"})
		}
		return
	}
	updates := map[string]interface{}{}
	if req.NomeCompleto != nil {
		updates["nome_completo"] = cleanText(*req.NomeCompleto, 100)
	}
	if req.CPF != nil {
		if !utils.ValidateCPF(*req.CPF) {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "CPF invalido"})
			return
		}
		updates["cpf_hash"] = utils.HashCPF(*req.CPF)
	}
	if req.Email != nil {
		if !utils.ValidateEmail(*req.Email) {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "Email invalido"})
			return
		}
		updates["email"] = cleanEmail(*req.Email)
	}
	if req.Telefone != nil {
		updates["telefone"] = cleanPhone(*req.Telefone)
	}
	if req.Senha != nil {
		if err := utils.ValidatePassword(*req.Senha); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
			return
		}
		hash, err := utils.HashPassword(*req.Senha)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao proteger senha"})
			return
		}
		updates["senha_hash"] = hash
	}
	if req.TipoUsuario != nil {
		updates["tipo_usuario"] = *req.TipoUsuario
	}
	if req.StatusConta != nil {
		updates["status_conta"] = *req.StatusConta
	}
	if len(updates) > 0 {
		if err := h.db.Model(&usuario).Updates(updates).Error; err != nil {
			log.Printf("Erro ao atualizar usuario: %v", err)
			c.JSON(http.StatusConflict, gin.H{"erro": "Nao foi possivel atualizar usuario"})
			return
		}
	}
	h.db.First(&usuario, id)
	usuario.SenhaHash = ""
	c.JSON(http.StatusOK, usuario)
}

func (h *AdminHandler) DeleteUsuario(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	result := h.db.Model(&models.Usuario{}).Where("id_usuario = ?", id).Update("status_conta", models.StatusContaInativa)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao inativar usuario"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Usuario nao encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mensagem": "Usuario inativado com sucesso"})
}

// Planos
func (h *AdminHandler) ListPlanos(c *gin.Context) {
	var items []models.Plano
	if h.list(c, &items) {
		c.JSON(http.StatusOK, gin.H{"planos": items, "total": len(items)})
	}
}
func (h *AdminHandler) GetPlano(c *gin.Context) { h.getByID(c, "Plano", &models.Plano{}) }
func (h *AdminHandler) CreatePlano(c *gin.Context) {
	var req models.PlanoRequest
	if !bindJSON(c, &req) {
		return
	}
	item := models.Plano{NomePlano: clean(req.NomePlano), Descricao: clean(req.Descricao), Valor: req.Valor, DuracaoMeses: req.DuracaoMeses, DuracaoTipo: req.DuracaoTipo, Ativo: boolValue(req.Ativo, true)}
	h.create(c, &item)
}
func (h *AdminHandler) UpdatePlano(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req models.PlanoRequest
	if !bindJSON(c, &req) {
		return
	}
	var item models.Plano
	if err := h.db.First(&item, id).Error; err != nil {
		if !notFound(c, "Plano", err) {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar plano"})
		}
		return
	}
	item.NomePlano, item.Descricao, item.Valor, item.DuracaoMeses, item.DuracaoTipo, item.Ativo = clean(req.NomePlano), clean(req.Descricao), req.Valor, req.DuracaoMeses, req.DuracaoTipo, boolValue(req.Ativo, item.Ativo)
	h.save(c, &item)
}
func (h *AdminHandler) DeletePlano(c *gin.Context) {
	h.softUpdate(c, &models.Plano{}, "id_plano", "ativo", false, "Plano desativado com sucesso")
}

// Parceiros
func (h *AdminHandler) ListParceiros(c *gin.Context) {
	var items []models.Parceiro
	if h.list(c, &items) {
		c.JSON(http.StatusOK, gin.H{"parceiros": items, "total": len(items)})
	}
}
func (h *AdminHandler) GetParceiro(c *gin.Context) {
	h.getByID(c, "Parceiro", &models.Parceiro{}, "Beneficios", "Cursos")
}
func (h *AdminHandler) CreateParceiro(c *gin.Context) {
	var req models.ParceiroRequest
	if !bindJSON(c, &req) {
		return
	}
	categoria := cleanSlug(req.Categoria)
	if !h.categoryExists(categoria) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Categoria nao existe"})
		return
	}
	email := cleanEmail(req.Email)
	if email != "" && !utils.ValidateEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email invalido"})
		return
	}
	status := req.Status
	if status == "" {
		status = models.StatusParceiroAtivo
	}
	item := models.Parceiro{NomeEmpresa: cleanText(req.NomeEmpresa, 150), Categoria: categoria, Descricao: cleanText(req.Descricao, 3000), PercentualDesconto: req.PercentualDesconto, Endereco: cleanText(req.Endereco, 1000), Telefone: cleanPhone(req.Telefone), Email: email, Site: cleanURL(req.Site), LogoURL: cleanURL(req.LogoURL), Status: status}
	h.create(c, &item)
}
func (h *AdminHandler) UpdateParceiro(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req models.ParceiroRequest
	if !bindJSON(c, &req) {
		return
	}
	var item models.Parceiro
	if err := h.db.First(&item, id).Error; err != nil {
		if !notFound(c, "Parceiro", err) {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar parceiro"})
		}
		return
	}
	categoria := cleanSlug(req.Categoria)
	if !h.categoryExists(categoria) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Categoria nao existe"})
		return
	}
	email := cleanEmail(req.Email)
	if email != "" && !utils.ValidateEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Email invalido"})
		return
	}
	status := req.Status
	if status == "" {
		status = item.Status
	}
	item.NomeEmpresa, item.Categoria, item.Descricao, item.PercentualDesconto = cleanText(req.NomeEmpresa, 150), categoria, cleanText(req.Descricao, 3000), req.PercentualDesconto
	item.Endereco, item.Telefone, item.Email, item.Site, item.LogoURL, item.Status = cleanText(req.Endereco, 1000), cleanPhone(req.Telefone), email, cleanURL(req.Site), cleanURL(req.LogoURL), status
	h.save(c, &item)
}
func (h *AdminHandler) DeleteParceiro(c *gin.Context) {
	h.softUpdate(c, &models.Parceiro{}, "id_parceiro", "status", models.StatusParceiroInativo, "Parceiro inativado com sucesso")
}

// Assinaturas
func (h *AdminHandler) ListAssinaturas(c *gin.Context) {
	var items []models.Assinatura
	if h.list(c, &items, "Usuario", "Plano") {
		for i := range items {
			items[i].Usuario.SenhaHash = ""
		}
		c.JSON(http.StatusOK, gin.H{"assinaturas": items, "total": len(items)})
	}
}
func (h *AdminHandler) GetAssinatura(c *gin.Context) {
	h.getByID(c, "Assinatura", &models.Assinatura{}, "Usuario", "Plano", "Pagamentos")
}
func (h *AdminHandler) CreateAssinatura(c *gin.Context) {
	var req models.AssinaturaCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	if !h.exists(&models.Usuario{}, req.IDUsuario) || !h.exists(&models.Plano{}, req.IDPlano) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Usuario ou plano nao existe"})
		return
	}
	status := req.Status
	if status == "" {
		status = models.StatusAssinaturaPendente
	}
	item := models.Assinatura{IDUsuario: req.IDUsuario, IDPlano: req.IDPlano, DataInicio: req.DataInicio, DataFim: req.DataFim, Status: status, ValorPago: req.ValorPago, RenovacaoAutomatica: req.RenovacaoAutomatica, CodigoTransacao: clean(req.CodigoTransacao), DataAtivacao: req.DataAtivacao, EmailBoasVindasEnviado: req.EmailBoasVindasEnviado}
	h.create(c, &item)
}
func (h *AdminHandler) UpdateAssinatura(c *gin.Context) { h.updateAssinatura(c) }
func (h *AdminHandler) DeleteAssinatura(c *gin.Context) {
	h.softUpdate(c, &models.Assinatura{}, "id_assinatura", "status", models.StatusAssinaturaCancelada, "Assinatura cancelada com sucesso")
}

// Pagamentos
func (h *AdminHandler) ListPagamentos(c *gin.Context) {
	var items []models.Pagamento
	if h.list(c, &items, "Assinatura") {
		c.JSON(http.StatusOK, gin.H{"pagamentos": items, "total": len(items)})
	}
}
func (h *AdminHandler) GetPagamento(c *gin.Context) {
	h.getByID(c, "Pagamento", &models.Pagamento{}, "Assinatura")
}
func (h *AdminHandler) CreatePagamento(c *gin.Context) {
	var req models.PagamentoRequest
	if !bindJSON(c, &req) {
		return
	}
	h.upsertPagamento(c, 0, req)
}
func (h *AdminHandler) UpdatePagamento(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req models.PagamentoRequest
	if !bindJSON(c, &req) {
		return
	}
	h.upsertPagamento(c, id, req)
}
func (h *AdminHandler) DeletePagamento(c *gin.Context) {
	h.hardDelete(c, &models.Pagamento{}, "id_pagamento", "Pagamento deletado com sucesso")
}

// Beneficios
func (h *AdminHandler) ListBeneficios(c *gin.Context) {
	var items []models.Beneficio
	if h.list(c, &items, "Parceiro") {
		c.JSON(http.StatusOK, gin.H{"beneficios": items, "total": len(items)})
	}
}
func (h *AdminHandler) GetBeneficio(c *gin.Context) {
	h.getByID(c, "Beneficio", &models.Beneficio{}, "Parceiro")
}
func (h *AdminHandler) CreateBeneficio(c *gin.Context) {
	var req models.BeneficioRequest
	if !bindJSON(c, &req) {
		return
	}
	h.upsertBeneficio(c, 0, req)
}
func (h *AdminHandler) UpdateBeneficio(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req models.BeneficioRequest
	if !bindJSON(c, &req) {
		return
	}
	h.upsertBeneficio(c, id, req)
}
func (h *AdminHandler) DeleteBeneficio(c *gin.Context) {
	h.softUpdate(c, &models.Beneficio{}, "id_beneficio", "ativo", false, "Beneficio desativado com sucesso")
}

// Cursos
func (h *AdminHandler) ListCursos(c *gin.Context) {
	var items []models.Curso
	if h.list(c, &items, "Parceiro") {
		c.JSON(http.StatusOK, gin.H{"cursos": items, "total": len(items)})
	}
}
func (h *AdminHandler) GetCurso(c *gin.Context) { h.getByID(c, "Curso", &models.Curso{}, "Parceiro") }
func (h *AdminHandler) CreateCurso(c *gin.Context) {
	var req models.CursoRequest
	if !bindJSON(c, &req) {
		return
	}
	h.upsertCurso(c, 0, req)
}
func (h *AdminHandler) UpdateCurso(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req models.CursoRequest
	if !bindJSON(c, &req) {
		return
	}
	h.upsertCurso(c, id, req)
}
func (h *AdminHandler) DeleteCurso(c *gin.Context) {
	h.softUpdate(c, &models.Curso{}, "id_curso", "ativo", false, "Curso desativado com sucesso")
}

// Responsaveis
func (h *AdminHandler) ListResponsaveis(c *gin.Context) {
	var items []models.Responsavel
	if h.list(c, &items, "Usuario") {
		c.JSON(http.StatusOK, gin.H{"responsaveis": items, "total": len(items)})
	}
}
func (h *AdminHandler) GetResponsavel(c *gin.Context) {
	h.getByID(c, "Responsavel", &models.Responsavel{}, "Usuario")
}
func (h *AdminHandler) CreateResponsavel(c *gin.Context) {
	var req models.ResponsavelRequest
	if !bindJSON(c, &req) {
		return
	}
	h.upsertResponsavel(c, 0, req)
}
func (h *AdminHandler) UpdateResponsavel(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req models.ResponsavelRequest
	if !bindJSON(c, &req) {
		return
	}
	h.upsertResponsavel(c, id, req)
}
func (h *AdminHandler) DeleteResponsavel(c *gin.Context) {
	h.hardDelete(c, &models.Responsavel{}, "id_responsavel", "Responsavel deletado com sucesso")
}

// Usuario cursos
func (h *AdminHandler) ListUsuarioCursos(c *gin.Context) {
	var items []models.UsuarioCurso
	if h.list(c, &items, "Usuario", "Curso") {
		c.JSON(http.StatusOK, gin.H{"usuario_cursos": items, "total": len(items)})
	}
}
func (h *AdminHandler) GetUsuarioCurso(c *gin.Context) {
	h.getByID(c, "UsuarioCurso", &models.UsuarioCurso{}, "Usuario", "Curso")
}
func (h *AdminHandler) CreateUsuarioCurso(c *gin.Context) {
	var req models.UsuarioCursoRequest
	if !bindJSON(c, &req) {
		return
	}
	h.upsertUsuarioCurso(c, 0, req)
}
func (h *AdminHandler) UpdateUsuarioCurso(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req models.UsuarioCursoRequest
	if !bindJSON(c, &req) {
		return
	}
	h.upsertUsuarioCurso(c, id, req)
}
func (h *AdminHandler) DeleteUsuarioCurso(c *gin.Context) {
	h.softUpdate(c, &models.UsuarioCurso{}, "id_usuario_curso", "status", models.StatusUsuarioCursoCancelado, "Vinculo de curso cancelado com sucesso")
}

// Logs
func (h *AdminHandler) ListLogsAuditoria(c *gin.Context) {
	var items []models.LogAuditoria
	if h.list(c, &items, "Usuario") {
		c.JSON(http.StatusOK, gin.H{"logs_auditoria": items, "total": len(items)})
	}
}
func (h *AdminHandler) GetLogAuditoria(c *gin.Context) {
	h.getByID(c, "LogAuditoria", &models.LogAuditoria{}, "Usuario")
}
func (h *AdminHandler) CreateLogAuditoria(c *gin.Context) {
	var req models.LogAuditoriaRequest
	if !bindJSON(c, &req) {
		return
	}
	if !h.exists(&models.Usuario{}, req.IDUsuario) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Usuario nao existe"})
		return
	}
	h.create(c, &models.LogAuditoria{IDUsuario: req.IDUsuario, Acao: clean(req.Acao), Descricao: clean(req.Descricao), IPUsuario: clean(req.IPUsuario)})
}
func (h *AdminHandler) UpdateLogAuditoria(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req models.LogAuditoriaRequest
	if !bindJSON(c, &req) {
		return
	}
	if !h.exists(&models.Usuario{}, req.IDUsuario) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Usuario nao existe"})
		return
	}
	var item models.LogAuditoria
	if err := h.db.First(&item, id).Error; err != nil {
		if !notFound(c, "LogAuditoria", err) {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar log"})
		}
		return
	}
	item.IDUsuario = req.IDUsuario
	item.Acao = clean(req.Acao)
	item.Descricao = clean(req.Descricao)
	item.IPUsuario = clean(req.IPUsuario)
	h.save(c, &item)
}
func (h *AdminHandler) DeleteLogAuditoria(c *gin.Context) {
	h.hardDelete(c, &models.LogAuditoria{}, "id_log", "Log deletado com sucesso")
}

func (h *AdminHandler) getByID(c *gin.Context, name string, out interface{}, preload ...string) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	query := h.db
	for _, item := range preload {
		query = query.Preload(item)
	}
	if err := query.First(out, id).Error; err != nil {
		if !notFound(c, name, err) {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar registro"})
		}
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *AdminHandler) create(c *gin.Context, item interface{}) {
	if err := h.db.Create(item).Error; err != nil {
		log.Printf("Erro ao criar registro: %v", err)
		c.JSON(http.StatusConflict, gin.H{"erro": "Nao foi possivel criar registro"})
		return
	}
	c.JSON(http.StatusCreated, item)
}
func (h *AdminHandler) save(c *gin.Context, item interface{}) {
	if err := h.db.Save(item).Error; err != nil {
		log.Printf("Erro ao salvar registro: %v", err)
		c.JSON(http.StatusConflict, gin.H{"erro": "Nao foi possivel salvar registro"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *AdminHandler) softUpdate(c *gin.Context, model interface{}, idColumn, field string, value interface{}, success string) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	result := h.db.Model(model).Where(idColumn+" = ?", id).Update(field, value)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao atualizar registro"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Registro nao encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mensagem": success})
}

func (h *AdminHandler) hardDelete(c *gin.Context, model interface{}, idColumn, success string) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	result := h.db.Where(idColumn+" = ?", id).Delete(model)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao deletar registro"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Registro nao encontrado"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mensagem": success})
}

func (h *AdminHandler) updateAssinatura(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req models.AssinaturaCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	if !h.exists(&models.Usuario{}, req.IDUsuario) || !h.exists(&models.Plano{}, req.IDPlano) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Usuario ou plano nao existe"})
		return
	}
	var item models.Assinatura
	if err := h.db.First(&item, id).Error; err != nil {
		if !notFound(c, "Assinatura", err) {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar assinatura"})
		}
		return
	}
	status := req.Status
	if status == "" {
		status = item.Status
	}
	item.IDUsuario, item.IDPlano, item.DataInicio, item.DataFim, item.Status, item.ValorPago = req.IDUsuario, req.IDPlano, req.DataInicio, req.DataFim, status, req.ValorPago
	item.RenovacaoAutomatica, item.CodigoTransacao, item.DataAtivacao, item.EmailBoasVindasEnviado = req.RenovacaoAutomatica, clean(req.CodigoTransacao), req.DataAtivacao, req.EmailBoasVindasEnviado
	h.save(c, &item)
}

func (h *AdminHandler) upsertPagamento(c *gin.Context, id uint, req models.PagamentoRequest) {
	if !h.exists(&models.Assinatura{}, req.IDAssinatura) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Assinatura nao existe"})
		return
	}
	status := req.StatusPagamento
	if status == "" {
		status = models.StatusPagamentoPendente
	}
	item := models.Pagamento{IDAssinatura: req.IDAssinatura, MetodoPagamento: cleanSlug(req.MetodoPagamento), Valor: req.Valor, StatusPagamento: status, DataPagamento: req.DataPagamento, IDGateway: clean(req.IDGateway), ComprovanteURL: cleanURL(req.ComprovanteURL)}
	if id > 0 {
		item.IDPagamento = id
		var current models.Pagamento
		if err := h.db.First(&current, id).Error; err != nil {
			if !notFound(c, "Pagamento", err) {
				c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar pagamento"})
			}
			return
		}
	}
	if id > 0 {
		h.save(c, &item)
	} else {
		h.create(c, &item)
	}
}

func (h *AdminHandler) upsertBeneficio(c *gin.Context, id uint, req models.BeneficioRequest) {
	if !h.exists(&models.Parceiro{}, req.IDParceiro) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Parceiro nao existe"})
		return
	}
	item := models.Beneficio{IDParceiro: req.IDParceiro, Titulo: clean(req.Titulo), Descricao: clean(req.Descricao), TipoDesconto: req.TipoDesconto, ValorDesconto: req.ValorDesconto, RegrasUso: clean(req.RegrasUso), DataInicio: req.DataInicio, DataFim: req.DataFim, Ativo: boolValue(req.Ativo, true)}
	if id > 0 {
		item.IDBeneficio = id
		var current models.Beneficio
		if err := h.db.First(&current, id).Error; err != nil {
			if !notFound(c, "Beneficio", err) {
				c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar beneficio"})
			}
			return
		}
		item.CreatedAt = current.CreatedAt
	}
	if id > 0 {
		h.save(c, &item)
	} else {
		h.create(c, &item)
	}
}

func (h *AdminHandler) upsertCurso(c *gin.Context, id uint, req models.CursoRequest) {
	if !h.exists(&models.Parceiro{}, req.IDParceiro) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Parceiro nao existe"})
		return
	}
	item := models.Curso{NomeCurso: clean(req.NomeCurso), Instituicao: clean(req.Instituicao), IDParceiro: req.IDParceiro, PercentualDesconto: req.PercentualDesconto, TipoMatricula: req.TipoMatricula, Ativo: boolValue(req.Ativo, true)}
	if id > 0 {
		item.IDCurso = id
		var current models.Curso
		if err := h.db.First(&current, id).Error; err != nil {
			if !notFound(c, "Curso", err) {
				c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar curso"})
			}
			return
		}
		item.CreatedAt = current.CreatedAt
	}
	if id > 0 {
		h.save(c, &item)
	} else {
		h.create(c, &item)
	}
}

func (h *AdminHandler) upsertResponsavel(c *gin.Context, id uint, req models.ResponsavelRequest) {
	if !h.exists(&models.Usuario{}, req.IDUsuario) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Usuario nao existe"})
		return
	}
	if !utils.ValidateCPF(req.CPF) || !utils.ValidateEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "CPF ou email invalido"})
		return
	}
	item := models.Responsavel{IDUsuario: req.IDUsuario, NomeCompleto: cleanText(req.NomeCompleto, 100), CPFHash: utils.HashCPF(req.CPF), Email: cleanEmail(req.Email), Telefone: cleanPhone(req.Telefone), Parentesco: cleanText(req.Parentesco, 50)}
	if id > 0 {
		item.IDResponsavel = id
		var current models.Responsavel
		if err := h.db.First(&current, id).Error; err != nil {
			if !notFound(c, "Responsavel", err) {
				c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar responsavel"})
			}
			return
		}
		item.CreatedAt = current.CreatedAt
	}
	if id > 0 {
		h.save(c, &item)
	} else {
		h.create(c, &item)
	}
}

func (h *AdminHandler) upsertUsuarioCurso(c *gin.Context, id uint, req models.UsuarioCursoRequest) {
	if !h.exists(&models.Usuario{}, req.IDUsuario) || !h.exists(&models.Curso{}, req.IDCurso) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Usuario ou curso nao existe"})
		return
	}
	status := req.Status
	if status == "" {
		status = models.StatusUsuarioCursoPendente
	}
	item := models.UsuarioCurso{IDUsuario: req.IDUsuario, IDCurso: req.IDCurso, TipoMatricula: req.TipoMatricula, Matricula: clean(req.Matricula), ComprovanteURL: cleanURL(req.ComprovanteURL), Status: status}
	if id > 0 {
		item.IDUsuarioCurso = id
		var current models.UsuarioCurso
		if err := h.db.First(&current, id).Error; err != nil {
			if !notFound(c, "UsuarioCurso", err) {
				c.JSON(http.StatusInternalServerError, gin.H{"erro": "Erro ao buscar vinculo"})
			}
			return
		}
		item.DataVinculo = current.DataVinculo
	}
	if id > 0 {
		h.save(c, &item)
	} else {
		h.create(c, &item)
	}
}
