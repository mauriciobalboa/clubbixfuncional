package models

import "time"

const (
	StatusParceiroAtivo    = "ativo"
	StatusParceiroInativo  = "inativo"
	StatusParceiroPendente = "pendente"
)

type Parceiro struct {
	IDParceiro         uint      `gorm:"primaryKey;column:id_parceiro" json:"id_parceiro"`
	NomeEmpresa        string    `gorm:"column:nome_empresa;size:150;not null" json:"nome_empresa"`
	Categoria          string    `gorm:"column:categoria;size:100;index;not null" json:"categoria"`
	Descricao          string    `gorm:"column:descricao;type:text" json:"descricao"`
	PercentualDesconto float64   `gorm:"column:percentual_desconto;type:numeric;not null" json:"percentual_desconto"`
	Endereco           string    `gorm:"column:endereco;type:text" json:"endereco"`
	Telefone           string    `gorm:"column:telefone;size:20" json:"telefone"`
	Email              string    `gorm:"column:email;size:150" json:"email"`
	Site               string    `gorm:"column:site;size:200" json:"site"`
	LogoURL            string    `gorm:"column:logo_url;size:500" json:"logo_url"`
	Status             string    `gorm:"column:status;size:30;not null;default:'ativo'" json:"status"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Beneficios []Beneficio `gorm:"foreignKey:IDParceiro" json:"beneficios,omitempty"`
	Cursos     []Curso     `gorm:"foreignKey:IDParceiro" json:"cursos,omitempty"`
}

func (Parceiro) TableName() string {
	return "parceiros"
}

type ParceiroRequest struct {
	NomeEmpresa        string  `json:"nome_empresa" binding:"required,min=2,max=150"`
	Categoria          string  `json:"categoria" binding:"required,min=2,max=100"`
	Descricao          string  `json:"descricao" binding:"omitempty,max=3000"`
	PercentualDesconto float64 `json:"percentual_desconto" binding:"required,min=0,max=100"`
	Endereco           string  `json:"endereco" binding:"omitempty,max=1000"`
	Telefone           string  `json:"telefone" binding:"omitempty,max=20"`
	Email              string  `json:"email" binding:"omitempty,email,max=150"`
	Site               string  `json:"site" binding:"omitempty,url,max=200"`
	LogoURL            string  `json:"logo_url" binding:"omitempty,url,max=500"`
	Status             string  `json:"status" binding:"omitempty,oneof=ativo inativo pendente"`
}
