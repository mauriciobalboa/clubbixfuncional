package models

import "time"

const (
	MetodoPagamentoCartao = "cartao"
	MetodoPagamentoPix    = "pix"
	MetodoPagamentoBoleto = "boleto"

	StatusPagamentoPendente  = "pendente"
	StatusPagamentoAprovado  = "aprovado"
	StatusPagamentoRecusado  = "recusado"
	StatusPagamentoEstornado = "estornado"
)

type Pagamento struct {
	IDPagamento     uint       `gorm:"primaryKey;column:id_pagamento" json:"id_pagamento"`
	IDAssinatura    uint       `gorm:"column:id_assinatura;index;not null" json:"id_assinatura"`
	MetodoPagamento string     `gorm:"column:metodo_pagamento;size:30;not null" json:"metodo_pagamento"`
	Valor           float64    `gorm:"column:valor;type:numeric;not null" json:"valor"`
	StatusPagamento string     `gorm:"column:status_pagamento;size:30;not null;default:'pendente'" json:"status_pagamento"`
	DataPagamento   *time.Time `gorm:"column:data_pagamento" json:"data_pagamento"`
	IDGateway       string     `gorm:"column:id_gateway;size:150" json:"id_gateway"`
	ComprovanteURL  string     `gorm:"column:comprovante_url;size:500" json:"comprovante_url"`

	Assinatura Assinatura `gorm:"foreignKey:IDAssinatura" json:"assinatura,omitempty"`
}

func (Pagamento) TableName() string {
	return "pagamentos"
}

type PagamentoRequest struct {
	IDAssinatura    uint       `json:"id_assinatura" binding:"required"`
	MetodoPagamento string     `json:"metodo_pagamento" binding:"required,oneof=cartao pix boleto"`
	Valor           float64    `json:"valor" binding:"required,min=0"`
	StatusPagamento string     `json:"status_pagamento" binding:"omitempty,oneof=pendente aprovado recusado estornado"`
	DataPagamento   *time.Time `json:"data_pagamento"`
	IDGateway       string     `json:"id_gateway" binding:"omitempty,max=150"`
	ComprovanteURL  string     `json:"comprovante_url" binding:"omitempty,url,max=500"`
}

type PagamentoClienteRequest struct {
	Nome      string `json:"nome" binding:"required,min=3,max=100"`
	Email     string `json:"email" binding:"required,email,max=150"`
	Documento string `json:"documento" binding:"required,max=20"`
	Telefone  string `json:"telefone" binding:"required,max=20"`
}

type PagamentoCartaoRequest struct {
	Numero       string `json:"numero" binding:"required,min=13,max=19"`
	NomeImpresso string `json:"nome_impresso" binding:"required,min=3,max=100"`
	MesExpiracao string `json:"mes_expiracao" binding:"required,len=2"`
	AnoExpiracao string `json:"ano_expiracao" binding:"required,min=2,max=4"`
	CVV          string `json:"cvv" binding:"required,min=3,max=4"`
	Parcelas     int    `json:"parcelas" binding:"omitempty,min=1,max=12"`
}

type CriarPagamentoRequest struct {
	IDAssinatura uint                    `json:"id_assinatura" binding:"required"`
	Cliente      PagamentoClienteRequest `json:"cliente" binding:"required"`
	Cartao       PagamentoCartaoRequest  `json:"cartao" binding:"required"`
}

type CriarPagamentoResponse struct {
	PagamentoID uint                   `json:"id_pagamento"`
	Status      string                 `json:"status_pagamento"`
	IDGateway   string                 `json:"id_gateway,omitempty"`
	Gateway     map[string]interface{} `json:"gateway,omitempty"`
}
