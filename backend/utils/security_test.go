package utils

import (
	"testing"
)

// TestHashPassword testa se a senha é corretamente hasheada
func TestHashPassword(t *testing.T) {
	password := "TestPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Erro ao fazer hash: %v", err)
	}

	if hash == "" {
		t.Errorf("Hash não pode ser vazio")
	}

	if hash == password {
		t.Errorf("Hash não pode ser igual à senha")
	}

	// Testar se diferentes senhas geram hashes diferentes
	hash2, _ := HashPassword(password + "2")
	if hash == hash2 {
		t.Errorf("Hashes diferentes devem ser diferentes")
	}
}

// TestVerifyPassword testa se a verificação de senha funciona
func TestVerifyPassword(t *testing.T) {
	password := "TestPassword123!"
	hash, _ := HashPassword(password)

	// Teste com senha correta
	if !VerifyPassword(hash, password) {
		t.Errorf("VerifyPassword deveria retornar true com senha correta")
	}

	// Teste com senha errada
	if VerifyPassword(hash, "WrongPassword") {
		t.Errorf("VerifyPassword deveria retornar false com senha errada")
	}

	// Teste com string vazia
	if VerifyPassword(hash, "") {
		t.Errorf("VerifyPassword deveria retornar false com string vazia")
	}
}

// TestValidateEmail testa validação de email
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"user.name@example.co.uk", true},
		{"user+tag@example.com", true},
		{"invalid.email", false},
		{"@example.com", false},
		{"user@", false},
		{"user@@example.com", false},
		{"user@.com", false},
		{"", false},
	}

	for _, test := range tests {
		result := ValidateEmail(test.email)
		if result != test.valid {
			t.Errorf("ValidateEmail(%s) = %v, esperado %v", test.email, result, test.valid)
		}
	}
}

// TestValidateCPF testa validação de CPF
func TestValidateCPF(t *testing.T) {
	tests := []struct {
		cpf   string
		valid bool
	}{
		{"11144477735", true},    // CPF válido
		{"111.444.777-35", true}, // CPF com formatação
		{"12345678901", false},   // CPF com dígitos verificadores errados
		{"11111111111", false},   // CPF com todos os dígitos iguais
		{"123", false},           // CPF muito curto
		{"", false},              // CPF vazio
		{"abcdefghijk", false},   // CPF com letras
	}

	for _, test := range tests {
		result := ValidateCPF(test.cpf)
		if result != test.valid {
			t.Errorf("ValidateCPF(%s) = %v, esperado %v", test.cpf, result, test.valid)
		}
	}
}

// TestValidatePassword testa validação de força de senha
func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"ValidPass123!", true},
		{"short", false},            // Muito curta
		{"nouppercase123!", false},  // Sem maiúscula
		{"NOLOWERCASE123!", false},  // Sem minúscula
		{"NoNumbers!", false},       // Sem número
		{"NoSpecial123", false},     // Sem caractere especial
		{"ValidPassword123", false}, // Sem caractere especial
		{"", false},                 // Vazia
	}

	for _, test := range tests {
		err := ValidatePassword(test.password)
		isValid := err == nil
		if isValid != test.valid {
			t.Errorf("ValidatePassword(%s): isValid=%v, esperado %v, erro=%v",
				test.password, isValid, test.valid, err)
		}
	}
}

// TestSanitizeString testa sanitização de strings
func TestSanitizeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "Hello World"},
		{"Hello\x00World", "HelloWorld"},                                  // Remove caractere nulo
		{"  Spaced  ", "Spaced"},                                          // Remove espaços extras
		{"Normal\nNewline", "Normal\nNewline"},                            // Mantém newline válido
		{"HTML<script>alert('xss')</script>", "HTMLscriptalertxssscript"}, // Remove tags
	}

	for _, test := range tests {
		result := SanitizeString(test.input)
		if result != test.expected {
			t.Errorf("SanitizeString(%q) = %q, esperado %q", test.input, result, test.expected)
		}
	}
}

// TestTruncateString testa truncagem de strings
func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"Hello World", 5, "Hello"},
		{"Hi", 5, "Hi"},
		{"Exactly Five", 5, "Exact"},
		{"", 5, ""},
	}

	for _, test := range tests {
		result := TruncateString(test.input, test.maxLen)
		if result != test.expected {
			t.Errorf("TruncateString(%q, %d) = %q, esperado %q",
				test.input, test.maxLen, result, test.expected)
		}
	}
}

// TestGenerateMatricula testa geração de matrícula
func TestGenerateMatricula(t *testing.T) {
	mat1 := GenerateMatricula()
	mat2 := GenerateMatricula()

	if mat1 == "" {
		t.Errorf("Matrícula não pode ser vazia")
	}

	if mat1 == mat2 {
		t.Errorf("Matrículas geradas consecutivamente não deveriam ser iguais")
	}

	if len(mat1) < 3 {
		t.Errorf("Matrícula deve ter pelo menos 3 caracteres")
	}
}

// BenchmarkHashPassword faz benchmark de hash de senha
func BenchmarkHashPassword(b *testing.B) {
	password := "TestPassword123!"
	for i := 0; i < b.N; i++ {
		HashPassword(password)
	}
}

// BenchmarkVerifyPassword faz benchmark de verificação de senha
func BenchmarkVerifyPassword(b *testing.B) {
	password := "TestPassword123!"
	hash, _ := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyPassword(hash, password)
	}
}

// BenchmarkValidateEmail faz benchmark de validação de email
func BenchmarkValidateEmail(b *testing.B) {
	email := "user@example.com"
	for i := 0; i < b.N; i++ {
		ValidateEmail(email)
	}
}
