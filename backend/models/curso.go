package models

import "time"

const (
	TipoMatriculaNova      = "nova"
	TipoMatriculaRenovacao = "renovacao"
	TipoMatriculaTransfer  = "transferencia"
)

type Curso struct {
	IDCurso            uint      `gorm:"primaryKey;column:id_curso" json:"id_curso"`
	NomeCurso          string    `gorm:"column:nome_curso;size:150;not null" json:"nome_curso"`
	Instituicao        string    `gorm:"column:instituicao;size:150;not null" json:"instituicao"`
	IDParceiro         uint      `gorm:"column:id_parceiro;index;not null" json:"id_parceiro"`
	PercentualDesconto float64   `gorm:"column:percentual_desconto;type:numeric;not null" json:"percentual_desconto"`
	TipoMatricula      string    `gorm:"column:tipo_matricula;size:30;not null" json:"tipo_matricula"`
	Ativo              bool      `gorm:"column:ativo;not null;default:1" json:"ativo"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Parceiro      Parceiro       `gorm:"foreignKey:IDParceiro" json:"parceiro,omitempty"`
	UsuarioCursos []UsuarioCurso `gorm:"foreignKey:IDCurso" json:"usuario_cursos,omitempty"`
}

func (Curso) TableName() string {
	return "cursos"
}

type CursoRequest struct {
	NomeCurso          string  `json:"nome_curso" binding:"required,min=2,max=150"`
	Instituicao        string  `json:"instituicao" binding:"required,min=2,max=150"`
	IDParceiro         uint    `json:"id_parceiro" binding:"required"`
	PercentualDesconto float64 `json:"percentual_desconto" binding:"required,min=0,max=100"`
	TipoMatricula      string  `json:"tipo_matricula" binding:"required,oneof=nova renovacao transferencia"`
	Ativo              *bool   `json:"ativo"`
}
