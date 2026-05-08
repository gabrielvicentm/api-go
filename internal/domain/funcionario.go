package domain

import "time"

type FuncionarioListFilter struct {
	Search           string
	Status           string
	Tipo             string
	Page             int
	Limit            int
	IncludeMotorista bool
}

type FuncionarioCreateRequest struct {
	Nome             string  `json:"nome" binding:"required,min=3"`
	CPF              string  `json:"cpf" binding:"required"`
	RG               string  `json:"rg"`
	DataNascimento   string  `json:"data_nascimento"`
	Telefone         string  `json:"telefone"`
	Email            string  `json:"email"`
	CEP              string  `json:"cep"`
	Endereco         string  `json:"endereco"`
	Complemento      string  `json:"complemento"`
	Numero           string  `json:"numero"`
	Bairro           string  `json:"bairro"`
	Cidade           string  `json:"cidade"`
	Estado           string  `json:"estado"`
	Cargo            string  `json:"cargo"`
	Setor            string  `json:"setor"`
	TipoContrato     string  `json:"tipo_contrato"`
	DataAdmissao     string  `json:"data_admissao"`
	DataDemissao     string  `json:"data_demissao"`
	Status           string  `json:"status"`
	SalarioBase      float64 `json:"salario_base"`
	TipoPagamento    string  `json:"tipo_pagamento"`
	ValorHoraExtra   float64 `json:"valor_hora_extra"`
	AdicionalNoturno float64 `json:"adicional_noturno"`
	ValeAlimentacao  float64 `json:"vale_alimentacao"`
	OutrosDescontos  float64 `json:"outros_descontos"`
	Banco            string  `json:"banco"`
	Agencia          string  `json:"agencia"`
	Conta            string  `json:"conta"`
	TipoConta        string  `json:"tipo_conta"`
	ChavePix         string  `json:"chave_pix"`
	HorarioEntrada   string  `json:"horario_entrada"`
	HorarioSaida     string  `json:"horario_saida"`
	HorarioAlmoco    string  `json:"horario_almoco"`
	HorasExtras      float64 `json:"horas_extras"`
	Faltas           int     `json:"faltas"`
	Atestados        int     `json:"atestados"`
	Observacoes      string  `json:"observacoes"`
}

type FuncionarioUpdateRequest = FuncionarioCreateRequest

type FuncionarioStatusUpdateRequest struct {
	Status string `json:"status" binding:"required"`
}

type FuncionarioListItem struct {
	ID           string     `json:"id"`
	Nome         string     `json:"nome"`
	CPF          string     `json:"cpf"`
	Telefone     string     `json:"telefone,omitempty"`
	Email        string     `json:"email,omitempty"`
	Cargo        string     `json:"cargo,omitempty"`
	Setor        string     `json:"setor,omitempty"`
	Status       string     `json:"status"`
	Tipo         string     `json:"tipo"`
	IsMotorista  bool       `json:"is_motorista"`
	DataAdmissao string     `json:"data_admissao,omitempty"`
	SalarioBase  float64    `json:"salario_base"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
}

type FuncionarioDetail struct {
	ID               string     `json:"id"`
	Nome             string     `json:"nome"`
	CPF              string     `json:"cpf"`
	RG               string     `json:"rg,omitempty"`
	DataNascimento   string     `json:"data_nascimento,omitempty"`
	Telefone         string     `json:"telefone,omitempty"`
	Email            string     `json:"email,omitempty"`
	CEP              string     `json:"cep,omitempty"`
	Endereco         string     `json:"endereco,omitempty"`
	Complemento      string     `json:"complemento,omitempty"`
	Numero           string     `json:"numero,omitempty"`
	Bairro           string     `json:"bairro,omitempty"`
	Cidade           string     `json:"cidade,omitempty"`
	Estado           string     `json:"estado,omitempty"`
	Cargo            string     `json:"cargo,omitempty"`
	Setor            string     `json:"setor,omitempty"`
	TipoContrato     string     `json:"tipo_contrato,omitempty"`
	DataAdmissao     string     `json:"data_admissao,omitempty"`
	DataDemissao     string     `json:"data_demissao,omitempty"`
	Status           string     `json:"status"`
	SalarioBase      float64    `json:"salario_base"`
	TipoPagamento    string     `json:"tipo_pagamento,omitempty"`
	ValorHoraExtra   float64    `json:"valor_hora_extra"`
	AdicionalNoturno float64    `json:"adicional_noturno"`
	ValeAlimentacao  float64    `json:"vale_alimentacao"`
	OutrosDescontos  float64    `json:"outros_descontos"`
	Banco            string     `json:"banco,omitempty"`
	Agencia          string     `json:"agencia,omitempty"`
	Conta            string     `json:"conta,omitempty"`
	TipoConta        string     `json:"tipo_conta,omitempty"`
	ChavePix         string     `json:"chave_pix,omitempty"`
	HorarioEntrada   string     `json:"horario_entrada,omitempty"`
	HorarioSaida     string     `json:"horario_saida,omitempty"`
	HorarioAlmoco    string     `json:"horario_almoco,omitempty"`
	HorasExtras      float64    `json:"horas_extras"`
	Faltas           int        `json:"faltas"`
	Atestados        int        `json:"atestados"`
	Observacoes      string     `json:"observacoes,omitempty"`
	Tipo             string     `json:"tipo"`
	IsMotorista      bool       `json:"is_motorista"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
}
