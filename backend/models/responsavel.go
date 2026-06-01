package models

import "time"

type Responsavel struct {
	IDResponsavel uint      `gorm:"primaryKey;column:id_responsavel" json:"id_responsavel"`
	IDUsuario     uint      `gorm:"column:id_usuario;index;not null" json:"id_usuario"`
	NomeCompleto  string    `gorm:"column:nome_completo;size:100;not null" json:"nome_completo"`
	CPFHash       string    `gorm:"column:cpf_hash;size:255;not null" json:"-"`
	Email         string    `gorm:"column:email;size:150;not null" json:"email"`
	Telefone      string    `gorm:"column:telefone;size:20" json:"telefone"`
	Parentesco    string    `gorm:"column:parentesco;size:50;not null" json:"parentesco"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Usuario Usuario `gorm:"foreignKey:IDUsuario" json:"usuario,omitempty"`
}

func (Responsavel) TableName() string {
	return "responsaveis"
}

type ResponsavelRequest struct {
	IDUsuario    uint   `json:"id_usuario" binding:"required"`
	NomeCompleto string `json:"nome_completo" binding:"required,min=3,max=100"`
	CPF          string `json:"cpf" binding:"required,max=20"`
	Email        string `json:"email" binding:"required,email,max=150"`
	Telefone     string `json:"telefone" binding:"omitempty,max=20"`
	Parentesco   string `json:"parentesco" binding:"required,min=2,max=50"`
}
