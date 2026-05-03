package domain

import "time"

type ViagemListFilter struct {
	Search       string
	Status       string
	MotoristaID  string
	VeiculoID    string
	ClienteID    string
	DataSaidaDe  string
	DataSaidaAte string
	Page         int
	Limit        int
}

type ViagemCreateRequest struct {
	MotoristaID         string `json:"motorista_id" binding:"required,uuid"`
	VeiculoID           string `json:"veiculo_id" binding:"required,uuid"`
	ClienteID           string `json:"cliente_id" binding:"omitempty,uuid"`
	OrigemCidade        string `json:"origem_cidade" binding:"required,min=2"`
	OrigemUF            string `json:"origem_uf" binding:"required,len=2"`
	DestinoCidade       string `json:"destino_cidade" binding:"required,min=2"`
	DestinoUF           string `json:"destino_uf" binding:"required,len=2"`
	DataSaida           string `json:"data_saida" binding:"required"`
	DataChegadaPrevista string `json:"data_chegada_prevista"`
	DistanciaKM         string `json:"distancia_km"`
	TipoCargaID         string `json:"tipo_carga_id" binding:"omitempty,uuid"`
	PesoCargaKG         string `json:"peso_carga_kg"`
	ValorFrete          string `json:"valor_frete"`
	KMInicial           string `json:"km_inicial" binding:"required"`
	Observacoes         string `json:"observacoes"`
}

type ViagemUpdateRequest struct {
	MotoristaID         string `json:"motorista_id" binding:"required,uuid"`
	VeiculoID           string `json:"veiculo_id" binding:"required,uuid"`
	ClienteID           string `json:"cliente_id" binding:"omitempty,uuid"`
	OrigemCidade        string `json:"origem_cidade" binding:"required,min=2"`
	OrigemUF            string `json:"origem_uf" binding:"required,len=2"`
	DestinoCidade       string `json:"destino_cidade" binding:"required,min=2"`
	DestinoUF           string `json:"destino_uf" binding:"required,len=2"`
	DataSaida           string `json:"data_saida" binding:"required"`
	DataChegadaPrevista string `json:"data_chegada_prevista"`
	DataChegadaReal     string `json:"data_chegada_real"`
	DistanciaKM         string `json:"distancia_km"`
	TipoCargaID         string `json:"tipo_carga_id" binding:"omitempty,uuid"`
	PesoCargaKG         string `json:"peso_carga_kg"`
	ValorFrete          string `json:"valor_frete"`
	KMInicial           string `json:"km_inicial" binding:"required"`
	KMFinal             string `json:"km_final"`
	Status              string `json:"status"`
	Observacoes         string `json:"observacoes"`
}

type ViagemDetail struct {
	ID                  string     `json:"id"`
	MotoristaID         string     `json:"motorista_id"`
	VeiculoID           string     `json:"veiculo_id"`
	ClienteID           string     `json:"cliente_id,omitempty"`
	OrigemCidade        string     `json:"origem_cidade"`
	OrigemUF            string     `json:"origem_uf"`
	DestinoCidade       string     `json:"destino_cidade"`
	DestinoUF           string     `json:"destino_uf"`
	DataSaida           *time.Time `json:"data_saida,omitempty"`
	DataChegadaPrevista *time.Time `json:"data_chegada_prevista,omitempty"`
	DataChegadaReal     *time.Time `json:"data_chegada_real,omitempty"`
	DistanciaKM         string     `json:"distancia_km,omitempty"`
	TipoCargaID         string     `json:"tipo_carga_id,omitempty"`
	PesoCargaKG         string     `json:"peso_carga_kg,omitempty"`
	ValorFrete          string     `json:"valor_frete,omitempty"`
	KMInicial           string     `json:"km_inicial"`
	KMFinal             string     `json:"km_final,omitempty"`
	Status              string     `json:"status"`
	Observacoes         string     `json:"observacoes,omitempty"`
	CreatedAt           *time.Time `json:"created_at,omitempty"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`
}

type ViagemDocumentoItem struct {
	ID           string     `json:"id"`
	ViagemID     string     `json:"viagem_id"`
	Nome         string     `json:"nome"`
	Tipo         string     `json:"tipo"`
	URL          string     `json:"url"`
	TamanhoBytes int64      `json:"tamanho_bytes,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
}

type ViagemDocumentoCreateInput struct {
	ViagemID     string
	Nome         string
	Tipo         string
	URL          string
	TamanhoBytes int64
}

type ViagemHistoricoItem struct {
	ID            string     `json:"id"`
	ViagemID      string     `json:"viagem_id"`
	UsuarioTipo   string     `json:"usuario_tipo"`
	UsuarioID     string     `json:"usuario_id"`
	CampoAlterado string     `json:"campo_alterado,omitempty"`
	ValorAnterior string     `json:"valor_anterior,omitempty"`
	ValorNovo     string     `json:"valor_novo,omitempty"`
	Descricao     string     `json:"descricao,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
}

type ViagemParadaItem struct {
	ID           string     `json:"id"`
	ViagemID     string     `json:"viagem_id"`
	Descricao    string     `json:"descricao"`
	Latitude     string     `json:"latitude,omitempty"`
	Longitude    string     `json:"longitude,omitempty"`
	RegistradoEm *time.Time `json:"registrado_em,omitempty"`
}

type ViagemFinalizacaoItem struct {
	ID                  string     `json:"id"`
	ViagemID            string     `json:"viagem_id"`
	KMFinal             string     `json:"km_final"`
	Status              string     `json:"status"`
	ObservacaoMotorista string     `json:"observacao_motorista,omitempty"`
	ObservacaoAdmin     string     `json:"observacao_admin,omitempty"`
	SolicitadoEm        *time.Time `json:"solicitado_em,omitempty"`
	RespondidoEm        *time.Time `json:"respondido_em,omitempty"`
}

type ViagemHistoricoCreateInput struct {
	ViagemID      string
	UsuarioTipo   string
	UsuarioID     string
	CampoAlterado string
	ValorAnterior string
	ValorNovo     string
	Descricao     string
}
