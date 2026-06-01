package models

type PublicAssinaturaRequest struct {
	NomeCompleto        string `json:"nome" binding:"required,min=3,max=100"`
	Email               string `json:"email" binding:"required,email,max=150"`
	CPF                 string `json:"cpf" binding:"required,max=20"`
	Telefone            string `json:"telefone" binding:"required,max=20"`
	Plano               string `json:"plano" binding:"required,oneof=basico plus max"`
	MetodoPagamento     string `json:"pagamento" binding:"required,oneof=cartao pix boleto"`
	Convite             string `json:"convite" binding:"omitempty,max=40"`
	RenovacaoAutomatica bool   `json:"renovacao_automatica"`
}

type PublicAssinaturaResponse struct {
	Mensagem    string           `json:"mensagem"`
	Token       string           `json:"token"`
	Session     *FrontendSession `json:"session"`
	Usuario     *Usuario         `json:"usuario"`
	Assinatura  *Assinatura      `json:"assinatura"`
	Pagamento   *Pagamento       `json:"pagamento"`
	PaymentNote string           `json:"payment_note,omitempty"`
}

type ContactRequest struct {
	Nome     string `json:"nome" binding:"required,min=2,max=100"`
	Assunto  string `json:"assunto" binding:"omitempty,max=120"`
	Mensagem string `json:"mensagem" binding:"required,min=3,max=1000"`
	Email    string `json:"email" binding:"omitempty,email,max=150"`
	Telefone string `json:"telefone" binding:"omitempty,max=20"`
}
