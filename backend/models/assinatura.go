package models

import "time"

const (
	StatusAssinaturaAtiva     = "ativa"
	StatusAssinaturaPendente  = "pendente"
	StatusAssinaturaCancelada = "cancelada"
	StatusAssinaturaExpirada  = "expirada"
)

type Assinatura struct {
	IDAssinatura           uint       `gorm:"primaryKey;column:id_assinatura" json:"id_assinatura"`
	IDUsuario              uint       `gorm:"column:id_usuario;index;not null" json:"id_usuario"`
	IDPlano                uint       `gorm:"column:id_plano;index;not null" json:"id_plano"`
	DataInicio             time.Time  `gorm:"column:data_inicio;type:date;not null" json:"data_inicio"`
	DataFim                time.Time  `gorm:"column:data_fim;type:date;not null" json:"data_fim"`
	Status                 string     `gorm:"column:status;size:30;not null;default:'pendente'" json:"status"`
	ValorPago              float64    `gorm:"column:valor_pago;type:numeric;not null" json:"valor_pago"`
	RenovacaoAutomatica    bool       `gorm:"column:renovacao_automatica;not null;default:0" json:"renovacao_automatica"`
	CodigoTransacao        string     `gorm:"column:codigo_transacao;size:100" json:"codigo_transacao"`
	DataAtivacao           *time.Time `gorm:"column:data_ativacao" json:"data_ativacao"`
	EmailBoasVindasEnviado bool       `gorm:"column:email_boas_vindas_enviado;not null;default:0" json:"email_boas_vindas_enviado"`
	CreatedAt              time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Usuario    Usuario     `gorm:"foreignKey:IDUsuario" json:"usuario,omitempty"`
	Plano      Plano       `gorm:"foreignKey:IDPlano" json:"plano,omitempty"`
	Pagamentos []Pagamento `gorm:"foreignKey:IDAssinatura" json:"pagamentos,omitempty"`
}

func (Assinatura) TableName() string {
	return "assinaturas"
}

type AssinaturaRequest struct {
	IDPlano             uint `json:"id_plano" binding:"required"`
	RenovacaoAutomatica bool `json:"renovacao_automatica"`
}

type AssinaturaCreateRequest struct {
	IDUsuario              uint       `json:"id_usuario" binding:"required"`
	IDPlano                uint       `json:"id_plano" binding:"required"`
	DataInicio             time.Time  `json:"data_inicio" binding:"required"`
	DataFim                time.Time  `json:"data_fim" binding:"required"`
	Status                 string     `json:"status" binding:"omitempty,oneof=ativa pendente cancelada expirada"`
	ValorPago              float64    `json:"valor_pago" binding:"omitempty,min=0"`
	RenovacaoAutomatica    bool       `json:"renovacao_automatica"`
	CodigoTransacao        string     `json:"codigo_transacao" binding:"omitempty,max=100"`
	DataAtivacao           *time.Time `json:"data_ativacao"`
	EmailBoasVindasEnviado bool       `json:"email_boas_vindas_enviado"`
}
