# API Go

## Autores

- [Gabriel Vicente](https://github.com/gabrielvicentm)
- [João Vitor Gouvea](https://github.com/gouveazs)
- [Leon Kennedy](https://github.com/Kennedys-Leon)
- [Gabriel Gusmão](https://github.com/Gusmaozz)

## Sobre o projeto

A **API Go** e o backend principal do projeto. Ela fornece os recursos consumidos pelo painel administrativo e pelas funcionalidades voltadas ao motorista, concentrando autenticacao, controle de acesso, cadastros operacionais, viagens, ocorrencias, abastecimentos, manutencoes, notificacoes, arquivos e relatorios.

O projeto segue uma organizacao em camadas, separando handlers HTTP, regras de negocio, repositorios, entidades de dominio, middlewares e configuracoes.

## Tecnologias utilizadas

- **Go**: linguagem principal do backend.
- **Gin**: framework HTTP usado para declarar rotas, middlewares e handlers.
- **PostgreSQL**: banco de dados relacional da aplicacao.
- **pgx**: driver e pool de conexoes com PostgreSQL.
- **JWT**: autenticacao por tokens de acesso e refresh.
- **godotenv**: carregamento das variaveis de ambiente.
- **AWS SDK S3 compativel**: integracao com armazenamento de arquivos, como Cloudflare R2.
- **Expo Push Notifications**: envio de notificacoes push.
- **SMTP**: envio de mensagens para fluxo de redefinicao de senha.

## Estrutura do projeto

```text
api-go/
├── cmd/
│   └── api/
│       └── main.go                  # Ponto de entrada da API
├── config/
│   └── config.go                    # Configuracao de ambiente e banco
├── docs/
│   └── push_notifications.md        # Documentacao auxiliar de notificacoes
├── internal/
│   ├── domain/                      # Entidades, contratos e erros de dominio
│   ├── handler/                     # Handlers HTTP e registro das rotas
│   ├── middleware/                  # CORS, autenticacao e autorizacao
│   ├── repository/                  # Acesso ao banco de dados
│   ├── security/                    # Tokens, chaves e seguranca
│   └── service/                     # Regras de negocio e integracoes externas
├── esquilo.sql                      # Script de criacao do banco
├── inserts_teste.sql                # Dados de teste
├── go.mod
└── go.sum
```

## Arquitetura

A aplicacao e inicializada em `cmd/api/main.go`. Nesse arquivo sao carregadas as variaveis de ambiente, criada a conexao com o banco, inicializados os repositorios, servicos, middlewares e handlers, e registradas as rotas da API.

O fluxo principal segue esta divisao:

- **Handlers** recebem as requisicoes HTTP, validam entradas e montam respostas.
- **Services** concentram regras de negocio e integracoes externas.
- **Repositories** fazem a comunicacao com o PostgreSQL.
- **Domain** define os modelos e contratos usados entre as camadas.
- **Middlewares** aplicam CORS, autenticacao, autorizacao e seguranca de rotas internas.
- **Security** gerencia tokens JWT, chaves e criptografia de dados sensiveis.

## Grupos de rotas

### Autenticacao

Prefixo: `/auth`

- `POST /auth/login`: login administrativo.
- `POST /auth/admin/login`: login administrativo.
- `POST /auth/motorista/login`: login do motorista.
- `POST /auth/refresh`: renovacao de token.
- `POST /auth/logout`: encerramento de sessao.
- `POST /auth/reset-password`: redefinicao de senha.
- `GET /auth/me`: dados do usuario autenticado.
- `POST /auth/change-password`: alteracao de senha.

### Rotas administrativas

Prefixo: `/admin`

As rotas administrativas exigem autenticacao e ator do tipo `admin`.

Principais modulos:

- Dashboard e alertas.
- Usuarios administrativos.
- Funcionarios.
- Folha de pagamento.
- Motoristas.
- Veiculos.
- Clientes.
- Tipos de carga.
- Viagens.
- Ocorrencias.
- Abastecimentos.
- Notificacoes.
- Manutencoes.
- Relatorios.
- Historico de alteracoes.

### Rotas do motorista

Prefixo: `/motorista`

As rotas do motorista exigem autenticacao e ator do tipo `motorista`.

Principais recursos:

- Perfil do motorista.
- Listagem de viagens.
- Viagem atual.
- Historico de viagens.
- Registro e finalizacao de paradas.
- Solicitacao de finalizacao de viagem.
- Cadastro e consulta de ocorrencias.
- Cadastro e consulta de abastecimentos.
- Consulta e leitura de notificacoes.
- Registro e remocao de token push.

### Rotas internas

Prefixo: `/internal`

As rotas internas sao protegidas por token proprio, definido em variavel de ambiente. Elas sao usadas para comunicacao entre servicos, como criacao interna de notificacoes.

### Arquivos estaticos

- `/uploads`: exposicao de arquivos estaticos locais.

## Principais rotas administrativas

### Dashboard

- `GET /admin/dashboard`
- `GET /admin/alertas`

### Usuarios

- `GET /admin/usuarios`
- `POST /admin/usuarios/password-reset-token`

### Funcionarios e folha de pagamento

- `GET /admin/funcionarios`
- `POST /admin/funcionarios`
- `GET /admin/funcionarios/:id`
- `PUT /admin/funcionarios/:id`
- `DELETE /admin/funcionarios/:id`
- `PATCH /admin/funcionarios/:id/status`
- `POST /admin/funcionarios/:id/foto`
- `GET /admin/folha-pagamento`
- `GET /admin/funcionarios/:id/folha-pagamento`
- `PUT /admin/funcionarios/:id/folha-pagamento`

### Motoristas

- `GET /admin/motoristas`
- `POST /admin/motoristas`
- `GET /admin/motoristas/:id`
- `PUT /admin/motoristas/:id`
- `DELETE /admin/motoristas/:id`
- `PATCH /admin/motoristas/:id/status`
- `POST /admin/motoristas/:id/foto`
- `GET /admin/motoristas/:id/indicadores`
- `GET /admin/motoristas/:id/viagens`
- `GET /admin/motoristas/:id/ocorrencias`

### Veiculos

- `GET /admin/veiculos`
- `GET /admin/veiculos/custos-totais`
- `GET /admin/veiculos/consumo-medio`
- `POST /admin/veiculos`
- `GET /admin/veiculos/:id`
- `PUT /admin/veiculos/:id`
- `DELETE /admin/veiculos/:id`
- `GET /admin/veiculos/:id/custos`
- `GET /admin/veiculos/:id/consumo`
- `GET /admin/veiculos/:id/historico`

### Clientes

- `GET /admin/clientes`
- `POST /admin/clientes`
- `GET /admin/clientes/:id`
- `PUT /admin/clientes/:id`
- `DELETE /admin/clientes/:id`

### Tipos de carga

- `GET /admin/tipos-carga`
- `POST /admin/tipos-carga`
- `GET /admin/tipos-carga/:id`
- `PUT /admin/tipos-carga/:id`
- `DELETE /admin/tipos-carga/:id`

### Viagens

- `GET /admin/viagens`
- `POST /admin/viagens`
- `GET /admin/viagens/:id`
- `PUT /admin/viagens/:id`
- `DELETE /admin/viagens/:id`
- `POST /admin/viagens/:id/finalizar`
- `GET /admin/viagens/:id/historico`
- `GET /admin/viagens/:id/documentos`
- `POST /admin/viagens/:id/documentos`
- `GET /admin/viagens/:id/documentos/:documentoId`
- `GET /admin/viagens/:id/finalizacoes`
- `POST /admin/viagens/:id/finalizacoes/:finalizacaoId/aprovar`
- `POST /admin/viagens/:id/finalizacoes/:finalizacaoId/rejeitar`

### Ocorrencias e abastecimentos

- `GET /admin/ocorrencias`
- `GET /admin/ocorrencias/:id`
- `GET /admin/abastecimentos`
- `GET /admin/abastecimentos/:id`

### Notificacoes

- `GET /admin/notificacoes`
- `GET /admin/notificacoes/stream`
- `POST /admin/notificacoes/push-token`
- `DELETE /admin/notificacoes/push-token`
- `PATCH /admin/notificacoes/:id/lida`

### Manutencoes

- `GET /admin/manutencoes`
- `POST /admin/manutencoes`
- `GET /admin/manutencoes/:id`
- `PUT /admin/manutencoes/:id`
- `GET /admin/veiculos/:id/manutencoes`

### Relatorios

- `GET /admin/relatorios/viagens`
- `GET /admin/relatorios/combustivel`
- `GET /admin/relatorios/manutencoes`
- `GET /admin/relatorios/custos`
- `GET /admin/relatorios/desempenho`
- `GET /admin/relatorios/lucro-por-viagem`
- `GET /admin/relatorios/exportacoes/xlsx`
- `GET /admin/relatorios/exportacoes/csv`

## Variaveis de ambiente

A API carrega as configuracoes a partir de um arquivo `.env`.

### Banco de dados

```env
DB_USER=
DB_PASSWORD=
DB_HOST=
DB_PORT=
DB_NAME=
```

### Autenticacao e seguranca

```env
JWT_ACCESS_SECRET=
JWT_REFRESH_SECRET=
JWT_ACCESS_TTL=
JWT_REFRESH_TTL=
JWT_ISSUER=
REFRESH_TOKEN_PEPPER=
DATA_ENCRYPTION_KEY=
INTERNAL_API_TOKEN=
PASSWORD_RESET_TOKEN_TTL=
```

### CORS

```env
CORS_ALLOW_ORIGINS=
```

### Armazenamento de arquivos

```env
R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET_NAME=
R2_REGION=
R2_ENDPOINT=
R2_PUBLIC_BASE_URL=
R2_FUNCIONARIOS_PREFIX=
R2_MOTORISTAS_PREFIX=
R2_VIAGENS_DOCUMENTOS_PREFIX=
R2_ABASTECIMENTOS_PREFIX=
R2_OCORRENCIAS_PREFIX=
```

### Notificacoes push

```env
EXPO_PUSH_ENDPOINT=
EXPO_ACCESS_TOKEN=
```

### SMTP

```env
SMTP_HOST=
SMTP_PORT=
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM_EMAIL=
SMTP_FROM_NAME=
```

## Banco de dados

O projeto possui dois scripts SQL principais:

- `esquilo.sql`: estrutura do banco de dados.
- `inserts_teste.sql`: dados de teste para desenvolvimento.

Antes de iniciar a API, crie o banco PostgreSQL e execute o script `esquilo.sql`. Se quiser popular o ambiente com dados de exemplo, execute tambem `inserts_teste.sql`.

## Como executar

### Requisitos

- Go instalado.
- PostgreSQL configurado.
- Arquivo `.env` preenchido.
- Banco criado com o script `esquilo.sql`.

### Instalar dependencias

```bash
go mod download
```

### Rodar a API

```bash
go run ./cmd/api
```

A API sera iniciada em:

```text
http://localhost:8080
```

## Resumo das funcionalidades

- Autenticacao de administradores e motoristas.
- Controle de sessoes com access token e refresh token.
- Protecao de rotas por tipo de usuario.
- CRUD de funcionarios, motoristas, veiculos, clientes e tipos de carga.
- Gestao de viagens, paradas, documentos e finalizacoes.
- Registro e consulta de ocorrencias.
- Registro e consulta de abastecimentos.
- Controle de manutencoes.
- Folha de pagamento de funcionarios.
- Dashboard com indicadores e alertas.
- Notificacoes em tempo real e notificacoes push.
- Upload e consulta de arquivos em storage compativel com S3.
- Relatorios operacionais e exportacoes.
- Historico de alteracoes para auditoria.
