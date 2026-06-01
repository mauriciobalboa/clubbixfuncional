package database

import (
	"log"
	"time"

	"clubbix/backend/config"
	"clubbix/backend/models"
	"clubbix/backend/utils"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) *gorm.DB {
	dbPath := cfg.DBPath
	environment := cfg.Environment
	logLevel := logger.Info
	if environment == "production" {
		logLevel = logger.Error
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logLevel),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}

	db.Exec("PRAGMA foreign_keys = ON")

	err = db.AutoMigrate(
		&models.Usuario{},
		&models.Categoria{},
		&models.Plano{},
		&models.Parceiro{},
		&models.Assinatura{},
		&models.Pagamento{},
		&models.Beneficio{},
		&models.Curso{},
		&models.Responsavel{},
		&models.UsuarioCurso{},
		&models.LogAuditoria{},
	)
	if err != nil {
		log.Fatalf("Erro ao fazer migration: %v", err)
	}

	log.Println("Banco de dados inicializado com sucesso")
	seedDatabase(db, cfg)

	return db
}

func nowDate() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func seedDatabase(db *gorm.DB, cfg *config.Config) {
	log.Println("Sincronizando dados iniciais...")

	// Admin — só cria se as env vars estiverem definidas
	if cfg.AdminEmail != "" && cfg.AdminPassword != "" {
		adminPassword, _ := utils.HashPassword(cfg.AdminPassword)
		admin := models.Usuario{
			NomeCompleto: "Administrador Clubbix",
			CPFHash:      utils.HashCPF(cfg.AdminCPF),
			Email:        cfg.AdminEmail,
			Telefone:     "",
			SenhaHash:    adminPassword,
			TipoUsuario:  models.TipoUsuarioAdmin,
			StatusConta:  models.StatusContaAtiva,
		}
		var existingAdmin models.Usuario
		if err := db.Where("email = ?", admin.Email).First(&existingAdmin).Error; err != nil {
			db.Create(&admin)
		}
	}

	// Demo — só cria se as env vars estiverem definidas
	if cfg.DemoEmail != "" && cfg.DemoPassword != "" {
		demoPassword, _ := utils.HashPassword(cfg.DemoPassword)
		demo := models.Usuario{
			NomeCompleto: "Visitante Clubbix",
			CPFHash:      utils.HashCPF(cfg.DemoCPF),
			Email:        cfg.DemoEmail,
			Telefone:     "",
			SenhaHash:    demoPassword,
			TipoUsuario:  models.TipoUsuarioAssociado,
			StatusConta:  models.StatusContaAtiva,
		}
		var existingDemo models.Usuario
		if err := db.Where("email = ?", demo.Email).First(&existingDemo).Error; err != nil {
			db.Create(&demo)
		}
	}

	planos := []models.Plano{
		{NomePlano: "Basico", Descricao: "Acesso essencial ao clube e aos beneficios", Valor: 409.99, DuracaoMeses: 12, DuracaoTipo: models.DuracaoTipoMeses, Ativo: true},
		{NomePlano: "Plus", Descricao: "Acesso estendido ao clube com melhor economia por periodo", Valor: 699.99, DuracaoMeses: 24, DuracaoTipo: models.DuracaoTipoMeses, Ativo: true},
		{NomePlano: "Max", Descricao: "Acesso completo ao clube pelo maior periodo disponivel", Valor: 1199.88, DuracaoMeses: 48, DuracaoTipo: models.DuracaoTipoMeses, Ativo: true},
	}
	for _, plano := range planos {
		var existing models.Plano
		if err := db.Where("nome_plano = ?", plano.NomePlano).First(&existing).Error; err != nil {
			db.Create(&plano)
			continue
		}
		db.Model(&existing).Updates(map[string]interface{}{
			"descricao":     plano.Descricao,
			"valor":         plano.Valor,
			"duracao_meses": plano.DuracaoMeses,
			"duracao_tipo":  plano.DuracaoTipo,
			"ativo":         plano.Ativo,
		})
	}

	categorias := []models.Categoria{
		{Slug: "restaurantes", Nome: "Restaurantes", Icone: "🍔", Ordem: 10, Ativo: true},
		{Slug: "saude", Nome: "Saúde", Icone: "🩺", Ordem: 20, Ativo: true},
		{Slug: "educacao", Nome: "Educação", Icone: "📚", Ordem: 30, Ativo: true},
		{Slug: "beleza", Nome: "Beleza", Icone: "💇", Ordem: 40, Ativo: true},
		{Slug: "academias", Nome: "Academias", Icone: "🏋️", Ordem: 50, Ativo: true},
		{Slug: "moda", Nome: "Moda", Icone: "🛍️", Ordem: 60, Ativo: true},
		{Slug: "lazer", Nome: "Lazer", Icone: "🎬", Ordem: 70, Ativo: true},
		{Slug: "farmacias", Nome: "Farmácias", Icone: "💊", Ordem: 80, Ativo: true},
		{Slug: "pet", Nome: "Pet", Icone: "🐾", Ordem: 90, Ativo: true},
		{Slug: "turismo", Nome: "Turismo", Icone: "✈️", Ordem: 100, Ativo: true},
		{Slug: "automotivo", Nome: "Automotivo", Icone: "🚗", Ordem: 110, Ativo: true},
		{Slug: "servicos", Nome: "Serviços", Icone: "🔧", Ordem: 120, Ativo: true},
	}
	for _, categoria := range categorias {
		var existing models.Categoria
		if err := db.Where("slug = ?", categoria.Slug).First(&existing).Error; err != nil {
			db.Create(&categoria)
			continue
		}
		db.Model(&existing).Updates(map[string]interface{}{
			"nome":  categoria.Nome,
			"icone": categoria.Icone,
			"ordem": categoria.Ordem,
			"ativo": categoria.Ativo,
		})
	}

	parceiros := []models.Parceiro{
		{NomeEmpresa: "Cantina Bella Massa", Categoria: "restaurantes", Descricao: "Restaurante parceiro com desconto para membros Clubbix", PercentualDesconto: 25, Endereco: "Centro, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Burger do Chef", Categoria: "restaurantes", Descricao: "Restaurante parceiro com desconto para membros Clubbix", PercentualDesconto: 20, Endereco: "Goiabeiras, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Sabor do Cerrado", Categoria: "restaurantes", Descricao: "Restaurante parceiro com desconto para membros Clubbix", PercentualDesconto: 15, Endereco: "Jd. das Americas, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Pizzaria Forno Real", Categoria: "restaurantes", Descricao: "Pizzaria parceira com desconto para membros Clubbix", PercentualDesconto: 30, Endereco: "Coxipo, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Clinica VidaPlena", Categoria: "saude", Descricao: "Clinica parceira para cuidados de saude", PercentualDesconto: 30, Endereco: "Centro, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Lab Diagnostico+", Categoria: "saude", Descricao: "Laboratorio parceiro com desconto para membros Clubbix", PercentualDesconto: 20, Endereco: "Bosque da Saude, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Odonto Sorriso", Categoria: "saude", Descricao: "Clinica odontologica parceira", PercentualDesconto: 25, Endereco: "CPA, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Idiomas Global", Categoria: "educacao", Descricao: "Escola de idiomas parceira", PercentualDesconto: 35, Endereco: "Centro, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Curso Aprova+", Categoria: "educacao", Descricao: "Curso preparatorio parceiro", PercentualDesconto: 25, Endereco: "Jd. das Americas, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Studio Bella Hair", Categoria: "beleza", Descricao: "Salao parceiro com desconto para membros Clubbix", PercentualDesconto: 30, Endereco: "Goiabeiras, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Espaco Estetica Viva", Categoria: "beleza", Descricao: "Estetica parceira com desconto para membros Clubbix", PercentualDesconto: 20, Endereco: "Centro, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "PowerFit Academia", Categoria: "academias", Descricao: "Academia com condicoes especiais para membros", PercentualDesconto: 40, Endereco: "Jd. das Americas, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Arena Cross Treino", Categoria: "academias", Descricao: "Centro de treinamento parceiro", PercentualDesconto: 25, Endereco: "CPA, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Urban Wear Store", Categoria: "moda", Descricao: "Loja de moda parceira", PercentualDesconto: 15, Endereco: "Centro, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Passo Certo Calcados", Categoria: "moda", Descricao: "Loja de calcados parceira", PercentualDesconto: 20, Endereco: "Varzea Grande-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "CineStar Cinemas", Categoria: "lazer", Descricao: "Cinema parceiro com desconto para membros Clubbix", PercentualDesconto: 30, Endereco: "Shopping Centro, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Boliche Strike", Categoria: "lazer", Descricao: "Boliche parceiro com desconto para membros Clubbix", PercentualDesconto: 25, Endereco: "Coxipo, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Farmacia BemEstar", Categoria: "farmacias", Descricao: "Farmacia parceira com desconto para membros Clubbix", PercentualDesconto: 15, Endereco: "Goiabeiras, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Mundo Pet", Categoria: "pet", Descricao: "Pet shop parceiro com desconto para membros Clubbix", PercentualDesconto: 20, Endereco: "Santa Rosa, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Pousada Aguas Claras", Categoria: "turismo", Descricao: "Pousada parceira com desconto para membros Clubbix", PercentualDesconto: 25, Endereco: "Chapada dos Guimaraes-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "AutoCenter Premium", Categoria: "automotivo", Descricao: "Auto center parceiro com desconto para membros Clubbix", PercentualDesconto: 20, Endereco: "Coxipo, Cuiaba-MT", Status: models.StatusParceiroAtivo},
		{NomeEmpresa: "Brilho Total Lava-Jato", Categoria: "servicos", Descricao: "Lava-jato parceiro com desconto para membros Clubbix", PercentualDesconto: 30, Endereco: "Centro, Cuiaba-MT", Status: models.StatusParceiroAtivo},
	}
	for i := range parceiros {
		var existing models.Parceiro
		if err := db.Where("nome_empresa = ?", parceiros[i].NomeEmpresa).First(&existing).Error; err != nil {
			db.Create(&parceiros[i])
			continue
		}
		parceiros[i].IDParceiro = existing.IDParceiro
		db.Model(&existing).Updates(map[string]interface{}{
			"categoria":           parceiros[i].Categoria,
			"descricao":           parceiros[i].Descricao,
			"percentual_desconto": parceiros[i].PercentualDesconto,
			"endereco":            parceiros[i].Endereco,
			"telefone":            parceiros[i].Telefone,
			"email":               parceiros[i].Email,
			"status":              parceiros[i].Status,
		})
	}

	beneficios := []models.Beneficio{
		{
			IDParceiro:    parceiros[0].IDParceiro,
			Titulo:        "Desconto em refeicoes",
			Descricao:     "Desconto valido para almoco e jantar",
			TipoDesconto:  models.TipoDescontoPercentual,
			ValorDesconto: 15,
			RegrasUso:     "Apresentar carteirinha Clubbix",
			DataInicio:    nowDate(),
			DataFim:       nowDate().AddDate(1, 0, 0),
			Ativo:         true,
		},
		{
			IDParceiro:    parceiros[1].IDParceiro,
			Titulo:        "Desconto em mensalidade",
			Descricao:     "Desconto no plano mensal",
			TipoDesconto:  models.TipoDescontoPercentual,
			ValorDesconto: 20,
			RegrasUso:     "Contrato direto com a academia",
			DataInicio:    nowDate(),
			DataFim:       nowDate().AddDate(1, 0, 0),
			Ativo:         true,
		},
	}
	for _, beneficio := range beneficios {
		var existing models.Beneficio
		if err := db.Where("id_parceiro = ? AND titulo = ?", beneficio.IDParceiro, beneficio.Titulo).First(&existing).Error; err != nil {
			db.Create(&beneficio)
		}
	}

	cursos := []models.Curso{
		{
			NomeCurso:          "Ingles Intensivo",
			Instituicao:        "Clubbix Education",
			IDParceiro:         parceiros[0].IDParceiro,
			PercentualDesconto: 10,
			TipoMatricula:      models.TipoMatriculaNova,
			Ativo:              true,
		},
	}
	for _, curso := range cursos {
		var existing models.Curso
		if err := db.Where("nome_curso = ? AND id_parceiro = ?", curso.NomeCurso, curso.IDParceiro).First(&existing).Error; err != nil {
			db.Create(&curso)
		}
	}

	log.Println("Dados iniciais sincronizados com sucesso")
}
