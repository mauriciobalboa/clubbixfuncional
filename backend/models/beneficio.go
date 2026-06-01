package models

import "time"

const (
	TipoDescontoPercentual = "percentual"
	TipoDescontoValorFixo  = "valor_fixo"
	TipoDescontoBrinde     = "brinde"
)

type Beneficio struct {
	IDBeneficio   uint      `gorm:"primaryKey;column:id_beneficio" json:"id_beneficio"`
	IDParceiro    uint      `gorm:"column:id_parceiro;index;not null" json:"id_parceiro"`
	Titulo        string    `gorm:"column:titulo;size:150;not null" json:"titulo"`
	Descricao     string    `gorm:"column:descricao;type:text" json:"descricao"`
	TipoDesconto  string    `gorm:"column:tipo_desconto;size:30;not null" json:"tipo_desconto"`
	ValorDesconto float64   `gorm:"column:valor_desconto;type:numeric;not null" json:"valor_desconto"`
	RegrasUso     string    `gorm:"column:regras_uso;type:text" json:"regras_uso"`
	DataInicio    time.Time `gorm:"column:data_inicio;type:date;not null" json:"data_inicio"`
	DataFim       time.Time `gorm:"column:data_fim;type:date;not null" json:"data_fim"`
	Ativo         bool      `gorm:"column:ativo;not null;default:1" json:"ativo"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Parceiro Parceiro `gorm:"foreignKey:IDParceiro" json:"parceiro,omitempty"`
}

func (Beneficio) TableName() string {
	return "beneficios"
}

type BeneficioRequest struct {
	IDParceiro    uint      `json:"id_parceiro" binding:"required"`
	Titulo        string    `json:"titulo" binding:"required,min=2,max=150"`
	Descricao     string    `json:"descricao" binding:"omitempty,max=3000"`
	TipoDesconto  string    `json:"tipo_desconto" binding:"required,oneof=percentual valor_fixo brinde"`
	ValorDesconto float64   `json:"valor_desconto" binding:"required,min=0"`
	RegrasUso     string    `json:"regras_uso" binding:"omitempty,max=3000"`
	DataInicio    time.Time `json:"data_inicio" binding:"required"`
	DataFim       time.Time `json:"data_fim" binding:"required"`
	Ativo         *bool     `json:"ativo"`
}
