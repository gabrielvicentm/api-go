-- Seed de testes para a API
-- Arquivo compativel com pgAdmin e SQL puro.
--
-- Antes de rodar em outro ambiente, ajuste a chave abaixo para a mesma
-- usada pela API em DATA_ENCRYPTION_KEY ou JWT_ACCESS_SECRET.
--
-- Importante: se a chave nao bater com a da API, CPF/RG/CNH serao gravados
-- mas a aplicacao nao conseguira descriptografar esses dados.

BEGIN;

CREATE OR REPLACE FUNCTION seed_data_key()
RETURNS text
LANGUAGE sql
AS $$
    SELECT 'dev-local-key'::text;
$$;

-- ============================================================
--  LIMPEZA DOS REGISTROS DE TESTE
-- ============================================================

DELETE FROM notificacoes
WHERE id IN (
    'd0000000-0000-0000-0000-000000000001',
    'd0000000-0000-0000-0000-000000000002',
    'd0000000-0000-0000-0000-000000000003'
);

DELETE FROM ocorrencia_midias
WHERE id IN (
    'c0000000-0000-0000-0000-000000000001',
    'c0000000-0000-0000-0000-000000000002'
);

DELETE FROM ocorrencias
WHERE id IN (
    'b0000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000002',
    'b0000000-0000-0000-0000-000000000003'
);

DELETE FROM abastecimentos
WHERE id IN (
    'a0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000002',
    'a0000000-0000-0000-0000-000000000003',
    'a0000000-0000-0000-0000-000000000004'
);

DELETE FROM viagem_documentos
WHERE id IN (
    '99000000-0000-0000-0000-000000000001',
    '99000000-0000-0000-0000-000000000002',
    '99000000-0000-0000-0000-000000000003'
);

DELETE FROM viagem_historico
WHERE id IN (
    '98000000-0000-0000-0000-000000000001',
    '98000000-0000-0000-0000-000000000002',
    '98000000-0000-0000-0000-000000000003',
    '98000000-0000-0000-0000-000000000004'
);

DELETE FROM viagem_paradas
WHERE id IN (
    '97000000-0000-0000-0000-000000000001',
    '97000000-0000-0000-0000-000000000002',
    '97000000-0000-0000-0000-000000000003'
);

DELETE FROM viagem_finalizacoes
WHERE id IN (
    '96000000-0000-0000-0000-000000000001',
    '96000000-0000-0000-0000-000000000002'
);

DELETE FROM manutencoes
WHERE id IN (
    '95000000-0000-0000-0000-000000000001',
    '95000000-0000-0000-0000-000000000002',
    '95000000-0000-0000-0000-000000000003'
);

DELETE FROM auth_refresh_tokens
WHERE actor_id IN (
    '20000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000003'
);

DELETE FROM viagens
WHERE id IN (
    '80000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000002',
    '80000000-0000-0000-0000-000000000003',
    '80000000-0000-0000-0000-000000000004'
);

DELETE FROM motorista_credenciais
WHERE motorista_id IN (
    '20000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000003'
);

DELETE FROM motoristas
WHERE id IN (
    '20000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000003'
);

DELETE FROM funcionario_folha_mensal
WHERE id IN (
    '50000000-0000-0000-0000-000000000001',
    '50000000-0000-0000-0000-000000000002',
    '50000000-0000-0000-0000-000000000003'
);

DELETE FROM funcionario_controle_ponto
WHERE funcionario_id IN (
    '10000000-0000-0000-0000-000000000001',
    '10000000-0000-0000-0000-000000000002',
    '10000000-0000-0000-0000-000000000003',
    '20000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000003'
);

DELETE FROM funcionarios
WHERE id IN (
    '10000000-0000-0000-0000-000000000001',
    '10000000-0000-0000-0000-000000000002',
    '10000000-0000-0000-0000-000000000003',
    '20000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000003'
);

DELETE FROM veiculos
WHERE id IN (
    '30000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000002',
    '30000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000004'
);

DELETE FROM clientes
WHERE id IN (
    '40000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000002',
    '40000000-0000-0000-0000-000000000003',
    '40000000-0000-0000-0000-000000000004'
);

-- ============================================================
--  FUNCIONARIOS
-- ============================================================

INSERT INTO funcionarios (
    id, nome, cpf, cpf_hash, rg, data_nascimento, telefone, email, cep, endereco,
    complemento, numero, bairro, cidade, estado, cargo, setor, tipo_contrato,
    data_admissao, status, salario_base, tipo_pagamento, valor_hora_extra,
    adicional_noturno, vale_alimentacao, outros_descontos, banco, agencia, conta,
    tipo_conta, chave_pix, observacoes
)
VALUES
(
    '10000000-0000-0000-0000-000000000001',
    'Mariana Costa',
    pgp_sym_encrypt('12345678901', seed_data_key()),
    encode(digest('12345678901', 'sha256'), 'hex'),
    pgp_sym_encrypt('123456789', seed_data_key()),
    DATE '1990-03-14',
    '11987654321',
    'mariana.costa@teste.local',
    '01310930',
    'Rua Vergueiro',
    'Apto 81',
    '1456',
    'Vila Mariana',
    'Sao Paulo',
    'SP',
    'Analista Financeiro',
    'Financeiro',
    'clt',
    DATE '2023-01-10',
    'ativo',
    4200.00,
    'mensal',
    35.00,
    180.00,
    650.00,
    120.00,
    'Banco do Brasil',
    '1234',
    '567890',
    'corrente',
    'mariana.costa@teste.local',
    'Responsavel por faturamento e conciliacao.'
),
(
    '10000000-0000-0000-0000-000000000002',
    'Paulo Henrique',
    pgp_sym_encrypt('23456789012', seed_data_key()),
    encode(digest('23456789012', 'sha256'), 'hex'),
    pgp_sym_encrypt('234567890', seed_data_key()),
    DATE '1987-11-02',
    '21999887766',
    'paulo.henrique@teste.local',
    '20040002',
    'Avenida Presidente Vargas',
    'Sala 402',
    '300',
    'Centro',
    'Rio de Janeiro',
    'RJ',
    'Coordenador Operacional',
    'Operacao',
    'clt',
    DATE '2022-06-01',
    'ativo',
    5800.00,
    'mensal',
    42.00,
    220.00,
    700.00,
    180.00,
    'Itau',
    '4321',
    '998877',
    'corrente',
    '21999887766',
    'Coordena escalas e acompanhamento de viagens.'
),
(
    '10000000-0000-0000-0000-000000000003',
    'Luciana Alves',
    pgp_sym_encrypt('34567890123', seed_data_key()),
    encode(digest('34567890123', 'sha256'), 'hex'),
    pgp_sym_encrypt('345678901', seed_data_key()),
    DATE '1994-08-19',
    '31997766554',
    'luciana.alves@teste.local',
    '30140071',
    'Rua da Bahia',
    '',
    '900',
    'Centro',
    'Belo Horizonte',
    'MG',
    'Assistente Administrativo',
    'Administrativo',
    'clt',
    DATE '2024-02-05',
    'ferias',
    3100.00,
    'mensal',
    28.00,
    120.00,
    500.00,
    90.00,
    'Caixa',
    '1111',
    '223344',
    'poupanca',
    '31997766554',
    'Em ferias durante o periodo atual.'
),
(
    '20000000-0000-0000-0000-000000000001',
    'Joao Pedro Lima',
    pgp_sym_encrypt('45678901234', seed_data_key()),
    encode(digest('45678901234', 'sha256'), 'hex'),
    pgp_sym_encrypt('456789012', seed_data_key()),
    DATE '1985-06-23',
    '41991112233',
    'joao.lima@teste.local',
    '80010000',
    'Rua Marechal Deodoro',
    '',
    '120',
    'Centro',
    'Curitiba',
    'PR',
    'Motorista',
    'Operacao',
    'clt',
    DATE '2021-03-15',
    'ativo',
    3900.00,
    'mensal',
    30.00,
    160.00,
    600.00,
    75.00,
    'Bradesco',
    '5566',
    '123456',
    'salario',
    '45678901234',
    'Motorista com foco em viagens interestaduais.'
),
(
    '20000000-0000-0000-0000-000000000002',
    'Carlos Eduardo Souza',
    pgp_sym_encrypt('56789012345', seed_data_key()),
    encode(digest('56789012345', 'sha256'), 'hex'),
    pgp_sym_encrypt('567890123', seed_data_key()),
    DATE '1979-12-08',
    '51992223344',
    'carlos.souza@teste.local',
    '90010000',
    'Avenida Farrapos',
    'Fundos',
    '455',
    'Floresta',
    'Porto Alegre',
    'RS',
    'Motorista',
    'Operacao',
    'clt',
    DATE '2020-08-10',
    'ativo',
    4100.00,
    'mensal',
    32.00,
    170.00,
    620.00,
    80.00,
    'Santander',
    '7788',
    '654321',
    'salario',
    '51992223344',
    'Motorista veterano com atuacao em carga refrigerada.'
),
(
    '20000000-0000-0000-0000-000000000003',
    'Rafael Martins',
    pgp_sym_encrypt('67890123456', seed_data_key()),
    encode(digest('67890123456', 'sha256'), 'hex'),
    pgp_sym_encrypt('678901234', seed_data_key()),
    DATE '1992-09-30',
    '62993334455',
    'rafael.martins@teste.local',
    '74000000',
    'Rua 7',
    '',
    '88',
    'Setor Central',
    'Goiania',
    'GO',
    'Motorista',
    'Operacao',
    'clt',
    DATE '2024-01-12',
    'afastado',
    3600.00,
    'mensal',
    26.00,
    140.00,
    550.00,
    60.00,
    'Sicredi',
    '9900',
    '112233',
    'pix',
    '67890123456',
    'Afastado temporariamente para recuperacao medica.'
);

INSERT INTO funcionario_controle_ponto (
    funcionario_id, horario_entrada, horario_saida, horario_almoco, horas_extras, faltas, atestados
)
VALUES
('10000000-0000-0000-0000-000000000001', '08:00', '17:30', '12:00', 6.50, 0, 0),
('10000000-0000-0000-0000-000000000002', '08:30', '18:00', '12:30', 9.00, 1, 0),
('10000000-0000-0000-0000-000000000003', '09:00', '18:00', '13:00', 2.00, 0, 1),
('20000000-0000-0000-0000-000000000001', '06:00', '15:00', '11:00', 12.00, 0, 0),
('20000000-0000-0000-0000-000000000002', '05:30', '14:30', '10:30', 8.00, 0, 0),
('20000000-0000-0000-0000-000000000003', '07:00', '16:00', '12:00', 1.50, 2, 1);

INSERT INTO funcionario_folha_mensal (
    id, funcionario_id, competencia, salario_base_snapshot, valor_hora_extra_snapshot,
    vale_alimentacao_snapshot, outros_descontos_snapshot, dias_faltas, dias_atestado,
    dias_ferias, dias_afastamento, horas_extras_50, horas_extras_100,
    horas_adicional_noturno, bonus, comissoes, outros_proventos, adiantamentos,
    desconto_inss, desconto_irrf, desconto_vale_transporte, descontos_manuais,
    observacoes, status, pago_em
)
VALUES
(
    '50000000-0000-0000-0000-000000000001',
    '10000000-0000-0000-0000-000000000001',
    DATE '2026-04-01',
    4200.00, 35.00, 650.00, 120.00, 0, 0, 0, 0, 6.50, 0, 0, 250.00, 0, 0, 0,
    420.00, 95.00, 252.00, 120.00,
    'Folha regular do financeiro.',
    'paga',
    TIMESTAMPTZ '2026-05-05 10:00:00-03'
),
(
    '50000000-0000-0000-0000-000000000002',
    '10000000-0000-0000-0000-000000000002',
    DATE '2026-04-01',
    5800.00, 42.00, 700.00, 180.00, 1, 0, 0, 0, 9.00, 0, 0, 300.00, 450.00, 0, 0,
    580.00, 180.00, 348.00, 180.00,
    'Comissao operacional incluida.',
    'fechada',
    NULL
),
(
    '50000000-0000-0000-0000-000000000003',
    '20000000-0000-0000-0000-000000000001',
    DATE '2026-04-01',
    3900.00, 30.00, 600.00, 75.00, 0, 0, 0, 0, 10.00, 2.00, 6.00, 0, 850.00, 0, 500.00,
    390.00, 70.00, 234.00, 75.00,
    'Motorista com comissao por fretes concluidos.',
    'aberta',
    NULL
);

-- ============================================================
--  MOTORISTAS
-- ============================================================

INSERT INTO motoristas (
    id, numero_cnh, numero_cnh_hash, tipo_cnh, validade_cnh, foto_url, observacoes
)
VALUES
(
    '20000000-0000-0000-0000-000000000001',
    pgp_sym_encrypt('12345678901', seed_data_key()),
    encode(digest('12345678901', 'sha256'), 'hex'),
    'E',
    DATE '2028-07-10',
    'https://cdn.teste.local/motoristas/joao-lima.jpg',
    'Apto para cargas gerais e viagens longas.'
),
(
    '20000000-0000-0000-0000-000000000002',
    pgp_sym_encrypt('23456789012', seed_data_key()),
    encode(digest('23456789012', 'sha256'), 'hex'),
    'E',
    DATE '2027-12-18',
    'https://cdn.teste.local/motoristas/carlos-souza.jpg',
    'Especialista em operacao frigorificada.'
),
(
    '20000000-0000-0000-0000-000000000003',
    pgp_sym_encrypt('34567890123', seed_data_key()),
    encode(digest('34567890123', 'sha256'), 'hex'),
    'D',
    DATE '2026-11-02',
    'https://cdn.teste.local/motoristas/rafael-martins.jpg',
    'Registro mantido para cenarios de afastamento.'
);

INSERT INTO motorista_credenciais (
    id, motorista_id, senha_hash, deve_trocar_senha, ativo, ultimo_acesso
)
VALUES
(
    '21000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    crypt('Motor@123', gen_salt('bf', 10)),
    FALSE,
    TRUE,
    TIMESTAMPTZ '2026-05-10 18:10:00-03'
),
(
    '21000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000002',
    crypt('Motor@123', gen_salt('bf', 10)),
    FALSE,
    TRUE,
    TIMESTAMPTZ '2026-05-10 19:40:00-03'
),
(
    '21000000-0000-0000-0000-000000000003',
    '20000000-0000-0000-0000-000000000003',
    crypt('Motor@123', gen_salt('bf', 10)),
    TRUE,
    FALSE,
    NULL
);

-- ============================================================
--  VEICULOS
-- ============================================================

INSERT INTO veiculos (
    id, placa, modelo, marca, ano, tipo, capacidade_carga_kg, renavam, km_atual,
    status, vencimento_seguro, vencimento_licenciamento, vencimento_ipva,
    seguradora, numero_apolice, observacoes
)
VALUES
(
    '30000000-0000-0000-0000-000000000001',
    'ABC1D23',
    'FH 540',
    'Volvo',
    2022,
    'carreta',
    30000.00,
    '12345678901',
    152340.00,
    'em_uso',
    DATE '2026-12-15',
    DATE '2026-08-31',
    DATE '2026-09-30',
    'Porto Seguro',
    'APL-0001',
    'Veiculo principal para cargas secas.'
),
(
    '30000000-0000-0000-0000-000000000002',
    'EFG4H56',
    'R 450',
    'Scania',
    2021,
    'bitruck',
    22000.00,
    '10987654321',
    98750.00,
    'disponivel',
    DATE '2026-11-10',
    DATE '2026-07-15',
    DATE '2026-07-31',
    'Tokio Marine',
    'APL-0002',
    'Reservado para rotas de medio porte.'
),
(
    '30000000-0000-0000-0000-000000000003',
    'IJK7L89',
    'Atego 2426',
    'Mercedes-Benz',
    2020,
    'truck',
    18000.00,
    '11223344556',
    204120.00,
    'manutencao',
    DATE '2026-10-01',
    DATE '2026-06-30',
    DATE '2026-06-30',
    'Allianz',
    'APL-0003',
    'Em revisao corretiva de freios.'
),
(
    '30000000-0000-0000-0000-000000000004',
    'MNO0P12',
    'Daily 70C17',
    'Iveco',
    2023,
    'vuc',
    3500.00,
    '66554433221',
    45120.00,
    'disponivel',
    DATE '2027-01-20',
    DATE '2026-12-20',
    DATE '2026-12-31',
    'Mapfre',
    'APL-0004',
    'Utilizado para entregas urbanas e apoio.'
);

-- ============================================================
--  CLIENTES
-- ============================================================

INSERT INTO clientes (id, nome, cpf_cnpj, telefone, email)
VALUES
(
    '40000000-0000-0000-0000-000000000001',
    'Comercial Atlas LTDA',
    '12.345.678/0001-90',
    '1130304040',
    'logistica@atlas.teste'
),
(
    '40000000-0000-0000-0000-000000000002',
    'Frigorifico Serra Azul SA',
    '98.765.432/0001-10',
    '5133332222',
    'expedicao@serraazul.teste'
),
(
    '40000000-0000-0000-0000-000000000003',
    'Distribuidora Horizonte',
    '45.678.901/0001-55',
    '3135009090',
    'compras@horizonte.teste'
),
(
    '40000000-0000-0000-0000-000000000004',
    'Mercado Nova Era',
    '22.111.333/0001-44',
    '6230302020',
    'recebimento@novaera.teste'
);

-- ============================================================
--  MANUTENCOES
-- ============================================================

INSERT INTO manutencoes (
    id, veiculo_id, tipo, status, descricao, oficina, km_na_manutencao,
    km_proxima_manutencao, data_agendada, data_conclusao, custo, observacoes
)
VALUES
(
    '95000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000003',
    'corretiva',
    'em_andamento',
    'Troca de discos e pastilhas de freio.',
    'Oficina Freio Forte',
    204000.00,
    214000.00,
    DATE '2026-05-09',
    NULL,
    4800.00,
    'Veiculo temporariamente indisponivel.'
),
(
    '95000000-0000-0000-0000-000000000002',
    '30000000-0000-0000-0000-000000000001',
    'preventiva',
    'agendada',
    'Revisao preventiva completa antes de viagem longa.',
    'Oficina Rodomax',
    153000.00,
    163000.00,
    DATE '2026-05-20',
    NULL,
    2300.00,
    'Agendamento feito pelo operacional.'
),
(
    '95000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000004',
    'revisao',
    'concluida',
    'Revisao dos 45 mil km.',
    'Auto Center Sul',
    45000.00,
    55000.00,
    DATE '2026-05-02',
    DATE '2026-05-02',
    980.00,
    'Servico concluido sem apontamentos.'
);

-- ============================================================
--  VIAGENS
-- ============================================================

INSERT INTO viagens (
    id, motorista_id, veiculo_id, cliente_id, origem_cidade, origem_uf,
    destino_cidade, destino_uf, data_saida, data_chegada_prevista,
    data_chegada_real, distancia_km, tipo_carga_id, peso_carga_kg,
    valor_frete, km_inicial, km_final, status, observacoes
)
VALUES
(
    '80000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000001',
    '40000000-0000-0000-0000-000000000001',
    'Sao Paulo',
    'SP',
    'Campinas',
    'SP',
    TIMESTAMPTZ '2026-05-11 06:00:00-03',
    TIMESTAMPTZ '2026-05-11 10:30:00-03',
    NULL,
    98.00,
    (SELECT id FROM tipos_carga WHERE nome = 'Carga Geral' LIMIT 1),
    12000.00,
    8500.00,
    152340.00,
    NULL,
    'em_andamento',
    'Entrega expressa de material seco.'
),
(
    '80000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000002',
    '30000000-0000-0000-0000-000000000002',
    '40000000-0000-0000-0000-000000000002',
    'Porto Alegre',
    'RS',
    'Caxias do Sul',
    'RS',
    TIMESTAMPTZ '2026-05-12 05:30:00-03',
    TIMESTAMPTZ '2026-05-12 09:30:00-03',
    NULL,
    127.00,
    (SELECT id FROM tipos_carga WHERE nome = 'Frigorificada' LIMIT 1),
    15000.00,
    12600.00,
    98750.00,
    NULL,
    'pendente',
    'Carga refrigerada com janela fixa de recebimento.'
),
(
    '80000000-0000-0000-0000-000000000003',
    '20000000-0000-0000-0000-000000000002',
    '30000000-0000-0000-0000-000000000004',
    '40000000-0000-0000-0000-000000000004',
    'Goiania',
    'GO',
    'Brasilia',
    'DF',
    TIMESTAMPTZ '2026-05-08 07:15:00-03',
    TIMESTAMPTZ '2026-05-08 10:45:00-03',
    TIMESTAMPTZ '2026-05-08 10:20:00-03',
    209.00,
    (SELECT id FROM tipos_carga WHERE nome = 'Carga Geral' LIMIT 1),
    2800.00,
    4200.00,
    44900.00,
    45120.00,
    'concluida',
    'Entrega urbana concluida antes do horario previsto.'
),
(
    '80000000-0000-0000-0000-000000000004',
    '20000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000002',
    '40000000-0000-0000-0000-000000000003',
    'Belo Horizonte',
    'MG',
    'Vitoria',
    'ES',
    TIMESTAMPTZ '2026-05-05 04:45:00-03',
    TIMESTAMPTZ '2026-05-05 14:00:00-03',
    NULL,
    524.00,
    (SELECT id FROM tipos_carga WHERE nome = 'Conteinerizada' LIMIT 1),
    17000.00,
    18900.00,
    98200.00,
    NULL,
    'cancelada',
    'Viagem cancelada por indisponibilidade de coleta na origem.'
);

-- ============================================================
--  ITENS COMPLEMENTARES DE VIAGEM
-- ============================================================

INSERT INTO viagem_documentos (id, viagem_id, nome, tipo, url, tamanho_bytes)
VALUES
(
    '99000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000001',
    'cte-viagem-001.pdf',
    'pdf',
    'https://cdn.teste.local/documentos/cte-001.pdf',
    248320
),
(
    '99000000-0000-0000-0000-000000000002',
    '80000000-0000-0000-0000-000000000001',
    'nf-viagem-001.xml',
    'xml',
    'https://cdn.teste.local/documentos/nf-001.xml',
    18240
),
(
    '99000000-0000-0000-0000-000000000003',
    '80000000-0000-0000-0000-000000000003',
    'comprovante-entrega-003.pdf',
    'pdf',
    'https://cdn.teste.local/documentos/comprovante-003.pdf',
    93440
);

INSERT INTO viagem_historico (
    id, viagem_id, usuario_tipo, usuario_id, campo_alterado, valor_anterior, valor_novo, descricao
)
VALUES
(
    '98000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000001',
    'admin',
    (SELECT id FROM usuarios ORDER BY created_at LIMIT 1),
    'status',
    'pendente',
    'em_andamento',
    'Viagem liberada pela operacao.'
),
(
    '98000000-0000-0000-0000-000000000002',
    '80000000-0000-0000-0000-000000000001',
    'motorista',
    '20000000-0000-0000-0000-000000000001',
    'observacoes',
    '',
    'Saida confirmada no patio.',
    'Motorista registrou inicio da viagem.'
),
(
    '98000000-0000-0000-0000-000000000003',
    '80000000-0000-0000-0000-000000000003',
    'admin',
    (SELECT id FROM usuarios ORDER BY created_at LIMIT 1),
    'status',
    'em_andamento',
    'concluida',
    'Viagem encerrada com comprovante.'
),
(
    '98000000-0000-0000-0000-000000000004',
    '80000000-0000-0000-0000-000000000004',
    'admin',
    (SELECT id FROM usuarios ORDER BY created_at LIMIT 1),
    'status',
    'pendente',
    'cancelada',
    'Cliente solicitou remarcacao da coleta.'
);

INSERT INTO viagem_paradas (id, viagem_id, descricao, latitude, longitude, registrado_em)
VALUES
(
    '97000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000001',
    'Parada para conferencia da amarracao da carga.',
    -23.4167000,
    -46.3833000,
    TIMESTAMPTZ '2026-05-11 07:25:00-03'
),
(
    '97000000-0000-0000-0000-000000000002',
    '80000000-0000-0000-0000-000000000001',
    'Abastecimento no posto parceiro da rodovia.',
    -23.1530000,
    -47.0610000,
    TIMESTAMPTZ '2026-05-11 08:10:00-03'
),
(
    '97000000-0000-0000-0000-000000000003',
    '80000000-0000-0000-0000-000000000003',
    'Chegada ao centro de distribuicao.',
    -15.7938890,
    -47.8827780,
    TIMESTAMPTZ '2026-05-08 10:22:00-03'
);

INSERT INTO viagem_finalizacoes (
    id, viagem_id, km_final, status, observacao_motorista, observacao_admin, solicitado_em, respondido_em
)
VALUES
(
    '96000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000003',
    45120.00,
    'aprovada',
    'Entrega realizada sem divergencias.',
    'Finalizacao aprovada com comprovante anexado.',
    TIMESTAMPTZ '2026-05-08 10:25:00-03',
    TIMESTAMPTZ '2026-05-08 10:40:00-03'
),
(
    '96000000-0000-0000-0000-000000000002',
    '80000000-0000-0000-0000-000000000001',
    152438.00,
    'pendente',
    'Previsao de encerramento ao chegar no cliente.',
    NULL,
    TIMESTAMPTZ '2026-05-11 09:50:00-03',
    NULL
);

-- ============================================================
--  ABASTECIMENTOS
-- ============================================================

INSERT INTO abastecimentos (
    id, viagem_id, veiculo_id, motorista_id, tipo_combustivel, km_atual,
    litros, valor_por_litro, fornecedor, foto_url, registrado_em
)
VALUES
(
    'a0000000-0000-0000-0000-000000000004',
    NULL,
    '30000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    'diesel',
    152330.00,
    75.300,
    6.270,
    'Auto Posto Central',
    'https://cdn.teste.local/abastecimentos/abast-004.jpg',
    TIMESTAMPTZ '2026-05-10 06:10:00-03'
),
(
    'a0000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    'diesel',
    152390.00,
    120.500,
    6.290,
    'Posto Bandeirantes',
    'https://cdn.teste.local/abastecimentos/abast-001.jpg',
    TIMESTAMPTZ '2026-05-11 08:12:00-03'
),
(
    'a0000000-0000-0000-0000-000000000002',
    '80000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000004',
    '20000000-0000-0000-0000-000000000002',
    'diesel',
    45020.00,
    48.000,
    6.150,
    'Rede Planalto',
    'https://cdn.teste.local/abastecimentos/abast-002.jpg',
    TIMESTAMPTZ '2026-05-08 08:05:00-03'
),
(
    'a0000000-0000-0000-0000-000000000003',
    NULL,
    '30000000-0000-0000-0000-000000000002',
    '20000000-0000-0000-0000-000000000002',
    'diesel',
    98620.00,
    90.000,
    6.180,
    'Posto da Serra',
    'https://cdn.teste.local/abastecimentos/abast-003.jpg',
    TIMESTAMPTZ '2026-05-10 17:40:00-03'
);

-- ============================================================
--  OCORRENCIAS
-- ============================================================

INSERT INTO ocorrencias (
    id, viagem_id, veiculo_id, motorista_id, tipo, descricao, audio_url,
    latitude, longitude, registrado_em
)
VALUES
(
    'b0000000-0000-0000-0000-000000000001',
    '80000000-0000-0000-0000-000000000001',
    '30000000-0000-0000-0000-000000000001',
    '20000000-0000-0000-0000-000000000001',
    'atraso',
    'Trecho com congestionamento na chegada a Campinas.',
    'https://cdn.teste.local/ocorrencias/audio-001.mp3',
    -22.9055600,
    -47.0608300,
    TIMESTAMPTZ '2026-05-11 09:15:00-03'
),
(
    'b0000000-0000-0000-0000-000000000002',
    '80000000-0000-0000-0000-000000000003',
    '30000000-0000-0000-0000-000000000004',
    '20000000-0000-0000-0000-000000000002',
    'multa',
    'Registro de autuacao por excesso de velocidade.',
    NULL,
    -16.0200000,
    -48.0300000,
    TIMESTAMPTZ '2026-05-08 09:00:00-03'
),
(
    'b0000000-0000-0000-0000-000000000003',
    NULL,
    '30000000-0000-0000-0000-000000000003',
    '20000000-0000-0000-0000-000000000003',
    'pane_mecanica',
    'Veiculo direcionado para manutencao devido a falha no sistema de freio.',
    NULL,
    -16.6869000,
    -49.2648000,
    TIMESTAMPTZ '2026-05-09 07:20:00-03'
);

INSERT INTO ocorrencia_midias (id, ocorrencia_id, tipo, url)
VALUES
(
    'c0000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    'foto',
    'https://cdn.teste.local/ocorrencias/foto-001.jpg'
),
(
    'c0000000-0000-0000-0000-000000000002',
    'b0000000-0000-0000-0000-000000000003',
    'video',
    'https://cdn.teste.local/ocorrencias/video-003.mp4'
);

-- ============================================================
--  NOTIFICACOES
-- ============================================================

INSERT INTO notificacoes (
    id, destinatario_tipo, destinatario_id, origem_tipo, origem_id, titulo,
    mensagem, lida, referencia_tipo, referencia_id, created_at
)
VALUES
(
    'd0000000-0000-0000-0000-000000000001',
    'admin',
    NULL,
    'motorista',
    '20000000-0000-0000-0000-000000000001',
    'Ocorrencia registrada em viagem',
    'Motorista Joao Pedro Lima informou atraso na viagem 80000000-0000-0000-0000-000000000001.',
    FALSE,
    'ocorrencia',
    'b0000000-0000-0000-0000-000000000001',
    TIMESTAMPTZ '2026-05-11 09:16:00-03'
),
(
    'd0000000-0000-0000-0000-000000000002',
    'motorista',
    '20000000-0000-0000-0000-000000000001',
    'sistema',
    NULL,
    'Finalizacao pendente',
    'Sua solicitacao de finalizacao aguarda avaliacao do administrativo.',
    FALSE,
    'viagem_finalizacao',
    '96000000-0000-0000-0000-000000000002',
    TIMESTAMPTZ '2026-05-11 09:55:00-03'
),
(
    'd0000000-0000-0000-0000-000000000003',
    'admin',
    NULL,
    'sistema',
    NULL,
    'Manutencao em andamento',
    'O veiculo IJK7L89 segue indisponivel por manutencao corretiva.',
    TRUE,
    'manutencao',
    '95000000-0000-0000-0000-000000000001',
    TIMESTAMPTZ '2026-05-09 08:00:00-03'
);

COMMIT;

DROP FUNCTION IF EXISTS seed_data_key();
