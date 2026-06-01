package models

import "time"

const (
	TipoUsuarioAdmin       = "admin"
	TipoUsuarioAssociado   = "associado"
	TipoUsuarioResponsavel = "responsavel"

	StatusContaAtiva     = "ativa"
	StatusContaInativa   = "inativa"
	StatusContaBloqueada = "bloqueada"
)

type Usuario struct {
	IDUsuario        uint       `gorm:"primaryKey;column:id_usuario" json:"id_usuario"`
	NomeCompleto     string     `gorm:"column:nome_completo;size:100;not null" json:"nome_completo"`
	CPFHash          string     `gorm:"column:cpf_hash;size:255;uniqueIndex;not null" json:"-"`
	Email            string     `gorm:"column:email;size:150;uniqueIndex;not null" json:"email"`
	Telefone         string     `gorm:"column:telefone;size:20" json:"telefone"`
	SenhaHash        string     `gorm:"column:senha_hash;size:255;not null" json:"-"`
	TipoUsuario      string     `gorm:"column:tipo_usuario;size:30;not null;default:'associado'" json:"tipo_usuario"`
	StatusConta      string     `gorm:"column:status_conta;size:30;not null;default:'ativa'" json:"status_conta"`
	DataCadastro     time.Time  `gorm:"column:data_cadastro;autoCreateTime" json:"data_cadastro"`
	DataUltimoAcesso *time.Time `gorm:"column:data_ultimo_acesso" json:"data_ultimo_acesso"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Assinaturas   []Assinatura   `gorm:"foreignKey:IDUsuario" json:"assinaturas,omitempty"`
	Responsaveis  []Responsavel  `gorm:"foreignKey:IDUsuario" json:"responsaveis,omitempty"`
	UsuarioCursos []UsuarioCurso `gorm:"foreignKey:IDUsuario" json:"usuario_cursos,omitempty"`
	LogsAuditoria []LogAuditoria `gorm:"foreignKey:IDUsuario" json:"logs_auditoria,omitempty"`
}

func (Usuario) TableName() string {
	return "usuarios"
}

type UsuarioCreateRequest struct {
	NomeCompleto string `json:"nome_completo" binding:"required,min=3,max=100"`
	CPF          string `json:"cpf" binding:"required,max=20"`
	Email        string `json:"email" binding:"required,email,max=150"`
	Telefone     string `json:"telefone" binding:"omitempty,max=20"`
	Senha        string `json:"senha" binding:"required,min=8,max=128"`
	TipoUsuario  string `json:"tipo_usuario" binding:"omitempty,oneof=admin associado responsavel"`
	StatusConta  string `json:"status_conta" binding:"omitempty,oneof=ativa inativa bloqueada"`
}

type UsuarioUpdateRequest struct {
	NomeCompleto *string `json:"nome_completo" binding:"omitempty,min=3,max=100"`
	CPF          *string `json:"cpf" binding:"omitempty,max=20"`
	Email        *string `json:"email" binding:"omitempty,email,max=150"`
	Telefone     *string `json:"telefone" binding:"omitempty,max=20"`
	Senha        *string `json:"senha" binding:"omitempty,min=8,max=128"`
	TipoUsuario  *string `json:"tipo_usuario" binding:"omitempty,oneof=admin associado responsavel"`
	StatusConta  *string `json:"status_conta" binding:"omitempty,oneof=ativa inativa bloqueada"`
}
