package models

import "time"

type Categoria struct {
	IDCategoria uint      `gorm:"primaryKey;column:id_categoria" json:"id_categoria"`
	Slug        string    `gorm:"column:slug;size:100;uniqueIndex;not null" json:"slug"`
	Nome        string    `gorm:"column:nome;size:100;not null" json:"nome"`
	Icone       string    `gorm:"column:icone;size:20" json:"icone"`
	Ordem       int       `gorm:"column:ordem;not null;default:0" json:"ordem"`
	Ativo       bool      `gorm:"column:ativo;not null;default:1" json:"ativo"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Categoria) TableName() string {
	return "categorias"
}
