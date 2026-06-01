package models

import "time"

const (
	StatusUsuarioCursoPendente  = "pendente"
	StatusUsuarioCursoAprovado  = "aprovado"
	StatusUsuarioCursoRecusado  = "recusado"
	StatusUsuarioCursoCancelado = "cancelado"
)

type UsuarioCurso struct {
	IDUsuarioCurso uint      `gorm:"primaryKey;column:id_usuario_curso" json:"id_usuario_curso"`
	IDUsuario      uint      `gorm:"column:id_usuario;index;not null" json:"id_usuario"`
	IDCurso        uint      `gorm:"column:id_curso;index;not null" json:"id_curso"`
	TipoMatricula  string    `gorm:"column:tipo_matricula;size:30;not null" json:"tipo_matricula"`
	Matricula      string    `gorm:"column:matricula;size:50" json:"matricula"`
	ComprovanteURL string    `gorm:"column:comprovante_url;size:500" json:"comprovante_url"`
	Status         string    `gorm:"column:status;size:30;not null;default:'pendente'" json:"status"`
	DataVinculo    time.Time `gorm:"column:data_vinculo;autoCreateTime" json:"data_vinculo"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Usuario Usuario `gorm:"foreignKey:IDUsuario" json:"usuario,omitempty"`
	Curso   Curso   `gorm:"foreignKey:IDCurso" json:"curso,omitempty"`
}

func (UsuarioCurso) TableName() string {
	return "usuario_cursos"
}

type UsuarioCursoRequest struct {
	IDUsuario      uint   `json:"id_usuario" binding:"required"`
	IDCurso        uint   `json:"id_curso" binding:"required"`
	TipoMatricula  string `json:"tipo_matricula" binding:"required,oneof=nova renovacao transferencia"`
	Matricula      string `json:"matricula" binding:"omitempty,max=50"`
	ComprovanteURL string `json:"comprovante_url" binding:"omitempty,url,max=500"`
	Status         string `json:"status" binding:"omitempty,oneof=pendente aprovado recusado cancelado"`
}
