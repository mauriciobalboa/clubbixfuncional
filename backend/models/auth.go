package models

import "time"

type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required,min=6"`
}

type RegisterRequest struct {
	NomeCompleto string `json:"nome_completo" binding:"required,min=3,max=100"`
	CPF          string `json:"cpf" binding:"required,max=20"`
	Email        string `json:"email" binding:"required,email,max=150"`
	Telefone     string `json:"telefone" binding:"omitempty,max=20"`
	Senha        string `json:"senha" binding:"required,min=8,max=128"`
}

type AuthResponse struct {
	Token          string           `json:"token"`
	Usuario        *Usuario         `json:"usuario"`
	ExpirationTime time.Time        `json:"expiration_time"`
	Session        *FrontendSession `json:"session,omitempty"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type FrontendSession struct {
	Nome     string `json:"nome"`
	Email    string `json:"email"`
	Plano    string `json:"plano,omitempty"`
	MemberID string `json:"memberId"`
	Token    string `json:"token,omitempty"`
}
