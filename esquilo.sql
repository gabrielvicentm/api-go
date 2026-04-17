-- ============================================================
--  SISTEMA DE TRANSPORTADORA
--  PostgreSQL Schema
-- ============================================================

-- ============================================================
--  EXTENSÕES
-- ============================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "unaccent";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
--  TIPOS ENUMERADOS
-- ============================================================

CREATE TYPE status_viagem     AS ENUM ('pendente', 'em_andamento', 'concluida', 'cancelada');
CREATE TYPE status_veiculo    AS ENUM ('disponivel', 'em_uso', 'manutencao', 'inativo');
CREATE TYPE status_motorista  AS ENUM ('ativo', 'inativo', 'ferias', 'afastado');
CREATE TYPE tipo_ocorrencia   AS ENUM (
    'acidente', 'multa', 'pane_mecanica', 'pane_eletrica',
    'furto', 'avaria_carga', 'atraso', 'outro'
);
CREATE TYPE tipo_manutencao   AS ENUM ('preventiva', 'corretiva', 'revisao');
CREATE TYPE status_manutencao AS ENUM ('agendada', 'em_andamento', 'concluida', 'cancelada');
CREATE TYPE tipo_combustivel  AS ENUM ('diesel', 'gasolina', 'etanol', 'gnv', 'eletrico');
CREATE TYPE tipo_veiculo      AS ENUM (
    'truck', 'bitruck', 'carreta', 'toco', 'vuc',
    'van', 'utilitario', 'outro'
);
CREATE TYPE tipo_cnh          AS ENUM ('A', 'B', 'C', 'D', 'E', 'AB', 'AC', 'AD', 'AE');
CREATE TYPE status_finalizacao AS ENUM ('pendente', 'aprovada', 'rejeitada');

-- ============================================================
--  MOTORISTAS
-- ============================================================

CREATE TABLE motoristas (
    id                  UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome                VARCHAR(150)    NOT NULL,
    cpf                 BYTEA           NOT NULL UNIQUE,  -- pgp_sym_encrypt
    cpf_hash            TEXT            NOT NULL UNIQUE,  -- digest(cpf, 'sha256') para lookup
    numero_cnh          BYTEA           NOT NULL UNIQUE,  -- pgp_sym_encrypt
    tipo_cnh            tipo_cnh        NOT NULL,
    validade_cnh        DATE            NOT NULL,
    telefone            VARCHAR(20),
    email               VARCHAR(150),
    -- endereço embutido (simples, sem tabela separada por ora)
    endereco_logradouro VARCHAR(200),
    endereco_numero     VARCHAR(10),
    endereco_complemento VARCHAR(60),
    endereco_bairro     VARCHAR(100),
    endereco_cidade     VARCHAR(100),
    endereco_uf         CHAR(2),
    endereco_cep        VARCHAR(10),
    -- dados profissionais
    data_admissao       DATE,
    status              status_motorista NOT NULL DEFAULT 'ativo',
    foto_url            TEXT,
    observacoes         TEXT,
    -- auditoria
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_motoristas_cpf_hash ON motoristas (cpf_hash);
CREATE INDEX idx_motoristas_status ON motoristas (status);
CREATE INDEX idx_motoristas_validade_cnh ON motoristas (validade_cnh);

-- ============================================================
--  VEÍCULOS
-- ============================================================

CREATE TABLE veiculos (
    id                      UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    placa                   VARCHAR(10)     NOT NULL UNIQUE,
    modelo                  VARCHAR(100)    NOT NULL,
    marca                   VARCHAR(100)    NOT NULL,
    ano                     SMALLINT        NOT NULL,
    tipo                    tipo_veiculo    NOT NULL,
    capacidade_carga_kg     NUMERIC(10,2),
    renavam                 VARCHAR(11)     UNIQUE,
    km_atual                NUMERIC(12,2)   NOT NULL DEFAULT 0,
    status                  status_veiculo  NOT NULL DEFAULT 'disponivel',
    -- documentação e vencimentos
    vencimento_seguro       DATE,
    vencimento_licenciamento DATE,
    vencimento_ipva         DATE,
    seguradora              VARCHAR(100),
    numero_apolice          VARCHAR(60),
    -- observações
    observacoes             TEXT,
    -- auditoria
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_veiculos_placa  ON veiculos (placa);
CREATE INDEX idx_veiculos_status ON veiculos (status);

-- ============================================================
--  TIPOS DE CARGA
-- ============================================================

CREATE TABLE tipos_carga (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome        VARCHAR(100) NOT NULL UNIQUE,
    descricao   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- seeds básicos
INSERT INTO tipos_carga (nome) VALUES
    ('Carga Geral'), ('Frigorificada'), ('Perigosa'),
    ('Granel Sólido'), ('Granel Líquido'), ('Conteinerizada'),
    ('Viva'), ('Veículos'), ('Mudança');

-- ============================================================
--  CLIENTES
-- ============================================================

CREATE TABLE clientes (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome        VARCHAR(150) NOT NULL,
    cpf_cnpj    VARCHAR(20),
    telefone    VARCHAR(20),
    email       VARCHAR(150),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_clientes_nome ON clientes USING gin (to_tsvector('portuguese', nome));

-- ============================================================
--  VIAGENS
-- ============================================================

CREATE TABLE viagens (
    id                  UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    motorista_id        UUID            NOT NULL REFERENCES motoristas (id),
    veiculo_id          UUID            NOT NULL REFERENCES veiculos (id),
    cliente_id          UUID            REFERENCES clientes (id),
    -- itinerário
    origem_cidade       VARCHAR(100)    NOT NULL,
    origem_uf           CHAR(2)         NOT NULL,
    destino_cidade      VARCHAR(100)    NOT NULL,
    destino_uf          CHAR(2)         NOT NULL,
    data_saida          TIMESTAMPTZ     NOT NULL,
    data_chegada_prevista TIMESTAMPTZ,
    data_chegada_real   TIMESTAMPTZ,
    distancia_km        NUMERIC(10,2),
    -- carga
    tipo_carga_id       UUID            REFERENCES tipos_carga (id),
    peso_carga_kg       NUMERIC(10,2),
    -- financeiro
    valor_frete         NUMERIC(12,2),
    -- odômetro
    km_inicial          NUMERIC(12,2)   NOT NULL,
    km_final            NUMERIC(12,2),
    -- controle
    status              status_viagem   NOT NULL DEFAULT 'pendente',
    observacoes         TEXT,
    -- auditoria
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_viagens_motorista ON viagens (motorista_id);
CREATE INDEX idx_viagens_veiculo   ON viagens (veiculo_id);
CREATE INDEX idx_viagens_status    ON viagens (status);
CREATE INDEX idx_viagens_data_saida ON viagens (data_saida);

-- ============================================================
--  DOCUMENTOS DAS VIAGENS
-- ============================================================

CREATE TABLE viagem_documentos (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    viagem_id   UUID        NOT NULL REFERENCES viagens (id) ON DELETE CASCADE,
    nome        VARCHAR(200) NOT NULL,
    tipo        VARCHAR(10)  NOT NULL CHECK (tipo IN ('pdf', 'xml')),
    url         TEXT        NOT NULL,
    tamanho_bytes BIGINT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_viagem_docs_viagem ON viagem_documentos (viagem_id);

-- ============================================================
--  HISTÓRICO DE ALTERAÇÕES DAS VIAGENS
-- ============================================================

CREATE TABLE viagem_historico (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    viagem_id       UUID        NOT NULL REFERENCES viagens (id) ON DELETE CASCADE,
    usuario_tipo    VARCHAR(20) NOT NULL, -- 'admin' | 'motorista'
    usuario_id      UUID        NOT NULL,
    campo_alterado  VARCHAR(60),
    valor_anterior  TEXT,
    valor_novo      TEXT,
    descricao       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_viagem_historico_viagem ON viagem_historico (viagem_id);

-- ============================================================
--  PARADAS DA VIAGEM
-- ============================================================

CREATE TABLE viagem_paradas (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    viagem_id   UUID        NOT NULL REFERENCES viagens (id) ON DELETE CASCADE,
    descricao   TEXT        NOT NULL,
    latitude    NUMERIC(10,7),
    longitude   NUMERIC(10,7),
    registrado_em TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_viagem_paradas_viagem ON viagem_paradas (viagem_id);

-- ============================================================
--  SOLICITAÇÕES DE FINALIZAÇÃO DE VIAGEM
-- ============================================================

CREATE TABLE viagem_finalizacoes (
    id              UUID                PRIMARY KEY DEFAULT uuid_generate_v4(),
    viagem_id       UUID                NOT NULL REFERENCES viagens (id) ON DELETE CASCADE,
    km_final        NUMERIC(12,2)       NOT NULL,
    status          status_finalizacao  NOT NULL DEFAULT 'pendente',
    observacao_motorista TEXT,
    observacao_admin     TEXT,
    solicitado_em   TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    respondido_em   TIMESTAMPTZ
);

CREATE INDEX idx_viagem_finalizacoes_viagem ON viagem_finalizacoes (viagem_id);
CREATE INDEX idx_viagem_finalizacoes_status ON viagem_finalizacoes (status);

-- ============================================================
--  ABASTECIMENTOS
-- ============================================================

CREATE TABLE abastecimentos (
    id                  UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    viagem_id           UUID            REFERENCES viagens (id),
    veiculo_id          UUID            NOT NULL REFERENCES veiculos (id),
    motorista_id        UUID            NOT NULL REFERENCES motoristas (id),
    tipo_combustivel    tipo_combustivel NOT NULL DEFAULT 'diesel',
    km_atual            NUMERIC(12,2)   NOT NULL,
    litros              NUMERIC(8,3)    NOT NULL,
    valor_por_litro     NUMERIC(8,3)    NOT NULL,
    valor_total         NUMERIC(12,2)   GENERATED ALWAYS AS (litros * valor_por_litro) STORED,
    fornecedor          VARCHAR(150),
    foto_url            TEXT,
    registrado_em       TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_abastecimentos_viagem   ON abastecimentos (viagem_id);
CREATE INDEX idx_abastecimentos_veiculo  ON abastecimentos (veiculo_id);
CREATE INDEX idx_abastecimentos_data     ON abastecimentos (registrado_em);

-- ============================================================
--  OCORRÊNCIAS
-- ============================================================

CREATE TABLE ocorrencias (
    id              UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    viagem_id       UUID            REFERENCES viagens (id),
    veiculo_id      UUID            REFERENCES veiculos (id),
    motorista_id    UUID            NOT NULL REFERENCES motoristas (id),
    tipo            tipo_ocorrencia NOT NULL,
    descricao       TEXT,
    audio_url       TEXT,
    latitude        NUMERIC(10,7),
    longitude       NUMERIC(10,7),
    registrado_em   TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ocorrencias_viagem    ON ocorrencias (viagem_id);
CREATE INDEX idx_ocorrencias_veiculo   ON ocorrencias (veiculo_id);
CREATE INDEX idx_ocorrencias_motorista ON ocorrencias (motorista_id);
CREATE INDEX idx_ocorrencias_data      ON ocorrencias (registrado_em);

-- ============================================================
--  MÍDIAS DAS OCORRÊNCIAS
-- ============================================================

CREATE TABLE ocorrencia_midias (
    id            UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    ocorrencia_id UUID        NOT NULL REFERENCES ocorrencias (id) ON DELETE CASCADE,
    tipo          VARCHAR(10) NOT NULL CHECK (tipo IN ('foto', 'video')),
    url           TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ocorrencia_midias_ocorrencia ON ocorrencia_midias (ocorrencia_id);

-- ============================================================
--  MANUTENÇÕES
-- ============================================================

CREATE TABLE manutencoes (
    id              UUID                PRIMARY KEY DEFAULT uuid_generate_v4(),
    veiculo_id      UUID                NOT NULL REFERENCES veiculos (id),
    tipo            tipo_manutencao     NOT NULL,
    status          status_manutencao   NOT NULL DEFAULT 'agendada',
    descricao       TEXT                NOT NULL,
    oficina         VARCHAR(150),
    km_na_manutencao NUMERIC(12,2),
    km_proxima_manutencao NUMERIC(12,2),
    data_agendada   DATE,
    data_conclusao  DATE,
    custo           NUMERIC(12,2),
    observacoes     TEXT,
    created_at      TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_manutencoes_veiculo ON manutencoes (veiculo_id);
CREATE INDEX idx_manutencoes_status  ON manutencoes (status);
CREATE INDEX idx_manutencoes_data    ON manutencoes (data_agendada);

-- ============================================================
--  NOTIFICAÇÕES
-- ============================================================

CREATE TABLE notificacoes (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    -- destinatário (null = broadcast para admins)
    destinatario_tipo VARCHAR(20),    -- 'admin' | 'motorista'
    destinatario_id   UUID,
    -- origem da ação
    origem_tipo     VARCHAR(30),      -- 'motorista' | 'sistema'
    origem_id       UUID,
    titulo          VARCHAR(200)    NOT NULL,
    mensagem        TEXT,
    lida            BOOLEAN         NOT NULL DEFAULT FALSE,
    referencia_tipo VARCHAR(40),     -- 'viagem' | 'abastecimento' | 'ocorrencia' | ...
    referencia_id   UUID,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notificacoes_destinatario ON notificacoes (destinatario_tipo, destinatario_id, lida);
CREATE INDEX idx_notificacoes_data         ON notificacoes (created_at DESC);

-- ============================================================
--  USUÁRIOS ADMINISTRATIVOS
-- ============================================================

CREATE TABLE usuarios (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    nome            VARCHAR(150) NOT NULL,
    email           VARCHAR(150) NOT NULL UNIQUE,
    senha_hash      TEXT        NOT NULL,
    role            VARCHAR(20) NOT NULL DEFAULT 'admin' CHECK (role IN ('superadmin', 'admin', 'operador')),
    ativo           BOOLEAN     NOT NULL DEFAULT TRUE,
    ultimo_acesso   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
--  CREDENCIAIS DE ACESSO DOS MOTORISTAS
-- ============================================================

CREATE TABLE motorista_credenciais (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    motorista_id    UUID        NOT NULL UNIQUE REFERENCES motoristas (id) ON DELETE CASCADE,
    senha_hash      TEXT        NOT NULL,
    deve_trocar_senha BOOLEAN   NOT NULL DEFAULT TRUE,
    ativo           BOOLEAN     NOT NULL DEFAULT TRUE,
    ultimo_acesso   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
--  CONVITES DE ADMIN
-- ============================================================

CREATE TABLE convites_admin (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    email       VARCHAR(150) NOT NULL,
    token_hash  TEXT        NOT NULL UNIQUE,
    role        VARCHAR(20) NOT NULL DEFAULT 'admin' CHECK (role IN ('superadmin', 'admin', 'operador')),
    usado       BOOLEAN     NOT NULL DEFAULT FALSE,
    expira_em   TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours',
    criado_por  UUID        REFERENCES usuarios (id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_convites_token ON convites_admin (token_hash);
CREATE INDEX idx_convites_email ON convites_admin (email);

-- ============================================================
--  REFRESH TOKENS DE AUTENTICAÇÃO
-- ============================================================

CREATE TABLE auth_refresh_tokens (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    token_id    TEXT        NOT NULL UNIQUE,
    actor_id    UUID        NOT NULL,
    actor_type  VARCHAR(20) NOT NULL CHECK (actor_type IN ('admin', 'motorista')),
    token_hash  TEXT        NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_refresh_tokens_actor ON auth_refresh_tokens (actor_type, actor_id);
CREATE INDEX idx_auth_refresh_tokens_token_id ON auth_refresh_tokens (token_id);
CREATE INDEX idx_auth_refresh_tokens_expires_at ON auth_refresh_tokens (expires_at);

-- ============================================================
--  VIEWS ÚTEIS
-- ============================================================

-- Resumo diário (usado pelo dashboard)
CREATE OR REPLACE VIEW vw_dashboard_hoje AS
SELECT
    (SELECT COUNT(*) FROM viagens WHERE DATE(data_saida) = CURRENT_DATE)                         AS total_viagens_hoje,
    (SELECT COUNT(*) FROM viagens WHERE status = 'em_andamento')                                  AS viagens_em_andamento,
    (SELECT COUNT(*) FROM viagens WHERE status = 'pendente')                                      AS viagens_pendentes,
    (SELECT COUNT(*) FROM viagens WHERE status = 'concluida' AND DATE(data_chegada_real) = CURRENT_DATE) AS viagens_concluidas_hoje,
    (SELECT COUNT(*) FROM veiculos WHERE status = 'em_uso')                                       AS veiculos_em_uso,
    (SELECT COUNT(*) FROM veiculos WHERE status = 'disponivel')                                   AS veiculos_disponiveis,
    (SELECT COUNT(*) FROM veiculos WHERE status = 'manutencao')                                   AS veiculos_em_manutencao,
    (SELECT COALESCE(SUM(valor_total), 0) FROM abastecimentos WHERE DATE(registrado_em) = CURRENT_DATE) AS gasto_abastecimento_hoje,
    (SELECT COALESCE(SUM(custo), 0)       FROM manutencoes     WHERE data_conclusao = CURRENT_DATE)     AS gasto_manutencao_hoje;

-- Consumo médio por veículo (km/l)
CREATE OR REPLACE VIEW vw_consumo_veiculo AS
SELECT
    v.id             AS veiculo_id,
    v.placa,
    v.modelo,
    COUNT(a.id)      AS total_abastecimentos,
    SUM(a.litros)    AS total_litros,
    MAX(a.km_atual) - MIN(a.km_atual) AS km_percorridos,
    CASE
        WHEN SUM(a.litros) > 0
        THEN ROUND((MAX(a.km_atual) - MIN(a.km_atual)) / SUM(a.litros), 2)
    END              AS consumo_km_por_litro,
    SUM(a.valor_total) AS custo_combustivel
FROM veiculos v
LEFT JOIN abastecimentos a ON a.veiculo_id = v.id
GROUP BY v.id, v.placa, v.modelo;

-- Custo total por veículo
CREATE OR REPLACE VIEW vw_custo_total_veiculo AS
SELECT
    v.id        AS veiculo_id,
    v.placa,
    v.modelo,
    COALESCE(SUM(a.valor_total), 0)  AS custo_combustivel,
    COALESCE(SUM(m.custo), 0)        AS custo_manutencao,
    COALESCE(SUM(a.valor_total), 0) + COALESCE(SUM(m.custo), 0) AS custo_total
FROM veiculos v
LEFT JOIN abastecimentos a ON a.veiculo_id = v.id
LEFT JOIN manutencoes    m ON m.veiculo_id = v.id
GROUP BY v.id, v.placa, v.modelo;

-- Indicadores do motorista
CREATE OR REPLACE VIEW vw_indicadores_motorista AS
SELECT
    m.id                                                                        AS motorista_id,
    m.nome,
    COUNT(DISTINCT v.id)                                                        AS total_viagens,
    COALESCE(SUM(v.km_final - v.km_inicial) FILTER (WHERE v.km_final IS NOT NULL), 0) AS total_km_rodados,
    COUNT(DISTINCT o.id)                                                        AS total_ocorrencias,
    COALESCE(SUM(DISTINCT v.valor_frete), 0)                                    AS total_frete_gerado
FROM motoristas m
LEFT JOIN viagens     v ON v.motorista_id = m.id
LEFT JOIN ocorrencias o ON o.motorista_id = m.id
GROUP BY m.id, m.nome;

-- Alertas de vencimento (CNH, seguro, licenciamento)
CREATE OR REPLACE VIEW vw_alertas AS
-- CNH vencendo em 30 dias
SELECT
    'cnh_vencimento'    AS tipo_alerta,
    'motorista'         AS entidade,
    m.id                AS entidade_id,
    m.nome              AS descricao,
    m.validade_cnh      AS data_referencia,
    m.validade_cnh - CURRENT_DATE AS dias_restantes
FROM motoristas m
WHERE m.validade_cnh BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days'
  AND m.status = 'ativo'

UNION ALL

-- Seguro vencendo em 30 dias
SELECT
    'seguro_vencimento',
    'veiculo',
    v.id,
    v.placa || ' - ' || v.modelo,
    v.vencimento_seguro,
    v.vencimento_seguro - CURRENT_DATE
FROM veiculos v
WHERE v.vencimento_seguro BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days'
  AND v.status != 'inativo'

UNION ALL

-- Licenciamento vencendo em 30 dias
SELECT
    'licenciamento_vencimento',
    'veiculo',
    v.id,
    v.placa || ' - ' || v.modelo,
    v.vencimento_licenciamento,
    v.vencimento_licenciamento - CURRENT_DATE
FROM veiculos v
WHERE v.vencimento_licenciamento BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days'
  AND v.status != 'inativo'

UNION ALL

-- Manutenção preventiva próxima (por km)
SELECT
    'manutencao_preventiva',
    'veiculo',
    ve.id,
    ve.placa || ' - ' || ve.modelo,
    NULL::DATE,
    NULL::INTEGER
FROM manutencoes m
JOIN veiculos ve ON ve.id = m.veiculo_id
WHERE m.tipo = 'preventiva'
  AND m.km_proxima_manutencao IS NOT NULL
  AND ve.km_atual >= (m.km_proxima_manutencao - 1000)
  AND m.status = 'agendada'

ORDER BY dias_restantes NULLS LAST;

-- ============================================================
--  FUNÇÕES / TRIGGERS
-- ============================================================

-- Atualiza updated_at automaticamente
CREATE OR REPLACE FUNCTION fn_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_motoristas_updated_at  BEFORE UPDATE ON motoristas  FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
CREATE TRIGGER trg_veiculos_updated_at    BEFORE UPDATE ON veiculos     FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
CREATE TRIGGER trg_viagens_updated_at     BEFORE UPDATE ON viagens      FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
CREATE TRIGGER trg_manutencoes_updated_at BEFORE UPDATE ON manutencoes  FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
CREATE TRIGGER trg_clientes_updated_at    BEFORE UPDATE ON clientes      FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
CREATE TRIGGER trg_usuarios_updated_at    BEFORE UPDATE ON usuarios      FOR EACH ROW EXECUTE FUNCTION fn_updated_at();
CREATE TRIGGER trg_auth_refresh_tokens_updated_at BEFORE UPDATE ON auth_refresh_tokens FOR EACH ROW EXECUTE FUNCTION fn_updated_at();

-- Ao concluir abastecimento, atualiza km_atual do veículo
CREATE OR REPLACE FUNCTION fn_atualiza_km_veiculo()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    UPDATE veiculos SET km_atual = NEW.km_atual WHERE id = NEW.veiculo_id AND km_atual < NEW.km_atual;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_abastecimento_km AFTER INSERT ON abastecimentos FOR EACH ROW EXECUTE FUNCTION fn_atualiza_km_veiculo();

-- Valida que km informado no abastecimento não é menor que o último
CREATE OR REPLACE FUNCTION fn_valida_km_abastecimento()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
DECLARE
    ultimo_km NUMERIC;
BEGIN
    SELECT km_atual INTO ultimo_km
    FROM abastecimentos
    WHERE veiculo_id = NEW.veiculo_id
    ORDER BY registrado_em DESC
    LIMIT 1;

    IF ultimo_km IS NOT NULL AND NEW.km_atual < ultimo_km THEN
        RAISE EXCEPTION 'O km informado (%) é menor que o último registrado (%)', NEW.km_atual, ultimo_km;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_valida_km_abast BEFORE INSERT ON abastecimentos FOR EACH ROW EXECUTE FUNCTION fn_valida_km_abastecimento();

-- Ao atualizar status da viagem para 'em_andamento', marca veículo como 'em_uso'
-- Ao concluir/cancelar, libera o veículo
CREATE OR REPLACE FUNCTION fn_status_veiculo_por_viagem()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.status = 'em_andamento' THEN
        UPDATE veiculos SET status = 'em_uso'      WHERE id = NEW.veiculo_id;
    ELSIF NEW.status IN ('concluida', 'cancelada') THEN
        UPDATE veiculos SET status = 'disponivel'  WHERE id = NEW.veiculo_id;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_viagem_status_veiculo AFTER UPDATE OF status ON viagens FOR EACH ROW EXECUTE FUNCTION fn_status_veiculo_por_viagem();

-- ============================================================
--  LUCRO POR VIAGEM (função auxiliar para relatórios)
-- ============================================================
CREATE OR REPLACE FUNCTION fn_lucro_viagem(p_viagem_id UUID)
RETURNS NUMERIC LANGUAGE sql STABLE AS $$
    SELECT
        COALESCE(v.valor_frete, 0)
        - COALESCE((SELECT SUM(a.valor_total) FROM abastecimentos a WHERE a.viagem_id = p_viagem_id), 0)
        - COALESCE((SELECT SUM(m.custo)       FROM manutencoes    m
                    JOIN ocorrencias o ON o.viagem_id = p_viagem_id AND o.veiculo_id = m.veiculo_id), 0)
    FROM viagens v
    WHERE v.id = p_viagem_id;
$$;


-- ============================================================
--  ADMINS DE TESTE
--  Requer: CREATE EXTENSION pgcrypto (já está no schema)
--  Senhas em texto puro estão nos comentários — apague depois
-- ============================================================

INSERT INTO usuarios (id, nome, email, senha_hash, role, ativo)
VALUES
  (
    uuid_generate_v4(),
    'Super Admin',
    'super@transportadora.com',
    crypt('Super@123', gen_salt('bf', 10)),  -- bcrypt custo 10
    'superadmin',
    TRUE
  ),
  (
    uuid_generate_v4(),
    'Admin Geral',
    'admin@transportadora.com',
    crypt('Admin@123', gen_salt('bf', 10)),
    'admin',
    TRUE
  ),
  (
    uuid_generate_v4(),
    'Operador',
    'operador@transportadora.com',
    crypt('Oper@123', gen_salt('bf', 10)),
    'operador',
    TRUE
  );

SELECT nome, email, role, LEFT(senha_hash, 30) || '...' AS hash_parcial
FROM usuarios
ORDER BY created_at;


-- ============================================================
--  FIM DO SCHEMA
-- ============================================================
