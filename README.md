# api-go

API Go inicial com Docker e Docker Compose.

## O que foi criado

- Uma API HTTP minima em Go
- Um `Dockerfile` para empacotar a aplicacao
- Um `.dockerignore` para evitar copiar arquivos desnecessarios
- Um `docker-compose.yml` para subir a API facilmente

## Estrutura

```text
.
├── .dockerignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── README.md
└── cmd
    └── api
        └── main.go
```

## Como o Docker entra no seu projeto

No seu caso, o Docker serve para colocar a API dentro de um ambiente padrao.
Em vez de depender do Go instalado localmente para rodar a aplicacao, voce define
em arquivos como ela deve ser montada e executada.

O fluxo e este:

1. O `Dockerfile` diz como montar a imagem da API.
2. O Docker constroi uma imagem com seu projeto.
3. Um container e criado a partir dessa imagem.
4. O `docker-compose.yml` facilita subir esse container com um comando so.

## Arquivos importantes

### `cmd/api/main.go`

E o ponto de entrada da API. Hoje ele sobe um servidor HTTP simples com:

- `GET /health` para verificar se a API esta no ar
- `GET /` para uma mensagem inicial

### `Dockerfile`

Usa build em duas etapas:

1. Uma imagem com Go para compilar a aplicacao
2. Uma imagem final menor, so com o binario gerado

Isso deixa a imagem final mais leve.

### `.dockerignore`

Funciona como um `.gitignore` do Docker. Ele evita mandar arquivos desnecessarios
para o contexto de build, o que deixa a construcao mais rapida e limpa.

### `docker-compose.yml`

Define o servico `api`, constroi a imagem a partir do `Dockerfile` e publica a
porta `8080` da sua maquina para a `8080` do container.

## Como rodar sem Docker

Se quiser testar a API direto com Go:

```bash
go run ./cmd/api
```

Depois acesse:

- `http://localhost:8080/`
- `http://localhost:8080/health`

## Como rodar com Docker

### Build manual

```bash
docker build -t api-go .
```

### Rodar manualmente

```bash
docker run --rm -p 8080:8080 api-go
```

## Como rodar com Docker Compose

```bash
docker compose up --build
```

Depois acesse:

- `http://localhost:8080/`
- `http://localhost:8080/health`

Para derrubar os containers:

```bash
docker compose down
```

## O que voce vai fazer depois

Quando sua API crescer, o caminho natural e:

1. Adicionar variaveis de ambiente
2. Adicionar banco de dados no `docker-compose.yml`
3. Criar ambientes de desenvolvimento e producao
4. Integrar o `app-motorista` e o `web-admin` com a API

## Resumo pratico

- `Dockerfile`: ensina o Docker a montar sua API
- `image`: o pacote pronto da aplicacao
- `container`: a aplicacao rodando
- `docker compose`: sobe os servicos juntos com menos trabalho
