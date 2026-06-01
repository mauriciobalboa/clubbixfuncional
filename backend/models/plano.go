package models

import "time"

const (
	DuracaoTipoMeses     = "meses"
	DuracaoTipoDias      = "dias"
	DuracaoTipoVitalicio = "vitalicio"
)

type Plano struct {
	IDPlano      uint      `gorm:"primaryKey;column:id_plano" json:"id_plano"`
	NomePlano    string    `gorm:"column:nome_plano;size:100;not null" json:"nome_plano"`
	Descricao    string    `gorm:"column:descricao;type:text" json:"descricao"`
	Valor        float64   `gorm:"column:valor;type:numeric;not null" json:"valor"`
	DuracaoMeses int       `gorm:"column:duracao_meses;not null" json:"duracao_meses"`
	DuracaoTipo  string    `gorm:"column:duracao_tipo;size:30;not null;default:'meses'" json:"duracao_tipo"`
	Ativo        bool      `gorm:"column:ativo;not null;default:1" json:"ativo"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Assinaturas []Assinatura `gorm:"foreignKey:IDPlano" json:"assinaturas,omitempty"`
}

func (Plano) TableName() string {
	return "planos"
}

type PlanoRequest struct {
	NomePlano    string  `json:"nome_plano" binding:"required,min=2,max=100"`
	Descricao    string  `json:"descricao" binding:"omitempty,max=2000"`
	Valor        float64 `json:"valor" binding:"required,min=0"`
	DuracaoMeses int     `json:"duracao_meses" binding:"required,min=0,max=1200"`
	DuracaoTipo  string  `json:"duracao_tipo" binding:"required,oneof=meses dias vitalicio"`
	Ativo        *bool   `json:"ativo"`
}
