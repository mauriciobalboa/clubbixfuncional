package models

import "time"

type LogAuditoria struct {
	IDLog     uint      `gorm:"primaryKey;column:id_log" json:"id_log"`
	IDUsuario uint      `gorm:"column:id_usuario;index;not null" json:"id_usuario"`
	Acao      string    `gorm:"column:acao;size:200;not null" json:"acao"`
	Descricao string    `gorm:"column:descricao;type:text" json:"descricao"`
	IPUsuario string    `gorm:"column:ip_usuario;size:50" json:"ip_usuario"`
	DataHora  time.Time `gorm:"column:data_hora;autoCreateTime" json:"data_hora"`

	Usuario Usuario `gorm:"foreignKey:IDUsuario" json:"usuario,omitempty"`
}

func (LogAuditoria) TableName() string {
	return "logs_auditoria"
}

type LogAuditoriaRequest struct {
	IDUsuario uint   `json:"id_usuario" binding:"required"`
	Acao      string `json:"acao" binding:"required,min=2,max=200"`
	Descricao string `json:"descricao" binding:"omitempty,max=3000"`
	IPUsuario string `json:"ip_usuario" binding:"omitempty,max=50"`
}
