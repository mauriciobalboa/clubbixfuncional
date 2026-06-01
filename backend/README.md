# Clubbix Backend - Go

Backend seguro em Go para a plataforma Clubbix, com proteção completa contra OWASP Top 10.

## 🔒 Segurança

Este backend implementa proteção contra os 10 principais riscos de segurança da web:

- **A01:2021 - Broken Access Control**: Validação rigorosa de tokens JWT, autorização por middleware
- **A02:2021 - Cryptographic Failures**: Hash bcrypt com custo 12, senhas fortes obrigatórias
- **A03:2021 - Injection**: ORM GORM previne SQL injection, sanitização de input
- **A04:2021 - Insecure Design**: Validação robusta de CPF, email, formato de dados
- **A05:2021 - Security Misconfiguration**: Headers de segurança, CORS restritivo, HSTS
- **A06:2021 - Vulnerable Components**: Dependências atualizadas, sem bibliotecas desatualizadas
- **A07:2021 - Authentication Failures**: JWT com expiração, lockout após falhas, token validation
- **A08:2021 - Software/Data Integrity**: Validações de entrada em todas as camadas
- **A09:2021 - Logging & Monitoring**: Logs estruturados, sem exposição de dados sensíveis
- **A10:2021 - SSRF**: Isolamento de recursos, uso seguro de URLs

## 📋 Requisitos

- Go 1.21+
- SQLite 3 (incluído no projeto)

## 🚀 Instalação e Setup

### 1. Clone ou abra o projeto

```bash
cd clubbix/backend
```

> Docker fica centralizado na raiz do projeto. Use `cd clubbix` para comandos com `docker compose`.

### 2. Instale as dependências

```bash
go mod download
go mod tidy
```

### 3. Configure as variáveis de ambiente

Crie um arquivo `.env` baseado em `.env.example`:

```bash
cp .env.example .env
```

### 4. Execute o servidor

```bash
go run main.go
```

O servidor iniciará em `http://localhost:8080`

## Docker

Todos os arquivos Docker ficam na raiz do projeto:

- `Dockerfile`
- `docker-compose.yml`
- `.dockerignore`

Comando principal:

```bash
cd clubbix
docker compose up --build
```

O container expõe backend e frontend HTML em `http://localhost:8080`.

O Docker não usa banco separado: o Compose monta o arquivo local `backend/clubbix.db` dentro do container e define `DB_PATH=/app/clubbix.db`.

## 🛠️ Desenvolvimento

### Compilar para produção

```bash
go build -o clubbix-server main.go
```

### Modo hot-reload (recomendado com Air)

```bash
go install github.com/cosmtrek/air@latest
air
```

Crie um arquivo `.air.toml`:

```toml
[build]
cmd = "go build -o ./bin/clubbix main.go"
bin = "bin/clubbix"
```

## 📡 Endpoints

### Autenticação (Público)

- `POST /api/auth/login` - Login
- `POST /api/auth/register` - Registro
- `POST /api/auth/esqueci-senha` - Recuperação de senha
- `POST /auth/redefinir-senha` - Redefinir senha

### Autenticação (Protegido)

- `GET /api/auth/me` - Dados do usuário
- `POST /api/auth/logout` - Logout

### Planos (Público)

- `GET /api/planos` - Listar planos
- `GET /api/planos/:id` - Detalhes do plano

### Parceiros (Público)

- `GET /api/parceiros?categoria=&busca=` - Listar parceiros com filtros
- `GET /api/parceiros/:id` - Detalhes do parceiro
- `GET /api/categorias` - Listar categorias

### Assinatura (Protegido)

- `POST /api/assinatura` - Criar assinatura
- `GET /api/assinatura/me` - Assinatura do usuário

### Contato (Público)

- `POST /api/contato` - Enviar mensagem

### FAQ (Público)

- `GET /api/faq?q=` - Buscar FAQs

### Admin

- `POST /api/admin/planos` - Criar plano
- `PUT /api/admin/planos/:id` - Atualizar plano
- `DELETE /api/admin/planos/:id` - Deletar plano
- `POST /api/admin/parceiros` - Criar parceiro
- `PUT /api/admin/parceiros/:id` - Atualizar parceiro
- `PATCH /api/admin/parceiros/:id/toggle` - Ativar/desativar parceiro

## 🗄️ Banco de Dados

O projeto usa SQLite com GORM ORM. Migrations são executadas automaticamente na inicialização.

### Estrutura de Tabelas

- `membros` - Usuários/membros
- `planos` - Planos de associação
- `assinaturas` - Registro de assinaturas
- `parceiros` - Parceiros/beneficiários
- `categorias` - Categorias de parceiros
- `mensagens_contato` - Mensagens de contato
- `faq` - Perguntas frequentes

## 📚 Estrutura do Projeto

```
clubbix/backend/
├── config/           # Configurações (JWT, DB, ports)
├── database/         # Inicialização e seed do banco
├── handlers/         # Handlers HTTP (lógica de rotas)
├── middleware/       # Middlewares (auth, security, CORS)
├── models/           # Modelos GORM e DTOs
├── utils/            # Utilitários (hash, JWT, validações)
├── main.go           # Arquivo principal
├── go.mod            # Dependências Go
└── README.md         # Este arquivo
```

## 🔐 Variáveis de Ambiente

```env
# Server
PORT=8080
ENVIRONMENT=development

# JWT
JWT_SECRET=sua-chave-secreta-muito-segura-aqui-mudaar

# Database
DB_PATH=./clubbix.db
```

## 🚨 Proteção contra Ataques

### Rate Limiting
- Limite de 60 requisições por minuto por IP
- Proteção contra DoS e brute force

### Autenticação
- JWT com expiração de 24h
- Lockout automático após 5 tentativas de login falhas
- Hash bcrypt com custo 12

### Validação de Entrada
- Sanitização de strings (remove caracteres de controle)
- Validação de email, CPF, tamanho de corpo
- Prevenção de SQL injection via ORM
- Limite de tamanho de requisições (10 MB)

### Headers de Segurança
- X-Frame-Options: DENY (previne clickjacking)
- X-Content-Type-Options: nosniff (previne MIME sniffing)
- Content-Security-Policy (previne XSS)
- HSTS (em produção)

### CORS
- Apenas origens configuradas
- Sem credenciais cruzadas não autorizadas

## 📝 Logging

Logs são registrados em console com timestamp, método HTTP, path, status e IP.

**Dados sensíveis nunca são logados**:
- Senhas
- Tokens JWT
- Dados financeiros

## 🧪 Testando Endpoints

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@clubbix.com.br","senha":"clubbix123"}'
```

### Listar Planos
```bash
curl http://localhost:8080/api/planos
```

### Acessar rota protegida
```bash
curl -H "Authorization: Bearer <seu-token-jwt>" \
  http://localhost:8080/api/auth/me
```

## 📄 Licença

Uso interno apenas - Clubbix
