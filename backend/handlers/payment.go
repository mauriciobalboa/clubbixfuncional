package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"clubbix/backend/config"
	"clubbix/backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaymentHandler struct {
	db         *gorm.DB
	cfg        *config.Config
	httpClient *http.Client
}

func NewPaymentHandler(db *gorm.DB, cfg *config.Config) *PaymentHandler {
	return &PaymentHandler{
		db:  db,
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.RequestTimeout,
		},
	}
}

type appmaxProduct struct {
	SKU   string `json:"sku"`
	Qty   int    `json:"qty"`
	Price int    `json:"price"`
}

type appmaxCustomer struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Document string `json:"document"`
	Phone    string `json:"phone"`
}

type appmaxCreditCard struct {
	Number          string `json:"number"`
	HolderName      string `json:"holder_name"`
	ExpirationMonth string `json:"expiration_month"`
	ExpirationYear  string `json:"expiration_year"`
	CVV             string `json:"cvv"`
	Installments    int    `json:"installments,omitempty"`
}

type appmaxPayment struct {
	Method     string           `json:"method"`
	CreditCard appmaxCreditCard `json:"credit_card"`
}

type appmaxOrderRequest struct {
	Products []appmaxProduct `json:"products"`
	Customer appmaxCustomer  `json:"customer"`
	Payment  appmaxPayment   `json:"payment"`
}

func (h *PaymentHandler) CreatePagamento(c *gin.Context) {
	if h.cfg.AppmaxToken == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"erro": "Gateway de pagamento nao configurado"})
		return
	}

	usuarioID, ok := currentUsuarioID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "Nao autenticado"})
		return
	}

	var req models.CriarPagamentoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados invalidos", "detalhe": err.Error()})
		return
	}

	if !validCard(req.Cartao) {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "Dados do cartao invalidos"})
		return
	}

	var assinatura models.Assinatura
	if err := h.db.Preload("Plano").Where("id_assinatura = ? AND id_usuario = ?", req.IDAssinatura, usuarioID).First(&assinatura).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": "Assinatura nao encontrada"})
		return
	}

	payload := appmaxOrderRequest{
		Products: []appmaxProduct{
			{
				SKU:   fmt.Sprintf("PLANO-%d", assinatura.IDPlano),
				Qty:   1,
				Price: toCents(assinatura.ValorPago),
			},
		},
		Customer: appmaxCustomer{
			Name:     clean(req.Cliente.Nome),
			Email:    clean(req.Cliente.Email),
			Document: onlyDigits(req.Cliente.Documento),
			Phone:    onlyDigits(req.Cliente.Telefone),
		},
		Payment: appmaxPayment{
			Method: "credit_card",
			CreditCard: appmaxCreditCard{
				Number:          onlyDigits(req.Cartao.Numero),
				HolderName:      clean(req.Cartao.NomeImpresso),
				ExpirationMonth: req.Cartao.MesExpiracao,
				ExpirationYear:  req.Cartao.AnoExpiracao,
				CVV:             req.Cartao.CVV,
				Installments:    installments(req.Cartao.Parcelas),
			},
		},
	}

	gatewayResp, statusCode, err := h.createAppmaxOrder(c, payload)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"erro": "Erro ao comunicar com gateway", "detalhe": err.Error()})
		return
	}
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		c.JSON(http.StatusBadGateway, gin.H{"erro": "Gateway recusou a requisicao", "status_gateway": statusCode, "gateway": sanitizeGatewayResponse(gatewayResp)})
		return
	}

	idGateway := extractGatewayID(gatewayResp)
	now := time.Now()
	pagamento := models.Pagamento{
		IDAssinatura:    assinatura.IDAssinatura,
		MetodoPagamento: models.MetodoPagamentoCartao,
		Valor:           assinatura.ValorPago,
		StatusPagamento: models.StatusPagamentoPendente,
		DataPagamento:   &now,
		IDGateway:       idGateway,
	}

	if err := h.db.Create(&pagamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "Pagamento enviado, mas nao foi possivel registrar no banco", "id_gateway": idGateway})
		return
	}

	c.JSON(http.StatusCreated, models.CriarPagamentoResponse{
		PagamentoID: pagamento.IDPagamento,
		Status:      pagamento.StatusPagamento,
		IDGateway:   pagamento.IDGateway,
		Gateway:     sanitizeGatewayResponse(gatewayResp),
	})
}

func (h *PaymentHandler) createAppmaxOrder(c *gin.Context, payload appmaxOrderRequest) (map[string]interface{}, int, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, h.cfg.AppmaxOrderURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.cfg.AppmaxToken)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	var decoded map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &decoded); err != nil {
			decoded = map[string]interface{}{"raw": string(body)}
		}
	} else {
		decoded = map[string]interface{}{}
	}

	return decoded, resp.StatusCode, nil
}

func currentUsuarioID(c *gin.Context) (uint, bool) {
	rawID, exists := c.Get("usuarioID")
	if !exists {
		rawID, exists = c.Get("membroID")
	}
	if !exists {
		return 0, false
	}
	usuarioID, err := strconv.ParseUint(rawID.(string), 10, 32)
	if err != nil || usuarioID == 0 {
		return 0, false
	}
	return uint(usuarioID), true
}

func validCard(card models.PagamentoCartaoRequest) bool {
	number := onlyDigits(card.Numero)
	cvv := onlyDigits(card.CVV)
	month, err := strconv.Atoi(card.MesExpiracao)
	if err != nil || month < 1 || month > 12 {
		return false
	}
	return len(number) >= 13 && len(number) <= 19 && len(cvv) >= 3 && len(cvv) <= 4
}

func onlyDigits(value string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, value)
}

func toCents(value float64) int {
	return int(value*100 + 0.5)
}

func installments(value int) int {
	if value <= 0 {
		return 1
	}
	return value
}

func extractGatewayID(resp map[string]interface{}) string {
	for _, key := range []string{"id", "order_id", "id_order"} {
		if value, ok := resp[key]; ok {
			return fmt.Sprintf("%v", value)
		}
	}
	if data, ok := resp["data"].(map[string]interface{}); ok {
		for _, key := range []string{"id", "order_id", "id_order"} {
			if value, ok := data[key]; ok {
				return fmt.Sprintf("%v", value)
			}
		}
	}
	return ""
}

func sanitizeGatewayResponse(resp map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{}, len(resp))
	for key, value := range resp {
		lowerKey := strings.ToLower(key)
		if strings.Contains(lowerKey, "card") || strings.Contains(lowerKey, "cartao") || strings.Contains(lowerKey, "number") || strings.Contains(lowerKey, "cvv") {
			sanitized[key] = "[redacted]"
			continue
		}

		if nested, ok := value.(map[string]interface{}); ok {
			sanitized[key] = sanitizeGatewayResponse(nested)
			continue
		}

		sanitized[key] = value
	}
	return sanitized
}
