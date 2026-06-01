package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var matriculaCounter uint64

// HashPassword gera um hash bcrypt da senha (proteção contra A02 - Cryptographic Failures)
func HashPassword(password string) (string, error) {
	// Usar custo 12 para um bom balance entre segurança e performance
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword compara uma senha com seu hash
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateJWT gera um token JWT (proteção contra A07 - Authentication Failures)
func GenerateJWT(membroID uint, email string, jwtSecret string, expiration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", membroID), // subject (membro ID)
		"email": email,
		"iat":   time.Now().Unix(),                 // issued at
		"exp":   time.Now().Add(expiration).Unix(), // expiration
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseJWT faz parse e valida um token JWT
func ParseJWT(tokenString, jwtSecret string) (jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar o método de assinatura (proteção contra A05 - Security Misconfiguration)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}

	return claims, nil
}

// ValidateEmail valida o formato de um email
func ValidateEmail(email string) bool {
	email = NormalizeEmail(email)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateCPF faz validação básica do CPF (proteção contra A04 - Insecure Design)
func ValidateCPF(cpf string) bool {
	// Remove caracteres especiais
	cpf = DigitsOnly(cpf)

	// CPF deve ter 11 dígitos
	if len(cpf) != 11 {
		return false
	}

	// Verificar se todos os dígitos são iguais (CPF inválido)
	allSame := true
	for i := 1; i < len(cpf); i++ {
		if cpf[i] != cpf[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	// Calcular primeiro dígito verificador
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(cpf[i]-'0') * (10 - i)
	}
	remainder := sum % 11
	var digit1 int
	if remainder < 2 {
		digit1 = 0
	} else {
		digit1 = 11 - remainder
	}

	if int(cpf[9]-'0') != digit1 {
		return false
	}

	// Calcular segundo dígito verificador
	sum = 0
	for i := 0; i < 10; i++ {
		sum += int(cpf[i]-'0') * (11 - i)
	}
	remainder = sum % 11
	var digit2 int
	if remainder < 2 {
		digit2 = 0
	} else {
		digit2 = 11 - remainder
	}

	if int(cpf[10]-'0') != digit2 {
		return false
	}

	return true
}

// ValidatePassword valida a força da senha (proteção contra A02 - Cryptographic Failures)
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("senha deve ter pelo menos 8 caracteres")
	}

	hasUppercase := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLowercase := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	if !hasUppercase {
		return errors.New("senha deve conter pelo menos uma letra maiúscula")
	}
	if !hasLowercase {
		return errors.New("senha deve conter pelo menos uma letra minúscula")
	}
	if !hasNumber {
		return errors.New("senha deve conter pelo menos um número")
	}
	if !hasSpecial {
		return errors.New("senha deve conter pelo menos um caractere especial")
	}

	return nil
}

// SanitizeString remove caracteres perigosos de uma string (proteção contra A03 - Injection)
func SanitizeString(input string) string {
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		switch r {
		case '<', '>', '/', '\\', '\'', '"', '(', ')':
			return -1
		}
		return r
	}, input)

	return strings.TrimSpace(sanitized)
}

func SanitizeText(input string, maxLength int) string {
	return TruncateString(SanitizeString(input), maxLength)
}

func NormalizeEmail(email string) string {
	return strings.ToLower(SanitizeText(email, 150))
}

func DigitsOnly(input string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, input)
}

func SanitizePhone(phone string) string {
	return TruncateString(DigitsOnly(phone), 20)
}

func SanitizeSlug(slug string) string {
	slug = strings.ToLower(strings.TrimSpace(slug))
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return -1
	}, slug)
}

func SanitizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Map(func(r rune) rune {
		if r < 32 || r == '<' || r == '>' || r == '\'' || r == '"' {
			return -1
		}
		return r
	}, raw)
	raw = TruncateString(raw, 500)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	return parsed.String()
}

// TruncateString trunca uma string a um máximo de caracteres
func TruncateString(s string, maxLength int) string {
	if maxLength <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) > maxLength {
		return string(runes[:maxLength])
	}
	return s
}

// GenerateMatricula gera um número de matrícula único
func GenerateMatricula() string {
	counter := atomic.AddUint64(&matriculaCounter, 1)
	return fmt.Sprintf("CLB%d%04d", time.Now().UnixNano(), counter)
}

// GenerateCodigoDesconto gera um código de desconto único
func GenerateCodigoDesconto(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano()%10000)
}

func GenerateRandomPassword(length int) (string, error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%&*"
	if length < 12 {
		length = 12
	}
	out := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = alphabet[n.Int64()]
	}
	return string(out), nil
}

func HashCPF(cpf string) string {
	digits := DigitsOnly(cpf)
	sum := sha256.Sum256([]byte(digits))
	return hex.EncodeToString(sum[:])
}
