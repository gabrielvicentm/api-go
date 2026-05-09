package domain

type FolhaPagamentoListFilter struct {
	Search      string
	Status      string
	Competencia string
}

type FolhaPagamentoUpsertRequest struct {
	Competencia               string  `json:"competencia" binding:"required"`
	SalarioBaseSnapshot       float64 `json:"salario_base_snapshot"`
	ValorHoraExtraSnapshot    float64 `json:"valor_hora_extra_snapshot"`
	ValeAlimentacaoSnapshot   float64 `json:"vale_alimentacao_snapshot"`
	OutrosDescontosSnapshot   float64 `json:"outros_descontos_snapshot"`
	DiasFaltas                int     `json:"dias_faltas"`
	DiasAtestado              int     `json:"dias_atestado"`
	DiasFerias                int     `json:"dias_ferias"`
	DiasAfastamento           int     `json:"dias_afastamento"`
	HorasExtras50             float64 `json:"horas_extras_50"`
	HorasExtras100            float64 `json:"horas_extras_100"`
	HorasAdicionalNoturno     float64 `json:"horas_adicional_noturno"`
	Bonus                     float64 `json:"bonus"`
	Comissoes                 float64 `json:"comissoes"`
	OutrosProventos           float64 `json:"outros_proventos"`
	Adiantamentos             float64 `json:"adiantamentos"`
	DescontoINSS              float64 `json:"desconto_inss"`
	DescontoIRRF              float64 `json:"desconto_irrf"`
	DescontoValeTransporte    float64 `json:"desconto_vale_transporte"`
	DescontosManuais          float64 `json:"descontos_manuais"`
	Observacoes               string  `json:"observacoes"`
	Status                    string  `json:"status"`
}

type FolhaPagamentoResumo struct {
	FuncionarioID      string  `json:"funcionario_id"`
	Nome               string  `json:"nome"`
	Cargo              string  `json:"cargo"`
	Setor              string  `json:"setor"`
	StatusFuncionario  string  `json:"status_funcionario"`
	StatusFolha        string  `json:"status_folha"`
	Competencia        string  `json:"competencia"`
	SalarioBase        float64 `json:"salario_base"`
	TotalProventos     float64 `json:"total_proventos"`
	TotalDescontos     float64 `json:"total_descontos"`
	SalarioLiquido     float64 `json:"salario_liquido"`
	DiasFaltas         int     `json:"dias_faltas"`
	DiasFerias         int     `json:"dias_ferias"`
	HorasExtras50      float64 `json:"horas_extras_50"`
	HorasExtras100     float64 `json:"horas_extras_100"`
	RegistroExistente  bool    `json:"registro_existente"`
}

type FolhaPagamentoFuncionario struct {
	ID            string `json:"id"`
	Nome          string `json:"nome"`
	Cargo         string `json:"cargo"`
	Setor         string `json:"setor"`
	Status        string `json:"status"`
	TipoPagamento string `json:"tipo_pagamento"`
	IsMotorista   bool   `json:"is_motorista"`
}

type FolhaPagamentoCompetencia struct {
	Competencia               string  `json:"competencia"`
	SalarioBaseSnapshot       float64 `json:"salario_base_snapshot"`
	ValorHoraExtraSnapshot    float64 `json:"valor_hora_extra_snapshot"`
	ValeAlimentacaoSnapshot   float64 `json:"vale_alimentacao_snapshot"`
	OutrosDescontosSnapshot   float64 `json:"outros_descontos_snapshot"`
	DiasFaltas                int     `json:"dias_faltas"`
	DiasAtestado              int     `json:"dias_atestado"`
	DiasFerias                int     `json:"dias_ferias"`
	DiasAfastamento           int     `json:"dias_afastamento"`
	HorasExtras50             float64 `json:"horas_extras_50"`
	HorasExtras100            float64 `json:"horas_extras_100"`
	HorasAdicionalNoturno     float64 `json:"horas_adicional_noturno"`
	Bonus                     float64 `json:"bonus"`
	Comissoes                 float64 `json:"comissoes"`
	OutrosProventos           float64 `json:"outros_proventos"`
	Adiantamentos             float64 `json:"adiantamentos"`
	DescontoINSS              float64 `json:"desconto_inss"`
	DescontoIRRF              float64 `json:"desconto_irrf"`
	DescontoValeTransporte    float64 `json:"desconto_vale_transporte"`
	DescontosManuais          float64 `json:"descontos_manuais"`
	Observacoes               string  `json:"observacoes"`
	Status                    string  `json:"status"`
}

type FolhaPagamentoCalculo struct {
	SalarioBase             float64 `json:"salario_base"`
	ValorDia                float64 `json:"valor_dia"`
	ValorHora               float64 `json:"valor_hora"`
	ValorHoraExtra50        float64 `json:"valor_hora_extra_50"`
	ValorHoraExtra100       float64 `json:"valor_hora_extra_100"`
	ValorAdicionalNoturno   float64 `json:"valor_adicional_noturno"`
	ValorFerias             float64 `json:"valor_ferias"`
	TercoFerias             float64 `json:"terco_ferias"`
	DescontoFaltas          float64 `json:"desconto_faltas"`
	DescontoAfastamento     float64 `json:"desconto_afastamento"`
	DescontoFerias          float64 `json:"desconto_ferias"`
	TotalProventos          float64 `json:"total_proventos"`
	TotalDescontos          float64 `json:"total_descontos"`
	SalarioLiquido          float64 `json:"salario_liquido"`
}

type FolhaPagamentoDetalhe struct {
	Funcionario       FolhaPagamentoFuncionario  `json:"funcionario"`
	Folha             FolhaPagamentoCompetencia `json:"folha"`
	Calculo           FolhaPagamentoCalculo     `json:"calculo"`
	RegistroExistente bool                      `json:"registro_existente"`
}
